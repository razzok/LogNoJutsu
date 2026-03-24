# Tech Stack

**Analysis Date:** 2026-03-24

## Languages

**Primary:**
- Go 1.26.1 — all application logic, HTTP server, process execution, reporting

**Templating/Scripting (embedded, not standalone):**
- PowerShell — invoked at runtime via `exec.Command("powershell.exe")` for technique execution, DPAPI encryption/decryption, credential validation, and user discovery
- CMD (cmd.exe) — invoked at runtime for certain technique executors of type `cmd` / `command_prompt`
- YAML — playbook definitions (techniques and campaigns) embedded in the binary

**Frontend:**
- HTML/CSS/JS — single-file UI at `internal/server/static/index.html`, served via Go's `embed.FS`

## Runtime

**Environment:**
- Windows only (hard requirement — depends on `powershell.exe`, `cmd.exe`, `auditpol.exe`, `sc.exe`, Windows DPAPI, Windows Event Log)
- No Linux/macOS support; executor and userstore packages make direct Windows API calls

**Go toolchain:**
- Go 1.26.1 (declared in `go.mod`)
- No `.nvmrc`, `.python-version`, or other runtime version pins needed

## Package Manager

- Go modules (`go.mod` / `go.sum`)
- Lockfile: `go.sum` present
- No `vendor/` directory — dependencies fetched from module proxy

## Frameworks

**HTTP Server:**
- Go standard library `net/http` — no external web framework
- Routes registered manually via `http.NewServeMux()` in `internal/server/server.go`

**Templating:**
- Go standard library `text/template` — used in `internal/reporter/reporter.go` to generate HTML reports

**Embedding:**
- Go standard library `embed` — used in two places:
  - `internal/server/server.go`: `//go:embed static` embeds the UI
  - `internal/playbooks/loader.go`: `//go:embed embedded` embeds all YAML playbooks

**YAML parsing:**
- `gopkg.in/yaml.v3 v3.0.1` — the only external dependency; used in `internal/playbooks/loader.go` to deserialise technique and campaign YAML files

**Testing:**
- None configured — no test files, no test framework, no test runner config

## Build & Package Management

**Build:**
- Standard `go build` — no Makefile, no build scripts, no CI config detected
- Output: `lognojutsu.exe` (pre-built binary committed to repo root)
- Binary is self-contained: all YAML playbooks and the UI HTML file are embedded at compile time

**Run flags (declared in `cmd/lognojutsu/main.go`):**
```
-host     string   Bind address (default "127.0.0.1")
-port     int      HTTP port (default 8080)
-password string   Optional Basic Auth password
```

## Key Dependencies

**External (go.mod):**
- `gopkg.in/yaml.v3 v3.0.1` — YAML deserialisation for technique/campaign playbooks (`internal/playbooks/loader.go`)

**Standard library — critical packages:**
- `embed` — compile-time embedding of static assets and playbooks
- `net/http` — HTTP server and routing (`internal/server/server.go`)
- `os/exec` — spawning `powershell.exe` and `cmd.exe` processes (`internal/executor/executor.go`, `internal/userstore/userstore.go`, `internal/preparation/preparation.go`)
- `text/template` — HTML report generation (`internal/reporter/reporter.go`)
- `encoding/json` — REST API serialisation throughout `internal/server/server.go`
- `sync` — `sync.RWMutex` used in `internal/engine/engine.go`, `internal/simlog/simlog.go`, `internal/userstore/userstore.go`
- `flag` — CLI argument parsing in `cmd/lognojutsu/main.go`

## Configuration

**Runtime config:**
- No config files — all configuration is passed via CLI flags at startup or via the REST API at `/api/start`
- Engine config struct (`engine.Config` in `internal/engine/engine.go`) is JSON-decoded from HTTP request body
- No `.env` files or environment variable requirements detected

**Persistent state:**
- `lognojutsu_users.json` — user profile store written to the working directory by `internal/userstore/userstore.go`
- `lognojutsu_<timestamp>.log` — simulation log files written to working directory by `internal/simlog/simlog.go`
- `lognojutsu_report_<timestamp>.json` and `.html` — result reports written to working directory by `internal/reporter/reporter.go`

## Platform Requirements

**Development:**
- Windows with Go 1.26.1 toolchain
- `powershell.exe` available in PATH
- Administrator privileges required for preparation steps (audit policy, Sysmon install) and techniques marked `elevation_required: true`

**Production/Deployment:**
- Single self-contained Windows executable (`lognojutsu.exe`)
- No installer, no Docker, no cloud deployment detected
- Default bind: `127.0.0.1:8080` (loopback only unless `-host 0.0.0.0` is passed)

---

*Stack analysis: 2026-03-24*
