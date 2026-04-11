---
phase: 19-distributed-technique-scheduling
plan: "01"
subsystem: engine
tags: [go, poc, scheduling, tdd, engine, distributed]

requires:
  - phase: 19-00
    provides: Wave 0 test stubs (TestRandomSlotsInWindow, TestPoCPhase1_DistributedSlots, TestPoCPhase2_BatchedSlots)

provides:
  - PoCConfig with four window fields (Phase1WindowStart/End, Phase2WindowStart/End) replacing two daily-hour fields
  - randomSlotsInWindow helper generating sorted random slot durations within configurable window
  - runPoC() Phase 1 loop distributing one technique per randomly-timed slot
  - runPoC() Phase 2 loop batching 2-3 techniques per randomly-timed slot
  - Gap days using fixed 24h wait instead of nextOccurrenceOfHour
  - All three POC test stubs fleshed out and passing

affects:
  - 19-02-PLAN (Phase 2 is already implemented here; plan 19-02 was merged into 01)
  - Phase 20 (POC-04: update poc_test.go and engine_test.go for new distributed behavior)

tech-stack:
  added: ["sort (stdlib, added to engine.go imports)"]
  patterns:
    - "randomSlotsInWindow: sorted absolute-time offsets converted to inter-slot durations; derive per-day rand.Source from top-level rng.Int63()"
    - "Distributed Phase 1: one After() call per technique slot (not one per day)"
    - "Batched Phase 2: ceil(techniques/batchSize) After() calls per day (not one per day)"
    - "afterCountClock pattern: wrapper around fakeClock counting After() calls for distributed-slot assertions"

key-files:
  created: []
  modified:
    - internal/engine/engine.go
    - internal/engine/poc_schedule_test.go
    - internal/engine/engine_test.go
    - internal/engine/poc_test.go

key-decisions:
  - "Phase 1 uses no delayBetween() between slots — random jitter IS the inter-technique delay (D-09)"
  - "Batch size 2 or 3 chosen per day via rng.Intn(2)+2 (D-06); not configurable"
  - "rand.Source derived per-day from top-level rng.Int63() so randomSlotsInWindow is deterministically seeded without sharing state"
  - "engine_test.go and poc_test.go old field names fixed inline (Rule 3) — preserving existing test behavior with new field equivalents"

patterns-established:
  - "afterCountClock: embed fakeClock, override After() to increment counter; used to assert distributed (multiple After() calls) vs burst (single After()) scheduling"

requirements-completed: [POC-01, POC-02, POC-03]

duration: 15min
completed: 2026-04-10
---

# Phase 19 Plan 01: Distributed Technique Scheduling Summary

**runPoC() rewritten: Phase 1 distributes one technique per random slot, Phase 2 batches 2-3 techniques across window-bounded random slots using randomSlotsInWindow helper**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-04-10T20:38:40Z
- **Completed:** 2026-04-10T20:45:20Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added `randomSlotsInWindow()` helper that generates n sorted inter-slot durations randomly distributed within a configurable [windowStart:00, windowEnd:00) window
- Replaced `Phase1DailyHour`/`Phase2DailyHour` fields with four window fields in PoCConfig
- Phase 1 now fires each technique at its own randomly-timed slot (N After() calls per day instead of 1)
- Phase 2 now groups techniques in batches of 2-3 and fires each batch at a random slot (ceil(n/batchSize) After() calls instead of 1)
- Gap days use a fixed 24h wait instead of `nextOccurrenceOfHour`
- All three Wave 0 test stubs (TestRandomSlotsInWindow, TestPoCPhase1_DistributedSlots, TestPoCPhase2_BatchedSlots) replaced with real passing tests

## Task Commits

Each task was committed atomically:

1. **Task 1: Add window config fields and randomSlotsInWindow helper** - `263793c` (feat)
2. **Task 2: Rewrite runPoC() for distributed scheduling** - `7f0a352` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified

- `internal/engine/engine.go` - Four new PoCConfig window fields, `randomSlotsInWindow` helper, rewritten Phase 1/Gap/Phase 2 loops in `runPoC()`
- `internal/engine/poc_schedule_test.go` - TestRandomSlotsInWindow with window bounds + guard clause test; TestPoCPhase1_DistributedSlots + TestPoCPhase2_BatchedSlots using afterCountClock
- `internal/engine/engine_test.go` - Field names updated: Phase1DailyHour → Phase1WindowStart/End, Phase2DailyHour → Phase2WindowStart/End
- `internal/engine/poc_test.go` - Same field name updates as engine_test.go

## Decisions Made

- Phase 1 does not call `delayBetween()` between slots — random window jitter serves as the inter-technique delay (per D-09)
- Batch size 2-3 chosen randomly per day (not configurable in this release), per D-06
- Per-day rand.Source derived from top-level `rng.Int63()` to keep randomSlotsInWindow deterministically seeded without sharing mutable state
- `nextOccurrenceOfHour` retained (not removed) — still used by old tests and may be used by future phases

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated field names in engine_test.go and poc_test.go**
- **Found during:** Task 1 (PoCConfig struct field rename)
- **Issue:** engine_test.go and poc_test.go used `Phase1DailyHour`/`Phase2DailyHour` which no longer exist after the struct change — compilation blocked all tests
- **Fix:** Replaced `Phase1DailyHour: 8` with `Phase1WindowStart: 8, Phase1WindowEnd: 17` and `Phase2DailyHour: 9` with `Phase2WindowStart: 9, Phase2WindowEnd: 18` across all 14 occurrences in the two test files
- **Files modified:** `internal/engine/engine_test.go`, `internal/engine/poc_test.go`
- **Verification:** `go build ./internal/engine/...` succeeds, TestRandomSlotsInWindow passes
- **Committed in:** `263793c` (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (Rule 3 - blocking compile error)
**Impact on plan:** Necessary to enable test compilation. No scope creep. Existing test behavior preserved — old hour values mapped to equivalent window values.

## Issues Encountered

None.

## Known Stubs

None — all three POC requirement tests are implemented and passing.

Note: The existing poc_test.go tests (TestPoCDayCounter, TestPoCCurrentStepStrings, etc.) were written against the old single-wait-per-day behavior. The plan anticipates these will need updating in Phase 20 (POC-04). They continue to pass because the distributed slots with fakeClock still complete all techniques in the correct day sequence.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All three POC requirements (POC-01, POC-02, POC-03) implemented and test-covered
- randomSlotsInWindow is the single source of truth for slot generation; Phase 20 tests can exercise it directly
- Existing poc_test.go tests still pass — window-based scheduling is compatible with existing day-counter and step-string tests
- Phase 20 (POC-04) should update poc_test.go test assertions to verify window-based behavior explicitly

## Self-Check: PASSED

- `19-01-SUMMARY.md`: FOUND
- `263793c` (Task 1 commit): FOUND
- `7f0a352` (Task 2 commit): FOUND
- `go build ./internal/engine/...`: PASSED
- All three tests pass (not skipped): VERIFIED

---
*Phase: 19-distributed-technique-scheduling*
*Completed: 2026-04-10*
