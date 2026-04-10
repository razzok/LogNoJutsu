# Phase 18: Technique Realism Upgrades - Research

**Researched:** 2026-04-10
**Domain:** YAML technique metadata (tier classification + expected_events enrichment)
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** T1057 (Process Discovery) and T1482 (Domain Trust Discovery) are already Tier 1 native Go techniques from Phase 15. No further work needed — TECH-01 is satisfied for these two.
- **D-02:** T1069 (Group Discovery), T1082 (System Info), T1083 (File Discovery), T1135 (Network Share Discovery) are classified Tier 3 but already execute real Windows commands (net localgroup, systeminfo, Get-ChildItem, net view, etc.) generating authentic EID 4688/4104/5140 events. Upgrade strategy is **re-audit tiers** (bump from 3 to 1 or 2 based on event realism assessment), not rewrite to native Go.
- **D-03:** Tier bump includes **expected_events review** — ensure each technique's expected_events list accurately reflects which EIDs each command generates, for more precise SIEM validation.
- **D-04:** TECH-02 is **already satisfied** by existing techniques: T1053.005 (Scheduled Task, Tier 1), T1547.001 (Registry Run Key, Tier 1), T1197 (BITS Jobs, Tier 1), T1543.003 (New Service, Tier 1). All have cleanup commands and use the RunWithCleanup defer pattern from Phase 14.
- **D-05:** TECH-03 is **already satisfied** by existing techniques: T1027 (Encoded Commands, Tier 1), T1036.005 (Masquerading, Tier 1), T1218.011 (Rundll32 LOLBin, Tier 1), T1574.002 (DLL Sideloading, Tier 2). Tier 2 for T1574.002 is appropriate — uses a benign DLL, not a real sideload.
- **D-06:** TECH-04 is **already satisfied** by existing techniques: T1041 (HTTP Exfil, Tier 2), T1071.001 (HTTP C2, Tier 2), T1071.004 (DNS C2, Tier 2), T1560.001 (Archive/Encoding, Tier 2). Tier 2 is the correct classification — they use loopback/simulation patterns (real protocol, simulated target), which is the right safety tradeoff per Out of Scope constraints.

### Claude's Discretion

- Exact tier assignment for each of the 4 discovery techniques (1 vs 2) based on event realism assessment
- Expected_events enrichment details — which additional EIDs to add or descriptions to update
- Whether to update technique descriptions to better reflect their actual (non-stub) execution

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TECH-01 | Discovery stub techniques upgraded to real tool execution (T1057, T1069, T1082, T1083, T1135, T1482) | T1057 and T1482 already Tier 1 Go techniques (confirmed); T1069/T1082/T1083/T1135 already execute real commands — tier field and expected_events need correction only |
| TECH-02 | Persistence techniques added with mandatory cleanup (scheduled tasks, registry run keys, BITS jobs, service creation) | All 4 files confirmed at Tier 1 with non-empty cleanup fields: T1053.005, T1547.001, T1197, T1543.003 |
| TECH-03 | Defense evasion techniques added (encoded commands, masquerading, rundll32 LOLBin, DLL sideloading) | All 4 files confirmed at Tier 1/2: T1027 (1), T1036.005 (1), T1218.011 (1), T1574.002 (2) |
| TECH-04 | C2 and exfiltration techniques added using loopback/internal-only patterns | All 4 files confirmed at Tier 2: T1041, T1071.001, T1071.004, T1560.001 — loopback/simulation pattern confirmed |
</phase_requirements>

---

## Summary

Phase 18 is a YAML-only metadata correction pass. The four discovery techniques (T1069, T1082, T1083, T1135) have been misclassified as Tier 3 ("stub/echo") since their creation, but they already execute real Windows commands that generate authentic Windows Security and PowerShell event log entries. The tier field in each YAML and the TECHNIQUE-CLASSIFICATION.md table must be corrected to match what the techniques actually do.

TECH-02, TECH-03, and TECH-04 are fully satisfied by existing Tier 1/2 techniques. Confirmation required: read each technique file, verify tier and cleanup fields, and document the finding. No code, executor logic, or structural YAML changes are needed for these — only the REQUIREMENTS.md traceability checkboxes move from `[ ]` to `[x]`.

The expected_events arrays for the 4 discovery techniques also need review. All four currently list EIDs, but the descriptions and completeness vary. The planner must specify which EIDs to add or refine per technique.

**Primary recommendation:** Two sequential tasks — (1) re-audit and update the 4 YAML files + TECHNIQUE-CLASSIFICATION.md for TECH-01; (2) verify the already-satisfied technique files for TECH-02/03/04 and close out the requirements.

---

## Standard Stack

### Core (YAML-only phase)

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| YAML (technique files) | existing schema | Tier + expected_events metadata | Established format; Technique struct maps directly to YAML fields |
| Go test suite | go 1.26.1 | Validation that YAML parses correctly and tiers are valid | `TestTierClassified` enforces tier in {1,2,3}; `TestExpectedEvents` enforces non-empty expected_events |

No libraries to install. This phase adds no new dependencies.

---

## Architecture Patterns

### YAML Technique File Structure (authoritative — from types.go)

```yaml
id: T1069
name: Permission Groups Discovery
description: <human-readable description of what the technique does>
tactic: discovery
technique_id: T1069
platform: windows
phase: discovery           # "discovery" or "attack"
elevation_required: false
tier: 2                    # 1=real events, 2=some real+shortcuts, 3=echo/stub
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "net.exe 'localgroup Administrators' — standalone tier-1 SIEM rule"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Get-LocalGroup / Get-LocalGroupMember"
tags:
  - discovery
executor:
  type: powershell
  command: |-
    <commands here>
cleanup: ""
```

### TECHNIQUE-CLASSIFICATION.md Table Row Format

```markdown
| T1069 | Permission Groups Discovery | 2 | ps | No | No | <rationale sentence> |
```

The rationale column explains WHY the tier was assigned. Existing Tier 1 examples from the table: "Real powershell.exe invocation with full attacker flag set", "Real schtasks.exe /create with encoded PS payload". Tier 2 rationale pattern: "Real X but uses simulation Y".

### Tier Decision Logic (from Phase 14 definitions)

```
Tier 1: Generates REAL Windows events a SIEM would see in an actual attack.
         - Real attacker tools/APIs with realistic parameters
         - No simulation shortcuts, no fake targets, no fake data
         - Examples: net.exe generating EID 4688, wmic.exe generating EID 4688+1,
           systeminfo.exe generating EID 4688

Tier 2: Generates SOME real events but uses simulation shortcuts.
         - Real protocol/tool but simulated target (loopback, .invalid domain, fake creds)
         - Real event fires but scope is limited vs real attack

Tier 3: Echo/stub/query-only — proves technique runs but does NOT generate
         realistic SIEM events.
         - Pure enumeration that is already part of normal system use
         - No attacker-exclusive artifacts or behavioral signals
```

### Anti-Patterns to Avoid

- **Bumping all 4 to Tier 1 without justification:** T1135 accesses admin shares over SMB via `dir \\%COMPUTERNAME%\C$` — this generates EID 5140/5145 on loopback. Loopback SMB access is a simulation shortcut (Tier 2 pattern), not a real lateral movement. The tier decision needs per-technique analysis.
- **Leaving expected_events underdescribed:** The `description` field in EventSpec is what operators see in results. It should name the specific command and the event trigger so SIEM engineers know what to look for.
- **Modifying executor.command or cleanup:** Out of scope per D-02. Do not change any PowerShell command bodies.

---

## Per-Technique Tier Analysis

### T1069 — Permission Groups Discovery

**Commands executed:** `net localgroup`, `net localgroup Administrators`, `whoami /groups`, `Get-LocalGroup`, `Get-LocalGroupMember`, `wmic group get`, `WindowsIdentity.GetCurrent().Groups`, `net group /domain`

**Events generated:**
- EID 4688 (Security): `net.exe` process creation — command line contains "localgroup Administrators". This is a standalone SIEM rule in most products.
- EID 4688 (Security): `wmic.exe` process creation.
- EID 4688 (Security): `whoami.exe` process creation.
- EID 4104 (PowerShell/Operational): ScriptBlock for `Get-LocalGroup`, `Get-LocalGroupMember`, `.NET WindowsIdentity` calls.
- EID 1 (Sysmon): Same process creations captured via Sysmon if deployed.

**Tier verdict — Tier 1:** No simulation shortcuts. `net localgroup Administrators` invokes the real tool, generates real EID 4688 with the real command line visible in the CommandLine field. This is the exact same artifact that appears in real attacker activity. No loopback targets, no fake credentials, no benign substitutes. The `net group /domain` command will fail gracefully on non-domain hosts — this is expected, not a shortcut.

**Missing expected_events:** Sysmon EID 1 is not listed. While not strictly required (EID 4688 is the primary signal), adding it is consistent with other techniques that list both.

### T1082 — System Information Discovery

**Commands executed:** `systeminfo`, `hostname`, `whoami /all`, `reg query` (MachineGuid, BIOS), `wmic os get`, `wmic computersystem get`, `wmic bios get`, `net config workstation`, `Get-CimInstance Win32_OperatingSystem`, `Get-CimInstance Win32_ComputerSystem`

**Events generated:**
- EID 4688 (Security): `systeminfo.exe`, `hostname.exe`, `whoami.exe`, `reg.exe`, `wmic.exe`, `net.exe` — burst of process creations in rapid succession.
- EID 4104 (PowerShell/Operational): ScriptBlock for `Get-CimInstance` queries.
- EID 1 (Sysmon): Same process creations.

**Tier verdict — Tier 1:** Every command in the sequence uses real tools with real parameters. The YAML description correctly identifies the behavioral signal: "temporal clustering of multiple discovery tool invocations" — the burst pattern is a real Exabeam/UEBA behavioral alert trigger. No simulation shortcuts. `wmic.exe` with `/format:csv` is the attacker-style output format flag that SIEM rules specifically pattern-match on.

**Missing expected_events:** Sysmon EID 1 is already present. Current listing is accurate. Descriptions could be more specific about which exact processes fire EID 4688 (the burst members).

### T1083 — File and Directory Discovery

**Commands executed:** `cmd /c dir /s /b` on user profile paths, `tree /F`, `Get-ChildItem -Recurse` on sensitive paths, `Get-Item -Stream *` (ADS check), `cmd /c dir /s /b` with credential keyword filter

**Events generated:**
- EID 4688 (Security): `cmd.exe` process creation — dir arguments visible in CommandLine.
- EID 4688 (Security): `tree.com` process creation.
- EID 4104 (PowerShell/Operational): ScriptBlock for `Get-ChildItem` recursive calls, `Get-Item -Stream *`.
- EID 1 (Sysmon): Same process creations.

**Tier verdict — Tier 1:** Real `dir /s /b` and `tree /F` on real user directories. `cmd.exe` spawned from PowerShell with file enumeration arguments is a genuine attacker behavioral pattern. No simulation shortcuts — the commands operate on the real file system. The alternate data stream check via `Get-Item -Stream *` is an authentic attacker technique artifact.

**Possible Tier 2 argument:** The technique is non-destructive file enumeration with no lateral movement component. However, Tier 1 vs Tier 2 is about event authenticity, not destructiveness. The events generated (EID 4688 for cmd.exe with dir/s/b) are exactly what a SIEM would see in a real attack. Tier 1 is correct.

**Missing expected_events:** Sysmon EID 1 is already listed. Descriptions are adequate. The ADS check does not generate an additional unique EID beyond 4104 (it's part of the same ScriptBlock).

### T1135 — Network Share Discovery

**Commands executed:** `net share`, `Get-SmbShare`, `Get-WmiObject Win32_Share`, `net view`, `net view /all /domain`, `net view \\%COMPUTERNAME%`, `dir \\%COMPUTERNAME%\C$`, `dir \\%COMPUTERNAME%\ADMIN$`, `dir \\%COMPUTERNAME%\IPC$`

**Events generated:**
- EID 4688 (Security): `net.exe` with share/view arguments.
- EID 5140 (Security): Network share object accessed — fires when `dir \\COMPUTERNAME\C$` succeeds (admin share access audit). **Requires Object Access audit policy enabled.**
- EID 5145 (Security): Network share object checked for access rights. **Requires Object Access audit policy enabled.**
- EID 4104 (PowerShell/Operational): ScriptBlock for `Get-SmbShare`, `Get-WmiObject`.
- Sysmon EID 3: NetworkConnect to port 445 when accessing admin shares via UNC path.

**Tier verdict — Tier 2:** The admin share access via `dir \\%COMPUTERNAME%\C$` accesses the local machine's own admin shares (loopback SMB). This is structurally the same as the Tier 2 loopback pattern used in T1071.001 and UEBA-LATERAL-NEW-ASSET. The EID 5140/5145 events require Object Access auditing enabled — not guaranteed in default audit policy, unlike EID 4688. The local share enumeration (`net share`, `Get-SmbShare`) is genuine, but the admin share access is to loopback, not a remote host. Tier 2 is the correct assignment.

**Note:** EID 5140/5145 generation depends on audit policy. The expected_events list should note this condition.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead |
|---------|-------------|-------------|
| Tier validation in tests | Custom validation logic | `TestTierClassified` in loader_test.go already enforces tier in {1,2,3} — the test will auto-validate any updated tier value |
| expected_events presence validation | Custom check | `TestExpectedEvents` in loader_test.go enforces non-empty expected_events for every technique |
| YAML parsing | Custom parser | `LoadEmbedded()` in loader.go handles YAML unmarshalling — just edit the YAML files |

---

## Common Pitfalls

### Pitfall 1: Leaving TECHNIQUE-CLASSIFICATION.md out of sync

**What goes wrong:** The YAML tier field gets updated but the table in TECHNIQUE-CLASSIFICATION.md still shows "3". The doc is the human-readable audit trail — if it diverges from YAML, operators and SIEM engineers get conflicting information.

**How to avoid:** Update both atomically in the same commit. The table row includes ID, name, tier, exec type, elevation, has-cleanup, and rationale — all fields must be refreshed.

**Warning signs:** Summary statistics at the bottom of the doc will be wrong (Tier 3 count won't decrease by 4, Tier 1/2 won't increase).

### Pitfall 2: EID 5140/5145 audit policy dependency not documented

**What goes wrong:** T1135 lists EID 5140 and 5145 as expected events, but they only fire if "Object Access > File Share" audit policy is enabled. On default Windows installs this is off. A SIEM engineer will see "not_found" for these events on a standard system and assume the technique failed.

**How to avoid:** Add a note in the expected_events description field such as: "Requires Object Access > File Share audit policy enabled — not default". Alternatively, don't list 5140/5145 as primary expected events; list them as conditional.

### Pitfall 3: Confusing tier with danger level

**What goes wrong:** Thinking Tier 1 means "more dangerous" and hesitating to upgrade safe discovery techniques. Tier is about event authenticity, not risk. `net localgroup Administrators` is harmless but Tier 1 because it generates the exact EID 4688 pattern SIEM rules target.

**How to avoid:** Apply the Phase 14 D-02 definition precisely: "Tier 1 = generates real Windows events a SIEM would see in an actual attack." Dangerousness is tracked separately via `elevation_required` and `cleanup`.

### Pitfall 4: Expected_events descriptions too generic

**What goes wrong:** Description like "process creation" without naming the specific executable. SIEM verification uses the `description` field as human guidance — it's shown in the LogNoJutsu results UI. Generic descriptions make it hard to correlate which command triggered which event.

**How to avoid:** Name the executable and the key argument in the description. Pattern: `"<tool.exe> with <key-argument> — <what SIEM looks for>"`. See existing Tier 1 examples: `"schtasks.exe with /create and full arguments visible in CommandLine"`.

### Pitfall 5: Test `TestTierClassified` will catch invalid tier values

**What goes wrong:** Setting `tier: 0` or omitting the tier field after update.

**How to avoid:** Always set tier to 1, 2, or 3. Run `go test ./internal/playbooks/...` after editing YAML to confirm. The test runs in < 1 second (no I/O beyond embedded FS).

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go standard `testing` package, go 1.26.1 |
| Config file | None — `go test` discovers `*_test.go` automatically |
| Quick run command | `go test ./internal/playbooks/... -v -run TestTierClassified` |
| Full suite command | `go test ./internal/playbooks/... -v` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TECH-01 | All techniques have valid tier (1-3) after tier bump | unit | `go test ./internal/playbooks/... -run TestTierClassified` | Yes (loader_test.go) |
| TECH-01 | All techniques have non-empty expected_events | unit | `go test ./internal/playbooks/... -run TestExpectedEvents` | Yes (loader_test.go) |
| TECH-01 | Registry loads all YAML without parse errors | unit | `go test ./internal/playbooks/... -run TestNewTechniqueCount` | Yes (loader_test.go) |
| TECH-02 | T1053.005, T1547.001, T1197, T1543.003 have cleanup | unit | `go test ./internal/playbooks/... -run TestWriteArtifactsHaveCleanup` | Yes (loader_test.go) |
| TECH-03 | T1027, T1036.005, T1218.011, T1574.002 load and parse | unit | `go test ./internal/playbooks/... -run TestNewTechniqueCount` | Yes (loader_test.go) |
| TECH-04 | T1041, T1071.001, T1071.004, T1560.001 load and parse | unit | `go test ./internal/playbooks/... -run TestNewTechniqueCount` | Yes (loader_test.go) |

### Sampling Rate

- **Per task commit:** `go test ./internal/playbooks/... -run TestTierClassified -run TestExpectedEvents`
- **Per wave merge:** `go test ./internal/playbooks/...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

None — existing test infrastructure covers all phase requirements. No new test files needed.

---

## State of the Art

| Old Classification | Correct Classification | Impact |
|-------------------|----------------------|--------|
| T1069 tier: 3 | tier: 1 | Reflects that net.exe EID 4688 is a real, primary SIEM signal |
| T1082 tier: 3 | tier: 1 | Reflects that discovery burst (systeminfo/wmic/reg) is the real Exabeam behavioral trigger |
| T1083 tier: 3 | tier: 1 | Reflects that cmd.exe spawning dir /s /b on user profile paths is real attacker artifact |
| T1135 tier: 3 | tier: 2 | Loopback admin share access is a simulation shortcut; local enumeration is real |

---

## Open Questions

1. **Should T1135 EID 5140/5145 stay in expected_events?**
   - What we know: These events require Object Access audit policy enabled, which is off by default on most Windows systems.
   - What's unclear: Whether LogNoJutsu's target engagement environments have this policy enabled.
   - Recommendation: Keep 5140/5145 in expected_events but add a conditional note in the description. Do not remove them — they ARE generated when the policy is enabled, and SIEM rules depend on them. The planner should decide whether to add a `contains` qualifier or description note.

2. **Should Sysmon EID 1 be added to T1069 and T1082 expected_events?**
   - What we know: Both techniques trigger process creations detectable via Sysmon EID 1 as well as Security EID 4688. Other techniques like T1083 and T1135 already list Sysmon EID 1.
   - What's unclear: Project convention — some techniques list both channels, some only Security.
   - Recommendation: Add Sysmon EID 1 entries to T1069 and T1082 for consistency with T1083/T1135 and other Tier 1 techniques. The planner should confirm this is the pattern.

---

## Environment Availability

Step 2.6: SKIPPED — this phase is YAML-only file edits and Go test runs. No external tools, services, databases, or CLIs beyond the existing Go toolchain are required. The Go test suite uses embedded FS (no external dependencies).

---

## Sources

### Primary (HIGH confidence)

- Source files read directly from the repository:
  - `internal/playbooks/types.go` — authoritative Go struct definition; confirms all YAML field names and types
  - `internal/playbooks/loader_test.go` — confirms which tests validate tier, expected_events, cleanup
  - `internal/playbooks/embedded/techniques/T1069_group_discovery.yaml` — current state
  - `internal/playbooks/embedded/techniques/T1082_system_info_discovery.yaml` — current state
  - `internal/playbooks/embedded/techniques/T1083_file_discovery.yaml` — current state
  - `internal/playbooks/embedded/techniques/T1135_network_share_discovery.yaml` — current state
  - `docs/TECHNIQUE-CLASSIFICATION.md` — tier definitions, current table, summary statistics
  - `.planning/phases/18-technique-realism-upgrades/18-CONTEXT.md` — all locked decisions

### Secondary (MEDIUM confidence)

- Windows Event ID knowledge (EID 4688, 4104, 5140, 5145, Sysmon 1/3): Well-established Microsoft documentation mapping; consistent with how existing techniques in this project reference these EIDs. Not re-verified against Microsoft docs for this research pass — patterns are consistent with 50+ existing techniques in the codebase.

---

## Metadata

**Confidence breakdown:**
- Tier assessment for 4 techniques: HIGH — based on direct reading of YAML command bodies against the Phase 14 tier definitions
- TECH-02/03/04 satisfaction status: HIGH — confirmed by reading tier and cleanup fields from all 12 technique files
- Test coverage: HIGH — loader_test.go read in full; all relevant test functions identified
- EID accuracy: MEDIUM — consistent with in-codebase patterns; not re-verified against fresh Microsoft docs

**Research date:** 2026-04-10
**Valid until:** 2026-05-10 (stable domain — Windows Event IDs and YAML structure are not fast-moving)
