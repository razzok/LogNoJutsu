---
phase: 03-additional-techniques
plan: 02
subsystem: playbooks
tags: [yaml, ueba, exabeam, mitre-attack, tdd, go-test]

# Dependency graph
requires:
  - phase: 01-events-manifest-verification-engine
    provides: EventSpec struct and expected_events YAML format established
  - phase: 02-code-structure-test-coverage
    provides: Go test infrastructure and package playbooks white-box test pattern
provides:
  - 4 new UEBA scenario YAML files (UEBA-DATA-STAGING, UEBA-ACCOUNT-TAKEOVER, UEBA-PRIV-ESC, UEBA-LATERAL-NEW-ASSET)
  - TestExpectedEvents: machine-verifiable TECH-03 coverage test (all techniques have expected_events)
  - TestNewUEBACount: TECH-02 validation test (7+ UEBA scenarios, 4 new IDs present)
  - TestNewTechniqueCount: TECH-01 validation test (requires plan 03-01 to fully pass)
affects: [03-01-additional-techniques, wave-end verification]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "UEBA scenario YAML pattern: tactic: ueba-scenario, UEBA DETECTION EXPECTED block, 3+ expected_events"
    - "White-box loader test pattern: package playbooks, LoadEmbedded() call in test"
    - "Wave-parallel test design: tests committed that partially fail until sibling plan completes"

key-files:
  created:
    - internal/playbooks/embedded/techniques/UEBA_data_staging_exfil_chain.yaml
    - internal/playbooks/embedded/techniques/UEBA_account_takeover_chain.yaml
    - internal/playbooks/embedded/techniques/UEBA_privilege_escalation_chain.yaml
    - internal/playbooks/embedded/techniques/UEBA_lateral_movement_new_asset.yaml
    - internal/playbooks/loader_test.go
  modified: []

key-decisions:
  - "TestNewTechniqueCount committed in RED state (failing) because plan 03-01's ATT&CK files are created in a parallel wave — test passes at wave-end when all 9 files are present"
  - "UEBA_lateral_movement_new_asset covers network-based first-time asset access (SMB+RDP probe) — distinct from UEBA_lateral_discovery_chain which covers enumeration speed (12 commands burst)"
  - "cleanup field uses bare string for UEBA_lateral_movement_new_asset (net use /delete) to match pattern from existing files"

patterns-established:
  - "UEBA YAML pattern: every UEBA scenario ends with Write-Host UEBA DETECTION EXPECTED block listing Exabeam use case names"
  - "expected_events channels restricted to Security, Microsoft-Windows-PowerShell/Operational, Microsoft-Windows-Sysmon/Operational only"

requirements-completed: [TECH-02, TECH-03]

# Metrics
duration: 10min
completed: 2026-03-25
---

# Phase 3 Plan 02: Additional UEBA Scenarios Summary

**4 Exabeam UEBA scenario YAMLs (data-staging, account-takeover, priv-esc, lateral-new-asset) plus LoadEmbedded loader tests validating TECH-02/TECH-03 coverage**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-03-25T14:15:22Z
- **Completed:** 2026-03-25T14:25:00Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Created 4 UEBA scenario YAML files covering distinct Exabeam use case families (data exfiltration, account compromise, privilege escalation, lateral movement/new asset)
- Each UEBA file follows established pattern: tactic: ueba-scenario, 3-4 expected_events entries, UEBA DETECTION EXPECTED output block
- Added loader_test.go with TestExpectedEvents (all techniques have expected_events — TECH-03), TestNewUEBACount (7+ UEBA scenarios — TECH-02), and TestNewTechniqueCount (48+ ATT&CK — TECH-01, passes at wave-end)

## Task Commits

Each task was committed atomically:

1. **Task 1: Create 4 UEBA scenario YAML files** — `a30c4d8` (feat)
2. **Task 2: Add loader tests (TDD RED)** — `0c3b576` (test)

_Note: TDD RED commit — TestExpectedEvents and TestNewUEBACount PASS; TestNewTechniqueCount FAIL pending plan 03-01 ATT&CK files (Wave 1 parallel)._

## Files Created/Modified

- `internal/playbooks/embedded/techniques/UEBA_data_staging_exfil_chain.yaml` — Data staging + HTTP exfil chain (T1074, EID 4104/11/3)
- `internal/playbooks/embedded/techniques/UEBA_account_takeover_chain.yaml` — Failed logins + post-auth enumeration (T1078, EID 4625/4624/4688)
- `internal/playbooks/embedded/techniques/UEBA_privilege_escalation_chain.yaml` — Privilege enumeration chain (T1134.001, EID 4688/4104/4672)
- `internal/playbooks/embedded/techniques/UEBA_lateral_movement_new_asset.yaml` — SMB/RDP first-time asset access (T1021.002, EID 5140/4688/3/4624)
- `internal/playbooks/loader_test.go` — 3 loader tests (TestExpectedEvents, TestNewTechniqueCount, TestNewUEBACount)

## Decisions Made

- TestNewTechniqueCount committed in partial-fail state because plan 03-01's ATT&CK technique files (T1005, T1560.001, T1119, T1071.001, T1071.004) are being created in a parallel Wave 1 plan; the test is correctly authored and will pass once all 9 files are present
- UEBA-LATERAL-NEW-ASSET is distinct from UEBA-LATERAL-CHAIN: the existing file covers enumeration speed (12-command burst); the new file covers network-based first-time asset access (SMB admin share + RDP port probe) — different Exabeam use case family

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- All 4 UEBA YAML files loadable by the existing Go embed/WalkDir infrastructure (no code changes needed)
- TestExpectedEvents and TestNewUEBACount pass with current file set
- TestNewTechniqueCount will pass once plan 03-01's 5 ATT&CK technique files are merged (wave-end verification)
- Wave 1 completion requires both 03-01 and 03-02 to be merged before running `go test ./...`

## Self-Check: PASSED

- FOUND: internal/playbooks/embedded/techniques/UEBA_data_staging_exfil_chain.yaml
- FOUND: internal/playbooks/embedded/techniques/UEBA_account_takeover_chain.yaml
- FOUND: internal/playbooks/embedded/techniques/UEBA_privilege_escalation_chain.yaml
- FOUND: internal/playbooks/embedded/techniques/UEBA_lateral_movement_new_asset.yaml
- FOUND: internal/playbooks/loader_test.go
- FOUND: .planning/phases/03-additional-techniques/03-02-SUMMARY.md
- FOUND: commit a30c4d8 (Task 1: 4 UEBA YAML files)
- FOUND: commit 0c3b576 (Task 2: loader_test.go TDD RED)
- FOUND: commit f691f20 (docs: SUMMARY + STATE + ROADMAP)
- TestExpectedEvents: PASS (all 47 current techniques have expected_events)
- TestNewUEBACount: PASS (7 UEBA scenarios including all 4 new ones)
- TestNewTechniqueCount: FAIL (expected — plan 03-01 ATT&CK files absent in this worktree, passes at wave-end)

---
*Phase: 03-additional-techniques*
*Completed: 2026-03-25*
