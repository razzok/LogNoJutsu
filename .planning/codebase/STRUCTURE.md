# Codebase Structure

**Analysis Date:** 2026-03-24

## Directory Layout

```
LogNoJutsu/
├── cmd/
│   └── lognojutsu/
│       └── main.go              # CLI entry point — flag parsing and server.Start()
├── internal/
│   ├── engine/
│   │   └── engine.go            # Simulation engine: phase state machine, scheduling
│   ├── executor/
│   │   └── executor.go          # OS command execution: powershell/cmd, RunAs via DPAPI
│   ├── playbooks/
│   │   ├── types.go             # Technique, Campaign, ExecutionResult struct definitions
│   │   ├── loader.go            # Embedded YAML loader, Registry, query helpers
│   │   └── embedded/
│   │       ├── techniques/      # One YAML file per ATT&CK technique (~50 files)
│   │       └── campaigns/       # One YAML file per attack scenario (~12 files)
│   ├── preparation/
│   │   └── preparation.go       # Host setup: PowerShell logging, auditpol, Sysmon
│   ├── reporter/
│   │   └── reporter.go          # JSON + HTML report generation (template compiled-in)
│   ├── server/
│   │   ├── server.go            # HTTP server, all route handlers, auth middleware
│   │   └── static/
│   │       └── index.html       # Single-page UI (vanilla JS, embedded into binary)
│   ├── simlog/
│   │   └── simlog.go            # Dual-destination structured logger (memory + .log file)
│   └── userstore/
│       └── userstore.go         # Credential store: DPAPI encryption, JSON persistence
├── go.mod                       # Module: "lognojutsu", requires gopkg.in/yaml.v3
├── go.sum
├── lognojutsu.exe               # Compiled Windows binary (not committed to source)
└── README.md
```

## Entry Points

**Binary entry point:**
- `cmd/lognojutsu/main.go` — `func main()`: parses `-host`, `-port`, `-password` flags; calls `server.Start(cfg)`

**HTTP API entry points (all in `internal/server/server.go`):**

| Method | Path | Handler |
|---|---|---|
| GET | `/` | Static UI (`index.html` via `embed.FS`) |
| GET | `/api/status` | `handleStatus` — returns engine `Status` JSON |
| GET | `/api/techniques` | `handleTechniques` — all techniques from registry |
| GET | `/api/campaigns` | `handleCampaigns` — all campaigns from registry |
| GET | `/api/tactics` | `handleTactics` — deduplicated tactic name list |
| POST | `/api/start` | `handleStart` — decodes `engine.Config`, calls `eng.Start()` |
| POST | `/api/stop` | `handleStop` — calls `eng.Stop()` |
| GET | `/api/logs` | `handleLogs` — returns in-memory `simlog.Entry` slice |
| GET | `/api/report` | `handleReport` — serves the latest HTML report file |
| POST | `/api/prepare` | `handlePrepare` — runs all preparation steps |
| POST | `/api/prepare/step` | `handlePrepareStep` — runs a named step (`powershell`, `auditpol`, `sysmon`) |
| GET/POST | `/api/users` | `handleUsers` — list or add user profiles |
| POST | `/api/users/delete` | `handleUsersDelete` |
| POST | `/api/users/discover` | `handleUsersDiscover` — scan local + domain users |
| POST | `/api/users/test` | `handleUsersTest` — validate credentials |

**Simulation execution entry point:**
- `internal/engine/engine.go` → `Engine.Start(cfg Config)` starts a goroutine that runs `e.run()` (normal) or `e.runPoC()` (PoC mode)

## Module Organization

Each subdirectory of `internal/` is a focused, single-responsibility package. No package has circular imports. Dependency flow is strictly one-directional:

```
server
  ├── engine
  │     ├── executor
  │     │     ├── playbooks
  │     │     ├── simlog
  │     │     └── userstore
  │     ├── playbooks
  │     ├── reporter
  │     │     └── playbooks
  │     ├── simlog
  │     └── userstore
  ├── playbooks
  ├── preparation     (no internal deps)
  ├── simlog          (no internal deps)
  └── userstore       (no internal deps)
```

Each package exposes a small, stable public surface:
- `engine`: `Engine`, `New()`, `Config`, `Status`, `Phase`, `UserRotation` constants
- `playbooks`: `Registry`, `Technique`, `Campaign`, `CampaignStep`, `ExecutionResult`, `Executor`, `LoadEmbedded()`
- `executor`: `Run()`, `RunAs()`, `RunWithCleanup()`, `RunCleanupOnly()`
- `simlog`: package-level functions only (`Start`, `Stop`, `Phase`, `TechStart`, etc., `GetEntries`, `GetFilePath`)
- `userstore`: `Store`, `UserProfile`, `UserType`, `DiscoveredUser`, `Load()`, `DiscoverLocalUsers()`, `DiscoverRecentDomainUsers()`, `TestCredentials()`
- `preparation`: `Result`, `RunAll()`, `EnablePowerShellLogging()`, `ConfigureAuditPolicy()`, `InstallSysmon()`
- `reporter`: `Report`, `SaveResults()`

## File Naming Conventions

**Go source files:**
- One file per package, named after the package: `engine.go`, `executor.go`, `loader.go`, `types.go`, `reporter.go`, `simlog.go`, `userstore.go`, `preparation.go`, `server.go`
- `types.go` is the exception — it holds only struct definitions and is separated from logic in `loader.go` within the `playbooks` package
- No test files currently exist

**Playbook YAML files — techniques:**
- Pattern: `T{MITRE_ID}_{descriptive_slug}.yaml`
- Examples: `T1059_001_powershell.yaml`, `T1003_001_lsass.yaml`, `T1110_003_password_spraying.yaml`
- UEBA-specific chains use a different prefix: `UEBA_{slug}.yaml` (e.g., `UEBA_credential_spray_chain.yaml`)

**Playbook YAML files — campaigns:**
- Pattern: `{descriptive_slug}.yaml`, usually `{industry_or_actor}_{scenario}.yaml`
- Examples: `ransomware_full_chain.yaml`, `finance_fin7.yaml`, `ueba_exabeam_validation.yaml`

**Output files (written to working directory at runtime):**
- Log files: `lognojutsu_<YYYYMMDD_HHMMSS>[_<campaign_id>].log`
- JSON reports: `lognojutsu_report_<YYYYMMDD_HHMMSS>.json`
- HTML reports: `lognojutsu_report_<YYYYMMDD_HHMMSS>.html`
- User store: `lognojutsu_users.json`

## Playbook YAML Schema

**Technique file fields** (defined in `internal/playbooks/types.go`):
```yaml
id: T1059.001                    # unique ID used as registry key
name: PowerShell Execution
description: ...
tactic: execution                # MITRE tactic name (lowercase, hyphenated)
technique_id: T1059.001          # MITRE ATT&CK ID
platform: windows
phase: attack                    # "discovery" or "attack" — controls which engine phase runs it
elevation_required: false
expected_events:
  - "4104 (PowerShell ScriptBlock ...)"
tags:
  - execution
executor:
  type: powershell               # "powershell", "psh", "cmd", or "command_prompt"
  command: |
    Write-Host "..."
cleanup: ""                      # optional cleanup command (same executor type assumed)
input_args: {}                   # optional variable substitutions
nist_controls: []                # optional NIST 800-53 control references
```

**Campaign file fields:**
```yaml
id: ransomware-full-chain        # unique ID used as registry key
name: Ransomware Full Attack Chain
description: ...
industry: Cross-Industry
threat_actor: Ransomware Group (LockBit / Conti TTPs)
tags:
  - ransomware
steps:
  - technique_id: T1082          # references a Technique ID
    delay_after: 5               # seconds to wait after this step
    optional: false
```

## Where to Add New Code

**New ATT&CK technique:**
- Create `internal/playbooks/embedded/techniques/T{ID}_{slug}.yaml`
- Set `phase: discovery` or `phase: attack` to control which engine phase runs it
- No Go code changes required — the `//go:embed` directive picks up all `.yaml` files at build time

**New campaign (attack scenario):**
- Create `internal/playbooks/embedded/campaigns/{slug}.yaml`
- Reference existing technique IDs in `steps[].technique_id`
- No Go code changes required

**New API endpoint:**
- Add a handler function `handleXxx(w http.ResponseWriter, r *http.Request)` in `internal/server/server.go`
- Register it in `server.Start()` with `mux.HandleFunc("/api/...", authMiddleware(handleXxx))`

**New preparation step:**
- Add a function returning `preparation.Result` in `internal/preparation/preparation.go`
- Add it to the `steps` slice in `RunAll()`
- Add a `case` for it in `server.handlePrepareStep()`

**New engine config option:**
- Add field to `engine.Config` struct in `internal/engine/engine.go`
- Add corresponding field to `engine.Status` if it needs to be visible via `GET /api/status`
- Handle it in `engine.run()` or `engine.runPoC()`

**New log entry type:**
- Add an `EntryType` constant to `internal/simlog/simlog.go`
- Add a corresponding exported function (following the pattern of `TechStart`, `TechEnd`, etc.)

## Special Directories and Files

**`internal/playbooks/embedded/`:**
- Compiled into the binary via `//go:embed embedded` in `internal/playbooks/loader.go`
- Contains only YAML data files — no Go code
- Generated: No. Committed: Yes

**`internal/server/static/`:**
- Compiled into the binary via `//go:embed static` in `internal/server/server.go`
- Contains only `index.html` — a self-contained single-page application with all CSS and JS inline
- Generated: No. Committed: Yes

**`lognojutsu_users.json`** (runtime, not in repo):
- Created in the working directory by `userstore` on first user add
- File permissions set to `0600` (owner read/write only)
- Contains DPAPI-encrypted password blobs — machine- and user-account-specific

**`.planning/codebase/`:**
- GSD planning documents. Not part of the build.

---

*Structure analysis: 2026-03-24*
