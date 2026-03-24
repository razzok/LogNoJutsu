# Integrations

**Analysis Date:** 2026-03-24

## External Services

**Microsoft Sysinternals — Sysmon:**
- Purpose: Optional endpoint telemetry agent installed during preparation
- Integration point: `internal/preparation/preparation.go` — `InstallSysmon()`
- Download URL (runtime, attempted automatically): `https://download.sysinternals.com/files/Sysmon.zip`
- Install method: `exec.Command(sysmonPath, "-accepteula", "-i", configPath)` using a bundled XML config
- Dependency type: Optional; preparation step gracefully degrades if download fails
- No SDK, no API key required

## APIs

**Internal REST API (self-hosted):**
- Implemented in `internal/server/server.go` via `net/http`
- All routes require optional Basic Auth (password set via `-password` flag)
- Base URL: `http://<host>:<port>/api/`

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/status` | GET | Engine phase and results |
| `/api/techniques` | GET | List all loaded techniques |
| `/api/campaigns` | GET | List all loaded campaigns |
| `/api/tactics` | GET | List all distinct tactics |
| `/api/start` | POST | Start simulation (JSON body: `engine.Config`) |
| `/api/stop` | POST | Abort running simulation |
| `/api/logs` | GET | In-memory simulation log entries |
| `/api/report` | GET | Serve latest HTML report file |
| `/api/prepare` | POST | Run all preparation steps |
| `/api/prepare/step` | POST | Run one preparation step (`powershell`, `auditpol`, `sysmon`) |
| `/api/users` | GET/POST | List or add user profiles |
| `/api/users/discover` | POST | Enumerate local + recent domain users |
| `/api/users/test` | POST | Validate stored credentials |
| `/api/users/delete` | POST | Delete a user profile |

**No external API integrations** — the tool does not call any third-party REST APIs, cloud services, or SIEM APIs at runtime.

## Data Stores

**Databases:**
- None — no SQL, NoSQL, or embedded database used

**File-based persistence:**
- `lognojutsu_users.json` — JSON array of `UserProfile` structs; written to working directory by `internal/userstore/userstore.go`; permissions `0600` (owner read/write only)
- `lognojutsu_<timestamp>[_<campaign>].log` — plaintext structured simulation log written by `internal/simlog/simlog.go`
- `lognojutsu_report_<timestamp>.json` — JSON simulation results written by `internal/reporter/reporter.go`
- `lognojutsu_report_<timestamp>.html` — self-contained HTML report written by `internal/reporter/reporter.go`; served via `/api/report`

**Embedded read-only data (compiled into binary):**
- YAML technique definitions: `internal/playbooks/embedded/techniques/*.yaml` (55+ files)
- YAML campaign definitions: `internal/playbooks/embedded/campaigns/*.yaml` (12 files)
- Static UI: `internal/server/static/index.html`
- All embedded via Go `//go:embed` directives at build time; no filesystem access needed at runtime for these assets

**Temporary files (runtime, auto-deleted):**
- `%TEMP%\lnj_<timestamp>.ps1` — temporary PowerShell script for `RunAs` execution (`internal/executor/executor.go`)
- `%TEMP%\lnj_out_<timestamp>.txt` / `lnj_err_<timestamp>.txt` — captured stdout/stderr from cross-user process launch
- `%TEMP%\lognojutsu_sysmon.xml` — Sysmon config written during preparation

## Authentication

**UI / API Authentication:**
- Optional HTTP Basic Auth enforced by `authMiddleware` in `internal/server/server.go`
- Password set via `-password` CLI flag at startup; empty string disables auth entirely
- Password checked against `r.BasicAuth()` password field only (no username checked)
- No session tokens, no JWT, no OAuth

**Credential Storage for Simulation Users:**
- User passwords stored encrypted in `lognojutsu_users.json`
- Encryption: Windows DPAPI via PowerShell `ConvertTo-SecureString` / `ConvertFrom-SecureString` (`internal/userstore/userstore.go` — `encryptDPAPI()` / `decryptDPAPI()`)
- DPAPI keys are machine + user scoped — encrypted passwords cannot be decrypted on a different machine or user account
- Fallback: if DPAPI encryption fails, password is stored with a `PLAIN:` prefix (logged as warning behavior)
- Encrypted password field (`password_enc`) is stripped from all API list responses — never sent to UI

**Credential Validation:**
- `TestCredentials()` in `internal/userstore/userstore.go` validates credentials via PowerShell `System.DirectoryServices.AccountManagement.PrincipalContext`
- Supports both local machine context and domain context based on whether a domain is set on the profile

## Notable Third-Party Dependencies

**`gopkg.in/yaml.v3 v3.0.1`** (only external Go dependency)
- Used: `internal/playbooks/loader.go`
- Purpose: Deserialise YAML technique and campaign playbook files into Go structs
- Maintained by the Go community; stable v3 API
- No transitive dependencies beyond `gopkg.in/check.v1` (test-only, not compiled into binary)

**Windows system tools (runtime dependencies, not Go packages):**

| Tool | Used in | Purpose |
|---|---|---|
| `powershell.exe` | `internal/executor/executor.go`, `internal/userstore/userstore.go`, `internal/preparation/preparation.go` | Technique execution, DPAPI crypto, credential validation, user discovery |
| `cmd.exe` | `internal/executor/executor.go` | Technique execution for `cmd`/`command_prompt` executor type |
| `auditpol.exe` | `internal/preparation/preparation.go` | Configure Windows Security audit subcategories |
| `sc.exe` | `internal/preparation/preparation.go` | Check Sysmon service status |
| `Sysmon64.exe` (optional) | `internal/preparation/preparation.go` | Endpoint telemetry; downloaded from Sysinternals if not present |

**Windows Event Log (read, not written directly):**
- `internal/userstore/userstore.go` — `DiscoverRecentDomainUsers()` reads Security event log (Event ID 4624) via `Get-WinEvent` PowerShell cmdlet to discover recently-logged-on domain users

---

*Integration audit: 2026-03-24*
