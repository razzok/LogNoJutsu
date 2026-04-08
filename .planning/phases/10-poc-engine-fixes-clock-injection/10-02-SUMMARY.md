---
phase: 10-poc-engine-fixes-clock-injection
plan: 02
subsystem: testing
tags: [go, clock-injection, poc-mode, test-coverage, engine, fake-clock]

# Dependency graph
requires:
  - phase: 10-poc-engine-fixes-clock-injection
    plan: 01
    provides: Clock interface, globalDay counter, English CurrentStep strings, simlog.Phase separators

provides:
  - fakeClock struct implementing Clock for deterministic PoC testing
  - captureClock wrapper for reliable CurrentStep capture during fast fake-clock runs
  - newPoCEngine helper for minimal PoC engine with fake clock wired in
  - TestPoCDayCounter — globalDay monotonic counter validation (POCFIX-01)
  - TestPoCCurrentStepStrings — English-only CurrentStep verification (POCFIX-02)
  - TestPoCPhaseLogSeparators — simlog.Phase entry verification (POCFIX-03)
  - TestPoCClockInjection — fake clock eliminates real sleeps (TEST-01)
affects: [13-poc-engine-tests]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "captureClock pattern: embed fakeClock, add engine reference, capture CurrentStep on each After() call — reliable state capture when goroutine scheduler may not observe fast-transitioning states"

key-files:
  created: []
  modified:
    - internal/engine/engine_test.go

key-decisions:
  - "captureClock wraps fakeClock with engine reference to capture CurrentStep on After() calls — more reliable than polling goroutine when fake clock fires too fast for scheduler to observe intermediate states"
  - "TestPoCPhaseLogSeparators uses strings.Contains with uppercase expected strings — simlog.Phase() uppercases all messages via strings.ToUpper"
  - "Four tests kept independent (each calls newPoCEngine) — no shared state between test functions"

patterns-established:
  - "captureClock pattern: when testing fast state transitions with a fake clock, embed the clock in a capturing wrapper that reads engine status synchronously on each clock call"

requirements-completed: [POCFIX-01, POCFIX-02, POCFIX-03, TEST-01]

# Metrics
duration: 10min
completed: 2026-04-08
---

# Phase 10 Plan 02: PoC Engine Test Coverage Summary

**Four deterministic PoC engine tests using fakeClock and captureClock — validating day counter monotonicity, English-only CurrentStep strings, simlog.Phase separators, and fake clock eliminating real sleeps**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-04-08T19:55:50Z
- **Completed:** 2026-04-08T20:05:38Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments

- Added `fakeClock` struct implementing `Clock` interface with immediate `After()` firing and monotonic time advance
- Added `newPoCEngine` helper wiring a 3-technique registry, campaign, fake clock, and fake runner for minimal PoC test setups
- Added `captureClock` wrapper (deviation from plan's polling approach) that captures `CurrentStep` synchronously on `After()` calls — guarantees reliable capture even when the fake clock makes goroutines run too fast for a polling goroutine to observe
- Added `TestPoCDayCounter` — verifies `PoCDay` reaches 5 for a 2+1+2 day config, `PoCTotalDays` also equals 5
- Added `TestPoCCurrentStepStrings` — verifies no German words in captured PoC day step strings, and confirms English "Day N of M" pattern appears
- Added `TestPoCPhaseLogSeparators` — verifies three `simlog.TypePhase` entries exist for "POC PHASE 1: DISCOVERY", "POC GAP", "POC PHASE 2: ATTACK"
- Added `TestPoCClockInjection` — verifies entire PoC run completes in under 2 seconds and fake clock time advances

## Task Commits

Each task was committed atomically:

1. **Task 1: Add fakeClock and newPoCEngine helper** - `800dcfa` (test)
2. **Task 2: Write PoC engine validation tests** - `2a66457` (test)

**Plan metadata:** (docs commit follows)

## Files Created/Modified

- `internal/engine/engine_test.go` — fakeClock, captureClock, newPoCEngine, four TestPoC* functions

## Decisions Made

- `captureClock` pattern chosen over polling goroutine for `TestPoCCurrentStepStrings`: with a fake clock, the PoC goroutine transitions through states so quickly that a 1ms-sleeping polling goroutine misses the intermediate CurrentStep values; `captureClock.After()` fires synchronously and reads engine status at exactly the right moment
- Expected phase strings in `TestPoCPhaseLogSeparators` use uppercase ("POC PHASE 1: DISCOVERY") because `simlog.Phase()` uppercases messages; case-insensitive comparison was not needed since the actual format is predictable

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Replaced polling goroutine with captureClock in TestPoCCurrentStepStrings**
- **Found during:** Task 2 (TestPoCCurrentStepStrings)
- **Issue:** Plan used a polling goroutine with 1ms sleep to capture CurrentStep values, but with the fake clock, the runPoC goroutine completes all state transitions before the polling goroutine can observe them; test consistently failed with "no CurrentStep contains English 'Day N of M' pattern; steps: [T1078 — T1078 test]" when run alongside other tests
- **Fix:** Introduced `captureClock` struct that embeds `fakeClock` and holds an `*Engine` reference; `After()` method captures the current `CurrentStep` synchronously before advancing time — guarantees the PoC day strings are captured at the exact moment they're set
- **Files modified:** internal/engine/engine_test.go
- **Verification:** `go test ./internal/engine/... -run "TestPoC" -timeout 30s` exits 0; all four tests PASS consistently

---

**Total deviations:** 1 auto-fixed (Rule 1 — bug in test timing approach)
**Impact on plan:** No scope change; test intent is identical; captureClock is strictly more correct for deterministic clock scenarios.

## Issues Encountered

- Test ordering sensitivity: when `TestPoCCurrentStepStrings` ran after `TestPoCDayCounter`, the Go scheduler did not yield to the polling goroutine before the engine completed all PoC transitions. Resolved by switching to the `captureClock` approach.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All four PoC engine requirements (POCFIX-01, POCFIX-02, POCFIX-03, TEST-01) now have passing test coverage
- `fakeClock` and `newPoCEngine` helpers are available for Phase 13 scheduling tests
- Full test suite passes: `go test ./... -timeout 60s` — all 6 packages pass
- `go vet ./...` clean

---
*Phase: 10-poc-engine-fixes-clock-injection*
*Completed: 2026-04-08*
