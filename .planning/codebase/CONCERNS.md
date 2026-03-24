# Concerns & Technical Debt

**Analysis Date:** 2026-03-24

---

## Security Concerns

### Plain-text Password Fallback in User Store
- **Risk:** If Windows DPAPI encryption fails, `Add()` silently falls back to storing the password in the JSON file with a `PLAIN:` prefix. The file is written with `0600` permissions (owner-only), but the password is readable as plain text by anyone who can read `lognojutsu_users.json`.
- **Files:** `internal/userstore/userstore.go:120-124`, `internal/userstore/userstore.go:178-179`
- **Impact:** Credential leak if the file is exfiltrated or if the DPAPI call fails silently on some machines.
- **Fix approach:** Treat DPAPI failure as a hard error and refuse to store the profile rather than degrading to plain-text storage.

### Password Transmitted in Plaintext over HTTP
- **Risk:** The tool has no TLS support. Credentials submitted via `/api/users` (POST body contains plaintext password) and the optional UI password (HTTP Basic Auth) travel over plain HTTP. If `--host 0.0.0.0` is used on a network, credentials are sniffable.
- **Files:** `internal/server/server.go:88`, `cmd/lognojutsu/main.go:23-41`
- **Current mitigation:** Warning log printed when `0.0.0.0` is used; default bind is `127.0.0.1`.
- **Recommendations:** Add optional TLS (`--cert`/`--key` flags) or document that the tool must only be used loopback.

### Password Interpolated Directly into PowerShell Strings
- **Risk:** In `TestCredentials`, the password is embedded as a double-quoted string inside a PowerShell script built with `fmt.Sprintf`. The escaping only handles `"` → `\"`. Passwords containing backtick (`` ` ``), `$`, newline, or null bytes can break out of the string or cause silent truncation.
- **Files:** `internal/userstore/userstore.go:310-328`
- **Current mitigation:** The `runCommandAs` path in `executor.go` has more careful escaping (backtick, `"`, `$` handled), but `TestCredentials` does not replicate this.
- **Recommendations:** Use `-SecureString` constructed from a byte array rather than interpolating the raw password into a string literal.

### DPAPI Credential Encryption in `encryptDPAPI` Uses String Interpolation
- **Risk:** `ConvertTo-SecureString "PLAIN_PASSWORD" -AsPlainText -Force` is constructed with `fmt.Sprintf` and only escapes `"` → `` `" ``. Passwords with backticks break the script.
- **Files:** `internal/userstore/userstore.go:345-354`
- **Fix approach:** Accept the password via stdin using `-Command "$input | ConvertTo-SecureString -AsPlainText -Force"` to avoid shell injection.

### Wildcard CORS on All API Responses
- **Risk:** Every API response includes `Access-Control-Allow-Origin: *` regardless of endpoint sensitivity. While the server defaults to loopback, any page open in the browser can call the API if the user is logged in (CSRF-via-CORS for state-mutating endpoints).
- **Files:** `internal/server/server.go:102`
- **Fix approach:** Only set the CORS header on endpoints that genuinely need cross-origin access, or restrict the origin to the same host.

### No CSRF Protection on State-Mutating Endpoints
- **Risk:** `/api/start`, `/api/stop`, `/api/users`, `/api/users/delete`, `/api/prepare` are state-mutating POST endpoints. There is no CSRF token, SameSite cookie, or origin check. An attacker page could trigger a simulation start against a locally-running instance.
- **Files:** `internal/server/server.go:64-80`
- **Fix approach:** Check `Origin` header against allowed origins, or add a CSRF token header requirement.

### No Request Body Size Limit
- **Risk:** `json.NewDecoder(r.Body).Decode(...)` has no `http.MaxBytesReader` wrapper. A malicious or misbehaving client can send a multi-gigabyte body to the `/api/start` endpoint and exhaust memory.
- **Files:** `internal/server/server.go:162-176`, all other `handleXxx` functions
- **Fix approach:** Wrap `r.Body` with `http.MaxBytesReader(w, r.Body, 1<<20)` before decoding.

### Temp Script Files Contain User Commands
- **Risk:** `runCommandAs` writes the technique command to a temp `.ps1` file (e.g. `lnj_<nanos>.ps1`). Deferred `os.Remove` calls clean up afterward, but if the process is killed or the deferred call panics, the script file and output files persist in `%TEMP%` indefinitely.
- **Files:** `internal/executor/executor.go:105-110`
- **Fix approach:** Use `os.CreateTemp` to ensure cleanup, or wrap cleanup in a separate goroutine with a watchdog.

---

## Technical Debt

### `server.go` Uses Package-Level Global State
- **Issue:** `eng`, `registry`, `users`, and `cfg` are package-level `var` declarations. This prevents testing server handlers in isolation and makes the server non-restartable within a process (e.g. cannot swap the registry without restarting).
- **Files:** `internal/server/server.go:31-35`
- **Fix approach:** Encapsulate in a `Server` struct and inject dependencies through it.

### `simlog` Package Uses a Global Logger (`current *Logger`)
- **Issue:** `simlog.Start()` replaces the global `current` logger. If two simulations somehow ran concurrently, or if the UI polls logs while `Start()` is reinitializing, there is a small race window covered only by `globalMu`, not the `Logger.mu`. The two locks (`globalMu` and `Logger.mu`) are locked independently within `write()`, creating a risk of subtle ordering issues.
- **Files:** `internal/simlog/simlog.go:47-48`, `internal/simlog/simlog.go:238-257`
- **Fix approach:** Pass the logger instance explicitly rather than using a global.

### Double `userstore.Load()` on Failure — Second Call Ignored
- **Issue:** If the first `userstore.Load()` returns an error, `server.go` calls `Load()` again and discards that error too with `users, _`. This means that if the file is corrupt, the server still starts with an empty store, silently losing profiles.
- **Files:** `internal/server/server.go:48-51`
- **Fix approach:** Remove the second `Load()` call. If the file is corrupt, either fail fast or construct an empty store explicitly.

### `CampaignStep.DelayAfter` and `CampaignStep.Optional` Are Defined but Never Used
- **Issue:** Both fields exist in the `CampaignStep` struct and are present in campaign YAML files, but `getTechniquesForCampaign()` only looks up `step.TechniqueID` — it never reads `DelayAfter` or `Optional`. Campaign-level delays are silently ignored; optional techniques cannot be skipped.
- **Files:** `internal/playbooks/types.go:37-41`, `internal/engine/engine.go:580-597`
- **Impact:** Campaign pacing is broken; all campaigns run techniques back-to-back regardless of their configured `delay_after`.
- **Fix approach:** Apply `step.DelayAfter` as a per-step sleep in `getTechniquesForCampaign`, and honor `step.Optional` when a technique is missing from the registry.

### `Technique.InputArgs` and `Technique.NistControls` Are Dead Fields
- **Issue:** Both fields are defined in `Technique`, parsed from YAML, and serialized to JSON for the UI — but neither is used by the engine or executor. Technique commands that reference `#{input_arg}` placeholders would need substitution logic.
- **Files:** `internal/playbooks/types.go:17-18`, `internal/executor/executor.go` (no substitution logic)
- **Fix approach:** Either implement argument substitution in `runInternal`, or remove the fields if they are not planned.

### `GetTechniquesByPhase` Returns Techniques in Non-Deterministic Map Iteration Order
- **Issue:** The registry stores techniques in a `map[string]*Technique`. `GetTechniquesByPhase` iterates the map, so the order techniques are executed in non-PoC mode changes between runs.
- **Files:** `internal/playbooks/loader.go:72-80`
- **Impact:** Reproducible test runs are not possible. Two runs of the same campaign config may produce differently-ordered results.
- **Fix approach:** Sort the returned slice by `Technique.ID` before returning.

### Bubble Sort Used Instead of `sort.Slice`
- **Issue:** `GetAllTactics()` in `loader.go` and tactic stat sorting in `reporter.go` both implement an O(n²) bubble sort manually instead of using `sort.Slice`.
- **Files:** `internal/playbooks/loader.go:95-100`, `internal/reporter/reporter.go:129-135`
- **Impact:** No functional impact at current scale (~60 techniques), but is a code quality signal.
- **Fix approach:** Replace with `sort.Slice`.

### `go.mod` Declares `go 1.26.1` — Non-Existent Version
- **Issue:** As of the analysis date, Go 1.26.1 does not exist. The module file likely contains a future or incorrect version string.
- **Files:** `go.mod:3`
- **Impact:** Toolchain enforcement may fail on developer machines with current Go releases.
- **Fix approach:** Update to the actual installed Go version (`go 1.22` or `go 1.23`).

### HTML Report Template Uses German UI Labels Mixed with English Code
- **Issue:** The HTML report template (`reporter.go`) contains German strings (`Gesamt`, `Erfolgreich`, `Fehlgeschlagen`, `Ausgeführte Techniken`, `Generiert`) while the Go codebase is otherwise English. The WhatIf mode label in `engine.go` and PoC step labels also mix German text.
- **Files:** `internal/reporter/reporter.go:248-292`, `internal/engine/engine.go:336`, `internal/engine/engine.go:374`
- **Impact:** Internationalization is inconsistent; makes the tool harder to hand off to non-German-speaking users.
- **Fix approach:** Standardize on one language for UI-facing strings.

---

## Performance Risks

### In-Memory Log Entry List Grows Without Bound
- **Risk:** `simlog.Logger.entries` and `engine.Status.Results` are slices that are appended to throughout a simulation and never truncated. In PoC multi-day mode (`runPoC`), Phase 2 re-runs the full campaign every day, so entries accumulate across all days in a single slice.
- **Files:** `internal/simlog/simlog.go:249`, `internal/engine/engine.go:460`
- **Impact:** A PoC run with 5 Phase 2 days × 10 techniques × verbose output per technique can accumulate thousands of entries in memory, all returned on every `/api/logs` poll.
- **Fix approach:** Cap the in-memory entry list (e.g. last 2000 entries) and only persist to file.

### `/api/logs` Copies the Full Entry Slice on Every Poll
- **Risk:** `GetEntries()` copies the entire slice under lock. The UI polls `/api/status` and `/api/logs` on a timer, so this copy operation runs repeatedly during long simulations.
- **Files:** `internal/simlog/simlog.go:214-225`
- **Fix approach:** Add an `offset` query parameter so the UI can fetch only new entries since its last poll.

### `runCommandAs` Shells Out to PowerShell Twice Per Technique
- **Risk:** For techniques run as another user, two PowerShell processes are spawned: one to write the inner script, and one outer launcher. On slow machines, this adds measurable latency, especially with `DelayBetweenTechniques = 0`.
- **Files:** `internal/executor/executor.go:98-191`

---

## Architectural Risks

### No HTTP Server Timeouts Configured
- **Risk:** `http.ListenAndServe` is called without wrapping in an `http.Server` with `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`. Long-running technique executions or misbehaving clients can hold connections open indefinitely.
- **Files:** `internal/server/server.go:88`
- **Fix approach:** Replace with `&http.Server{ReadHeaderTimeout: 10*time.Second, IdleTimeout: 60*time.Second, ...}`.

### Engine Is a Singleton — Only One Simulation at a Time
- **Risk:** The engine enforces single-simulation semantics via phase checks, which is intentional. However, if `Start()` is called immediately after `Stop()` before the goroutine exits, the `stopCh` channel (buffered 1) may not have drained, and `isStopped()` will immediately return true on the next run.
- **Files:** `internal/engine/engine.go:121-128`, `internal/engine/engine.go:173-178`
- **Fix approach:** Wait for the goroutine to complete (`sync.WaitGroup`) before resetting state on a new `Start()`.

### `pickUser()` Takes a Full Write Lock for Sequential Rotation
- **Risk:** `pickUser()` acquires `e.mu.Lock()` (write lock) to increment `rotationIndex`. This blocks all concurrent `GetStatus()` reads (which use `RLock`) for the duration.  This is called inside `runTechnique`, which also calls `e.mu.Lock()` afterward to append results — creating repeated short write-lock windows inside the goroutine.
- **Files:** `internal/engine/engine.go:214-230`
- **Fix approach:** Use a separate `sync/atomic` counter for `rotationIndex` to avoid taking the shared engine lock.

### Windows-Only Tool with No Platform Guard
- **Risk:** The tool uses `powershell.exe`, `cmd.exe`, `auditpol.exe`, Windows DPAPI, and the Windows Security event log. There is no build constraint (`//go:build windows`) in any file, so `go build` on Linux/macOS will compile successfully but produce a binary that fails at runtime.
- **Files:** All `.go` files, especially `internal/executor/executor.go`, `internal/userstore/userstore.go`, `internal/preparation/preparation.go`
- **Fix approach:** Add `//go:build windows` to files with Windows-specific calls, or use a build tag at the module level.

### `stopCh` Channel Buffer Is 1 — Multiple `Stop()` Calls Silently Drop
- **Risk:** `Stop()` sends to a buffered channel of size 1 using a non-blocking select. If `Stop()` is called multiple times quickly (e.g. double-click in UI), only the first send succeeds. This is safe but means the UI stop button may appear to have no effect on a second press.
- **Files:** `internal/engine/engine.go:173-178`
- **Impact:** Low; functionally correct because one stop signal is sufficient.

---

## TODO/FIXME Items

No explicit `TODO`, `FIXME`, `HACK`, or `XXX` comments were found in the Go source files. Issues are instead tracked as warning log messages:

- **Campaign not found:** `engine.go:583` — logs `"WARNING: Campaign not found"` but returns `nil` silently, causing zero attack techniques to run without any user-visible error in the UI.
- **Technique missing from campaign:** `engine.go:591` — logs `"WARNING: Technique %s not found"` but silently skips the step. The UI results will show fewer techniques than the campaign step count with no explanation.
- **User store load failure:** `server.go:50` — logs a warning and continues with an empty store, potentially losing persisted user profiles silently.
- **Log file creation failure:** `simlog.go:65` — prints a warning to stdout but continues without file logging; the UI log viewer will still work but no file is written.

---

## Missing Features / Gaps

### No SIEM-Side Verification
- The tool executes techniques and records whether the OS command succeeded, but has no mechanism to query the SIEM (e.g. Exabeam, Splunk) to confirm alerts were generated. Validation is entirely manual.
- **Impact:** The primary purpose (SIEM validation) requires a manual lookup step that could be automated via a SIEM API.

### Campaign Step `delay_after` Not Applied
- The `CampaignStep.DelayAfter` field (defined in `types.go` and populated from YAML) is never read during campaign execution. All campaign steps run back-to-back, ignoring pacing configuration.
- **Files:** `internal/engine/engine.go:580-597`, `internal/playbooks/types.go:38`

### No Technique Input Argument Substitution
- `Technique.InputArgs` is parsed from YAML but never applied to `Executor.Command`. Techniques that use `#{arg}` placeholders in their commands will run with the literal placeholder string rather than a substituted value.
- **Files:** `internal/playbooks/types.go:17`, `internal/executor/executor.go:54-93`

### No Log Rotation or Cleanup for Output Files
- Each simulation run creates a `.log`, `.json`, and `.html` file in the working directory. There is no automatic cleanup of old files. Long-term use will accumulate files indefinitely.
- **Files:** `internal/simlog/simlog.go:56-63`, `internal/reporter/reporter.go:52-71`

### User Profile ID Collision Not Detected
- Profile IDs are derived from `domain_username` (lowercased, spaces replaced with `_`). Adding two users with the same username in the same domain silently overwrites the first profile.
- **Files:** `internal/userstore/userstore.go:112-116`

### No Tests
- The codebase has zero test files (`*_test.go`). There is no coverage for any package: engine, executor, userstore, simlog, reporter, or playbook loading.
- **Impact:** Regressions in core simulation logic, credential handling, or campaign loading cannot be detected automatically.

### PoC Mode Phase 2 Re-Runs Full Campaign Every Day
- In PoC mode, Phase 2 calls `getTechniquesForPhase()` fresh on each day loop iteration, re-running the full campaign every day without variation. If the intent is different technique subsets per day, the current design does not support it.
- **Files:** `internal/engine/engine.go:404`

---

## Fragile Areas

### `runCommandAs` — PowerShell String Escaping
- **Files:** `internal/executor/executor.go:128-166`
- **Why fragile:** The launcher script is built via `fmt.Sprintf` with manual backtick-escaping. Only three characters are escaped. Passwords or paths containing characters like `'`, newlines, null bytes, or Unicode can break the generated PowerShell, causing silent failures or unexpected behavior.
- **Safe modification:** Any change to the escaping logic requires thorough testing with edge-case passwords (special chars, very long strings, empty).
- **Test coverage:** None.

### `simlog` Dual-Lock Pattern
- **Files:** `internal/simlog/simlog.go:237-257`
- **Why fragile:** `write()` must be called with `globalMu` held, but then immediately acquires `l.mu` (a different lock on the Logger struct). The locking contract is enforced by convention only (comments). A future contributor adding a public logging function that calls `write()` without `globalMu` will introduce a data race.
- **Safe modification:** Add a `//go:noescape` note or redesign so `write` does not require the caller to hold an external lock.

### `userstore.Save()` Called Inside `mu.Lock()` Scope
- **Files:** `internal/userstore/userstore.go:146-149`
- **Why fragile:** `Add()` holds `s.mu.Lock()` while calling `s.save()`, which does I/O. If the disk is slow or full, this blocks all other store operations (including reads) for the duration of the write.
- **Safe modification:** Release the lock before writing, or write to a temp file and rename atomically.

### `handleTechniques` Returns Unsorted, Non-Deterministic List
- **Files:** `internal/server/server.go:125-131`
- **Why fragile:** Iterates the registry map directly; order changes between requests. UI rendering depends on JavaScript sort, but any consumer expecting stable ordering will be surprised.

### Embedded Binary Includes All Techniques at Compile Time
- **Files:** `internal/playbooks/loader.go:12-13`
- **Why fragile:** Adding a new technique YAML requires a recompile. There is no mechanism to load external techniques at runtime, meaning field teams cannot extend the technique library without access to Go toolchain.

---

*Concerns audit: 2026-03-24*
