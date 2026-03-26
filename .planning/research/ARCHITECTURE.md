# Architecture Research

**Domain:** Go single-binary SIEM validation tool — v1.1 integration points
**Researched:** 2026-03-26
**Confidence:** HIGH (derived entirely from direct codebase inspection)

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     cmd/lognojutsu                           │
│  main.go — entry point, ldflags version var, banner print    │
├─────────────────────────────────────────────────────────────┤
│                    internal/server                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Server struct│  │ registerRoutes│ │ static/index.html   │  │
│  │ (eng, reg,  │  │ /api/info    │ │ (embedded, served    │  │
│  │  users, cfg)│  │ /api/prepare │ │  by FileServer,      │  │
│  └──────┬──────┘  └──────┬───────┘ │  no template engine) │  │
│         │                │         └─────────────────────┘  │
├─────────┴────────────────┴──────────────────────────────────┤
│                   internal/preparation                        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ preparation.go — Result struct, RunAll(),             │   │
│  │ EnablePowerShellLogging(), ConfigureAuditPolicy(),    │   │
│  │ InstallSysmon()                                       │   │
│  └──────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│              internal/{engine,playbooks,verifier,...}         │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | v1.1 Change |
|-----------|----------------|-------------|
| `cmd/lognojutsu/main.go` | Entry point, banner, `server.Start()` call | Add `var version` for ldflags injection |
| `internal/server/server.go` | HTTP routing, handler methods on `Server` struct | Add `version` field to `Server`, add `/api/info` handler |
| `internal/server/static/index.html` | Embedded single-page UI; served by `http.FileServer` | JS fetches `/api/info` on load; prep error detail elements |
| `internal/preparation/preparation.go` | Runs auditpol, PowerShell, Sysmon; returns `Result` | Replace locale-dependent subcategory names with GUIDs |

## Integration Points — v1.1 Features

### 1. Build-time Version via ldflags

**Current state:** `main.go` line 19 has `const banner = "... v0.1.0"`. `index.html` line 167 has `<span class="version-badge">v0.1.0</span>` as static HTML in the embedded filesystem. There is no Go template engine in use — `index.html` is served verbatim by `http.FileServer(http.FS(staticFS))`.

**Constraint:** `embed.FS` content is fixed at compile time. A Go `html/template.Execute` approach would require changing the static file handler from `http.FileServer` to a template-parsed handler — a larger refactor that is not warranted for a version badge.

**Recommended pattern — API endpoint:**

Declare a package-level var in `main.go` that ldflags can target:

```go
// cmd/lognojutsu/main.go
var version = "dev"   // overridden at build: -ldflags "-X main.version=v1.1.0"
```

Pass it into `server.Config` (add a `Version string` field) so `Server` can expose it:

```go
// internal/server/server.go
type Config struct {
    Host     string
    Port     int
    Password string
    Version  string  // NEW
}
```

Add `version string` to the `Server` struct, populated from `cfg.Version` in `Start()`.

Add a new route and handler:

```
GET /api/info  →  {"version": "v1.1.0"}
```

In `index.html`, fetch on page load and update the badge element:

```javascript
async function loadInfo() {
  const info = await api('/api/info');
  const el = document.querySelector('.version-badge');
  if (el && info.version) el.textContent = info.version;
}
```

Call `loadInfo()` in the existing `init()` or `document.addEventListener('DOMContentLoaded', ...)` block that currently calls `refresh()`.

**Banner update:** Replace `const banner` with a function or `fmt.Sprintf` that interpolates the `version` var so the CLI banner is also dynamic.

**Build command target:**

```
go build -ldflags "-X main.version=v1.1.0" -o lognojutsu.exe ./cmd/lognojutsu
```

**Data flow:**

```
go build -ldflags "-X main.version=v1.1.0"
    → main.version = "v1.1.0"
    → server.Config{Version: version}
    → Server{version: cfg.Version}
    → GET /api/info → {"version":"v1.1.0"}
    → JS fetch → document.querySelector('.version-badge').textContent = "v1.1.0"
```

**Files modified:** `cmd/lognojutsu/main.go`, `internal/server/server.go`, `internal/server/static/index.html`

**Files NOT changed:** `embed.FS` embedding, `registerRoutes` signature (just add one line)

---

### 2. Locale-independent Audit Policy GUIDs

**Current state:** `ConfigureAuditPolicy()` in `preparation.go` calls `auditpol.exe /set /subcategory:Logon` using English-language subcategory names. On German (or any non-English) Windows these names resolve differently and auditpol exits non-zero, producing failures.

**Fix:** Replace each `/subcategory:NAME` with `/subcategory:{GUID}`. GUIDs are locale-independent and stable across Windows versions.

**Organization decision — inline vs separate file:**

The current file has 12 subcategory entries defined as an anonymous struct slice inside `ConfigureAuditPolicy()`. The GUIDs are straight substitutions for the existing strings — the data shape does not change.

Recommendation: Keep the definitions inline in `preparation.go`, replacing the string values. Do NOT create a separate constants file. Rationale:
- The GUIDs are only ever consumed in one place (`ConfigureAuditPolicy`).
- A `constants.go` file in the same package adds a navigation hop with no abstraction benefit.
- If the subcategory list grows significantly in a future milestone, extraction to a constants file is a valid follow-up.

The struct slice becomes:

```go
policies := []struct {
    subcategory string
    description string
}{
    {"{0CCE9215-69AE-11D9-BED3-505054503030}", "Logon/Logoff events (4624, 4625, 4634)"},
    {"{0CCE9216-69AE-11D9-BED3-505054503030}", "Logoff events"},
    // ... remaining GUIDs
}
```

The `auditpol.exe` call itself is unchanged — only the subcategory value changes from a locale-sensitive name to a GUID string.

**Files modified:** `internal/preparation/preparation.go` (12 string values replaced inline)

**Auditpol subcategory GUIDs (Windows Security audit policy):**

| Subcategory (English name) | GUID |
|---------------------------|------|
| Logon | {0CCE9215-69AE-11D9-BED3-505054503030} |
| Logoff | {0CCE9216-69AE-11D9-BED3-505054503030} |
| Account Lockout | {0CCE9217-69AE-11D9-BED3-505054503030} |
| Process Creation | {0CCE922B-69AE-11D9-BED3-505054503030} |
| Audit Policy Change | {0CCE922F-69AE-11D9-BED3-505054503030} |
| Security Group Management | {0CCE9237-69AE-11D9-BED3-505054503030} |
| User Account Management | {0CCE9235-69AE-11D9-BED3-505054503030} |
| Sensitive Privilege Use | {0CCE9228-69AE-11D9-BED3-505054503030} |
| Special Logon | {0CCE921B-69AE-11D9-BED3-505054503030} |
| Other Object Access Events | {0CCE9227-69AE-11D9-BED3-505054503030} |
| Scheduled Task | {0CCE9232-69AE-11D9-BED3-505054503030} |
| Filtering Platform Connection | {0CCE9226-69AE-11D9-BED3-505054503030} |

Confidence for these GUIDs: MEDIUM. The GUID format and most values are widely documented in Microsoft Security Compliance Toolkit and auditpol documentation. Verify against `auditpol /list /subcategory:* /v` output on a test machine before shipping — this command outputs the GUID next to each English name.

---

### 3. Preparation Error Surfacing to UI

**Current state:** The `Result` struct already carries a `Message string` field with actionable text (e.g., `"Partial failure: Logon: exit status 1 (error details)"` or the Sysmon download-failed message with the manual URL).

The UI currently surfaces errors in two ways, both inadequate:
- Line 1049: `document.getElementById('status-' + id).title = r.message` — puts message in a tooltip (invisible unless hovering)
- Line 1058: `alert('Step failed:\n' + r.message)` — only fires for per-step runs, not the Run All path (line 1044-1050 uses `title` only)

**API response format is already correct.** The `Result` struct returned by both `/api/prepare` (array) and `/api/prepare/step` (single object) is:

```json
{
  "step": "Windows Audit Policy",
  "success": false,
  "message": "Partial failure: Logon: exit status 1 (...); Process Creation: exit status 1 (...)"
}
```

No changes to the Go API layer are needed for error surfacing.

**Required changes are UI-only (`index.html`):**

Each preparation step row currently has this structure:

```html
<div class="prep-step" id="step-auditpol">
  <div class="prep-step-info">
    <h3>Windows Audit Policy</h3>
    <p>Enable audit subcategories for SIEM-relevant event IDs.</p>
  </div>
  <span class="prep-status" id="status-auditpol">—</span>
  <button ...>Run</button>
</div>
```

Add a detail element per step for error text:

```html
<div class="prep-error" id="error-auditpol" style="display:none; ..."></div>
```

Update `setPrepStatus()` and the Run All handler to populate this element when `r.success === false`:

```javascript
function setPrepStatus(id, state, text, message) {
  const statusEl = document.getElementById('status-' + id);
  statusEl.textContent = text;
  statusEl.className = 'prep-status ' + (state === 'ok' ? 'status-ok' : state === 'fail' ? 'status-fail' : 'status-running');

  const errEl = document.getElementById('error-' + id);
  if (errEl) {
    if (state === 'fail' && message) {
      errEl.textContent = message;
      errEl.style.display = 'block';
    } else {
      errEl.style.display = 'none';
    }
  }
}
```

The `runAllPrep()` handler at line 1045-1050 and `runPrepStep()` at line 1053-1058 both need to pass `r.message` through to `setPrepStatus`.

Remove the `alert()` call on line 1058 — inline display is better UX.

**Files modified:** `internal/server/static/index.html` only

---

## Recommended Project Structure (unchanged packages)

```
cmd/lognojutsu/
    main.go              # var version = "dev"  (ldflags target)

internal/server/
    server.go            # Config.Version, Server.version, /api/info handler
    server_test.go       # existing tests unaffected; optionally add /api/info test
    static/
        index.html       # loadInfo() JS, prep error detail elements

internal/preparation/
    preparation.go       # GUID strings replacing locale names (inline, same file)
```

No new files or packages are required for v1.1.

## Data Flow

### Version Flow

```
go build -ldflags "-X main.version=v1.1.0"
    ↓
main.version var = "v1.1.0"
    ↓
server.Config{Version: version}  →  server.Start(cfg)
    ↓
Server{version: cfg.Version}
    ↓
GET /api/info  →  {"version": "v1.1.0"}
    ↓
JS: fetch('/api/info').then(info => badge.textContent = info.version)
    ↓
<span class="version-badge">v1.1.0</span>
```

### Preparation Error Flow

```
User clicks "Run All" or "Run" on a step
    ↓
POST /api/prepare  or  POST /api/prepare/step
    ↓
preparation.RunAll()  or  preparation.ConfigureAuditPolicy()
    ↓
Result{Step, Success: false, Message: "auditpol GUID failed: exit status 1 (...)"}
    ↓
writeJSON(w, result)
    ↓
JS receives r.success=false, r.message="..."
    ↓
setPrepStatus(id, 'fail', '✗ Failed', r.message)
    ↓
#error-auditpol becomes visible with actionable message text
```

## Anti-Patterns

### Anti-Pattern 1: Go html/template for version badge

**What people do:** Replace `http.FileServer` with a `html/template.ParseFS` + `ExecuteTemplate` handler to inject version at serve time.

**Why it's wrong:** The entire UI is one large `index.html` file. Switching from `http.FileServer` to template execution means adding template delimiters throughout the file to avoid parse errors, changing all `{{` occurrences in JavaScript (Go templates use the same delimiters), and breaking the existing test pattern. The benefit (one fewer HTTP round-trip vs. `/api/info`) does not justify the churn.

**Do this instead:** Add `/api/info` JSON endpoint. One small JS fetch call on page load, zero changes to the file serving path.

### Anti-Pattern 2: Separate constants file for 12 GUID strings

**What people do:** Create `internal/preparation/constants.go` with named `const` values for each GUID.

**Why it's wrong:** The GUIDs are consumed exactly once, in one function. Named constants like `const AuditLogon = "{0CCE9215...}"` add no clarity over the inline struct entry with its `description` field already explaining what the GUID represents. Named constants are appropriate when the same value is referenced in multiple places.

**Do this instead:** Replace the English-language strings inline in the struct slice. The `description` field already documents what each GUID does.

### Anti-Pattern 3: Using `alert()` for preparation errors

**What people do:** Keep the existing `alert('Step failed:\n' + r.message)` for step failures.

**Why it's wrong:** `alert()` is blocking, not styled consistently with the UI, and doesn't allow the user to copy the error text easily. It also only fires for per-step runs, not the Run All path, creating inconsistent error behaviour.

**Do this instead:** Render error messages inline beneath the failing step using a styled `<div>` that the JS shows/hides based on `r.success`.

## Integration: New vs Modified

| Item | New or Modified | File | Scope |
|------|----------------|------|-------|
| `var version = "dev"` | Modified | `cmd/lognojutsu/main.go` | Replace `const`-embedded string with interpolated var |
| `Config.Version string` | Modified | `internal/server/server.go` | Add field to existing struct |
| `Server.version string` | Modified | `internal/server/server.go` | Add field to existing struct |
| `handleInfo` handler | New | `internal/server/server.go` | ~10 lines, returns `{"version":...}` |
| `GET /api/info` route | New | `internal/server/server.go` | One line in `registerRoutes` |
| GUID strings | Modified | `internal/preparation/preparation.go` | 12 string values replaced in-place |
| `loadInfo()` JS | New | `internal/server/static/index.html` | ~5 lines, called on page load |
| Prep error `<div>` elements | New | `internal/server/static/index.html` | 3 elements (one per step) |
| `setPrepStatus()` signature | Modified | `internal/server/static/index.html` | Add `message` param, show/hide error div |
| Remove `alert()` call | Modified | `internal/server/static/index.html` | Line 1058 — delete |

## Build Order Considerations

1. **`preparation.go` GUID change** has no dependencies — can be done and tested independently with `auditpol /list /subcategory:* /v` verification.
2. **`server.go` + `main.go` version plumbing** is a self-contained Go change — compile and test `/api/info` with `curl` before touching HTML.
3. **`index.html` version badge** depends on `/api/info` existing — do after step 2.
4. **`index.html` prep error surfacing** is independent of version work — can be done in parallel with steps 2-3.

## Sources

- Direct inspection: `cmd/lognojutsu/main.go`, `internal/server/server.go`, `internal/preparation/preparation.go`, `internal/server/static/index.html`
- Microsoft auditpol GUID reference: documented in Security Compliance Toolkit and via `auditpol /list /subcategory:* /v` on any Windows system
- Go ldflags documentation: `go help build`, `-X importpath.name=value` flag

---
*Architecture research for: LogNoJutsu v1.1 Bug Fixes & UI Polish*
*Researched: 2026-03-26*
