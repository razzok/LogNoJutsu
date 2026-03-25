# Phase 4: CrowdStrike SIEM Coverage - Research

**Researched:** 2026-03-25
**Domain:** CrowdStrike Falcon detection mappings, Go struct extension, HTML conditional column, YAML technique authoring
**Confidence:** HIGH (codebase analysis) / MEDIUM (Falcon detection names)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-01:** Add `SIEMCoverage map[string][]string` field to the `Technique` struct with YAML key `siem_coverage` and `omitempty`. The map key is the SIEM platform name (`crowdstrike`, `sentinel`), value is a list of detection rule names. This single field handles both Phase 4 (CrowdStrike) and Phase 5 (Sentinel) — no struct change needed in Phase 5.

**D-02:** Techniques with no SIEM mappings omit the `siem_coverage` field entirely (`omitempty`). Not all 52 existing techniques will have CrowdStrike mappings — only add where there is a genuine, accurate mapping. Keep YAML files clean.

**D-03:** Create 3 new YAML technique files (prefix `FALCON_`) rather than modifying existing techniques. Targets:
1. `FALCON_process_injection` — CreateRemoteThread / VirtualAllocEx pattern; triggers Falcon's process injection behavioral detection
2. `FALCON_lsass_access` — LSASS memory access via OpenProcess/ReadProcessMemory; triggers Falcon's credential theft detection
3. `FALCON_lateral_movement_psexec` — PsExec-style / WMI remote execution pattern; triggers Falcon's lateral movement detection

**D-04:** Each FALCON_ technique includes `expected_events` (structured EventSpec entries) for the Windows/Sysmon events generated, AND `siem_coverage.crowdstrike` with the official Falcon alert names those behaviors trigger.

**D-05:** CrowdStrike column shows: green **CS** badge + list of detection rule names when `siem_coverage.crowdstrike` is populated for that technique. Grey **N/A** cell when no CrowdStrike mappings exist for that technique.

**D-06:** The CrowdStrike column is **conditional** — it only renders when at least one technique in the results has `siem_coverage.crowdstrike` populated. This matches CROW-03 wording ("when Falcon events are present") and keeps the report clean on non-CrowdStrike environments. Consistent with how the Sentinel column will work in Phase 5.

**D-07:** Column style follows the Phase 1 badge pattern — the CS badge should visually match the existing pass/fail badge style already used in the verification column.

**D-08:** Use **official Falcon alert names** as they appear in the CrowdStrike Falcon console UI. These are the exact strings consultants will see in the client's Falcon dashboard — makes the report directly actionable.

### Claude's Discretion

- Specific official Falcon alert names to map to each existing technique — research actual Falcon detection names
- Which existing techniques (out of 52) merit CrowdStrike mappings — only map where Falcon genuinely fires
- Exact PowerShell/cmd commands in FALCON_ techniques — must generate the expected Windows/Sysmon events safely (no real attacks), following the multi-variant pattern from Phase 3 (D-06)
- Where to add CrowdStrike documentation in the README — extend existing German README following Phase 3's documentation pattern

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CROW-01 | CrowdStrike Falcon detection rule mappings documented per technique in events manifest | Covered by: `siem_coverage` field on Technique struct + YAML population for ~10-15 existing techniques + all 3 FALCON_ techniques |
| CROW-02 | At least 3 techniques that specifically generate Falcon sensor events | Covered by: 3 new FALCON_ YAML files, each authored with Falcon behavioral triggers in mind + confirmed detection policy names |
| CROW-03 | HTML report shows CrowdStrike-specific coverage column when Falcon events are present | Covered by: `HasCrowdStrike bool` flag in htmlData, conditional template rendering, CS badge CSS |
</phase_requirements>

---

## Summary

Phase 4 is a data + presentation phase: add one struct field, populate YAML data, and extend the HTML template. There are no new package dependencies, no engine changes, and no new Go packages. The hard problems are (1) choosing correct Falcon detection names and (2) authoring safe FALCON_ technique YAML files.

CrowdStrike Falcon exposes detection taxonomy through two overlapping concepts: **Prevention Policy features** (toggle names like `credential_dumping`, `code_injection`) and **IOA behavioral detections** shown as alert names in the console. The prevention policy names are machine-readable strings from the API/Terraform; the console-visible alert names are less formally documented publicly. Research found the prevention policy feature names with high confidence via the Pulumi/Terraform registry. Console-facing detection names can be derived from the prevention policy display names and CrowdStrike's published material, but should be treated as MEDIUM confidence and documented as such in YAML comments.

The three FALCON_ techniques map cleanly to existing analogues: `FALCON_lsass_access` is a Falcon-optimized variant of `T1003.001`, `FALCON_process_injection` is novel (no existing T1055 technique in the library), and `FALCON_lateral_movement_psexec` overlaps with `T1021.002` but targets PsExec/WMI remote execution patterns. All three generate authentic Windows/Sysmon events. The existing `T1003.001` technique already uses `OpenProcess(0x1010)` — `FALCON_lsass_access` needs to differentiate with a different attack angle or documented Falcon-specific sensor trigger.

**Primary recommendation:** Use Falcon prevention policy feature names as the primary detection identifiers (e.g., `"Credential Dumping"`, `"Code Injection"`, `"Suspicious Scripts and Commands"`) since these are the exact toggle-level labels visible in the Falcon console under Prevention Policies. These are authoritative, version-stable, and actionable for consultants tuning client policies.

---

## Standard Stack

### Core (no new dependencies — all built-in to existing project)

| Component | Version | Purpose | Notes |
|-----------|---------|---------|-------|
| `gopkg.in/yaml.v3` | current (in go.mod) | YAML unmarshalling for `SIEMCoverage map[string][]string` | Already in project — `map[string][]string` is natively supported |
| `text/template` | stdlib | HTML template conditional column rendering | Already used in reporter.go |
| Go stdlib | go 1.26.1 | No new imports needed | No external packages required |

**Installation:** No new packages. This phase is zero-dependency from a library perspective.

**Version verification:** `go.mod` shows `gopkg.in/yaml.v3` already present. `map[string][]string` unmarshals from YAML block-map syntax with no additional configuration.

---

## Architecture Patterns

### Recommended Project Structure (changes only)

```
internal/
├── playbooks/
│   ├── types.go                    # ADD: SIEMCoverage field to Technique struct
│   └── embedded/techniques/
│       ├── FALCON_process_injection.yaml      # NEW
│       ├── FALCON_lsass_access.yaml           # NEW
│       ├── FALCON_lateral_movement_psexec.yaml # NEW
│       └── T1059_001_powershell.yaml          # MODIFY: add siem_coverage
│       └── (10-14 other existing .yaml files) # MODIFY: add siem_coverage
└── reporter/
    └── reporter.go                 # MODIFY: HasCrowdStrike flag, CSS, template
```

### Pattern 1: Adding SIEMCoverage to Technique Struct

**What:** Add one field to the `Technique` struct following the existing `InputArgs` and `NistControls` omitempty pattern.

**Key insight from codebase:** The struct tag alignment convention uses spaces to visually align the backtick tags (see types.go lines 33–46). `SIEMCoverage` must follow this exact alignment style.

```go
// Source: internal/playbooks/types.go — follow existing alignment pattern
SIEMCoverage  map[string][]string `yaml:"siem_coverage,omitempty"  json:"siem_coverage,omitempty"`
```

Place after `NistControls` (line 45) — consistent with other optional map/slice fields at the end of the struct.

**YAML tag note:** `map[string][]string` in YAML is a block-map where each key maps to a sequence. The `gopkg.in/yaml.v3` library handles this natively with zero additional configuration.

### Pattern 2: YAML SIEM Coverage Block

**What:** The `siem_coverage` key in technique YAML files uses standard YAML block-map syntax.

```yaml
# Source: CONTEXT.md D-02, consistent with existing YAML field style
siem_coverage:
  crowdstrike:
    - "Credential Dumping"
    - "Suspicious Scripts and Commands"
```

**When to omit:** If no verified Falcon detection applies, omit the field entirely. Do not add placeholder or speculative entries. The `omitempty` tag means absent YAML key = nil map = no rendering in template.

### Pattern 3: HasCrowdStrike Flag in htmlData

**What:** Compute a boolean flag in `saveHTML()` by scanning results, following the exact same pattern used for `WhatIf bool`.

**Key insight from codebase (reporter.go line 83–95):** `htmlData` struct holds `WhatIf bool` which is set from `r.WhatIf`. The `HasCrowdStrike` flag follows the same pattern but requires scanning results + looking up the technique's `SIEMCoverage`.

**Problem:** `ExecutionResult` (reporter.go) currently does not carry `SIEMCoverage` data — it only has `TechniqueID` and result fields. The reporter's `SaveResults` receives `[]playbooks.ExecutionResult`, not `[]playbooks.Technique`. This means the reporter either needs:
- Option A: `ExecutionResult` carries `SIEMCoverage map[string][]string` (add field to ExecutionResult)
- Option B: Reporter receives the technique registry as an additional parameter
- Option C: Engine populates `SIEMCoverage` on `ExecutionResult` at run time

**Recommended approach (Option A):** Add `SIEMCoverage map[string][]string` to `ExecutionResult` as an `omitempty` JSON field. The engine's `runTechnique()` populates it from `t.SIEMCoverage` when building the result. This is the least invasive change — no new parameters, no registry threading, and the data is available wherever the result is used (JSON report, HTML report, future API endpoints).

```go
// internal/playbooks/types.go — ExecutionResult addition
SIEMCoverage  map[string][]string `json:"siem_coverage,omitempty"`
```

```go
// internal/engine/engine.go — runTechnique() — copy coverage into result
// After building result (line 452 or in WhatIf branch, line 452):
result.SIEMCoverage = t.SIEMCoverage
```

```go
// internal/reporter/reporter.go — saveHTML() — compute HasCrowdStrike
hasCrowdStrike := false
for _, res := range r.Results {
    if len(res.SIEMCoverage["crowdstrike"]) > 0 {
        hasCrowdStrike = true
        break
    }
}
```

### Pattern 4: Conditional HTML Column

**What:** The CrowdStrike column header and each row cell render only when `HasCrowdStrike` is true. Follows the `{{if .WhatIf}}` pattern already in the template.

```html
<!-- Column header — add after existing <th>Verifikation</th> -->
{{if .HasCrowdStrike}}<th>CrowdStrike</th>{{end}}

<!-- Row cell — add after existing </td> for verification -->
{{if $.HasCrowdStrike}}
<td>
  {{if index .SIEMCoverage "crowdstrike"}}
    <span class="cs-badge">CS</span>
    <ul class="cs-list">
      {{range index .SIEMCoverage "crowdstrike"}}
      <li>{{.}}</li>
      {{end}}
    </ul>
  {{else}}
    <span class="cs-na">N/A</span>
  {{end}}
</td>
{{end}}
```

**CSS additions (follow existing badge patterns):**
```css
.cs-badge{background:#e01b22;color:#fff;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600}
.cs-na{color:#8b949e;font-size:11px}
.cs-list{font-size:11px;margin-top:4px;padding-left:0;list-style:none;color:#e6edf3}
```

**Note on `$.HasCrowdStrike` vs `.HasCrowdStrike`:** Inside `{{range .Results}}`, `.` refers to the current result. Use `$.HasCrowdStrike` to access the root template data.

**Note on `index .SIEMCoverage "crowdstrike"`:** Go templates use `index` to access map values dynamically. `SIEMCoverage.crowdstrike` is NOT valid template syntax for map access.

### Pattern 5: FALCON_ Technique YAML Structure

**What:** FALCON_ files follow the exact existing YAML schema. Key decisions about tactic and phase values:

- **tactic:** Use the matching MITRE tactic (`credential-access`, `defense-evasion`, `lateral-movement`) — NOT `"crowdstrike-falcon"`. The engine's `filterByTactics` filters by `t.Tactic` value. Using standard MITRE tactic names ensures FALCON_ techniques run in the attack phase alongside related techniques and can be filtered/included by tactic name.
- **phase:** `"attack"` — all FALCON_ techniques are attack-phase behaviors
- **elevation_required:** `true` for FALCON_lsass_access (LSASS requires SeDebugPrivilege); `true` for FALCON_process_injection (kernel/API injection); `false` for FALCON_lateral_movement_psexec depending on approach

```yaml
# FALCON_ technique skeleton
id: FALCON_lsass_access
name: Falcon Sensor — LSASS Credential Theft Detection
description: >
  Simulates LSASS memory access patterns that trigger CrowdStrike Falcon's
  credential theft behavioral detection. Uses OpenProcess with elevated access
  mask targeting lsass.exe — same pattern as Mimikatz but simulation-only,
  no credentials extracted.
tactic: credential-access
technique_id: T1003.001
platform: windows
phase: attack
elevation_required: true
expected_events:
  - event_id: 10
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "ProcessAccess - TargetImage: lsass.exe GrantedAccess: 0x1010"
  - event_id: 4656
    channel: "Security"
    description: "Handle request to lsass.exe PROCESS_VM_READ"
siem_coverage:
  crowdstrike:
    - "Credential Dumping"
    - "Suspicious LSASS Access"
tags:
  - credential-access
  - lsass
  - requires-elevation
  - falcon-targeted
executor:
  type: powershell
  command: |-
    # ... multi-variant safe simulation ...
cleanup: ""
```

**ID collision risk:** The `FALCON_lsass_access` technique uses `technique_id: T1003.001` but has `id: FALCON_lsass_access`. The Registry indexes by `t.ID` (loader.go line 42), so no collision with existing `T1003.001` technique. Confirmed safe.

### Anti-Patterns to Avoid

- **Do not use `tactic: "crowdstrike-falcon"`:** The engine filters by tactic. An unknown tactic value means `filterByTactics` will include it only when no filter is active — but it creates a non-standard tactic that breaks the tactic stats in the HTML report. Use MITRE standard tactic names.
- **Do not thread the registry through `saveHTML()`:** Avoid changing `SaveResults` signature. Use ExecutionResult field propagation (Option A) instead.
- **Do not access map with dot notation in Go templates:** Use `index .SIEMCoverage "crowdstrike"` not `.SIEMCoverage.crowdstrike`.
- **Do not add `siem_coverage` to techniques where Falcon does not genuinely fire:** The CONTEXT.md D-02 is explicit — omit rather than speculate.
- **Do not duplicate FALCON_ vs existing techniques at the YAML id level:** `FALCON_` prefix in the `id` field prevents registry key collision.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| YAML `map[string][]string` parsing | Custom YAML decoder | `gopkg.in/yaml.v3` native | Already handles this type out of the box |
| HTML template map access | Custom funcMap for map lookup | `index` built-in template function | `{{index .SIEMCoverage "crowdstrike"}}` works natively |
| CrowdStrike name canonicalization | Normalization logic | Verbatim strings from prevention policy | Just use the exact string from the Falcon console |
| Technique auto-registration | Registration map or init() | `loader.go` fs.WalkDir embed | Drop YAML in embedded/techniques/ — automatically discovered |

---

## CrowdStrike Detection Names — Research Findings

### What "Detection Names" Mean in Falcon

CrowdStrike Falcon has two layers of detection nomenclature:

1. **Prevention Policy features** — toggle names in the Falcon console UI under "Prevention Policies → Windows". These appear as the feature label (e.g., "Credential Dumping"). When a detection fires due to this policy feature, the alert in the console references the policy feature name. These names are stable, officially documented via the Terraform/Pulumi registry, and directly visible to administrators.

2. **Behavioral IOA alert names** — Names shown in the Alerts page after a detection fires. These are not formally published as a public enumeration. CrowdStrike uses names like "SuspiciousMshta", "Known Malware", and similar patterns, but the complete taxonomy is not publicly indexed. Individual names appear in blog posts and case studies.

**Research conclusion:** Use Prevention Policy feature names as the `siem_coverage.crowdstrike` values. These are what a consultant configuring or reviewing the client's Falcon policy will immediately recognize. They directly correspond to the detection capabilities that fired.

### Verified Prevention Policy Feature Names (HIGH confidence — Pulumi Registry)

Source: https://www.pulumi.com/registry/packages/crowdstrike/api-docs/preventionpolicywindows/

| Falcon Feature Name | Applicable Techniques | Falcon Sensor Trigger |
|--------------------|-----------------------|-----------------------|
| `"Credential Dumping"` | T1003.001, FALCON_lsass_access | Kills processes reading LSASS; requires additional_user_mode_data |
| `"Code Injection"` | FALCON_process_injection | Kills processes injecting code into another process; requires additional_user_mode_data |
| `"Suspicious Scripts and Commands"` | T1059.001, T1059.003 | Blocks script-based and PowerShell-based threats |
| `"Interpreter-Only"` | T1059.001 | Visibility into malicious PowerShell interpreter usage |
| `"Engine (Full Visibility)"` | T1059.001 | Full visibility into PS SMA engine usage by any app |
| `"Script-Based Execution Monitoring"` | T1059.001, T1027 | Shell and PowerShell script content monitoring (Windows 10+/Server 2016+) |
| `"Windows Logon Bypass (Sticky Keys)"` | T1548.002 | Blocks sticky keys / logon bypass |
| `"Suspicious Registry Operations"` | T1547.001 | Blocks suspicious ASEP/security config registry changes |

### MEDIUM Confidence — Behavioral IOA Names (derived from documentation, case studies)

These names appear in CrowdStrike blog posts, incident reports, and case studies but are not in a formal public registry. Treat as MEDIUM confidence — they reflect actual Falcon alert taxonomy but may vary across Falcon versions.

| Detection Name | Source | Confidence |
|----------------|--------|------------|
| `"Suspicious LSASS Access"` | Referenced in LSASS analysis blog posts, bypass writeups | MEDIUM |
| `"Lateral Movement - PsExec"` | Referenced in lateral movement detection docs | MEDIUM |
| `"Malicious PowerShell Interpreter"` | Referenced in PowerShell detection blog posts | MEDIUM |
| `"Credential Access via Credential Dumping"` | CrowdStrike detection-container GitHub labels | MEDIUM |

**Recommendation:** For the 3 FALCON_ techniques, use Prevention Policy feature names as the primary identifiers (HIGH confidence). Optionally add one MEDIUM-confidence behavioral name per technique to give consultants a second search term for the alerts page.

### Existing Technique Mapping Recommendations (Claude's Discretion)

The following existing techniques have clear Falcon prevention policy triggers. Mapping is recommended where the technique's behavior directly corresponds to a named policy feature:

| Technique ID | Name | Recommended `siem_coverage.crowdstrike` |
|---|---|---|
| T1059.001 | PowerShell Execution | `["Suspicious Scripts and Commands", "Interpreter-Only"]` |
| T1059.003 | Cmd Shell | `["Suspicious Scripts and Commands"]` |
| T1003.001 | LSASS Memory Access | `["Credential Dumping"]` |
| T1547.001 | Registry Persistence | `["Suspicious Registry Operations"]` |
| T1548.002 | UAC Bypass | `["Windows Logon Bypass (Sticky Keys)"]` (partial — depends on technique) |
| T1027 | Obfuscated Commands | `["Suspicious Scripts and Commands", "Script-Based Execution Monitoring"]` |
| T1562.002 | Disable Logging | `["Suspicious Scripts and Commands"]` |
| T1218.011 | Rundll32 | `["Javascript via Rundll32"]` (for JS variant) |
| T1543.003 | New Service | `["Suspicious Registry Operations"]` (service key) |
| T1134.001 | Token Impersonation | `["Code Injection"]` (requires additional_user_mode_data) |

**Techniques with NO Falcon mapping (omit siem_coverage):**

Discovery techniques (T1016, T1049, T1057, T1069, T1082, T1083, T1087, T1135, T1482) — these are benign enumeration activities that do NOT trigger Falcon prevention policies by default. Mapping these would be misleading. UEBA scenarios similarly do not map to Falcon detections.

---

## Common Pitfalls

### Pitfall 1: FALCON_ Technique Duplicates T1003.001 Behavior

**What goes wrong:** `FALCON_lsass_access` and `T1003.001` both use `OpenProcess(0x1010)` against lsass.exe. Running both techniques in the same session generates indistinguishable Sysmon EID 10 events — SIEM operators cannot tell which technique fired which alert.

**Why it happens:** The existing T1003.001 already implements Method 2 with `OpenProcess(0x1010)`. FALCON_lsass_access risks being an exact duplicate.

**How to avoid:** `FALCON_lsass_access` should use a **different** access pattern — specifically focusing on the `comsvcs.dll MiniDump` path (which T1003.001 already does as Method 1) OR targeting `VirtualAllocEx`/`ReadProcessMemory` directly. Better differentiation: FALCON_lsass_access can use the `NtReadVirtualMemory` syscall pattern that triggers Falcon's usermode data collection. Document the distinction explicitly in the YAML description.

**Warning signs:** If the YAML commands for `FALCON_lsass_access` look identical to `T1003.001`, it needs redesign.

### Pitfall 2: Go Template Map Access Syntax

**What goes wrong:** Writing `{{.SIEMCoverage.crowdstrike}}` in the HTML template causes a template execution panic at runtime — Go templates do not support dot-notation map key access.

**Why it happens:** Developers familiar with other templating languages (Jinja2, Handlebars) expect dot notation to work for map keys.

**How to avoid:** Always use `{{index .SIEMCoverage "crowdstrike"}}`. Write a unit test (see Validation Architecture) that renders an HTML report with a result containing `SIEMCoverage` populated and asserts no template execution error.

**Warning signs:** Template panics during test runs; HTML file contains "can't evaluate field crowdstrike in type map[string][]string".

### Pitfall 3: HasCrowdStrike Computed but SIEMCoverage Not Propagated

**What goes wrong:** `HasCrowdStrike` is set to `true` in `htmlData` but the per-result `SIEMCoverage` field is nil because the engine never copied it from the Technique into ExecutionResult.

**Why it happens:** The engine's `runTechnique()` builds `ExecutionResult` without the new `SIEMCoverage` field — easy to miss since it's optional.

**How to avoid:** The engine must explicitly set `result.SIEMCoverage = t.SIEMCoverage` for ALL code paths in `runTechnique()`: the real executor path, the WhatIf path, and the injected runner path. Three code paths, all must propagate.

**Warning signs:** HTML renders the CS column header but all rows show N/A even for techniques with `siem_coverage` in their YAML.

### Pitfall 4: FALCON_ Technique Uses Non-Standard Tactic Value

**What goes wrong:** Setting `tactic: crowdstrike-falcon` means the tactic stats section in the HTML report shows a "crowdstrike-falcon" tactic row with no color match, and `filterByTactics` treats it as an unknown tactic.

**Why it happens:** D-03 in CONTEXT says "check how the engine filters by tactic" — the answer is that filterByTactics uses exact string match on `t.Tactic`. Using a non-MITRE tactic name requires all users to explicitly include it via `IncludedTactics` to run it.

**How to avoid:** Use standard MITRE tactic names. The tactic color map in reporter.go (line 179–193) already covers `credential-access`, `lateral-movement`, `defense-evasion` — all valid targets for FALCON_ techniques.

### Pitfall 5: SIEMCoverage Field Missing from ExecutionResult JSON Report

**What goes wrong:** Adding `SIEMCoverage` to `Technique` but not to `ExecutionResult` means the JSON report (lognojutsu_report_*.json) omits the field. Future tooling expecting coverage data in the JSON output will find nothing.

**Why it happens:** Types in two different places (`Technique` for the playbook definition, `ExecutionResult` for the run record) need to be updated independently.

**How to avoid:** The recommended Option A approach above adds `SIEMCoverage` to `ExecutionResult` explicitly. This keeps the JSON report as the single source of truth.

---

## Code Examples

### 1. SIEMCoverage Field in Technique Struct

```go
// Source: internal/playbooks/types.go — follow existing alignment pattern
// Insert after NistControls field (currently line 45)
SIEMCoverage   map[string][]string `yaml:"siem_coverage,omitempty"  json:"siem_coverage,omitempty"`
```

Note: count spaces to align with `NistControls` tag column. Check existing alignment in file.

### 2. SIEMCoverage Field in ExecutionResult

```go
// Source: internal/playbooks/types.go — ExecutionResult struct
// Insert after VerifyTime (currently line 84)
SIEMCoverage   map[string][]string `json:"siem_coverage,omitempty"`
```

### 3. Engine Propagation in runTechnique()

```go
// Source: internal/engine/engine.go — all three code paths in runTechnique()
// Add after the result variable is populated in each branch:
result.SIEMCoverage = t.SIEMCoverage
```

The three branches are:
- WhatIf branch (line ~452): result literal is built inline, then add the field
- runner != nil branch (line ~464): result returned from runner func, then add the field
- Real executor branches (lines ~471-479): result returned from executor, then add the field

### 4. HasCrowdStrike Computation in saveHTML()

```go
// Source: internal/reporter/reporter.go — add in saveHTML() after verifFailed loop
hasCrowdStrike := false
for _, res := range r.Results {
    if len(res.SIEMCoverage["crowdstrike"]) > 0 {
        hasCrowdStrike = true
        break
    }
}
// Add HasCrowdStrike to htmlData struct and data literal
```

### 5. CSS Badge Classes (append to existing <style> block)

```css
/* CrowdStrike coverage column */
.cs-badge{background:#e01b22;color:#fff;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}
.cs-na{color:#8b949e;font-size:11px}
.cs-list{font-size:11px;margin-top:4px;padding-left:0;list-style:none}
.cs-list li{margin:1px 0;color:#e6edf3}
```

Colors: `#e01b22` is CrowdStrike's brand red. Consistent with the dark theme of the existing report.

### 6. HTML Template — Conditional Column

```html
<!-- In <thead><tr> after existing <th>Verifikation</th> -->
{{if .HasCrowdStrike}}<th>CrowdStrike</th>{{end}}

<!-- In {{range .Results}}<tr> after existing verification </td> -->
{{if $.HasCrowdStrike}}
<td>
  {{if index .SIEMCoverage "crowdstrike"}}
    <span class="cs-badge">CS</span>
    <ul class="cs-list">
      {{range index .SIEMCoverage "crowdstrike"}}<li>{{.}}</li>{{end}}
    </ul>
  {{else}}
    <span class="cs-na">N/A</span>
  {{end}}
</td>
{{end}}
```

### 7. YAML siem_coverage Block Pattern

```yaml
# Add to existing technique YAML (e.g., T1059_001_powershell.yaml)
siem_coverage:
  crowdstrike:
    - "Suspicious Scripts and Commands"
    - "Interpreter-Only"
```

---

## State of the Art

| Old Approach | Current Approach | Impact |
|---|---|---|
| Hard-coded SIEM names in description fields | Structured `siem_coverage` map field | Machine-readable; Phase 5 (Sentinel) adds `sentinel` key without struct change |
| Manual report correlation | Inline column in existing results table | Consultant can compare execution result + SIEM detection rule in one row |

**Deprecated/outdated:** The Falcon `Detects` API collection is deprecated as of September 2025 per FalconPy docs. The replacement is the `Alerts` service collection. This is irrelevant to LogNoJutsu (we don't call Falcon APIs), but is context for why some detection name documentation references old API field names.

---

## Open Questions

1. **FALCON_lsass_access vs T1003.001 — Differentiation Strategy**
   - What we know: Both target LSASS. T1003.001 uses `OpenProcess(0x1010)` and `comsvcs.dll MiniDump`.
   - What's unclear: What access pattern does `FALCON_lsass_access` use that T1003.001 doesn't, to justify a separate technique?
   - Recommendation: Use `NtReadVirtualMemory` direct syscall pattern (C# via P/Invoke or PowerShell Add-Type) which targets Falcon's usermode data collection path specifically. Alternatively, focus on `MiniDumpWriteDump` via a custom process instead of rundll32/comsvcs — a different parent process spawning a dump generates a distinct Sysmon EID 10 entry that consultants can distinguish in the report.

2. **Falcon Alert Names in Console — MEDIUM Confidence**
   - What we know: Prevention policy feature names ("Credential Dumping", "Code Injection") are verified HIGH confidence. Console IOA alert names like "Suspicious LSASS Access" appear in blog posts but are not in a public registry.
   - What's unclear: Whether "Suspicious LSASS Access" and "Lateral Movement - PsExec" are the exact console-visible strings vs variant names like "SuspiciousLsassAccess" or "Credential Access via Credential Dumping".
   - Recommendation: Use Prevention Policy names as primary identifiers (they are the definitive toggle labels clients see). Add a YAML comment on FALCON_ techniques noting that console alert names may vary by Falcon version/policy configuration.

3. **Which existing 10-14 techniques to map**
   - What we know: ~10 techniques have clear Falcon prevention policy matches (see mapping table above).
   - What's unclear: T1134.001 (Token Impersonation) requires `additional_user_mode_data` — if the client hasn't enabled it, the detection won't fire. Should techniques with conditional coverage be mapped?
   - Recommendation: Map them with a note in the YAML (as a description annotation). The consultant report should show the detection name exists; whether it fires is a policy configuration question, not a mapping question.

---

## Environment Availability

Step 2.6: SKIPPED — This phase is code/config/YAML changes with no external tool dependencies. The FALCON_ technique commands use Windows built-ins (PowerShell, rundll32, net.exe) identical to patterns already used in existing techniques.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go standard testing (`testing` package) |
| Config file | None — `go test ./...` convention |
| Quick run command | `go test ./internal/playbooks/... ./internal/reporter/...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CROW-01 | `SIEMCoverage` field parses from YAML correctly | unit | `go test ./internal/playbooks/... -run TestSIEMCoverage` | ❌ Wave 0 |
| CROW-01 | FALCON_ techniques have non-empty `siem_coverage.crowdstrike` | unit | `go test ./internal/playbooks/... -run TestFalconTechniques` | ❌ Wave 0 |
| CROW-02 | 3 FALCON_ technique files load and have expected_events | unit | `go test ./internal/playbooks/... -run TestFalconTechniques` | ❌ Wave 0 |
| CROW-02 | Technique count includes 3 new FALCON_ files | unit | `go test ./internal/playbooks/... -run TestNewTechniqueCount` | ✅ (needs threshold bump) |
| CROW-03 | HTML report contains CS column when SIEMCoverage present | unit | `go test ./internal/reporter/... -run TestHTMLCrowdStrikeColumn` | ❌ Wave 0 |
| CROW-03 | HTML report omits CS column when no SIEMCoverage present | unit | `go test ./internal/reporter/... -run TestHTMLCrowdStrikeColumn` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/playbooks/... ./internal/reporter/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/playbooks/loader_test.go` — add `TestSIEMCoverage` (verify map[string][]string parses from YAML), `TestFalconTechniques` (3 FALCON_ files exist, have siem_coverage, have expected_events)
- [ ] `internal/playbooks/loader_test.go` — bump `TestNewTechniqueCount` threshold from 48 to 51 (adds 3 FALCON_ files)
- [ ] `internal/reporter/reporter_test.go` — add `TestHTMLCrowdStrikeColumn` (present when SIEMCoverage populated, absent when not)
- [ ] No new framework install needed — `testing` stdlib already in use across all packages

---

## Sources

### Primary (HIGH confidence)

- Pulumi CrowdStrike Provider — `PreventionPolicyWindows` resource schema — exact feature names: https://www.pulumi.com/registry/packages/crowdstrike/api-docs/preventionpolicywindows/
- `internal/playbooks/types.go` — current Technique and ExecutionResult struct layout
- `internal/reporter/reporter.go` — current htmlData struct, template pattern, badge CSS
- `internal/engine/engine.go` — runTechnique() code paths (WhatIf, runner, real executor)
- `internal/playbooks/loader.go` — registry load-by-id pattern (confirms FALCON_ ID safety)
- `internal/playbooks/loader_test.go` — existing test patterns (TestNewTechniqueCount threshold)
- `internal/reporter/reporter_test.go` — existing saveHTMLToDir helper pattern
- Go `text/template` docs — `index` function for map access in templates (stdlib, HIGH)

### Secondary (MEDIUM confidence)

- CrowdStrike Prevention Policy blog (alirodoplu.com) — console feature names cross-validated against Pulumi schema
- CrowdStrike detection-container GitHub — behavioral detection label examples
- Sekoia.io CrowdStrike Falcon integration docs — alert field structure examples
- FalconPy Detects API README — detection tactic/technique classification examples

### Tertiary (LOW confidence)

- WebSearch results for "Suspicious LSASS Access", "Lateral Movement - PsExec" alert names — appeared in bypass writeups and blog posts but not in an official CrowdStrike detection name registry

---

## Metadata

**Confidence breakdown:**

- Standard stack (Go struct + YAML + template): HIGH — zero new dependencies, all patterns directly in codebase
- Architecture (SIEMCoverage propagation via ExecutionResult): HIGH — based on direct code analysis
- Falcon prevention policy feature names: HIGH — verified via official Terraform/Pulumi registry schema
- Falcon console IOA alert names: MEDIUM — derived from blog posts and documentation, not a public enumeration
- FALCON_ technique content (which behaviors to trigger): MEDIUM — based on Falcon detection documentation and policy feature descriptions, not hands-on Falcon console access

**Research date:** 2026-03-25
**Valid until:** 2026-06-25 (stable — Falcon prevention policy names change infrequently; template/Go patterns are stable)
