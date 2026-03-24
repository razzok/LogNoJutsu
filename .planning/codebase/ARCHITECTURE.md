# Architecture

**Analysis Date:** 2026-03-24

## Overview

LogNoJutsu is a Windows-only SIEM validation and ATT&CK simulation tool. It runs as a single self-contained binary that embeds all playbook data. A local HTTP server exposes a JSON API consumed by a bundled single-page UI. The operator configures and triggers simulations through the UI; the engine executes MITRE ATT&CK technique definitions by shelling out to `powershell.exe` or `cmd.exe` on the host.

The overall shape is a **monolithic single-process server** with a clear internal layering:

```
UI (browser)  ──HTTP──▶  server  ──▶  engine  ──▶  executor  ──▶  OS (powershell/cmd)
                             │
                             ├──▶  playbooks (embedded YAML registry)
                             ├──▶  userstore  (credential store, DPAPI)
                             ├──▶  preparation (audit-policy / Sysmon setup)
                             ├──▶  simlog     (dual file+memory log)
                             └──▶  reporter   (JSON + HTML report generation)
```

## Design Patterns

**Embedded assets via `//go:embed`**
All playbook YAML files and the static UI are compiled into the binary at build time. `internal/playbooks/loader.go` embeds `embedded/` and `internal/server/server.go` embeds `static/`. The deployed artifact is a single `.exe` with no external data dependencies.

**Registry pattern for playbooks**
`playbooks.Registry` holds two maps (`map[string]*Technique`, `map[string]*Campaign`) keyed by ID. The registry is loaded once at startup and treated as read-only thereafter — no mutation after `LoadEmbedded()` returns.

**Engine as a state machine**
`engine.Engine` tracks a `Phase` enum that transitions through a defined set of states. Phases are typed string constants:
```
idle → discovery → attack → done
idle → poc_phase1 → poc_gap → poc_phase2 → done
(any running phase) → aborted
```
The engine holds a `stopCh chan struct{}` for cooperative cancellation. The goroutine started by `Engine.Start()` checks `isStopped()` between each technique.

**Mutex-guarded mutable state**
`engine.Engine` uses a `sync.RWMutex` (`mu`) to protect the `Status` struct. Reads use `RLock`; writes (phase transitions, result appends) use `Lock`. `simlog` uses a two-level lock: a global `globalMu` for session lifecycle and a per-logger `mu` for entry appends.

**Optional WhatIf (dry-run) mode**
When `Config.WhatIf` is true the engine skips `executor.RunAs`/`RunWithCleanup` entirely and records a synthetic `ExecutionResult` with a placeholder output. The rest of the pipeline (logging, reporting) operates unchanged.

**User rotation strategies**
The engine supports three strategies, selected at start time and applied per-technique via `pickUser()`:
- `none` — all techniques run as the process owner
- `sequential` — profiles are cycled in order with an index counter
- `random` — `rand.Intn` selects a profile each call

**PoC multi-day scheduling**
When `PoCMode` is true the engine switches to `runPoC()`, which schedules Phase 1 (discovery) and Phase 2 (attack) across multiple calendar days at configured hours. `nextOccurrenceOfHour(hour int)` computes the duration until the next occurrence of a given hour and the engine sleeps via `waitOrStop(d)`.

**Password security via Windows DPAPI**
`userstore` encrypts stored passwords with Windows DPAPI (`ConvertTo-SecureString` / `ConvertFrom-SecureString` via PowerShell). The encrypted blob is tied to the machine and Windows user account. The `PasswordEnc` field is stripped from all API responses before serialisation; the plaintext is decrypted only at simulation start and held only in memory for the duration of the run.

**Cleanup tracking**
When `RunCleanup` is false, techniques with non-empty `Cleanup` commands are accumulated in `engine.executedTechniques`. On `finish()` or `abort()`, `runPendingCleanups()` iterates that list and calls `executor.RunCleanupOnly`. When `RunCleanup` is true, cleanup runs immediately after each technique inside `executor.RunWithCleanup`.

## Key Components

**`cmd/lognojutsu/main.go`**
Entry point. Parses CLI flags (`-host`, `-port`, `-password`), prints banner, constructs `server.Config`, and calls `server.Start()`. Contains no logic beyond configuration wiring.

**`internal/server/server.go`**
HTTP layer. Owns global instances of `*engine.Engine`, `*playbooks.Registry`, `*userstore.Store`. Registers all routes on an `http.ServeMux`. Implements `authMiddleware` (optional HTTP Basic Auth against a single password). Handles serialisation errors and method validation. Delegates all business logic to the packages below.

API surface:
- `GET  /api/status` — engine phase and results
- `GET  /api/techniques`, `GET /api/campaigns`, `GET /api/tactics`
- `POST /api/start` — starts a simulation with a JSON `engine.Config` body
- `POST /api/stop`
- `GET  /api/logs` — in-memory simlog entries
- `GET  /api/report` — serves the latest HTML report file
- `POST /api/prepare`, `POST /api/prepare/step` — host preparation
- `GET/POST /api/users`, `POST /api/users/discover`, `POST /api/users/test`, `POST /api/users/delete`

**`internal/engine/engine.go`**
Simulation orchestrator. Resolves user profiles, selects techniques for each phase, invokes `executor`, collects `ExecutionResult` values, and transitions phases. Starts the simulation in a goroutine. Exposes thread-safe `GetStatus()` and `Stop()`. Also contains `runPoC()` for multi-day mode.

**`internal/playbooks/types.go`**
Data model definitions: `Technique`, `Executor`, `Campaign`, `CampaignStep`, `ExecutionResult`. These structs have both `yaml` and `json` struct tags — they are both parsed from YAML files and serialised to JSON for the API.

**`internal/playbooks/loader.go`**
Loads and parses the embedded YAML files using `fs.WalkDir` over the embedded filesystem. Populates `Registry.Techniques` and `Registry.Campaigns` maps. Provides `GetTechniquesByPhase()` and `GetAllTactics()` query helpers.

**`internal/executor/executor.go`**
Executes technique commands on the Windows host. Two code paths:
- `runCommand()` — runs as the current process user via `exec.Command("powershell.exe", ...)` or `exec.Command("cmd.exe", ...)`
- `runCommandAs()` — impersonates another user by writing the command to a temp `.ps1` file, then launching it via `Start-Process -Credential` from a PowerShell launcher script. This path intentionally generates Windows Event 4648 (explicit credential use) as a UEBA signal.

**`internal/simlog/simlog.go`**
Structured, dual-destination logger. Maintains a global `*Logger` instance per simulation session. Every write goes to both an in-memory `[]Entry` slice (for `GET /api/logs`) and a `.log` file on disk. Entry types are typed string constants (`TypeTechStart`, `TypeOutput`, `TypePhase`, etc.) that categorise log lines.

**`internal/reporter/reporter.go`**
Post-simulation report generator. Produces two output files:
- `lognojutsu_report_<stamp>.json` — full results as JSON
- `lognojutsu_report_<stamp>.html` — self-contained HTML report with tactic summary statistics and per-technique result rows, using `text/template` with a compiled-in HTML template string.

**`internal/preparation/preparation.go`**
Pre-simulation host setup. Three steps (callable individually via `POST /api/prepare/step` or all at once via `POST /api/prepare`): `EnablePowerShellLogging` (registry keys for EIDs 4103/4104), `ConfigureAuditPolicy` (`auditpol.exe` calls), `InstallSysmon` (download + install via PowerShell).

**`internal/userstore/userstore.go`**
Credential store backed by `lognojutsu_users.json`. Thread-safe via `sync.RWMutex`. Passwords encrypted/decrypted via DPAPI through PowerShell. Also contains `DiscoverLocalUsers()` (`Get-LocalUser`) and `DiscoverRecentDomainUsers()` (Security EventLog EID 4624 query) for UI-assisted profile creation. `TestCredentials()` validates a stored profile using `System.DirectoryServices.AccountManagement.PrincipalContext`.

## Data Flow

**Simulation start (normal mode):**

1. Browser POSTs `engine.Config` JSON to `POST /api/start`
2. `server.handleStart` decodes body, calls `eng.Start(cfg)`
3. `engine.Start()` validates state, resolves user profiles (decrypts passwords), calls `simlog.Start()`, spawns goroutine `go e.run()`
4. `e.run()` waits `DelayBeforeDiscovery`, enters Discovery phase:
   - Calls `registry.GetTechniquesByPhase("discovery")`, filters by tactic inclusion/exclusion list
   - For each technique: calls `pickUser()` → calls `executor.RunAs(t, profile, password)` or `executor.RunWithCleanup()`
   - `executor.runInternal()` shells out to `powershell.exe`/`cmd.exe`, captures stdout/stderr, returns `ExecutionResult`
   - Result is appended to `engine.status.Results` under lock
   - `simlog` writes to both memory and file
5. After `DelayBeforeAttack`, enters Attack phase; same loop for attack-phase or campaign techniques
6. On completion: `e.finish()` runs pending cleanups, calls `reporter.SaveResults()` which writes JSON + HTML files to the working directory, stores the HTML path in `status.ReportFile`
7. Phase transitions to `done`
8. Browser polls `GET /api/status` and `GET /api/logs` to update the UI live

**RunAs impersonation flow:**

1. Engine calls `executor.RunAs(t, profile, password)` with a resolved `UserProfile` and decrypted plaintext password
2. `runCommandAs()` writes the technique command to a temp `.ps1` file
3. Constructs a PowerShell launcher script that creates a `PSCredential` and calls `System.Diagnostics.Process.Start()` with `UseShellExecute=false`, `RedirectStandardOutput=true`, capturing output to temp files
4. The launcher is executed as the current process user; the inner script runs as the target user
5. Windows generates Event 4648 at this step (deliberate UEBA signal)
6. Temp files are read back for stdout/stderr, then deleted

**Report flow:**

1. `reporter.SaveResults()` collects all `ExecutionResult` values from the completed run
2. Writes `lognojutsu_report_<stamp>.json`
3. Builds per-tactic statistics (`tacticStat` aggregation)
4. Renders the compiled-in `htmlTemplate` via `text/template` to `lognojutsu_report_<stamp>.html`
5. Returns the HTML file path, which `engine.finish()` stores in `status.ReportFile`
6. `GET /api/report` reads and serves the file with `Content-Type: text/html`

## Module Responsibilities

| Package | Responsibility | Depends On |
|---|---|---|
| `cmd/lognojutsu` | CLI entry point, flag parsing | `internal/server` |
| `internal/server` | HTTP routing, auth middleware, request/response handling | `engine`, `playbooks`, `preparation`, `simlog`, `userstore` |
| `internal/engine` | Simulation lifecycle, phase state machine, user rotation, PoC scheduling | `executor`, `playbooks`, `reporter`, `simlog`, `userstore` |
| `internal/playbooks` | YAML type definitions, embedded registry, query helpers | `gopkg.in/yaml.v3` |
| `internal/executor` | OS command execution, RunAs impersonation via PowerShell | `playbooks`, `simlog`, `userstore` |
| `internal/simlog` | Dual-destination structured logger (memory + file) | stdlib only |
| `internal/reporter` | JSON + HTML report generation | `playbooks` |
| `internal/preparation` | Host pre-configuration (PowerShell logging, auditpol, Sysmon) | stdlib only |
| `internal/userstore` | Credential store (DPAPI encryption, JSON persistence, user discovery) | stdlib only |

---

*Architecture analysis: 2026-03-24*
