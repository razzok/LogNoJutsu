# Phase 20: Scheduling Test Coverage - Context

**Gathered:** 2026-04-11
**Status:** Ready for planning

<domain>
## Phase Boundary

Comprehensive test overhaul for the distributed scheduling engine shipped in Phase 19. Audit existing tests for implicit fire-all-at-once assumptions, fix any that pass by luck, and add DayDigest accuracy tests under distributed slots. All tests must use structural assertions (observable behavior) rather than exact timing.

</domain>

<decisions>
## Implementation Decisions

### Coverage Gap Assessment
- **D-01:** Phase 19 already delivered 3 tests in `poc_schedule_test.go` covering success criteria 2-3 (distributed Phase 1 slots, Phase 2 batching). Phase 20 focuses on DayDigest accuracy (criterion 4) and auditing existing tests for hidden assumptions.
- **D-02:** Comprehensive overhaul — review every existing test for implicit assumptions about the old fire-all-at-once behavior, not just add new tests.

### DayDigest Accuracy
- **D-03:** Count verification per slot: test that Phase 1 with multiple techs/day shows correct TechniqueCount and PassCount+FailCount after distributed execution. Test that Phase 2 campaign shows total technique count (not batch count).
- **D-04:** No need for full lifecycle-per-slot testing (heartbeat updates between individual slots). Count accuracy is sufficient.

### Deterministic Seeding
- **D-05:** Structural assertions only — assert on observable behavior (After() call count, technique execution count, DayDigest values) rather than exact slot timing. No rand.Source injection into runPoC(). This matches the Phase 19 test pattern and is resilient to algorithm changes.

### Existing Test Modernization
- **D-06:** Audit and fix implicit assumptions in existing tests. Review each test's `blockAt` values, After() count assumptions, and phase transition timing. Update any that pass by coincidence rather than correctness. Document which tests now exercise distributed code paths.
- **D-07:** Do NOT replace tests entirely — update in place. Lower risk, preserves test history.

### Claude's Discretion
- Edge case selection for new tests (window boundaries, single-technique days, etc.)
- Whether to add sub-tests or keep flat test functions
- Test naming conventions for new distributed-aware tests

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Test Files (primary targets)
- `internal/engine/poc_test.go` — 6 tests: day counter, DayDigest lifecycle, stop-signal handling (audit targets)
- `internal/engine/engine_test.go` — 17 tests: DayDigest Pre/Lifecycle/Counts/Heartbeat/GapDays/Reset, campaign delay, clock injection (audit targets)
- `internal/engine/poc_schedule_test.go` — 3 tests from Phase 19: randomSlotsInWindow, Phase 1 distributed slots, Phase 2 batched slots (reference pattern)

### Engine Implementation (context)
- `internal/engine/engine.go` — `randomSlotsInWindow()`, rewritten `runPoC()`, `PoCConfig` with window fields
- `internal/engine/engine.go` lines ~92-112 — `DayDigest` struct

### Test Infrastructure
- `internal/engine/engine_test.go` — `fakeClock`, `captureClock`, `newPoCEngineWithCampaign`, `pocConfig` helper, `waitForPhase`
- `internal/engine/poc_test.go` — `dayCaptureClock`, `digestCaptureClock`, `stopOnNthClock`, `newStopOnNthEngine`
- `internal/engine/poc_schedule_test.go` — `afterCountClock` pattern from Phase 19

### Prior Decisions
- Phase 10: Clock interface via `e.clock.Now()` / `e.clock.After()` — all timing injectable
- Phase 13: captureClock/digestCaptureClock snapshot patterns for race-free assertions; stopOnNthClock generalizer
- Phase 19: randomSlotsInWindow uses per-day rand.Source; Phase 1 has no delayBetween() between slots

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `afterCountClock` (poc_schedule_test.go): Counts After() calls — reuse for distributed slot assertions
- `digestCaptureClock` (poc_test.go): Snapshots DayDigest state on each After() — reuse for mid-run DayDigest verification
- `pocConfig()` helper (engine_test.go): Already updated with window fields — use for all new tests
- `newPoCEngineWithCampaign()` (engine_test.go): Pre-wired engine with fakeClock and campaign registry
- `fakeRunner(0)` / alternating runner: Simulates pass/fail without real technique execution

### Established Patterns
- Structural assertions: assert After() call counts, DayDigest field values, phase transitions — not exact timing
- `waitForPhase(eng, target, timeout)`: Poll-based phase completion check with timeout
- `stopOnNthClock`: Controls exactly which After() call blocks for stop-signal testing
- Tests use `time.Date(2026, ...)` anchors with `fakeClock` for deterministic clock behavior

### Integration Points
- `pocConfig()` helper is shared across poc_test.go and engine_test.go — changes affect both
- `newStopOnNthEngine()` hardcodes `blockAt` — may need updated values for distributed scheduling (more After() calls per day)
- All tests use `fakeClock` which fires After() channels immediately — distributed slots still complete instantly in tests

</code_context>

<specifics>
## Specific Ideas

- The `stopOnNthClock.blockAt` values were calibrated for fire-all-at-once (1 After() per day). Under distributed scheduling, Phase 1 with N techniques produces N After() calls per day. Existing blockAt values likely still work by accident because the tests use 1 tech/day, but this should be verified and documented.
- `TestDayDigest_Counts` tests with 1 tech/day (Phase 1) and 2-step campaign (Phase 2) — doesn't exercise multi-technique distributed behavior. A new test with 3+ techs/day would close the DayDigest accuracy gap.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 20-scheduling-test-coverage*
*Context gathered: 2026-04-11*
