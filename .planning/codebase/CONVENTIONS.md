# Code Conventions

**Analysis Date:** 2026-03-24

## Naming Conventions

**Packages:**
- All lowercase, single word matching the directory name: `engine`, `executor`, `playbooks`, `simlog`, `userstore`, `preparation`, `reporter`, `server`
- Package-level doc comments use `// Package <name> ...` format on the first line (seen in `simlog` and `userstore`)

**Types / Structs:**
- PascalCase for all exported types: `Technique`, `Campaign`, `ExecutionResult`, `UserProfile`, `Registry`, `Config`, `Status`, `Entry`
- Unexported types are also PascalCase when used as internal helpers: `resolvedProfile`, `tacticStat`, `htmlData`

**Constants:**
- Exported typed constants use PascalCase with a type prefix: `PhaseIdle`, `PhaseDiscovery`, `PhaseDone`, `RotationNone`, `RotationSequential`, `TypeSimStart`, `TypeTechEnd`
- Untyped string constants use `const name = value` directly: `banner` in `main.go`

**Variables:**
- Unexported package-level vars: camelCase — `current`, `globalMu`, `defaultFilePath`, `eng`, `registry`, `users`, `cfg`
- Local variables: camelCase — `encPassword`, `qualifiedUser`, `successRate`, `tacticMap`
- Short single-letter names for loop indices and common idioms: `i`, `j`, `p`, `t`, `r`, `d`

**Functions:**
- Exported functions: PascalCase — `LoadEmbedded`, `SaveResults`, `RunWithCleanup`, `TestCredentials`
- Unexported functions: camelCase — `runInternal`, `runCommandAs`, `runCommand`, `runPS`, `saveHTML`, `fileExists`
- HTTP handler functions: `handle` prefix + noun — `handleStatus`, `handleStart`, `handleUsers`, `handleUsersDiscover`
- Receiver methods: short (1-2 letter) variable matching type initial — `(e *Engine)`, `(s *Store)`, `(r *Registry)`, `(l *Logger)`, `(u *UserProfile)`

**Files:**
- All lowercase with underscores where needed: `engine.go`, `executor.go`, `userstore.go`, `simlog.go` — no underscores except in test files
- One file per package in most packages; multi-file split only in `playbooks` (`types.go` + `loader.go`)

**Struct Tags:**
- Both `yaml` and `json` tags aligned with spaces on the same field, columns visually aligned:
  ```go
  ID          string `yaml:"id"          json:"id"`
  Name        string `yaml:"name"        json:"name"`
  ```
- `json:"field,omitempty"` used selectively for optional fields: `InputArgs`, `NistControls`, `ReportFile`, `PoCDay`

## Code Style

**Formatting:**
- Standard `gofmt` formatting assumed throughout — consistent indentation, spacing
- No linter config files present (no `.golangci.yml`, `.revive.toml`, etc.)

**Imports:**
- Standard library imports first, then a blank line, then internal/third-party imports:
  ```go
  import (
      "fmt"
      "sync"
      "time"

      "lognojutsu/internal/executor"
      "lognojutsu/internal/playbooks"
  )
  ```
- No aliased imports; all packages imported by their declared package name

**Constants grouped by type:**
- Related constants declared in a single `const (...)` block with the typed constant group at top:
  ```go
  const (
      PhaseIdle    Phase = "idle"
      PhaseDiscovery Phase = "discovery"
      // ...
  )
  ```

**Section separators in large files:**
- Unicode box-drawing comment banners used to visually divide sections in `server.go` and `executor.go`:
  ```go
  // ── Simulation ────────────────────────────────────────────────────────────────
  // ── internal ──────────────────────────────────────────────────────────────────
  ```

**Inline anonymous structs for one-off request decoding:**
- Used throughout `server.go` to decode JSON request bodies without defining named types:
  ```go
  var req struct {
      Step string `json:"step"`
  }
  ```

**Slice initialization:**
- `make([]T, 0, cap)` used when capacity is known: `make([]Entry, len(current.entries))`
- `var result []T` + `append` used when capacity unknown
- Explicit nil-to-empty-slice normalisation before JSON responses: `if entries == nil { entries = []simlog.Entry{} }`

## Error Handling

**Pattern: always check error, wrap with context using `%w`:**
```go
data, readErr := embeddedFS.ReadFile(path)
if readErr != nil {
    return fmt.Errorf("reading %s: %w", path, readErr)
}
```

**Error variable naming:**
- The first error in a function is `err`; subsequent errors in the same scope get descriptive names: `readErr`, `parseErr`, `cleanErr`, `runErr`
- This avoids shadowing and makes error context obvious at call sites

**Fatal errors at startup only:**
- `log.Fatalf` used exclusively in `main.go` for unrecoverable startup failures:
  ```go
  if err := server.Start(cfg); err != nil {
      log.Fatalf("Server error: %v", err)
  }
  ```

**HTTP handler error pattern:**
- Method validation first, then decode, then business logic, each returning early via `writeError`:
  ```go
  if r.Method != http.MethodPost {
      writeError(w, "POST required", http.StatusMethodNotAllowed)
      return
  }
  if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
      writeError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
      return
  }
  ```

**Degraded operation on non-critical failure:**
- `server.go` warns and continues if user store fails to load (non-fatal):
  ```go
  if err != nil {
      log.Printf("WARNING: Could not load user store: %v (starting with empty store)", err)
      users, _ = userstore.Load()
  }
  ```
- DPAPI encrypt failure falls back to `"PLAIN:"` prefix rather than failing outright

**Ignored errors marked with `_`:**
- Only for cleanup operations or best-effort file writes:
  ```go
  defer os.Remove(scriptFile)
  _, _ = l.file.WriteString(line)
  _ = s.save()
  ```

**Error strings:**
- Lowercase, no trailing period, describe the operation: `"reading user store"`, `"decrypting password for %s"`
- HTTP error strings are capitalised (user-visible): `"POST required"`, `"Profile not found"`

## Comments and Documentation

**Package-level doc comments:**
- Present on packages with non-obvious purpose: `simlog`, `userstore`
- Format: `// Package <name> <verb phrase>.`

**Exported type and function comments:**
- All exported types and most exported functions have a single-line doc comment immediately above the declaration:
  ```go
  // Registry holds all loaded techniques and campaigns.
  type Registry struct { ... }

  // LoadEmbedded loads all playbooks bundled into the binary.
  func LoadEmbedded() (*Registry, error) { ... }
  ```

**Inline comments:**
- Used for non-obvious decisions, security rationale, and Windows-specific behaviour:
  ```go
  // Event 4648 is generated here — "A logon was attempted using explicit credentials"
  // 0600 = owner-only read/write
  ```

**Section header comments in server.go:**
- `// ── Section Name ──` separators group handler registrations and functions visually

**TODO/WARNING comments:**
- `WARNING:` prefix used in log messages for operational warnings, not code comments
- No `TODO`/`FIXME` markers found in any source file

**German strings in UI-facing output:**
- `WhatIf` mode labels, PoC phase status strings, and the HTML report are in German
- Go source code comments and log messages are in English throughout

## Patterns Used

**Typed string constants for enumerations:**
- All domain enumerations use `type X string` + `const` block rather than `iota` integers:
  ```go
  type Phase string
  const (
      PhaseIdle    Phase = "idle"
      PhaseDiscovery Phase = "discovery"
  )
  ```
  This makes JSON serialisation trivial and values are human-readable in logs.

**sync.RWMutex for concurrent state:**
- `Engine` and `Store` both embed `sync.RWMutex` as the first field
- Read operations use `e.mu.RLock()` / `defer e.mu.RUnlock()`
- Write operations use `e.mu.Lock()` / `e.mu.Unlock()` (defer avoided when unlock must happen before subsequent reads in the same function)

**Constructor functions named `New`:**
- `engine.New(registry, users)` — standard Go constructor convention

**Global singleton with mutex for logging:**
- `simlog` uses a package-level `current *Logger` plus `globalMu sync.Mutex` for the active session
- All public functions guard with `if current == nil { return }` before accessing

**`embed.FS` for bundled assets:**
- `//go:embed embedded` in `playbooks/loader.go` bundles YAML playbooks
- `//go:embed static` in `server/server.go` bundles the HTML UI
- Both walked with `fs.WalkDir` or `fs.Sub`

**`defer` for cleanup:**
- `defer f.Close()` for file handles
- `defer os.Remove(...)` for temp files in `executor.go`
- `defer e.mu.RUnlock()` / `defer s.mu.RUnlock()` for read locks

**Stop-channel pattern for goroutine control:**
- `stopCh chan struct{}` with buffered size 1; `Stop()` uses non-blocking send:
  ```go
  select {
  case e.stopCh <- struct{}{}:
  default:
  }
  ```
- `isStopped()` uses non-blocking receive on the same channel

**`select` on timer + stop channel:**
```go
func (e *Engine) waitOrStop(d time.Duration) bool {
    select {
    case <-time.After(d):
        return true
    case <-e.stopCh:
        return false
    }
}
```

**JSON response helpers (`writeJSON` / `writeError`):**
- All HTTP handlers use two shared helpers rather than inline `json.NewEncoder` calls
- `writeError` always sets status code before writing body

**Bubble-up error wrapping through all layers:**
- `loader.go` → `server.go` → `main.go` — errors wrapped at each boundary with `fmt.Errorf("...: %w", err)` so the full context chain is preserved in `log.Fatal` output
