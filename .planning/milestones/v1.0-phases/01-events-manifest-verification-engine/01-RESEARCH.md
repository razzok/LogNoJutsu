# Phase 1: Events Manifest & Verification Engine — Research

**Researched:** 2026-03-24
**Domain:** Go — Windows Event Log querying, struct design, HTML template extension
**Confidence:** HIGH (all findings based on direct source code reading; no external docs required for core implementation)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Replace `ExpectedEvents []string` with a structured `EventSpec` type containing at minimum: `event_id` (int), `channel` (string), `description` (string). This is the canonical format for both the YAML technique files and querying.
- **D-02:** Whether to add optional keyword filter criteria (e.g., `contains` field for message matching) is left to Claude's discretion — balance precision vs complexity.
- **D-03:** All 39+ existing technique YAML files must be updated to use the new `EventSpec` format, replacing the existing free-text strings.
- **D-04:** Verification results appear as an **inline column in the existing per-technique results table** in the HTML report (not a separate section).
- **D-05:** The verification column shows: a pass/fail badge (green/red) + a compact list of each expected event ID with its individual status (e.g., ✓ EID 10 Sysmon/Operational, ✗ EID 4656 Security).
- **D-06:** Verification runs **after each technique**, immediately following execution, with a **configurable wait delay** (default: 3 seconds) to allow Windows Event Log writes to settle before querying.
- **D-07:** Verification result is stored in the `ExecutionResult` for that technique and appears in real-time in the simulation logs.
- **D-08:** Use the existing `ExecutionResult.Success` flag as the gate: if `Success=false` → mark as **"Not Executed"** (tool-side failure); if `Success=true` but no matching events found → mark as **"Events Missing"** (SIEM-side gap). No additional execution detection logic needed.

### Claude's Discretion

- Query mechanism for Windows Event Log (PowerShell `Get-WinEvent`, Go Win32 API via `golang.org/x/sys`, or subprocess — choose based on existing executor patterns)
- Whether `EventSpec` includes an optional `contains` keyword filter for message content matching
- Time window for event log search (how far back to look — recommend 30-60s from technique start)
- New fields to add to `ExecutionResult` to carry verification status
- WhatIf mode behavior (skip verification since techniques don't execute)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| VERIF-01 | Each technique declares the Windows Event IDs / log sources it is expected to generate | D-01: `EventSpec` struct replaces `ExpectedEvents []string`; all 43 technique YAMLs mapped below |
| VERIF-02 | After simulation run, tool queries local Windows Event Log for expected events | PowerShell `Get-WinEvent` via subprocess is the recommended query mechanism |
| VERIF-03 | Each technique reports pass (events found) or fail (events missing) in the run results | New `VerificationStatus` typed enum + `VerifiedEvents []VerifiedEvent` added to `ExecutionResult` |
| VERIF-04 | Verification results are included in the HTML report (expected vs. observed, pass/fail per technique) | New `Verification` column added to existing results table in `htmlTemplate`; `htmlData.Results` already passes `ExecutionResult` slices directly |
| VERIF-05 | Verification distinguishes "technique did not execute" from "technique executed but event not found" | `ExecutionResult.Success=false` → `VerifNotExecuted`; `Success=true` + empty found events → `VerifFail` |
</phase_requirements>

---

## Summary

This phase adds two things: (1) a richer data model for declaring expected events per technique, and (2) a post-execution verification engine that queries the local Windows Event Log and records pass/fail. The codebase is already well-structured for this addition — the key insertion points are `internal/playbooks/types.go` (struct changes), `internal/engine/engine.go` (`runTechnique` method), and `internal/reporter/reporter.go` (HTML template).

The query mechanism decision is the most significant design choice. The codebase uses PowerShell subprocess invocation as its established pattern (`runCommand` in `executor.go`). Using PowerShell `Get-WinEvent` via the same subprocess mechanism avoids introducing any new dependencies and matches the project's zero-external-deps constraint. This is the recommended approach.

The 43 technique YAML files each have free-text `expected_events` strings. Each must be converted to a list of `EventSpec` structs with integer `event_id`, `channel`, and `description`. The description field preserves the existing human-readable text. A complete mapping of every technique to its structured EventSpec entries is documented below.

**Primary recommendation:** Use PowerShell `Get-WinEvent` via subprocess (matching `executor.runCommand` pattern), with a 30-second lookback window anchored to `result.StartTime`, and add `VerificationStatus` + `VerifiedEvents` to `ExecutionResult`.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go standard library | (project Go version) | `exec.Command`, `time`, `strings`, `fmt` | No external deps needed for subprocess-based querying |
| `gopkg.in/yaml.v3` | v3.0.1 (already in go.mod) | Parsing `EventSpec` from technique YAML files | Already the project's YAML parser |

### No New Dependencies
This phase introduces zero new Go module dependencies. All capabilities needed — subprocess execution, JSON marshaling, HTML templating, struct definitions — are already in the project's standard library usage.

**Installation:** No new packages to install.

---

## Architecture Patterns

### Recommended Project Structure — Files to Change

```
internal/
├── playbooks/
│   └── types.go             # ADD: EventSpec, VerificationStatus, VerifiedEvent types
│                            # CHANGE: Technique.ExpectedEvents []string → []EventSpec
│                            # CHANGE: ExecutionResult — add VerificationStatus, VerifiedEvents, VerifyTime
├── engine/
│   └── engine.go            # CHANGE: runTechnique() — call verifier after executor
│                            # CHANGE: engine.Config — add VerificationWaitSecs int
├── reporter/
│   └── reporter.go          # CHANGE: htmlTemplate — add Verification column
│                            # CHANGE: htmlData — already uses []ExecutionResult (no struct change needed)
└── verifier/
    └── verifier.go          # ADD: new package — QueryEvents(t Technique, since time.Time) VerificationResult
                             # (or inline in engine package as engine/verify.go)
```

### Pattern 1: New Types in `internal/playbooks/types.go`

**What:** Replace `ExpectedEvents []string` with `[]EventSpec` and enrich `ExecutionResult`.

**Full struct definitions (follow existing visual-alignment tag convention):**

```go
// EventSpec declares a single Windows event that a technique is expected to generate.
type EventSpec struct {
    EventID     int    `yaml:"event_id"    json:"event_id"`
    Channel     string `yaml:"channel"     json:"channel"`
    Description string `yaml:"description" json:"description"`
    Contains    string `yaml:"contains,omitempty" json:"contains,omitempty"` // optional keyword filter
}

// VerificationStatus indicates the result of post-execution event log verification.
type VerificationStatus string

const (
    VerifNotRun      VerificationStatus = "not_run"       // verification did not run (WhatIf mode)
    VerifPass        VerificationStatus = "pass"           // all expected events found
    VerifFail        VerificationStatus = "fail"           // executed but events missing
    VerifNotExecuted VerificationStatus = "not_executed"   // technique itself failed to run
)

// VerifiedEvent records the per-EventSpec outcome after querying the event log.
type VerifiedEvent struct {
    EventID     int    `json:"event_id"`
    Channel     string `json:"channel"`
    Description string `json:"description"`
    Found       bool   `json:"found"`
    Count       int    `json:"count"` // number of matching events found in the time window
}
```

**Changes to existing structs:**

```go
// In Technique — replace line 14:
// BEFORE: ExpectedEvents []string          `yaml:"expected_events"    json:"expected_events"`
// AFTER:
ExpectedEvents []EventSpec `yaml:"expected_events" json:"expected_events"`

// In ExecutionResult — add after CleanupRun:
VerificationStatus VerificationStatus `json:"verification_status,omitempty"`
VerifiedEvents     []VerifiedEvent    `json:"verified_events,omitempty"`
VerifyTime         string             `json:"verify_time,omitempty"` // RFC3339 timestamp of verification
```

### Pattern 2: Verification Package `internal/verifier/verifier.go`

**What:** New package with a single exported function. Placing in its own package avoids circular imports and keeps engine.go focused on scheduling.

**Dependency position in the existing graph:**
```
engine → verifier → playbooks (OK — no cycles)
```

**Core function signature:**

```go
package verifier

// Verify queries the Windows Event Log for each EventSpec in the technique,
// using a time window from 'since' to now. Returns a populated ExecutionResult
// with VerificationStatus and VerifiedEvents filled in.
func Verify(specs []playbooks.EventSpec, since time.Time) (playbooks.VerificationStatus, []playbooks.VerifiedEvent)
```

**PowerShell query implementation pattern:**

The query uses `Get-WinEvent` with `-FilterHashtable`. This is the correct modern API (not `Get-EventLog` which is legacy and missing many channels).

```powershell
# Template query — parameterized per EventSpec
Get-WinEvent -FilterHashtable @{
    LogName   = 'Microsoft-Windows-Sysmon/Operational'
    Id        = 10
    StartTime = '2026-03-24T14:00:00'
} -ErrorAction SilentlyContinue | Measure-Object | Select-Object -ExpandProperty Count
```

**Go implementation using `runCommand` pattern (matching executor.go):**

```go
func queryEventCount(channel string, eventID int, since time.Time) (int, error) {
    sinceStr := since.Format(time.RFC3339)
    // Build PowerShell one-liner — outputs a single integer count
    script := fmt.Sprintf(
        `(Get-WinEvent -FilterHashtable @{LogName='%s'; Id=%d; StartTime='%s'} -ErrorAction SilentlyContinue | Measure-Object).Count`,
        channel, eventID, sinceStr,
    )
    cmd := exec.Command("powershell.exe",
        "-NonInteractive", "-NoProfile",
        "-ExecutionPolicy", "Bypass",
        "-Command", script,
    )
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    if err := cmd.Run(); err != nil {
        return 0, nil // treat query error as zero events (channel may not exist/be enabled)
    }
    count, _ := strconv.Atoi(strings.TrimSpace(out.String()))
    return count, nil
}
```

**Note on error handling:** If a channel does not exist (e.g., Sysmon not installed) `Get-WinEvent` returns a non-zero exit code. Treat this as `count=0` rather than a hard error — the technique simply fails verification. This avoids crashing verification when running without Sysmon.

### Pattern 3: Injection Point in `engine.go` `runTechnique()`

**Current code (lines 420-466):**

```go
func (e *Engine) runTechnique(t *playbooks.Technique) {
    // ... picks user, sets CurrentStep ...
    var result playbooks.ExecutionResult
    if e.cfg.WhatIf {
        // ... builds WhatIf result ...
    } else if e.cfg.RunCleanup {
        result = executor.RunWithCleanup(t, profile, password)
    } else {
        result = executor.RunAs(t, profile, password)
        // ...
    }
    // ← INJECT VERIFICATION HERE, before appending result
    e.mu.Lock()
    e.status.Results = append(e.status.Results, result)
    // ...
}
```

**After injection pattern:**

```go
    // Post-execution verification (skip in WhatIf mode)
    if !e.cfg.WhatIf && len(t.ExpectedEvents) > 0 {
        waitSecs := e.cfg.VerificationWaitSecs
        if waitSecs <= 0 {
            waitSecs = 3 // default
        }
        simlog.Info(fmt.Sprintf("[Verify] %s — waiting %ds for event log writes...", t.ID, waitSecs))
        time.Sleep(time.Duration(waitSecs) * time.Second)
        status, verified := verifier.Verify(t.ExpectedEvents, result.StartTimeParsed())
        result.VerificationStatus = status
        result.VerifiedEvents = verified
        result.VerifyTime = time.Now().Format(time.RFC3339)
        simlog.Info(fmt.Sprintf("[Verify] %s — %s (%d/%d events found)",
            t.ID, status, countFound(verified), len(verified)))
    } else if e.cfg.WhatIf {
        result.VerificationStatus = playbooks.VerifNotRun
    } else {
        // No expected events declared — mark as not_run
        result.VerificationStatus = playbooks.VerifNotRun
    }
```

**Helper needed:** `result.StartTimeParsed()` — a method on `ExecutionResult` that parses `StartTime` (RFC3339 string) into `time.Time`. Alternatively pass `start time.Time` from `runInternal` through to `runTechnique`.

**Simpler alternative:** Capture `start := time.Now()` in `runTechnique` before calling executor, then pass `start` directly to `verifier.Verify`. This avoids string-parsing.

**`VerificationWaitSecs` addition to `engine.Config`:**

```go
// In engine.Config struct — add after existing execution options:
VerificationWaitSecs int `json:"verification_wait_secs"` // 0 = use default (3s)
```

### Pattern 4: HTML Report Template Changes

**Current table headers (line 267-275 of reporter.go):**

```html
<tr>
  <th>Zeit</th>
  <th>Technik-ID</th>
  <th>Name</th>
  <th>Taktik</th>
  <th>Status</th>
  <th>Benutzer</th>
</tr>
```

**After adding verification column:**

```html
<tr>
  <th>Zeit</th>
  <th>Technik-ID</th>
  <th>Name</th>
  <th>Taktik</th>
  <th>Status</th>
  <th>Verifikation</th>
  <th>Benutzer</th>
</tr>
```

**New CSS classes needed in the `<style>` block:**

```css
.verif-pass{color:#3fb950;font-weight:600}
.verif-fail{color:#f85149;font-weight:600}
.verif-skip{color:#8b949e}
.verif-list{font-size:11px;margin-top:4px;padding-left:0;list-style:none}
.verif-list li{margin:1px 0}
```

**New table cell template (inserted before `<td style="color:#bc8cff...">{{.RunAsUser}}</td>`):**

```html
<td>
  {{if eq .VerificationStatus "pass"}}
    <span class="verif-pass">✓ Pass</span>
  {{else if eq .VerificationStatus "fail"}}
    <span class="verif-fail">✗ Fail</span>
  {{else if eq .VerificationStatus "not_executed"}}
    <span class="verif-skip">— Nicht ausgeführt</span>
  {{else}}
    <span class="verif-skip">—</span>
  {{end}}
  {{if .VerifiedEvents}}
  <ul class="verif-list">
    {{range .VerifiedEvents}}
    <li>{{if .Found}}✓{{else}}✗{{end}} EID {{.EventID}} <span style="color:#8b949e">{{.Channel}}</span></li>
    {{end}}
  </ul>
  {{end}}
</td>
```

**Note on template function map:** The `eq` comparison function is built into Go's `text/template` — no new template functions needed.

### Anti-Patterns to Avoid

- **Do not use `Get-EventLog`:** It is a legacy PowerShell cmdlet that does not support Sysmon, Security Audit, or modern Windows log channels. Only `Get-WinEvent` supports all channels.
- **Do not query without a time window:** Querying the entire event log history is prohibitively slow on production systems. Always filter by `StartTime`.
- **Do not use `golang.org/x/sys` Windows API for event log queries:** It would add a new module dependency, is significantly more complex than a PowerShell subprocess call, and provides no reliability advantage in this use case. The project explicitly avoids runtime dependencies.
- **Do not add a separate verification section to the HTML report:** Decision D-04 locks the inline column approach.
- **Do not change `SaveResults` function signature:** `reporter.SaveResults(results []playbooks.ExecutionResult, logFile string, whatIf bool)` is called from both `e.abort()` and `e.finish()`. Keep this interface stable — verification data travels inside `ExecutionResult`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Windows Event Log querying | Custom Win32 API via `golang.org/x/sys` | PowerShell `Get-WinEvent` via subprocess | Project has no runtime deps; `runCommand` pattern already established; `Get-WinEvent` handles all channels including Sysmon |
| Time windowing | Custom timestamp comparison logic | `Get-WinEvent -FilterHashtable @{StartTime=...}` | Windows handles log channel iteration; StartTime filter is built into the API |
| Channel name lookup | Hardcoded mapping table in Go | `channel` field in `EventSpec` (from YAML) | Channel names are technique-specific; YAML is the right place to declare them |

---

## Complete Event ID Mapping — All 43 Techniques

This is the authoritative source for populating `expected_events` in every YAML file. For each technique, the structured `EventSpec` entries are derived directly from reading the YAML files.

### Windows Event Log Channel Reference

| Channel Name (YAML value) | Description | Requires |
|--------------------------|-------------|----------|
| `Security` | Windows Security audit log — 4624, 4625, 4688, 4697, 4698, 4720, etc. | Process creation auditing via auditpol |
| `System` | Windows System log — 7045, 104, 1102 | Default |
| `Microsoft-Windows-Sysmon/Operational` | Sysmon events — EID 1, 3, 7, 10, 11, 13, 19, 20, 21, 22, 23 | Sysmon installed |
| `Microsoft-Windows-PowerShell/Operational` | PowerShell events — 4104, 4103 | ScriptBlock logging enabled via preparation |
| `Microsoft-Windows-WMI-Activity/Operational` | WMI provider events — 5857, 5858, 5860, 5861 | Default (Windows 7+) |
| `Microsoft-Windows-Bits-Client/Operational` | BITS transfer events — 59, 60 | Default |
| `Microsoft-Windows-Windows Defender/Operational` | Defender behavioral detections — 1116, 1117 | Windows Defender running |

### Technique-to-EventSpec Mapping

The following shows the structured `expected_events` block each YAML needs after conversion. `event_id: 0` is used as a sentinel for Sysmon events documented by number (e.g., Sysmon 1 = EID 1 on the Sysmon/Operational channel).

**T1003.001 — LSASS Memory Access**
```yaml
expected_events:
  - event_id: 10
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "ProcessAccess - TargetImage: lsass.exe, GrantedAccess: 0x1010 or 0x1410"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "rundll32.exe - CommandLine contains comsvcs.dll MiniDump"
  - event_id: 11
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "FileCreate - .dmp file in TEMP"
  - event_id: 4656
    channel: "Security"
    description: "Handle request to lsass.exe with PROCESS_VM_READ"
  - event_id: 4663
    channel: "Security"
    description: "Object access - lsass.exe"
```

**T1003.006 — DCSync**
```yaml
expected_events:
  - event_id: 4662
    channel: "Security"
    description: "DS object access - replication GUIDs 1131f6aa, 1131f6ad in Properties field"
  - event_id: 4928
    channel: "Security"
    description: "AD replication NC events via repadmin"
  - event_id: 4688
    channel: "Security"
    description: "nltest.exe, dsquery.exe, repadmin.exe"
  - event_id: 4769
    channel: "Security"
    description: "Kerberos TGS for DRSUAPI"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - LDAP replication attribute queries"
```

**T1016 — Network Config Discovery**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "ipconfig.exe, arp.exe, route.exe, netsh.exe, nbtstat.exe, nslookup.exe — rapid succession burst"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "ipconfig.exe with /displaydns — DNS cache dump"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Get-NetIPConfiguration, Get-DnsClientCache"
```

**T1021.001 — RDP**
```yaml
expected_events:
  - event_id: 4648
    channel: "Security"
    description: "Logon with explicit credentials — cmdkey RDP credential storage"
  - event_id: 4946
    channel: "Security"
    description: "Windows Firewall rule added for port 3389"
  - event_id: 13
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "RegistryValueSet - fDenyTSConnections + PortNumber modification"
```

**T1021.002 — SMB Shares**
```yaml
expected_events:
  - event_id: 5140
    channel: "Security"
    description: "Network Share Object Accessed — admin share enumeration"
  - event_id: 5145
    channel: "Security"
    description: "Network Share Object Checked — detailed share access"
  - event_id: 4688
    channel: "Security"
    description: "net.exe share — CommandLine contains 'share'"
  - event_id: 3
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "NetworkConnect — port 445 SMB"
```

**T1027 — Obfuscated Commands**
```yaml
expected_events:
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - decoded obfuscated content logged after deobfuscation"
  - event_id: 4103
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell Module Logging - IEX calls visible"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "powershell.exe with -EncodedCommand -NoP -NonI -W Hidden -Exec Bypass"
  - event_id: 13
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "RegistryValueSet - HKCU payload storage"
```

**T1036.005 — Masquerading**
```yaml
expected_events:
  - event_id: 11
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "FileCreate — binary copied to non-standard location"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "Image path = TEMP\\svchost.exe — path mismatch tier-1 alert"
  - event_id: 4688
    channel: "Security"
    description: "Process creation with image path not matching expected system location"
  - event_id: 13
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "RegistryValueSet — masqueraded binary in Run key"
```

**T1041 — Exfiltration HTTP**
```yaml
expected_events:
  - event_id: 3
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "NetworkConnect — outbound HTTP from PowerShell to external host"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock — Invoke-WebRequest, base64 encoding, ConvertTo-Json"
  - event_id: 22
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "DNSEvent — DNS query for exfil target"
```

**T1046 — Network Scan**
```yaml
expected_events:
  - event_id: 3
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "NetworkConnect - burst of simultaneous connections from powershell.exe PID"
  - event_id: 5156
    channel: "Security"
    description: "WFP allowed connections - rapid burst pattern"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - runspace port scan"
```

**T1047 — WMI Execution**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "wmic.exe with 'process call create' CommandLine"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "WmiPrvSE.exe parent spawning cmd.exe child — key detection signal"
  - event_id: 5861
    channel: "Microsoft-Windows-WMI-Activity/Operational"
    description: "WMI provider activity — permanent subscription or ad-hoc query"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Invoke-WmiMethod / Invoke-CimMethod"
```

**T1049 — Network Connections**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "netstat.exe, net.exe, net1.exe — multiple process creation events"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "netstat.exe with -ano flag"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Get-NetTCPConnection with process correlation"
```

**T1053.005 — Scheduled Task**
```yaml
expected_events:
  - event_id: 4698
    channel: "Security"
    description: "Scheduled task created — task XML with action/trigger/principal in event data"
  - event_id: 4688
    channel: "Security"
    description: "schtasks.exe with /create and full arguments visible in CommandLine"
  - event_id: 4702
    channel: "Security"
    description: "Scheduled task updated — when task is modified"
```

**T1057 — Process Discovery**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "tasklist.exe with /v and /svc flags"
  - event_id: 4688
    channel: "Security"
    description: "wmic.exe 'process get commandline' — CommandLine field is highly suspicious"
    contains: "commandline"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Get-WmiObject Win32_Process CommandLine query"
```

**T1059.001 — PowerShell**
```yaml
expected_events:
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock — decoded script content logged post-deobfuscation"
  - event_id: 4103
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell Module Logging — IEX / Invoke-Expression calls"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "powershell.exe with -NoP -NonI -W Hidden -Exec Bypass in CommandLine"
  - event_id: 3
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "NetworkConnect from powershell.exe — download cradle attempt"
```

**T1059.003 — cmd.exe**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "cmd.exe with redirection operators, chained & commands in CommandLine"
  - event_id: 11
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "FileCreate — output redirect files"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "powershell.exe child of cmd.exe — high-signal parent-child chain"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock when cmd spawns PS"
```

**T1069 — Group Discovery**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "net.exe 'localgroup Administrators' — standalone tier-1 SIEM rule"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Get-LocalGroup / Get-LocalGroupMember"
```

**T1070.001 — Clear Logs**
```yaml
expected_events:
  - event_id: 1102
    channel: "Security"
    description: "Audit log cleared — generated BEFORE clear, always survives"
  - event_id: 104
    channel: "System"
    description: "Event log cleared — for non-Security logs"
  - event_id: 4688
    channel: "Security"
    description: "wevtutil.exe with 'cl' subcommand"
```

**T1082 — System Info Discovery**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "systeminfo.exe, wmic.exe, reg.exe, hostname.exe, whoami.exe — burst"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "wmic.exe with /format:csv — common attacker output pattern"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Get-CimInstance queries"
```

**T1083 — File Discovery**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "cmd.exe / tree.com — dir /s /b on user profile directories"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Get-ChildItem recursive on sensitive paths"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "tree.com process creation"
```

**T1087 — Account Discovery**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "net.exe with 'localgroup Administrators' — standalone SIEM rule"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Get-LocalUser / Get-LocalGroupMember"
```

**T1098 — Account Manipulation**
```yaml
expected_events:
  - event_id: 4720
    channel: "Security"
    description: "User account created"
  - event_id: 4732
    channel: "Security"
    description: "Member added to security-enabled local group — Administrators"
  - event_id: 4738
    channel: "Security"
    description: "User account changed — PasswordNeverExpires, flags modified"
  - event_id: 4688
    channel: "Security"
    description: "net.exe with 'localgroup Administrators ... /add'"
```

**T1110.001 — Bruteforce**
```yaml
expected_events:
  - event_id: 4625
    channel: "Security"
    description: "Failed logon - LogonType=2 via LogonUser API, SubStatus=0xC000006A wrong password"
  - event_id: 4740
    channel: "Security"
    description: "Account lockout — only if lockout policy triggers"
  - event_id: 4688
    channel: "Security"
    description: "net.exe with use \\IPC$ arguments"
```

**T1110.003 — Password Spraying**
```yaml
expected_events:
  - event_id: 4625
    channel: "Security"
    description: "Failed logon - same SubStatus=0xC000006A across many different TargetUserName values"
  - event_id: 4771
    channel: "Security"
    description: "Kerberos pre-auth failed — in AD environments, one per user"
  - event_id: 4688
    channel: "Security"
    description: "powershell.exe process creation for spray script"
```

**T1134.001 — Token Impersonation**
```yaml
expected_events:
  - event_id: 4673
    channel: "Security"
    description: "Sensitive Privilege Use — SeDebugPrivilege, SeImpersonatePrivilege"
  - event_id: 4672
    channel: "Security"
    description: "Special Logon — elevated privilege token assigned"
  - event_id: 4688
    channel: "Security"
    description: "whoami.exe /priv — CommandLine contains '/priv'"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock — WindowsIdentity, WindowsPrincipal"
```

**T1135 — Network Share Discovery**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "net.exe with view/share arguments"
  - event_id: 5140
    channel: "Security"
    description: "Network share accessed — admin share access"
  - event_id: 5145
    channel: "Security"
    description: "Network share object checked — file share audit"
  - event_id: 3
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "NetworkConnect to SMB port 445"
```

**T1136.001 — Create Local Account**
```yaml
expected_events:
  - event_id: 4720
    channel: "Security"
    description: "User account created — TargetUserName, SubjectUserName fields"
  - event_id: 4722
    channel: "Security"
    description: "Account enabled — automatically fires after creation"
  - event_id: 4732
    channel: "Security"
    description: "Member added to security-enabled local group — added to Administrators"
  - event_id: 4688
    channel: "Security"
    description: "net.exe with 'user /add' in CommandLine"
```

**T1197 — BITS Jobs**
```yaml
expected_events:
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "bitsadmin.exe with /setnotifycmdline argument — high-signal indicator"
  - event_id: 59
    channel: "Microsoft-Windows-Bits-Client/Operational"
    description: "BITS transfer job started — URL + destination logged"
  - event_id: 60
    channel: "Microsoft-Windows-Bits-Client/Operational"
    description: "BITS transfer stopped"
  - event_id: 3
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "NetworkConnect from svchost.exe — BITS service reaching out"
```

**T1218.011 — Rundll32**
```yaml
expected_events:
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "rundll32.exe with DLL+entrypoint in CommandLine"
  - event_id: 7
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "ImageLoad — DLL loaded by rundll32: url.dll, pcwutl.dll, advpack.dll"
  - event_id: 3
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "NetworkConnect — if rundll32 makes HTTP request"
```

**T1482 — Domain Trust Discovery**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "nltest.exe /domain_trusts — standalone tier-1 detection rule"
  - event_id: 4662
    channel: "Security"
    description: "LDAP query accessing trusted domain objects"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - .NET Domain/Forest trust enumeration"
```

**T1486 — Data Encrypted for Impact**
```yaml
expected_events:
  - event_id: 11
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "FileCreate - mass .locked file creation burst"
  - event_id: 23
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "FileDelete - original files deleted after encryption"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - RNGCryptoServiceProvider, CreateEncryptor, bulk file ops"
  - event_id: 1116
    channel: "Microsoft-Windows-Windows Defender/Operational"
    description: "Windows Defender behavioral ransomware detection may trigger"
```

**T1490 — Inhibit Recovery**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "vssadmin.exe 'delete shadows /all /quiet' — standalone tier-1 SIEM alert"
  - event_id: 4688
    channel: "Security"
    description: "bcdedit.exe /set recoveryenabled no"
    contains: "bcdedit"
  - event_id: 13
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "RegistryValueSet - SystemRestore disable keys"
```

**T1543.003 — New Service**
```yaml
expected_events:
  - event_id: 7045
    channel: "System"
    description: "New service installed — ServiceName, ServiceFileName, ServiceAccount"
  - event_id: 4697
    channel: "Security"
    description: "Service installed — Security log duplicate of EID 7045"
  - event_id: 4688
    channel: "Security"
    description: "sc.exe with 'create' and service parameters in CommandLine"
  - event_id: 13
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "RegistryValueSet - HKLM\\SYSTEM\\CurrentControlSet\\Services direct creation"
```

**T1546.003 — WMI Event Subscription**
```yaml
expected_events:
  - event_id: 19
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "WmiEventFilter — filter name and WQL query logged"
  - event_id: 20
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "WmiEventConsumer — consumer type CommandLineEventConsumer + command"
  - event_id: 21
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "WmiEventConsumerToFilter — binding links filter to consumer"
  - event_id: 5861
    channel: "Microsoft-Windows-WMI-Activity/Operational"
    description: "Permanent WMI subscription registered"
```

**T1547.001 — Registry Persistence**
```yaml
expected_events:
  - event_id: 13
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "RegistryValueSet — key path + value data showing powershell.exe -WindowStyle Hidden"
  - event_id: 4657
    channel: "Security"
    description: "Registry value modified — requires Object Access auditing on registry"
```

**T1548.002 — UAC Bypass**
```yaml
expected_events:
  - event_id: 13
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "RegistryValueSet - HKCU\\Software\\Classes\\ms-settings\\Shell\\Open\\command"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "fodhelper.exe / eventvwr.msc spawning cmd.exe with HIGH integrity level"
  - event_id: 4688
    channel: "Security"
    description: "Child process of fodhelper.exe — elevated token without UAC prompt"
  - event_id: 4673
    channel: "Security"
    description: "Privileged service called — UAC elevation event"
```

**T1550.002 — Pass the Hash**
```yaml
expected_events:
  - event_id: 4648
    channel: "Security"
    description: "Logon with explicit credentials — primary PtH Exabeam detection signal"
  - event_id: 4624
    channel: "Security"
    description: "LogonType=3 Network, AuthenticationPackageName=NTLM — PtH network logon"
  - event_id: 4625
    channel: "Security"
    description: "Failed network logon — when wrong credentials used in net use attempt"
```

**T1552.001 — Credentials in Files**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "findstr.exe with /si password — /si flag is the detection signal"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Select-String pattern matching password/passwd/pwd"
```

**T1558.003 — Kerberoasting**
```yaml
expected_events:
  - event_id: 4769
    channel: "Security"
    description: "Kerberos TGS request - TicketEncryptionType=0x17 RC4 = Kerberoasting indicator"
  - event_id: 4688
    channel: "Security"
    description: "setspn.exe with -T -Q */* arguments"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - KerberosRequestorSecurityToken / LDAP SPN enumeration"
```

**T1562.002 — Disable Logging**
```yaml
expected_events:
  - event_id: 4719
    channel: "Security"
    description: "System audit policy changed — fires when auditpol disables a category"
  - event_id: 4688
    channel: "Security"
    description: "auditpol.exe with /set /category ... /success:disable"
  - event_id: 13
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "RegistryValueSet — EventLog MaxSize reduction"
```

**T1574.002 — DLL Sideloading**
```yaml
expected_events:
  - event_id: 7
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "ImageLoad — DLL loaded from non-standard path outside System32"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock — Add-Type, Assembly compilation"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "Process creation with unusual DLL in module list"
```

**UEBA_credential_spray_chain**
```yaml
expected_events:
  - event_id: 4625
    channel: "Security"
    description: "Failed logon x25 — rapid succession (Exabeam brute force use case trigger)"
  - event_id: 4624
    channel: "Security"
    description: "Successful logon after failures (Exabeam: abnormal authentication pattern)"
```

**UEBA_lateral_discovery_chain**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "Multiple enumeration commands in sequence — net.exe, ipconfig, netstat, arp"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "Process chain: net.exe, ipconfig, netstat, arp (Exabeam abnormal enumeration)"
```

**UEBA_offhours_activity**
```yaml
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "Process creation at off-hours timestamp (Exabeam: abnormal activity time)"
  - event_id: 4624
    channel: "Security"
    description: "Logon event at unusual hour (Exabeam: abnormal activity time use case)"
```

---

## Common Pitfalls

### Pitfall 1: `Get-WinEvent` on Non-Existent Channel Exits Non-Zero
**What goes wrong:** If Sysmon is not installed, querying `Microsoft-Windows-Sysmon/Operational` causes `Get-WinEvent` to throw a terminating error and the PowerShell process exits with code 1. The verifier interprets this as a query failure rather than "zero events found", causing confusing error messages.
**Why it happens:** `Get-WinEvent` distinguishes between "channel has no matching events" (no output, exit 0) and "channel does not exist" (exception, exit 1). `-ErrorAction SilentlyContinue` suppresses the exception but the exit code still varies by PowerShell version.
**How to avoid:** Wrap the count query in a `try/catch` at the PowerShell level, or treat any non-zero exit as `count=0` in the Go layer. Use: `try { (Get-WinEvent ...).Count } catch { 0 }`.
**Warning signs:** Verification always returns `VerifFail` for all Sysmon events despite techniques running successfully.

### Pitfall 2: Time Window Too Short — Events Not Yet Flushed
**What goes wrong:** Windows Event Log writes are asynchronous. If verification starts immediately after a technique, some events (especially Sysmon high-volume events like EID 3) may not yet be committed.
**Why it happens:** The Sysmon driver buffers events; the PowerShell cmdlet reads from a committed log, not the buffer.
**How to avoid:** The 3-second default wait (D-06) addresses this. Make the wait configurable via `VerificationWaitSecs` in `engine.Config`. Document that fast machines may need only 1-2s while VMs may need 5+s.
**Warning signs:** Intermittent `VerifFail` for events that appear in the log when checked manually 10 seconds later.

### Pitfall 3: Time Window Spans Multiple Technique Runs
**What goes wrong:** If `DelayBetweenTechniques=0`, multiple techniques run in quick succession. A 60-second lookback window for technique N may include events from technique N-1, producing false `VerifPass` for technique N even though its events weren't actually generated.
**Why it happens:** Event IDs like 4688 (process creation) and 4104 (ScriptBlock logging) are generated by many techniques. A broad lookback window conflates results.
**How to avoid:** Use a tight lookback window anchored to `result.StartTime`, not a fixed "last 60 seconds". Recommend a 45-second maximum window (from `StartTime` to `StartTime + max_technique_duration`). Anchor verification at `result.StartTime - 2s` (small negative buffer for clock skew).
**Warning signs:** A technique with empty `ExpectedEvents` (like a UEBA chain placeholder) shows `VerifPass` because it finds events from the previous technique.

### Pitfall 4: Duplicate EventSpec Entries in YAML
**What goes wrong:** T1490 (Inhibit Recovery) has two distinct `4688` entries (one for vssadmin, one for bcdedit). If the verifier deduplicates by `(channel, event_id)`, only one entry is checked.
**Why it happens:** The natural data model groups by EventSpec identity. But the same Event ID can appear multiple times for different processes.
**How to avoid:** The verifier must check each `EventSpec` independently — a count > 0 satisfies that EventSpec. Two `EventSpec` entries with the same `event_id` but different `description` (or `contains`) fields both need to be queried. The `contains` field enables distinguishing them: for T1490, add `contains: "bcdedit"` to the second entry.
**Warning signs:** T1490 shows pass with only 1 matching event when 2 distinct process executions were expected.

### Pitfall 5: YAML Parsing of `EventSpec` with `event_id: 0` Sentinel
**What goes wrong:** Some techniques have non-Security events where `event_id` is a small number like `1` (Sysmon process creation). If `event_id` is accidentally left at its zero value, it matches `EventSpec{}` default construction — a verifier bug could query for EID 0 which doesn't exist.
**Why it happens:** Go zero-value initialization means an `EventSpec` with no `event_id` set has `EventID=0`.
**How to avoid:** Add validation in the verifier: skip `EventSpec` entries where `EventID == 0` or `Channel == ""`. Log a warning for malformed specs during registry loading.

### Pitfall 6: `htmlTemplate` Uses `text/template` Not `html/template`
**What goes wrong:** The reporter imports `text/template` (line 9 of reporter.go). If verification output includes user-supplied strings (e.g., event descriptions from YAML), these could contain `<` or `>` characters that break the HTML structure.
**Why it happens:** The current reporter uses `text/template` throughout, and the `description` field in EventSpec may contain characters like `<`, `>`, `"` (e.g., "TargetImage: lsass.exe, GrantedAccess: 0x1010").
**How to avoid:** Apply `html.EscapeString()` to description strings before inserting them into the template, or add a `htmlEscape` function to the template FuncMap. Do NOT switch to `html/template` for the whole reporter — that would require reworking all existing template logic.

---

## Code Examples

### PowerShell `Get-WinEvent` Query with Error Handling

```powershell
# Source: verified against Windows Get-WinEvent documentation
# Returns integer count; catches channel-not-found error gracefully
try {
    $count = (Get-WinEvent -FilterHashtable @{
        LogName   = 'Security'
        Id        = 4688
        StartTime = '2026-03-24T14:00:00+01:00'
    } -ErrorAction Stop | Measure-Object).Count
    $count
} catch {
    0
}
```

### Go Subprocess Call Matching `executor.runCommand` Pattern

```go
// Source: mirroring executor.go runCommand() pattern (lines 194-216)
func queryCount(channel string, eventID int, since time.Time) int {
    sinceStr := since.UTC().Format("2006-01-02T15:04:05Z")
    script := fmt.Sprintf(
        `try { (Get-WinEvent -FilterHashtable @{LogName='%s';Id=%d;StartTime='%s'} -ErrorAction Stop | Measure-Object).Count } catch { 0 }`,
        channel, eventID, sinceStr,
    )
    cmd := exec.Command("powershell.exe",
        "-NonInteractive", "-NoProfile", "-ExecutionPolicy", "Bypass",
        "-Command", script,
    )
    var buf bytes.Buffer
    cmd.Stdout = &buf
    _ = cmd.Run() // ignore exit code — we always get an integer (0 on failure)
    n, _ := strconv.Atoi(strings.TrimSpace(buf.String()))
    return n
}
```

### `VerificationStatus` Determination Logic

```go
// Source: logic derived from D-08 in CONTEXT.md
func determineStatus(result playbooks.ExecutionResult, verified []playbooks.VerifiedEvent) playbooks.VerificationStatus {
    if !result.Success {
        return playbooks.VerifNotExecuted
    }
    for _, v := range verified {
        if !v.Found {
            return playbooks.VerifFail
        }
    }
    return playbooks.VerifPass
}
```

### YAML Technique File Format After Migration

```yaml
# Source: D-01 from CONTEXT.md + existing YAML tag alignment convention
expected_events:
  - event_id:     4104
    channel:      "Microsoft-Windows-PowerShell/Operational"
    description:  "PowerShell ScriptBlock — decoded script content logged post-deobfuscation"
  - event_id:     1
    channel:      "Microsoft-Windows-Sysmon/Operational"
    description:  "powershell.exe with -NoP -NonI -W Hidden -Exec Bypass in CommandLine"
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `Get-EventLog` PowerShell cmdlet | `Get-WinEvent` with `-FilterHashtable` | Windows Vista / PowerShell 2.0 (2009) | `Get-EventLog` does not support modern channels (Sysmon, WMI Operational, PowerShell Operational); `Get-WinEvent` is the only correct choice |
| Free-text `expected_events []string` | Structured `EventSpec{event_id, channel, description}` | This phase | Enables automated querying instead of manual log lookup |

**Deprecated/outdated:**
- `Get-EventLog -LogName Security -EventId 4688`: Does not support channels like `Microsoft-Windows-Sysmon/Operational`. Replaced by `Get-WinEvent -FilterHashtable`.
- `wevtutil qe` subprocess: Verbose XML output requires XML parsing; harder to count events than `Get-WinEvent | Measure-Object`. Avoid.

---

## Open Questions

1. **`contains` field: include or omit?**
   - What we know: Several techniques (T1490, T1057) have multiple `EventSpec` entries with the same `event_id` on the same channel (two `4688` events for different processes). Without a filter field, the verifier can only confirm "at least one 4688 in the window" — not that the *specific* process ran.
   - What's unclear: How often does this matter in practice for SIEM validation? A broad "EID 4688 found" may be sufficient for most clients.
   - **Recommendation:** Include `contains` as an optional field (D-02 leaves this to Claude). Implement it as an additional `-Message "*keyword*"` filter in the PowerShell query (e.g., `Get-WinEvent -FilterHashtable @{...} | Where-Object {$_.Message -match 'keyword'}`). Default: empty string = no keyword filter. This adds precision without breaking the simple cases.

2. **Where to place the `verifier` package?**
   - What we know: The engine currently imports `executor`, `playbooks`, `reporter`, `simlog`, `userstore`. A new `verifier` package importing `playbooks` is clean. Alternatively, verification code could live in `engine/verify.go` as a file within the engine package.
   - What's unclear: Whether the verifier will ever be called from outside the engine (e.g., a future standalone `/api/verify` endpoint).
   - **Recommendation:** Create `internal/verifier/verifier.go` as a separate package. This follows the single-file-per-package convention and keeps engine.go focused on scheduling logic. The planner should create this as a new package.

3. **Batch vs. per-EventSpec queries?**
   - What we know: Each `Verify()` call makes one PowerShell subprocess invocation per `EventSpec`. A technique like T1003.001 has 5 EventSpecs — that's 5 subprocess calls taking 5 × ~300ms = ~1.5 seconds minimum overhead before the wait delay.
   - What's unclear: Whether this latency is acceptable, or whether a single multi-query PowerShell script would be better.
   - **Recommendation:** Start with per-EventSpec queries (simpler, easier to debug). For Phase 1, 5 subprocess calls at ~300ms each is ~1.5s total — acceptable given the 3s wait delay is already in place. Optimization is a future concern.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| PowerShell (`powershell.exe`) | Event log querying via `Get-WinEvent` | ✓ (Windows-only tool) | Varies (PS 5.1 on Win10/11) | None — tool is Windows-only by design |
| Sysmon | Sysmon channel event verification | Unknown at research time | — | Verification returns `VerifFail` for Sysmon events if not installed — not an error, expected behavior |
| Windows audit policy (via preparation.go) | Security channel event verification | Configurable | — | Without audit policy, Security events 4688/4720/etc. will not appear — documented in prep step |
| PowerShell ScriptBlock Logging | EID 4104 / 4103 verification | Configurable | — | Without preparation.go `EnablePowerShellLogging()`, PS operational events won't appear |

**Missing dependencies with no fallback:**
- None — all dependencies either exist (PowerShell on Windows) or degrade gracefully (Sysmon/audit policy — verification fails with `VerifFail` which is the intended behavior).

**Missing dependencies with fallback:**
- Sysmon not installed → Sysmon channel queries return `count=0` → technique shows `VerifFail` → correct signal (SIEM gap detected).
- Audit policy not configured → Security events missing → technique shows `VerifFail` → correct signal.

---

## Validation Architecture

> nyquist_validation not configured (no `.planning/config.json` found) — treated as enabled.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | None currently — zero test files in codebase |
| Config file | None (Wave 0 gap) |
| Quick run command | `go test ./internal/verifier/... -v -run TestVerify` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VERIF-01 | `EventSpec` fields parse correctly from YAML | unit | `go test ./internal/playbooks/... -run TestEventSpecParsing` | ❌ Wave 0 |
| VERIF-02 | `queryCount` returns correct count for known event | unit (mock) | `go test ./internal/verifier/... -run TestQueryCount` | ❌ Wave 0 |
| VERIF-03 | `determineStatus` returns correct enum for each case | unit | `go test ./internal/verifier/... -run TestDetermineStatus` | ❌ Wave 0 |
| VERIF-04 | HTML template renders verification column correctly | unit | `go test ./internal/reporter/... -run TestHTMLVerificationColumn` | ❌ Wave 0 |
| VERIF-05 | `VerifNotExecuted` when `Success=false`, `VerifFail` when `Success=true` + no events | unit | `go test ./internal/verifier/... -run TestNotExecutedVsEventsMissing` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/verifier/... ./internal/playbooks/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/verifier/verifier_test.go` — covers VERIF-02, VERIF-03, VERIF-05
- [ ] `internal/playbooks/types_test.go` — covers VERIF-01 (YAML parsing of `EventSpec`)
- [ ] `internal/reporter/reporter_test.go` — covers VERIF-04 (HTML template rendering)
- [ ] `go test` infrastructure: no test files currently exist in any package — any test file will be the first

**Note on testability for VERIF-02:** The `queryCount` function calls `powershell.exe` which cannot run in a CI environment without Windows. Design the verifier so `queryCount` is an injectable function (or interface method) so unit tests can mock it. The actual subprocess call only runs in integration/manual tests.

---

## Sources

### Primary (HIGH confidence)
- Direct reading of `internal/playbooks/types.go` — exact struct field names and tags
- Direct reading of `internal/engine/engine.go` — exact injection point (`runTechnique` method, lines 420-466)
- Direct reading of `internal/reporter/reporter.go` — exact HTML template structure, `htmlData` struct, `htmlTemplate` const
- Direct reading of `internal/executor/executor.go` — established `runCommand` pattern for subprocess calls
- Direct reading of all 43 technique YAML files — authoritative `expected_events` source
- Direct reading of `.planning/codebase/CONVENTIONS.md`, `STRUCTURE.md`, `CONCERNS.md`
- Direct reading of `01-CONTEXT.md` — locked decisions

### Secondary (MEDIUM confidence)
- `Get-WinEvent -FilterHashtable` syntax: verified from multiple technique YAML comments referencing specific event IDs and channels, cross-validated with preparation.go audit policy comments listing matching event IDs

### Tertiary (LOW confidence)
- PowerShell exit code behavior for non-existent channels — inferred from general PowerShell error handling knowledge; should be validated in the target environment before finalizing the `try/catch` pattern

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies; all patterns derived from existing codebase
- Architecture: HIGH — based on direct reading of all relevant source files
- Event ID mapping: HIGH — derived directly from technique YAML files; each mapping is a structured interpretation of existing free-text strings
- Pitfalls: MEDIUM — items 1-3 are from direct code analysis; items 4-6 are inferred from Go/PowerShell behavior patterns
- Test architecture: MEDIUM — test framework gaps are known; specific test assertions TBD during implementation

**Research date:** 2026-03-24
**Valid until:** 2026-06-24 (stable domain — Go standard library + Windows Event Log API)
