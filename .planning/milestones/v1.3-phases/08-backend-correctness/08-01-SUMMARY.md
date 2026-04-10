---
phase: 08-backend-correctness
plan: 01
subsystem: audit
tags: [go, auditpol, windows, audit-policy, guid, locale]

# Dependency graph
requires:
  - phase: 07-nyquist-validation
    provides: validated v1.0 codebase baseline
provides:
  - GUID-based audit policy configuration in preparation.go
  - preparation_test.go with tests for GUID usage, deduplication, and error format
affects: [08-02, any future changes to ConfigureAuditPolicy]

# Tech tracking
tech-stack:
  added: []
  patterns: [package-level var extracted from function for testability, TDD RED/GREEN cycle]

key-files:
  created:
    - internal/preparation/preparation_test.go
  modified:
    - internal/preparation/preparation.go

key-decisions:
  - "auditPolicies extracted to package-level var so tests can inspect entries directly (D-01, D-02)"
  - "11 entries after deduplication: Other Object Access Events + Scheduled Task share GUID 0CCE9227 (D-03)"
  - "Error format uses p.description not raw GUID: '<desc>: failed (exit status N)' (D-04)"
  - "Two disputed GUIDs (Audit Policy Change, Object access/Scheduled Task) marked with VERIFY comments (D-03)"

patterns-established:
  - "Package-level var pattern: extract local function vars to package scope when test access is needed"

requirements-completed: [BUG-01, BUG-02]

# Metrics
duration: 15min
completed: 2026-03-26
---

# Phase 8 Plan 01: Backend Correctness — Audit Policy GUID Migration Summary

**GUID-based auditpol calls replacing 12 locale-dependent English names with 11 stable Microsoft GUIDs, plus human-readable error messages using description field**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-03-26T00:00:00Z
- **Completed:** 2026-03-26
- **Tasks:** 2 (TDD: RED then GREEN)
- **Files modified:** 2

## Accomplishments

- Replaced 12 English subcategory names in ConfigureAuditPolicy with 11 locale-independent GUIDs (BUG-01)
- Deduplicated "Other Object Access Events" and "Scheduled Task" into a single GUID entry (0CCE9227) since they share the same GUID
- Error messages now show human-readable description first: "Logon/Logoff events (4624, 4625, 4634): failed (exit status 87)" instead of raw GUID (BUG-02)
- Two disputed GUIDs marked with VERIFY comments for on-machine validation
- TDD cycle: failing tests created first (RED), then implementation made them pass (GREEN)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add preparation_test.go with GUID and error format tests** - `59b0c7a` (test — RED phase)
2. **Task 2: Migrate ConfigureAuditPolicy to GUIDs and fix error format** - `8ac7081` (feat — GREEN phase)

**Plan metadata:** (final commit hash — see below)

_Note: TDD tasks have two commits (test RED → feat GREEN)_

## Files Created/Modified

- `internal/preparation/preparation_test.go` - Three tests: TestPoliciesUseGUIDs, TestPoliciesNoDuplicateGUIDs, TestAuditFailureMessageFormat
- `internal/preparation/preparation.go` - auditPolicies package-level var with 11 GUID entries; updated loop, error format, and success count

## Decisions Made

- Used package-level `auditPolicies` var (instead of keeping policies local to function) to make entries directly inspectable by tests — required for TDD
- Struct field renamed from `subcategory` to `guid` to clearly signal the value type
- Kept `description` field name unchanged — it's used in both error messages and success count
- `out` variable removed from error handling branch (no longer needed since error format uses only description and err)

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None — tests compiled and passed without any build issues.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- preparation.go BUG-01 and BUG-02 are resolved; auditPolicies is ready for future GUID updates
- Two disputed GUIDs (Audit Policy Change, Object access/Scheduled Task) require on-machine verification via `auditpol /list /subcategory:* /v` before shipping Phase 8
- Phase 8 Plan 02 (version injection + /api/info endpoint) can proceed independently

## Self-Check: PASSED

- FOUND: internal/preparation/preparation_test.go
- FOUND: internal/preparation/preparation.go
- FOUND: .planning/phases/08-backend-correctness/08-01-SUMMARY.md
- FOUND commits: 59b0c7a (test RED), 8ac7081 (feat GREEN)

---
*Phase: 08-backend-correctness*
*Completed: 2026-03-26*
