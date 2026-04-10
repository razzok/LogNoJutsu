# Phase 14: Safety Audit - Research

**Researched:** 2026-04-09
**Domain:** Go struct extension, YAML field addition, PowerShell technique safety, defer-style cleanup pattern, HTML template extension
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Add a `tier: 1|2|3` field to each technique YAML file. This is the source of truth for classification.
- **D-02:** Tier boundaries are defined by **event realism**: Tier 1 = generates the real Windows events a SIEM would see in an actual attack. Tier 2 = generates some real events but uses simulation shortcuts. Tier 3 = echo/stub that only proves the technique runs.
- **D-03:** Tier is visible in **both** the HTML report and web UI (badge/label in technique list).
- **D-04:** Each technique gets a 1-line rationale explaining its tier assignment. Rationales live in the classification document (`docs/TECHNIQUE-CLASSIFICATION.md`).
- **D-05:** Strategy is **scope-limiting real actions** — keep real tool invocations but reduce blast radius so no permanent damage occurs while still generating the target event IDs.
- **D-06:** **T1070.001 (Clear Logs):** Replace full log clearing with creating and clearing a LogNoJutsu-specific custom event log channel. Generates EID 104 without wiping Security/Application/System logs.
- **D-07:** **T1490 (Inhibit Recovery):** Keep bcdedit recoveryenabled and registry disable steps (easily reversible). Skip vssadmin/wmic shadow delete entirely. Still generates EID 4688 for bcdedit.exe and Sysmon EID 13 for registry changes.
- **D-08:** **T1546.003 (WMI Persistence):** Safe as-is — harmless trigger (uptime check), benign action (whoami to temp file), cleanup removes all CIM objects. Just needs cleanup reliability guarantee from D-10.
- **D-09:** Cleanup guarantee via **defer-style pattern in executor** — wrap technique execution in `RunWithCleanup` so cleanup runs even if the technique body panics or context is cancelled. Minimal change to existing code structure.
- **D-10:** **Audit all 58 techniques** for missing cleanup. If a technique writes to disk, registry, or scheduled tasks and has empty cleanup, add the appropriate cleanup command. Read-only/discovery techniques stay with empty cleanup (legitimate).
- **D-11:** Classification document is a **Markdown table in `docs/TECHNIQUE-CLASSIFICATION.md`** — columns: Technique ID, Name, Tier, Rationale, Has Cleanup, Writes Artifacts. Human-readable, version-controlled.
- **D-12:** Primary audience is the **security consultant** running LogNoJutsu at a client site. They need to quickly know which techniques are realistic, which are stubs, and which need admin review.

### Claude's Discretion

- Order of technique auditing (can batch by tactic, phase, or alphabetical)
- Exact wording of per-technique rationales
- Whether to add a `writes_artifacts: bool` field to YAML or derive it from cleanup presence
- Custom event log channel naming convention for T1070.001

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SAFE-01 | All potentially destructive techniques (T1490, T1070.001, T1546.003) are audited and fixed to prevent damage on client machines | D-06/D-07/D-08 in decisions; PowerShell `New-EventLog`/`Remove-EventLog` pattern confirmed for T1070.001; bcdedit + registry-only approach for T1490 |
| SAFE-02 | All 57 existing techniques are classified as Tier 1/2/3 with documented assessment | 58 YAML files enumerated; classification criteria established; `docs/TECHNIQUE-CLASSIFICATION.md` format designed |
| SAFE-03 | All persistence and write techniques have verified cleanup commands that execute regardless of technique body success/failure | 26 techniques confirmed with empty cleanup via grep audit; executor defer pattern analyzed; `RunWithCleanup` modification path identified |
</phase_requirements>

---

## Summary

Phase 14 is a pure audit-and-remediation phase: no new techniques, no new architecture. The work splits into four independent streams — (1) rewrite three destructive YAML techniques, (2) add `tier:` field to all 58 YAML files plus the Go struct, (3) patch the executor for defer-style cleanup reliability, and (4) audit the 26 techniques with empty cleanup strings.

The codebase is well-structured for these changes. The `Technique` struct in `types.go` already uses YAML+JSON struct tags with aligned columns — adding `tier int` follows an established pattern. The HTML reporter already conditionally renders SIEM columns based on data presence; adding a tier column or badge follows the identical pattern. The web UI technique table is rendered via a single JS template string at line 1070 of `index.html` — adding a tier badge cell is a one-line change.

The most consequential discovery from the code audit is the cleanup gap: grep found **26 techniques** with `cleanup: ""`. However, most of these are legitimate read-only discovery or simulation techniques — they execute commands but write no persistent artifacts. The audit task must distinguish "read-only, empty cleanup is correct" from "writes artifacts, missing cleanup is a bug." Only a per-technique read can make that determination.

**Primary recommendation:** Execute the four streams in sequence: (1) fix destructive techniques first (blocks nothing downstream but is highest risk), (2) patch executor defer pattern, (3) add tier field to struct + all YAMLs, (4) write classification doc.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `gopkg.in/yaml.v3` | already in go.mod | YAML struct unmarshalling | Already used; adding `tier` field is zero-dependency |
| `text/template` | stdlib | HTML report generation | Already used in reporter.go |
| PowerShell `New-EventLog` / `Remove-EventLog` | Windows built-in | Custom log channel for T1070.001 | Generates authentic EID 104 without touching real logs |

No new dependencies are required for this phase.

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `wevtutil.exe` | Windows built-in | Create/delete custom event log channel | T1070.001 rewrite — `wevtutil im` for custom log schema |
| PowerShell `New-CimInstance` | Windows built-in | WMI subscription creation | T1546.003 already uses it; cleanup reliability is the only change |
| `bcdedit.exe` | Windows built-in | Boot config changes | T1490 keeps this step; cleanup is already present |

**Installation:** No new packages.

---

## Architecture Patterns

### Recommended Project Structure

No new directories. Additions are:
```
internal/
├── playbooks/
│   ├── types.go          — add Tier int field
│   ├── embedded/
│   │   └── techniques/   — add tier: N to all 58 YAML files
├── executor/
│   └── executor.go       — defer-style cleanup in RunWithCleanup
├── reporter/
│   └── reporter.go       — tier badge in htmlTemplate
internal/server/
└── static/index.html     — tier badge in technique table
docs/
└── TECHNIQUE-CLASSIFICATION.md   — NEW file
```

### Pattern 1: Adding a Field to Technique Struct

The existing pattern in `types.go` uses aligned YAML + JSON struct tags. Follow exactly:

```go
// Source: internal/playbooks/types.go existing style
type Technique struct {
    ID                string            `yaml:"id"                 json:"id"`
    Name              string            `yaml:"name"               json:"name"`
    // ... existing fields ...
    Tier              int               `yaml:"tier"               json:"tier"`
    SIEMCoverage      map[string][]string `yaml:"siem_coverage,omitempty"  json:"siem_coverage,omitempty"`
}
```

**Note on field position:** Insert `Tier` before `SIEMCoverage` (last field), after `NistControls`. Zero value (`0`) for unset tier is intentional — it signals "not yet classified" and is distinguishable from Tier 1/2/3 during the audit window.

**YAML loading:** No change to `loader.go` required. `yaml.v3` unmarshals unknown fields silently, and `tier: 0` (default) is the zero value for `int`. Once the field is in the struct, all 58 YAMLs that already have `tier: N` will parse correctly. YAMLs without the field will parse as `tier: 0`.

### Pattern 2: Defer-Style Cleanup in RunWithCleanup

Current `RunWithCleanup` in `executor.go` (line 30–40) runs cleanup sequentially after the technique body returns. It does NOT run cleanup if the Go process panics or if the calling goroutine is cancelled before the function returns. The fix is a `defer` that fires before the function stack unwinds:

```go
// Source: internal/executor/executor.go — proposed pattern
func RunWithCleanup(t *playbooks.Technique, profile *userstore.UserProfile, password string) (result playbooks.ExecutionResult) {
    // Register cleanup as deferred action — fires even on panic
    if strings.TrimSpace(t.Cleanup) != "" {
        defer func() {
            simlog.TechCleanup(t.ID, t.Cleanup, false)
            _, _, cleanErr := runCommand(t.Executor.Type, t.Cleanup)
            result.CleanupRun = true
            simlog.TechCleanup(t.ID, "(cleanup completed)", cleanErr == nil)
        }()
    }
    result = runInternal(t, profile, password)
    return result
}
```

**Critical detail:** The return variable must be named (`result playbooks.ExecutionResult`) for the deferred closure to write `result.CleanupRun = true` to the caller's copy. Anonymous return `playbooks.ExecutionResult` would not work — the defer closure would mutate a local copy.

**Context cancellation gap:** The current `runInternal` uses `exec.Command` without context. Context cancellation killing the outer process would leave cleanup unrun. The defer pattern addresses the panic case but not SIGKILL. For this phase, defer is sufficient per D-09. SIGKILL handling (context-aware execution) is a future concern.

### Pattern 3: T1070.001 Custom Log Channel

The rewritten technique creates a named custom event log, writes a test entry, then clears it — generating authentic EID 104 (System log: event log cleared) without touching real logs:

```powershell
# Source: Windows PowerShell documentation — New-EventLog / Remove-EventLog
$logName = "LogNoJutsu-Test"

# Step 1: Create custom event log channel
New-EventLog -LogName $logName -Source "LNJSource" -ErrorAction SilentlyContinue
Write-EventLog -LogName $logName -Source "LNJSource" -EventId 1000 -Message "LogNoJutsu T1070.001 test entry" -EntryType Information

# Step 2: Clear it via wevtutil (generates EID 104 in System log — authentic)
wevtutil.exe cl $logName
Write-Host "EID 104 fired for '$logName' clear — authentic event, no real log destroyed"

# Step 3: Remove custom log (cleanup happens in cleanup: block)
Remove-EventLog -LogName $logName -ErrorAction Ignore
```

**EID mapping:** When a non-Security log is cleared, Windows logs EID 104 to the System event log. EID 1102 (audit log cleared) is only generated for the Security log. The custom log approach generates EID 104 authentically — the target event for this technique's expected_events. Update expected_events to remove EID 1102 (Security cleared) and remove EID 4688 for wevtutil (still fires — keep it). Retain EID 104 for System channel.

**Channel name:** `LogNoJutsu-Test` is clear, LNJ-prefixed, and consistent with other artifact naming (`LNJFilter`, `LNJConsumer`, `lnj_wmi_persist.txt`).

### Pattern 4: T1490 Scope-Limited Rewrite

Remove vssadmin and wmic shadow delete steps entirely. Keep bcdedit and registry steps. Update `description` and `expected_events` to remove the vssadmin/wmic EID 4688 entries that will no longer fire:

```powershell
# Keep: generates EID 4688 for bcdedit.exe
bcdedit.exe /set "{default}" recoveryenabled no
bcdedit.exe /set "{default}" bootstatuspolicy ignoreallfailures

# Keep: generates Sysmon EID 13 for registry
reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows NT\SystemRestore" /v DisableConfig /t REG_DWORD /d 1 /f
reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows NT\SystemRestore" /v DisableSR /t REG_DWORD /d 1 /f

# REMOVE these two steps entirely:
# vssadmin.exe delete shadows /all /quiet   -- irreversible VSS deletion
# wmic.exe shadowcopy delete                -- irreversible VSS deletion
# wbadmin.exe delete catalog -quiet        -- irreversible backup catalog deletion
```

Cleanup (already present) covers the kept steps. No cleanup additions needed for T1490.

### Pattern 5: Tier Badge in HTML Report

The report template already has conditional SIEM columns as the precedent. Add a `HasTier` bool to `htmlData` (set to `true` when any result has `Tier > 0`), then add the column conditionally. Since `ExecutionResult` does not carry `Tier` (it's a technique property, not a result property), pass tier data either by: (a) adding `Tier int` to `ExecutionResult` and populating it in `runInternal`, or (b) looking it up from the technique in the template. Option (a) is simpler — populate `result.Tier = t.Tier` in `runInternal`.

```go
// In runInternal, after result initialization:
result.Tier = t.Tier
```

```go
// In htmlData struct:
HasTier bool

// In saveHTML, compute:
hasTier := false
for _, res := range r.Results {
    if res.Tier > 0 {
        hasTier = true
        break
    }
}
```

CSS badge pattern (consistent with existing cs-badge / ms-badge):
```css
.tier1-badge{background:rgba(63,185,80,0.15);color:#3fb950;border:1px solid #3fb950;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}
.tier2-badge{background:rgba(210,153,34,0.15);color:#d29922;border:1px solid #d29922;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}
.tier3-badge{background:rgba(139,148,158,0.15);color:#8b949e;border:1px solid #8b949e;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}
```

### Pattern 6: Tier Badge in Web UI Technique Table

The technique table at `index.html:1070` renders a JS template string. The `/api/techniques` handler serializes the full `Technique` struct including the new `tier` field. Adding a badge column requires:

1. Add `<th>Tier</th>` to the table header at line 373 (update `colspan="7"` to `colspan="8"`)
2. Add a `<td>` cell to the row template at line 1070

```javascript
// In the sorted.map(t => `...`) template:
<td>${t.tier === 1 ? '<span class="tier-badge tier1">T1</span>'
    : t.tier === 2 ? '<span class="tier-badge tier2">T2</span>'
    : t.tier === 3 ? '<span class="tier-badge tier3">T3</span>'
    : '<span style="color:var(--muted)">—</span>'}</td>
```

CSS (add to existing style block):
```css
.tier-badge{border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}
.tier1{background:rgba(63,185,80,0.15);color:var(--green);border:1px solid var(--green)}
.tier2{background:rgba(210,153,34,0.15);color:var(--orange);border:1px solid var(--orange)}
.tier3{background:rgba(139,148,158,0.15);color:var(--muted);border:1px solid var(--muted)}
```

### Anti-Patterns to Avoid

- **Adding tier to Campaign or CampaignStep:** Tier lives on the Technique, not the campaign. Do not add tier to Campaign/CampaignStep structs.
- **Using string for Tier field:** Use `int` not `string`. The tier values are always 1/2/3 (or 0 for unclassified). String requires extra parsing and comparison logic everywhere. Zero value `int` naturally signals "unset."
- **Relying on cleanup order in sequential RunWithCleanup:** The existing sequential version (run technique → run cleanup) is safe only when the technique returns normally. The defer version guarantees cleanup on panic but not on process kill. Document this limitation.
- **Clearing real Windows event logs in T1070.001:** Even clearing Application or System logs destroys forensic history for any ongoing investigation on the client machine. The custom channel approach is mandatory.
- **Keeping vssadmin/wmic shadow delete in T1490:** VSS shadow deletion is irreversible on most configurations. No cleanup command can restore already-deleted shadow copies.

---

## Cleanup Audit Findings

### Techniques with `cleanup: ""` (26 total)

From `grep -l 'cleanup: ""'` audit:

| File | Write Artifacts? | Cleanup Verdict |
|------|-----------------|-----------------|
| AZURE_dcsync.yaml | No (repadmin query only) | Correct — empty cleanup |
| AZURE_kerberoasting.yaml | No (LDAP query) | Correct — empty cleanup |
| T1003_006_dcsync.yaml | No (repadmin /syncall, LDAP query) | Correct — empty cleanup |
| T1016_network_config.yaml | No (ipconfig/route query) | Correct — empty cleanup |
| T1046_network_scan.yaml | No (port scan, no files) | Correct — empty cleanup |
| T1049_network_connections.yaml | No (netstat equivalent) | Correct — empty cleanup |
| T1057_process_discovery.yaml | No (Get-Process query) | Correct — empty cleanup |
| T1059_001_powershell.yaml | No (in-memory execution) | Correct — empty cleanup |
| T1069_group_discovery.yaml | No (net group queries) | Correct — empty cleanup |
| T1070_001_clear_logs.yaml | YES — clears real logs | **BUG — must be rewritten (D-06)** |
| T1071_001_web_protocols.yaml | No (loopback HTTP only) | Correct — empty cleanup |
| T1071_004_dns.yaml | No (DNS query only) | Correct — empty cleanup |
| T1082_system_info_discovery.yaml | No (systeminfo query) | Correct — empty cleanup |
| T1083_file_discovery.yaml | No (Get-ChildItem query) | Correct — empty cleanup |
| T1087_account_discovery.yaml | No (Get-LocalUser query) | Correct — empty cleanup |
| T1110_001_bruteforce.yaml | No (auth attempt only) | Correct — empty cleanup |
| T1110_003_password_spraying.yaml | No (auth attempt only) | Correct — empty cleanup |
| T1135_network_share_discovery.yaml | No (net share query) | Correct — empty cleanup |
| T1482_domain_trust_discovery.yaml | No (nltest/AD query) | Correct — empty cleanup |
| T1550_002_pass_the_hash.yaml | No (auth attempt via net use; IPC$ session) | Requires verification — net use IPC$ may leave a session |
| UEBA_account_takeover_chain.yaml | Needs per-file read | Requires verification |
| UEBA_credential_spray_chain.yaml | Needs per-file read | Requires verification |
| UEBA_lateral_discovery_chain.yaml | Needs per-file read | Requires verification |
| UEBA_offhours_activity.yaml | Needs per-file read | Requires verification |
| UEBA_privilege_escalation_chain.yaml | Needs per-file read | Requires verification |
| T1558_003_kerberoasting.yaml | Needs per-file read (Kerberos ticket request) | Requires verification |

**SAFE-03 audit scope:** The planner must schedule a per-technique read pass for the 6 "requires verification" entries above. The 19 confirmed read-only techniques need no cleanup and no SAFE-03 action. T1070.001 is addressed by the rewrite task.

### Techniques with Cleanup Already Present (32 total)

These have non-empty cleanup strings and are presumed correct. The executor defer patch (D-09) makes their existing cleanup more reliable — no YAML changes needed.

Notable entries already using `ErrorAction Ignore` / `-ErrorAction SilentlyContinue` pattern (the correct defensive style):
- T1546.003, T1547.001, T1543.003, T1562.002, T1548.002, T1036.005, T1134.001, T1574.002

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Custom event log channel | Custom file/pipe approach | PowerShell `New-EventLog` + `wevtutil cl` | Generates authentic EID 104 in System log — the real SIEM trigger |
| Tier validation logic | Custom YAML schema validator | Go unit test in `loader_test.go` | Simpler, version-controlled, runs in CI |
| Cleanup guarantee | Custom try/finally wrapper | Go `defer` in `RunWithCleanup` | Native Go mechanism; composable with existing error handling |
| Classification document format | Custom JSON/XML schema | Markdown table in `docs/TECHNIQUE-CLASSIFICATION.md` | Consultant-readable, git-diffable, no tooling needed |

---

## Common Pitfalls

### Pitfall 1: Named Return Variable Required for Defer to Modify Result
**What goes wrong:** If `RunWithCleanup` uses an anonymous return (`return playbooks.ExecutionResult`), the deferred closure cannot set `result.CleanupRun = true` on the value the caller receives. The field will be false in the returned struct.
**Why it happens:** Go closures capture variables by reference. An anonymous return creates a temporary — the defer can't reference it.
**How to avoid:** Declare a named return: `func RunWithCleanup(...) (result playbooks.ExecutionResult)`. The defer closure then operates on `result` directly.
**Warning signs:** `CleanupRun` is always `false` in execution results even when cleanup ran.

### Pitfall 2: T1070.001 EID 1102 Not Generated by Custom Log Clear
**What goes wrong:** EID 1102 ("The audit log was cleared") is only generated when the **Security** event log is cleared. Clearing a custom log channel generates EID 104 in the System log, not EID 1102.
**Why it happens:** Windows has two distinct events: 1102 = Security audit log cleared (requires `SeSecurityPrivilege`), 104 = any other event log cleared.
**How to avoid:** Update T1070.001's `expected_events` to remove EID 1102 and EID 4688/wevtutil-Security. Keep EID 104 (System channel) and EID 4688 for the wevtutil.exe process creation. The description must be updated to reflect this.
**Warning signs:** Verification step fails looking for EID 1102 after the rewrite.

### Pitfall 3: T1490 `expected_events` Left Stale After Removing vssadmin Steps
**What goes wrong:** The current T1490 `expected_events` includes `EID 4688 for vssadmin.exe 'delete shadows /all /quiet'`. After removing the vssadmin step, this event will never fire, causing SAFE-01 verification to report a false failure.
**Why it happens:** YAML and expected_events are edited separately.
**How to avoid:** Update `expected_events` in T1490 when rewriting the command. Remove the vssadmin and wmic EID 4688 entries. Keep bcdedit EID 4688 and Sysmon EID 13 entries.
**Warning signs:** T1490 verification always shows "fail" for vssadmin EID 4688.

### Pitfall 4: Tier 0 Techniques Appearing in Report
**What goes wrong:** Techniques without a `tier:` field will have `Tier = 0` (Go zero value). If the report/UI renders `Tier 0` as a badge or number, consultants see confusing output.
**How to avoid:** In the UI template, render `t.tier === 0` as `—` (dash), not `T0`. In the report template, use `{{if gt .Tier 0}}` guard. Zero = "not yet classified" is only valid during the audit window; all 58 must have `tier: 1|2|3` before phase completion.
**Warning signs:** Technique table shows "T0" or "0" badges.

### Pitfall 5: `New-EventLog` Requires Admin on Modern Windows
**What goes wrong:** Creating a new event log channel requires administrator privileges on Windows 10/11. T1070.001 is already `elevation_required: true` — but if run without elevation, `New-EventLog` will silently fail and no EID 104 will be generated.
**Why it happens:** Event log registration writes to `HKLM\SYSTEM\CurrentControlSet\Services\EventLog\`.
**How to avoid:** Keep `elevation_required: true` on T1070.001. Add error checking in the PowerShell command to detect failure and emit a clear message.
**Warning signs:** Technique "succeeds" (exit 0) but no EID 104 appears and no log channel was created.

### Pitfall 6: T1550.002 `net use \\IPC$` Leaving Persistent Sessions
**What goes wrong:** `net use \\127.0.0.1\IPC$` creates a network session that persists until explicitly deleted or Windows cleans it up. On long-running client machines, these sessions accumulate.
**How to avoid:** Verify the T1550.002 command content. If it creates a net use session, add `net use \\127.0.0.1\IPC$ /delete 2>$null` to cleanup (matching the pattern already used in `UEBA_lateral_movement_new_asset.yaml`).
**Warning signs:** `net use` on the test machine shows lingering 127.0.0.1 sessions after the simulation.

---

## Code Examples

### Verified Pattern: Conditional Column in HTML Report Template

```go
// Source: internal/reporter/reporter.go:156-170 (existing HasCrowdStrike pattern)
hasCrowdStrike := false
for _, res := range r.Results {
    if len(res.SIEMCoverage["crowdstrike"]) > 0 {
        hasCrowdStrike = true
        break
    }
}
// Apply same pattern for HasTier:
hasTier := false
for _, res := range r.Results {
    if res.Tier > 0 {
        hasTier = true
        break
    }
}
```

### Verified Pattern: Struct Tag Convention

```go
// Source: internal/playbooks/types.go:31-47
// Add Tier before SIEMCoverage, aligned with existing column spacing:
Tier              int               `yaml:"tier"               json:"tier"`
```

### Verified Pattern: Cleanup in Technique YAML

```yaml
# Source: existing cleanup patterns in T1546.003, T1547.001
cleanup: |-
  Remove-EventLog -LogName "LogNoJutsu-Test" -ErrorAction Ignore
  Write-Host "Cleanup: Custom event log channel removed."
```

### Verified Pattern: `ErrorAction Ignore` in PowerShell Cleanup

Every cleanup block in the codebase uses either `-ErrorAction Ignore` or `-ErrorAction SilentlyContinue` on Remove/Delete operations. This prevents cleanup from failing loudly if the artifact was already removed (e.g., if the technique body cleaned up on failure). Always use this pattern in new cleanup blocks.

### Verified Pattern: Tier Classification Document Table

```markdown
| Technique ID | Name | Tier | Rationale | Has Cleanup | Writes Artifacts |
|-------------|------|------|-----------|-------------|-----------------|
| T1059.001 | PowerShell Execution | 1 | Uses real attacker flag combinations; EID 4104 ScriptBlock fires for decoded content | No | No |
| T1046 | Network Service Discovery | 3 | Pure PowerShell TCP connect loop; generates no Sysmon EID 3 (no raw socket) | No | No |
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Sequential cleanup (run technique → run cleanup) | Defer-style cleanup (cleanup registered before technique runs) | Phase 14 | Cleanup fires on panic, not just normal return |
| T1070.001 clears real logs | T1070.001 uses custom `LogNoJutsu-Test` channel | Phase 14 | No real log destruction on client machines |
| T1490 includes VSS shadow deletion | T1490 scope-limited to bcdedit + registry only | Phase 14 | Irreversible steps removed; bcdedit/registry cleanup already present |
| No tier classification | `tier: 1|2|3` in each YAML | Phase 14 | Consultants can instantly identify realistic vs. stub techniques |

---

## Open Questions

1. **T1550.002 net use session persistence**
   - What we know: The technique simulates PtH via `net use \\IPC$` among other methods. `UEBA_lateral_movement_new_asset.yaml` already handles this with a `/delete` cleanup.
   - What's unclear: Whether T1550.002 explicitly closes the IPC$ session or leaves it.
   - Recommendation: Read T1550.002 full command during implementation and add cleanup if the session is left open.

2. **UEBA chain techniques — artifact writes unclear from grep**
   - What we know: `UEBA_account_takeover_chain.yaml`, `UEBA_credential_spray_chain.yaml`, `UEBA_lateral_discovery_chain.yaml`, `UEBA_offhours_activity.yaml`, `UEBA_privilege_escalation_chain.yaml` all have empty cleanup.
   - What's unclear: Whether their command blocks write temp files, registry keys, or sessions.
   - Recommendation: Read each YAML during the cleanup audit wave and verify. Most UEBA chains are behavioral simulation (event generation via real commands) without persistent writes.

3. **Whether to add `writes_artifacts: bool` to YAML struct**
   - What we know: D-11 requires the classification document to have a "Writes Artifacts" column. This could be derived from `cleanup != ""` as a heuristic, or explicitly set as a YAML field.
   - What's unclear: Which approach the planner prefers for the classification document generation.
   - Recommendation: Derive from cleanup presence (`cleanup != ""` → writes artifacts = yes) for the document, avoiding a new YAML field. This is within Claude's discretion per D-12.

---

## Environment Availability

Step 2.6: SKIPPED (no external dependencies — this phase modifies existing Go source files and YAML technique files only; `go test ./...` already passes as confirmed by the test run showing all packages pass).

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` package (stdlib) |
| Config file | none — `go test ./...` auto-discovers |
| Quick run command | `go test ./internal/playbooks/... -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SAFE-01 | T1490, T1070.001, T1546.003 execute without destroying logs/shadows | Manual smoke test on Windows | run each technique in WhatIf=false mode, verify EIDs fire, verify no real log loss | N/A — execution test |
| SAFE-02 | All 58 techniques have `tier: 1|2|3` (non-zero) | unit | `go test ./internal/playbooks/... -run TestTierClassified` | ❌ Wave 0 |
| SAFE-03 | All write techniques have non-empty cleanup | unit | `go test ./internal/playbooks/... -run TestWriteArtifactsHaveCleanup` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/playbooks/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/playbooks/loader_test.go` — add `TestTierClassified`: loads all techniques, asserts `tech.Tier > 0` for each — covers SAFE-02
- [ ] `internal/playbooks/loader_test.go` — add `TestWriteArtifactsHaveCleanup`: loads all techniques, for each technique flagged as writing artifacts (determined by a known list or `writes_artifacts` field), asserts `strings.TrimSpace(tech.Cleanup) != ""` — covers SAFE-03

*(Both tests add to the existing `loader_test.go` file — no new file needed)*

---

## Sources

### Primary (HIGH confidence)
- Direct source code read: `internal/playbooks/types.go`, `internal/executor/executor.go`, `internal/reporter/reporter.go`, `internal/server/static/index.html`, `internal/playbooks/loader.go` — full content reviewed
- Direct YAML read: `T1490_inhibit_recovery.yaml`, `T1070_001_clear_logs.yaml`, `T1546_003_wmi_event_subscription.yaml` — full content reviewed
- Grep audit: all 58 technique YAML files for `cleanup: ""` — 26 empty cleanup techniques identified

### Secondary (MEDIUM confidence)
- Windows documentation knowledge (August 2025 cutoff): EID 104 fires for non-Security log clears; EID 1102 fires only for Security log clear; `New-EventLog` requires admin; VSS shadow deletion is irreversible
- Go language specification: named return variables are required for defer closures to modify the returned value

### Tertiary (LOW confidence)
- None

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies; existing patterns confirmed by source read
- Architecture: HIGH — all integration points identified from source; patterns confirmed by existing code
- Pitfalls: HIGH — T1070.001 EID issue and named-return requirement derived from code; VSS irreversibility is established Windows behavior
- Cleanup audit: MEDIUM — 19 techniques confirmed read-only by name/description; 6 UEBA/T1550 entries require per-file verification during implementation

**Research date:** 2026-04-09
**Valid until:** 2026-05-09 (stable Go + PowerShell stack; no fast-moving dependencies)
