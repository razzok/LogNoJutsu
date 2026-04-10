---
phase: 18-technique-realism-upgrades
plan: "01"
subsystem: playbooks
tags:
  - technique-classification
  - yaml-metadata
  - tier-correction
  - expected-events
dependency_graph:
  requires:
    - internal/playbooks/embedded/techniques/*.yaml
    - docs/TECHNIQUE-CLASSIFICATION.md
  provides:
    - Corrected tier classifications for T1069, T1082, T1083, T1135
    - Enriched expected_events arrays with specific EID descriptions
    - Updated TECHNIQUE-CLASSIFICATION.md summary statistics
    - All TECH-01 through TECH-04 requirements formally closed
  affects:
    - SIEM validation accuracy (tier 1/2 techniques counted correctly)
    - Operator-visible expected events descriptions in results UI
tech_stack:
  added: []
  patterns:
    - YAML technique metadata schema (tier + expected_events fields)
key_files:
  created:
    - docs/TECHNIQUE-CLASSIFICATION.md (copied to worktree with updates)
    - .planning/REQUIREMENTS.md (copied to worktree with updates)
  modified:
    - internal/playbooks/embedded/techniques/T1069_group_discovery.yaml
    - internal/playbooks/embedded/techniques/T1082_system_info_discovery.yaml
    - internal/playbooks/embedded/techniques/T1083_file_discovery.yaml
    - internal/playbooks/embedded/techniques/T1135_network_share_discovery.yaml
decisions:
  - T1069/T1082/T1083 upgraded to Tier 1 — all execute real Windows tools with real parameters generating authentic EID 4688/4104 events
  - T1135 upgraded to Tier 2 — local share enumeration is real but admin share access uses loopback SMB (simulation shortcut); EID 5140/5145 require non-default audit policy
  - TECH-02/03/04 requirements satisfied by existing Phase 14-16 techniques — no new code needed, only formal sign-off
metrics:
  duration: "~5 min"
  completed: "2026-04-10T19:22:46Z"
  tasks_completed: 2
  tasks_total: 2
  files_modified: 6
---

# Phase 18 Plan 01: Technique Realism Re-Audit Summary

**One-liner:** Re-classified 4 misclassified discovery techniques (T1069/T1082/T1083 to Tier 1, T1135 to Tier 2) and formally closed TECH-01 through TECH-04 requirements via verified YAML inspection.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Update tier and expected_events in 4 discovery YAML files + TECHNIQUE-CLASSIFICATION.md | b0d1e77 | T1069, T1082, T1083, T1135 YAMLs + TECHNIQUE-CLASSIFICATION.md |
| 2 | Verify TECH-02/03/04 satisfaction and update REQUIREMENTS.md | 00fe550 | .planning/REQUIREMENTS.md |

## What Was Done

### Task 1: Tier Corrections and expected_events Enrichment

Four discovery techniques were misclassified as Tier 3 ("stub/echo") since their creation in earlier phases, despite executing real Windows commands that generate authentic SIEM events.

**T1069 Permission Groups Discovery** (Tier 3 → 1):
- `net localgroup Administrators` generates real EID 4688 with attacker command line visible — this is a standalone SIEM detection rule in most products
- Added Sysmon EID 1 entry for consistency with T1083/T1135
- Split EID 4688 into two entries: one for net.exe, one for whoami.exe/wmic.exe burst
- Updated descriptions to name specific executables per Pitfall 4 guidance

**T1082 System Information Discovery** (Tier 3 → 1):
- Full attacker recon burst (systeminfo/wmic/reg/hostname/whoami in rapid succession) is the Exabeam behavioral trigger — all real tools with real parameters
- Updated EID 4688 description to enumerate all burst members explicitly
- Updated Sysmon EID 1 description to name attacker-specific patterns (wmic with /format:csv)

**T1083 File and Directory Discovery** (Tier 3 → 1):
- Real `cmd.exe dir /s /b` and `tree /F` on real user directories — genuine attacker behavioral pattern
- Split EID 4688 into two entries: cmd.exe (dir pattern) and tree.com
- Updated EID 4104 description to include ADS check via `Get-Item -Stream *`

**T1135 Network Share Discovery** (Tier 3 → 2):
- Local share enumeration (`net share`, `Get-SmbShare`) is real, but admin share access via `dir \\COMPUTERNAME\C$` uses loopback SMB — simulation shortcut consistent with Tier 2 pattern
- Added EID 4104 entry for PowerShell share enumeration commands
- Updated EID 5140/5145 descriptions to explicitly note "requires Object Access > File Share audit policy enabled (not default)" — per Pitfall 2 guidance from research

**TECHNIQUE-CLASSIFICATION.md** updated:
- 4 table rows updated with new tier values and detailed rationale
- Summary statistics: Tier 1: 26+3=29, Tier 2: 18+1=19, Tier 3: 14-4=10 (Total: 58)

### Task 2: TECH-02/03/04 Verification and Requirements Sign-off

Verified each technique file from the main repo (which has tier fields from Phase 14):

| Requirement | Techniques | Confirmed Tier | Has Cleanup |
|-------------|-----------|----------------|-------------|
| TECH-02 | T1053.005, T1547.001, T1197, T1543.003 | All Tier 1 | Yes (all 4) |
| TECH-03 | T1027, T1036.005, T1218.011, T1574.002 | T1/T1/T1/T2 | Yes (all 4) |
| TECH-04 | T1041, T1071.001, T1071.004, T1560.001 | All Tier 2 | Yes (where applicable) |

All requirements formally marked `[x]` in REQUIREMENTS.md. Traceability table updated from Pending to Complete for TECH-01 through TECH-04.

## Verification Results

```
=== RUN   TestTierClassified
--- PASS: TestTierClassified (0.01s)
=== RUN   TestExpectedEvents
--- PASS: TestExpectedEvents (0.01s)
=== RUN   TestNewTechniqueCount
--- PASS: TestNewTechniqueCount (0.01s)
=== RUN   TestWriteArtifactsHaveCleanup
--- PASS: TestWriteArtifactsHaveCleanup (0.00s)
PASS ok lognojutsu/internal/playbooks
```

Note: Tests run against main repo (`D:/Code/LogNoJutsu`) for `TestWriteArtifactsHaveCleanup` and tier tests because the worktree branch predates Phase 14's tier classification work. The worktree's `TestExpectedEvents` and `TestNewTechniqueCount` also pass with the updated YAML files.

## Deviations from Plan

None — plan executed exactly as written. The worktree branching context required copying `docs/TECHNIQUE-CLASSIFICATION.md` and `.planning/REQUIREMENTS.md` from master (these files didn't exist in the older worktree branch), but this is expected parallel execution behavior, not a deviation from plan intent.

## Decisions Made

1. **T1069 → Tier 1**: `net localgroup Administrators` generates real EID 4688 with attacker command line — no simulation shortcuts. `net group /domain` fails gracefully on non-domain hosts — expected, not a shortcut.
2. **T1082 → Tier 1**: Temporal clustering of systeminfo/wmic/reg/hostname/whoami burst is the Exabeam behavioral trigger — every tool uses real parameters.
3. **T1083 → Tier 1**: `cmd.exe dir /s /b` on user profile paths is genuine attacker behavioral pattern — Tier 1 is about event authenticity, not destructiveness.
4. **T1135 → Tier 2**: Admin share access via loopback SMB is a simulation shortcut (same pattern as T1071.001 and UEBA-LATERAL-NEW-ASSET). EID 5140/5145 are conditional on non-default audit policy.
5. **TECH-02/03/04 sign-off**: No new code needed — all 12 supporting techniques are already at correct tiers with cleanup. Only REQUIREMENTS.md traceability updates required.

## Known Stubs

None — all 4 discovery techniques generate authentic Windows events. Tier 2 for T1135 reflects the loopback simulation shortcut, which is the intended safety tradeoff per project Out of Scope constraints.

## Self-Check: PASSED
