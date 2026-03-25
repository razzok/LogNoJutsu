---
phase: 04-crowdstrike-siem-coverage
plan: 02
subsystem: playbooks
tags: [crowdstrike, falcon, yaml, techniques, mitre-attack, sysmon]

# Dependency graph
requires:
  - phase: 04-crowdstrike-siem-coverage/04-01
    provides: SIEMCoverage field in Technique struct, siem_coverage YAML block on existing techniques

provides:
  - 3 FALCON_-prefixed technique YAML files targeting Falcon behavioral detections
  - FALCON_process_injection: defense-evasion/T1055.001 targeting Falcon 'Code Injection'
  - FALCON_lsass_access: credential-access/T1003.001 with 0x0410 mask targeting Falcon 'Credential Dumping'/'Suspicious LSASS Access'
  - FALCON_lateral_movement_psexec: lateral-movement/T1021.002 targeting Falcon 'Lateral Movement (PsExec)'
  - TestFalconTechniques and updated TestNewTechniqueCount (threshold 51) in loader_test.go

affects: [04-03, microsoft-sentinel-coverage, README]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - FALCON_ prefix convention for Falcon-sensor-specific technique IDs
    - Single-line Add-Type signatures to avoid YAML @' here-string parsing issues
    - Multi-method approach: 2-3 PowerShell execution paths per technique for coverage depth

key-files:
  created:
    - internal/playbooks/embedded/techniques/FALCON_process_injection.yaml
    - internal/playbooks/embedded/techniques/FALCON_lsass_access.yaml
    - internal/playbooks/embedded/techniques/FALCON_lateral_movement_psexec.yaml
  modified:
    - internal/playbooks/loader_test.go

key-decisions:
  - "Here-strings (@'...'@) in PowerShell YAML commands must be flattened to single-line strings to avoid YAML parser errors (line-start '@ is YAML flow scalar)"
  - "FALCON_lsass_access uses GrantedAccess 0x0410 (PROCESS_QUERY_INFORMATION|PROCESS_VM_READ) vs T1003.001's 0x1010 — different access mask triggers different Falcon ML detection"
  - "FALCON_ techniques use standard MITRE tactic names (defense-evasion, credential-access, lateral-movement) not vendor-specific names"

patterns-established:
  - "FALCON_ prefix: vendor-targeted technique IDs follow FALCON_{technique_name} naming"
  - "TDD RED/GREEN: write failing tests first (Task 1), then create YAML files to pass (Task 2)"
  - "Single-line P/Invoke signatures: avoids @' here-string YAML conflict"

requirements-completed: [CROW-02]

# Metrics
duration: 4min
completed: 2026-03-25
---

# Phase 04 Plan 02: FALCON_ Technique YAML Files Summary

**3 CrowdStrike Falcon-targeted technique YAML files (process injection, LSASS credential dump, PsExec lateral movement) with authentic Sysmon events and official Falcon detection name mappings**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-03-25T18:44:24Z
- **Completed:** 2026-03-25T18:47:39Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Created 3 FALCON_-prefixed technique YAML files targeting CrowdStrike Falcon behavioral detections
- Each file has multi-method PowerShell execution patterns, expected_events (3-4 Sysmon/Security events each), and siem_coverage.crowdstrike with official Falcon detection names
- All 3 use standard MITRE tactic names (defense-evasion, credential-access, lateral-movement)
- FALCON_lsass_access uses 0x0410 access mask, distinct from T1003.001's 0x1010 — satisfies RESEARCH.md Pitfall 1
- TestFalconTechniques and TestNewTechniqueCount (threshold 51) added and passing; full suite green

## Task Commits

1. **Task 1: Add TestFalconTechniques and bump TestNewTechniqueCount** - `48827ec` (test) — TDD RED state
2. **Task 2: Create 3 FALCON_ technique YAML files** - `726c30e` (feat) — TDD GREEN state

## Files Created/Modified

- `internal/playbooks/embedded/techniques/FALCON_process_injection.yaml` - Process injection simulation targeting Falcon 'Code Injection'; defense-evasion / T1055.001; 4 expected events (Sysmon EID 8, 10, 1; Security 4688)
- `internal/playbooks/embedded/techniques/FALCON_lsass_access.yaml` - LSASS credential theft via MiniDumpWriteDump+ReadProcessMemory targeting Falcon 'Credential Dumping'/'Suspicious LSASS Access'; credential-access / T1003.001; uses 0x0410 mask; 4 expected events
- `internal/playbooks/embedded/techniques/FALCON_lateral_movement_psexec.yaml` - PsExec-style service creation+WMI+named pipe targeting Falcon 'Lateral Movement (PsExec)'; lateral-movement / T1021.002; 4 expected events (System 7045, Security 4688/4697, Sysmon EID 1)
- `internal/playbooks/loader_test.go` - Added TestFalconTechniques (3 FALCON_ IDs, crowdstrike coverage, valid MITRE tactics, phase=attack); bumped TestNewTechniqueCount threshold from 48 to 51; added FALCON_ IDs to required slice

## Decisions Made

- **Here-string flattening:** PowerShell `@'...'@` here-strings cause YAML parse errors because `'@` at line-start is a YAML flow scalar indicator. Solution: use single-line Add-Type signature strings (matching pattern of T1003_001_lsass.yaml Method 2). Fixed during Task 2 as Rule 1 (Bug).
- **0x0410 access mask for FALCON_lsass_access:** Follows RESEARCH.md Pitfall 1 requirement — FALCON_lsass_access must use a different access pattern than T1003.001's 0x1010.
- **Standard MITRE tactics:** All FALCON_ techniques use defense-evasion, credential-access, lateral-movement — not vendor-specific names.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed YAML parse error from PowerShell here-string delimiters**
- **Found during:** Task 2 (Create 3 FALCON_ technique YAML files)
- **Issue:** `@'...'@` PowerShell here-strings in YAML block scalar caused parse error: `yaml: line 57: could not find expected ':'` — the `'@` at line start is parsed as a YAML flow scalar indicator
- **Fix:** Converted multi-line here-strings to single-line quoted strings for Add-Type signatures in both FALCON_lsass_access.yaml and FALCON_process_injection.yaml
- **Files modified:** internal/playbooks/embedded/techniques/FALCON_lsass_access.yaml, FALCON_process_injection.yaml
- **Verification:** `go test ./internal/playbooks/...` passes
- **Committed in:** 726c30e (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug)
**Impact on plan:** Required fix — here-string pattern prevents YAML loading. Single-line strings achieve the same simulation behavior. No scope creep.

## Issues Encountered

YAML parse error from PowerShell here-string syntax — resolved by converting to single-line Add-Type signatures (same pattern used in existing T1003_001_lsass.yaml).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 3 FALCON_-prefixed technique files are present and loading successfully
- CROW-02 satisfied: at least 3 techniques generate Falcon sensor events with official detection name mappings
- Ready for Plan 03 (validation/count verification) or any remaining Phase 4 plans
- Total technique count: 54 (51+ threshold met)

---
*Phase: 04-crowdstrike-siem-coverage*
*Completed: 2026-03-25*
