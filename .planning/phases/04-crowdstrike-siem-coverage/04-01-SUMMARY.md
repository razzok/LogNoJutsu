---
phase: 04-crowdstrike-siem-coverage
plan: 01
subsystem: playbooks
tags: [siem-coverage, crowdstrike, falcon, mitre-attack, yaml, go-structs]

# Dependency graph
requires:
  - phase: 03-additional-techniques
    provides: embedded YAML technique registry with EventSpec patterns and loader_test.go baseline

provides:
  - SIEMCoverage map[string][]string field on Technique and ExecutionResult structs
  - YAML siem_coverage key parsing (via yaml:"siem_coverage,omitempty")
  - Engine propagation: result.SIEMCoverage = t.SIEMCoverage in runTechnique (all code paths)
  - 10 existing technique YAMLs with siem_coverage.crowdstrike Falcon prevention policy names
  - TestSIEMCoverage unit test validating round-trip YAML parsing

affects:
  - 04-02: FALCON_ technique YAML files will use the same siem_coverage.crowdstrike key
  - 04-03: HTML report column needs SIEMCoverage populated on ExecutionResult to render

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "siem_coverage: map[string][]string — SIEM vendor keyed to list of policy/rule names"
    - "TDD RED/GREEN: test committed in failing state, then struct + YAML changes to pass"
    - "Single propagation point in runTechnique after if/else block — covers all four code paths"

key-files:
  created: []
  modified:
    - internal/playbooks/types.go
    - internal/engine/engine.go
    - internal/playbooks/loader_test.go
    - internal/playbooks/embedded/techniques/T1059_001_powershell.yaml
    - internal/playbooks/embedded/techniques/T1059_003_cmd_shell.yaml
    - internal/playbooks/embedded/techniques/T1003_001_lsass.yaml
    - internal/playbooks/embedded/techniques/T1547_001_registry_persistence.yaml
    - internal/playbooks/embedded/techniques/T1548_002_uac_bypass.yaml
    - internal/playbooks/embedded/techniques/T1027_obfuscated_commands.yaml
    - internal/playbooks/embedded/techniques/T1562_002_disable_logging.yaml
    - internal/playbooks/embedded/techniques/T1218_011_rundll32.yaml
    - internal/playbooks/embedded/techniques/T1543_003_new_service.yaml
    - internal/playbooks/embedded/techniques/T1134_001_token_impersonation.yaml

key-decisions:
  - "Single result.SIEMCoverage = t.SIEMCoverage line placed after the entire if/else block in runTechnique — covers WhatIf, runner, RunCleanup, and RunAs paths with one line"
  - "Discovery techniques (T1016, T1049, etc.) intentionally omitted — benign enumeration does not trigger Falcon prevention policies"
  - "siem_coverage inserted after tags: block and before executor: block in all YAMLs for consistent placement"

patterns-established:
  - "siem_coverage YAML placement: after tags block, before executor block"
  - "CrowdStrike key is 'crowdstrike' (lowercase) — matches map key in TestSIEMCoverage assertion"

requirements-completed: [CROW-01]

# Metrics
duration: 3min
completed: 2026-03-25
---

# Phase 04 Plan 01: SIEMCoverage Data Model Summary

**SIEMCoverage map[string][]string data layer added to Technique and ExecutionResult structs, propagated through the engine, and 10 existing technique YAMLs populated with official CrowdStrike Falcon prevention policy names**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-25T18:38:11Z
- **Completed:** 2026-03-25T18:41:48Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments

- Added `SIEMCoverage map[string][]string` to `Technique` struct with `yaml:"siem_coverage,omitempty"` tag — YAML round-trip parsing confirmed
- Added `SIEMCoverage map[string][]string` to `ExecutionResult` struct — data flows from technique definition to execution output
- Engine propagation via single `result.SIEMCoverage = t.SIEMCoverage` line covering all four runTechnique code paths (WhatIf, runner, RunCleanup, RunAs)
- 10 attack-phase YAML techniques populated with Falcon prevention policy names from Pulumi registry research
- TestSIEMCoverage validates YAML parsing round-trip; all 7 loader tests green

## Task Commits

Each task was committed atomically (TDD = multiple commits):

1. **Task 1 RED: TestSIEMCoverage (failing)** - `3cd32d1` (test)
2. **Task 1 GREEN struct + engine** - `e1254dc` (feat)
3. **Task 1 GREEN YAML + passing test** - `e6cacd3` (feat)
4. **Task 2: Populate 9 remaining YAML files** - `a96bf7b` (feat)

## Files Created/Modified

- `internal/playbooks/types.go` - SIEMCoverage field added to Technique and ExecutionResult
- `internal/engine/engine.go` - result.SIEMCoverage = t.SIEMCoverage propagation in runTechnique
- `internal/playbooks/loader_test.go` - TestSIEMCoverage function added
- `internal/playbooks/embedded/techniques/T1059_001_powershell.yaml` - Suspicious Scripts and Commands, Interpreter-Only
- `internal/playbooks/embedded/techniques/T1059_003_cmd_shell.yaml` - Suspicious Scripts and Commands
- `internal/playbooks/embedded/techniques/T1003_001_lsass.yaml` - Credential Dumping
- `internal/playbooks/embedded/techniques/T1547_001_registry_persistence.yaml` - Suspicious Registry Operations
- `internal/playbooks/embedded/techniques/T1548_002_uac_bypass.yaml` - Windows Logon Bypass (Sticky Keys)
- `internal/playbooks/embedded/techniques/T1027_obfuscated_commands.yaml` - Suspicious Scripts and Commands, Script-Based Execution Monitoring
- `internal/playbooks/embedded/techniques/T1562_002_disable_logging.yaml` - Suspicious Scripts and Commands
- `internal/playbooks/embedded/techniques/T1218_011_rundll32.yaml` - Javascript via Rundll32
- `internal/playbooks/embedded/techniques/T1543_003_new_service.yaml` - Suspicious Registry Operations
- `internal/playbooks/embedded/techniques/T1134_001_token_impersonation.yaml` - Code Injection

## Decisions Made

- Single `result.SIEMCoverage = t.SIEMCoverage` placed after the if/else block in runTechnique rather than duplicating in each branch — cleaner and covers all paths atomically
- Discovery/UEBA techniques deliberately excluded from siem_coverage population — benign enumeration activities do not trigger Falcon prevention policies (decision D-02 from research)
- siem_coverage block placed after tags and before executor in all YAMLs — consistent, scannable position

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Known Stubs

None. All siem_coverage data is real Falcon prevention policy names from Pulumi registry research (HIGH confidence). No placeholder values.

## Next Phase Readiness

- SIEMCoverage data layer is complete — Plan 02 (FALCON_ techniques) can add new YAML technique files using the same siem_coverage.crowdstrike key pattern
- Plan 03 (HTML column) can read ExecutionResult.SIEMCoverage directly from the results slice already stored on Status
- No blockers

---
*Phase: 04-crowdstrike-siem-coverage*
*Completed: 2026-03-25*
