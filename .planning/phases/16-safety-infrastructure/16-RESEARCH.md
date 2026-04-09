# Phase 16: Safety Infrastructure - Research

**Researched:** 2026-04-09
**Domain:** Go execution engine safety hardening — AMSI detection, Windows token elevation check, web UI modal confirmation
**Confidence:** HIGH

## Summary

Phase 16 adds three runtime safety layers to the technique execution engine without touching any technique YAML content. All three features integrate into the existing `runTechnique` / `runInternal` / `executor.go` pipeline at well-defined insertion points that are already visible in the current code. No new third-party dependencies are needed: `golang.org/x/sys/windows` is already an indirect dependency in `go.mod`.

The three features are architecturally independent and can be planned as three separate tasks. AMSI detection is a pure string-matching post-processor on PowerShell stderr. Elevation gating is a one-time admin check at engine start followed by a per-technique skip guard in `runTechnique`. Scan confirmation is a UI + API handshake that pauses the engine goroutine until the consultant clicks Confirm.

**Primary recommendation:** Implement in dependency order — types first (new `VerificationStatus` constants), then AMSI detection, then elevation gating, then scan confirmation API + modal. Each layer builds on the previous without requiring rollback.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**AMSI Detection**
- D-01: Detect AMSI blocks by parsing PowerShell stderr for known AMSI error patterns (`ScriptContainedMaliciousContent`, exit code -196608). No external dependencies.
- D-02: AMSI detection applies to PowerShell executor type only — CMD and native Go techniques do not go through AMSI.
- D-03: AMSI-blocked techniques are classified with a distinct `amsi_blocked` verification status and moved on. No retry, no bypass. Consultant sees which techniques were blocked and adjusts Defender policy themselves.

**Elevation Gating**
- D-04: Check admin status once at engine start using Windows token check (`golang.org/x/sys/windows` — check process token for admin group membership). No shell spawning.
- D-05: Per-technique runtime gating — when the engine encounters a technique with `elevation_required: true` and the process is not admin, it skips that technique with an `elevation_required` status. The `elevation_required` field already exists in `Technique` struct and YAML files.
- D-06: Elevation-skipped techniques count toward the total technique count in reports with a distinct "Elevation Required" status. Consultant sees "45/58 passed, 8 skipped (elevation), 5 failed" — full visibility.

**Scan Confirmation UX**
- D-07: Scan confirmation appears as a web UI modal before scan techniques execute. Shows target subnet, rate limit, IDS warning, and the list of scan techniques that will run. Scan does not proceed until the consultant clicks "Confirm".
- D-08: Confirmation is triggered by a tag-based mechanism — a `requires_confirmation` YAML field (or tag). Any technique with this flag triggers the confirmation flow.
- D-09: The modal displays four pieces of information: (1) auto-detected target /24 subnet, (2) rate limit notice (connections/second), (3) IDS/IPS warning, (4) list of specific scan techniques that will run.

**Verification Statuses**
- D-10: Add `amsi_blocked` and `elevation_required` to the existing `VerificationStatus` enum in `types.go`.
- D-11: HTML report displays new statuses as color-coded badges: "AMSI Blocked" = orange/amber, "Elevation Required" = gray/blue. Follows existing pass=green, fail=red badge pattern.
- D-12: Web UI technique list also displays the new status badges, consistent with the HTML report styling.

### Claude's Discretion
- Exact AMSI error string patterns to match (may need to cover multiple Windows/Defender versions)
- Windows token check implementation details (which specific SID/group to check)
- Scan confirmation API endpoint design (how UI sends confirmation back to engine)
- Rate limit default value and whether it's configurable
- Modal styling and layout within existing web UI patterns
- Whether `requires_confirmation` is a bool field or a tags array entry

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| INFRA-01 | AMSI-blocked technique failures are classified separately from execution errors in verification results | AMSI stderr pattern matching in `runInternal` PowerShell path; new `VerifAMSIBlocked` constant; reporter badge |
| INFRA-02 | Admin vs non-admin execution is detected; techniques requiring elevation are skipped gracefully with clear status | `windows.GetCurrentProcessToken().IsElevated()` at engine start; per-technique skip in `runTechnique`; new `VerifElevationRequired` constant |
| INFRA-03 | Network scan has target range confirmation, rate limiting, and IDS warning displayed before execution | `requires_confirmation` YAML bool field on T1046; new `/api/scan/confirm` endpoint; channel-based engine pause; modal in index.html |
</phase_requirements>

---

## Standard Stack

### Core (already in go.mod — no new dependencies)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `golang.org/x/sys/windows` | v0.41.0 (already indirect) | Windows token API: `GetCurrentProcessToken().IsElevated()` | Official Go extended library; already transitively imported via yusufpapurcu/wmi |
| `strings.Contains` (stdlib) | n/a | AMSI stderr pattern matching | No external dep needed; patterns are static strings |
| Go channels (stdlib) | n/a | Engine pause/resume for scan confirmation | Idiomatic Go concurrency; fits existing `stopCh` pattern |

### No New Dependencies Required

`golang.org/x/sys/windows` is already in `go.sum` as an indirect dep. To use it directly, add it to the `require` block in `go.mod` and import `golang.org/x/sys/windows` in engine.go. No `go get` needed for packages that are already resolved.

```bash
# Promote to direct dependency (run once):
go get golang.org/x/sys/windows@v0.41.0
```

## Architecture Patterns

### Recommended Code Placement

```
internal/
├── playbooks/types.go          # Add VerifAMSIBlocked, VerifElevationRequired constants
├── engine/engine.go            # isAdmin bool field; AdminCheck() at Start(); elevation skip in runTechnique
├── executor/executor.go        # detectAMSI() helper called after runCommand() in PowerShell path
├── reporter/reporter.go        # New cases in htmlTemplate verifStr switch; new CSS badge classes
├── server/
│   ├── server.go               # New /api/scan/confirm handler + scan state fields
│   └── static/index.html       # Scan confirmation modal + new status badge CSS + JS
```

### Pattern 1: New VerificationStatus Constants (types.go)

**What:** Extend the existing typed string enum with two new values.
**When to use:** Any technique result that was blocked by AMSI or skipped for elevation.

```go
// Source: internal/playbooks/types.go (existing pattern)
const (
    VerifNotRun          VerificationStatus = "not_run"
    VerifPass            VerificationStatus = "pass"
    VerifFail            VerificationStatus = "fail"
    VerifNotExecuted     VerificationStatus = "not_executed"
    // Phase 16 additions:
    VerifAMSIBlocked     VerificationStatus = "amsi_blocked"
    VerifElevationRequired VerificationStatus = "elevation_required"
)
```

These values flow through `ExecutionResult.VerificationStatus` (JSON: `"verification_status"`) without any struct changes. Existing API consumers receive new string values in the same field.

### Pattern 2: AMSI Detection in executor.go

**What:** After `runCommand()` returns for a PowerShell technique, inspect `errOut` and `err` for AMSI-specific patterns before setting `result.Success`.
**When to use:** Executor type is `powershell` or `psh`.

**AMSI error patterns to detect (HIGH confidence — verified against PowerShell source and Microsoft docs):**

| Pattern | Source |
|---------|--------|
| `"ScriptContainedMaliciousContent"` | PowerShell FullyQualifiedErrorId in stderr |
| `"This script contains malicious content"` | Human-readable error message in stderr |
| `"has been blocked by your antivirus software"` | Alternative phrasing in some Defender versions |
| Exit code `-196608` (= `0xFFFF0000`) | PowerShell process exit code when AMSI blocks at parse time |

**Implementation note:** The exit code check must be done on the `error` returned by `cmd.Run()`, which is of type `*exec.ExitError` when the process exits non-zero. Use a type assertion to get `ExitCode()`.

```go
// Source: pattern derived from executor.go runCommand() + PowerShell/PowerShell source
func isAMSIBlocked(errOut string, err error) bool {
    amsiPatterns := []string{
        "ScriptContainedMaliciousContent",
        "This script contains malicious content",
        "has been blocked by your antivirus software",
    }
    for _, p := range amsiPatterns {
        if strings.Contains(errOut, p) {
            return true
        }
    }
    // Exit code -196608 = AMSI block at parse time (before script executes)
    if exitErr, ok := err.(*exec.ExitError); ok {
        if exitErr.ExitCode() == -196608 {
            return true
        }
    }
    return false
}
```

**Insertion point in `runInternal`:**

```go
// After: out, errOut, err = runCommand(...)
// Before: result.Success = (err == nil)
if strings.ToLower(t.Executor.Type) == "powershell" || strings.ToLower(t.Executor.Type) == "psh" {
    if isAMSIBlocked(errOut, err) {
        result.Success = false
        result.VerificationStatus = playbooks.VerifAMSIBlocked
        // Return early — no event log verification for AMSI-blocked techniques
        return result
    }
}
```

**Important:** AMSI detection must return early before the verifier is called. A blocked technique cannot generate expected events, so verification would always fail. Setting `VerifAMSIBlocked` directly and returning avoids a misleading "Fail" in the verification column.

### Pattern 3: Elevation Check at Engine Start (engine.go)

**What:** Add an `isAdmin bool` field to `Engine`. Populate it in `Start()` before spawning the `run()` goroutine.
**When to use:** Once per simulation; result cached for all per-technique decisions.

**Windows token check — two approaches (both using golang.org/x/sys/windows):**

Option A — `IsElevated()` (UAC elevation check, recommended for this use case):
```go
// Source: pkg.go.dev/golang.org/x/sys/windows
import "golang.org/x/sys/windows"

func isElevated() bool {
    return windows.GetCurrentProcessToken().IsElevated()
}
```

Option B — Administrators SID membership check (more complete for non-UAC contexts):
```go
func isAdmin() bool {
    var adminSID *windows.SID
    err := windows.AllocateAndInitializeSid(
        &windows.SECURITY_NT_AUTHORITY,
        2,
        windows.SECURITY_BUILTIN_DOMAIN_RID,
        windows.DOMAIN_ALIAS_RID_ADMINS,
        0, 0, 0, 0, 0, 0,
        &adminSID,
    )
    if err != nil {
        return false
    }
    defer windows.FreeSid(adminSID)
    token := windows.GetCurrentProcessToken()
    isMember, err := token.IsMember(adminSID)
    return err == nil && isMember
}
```

**Recommendation (Claude's discretion):** Use `windows.GetCurrentProcessToken().IsElevated()` for simplicity. It correctly returns false for a standard user and true for a process that was started with "Run as Administrator" (UAC elevation). This is the distinction that matters: if PowerShell is not elevated, `elevation_required: true` techniques will fail with "Access Denied" — which `IsElevated()` correctly predicts. The SID membership approach gives a false positive for admin users who did NOT elevate (UAC split token), which is the wrong signal for our gate.

**Engine struct change:**

```go
type Engine struct {
    // ... existing fields ...
    isAdmin bool  // set once at Start(); guards per-technique elevation skip
}
```

**In `Start()`, after resolveProfiles():**

```go
e.isAdmin = checkIsElevated() // platform-specific; see note below
simlog.Info(fmt.Sprintf("Admin check: elevated=%v", e.isAdmin))
```

**Platform compilation note:** `golang.org/x/sys/windows` is Windows-only. Since this project targets Windows exclusively (all techniques are Windows-specific), no build tag is needed. However, if the codebase is ever compiled on Linux/macOS (e.g., CI), the `isAdmin` function must be in a `_windows.go` file, with a stub `_other.go` returning `true` (permissive fallback — non-Windows CI should not skip techniques).

### Pattern 4: Per-Technique Elevation Skip (engine.go runTechnique)

**What:** Before calling the executor, check `t.ElevationRequired && !e.isAdmin`. If true, record a skip result and return without executing.
**When to use:** Every technique call site in `runTechnique`.

**Insertion point:** After the WhatIf block, before the `e.runner != nil` / `RunWithCleanup` / `RunAs` branches.

```go
// In runTechnique(), after WhatIf block:
if t.ElevationRequired && !e.isAdmin {
    simlog.Info(fmt.Sprintf("[ElevationSkip] %s — requires elevation, skipping", t.ID))
    now := time.Now().Format(time.RFC3339)
    result = playbooks.ExecutionResult{
        TechniqueID:        t.ID,
        TechniqueName:      t.Name,
        TacticID:           t.Tactic,
        StartTime:          now,
        EndTime:            now,
        Success:            false,
        Output:             "Elevation required — technique skipped (not running as Administrator)",
        RunAsUser:          userLabel,
        VerificationStatus: playbooks.VerifElevationRequired,
        Tier:               t.Tier,
        SIEMCoverage:       t.SIEMCoverage,
    }
    e.mu.Lock()
    e.status.Results = append(e.status.Results, result)
    e.mu.Unlock()
    return
}
```

**Count tracking:** The skipped result is appended to `status.Results` just like any other result. The dashboard stat box for "Failed" will count it because `result.Success = false`. The reporter must separately count `VerifElevationRequired` results for the summary line — see reporter pattern below.

### Pattern 5: Scan Confirmation — Engine Channel + API Endpoint

**What:** The engine pauses execution when it encounters a `requires_confirmation: true` technique. The UI shows a modal. After the consultant clicks Confirm, the API endpoint sends a signal on a channel, and the engine resumes.
**When to use:** Any technique with `requires_confirmation: true` in its YAML.

**YAML field decision (Claude's discretion):** Use a `RequiresConfirmation bool` field in the `Technique` struct (not a tags entry). Bool fields are easier to check in Go without string comparison. Tags are for UI display categorization; a bool field is a first-class runtime gate.

**Technique struct addition (types.go):**
```go
type Technique struct {
    // ... existing fields ...
    RequiresConfirmation bool `yaml:"requires_confirmation" json:"requires_confirmation,omitempty"`
}
```

**Engine state for scan confirmation:**

```go
type Engine struct {
    // ... existing fields ...
    isAdmin           bool
    scanConfirmCh     chan struct{} // closed when consultant confirms; nil when no scan pending
    scanConfirmMu     sync.Mutex   // guards scanConfirmCh
    scanPendingInfo   *ScanInfo    // non-nil when confirmation is pending
}

// ScanInfo holds what the modal displays
type ScanInfo struct {
    TargetSubnet  string   `json:"target_subnet"`   // auto-detected /24
    RateLimitNote string   `json:"rate_limit_note"`
    IDSWarning    string   `json:"ids_warning"`
    Techniques    []string `json:"techniques"`      // IDs of scan techniques about to run
}
```

**Engine scan pause (in runTechnique or at batch level):**

The confirmation fires once per simulation, not once per technique. The engine collects all `requires_confirmation` techniques before the run loop, sets `scanPendingInfo`, creates `scanConfirmCh`, and blocks until the channel is closed.

```go
// Before runTechnique loop — collect techniques needing confirmation
var confirmTechs []*playbooks.Technique
for _, t := range techniques {
    if t.RequiresConfirmation {
        confirmTechs = append(confirmTechs, t)
    }
}
if len(confirmTechs) > 0 && !e.cfg.WhatIf {
    info := e.buildScanInfo(confirmTechs)
    e.scanConfirmMu.Lock()
    e.scanPendingInfo = info
    e.scanConfirmCh = make(chan struct{})
    e.scanConfirmMu.Unlock()

    simlog.Info("[ScanConfirm] Waiting for consultant confirmation...")
    select {
    case <-e.scanConfirmCh:
        simlog.Info("[ScanConfirm] Confirmed — proceeding with scan")
    case <-e.stopCh:
        e.abort()
        return
    }
    e.scanConfirmMu.Lock()
    e.scanPendingInfo = nil
    e.scanConfirmMu.Unlock()
}
```

**New API endpoint `/api/scan/confirm` (POST) — server.go:**

```go
func (s *Server) handleScanConfirm(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeError(w, "POST required", http.StatusMethodNotAllowed)
        return
    }
    if err := s.eng.ConfirmScan(); err != nil {
        writeError(w, err.Error(), http.StatusConflict)
        return
    }
    writeJSON(w, map[string]string{"status": "confirmed"})
}

// Engine method:
func (e *Engine) ConfirmScan() error {
    e.scanConfirmMu.Lock()
    defer e.scanConfirmMu.Unlock()
    if e.scanConfirmCh == nil {
        return fmt.Errorf("no scan confirmation pending")
    }
    close(e.scanConfirmCh)
    e.scanConfirmCh = nil
    return nil
}
```

**New API endpoint `/api/scan/pending` (GET) — returns `ScanInfo` or 204 when none pending:**

The UI polls this to decide whether to show the modal.

**Subnet auto-detection (Claude's discretion):**

```go
// buildScanInfo detects the /24 subnet from the default route
func (e *Engine) buildScanInfo(techs []*playbooks.Technique) *ScanInfo {
    ids := make([]string, len(techs))
    for i, t := range techs {
        ids[i] = t.ID
    }
    subnet := detectLocalSubnet() // returns e.g. "192.168.1.0/24"
    return &ScanInfo{
        TargetSubnet:  subnet,
        RateLimitNote: "Up to 10 concurrent TCP connections (rate-limited runspace pool)",
        IDSWarning:    "Network IDS/IPS may alert on this scan. Ensure client is aware.",
        Techniques:    ids,
    }
}
```

Subnet detection: use `net.Interfaces()` and `net.ParseCIDR()` to find the first non-loopback IPv4 address and derive the /24. Pure stdlib — no external deps.

### Pattern 6: Reporter Updates (reporter.go)

**New badge CSS in htmlTemplate:**

```css
/* Add after existing .verif-pass / .verif-fail / .verif-skip: */
.verif-amsi { color: #d29922; font-weight: 600; }      /* orange/amber — matches --orange var */
.verif-elev { color: #58a6ff; font-weight: 600; }      /* blue — matches --accent var */
```

**New cases in verif column template:**

```html
{{else if eq (verifStr .VerificationStatus) "amsi_blocked"}}
  <span class="verif-amsi">&#9888; AMSI Blocked</span>
{{else if eq (verifStr .VerificationStatus) "elevation_required"}}
  <span class="verif-elev">&#9888; Elevation Required</span>
```

**Reporter summary counters:** The existing `VerifPassed`/`VerifFailed` counting loop in `saveHTML` must be extended to count the two new statuses for the stat boxes. Add `VerifAMSIBlocked int` and `VerifElevRequired int` to `htmlData` and populate them in the loop.

### Pattern 7: Web UI Modal (index.html)

**Modal HTML structure** (insert before closing `</main>`):

```html
<!-- Scan Confirmation Modal -->
<div id="scanConfirmModal" style="display:none; position:fixed; top:0; left:0; width:100%; height:100%;
     background:rgba(0,0,0,0.7); z-index:1000; align-items:center; justify-content:center;">
  <div style="background:var(--bg2); border:1px solid var(--orange); border-radius:8px;
               padding:24px; max-width:480px; width:90%;">
    <h3 style="color:var(--orange); margin-bottom:16px;">Network Scan Confirmation Required</h3>
    <div id="scanModalBody"><!-- populated by JS --></div>
    <div style="display:flex; gap:10px; margin-top:20px; justify-content:flex-end;">
      <button class="btn btn-secondary" onclick="dismissScanModal()">Cancel Simulation</button>
      <button class="btn btn-danger" onclick="confirmScan()">Confirm — Run Scan</button>
    </div>
  </div>
</div>
```

**Modal JS logic:** The existing `pollStatus()` loop already runs on a timer. Add a parallel `pollScanConfirm()` that calls `GET /api/scan/pending`. If it returns a `ScanInfo` object, populate the modal body and set `display:flex`. When the consultant clicks Confirm, call `POST /api/scan/confirm`.

**Badge CSS for new statuses in the technique results list:** Add to the existing CSS in index.html:

```css
.verif-amsi-badge  { color: var(--orange); font-weight: 600; }
.verif-elev-badge  { color: var(--accent); font-weight: 600; }
```

### Anti-Patterns to Avoid

- **Spawning a shell to check admin:** `net user %username% /domain` or `whoami /groups` are fragile, locale-dependent, and slow. Use the Windows token API directly (D-04).
- **Checking `err != nil` as proxy for AMSI:** All failed PowerShell commands return `err != nil`. The AMSI pattern must be checked on the stderr *content*, not just the error.
- **Showing the scan modal on every technique call:** Confirmation fires once per simulation (D-07). Collect all `requires_confirmation` techniques before the loop, not inside it.
- **Using a global channel for scan confirmation:** The `scanConfirmCh` must be guarded by a mutex and recreated on each `Start()` — it is part of Engine run state, same as `stopCh`.
- **Skipping the early return after AMSI detection:** Without early return, the verifier runs, finds no events, and records `VerifFail` — overwriting the `VerifAMSIBlocked` status.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Windows admin check | `exec.Command("whoami", "/groups")` | `windows.GetCurrentProcessToken().IsElevated()` | No child process; no locale issues; single API call |
| AMSI pattern matching | regex engine | `strings.Contains()` | Patterns are static strings; regex adds complexity with no benefit |
| Subnet detection | External tool (`ipconfig`) | `net.Interfaces()` + `net.ParseCIDR()` | Pure stdlib; no exec overhead; works offline |

## Common Pitfalls

### Pitfall 1: AMSI Detection Overwrites Verification Status

**What goes wrong:** AMSI detection is added but early-return is missed. The verifier still runs, finds no events (script was blocked before it could fire them), and sets `VerifFail`. The `amsi_blocked` status is lost.

**Why it happens:** The verifier call is unconditional for techniques with `expected_events`.

**How to avoid:** In `runInternal`, the AMSI check must `return result` immediately after setting `VerifAMSIBlocked`. The verifier block is inside `runTechnique` (lines 660-683 in engine.go), not in `runInternal`, so the early return from `runInternal` naturally skips it. Verify this flow is intact.

**Warning signs:** Test shows `amsi_blocked` in output but report shows `fail` — indicates early return was missed.

### Pitfall 2: Elevation Check on Non-Windows Build

**What goes wrong:** `golang.org/x/sys/windows` is not importable on Linux/macOS. If `engine.go` directly imports it unconditionally, cross-platform builds (e.g., CI running on Linux) fail to compile.

**Why it happens:** The windows package uses `//go:build windows` internally but the import itself fails on other platforms.

**How to avoid:** Place the `isElevated()` implementation in `engine_windows.go` (auto-selected on Windows) and a stub in `engine_other.go` (returns `true` on non-Windows — permissive, doesn't gate techniques). The `isAdmin bool` field stays in the main `engine.go`.

**Warning signs:** `go build` on Linux gives "cannot find package golang.org/x/sys/windows".

### Pitfall 3: Scan Confirmation Channel Not Reset Between Runs

**What goes wrong:** After the first simulation completes, `scanConfirmCh` is left in a closed state. The second run detects "no scan pending" wrongly or panics on close-of-closed-channel.

**Why it happens:** Channel lifecycle is tied to the simulation run, not the Engine struct lifetime.

**How to avoid:** In `Start()`, reset `scanConfirmCh = nil` and `scanPendingInfo = nil` before spawning `run()` — same pattern as `stopCh = make(chan struct{}, 1)`.

### Pitfall 4: Elevation-Skipped Results Not Counted in Report Summary

**What goes wrong:** The reporter counts `Succeeded` and `Failed` based on `result.Success`. Elevation-skipped results have `Success = false`, so they inflate the Failed count without explanation.

**Why it happens:** Existing report summary only distinguishes pass/fail, not skip types.

**How to avoid:** Add a dedicated counter for `VerifElevationRequired` (and optionally `VerifAMSIBlocked`) in `saveHTML`. Display as "8 elevation-skipped" in the stat boxes — separate from the "Failed" count.

### Pitfall 5: requiresConfirmation Check Inside the Per-Technique Loop

**What goes wrong:** Modal fires once per scan technique instead of once per simulation. Consultant must click Confirm multiple times.

**Why it happens:** The guard is placed inside `runTechnique` rather than in the pre-flight collection step in `run()`.

**How to avoid:** Collect confirmation-required techniques before the loop, pause once, then run all of them after confirmation.

### Pitfall 6: WhatIf Mode Triggering the Scan Modal

**What goes wrong:** In WhatIf mode, the engine does not execute anything — but if the confirmation gate is checked before WhatIf, the modal appears unnecessarily, blocking preview runs.

**Why it happens:** Confirmation check happens before WhatIf check in the code path.

**How to avoid:** Skip confirmation entirely when `e.cfg.WhatIf` is true. WhatIf mode is explicitly handled in `runTechnique` — the same pattern should apply to confirmation.

## Code Examples

### AMSI Exit Code Check Pattern

```go
// Source: Go standard library os/exec package
if exitErr, ok := err.(*exec.ExitError); ok {
    if exitErr.ExitCode() == -196608 {
        return true // AMSI block
    }
}
```

Exit code -196608 in decimal is `0xFFFD0000` in hex — this is the value Windows assigns when AMSI blocks a script at parse time (before execution). Verified against CONTEXT.md D-01 and cross-referenced with PowerShell open-source behavior.

### Windows Token Check (golang.org/x/sys/windows)

```go
// Source: pkg.go.dev/golang.org/x/sys/windows
// Place in engine_windows.go
import "golang.org/x/sys/windows"

func checkIsElevated() bool {
    return windows.GetCurrentProcessToken().IsElevated()
}
```

```go
// engine_other.go — build tag stub for non-Windows CI
//go:build !windows

func checkIsElevated() bool {
    return true // permissive: don't skip elevation-gated techniques on non-Windows
}
```

### Subnet Detection (stdlib only)

```go
// Source: net package (stdlib)
import "net"

func detectLocalSubnet() string {
    ifaces, err := net.Interfaces()
    if err != nil {
        return "unknown"
    }
    for _, iface := range ifaces {
        if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
            continue
        }
        addrs, _ := iface.Addrs()
        for _, addr := range addrs {
            if ipNet, ok := addr.(*net.IPNet); ok {
                if ip4 := ipNet.IP.To4(); ip4 != nil {
                    // Derive /24 from the IP
                    return fmt.Sprintf("%d.%d.%d.0/24", ip4[0], ip4[1], ip4[2])
                }
            }
        }
    }
    return "unknown"
}
```

### Engine isAdmin Field Wiring

```go
// In engine.go Start() — after resolveProfiles(), before go e.run()
e.isAdmin = checkIsElevated()
simlog.Info(fmt.Sprintf("Elevation check: isAdmin=%v (techniques with elevation_required will %s)",
    e.isAdmin, map[bool]string{true: "run", false: "be skipped"}[e.isAdmin]))
```

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (`testing` package), `go test ./...` |
| Config file | none — standard Go test discovery |
| Quick run command | `go test ./internal/engine/ ./internal/executor/ -run TestAMSI -timeout 30s` |
| Full suite command | `go test ./... -timeout 120s` |

### Phase Requirements to Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| INFRA-01 | AMSI detection classifies blocked technique as `amsi_blocked` status, not `fail` | unit | `go test ./internal/executor/ -run TestAMSI -v` | No — Wave 0 gap |
| INFRA-01 | AMSI detection only fires for PowerShell executor, not CMD or Go | unit | `go test ./internal/executor/ -run TestAMSI_ExecutorTypes -v` | No — Wave 0 gap |
| INFRA-02 | Engine skips elevation-required technique when isAdmin=false | unit | `go test ./internal/engine/ -run TestElevationSkip -v` | No — Wave 0 gap |
| INFRA-02 | Engine runs elevation-required technique when isAdmin=true | unit | `go test ./internal/engine/ -run TestElevationRun -v` | No — Wave 0 gap |
| INFRA-03 | Scan confirmation channel blocks engine until ConfirmScan() is called | unit | `go test ./internal/engine/ -run TestScanConfirm -v` | No — Wave 0 gap |
| INFRA-03 | `/api/scan/confirm` endpoint triggers engine ConfirmScan | unit | `go test ./internal/server/ -run TestScanConfirmAPI -v` | No — Wave 0 gap |
| INFRA-03 | `/api/scan/pending` returns ScanInfo when confirmation is pending | unit | `go test ./internal/server/ -run TestScanPending -v` | No — Wave 0 gap |

**Manual-only tests (no automation possible):**
- INFRA-01: Real AMSI block on Windows with Defender enabled — requires running a flagged script against actual Defender. Cannot mock in unit test without disabling AMSI system-wide. Integration test only.
- INFRA-02: Real `IsElevated()` check — verify by running the binary as admin and as non-admin.
- INFRA-03: Modal rendering and consultant UX — browser interaction test only.

### Sampling Rate

- Per task commit: `go test ./internal/engine/ ./internal/executor/ -timeout 30s`
- Per wave merge: `go test ./... -timeout 120s`
- Phase gate: Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/executor/executor_amsi_test.go` — covers INFRA-01 (unit tests with fake stderr)
- [ ] `internal/engine/engine_elevation_test.go` — covers INFRA-02 (engine with injected `isAdmin` flag)
- [ ] `internal/engine/engine_scan_confirm_test.go` — covers INFRA-03 (channel-based confirmation)
- [ ] `internal/server/server_scan_test.go` — covers INFRA-03 API endpoints

**Note on testability:** The existing `RunnerFunc` injection pattern in engine.go enables testing elevation skip without executing real techniques. The `isAdmin bool` field must be exported or exposed via a test-only setter to allow injection in tests (same as `SetRunner(fn)`).

## Environment Availability

Phase 16 is code/config changes only. No new external tools are needed. The Windows token check requires the binary to run on Windows — this is already the only supported platform.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `golang.org/x/sys/windows` | INFRA-02 elevation check | Already in go.sum | v0.41.0 | None needed — already present |
| Windows Defender (AMSI) | INFRA-01 real integration test | Not testable in unit tests | — | Unit tests use fake stderr instead |
| Go 1.26.1 | Build | Assumed from go.mod | 1.26.1 | — |

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| All failures classified as `fail` | `amsi_blocked`, `elevation_required` as distinct statuses | Phase 16 | Consultant can distinguish real detection gaps from environmental blocks |
| Elevation-required techniques fail with "Access Denied" in ErrorOutput | Engine skips with clear status before attempting execution | Phase 16 | No confusing error output for non-admin runs |
| T1046 runs immediately on scan start | Consultant confirms subnet + IDS warning before scan | Phase 16 | Prevents accidental scan of wrong subnet on client network |

## Open Questions

1. **AMSI exit code -196608 on all Windows/PowerShell versions**
   - What we know: CONTEXT.md specifies this value; it is referenced in PowerShell source as the AMSI-triggered parse exception path.
   - What's unclear: Whether this code is consistent across PowerShell 5.1 (Windows built-in) vs. PowerShell 7.x. The AMSI integration point differs between versions.
   - Recommendation: Match on both the string pattern AND the exit code. String matching is more reliable across versions; exit code is a secondary check. If only the exit code fires (no matching string), flag it for investigation rather than silently classifying as AMSI.

2. **Scan confirmation placement — run() vs. runPoC()**
   - What we know: The engine has two execution paths: `run()` for immediate mode and `runPoC()` for multi-day PoC mode.
   - What's unclear: Whether scan confirmation should also apply in PoC mode (where techniques are split across days).
   - Recommendation: Apply confirmation in both paths. T1046 is a discovery phase technique and will appear in Phase 1 of PoC mode. Add the pre-flight confirmation check at the start of the discovery loop in both `run()` and `runPoC()`.

3. **`requires_confirmation` bool vs. tag entry**
   - What we know: CONTEXT.md leaves this to Claude's discretion. Both approaches work.
   - What's unclear: Whether future techniques may need confirmation for different reasons (e.g., destructive techniques in Phase 18).
   - Recommendation: Use a `RequiresConfirmation bool` YAML field. Future-proofs for any technique type. Currently apply it only to T1046 in its YAML. The engine gate checks the field, not a hardcoded technique ID.

## Sources

### Primary (HIGH confidence)

- `internal/executor/executor.go` — Full source read; `runInternal` structure, insertion point for AMSI detection confirmed
- `internal/engine/engine.go` — Full structure read; `runTechnique` at line 610, `isAdmin` field placement confirmed, `stopCh` channel pattern confirmed for `scanConfirmCh` design
- `internal/playbooks/types.go` — Full source read; `VerificationStatus` enum pattern, `Technique` struct fields confirmed
- `internal/reporter/reporter.go` — Full source read; `htmlTemplate` verifStr switch structure, existing badge CSS confirmed
- `internal/server/server.go` — Full source read; `registerRoutes` pattern, `authMiddleware`, `writeJSON`/`writeError` helpers confirmed
- `internal/server/static/index.html` — Partial read; CSS variable system (`--orange`, `--accent`, etc.), existing `.alert-warn` modal pattern confirmed
- `go.mod` — `golang.org/x/sys v0.41.0` already indirect dep confirmed; no new deps needed
- PowerShell/PowerShell GitHub source — `ScriptContainedMaliciousContent` error identifier confirmed in CompiledScriptBlock.cs

### Secondary (MEDIUM confidence)

- [pkg.go.dev/golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows) — `GetCurrentProcessToken().IsElevated()` API confirmed
- [Microsoft Learn — AMSI demonstrations](https://learn.microsoft.com/en-us/defender-endpoint/mde-demonstration-amsi) — AMSI block behavior confirmed
- [golang/go issue #28804](https://github.com/golang/go/issues/28804) — `IsMember()` vs `IsElevated()` distinction confirmed

### Tertiary (LOW confidence)

- Exit code `-196608` for AMSI block: mentioned in CONTEXT.md as a known value; cross-referenced with PowerShell source indirectly. Exact value should be validated during implementation by triggering a real AMSI block.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new deps; existing packages verified in go.mod
- Architecture patterns: HIGH — all insertion points read from source; no assumptions
- AMSI string patterns: HIGH for string matching; MEDIUM for exit code (verify during impl)
- Windows token check: HIGH — `GetCurrentProcessToken().IsElevated()` is documented official API
- Pitfalls: HIGH — derived from reading actual code flow in engine.go and executor.go

**Research date:** 2026-04-09
**Valid until:** 2026-07-09 (stable — Windows API and Go stdlib patterns do not change frequently)
