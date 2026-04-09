---
phase: 16-safety-infrastructure
plan: 01
subsystem: engine, executor
tags: [amsi, elevation, windows, safety, verification]

# Dependency graph
requires:
  - phase: 13-poc-scheduling-tests
    provides: engine test patterns (RunnerFunc injection, fakeClock, testRegistry helpers)
  - phase: 01-events-manifest-verification-engine
    provides: VerificationStatus type and constants in playbooks/types.go
provides:
  - VerifAMSIBlocked and VerifElevationRequired verification status constants
  - isAMSIBlocked() AMSI detection function in executor (3 string patterns + exit code -196608)
  - AMSI early-return path in runInternal() for powershell/psh executor types
  - isAdmin field in Engine struct populated at Start() via platform-specific checkIsElevated()
  - Elevation skip in runTechnique() for techniques with ElevationRequired=true when not admin
  - SetAdmin() test helper for injection in unit tests
  - engine_windows.go: checkIsElevated() using windows.GetCurrentProcessToken().IsElevated()
  - engine_other.go: checkIsElevated() permissive stub for non-Windows builds
  - RequiresConfirmation bool field on Technique struct
affects: [16-safety-infrastructure, reporter, verifier]

# Tech tracking
tech-stack:
  added: [golang.org/x/sys v0.41.0 — Windows token API for elevation check]
  patterns:
    - "Platform-split files (engine_windows.go / engine_other.go) with build tags for platform-specific behavior"
    - "Early-return pattern in runInternal() for AMSI blocks — prevents verifier from running on blocked techniques"
    - "SetAdmin() injection alongside SetRunner() — consistent testability pattern"

key-files:
  created:
    - internal/engine/engine_windows.go
    - internal/engine/engine_other.go
    - internal/executor/executor_amsi_test.go
    - internal/engine/engine_elevation_test.go
  modified:
    - internal/playbooks/types.go
    - internal/playbooks/types_test.go
    - internal/executor/executor.go
    - internal/engine/engine.go
    - go.mod
    - go.sum

key-decisions:
  - "AMSI detection only fires for powershell/psh executor types — CMD and Go executors are never blocked by AMSI"
  - "AMSI early-return before result.Success assignment prevents verifier from running on blocked techniques"
  - "checkIsElevated() split into engine_windows.go (real API) and engine_other.go (permissive stub) — avoids cross-compilation issues"
  - "isAdmin set once at Start() not per-technique — admin status doesn't change mid-run"
  - "SetAdmin() test helper mirrors SetRunner() pattern — consistent injection interface"

patterns-established:
  - "Platform build-tag split: engine_windows.go + engine_other.go pattern for future Windows-specific features"
  - "Early-return safety gates in runInternal() before Success assignment"

requirements-completed: [INFRA-01, INFRA-02]

# Metrics
duration: 20min
completed: 2026-04-09
---

# Phase 16 Plan 01: Safety Infrastructure Summary

**AMSI block detection (3 string patterns + exit code) in PowerShell executor path with elevation gating in engine — consultants now see VerifAMSIBlocked / VerifElevationRequired instead of misleading Fail**

## Performance

- **Duration:** ~20 min
- **Started:** 2026-04-09T16:44:00Z
- **Completed:** 2026-04-09T17:04:07Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- Added `VerifAMSIBlocked` and `VerifElevationRequired` constants to the playbooks VerificationStatus type
- Implemented `isAMSIBlocked()` in executor detecting 3 AMSI stderr patterns and exit code -196608
- AMSI check fires only for `powershell`/`psh` executor types — returns early before the verifier is invoked
- Added `isAdmin` field to Engine, set once at `Start()` via `checkIsElevated()` platform function
- Elevation skip in `runTechnique()` — techniques with `ElevationRequired=true` get `VerifElevationRequired` status when not admin
- Platform split: `engine_windows.go` (real Windows token API) + `engine_other.go` (permissive stub for non-Windows builds)
- 6 unit tests across 2 new test files — all pass

## Task Commits

Each task was committed atomically:

1. **Task 1: Add verification status constants and RequiresConfirmation field** - `e8a8426` (feat)
2. **Task 2: AMSI detection in executor, elevation gating in engine** - `525c1fd` (feat)

## Files Created/Modified
- `internal/playbooks/types.go` - Added VerifAMSIBlocked, VerifElevationRequired constants and RequiresConfirmation field
- `internal/playbooks/types_test.go` - Added assertions for new constants
- `internal/executor/executor.go` - Added isAMSIBlocked() function and AMSI check in runInternal()
- `internal/executor/executor_amsi_test.go` - TestIsAMSIBlocked_Patterns, NormalError, ExitCode
- `internal/engine/engine.go` - Added isAdmin field, SetAdmin(), checkIsElevated() call, elevation skip in runTechnique()
- `internal/engine/engine_windows.go` - checkIsElevated() via windows.GetCurrentProcessToken().IsElevated()
- `internal/engine/engine_other.go` - checkIsElevated() permissive stub for non-Windows builds
- `internal/engine/engine_elevation_test.go` - TestElevationSkip, TestElevationRun, TestElevationNotRequired
- `go.mod` - Added golang.org/x/sys v0.41.0
- `go.sum` - Updated checksums

## Decisions Made
- AMSI patterns checked in order: string patterns first, then exit code — string patterns are the primary signal
- `isAMSIBlocked` is unexported (package-internal) — only used by `runInternal`, no need for external access
- checkIsElevated() uses build tags (not runtime OS check) — cleaner compilation, no dead code on each platform
- `SetAdmin(false)` test default matches real unelevated consultant usage — tests verify the common safety case

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Worktree has separate file copies from the main repo — initial edits went to the wrong path. Corrected by identifying worktree root and editing files in the correct location.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- AMSI and elevation status constants are ready for reporter to display in HTML output (16-02)
- `RequiresConfirmation` field on Technique is available for future confirmation workflow (16-03)
- All existing tests continue to pass; no regressions introduced

---
*Phase: 16-safety-infrastructure*
*Completed: 2026-04-09*
