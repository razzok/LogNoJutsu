# Phase 5: Microsoft Sentinel Coverage - Research

**Researched:** 2026-03-25
**Domain:** Microsoft Sentinel detection rule names, AZURE_ YAML technique authoring, Go HTML template extension
**Confidence:** HIGH (codebase analysis, official Sentinel GitHub) / MEDIUM (LDAP rule naming)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-01 (locked from Phase 4):** `SIEMCoverage map[string][]string` is already on the `Technique` struct with YAML key `siem_coverage`. Phase 5 requires zero struct changes — just populate `siem_coverage.sentinel` in YAML files. The `sentinel` map key is already supported by the existing data model.

**D-02 (locked from Phase 4):** Only map where Sentinel genuinely fires. Techniques with no Sentinel mappings omit `siem_coverage.sentinel` entirely. Keep YAML files clean.

**D-03:** Create 3 new YAML technique files with `AZURE_` prefix targeting Kerberos/LDAP attack patterns. These run locally on Windows, generate Windows Security Events, and are picked up by Sentinel's AMA/MMA agents — zero Azure connectivity required at runtime. Targets:
1. **AZURE_kerberoasting** — Request RC4-encrypted Kerberos service tickets (T1558.003); generates Windows Security Event 4769. Triggers Sentinel "Potential Kerberoasting" analytic rule.
2. **AZURE_ldap_recon** — LDAP directory enumeration (T1087.002 / T1018); generates Windows Event 1644 (expensive LDAP query). Triggers Sentinel "LDAP Query Reconnaissance" analytic rule.
3. **AZURE_dcsync** — Simulate DCSync directory replication access (T1003.006); generates Windows Security Event 4662 (DS-Replication-Get-Changes). Triggers Sentinel "Potential DCSync Attack" analytic rule.

**D-04:** Each AZURE_ technique includes `expected_events` entries for the Windows Security Events generated AND `siem_coverage.sentinel` with the official Sentinel analytic rule names.

**D-05 (mirrors Phase 4 D-05/D-06):** Sentinel column shows: blue **MS** badge (#0078D4) + list of analytic rule names when `siem_coverage.sentinel` is populated. Grey **N/A** cell when no Sentinel mappings exist for that technique.

**D-06 (mirrors Phase 4 D-06):** The Sentinel column is **conditional** — renders only when at least one technique in the results has `siem_coverage.sentinel` populated. Matches SENT-03 wording ("when Azure events are present") and keeps the report clean on non-Sentinel environments.

**D-07:** Badge text is `MS`, background color `#0078D4` (Microsoft blue), white text. CSS classes follow the `cs-badge`/`cs-na`/`cs-list` naming pattern — use `ms-badge`/`ms-na`/`ms-list`. Rendered inside `{{if .HasSentinel}}` conditional — absent HTML has zero vendor-specific markup.

**D-08 (mirrors Phase 4 D-08):** Use **official Microsoft Sentinel analytic rule names** as they appear in the Sentinel portal. These are the exact strings consultants will see in the client's Sentinel workspace — makes the report directly actionable.

### Claude's Discretion

- Specific official Sentinel analytic rule names to map to each existing technique — research actual Sentinel built-in rule names
- Which existing techniques (out of 52) merit Sentinel mappings — only map where Sentinel genuinely fires
- Exact PowerShell/cmd commands in AZURE_ techniques — must generate the expected Windows Security Events safely (no real attacks), following the multi-variant pattern from Phase 3
- `HasSentinel bool` flag on `htmlData` struct — compute in `SaveResults()` by scanning for any non-empty `SIEMCoverage["sentinel"]` slice, same pattern as `HasCrowdStrike`
- Where to add Sentinel documentation in the German README — extend existing README following Phase 4's documentation pattern

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SENT-01 | Microsoft Sentinel detection rule mappings documented per technique in events manifest | Covered by: `siem_coverage.sentinel` key in existing `SIEMCoverage map[string][]string` field — no struct changes; populate in 3 AZURE_ YAMLs + select existing techniques where Sentinel genuinely fires |
| SENT-02 | At least 3 techniques that target Azure AD / Microsoft Defender / Sentinel log sources | Covered by: 3 new AZURE_ YAML files, each generating Windows Security Events that Sentinel AMA/MMA agents forward to Azure — generates Security channel events (4662, 4769, 1644) that Sentinel built-in rules query |
| SENT-03 | HTML report shows Sentinel-specific coverage column when Azure events are present | Covered by: `HasSentinel bool` in `htmlData`, conditional `{{if .HasSentinel}}` block in HTML template, `ms-badge`/`ms-na`/`ms-list` CSS classes |
</phase_requirements>

---

## Summary

Phase 5 is structurally identical to Phase 4 — a data + presentation phase. All Go types are already correct: `SIEMCoverage map[string][]string` on both `Technique` and `ExecutionResult` is fully operational from Phase 4. The only code change is one new field in `htmlData` (`HasSentinel bool`), one new scan loop in `saveHTML()`, and the Sentinel column block in the HTML template (mirrors the CrowdStrike block exactly with different key and CSS classes). Three new AZURE_ YAML files provide the SENT-02 technique count. A small number of existing techniques get `siem_coverage.sentinel` entries for SENT-01 coverage breadth.

The hardest problem is **official Sentinel analytic rule naming** — these are what consultants see in the Sentinel portal, and accuracy matters for SENT-01. Research confirmed two official names directly from the Azure/Azure-Sentinel GitHub repository (HIGH confidence): "Potential Kerberoasting" (from `PotentialKerberoast.yaml`) and "Non Domain Controller Active Directory Replication" (from `NonDCActiveDirectoryReplication.yaml`). The DCSync rule is the authoritative name for T1003.006-style simulation; CONTEXT.md D-03 uses "Potential DCSync Attack" as a shorthand but the actual rule is named "Non Domain Controller Active Directory Replication". The LDAP recon rule is MEDIUM confidence — no exact match found in the Detections/SecurityEvent directory for Event 1644; the closest verified name is from the Defender for Identity analytics layer. YAML files should use a comment to flag MEDIUM confidence names.

The three AZURE_ techniques all run locally against AD without any Azure cloud connectivity. They generate authentic Windows Security Events that, in a Sentinel-connected environment, would be forwarded by the AMA/MMA agent and trigger the mapped analytic rules. The technique IDs and tactic assignments are distinct from existing T1558.003, T1003.006, and T1087 entries — the AZURE_ prefix marks them as Sentinel-optimized variants with different command profiles designed to maximize Sentinel rule trigger probability (e.g., RC4 downgrade path for EID 4769, broad LDAP filter for EID 1644, DS-Replication GUID access for EID 4662).

**Primary recommendation:** Mirror Phase 4 exactly. Use `HasSentinel` / `ms-badge` / `ms-list` / `ms-na` naming. Use verified Sentinel rule names from GitHub where available, flag MEDIUM confidence names inline in YAML comments.

---

## Standard Stack

### Core (zero new dependencies)

| Component | Version | Purpose | Notes |
|-----------|---------|---------|-------|
| `gopkg.in/yaml.v3` | in go.mod | YAML unmarshalling for `siem_coverage.sentinel` | Already present — `map[string][]string` natively handled |
| `text/template` | stdlib | HTML conditional Sentinel column | Already used in reporter.go |
| Go stdlib | go 1.26.1 | No new imports | Phase 4 set everything up |

**Installation:** No new packages. This phase is zero-dependency from a library perspective.

**Version verification:** Confirmed from go.mod — `gopkg.in/yaml.v3` is already present in the project.

---

## Architecture Patterns

### Files Changed / Created

```
internal/
├── playbooks/
│   └── embedded/techniques/
│       ├── AZURE_kerberoasting.yaml          # NEW (SENT-02)
│       ├── AZURE_ldap_recon.yaml             # NEW (SENT-02)
│       ├── AZURE_dcsync.yaml                 # NEW (SENT-02)
│       ├── T1003_001_lsass.yaml              # MODIFY: add siem_coverage.sentinel
│       ├── T1003_006_dcsync.yaml             # MODIFY: add siem_coverage.sentinel
│       ├── T1558_003_kerberoasting.yaml      # MODIFY: add siem_coverage.sentinel
│       ├── T1059_001_powershell.yaml         # MODIFY: add siem_coverage.sentinel
│       └── T1136_001_create_local_account.yaml # MODIFY: add siem_coverage.sentinel
└── reporter/
    └── reporter.go                           # MODIFY: HasSentinel flag, CSS, template block
README.md                                    # MODIFY: Sentinel section in German
```

### Pattern 1: YAML siem_coverage.sentinel Block

**What:** Identical structure to `siem_coverage.crowdstrike` from Phase 4. Key is `sentinel`, value is a YAML sequence of official Sentinel analytic rule names.

```yaml
# Source: CONTEXT.md D-08, mirrors Phase 4 pattern
siem_coverage:
  sentinel:
    - "Potential Kerberoasting"
    - "Non Domain Controller Active Directory Replication"
```

**When to omit:** If no verified Sentinel detection applies to the technique, omit the `sentinel` key entirely. Never add speculative entries. The `omitempty` tag on the struct field means absent YAML = nil = no rendering.

**Co-presence:** A technique can have both `crowdstrike` and `sentinel` keys in the same `siem_coverage` block. The `map[string][]string` type handles this natively.

```yaml
siem_coverage:
  crowdstrike:
    - "Credential Dumping"
  sentinel:
    - "Dumping LSASS Process Into a File"
```

### Pattern 2: HasSentinel Flag in htmlData

**What:** Add `HasSentinel bool` field to `htmlData` struct (line 95–96 in reporter.go, after `HasCrowdStrike bool`). Compute in `saveHTML()` by scanning results, identical logic to `hasCrowdStrike` scan.

```go
// reporter.go — htmlData struct addition (after HasCrowdStrike line)
HasSentinel bool

// reporter.go — saveHTML() — scan loop (after hasCrowdStrike block)
hasSentinel := false
for _, res := range r.Results {
    if len(res.SIEMCoverage["sentinel"]) > 0 {
        hasSentinel = true
        break
    }
}

// reporter.go — data := htmlData{...} — add field
HasSentinel: hasSentinel,
```

### Pattern 3: HTML Template Sentinel Column

**What:** Add Sentinel column header and row cell after the CrowdStrike column block. Mirrors Phase 4 pattern exactly with `sentinel` key and `ms-*` CSS classes.

```html
<!-- Column header — add after {{if .HasCrowdStrike}}<th>CrowdStrike</th>{{end}} -->
{{if .HasSentinel}}<th>Microsoft Sentinel</th>{{end}}

<!-- Row cell — add after the CrowdStrike {{end}} block -->
{{if $.HasSentinel}}
<td>
  {{$ms := siemCoverage .SIEMCoverage "sentinel"}}
  {{if $ms}}
    <span class="ms-badge">MS</span>
    <ul class="ms-list">
      {{range $ms}}<li>{{.}}</li>{{end}}
    </ul>
  {{else}}
    <span class="ms-na">N/A</span>
  {{end}}
</td>
{{end}}
```

**Note:** The existing `siemCoverage` funcMap helper in reporter.go (line 181–186) already performs nil-safe map access — it works for any key including `"sentinel"`. No changes to the funcMap needed.

### Pattern 4: CSS Classes

**What:** Add Sentinel CSS classes inside the existing `{{if .HasCrowdStrike}}` CSS block, or as a separate `{{if .HasSentinel}}` CSS block. Keep vendor CSS conditional so the HTML is clean on non-Sentinel runs.

```css
{{if .HasSentinel}}.ms-badge{background:#0078D4;color:#fff;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}
.ms-na{color:#8b949e;font-size:11px}
.ms-list{font-size:11px;margin-top:4px;padding-left:0;list-style:none}
.ms-list li{margin:1px 0;color:#e6edf3}{{end}}
```

Add this immediately after the `{{if .HasCrowdStrike}}...{{end}}` CSS block (reporter.go lines 269–272).

### Pattern 5: AZURE_ Technique YAML Schema

**What:** New AZURE_ technique files follow the same schema as FALCON_ files from Phase 4. Key differences from existing T1558.003/T1003.006:
- `id` uses `AZURE_` prefix
- `siem_coverage.sentinel` populated with official rule names
- Commands target distinct code paths that maximize Sentinel detection probability
- `tags` includes `sentinel-targeted` to mark intent
- Multi-variant commands (3–5 methods) for signal diversity

```yaml
id: AZURE_kerberoasting
name: Sentinel - Potential Kerberoasting Detection
description: >
  ...
tactic: credential-access
technique_id: T1558.003
platform: windows
phase: attack
elevation_required: false
expected_events:
  - event_id: 4769
    channel: "Security"
    description: "Kerberos TGS request - TicketEncryptionType=0x17 RC4-HMAC — Sentinel Potential Kerberoasting trigger"
  - event_id: 4688
    channel: "Security"
    description: "setspn.exe SPN enumeration"
tags:
  - credential-access
  - active-directory
  - sentinel-targeted
siem_coverage:
  sentinel:
    - "Potential Kerberoasting"
executor:
  type: powershell
  command: |-
    ...
cleanup: ""
```

### Anti-Patterns to Avoid

- **Speculative Sentinel rule names:** Never add a rule name that was not verified from the Azure/Azure-Sentinel GitHub repository or an official Microsoft source. Flag MEDIUM confidence names in YAML comments.
- **Duplicate technique logic:** AZURE_ techniques must differ meaningfully from existing T1558.003, T1003.006, and T1087 entries. Different command sequences, different access patterns, different Sentinel-optimized signal paths.
- **Using `siemCoverage .SIEMCoverage "sentinel"` outside the `{{if $.HasSentinel}}` guard:** The guard protects the table column structure — accessing the map without it would add empty cells on non-Sentinel runs.
- **CSS outside conditional block:** All `ms-*` CSS must be inside `{{if .HasSentinel}}...{{end}}` to prevent vendor markup leaking into clean reports.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Nil-safe map access in template | Custom map lookup function | Existing `siemCoverage` funcMap helper | Already handles nil map; reuse directly with key `"sentinel"` |
| HTML template language detection | Runtime string scan | `HasSentinel bool` pre-computed in `saveHTML()` | Template cannot efficiently iterate all results; pre-compute flag in Go |
| Sentinel rule names | Manual guessing | Azure/Azure-Sentinel GitHub + analyticsrules.exchange | Rule names change; must match portal exactly for consultant usability |

**Key insight:** The siemCoverage funcMap helper already exists and handles `"sentinel"` key just as well as `"crowdstrike"`. Zero new Go code for template data access.

---

## Official Sentinel Analytic Rule Names (Verified)

This table drives both YAML population (SENT-01) and AZURE_ technique mapping (SENT-02).

### HIGH Confidence (verified from Azure/Azure-Sentinel GitHub)

| Technique | MITRE ID | Official Sentinel Rule Name | Source File |
|-----------|----------|-----------------------------|-------------|
| Kerberoasting | T1558.003 | `"Potential Kerberoasting"` | `Detections/SecurityEvent/PotentialKerberoast.yaml` |
| DCSync (non-DC replication) | T1003.006 | `"Non Domain Controller Active Directory Replication"` | `Solutions/Windows Security Events/Analytic Rules/NonDCActiveDirectoryReplication.yaml` |
| LSASS dump to file | T1003.001 | `"Dumping LSASS Process Into a File"` | `Detections/SecurityEvent/DumpingLSASSProcessIntoaFile.yaml` |

**Verification:** All three names confirmed from raw YAML `name:` fields at Azure/Azure-Sentinel master branch.

### MEDIUM Confidence (inferred / indirect sources)

| Technique | MITRE ID | Candidate Rule Name | Basis |
|-----------|----------|--------------------|-------|
| PowerShell suspicious execution | T1059.001 | `"Suspicious Powershell Commandlet Executed"` | analyticsrules.exchange listing |
| Create local account | T1136.001 | `"User Created and Added to Built-in Administrators"` | analyticsrules.exchange listing (rule exists for admin group addition) |
| LDAP recon (Event 1644) | T1087.002 / T1018 | No confirmed dedicated built-in rule for raw EID 1644 | MDI handles LDAP recon; no SecurityEvent YAML found for EID 1644 specifically |

**LDAP finding:** Sentinel's SecurityEvent detection directory does not contain a dedicated analytic rule for Windows Event 1644. LDAP reconnaissance detection in Sentinel is primarily handled through Microsoft Defender for Identity (MDI) analytics — the `IdentityQueryEvents` table rather than the `SecurityEvent` table. AZURE_ldap_recon should target the MDI-aware vector (EID 4688 for LDAP tools) alongside EID 1644 as a supporting event, and use a conservative Sentinel rule name that reflects what actually fires. See Open Questions.

---

## Existing Techniques: Sentinel Mapping Candidates

Out of 55 existing techniques, Sentinel has verified built-in coverage for:

| Technique File | Technique | Sentinel Rule to Map |
|---------------|-----------|---------------------|
| `T1558_003_kerberoasting.yaml` | Kerberoasting | `"Potential Kerberoasting"` (HIGH) |
| `T1003_006_dcsync.yaml` | DCSync | `"Non Domain Controller Active Directory Replication"` (HIGH) |
| `T1003_001_lsass.yaml` | LSASS dump | `"Dumping LSASS Process Into a File"` (HIGH) |
| `T1059_001_powershell.yaml` | PowerShell execution | `"Suspicious Powershell Commandlet Executed"` (MEDIUM) |
| `T1136_001_create_local_account.yaml` | Create local account | `"User Created and Added to Built-in Administrators"` (MEDIUM) |

Techniques deliberately excluded (benign enumeration, no Sentinel rule fires by default):
- Discovery techniques (T1016, T1046, T1049, T1057, T1069, T1082, T1083, T1087, T1135) — Discovery enumeration does not trigger Sentinel prevention analytics
- UEBA techniques — Sentinel does not have built-in behavioral UEBA rules matching these patterns
- T1548.002 UAC bypass, T1574.002 DLL sideloading — Sentinel coverage is via Defender for Endpoint, not Windows Security Events

---

## Common Pitfalls

### Pitfall 1: Wrong DCSync Rule Name
**What goes wrong:** Using "Potential DCSync Attack" (from CONTEXT.md D-03 shorthand) as the Sentinel rule name — this string does not appear in the Sentinel portal.
**Why it happens:** CONTEXT.md used a shorthand description, not the official portal name.
**How to avoid:** Use the verified name `"Non Domain Controller Active Directory Replication"` from GitHub. The YAML file for AZURE_dcsync should include a comment: `# Official Sentinel rule: "Non Domain Controller Active Directory Replication"`.
**Warning signs:** If the planner/implementer copies the CONTEXT.md shorthand verbatim.

### Pitfall 2: AZURE_ Technique Duplicates Existing Technique Logic
**What goes wrong:** AZURE_kerberoasting becomes a copy of T1558_003_kerberoasting with just `siem_coverage.sentinel` added. The existing technique already has `expected_events` for EID 4769. The AZURE_ variant adds no incremental value.
**Why it happens:** Phase 3 added T1558.003 with a comprehensive command set; phase 5 must produce a distinct Sentinel-optimized variant.
**How to avoid:** AZURE_kerberoasting must use a different execution path — target volume threshold trigger (15+ distinct SPN requests in 1 hour as per the rule's KQL logic), not just one RC4 ticket. The existing T1558.003 already requests a few tickets; AZURE_kerberoasting should enumerate more SPNs to cross the Sentinel detection threshold.
**Warning signs:** Identical `executor.command` content between AZURE_ and T-prefix file.

### Pitfall 3: LDAP EID 1644 Not Enabled by Default
**What goes wrong:** AZURE_ldap_recon generates EID 1644 events that never appear because the DC diagnostic logging threshold is not configured (requires: `HKLM\SYSTEM\CurrentControlSet\Services\NTDS\Diagnostics` → set value 15 Field Engineering to at least 5).
**Why it happens:** EID 1644 is NOT enabled by default on Windows DCs. The Preparation phase enables security/Sysmon events but not NTDS diagnostic logging.
**How to avoid:** AZURE_ldap_recon's `expected_events` for EID 1644 should be supplementary; primary expected_events should be EID 4688 (for LDAP tool processes) and EID 4104 (PowerShell ScriptBlock). Document the EID 1644 prerequisite clearly in the technique description. The technique command should still attempt the expensive LDAP query (which would generate EID 1644 if configured) while producing EID 4688/4104 unconditionally.
**Warning signs:** Verification fails for EID 1644 on a standard DC without NTDS diagnostic logging.

### Pitfall 4: CSS Block Placement
**What goes wrong:** Sentinel CSS classes placed outside the `{{if .HasSentinel}}` guard — causes CSS to render in reports with no Sentinel data, bloating the HTML and potentially confusing users.
**Why it happens:** Copying the CrowdStrike CSS block (lines 269–272) without wrapping in a new conditional.
**How to avoid:** The Phase 4 CrowdStrike CSS is already inside `{{if .HasCrowdStrike}}...{{end}}`. Add Sentinel CSS as a separate `{{if .HasSentinel}}...{{end}}` block immediately after. Do not add sentinel classes inside the CrowdStrike block.

### Pitfall 5: Column Order Breaks on Non-Sentinel Runs
**What goes wrong:** When `HasCrowdStrike=true` and `HasSentinel=false`, the column count mismatch between `<thead>` and `<tbody>` causes display issues if not both conditional.
**Why it happens:** Forgetting to guard the `<td>` row cell with `{{if $.HasSentinel}}` while the `<th>` is guarded.
**How to avoid:** Both the header `<th>` and each row `<td>` must be inside identically matched conditionals. The Phase 4 pattern in reporter.go (lines 314, 348–360) shows this correctly — copy both blocks.

---

## Code Examples

### HasSentinel Scan (reporter.go saveHTML)

```go
// Source: reporter.go lines 155-161 (hasCrowdStrike block) — mirror exactly
hasSentinel := false
for _, res := range r.Results {
    if len(res.SIEMCoverage["sentinel"]) > 0 {
        hasSentinel = true
        break
    }
}
// Add HasSentinel: hasSentinel to the data := htmlData{...} literal
```

### HTML Template Header (after CrowdStrike header)

```html
<!-- Source: reporter.go line 314 pattern -->
{{if .HasSentinel}}<th>Microsoft Sentinel</th>{{end}}
```

### HTML Template Row Cell (after CrowdStrike cell)

```html
<!-- Source: reporter.go lines 348-360 pattern — replace cs-* with ms-* and "crowdstrike" with "sentinel" -->
{{if $.HasSentinel}}
<td>
  {{$ms := siemCoverage .SIEMCoverage "sentinel"}}
  {{if $ms}}
    <span class="ms-badge">MS</span>
    <ul class="ms-list">
      {{range $ms}}<li>{{.}}</li>{{end}}
    </ul>
  {{else}}
    <span class="ms-na">N/A</span>
  {{end}}
</td>
{{end}}
```

### YAML siem_coverage with both platforms

```yaml
# Example: T1003_001_lsass.yaml after modification
siem_coverage:
  crowdstrike:
    - "Credential Dumping"
    - "Suspicious LSASS Access"
  sentinel:
    - "Dumping LSASS Process Into a File"
```

### AZURE_kerberoasting Command Approach

The existing T1558.003 requests a small number of tickets. AZURE_kerberoasting must cross the Sentinel rule's threshold: 15+ distinct SPN requests within 1 hour. Strategy:

```powershell
# Phase 1: Broad SPN enumeration (all service accounts) — generates EID 4104
Add-Type -AssemblyName System.IdentityModel
$searcher = New-Object System.DirectoryServices.DirectorySearcher
$searcher.Filter = "(&(objectClass=user)(servicePrincipalName=*)(!(cn=krbtgt)))"
$searcher.PropertiesToLoad.AddRange(@("servicePrincipalName","sAMAccountName"))
$results = $searcher.FindAll()

# Phase 2: Request TGS for each SPN — EID 4769 on DC per request
foreach ($r in $results) {
    foreach ($spn in $r.Properties["serviceprincipalname"]) {
        try {
            New-Object System.IdentityModel.Tokens.KerberosRequestorSecurityToken -ArgumentList $spn
            # EID 4769 fires on DC with EncryptionType=0x17 if account supports RC4
        } catch {}
    }
}
# Also add fallback machine SPNs: HOST/COMPUTERNAME etc.
```

This generates volume that crosses the Sentinel "15 distinct service names in 1 hour" threshold.

### AZURE_dcsync Command Approach

Must generate EID 4662 with DS-Replication GUIDs. The existing T1003.006 already does this. AZURE_dcsync should use a more focused ACL-access approach with explicit LDAP access to the domain object's replication extended rights, generating a clean EID 4662 burst:

```powershell
# Target: domain object ACL access (generates EID 4662 with replication GUIDs)
$domainDN = ([ADSI]"LDAP://RootDSE").defaultNamingContext
$domainObj = [ADSI]"LDAP://$domainDN"
# Access replication-related attributes directly (triggers EID 4662 on DC)
$replGUIDs = @(
    "1131f6aa-9c07-11d1-f79f-00c04fc2dcd2",  # DS-Replication-Get-Changes
    "1131f6ad-9c07-11d1-f79f-00c04fc2dcd2"   # DS-Replication-Get-Changes-All
)
# Check ACL for these extended rights (audit access triggers EID 4662)
$acl = $domainObj.ObjectSecurity
$replAces = $acl.Access | Where-Object { $_.ObjectType -in $replGUIDs }
```

### AZURE_ldap_recon Command Approach

Primary events: EID 4688 (ldifde/dsquery/adfind process creation) + EID 4104 (PowerShell ScriptBlock). Secondary: EID 1644 (fires only if NTDS diagnostic logging is configured on the DC).

```powershell
# Method 1: Broad DirectorySearcher (tries to trigger EID 1644 on DC)
$searcher = New-Object System.DirectoryServices.DirectorySearcher
$searcher.Filter = "(objectClass=*)"
$searcher.PageSize = 0         # no paging = larger result set
$searcher.SizeLimit = 1000     # forces "expensive" query
try { $searcher.FindAll() | Out-Null } catch {}

# Method 2: ldifde export (EID 4688)
ldifde -f "$env:TEMP\lognojutsu_ldap_export.ldf" -r "(objectClass=user)" 2>&1

# Cleanup
Remove-Item "$env:TEMP\lognojutsu_ldap_export.ldf" -ErrorAction SilentlyContinue
```

---

## Environment Availability Audit

Step 2.6: SKIPPED — Phase 5 is purely code/config/YAML changes. No external tools, services, or runtimes beyond PowerShell (already verified present from Phase 1 Preparation) are required. AZURE_ techniques run against local Windows AD without Azure cloud connectivity.

Note: AZURE_ techniques DO require a domain-joined Windows machine with Active Directory available to generate meaningful events. This is a SIEM validation tool prerequisite (pre-existing), not a new Phase 5 requirement.

---

## Validation Architecture

`nyquist_validation` is not explicitly set to false in `.planning/config.json` (only `_auto_chain_active` is set). Treating as enabled.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (`testing` stdlib) |
| Config file | none — `go test ./...` from project root |
| Quick run command | `go test ./internal/reporter/...` |
| Full suite command | `go test ./...` |

### Phase Requirements to Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SENT-01 | `siem_coverage.sentinel` key populates in YAML-loaded Technique struct | unit | `go test ./internal/playbooks/... -run TestSentinelCoverage` | ❌ Wave 0 |
| SENT-02 | 3 AZURE_ technique YAML files load successfully and have non-empty `SIEMCoverage["sentinel"]` | unit | `go test ./internal/playbooks/... -run TestAZURETechniqueCount` | ❌ Wave 0 |
| SENT-03 | `HasSentinel=true` when results contain sentinel coverage; HTML template renders ms-badge column | unit | `go test ./internal/reporter/... -run TestHasSentinel` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/playbooks/... ./internal/reporter/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/playbooks/playbooks_test.go` or similar — add `TestSentinelCoverage` and `TestAZURETechniqueCount` tests (check existing test file name — Phase 4 may have added technique count tests already)
- [ ] `internal/reporter/reporter_test.go` — add `TestHasSentinel` covering both `HasSentinel=true` path (results with sentinel data) and `HasSentinel=false` path (no sentinel data)

**Existing coverage check:** Phase 3 added `TestNewTechniqueCount`. Phase 4 may have added `HasCrowdStrike` reporter tests. Check these files first — the AZURE_ count test can extend the existing technique count test, and the `HasSentinel` test can extend the existing reporter test pattern.

---

## Open Questions

1. **LDAP EID 1644 — Sentinel rule name**
   - What we know: No dedicated analytic rule for SecurityEvent 1644 found in `Azure/Azure-Sentinel/Detections/SecurityEvent/`. MDI handles LDAP recon via `IdentityQueryEvents`.
   - What's unclear: Does a Windows Security Events solution pack contain a 1644-based rule? Is the CONTEXT.md "LDAP Query Reconnaissance" name a real rule or a shorthand?
   - Recommendation: AZURE_ldap_recon should not map `siem_coverage.sentinel` to a speculative rule name. Instead, map to `"Anomalous Amount of LDAP traffic"` (MEDIUM confidence, from Bert-JanP/Hunting-Queries repository — community rule, not built-in) or omit the sentinel mapping from AZURE_ldap_recon entirely and rely on EID 4688/4104 for SENT-02 event coverage. The 3-technique SENT-02 count does not require all 3 to have sentinel rule names — only SENT-01 requires mappings per technique where applicable.
   - **Decision needed by planner:** Use `"Anomalous Amount of LDAP traffic"` (community, MEDIUM) or omit sentinel mapping from AZURE_ldap_recon YAML.

2. **Existing technique siem_coverage.sentinel breadth**
   - What we know: 5 existing techniques identified as Sentinel mapping candidates (see table above). Only 3 are HIGH confidence.
   - What's unclear: Whether MEDIUM confidence names ("Suspicious Powershell Commandlet Executed", "User Created and Added to Built-in Administrators") are accurate enough to include without portal verification.
   - Recommendation: Include MEDIUM confidence names with an inline YAML comment: `# MEDIUM confidence — verify rule name in Sentinel portal`. This gives consultants a starting point while being honest about confidence level. SENT-01 says "documented per technique where applicable" — a commented mapping satisfies this.

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Sentinel rules named "Azure Sentinel" in docs | Now "Microsoft Sentinel" throughout | 2022 rebranding | Rule names in GitHub repo use "Microsoft Sentinel"; older blog posts use "Azure Sentinel" — trust GitHub YAML `name:` field |
| MMA agent (Log Analytics agent) for Windows event forwarding | AMA (Azure Monitor Agent) is current standard | 2022–2024 migration | For documentation: recommend AMA, note MMA is in deprecation path. Neither affects YAML content. |
| SecurityEvent table | SecurityEvent table (unchanged) | n/a | Security channel events (4662, 4769, 4688) still ingest via SecurityEvent table in both MMA and AMA paths |

**Deprecated/outdated:**
- "Azure Sentinel" name: replaced by "Microsoft Sentinel" — use "Microsoft Sentinel" in README documentation
- MMA agent: officially deprecated by Microsoft; recommend AMA in Sentinel prerequisites section

---

## Sources

### Primary (HIGH confidence)
- `github.com/Azure/Azure-Sentinel/blob/master/Detections/SecurityEvent/PotentialKerberoast.yaml` — displayName `"Potential Kerberoasting"` verified from raw YAML `name:` field
- `github.com/Azure/Azure-Sentinel/blob/master/Solutions/Windows Security Events/Analytic Rules/NonDCActiveDirectoryReplication.yaml` — displayName `"Non Domain Controller Active Directory Replication"` verified from raw YAML
- `github.com/Azure/Azure-Sentinel/blob/master/Detections/SecurityEvent/DumpingLSASSProcessIntoaFile.yaml` — displayName `"Dumping LSASS Process Into a File"` verified
- `D:/Code/LogNoJutsu/internal/reporter/reporter.go` — existing `hasCrowdStrike` pattern, `siemCoverage` funcMap, CrowdStrike CSS and template blocks (exact code to mirror for Sentinel)
- `D:/Code/LogNoJutsu/internal/playbooks/types.go` — `SIEMCoverage map[string][]string` already on both `Technique` and `ExecutionResult`; zero struct changes needed

### Secondary (MEDIUM confidence)
- `analyticsrules.exchange` — rule listings for PowerShell and account creation rules (names not fully verified against GitHub YAML)
- `managedsentinel.com/ms-a143` — corroborates "Potential Kerberoasting" as the Sentinel portal-visible name (MS-A143)
- Phase 4 RESEARCH.md and CONTEXT.md — established patterns for FALCON_ techniques that AZURE_ files mirror exactly

### Tertiary (LOW confidence)
- `github.com/Bert-JanP/Hunting-Queries-Detection-Rules` — "Anomalous Amount of LDAP traffic" as candidate LDAP recon detection (community hunting query, not built-in Sentinel analytics rule)

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies; existing types confirmed from codebase
- Architecture: HIGH — direct mirror of Phase 4 patterns confirmed from reporter.go source
- Sentinel rule names (Kerberoasting, DCSync, LSASS): HIGH — verified from Azure/Azure-Sentinel GitHub YAML files
- Sentinel rule names (PowerShell, account creation): MEDIUM — from community sources, not GitHub YAML
- LDAP Sentinel rule: LOW — no built-in SecurityEvent rule for EID 1644 found; MDI-based coverage only

**Research date:** 2026-03-25
**Valid until:** 2026-06-25 (Sentinel rule names are relatively stable; built-in rules in the GitHub repo change slowly)
