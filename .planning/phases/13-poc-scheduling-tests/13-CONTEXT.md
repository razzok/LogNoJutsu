# Phase 13: PoC Scheduling Tests - Context

**Gathered:** 2026-04-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Write deterministic tests for runPoC() scheduling logic using the fake clock — day counter transitions, stop-signal handling, DayDigest lifecycle. No new features, no engine changes, pure test coverage for TEST-02, TEST-03, TEST-04.

</domain>

<decisions>
## Implementation Decisions

### Coverage Depth
- **D-01:** Requirements-only coverage — fill gaps for TEST-02, TEST-03, TEST-04. Do not add edge cases, race condition tests, or boundary condition tests beyond what the three requirements specify.
- **D-02:** Existing tests (TestPoCDayCounter, TestDayDigest_Lifecycle, etc.) already cover basics — new tests should target specific gaps, not duplicate existing coverage.

### Stop-Signal Scenarios (TEST-03)
- **D-03:** Test stop signal during day wait (nextOccurrenceOfHour sleep) — most common real-world scenario
- **D-04:** Test stop signal between phase transitions (Phase1→Gap, Gap→Phase2) — verify phase boundary doesn't swallow signal
- **D-05:** Test stop signal during gap days — gap days have no techniques, just sleep cycle
- **D-06:** Test immediate stop after start — stop before first day executes, quick cancellation edge case

### Test Organization
- **D-07:** Create new file `internal/engine/poc_test.go` for all Phase 13 tests. engine_test.go is 900+ lines; separate file keeps PoC scheduling tests grouped and discoverable.
- **D-08:** Reuse existing helpers from engine_test.go (fakeClock, captureClock, newPoCEngine, newPoCEngineWithCampaign) — they're accessible from the same package.

### Claude's Discretion
- Specific test function naming and subtesting structure
- Which existing clock wrapper patterns (captureClock, afterTrackClock, blockingClock) to reuse vs create new ones for stop-signal tests
- Whether to use table-driven tests or individual test functions

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Test Infrastructure
- `internal/engine/engine_test.go` — Existing fakeClock, captureClock, newPoCEngine helpers and all current PoC tests
- `internal/engine/engine.go` — Clock interface, runPoC(), waitOrStop(), setPhase() implementations

### Requirements
- `.planning/REQUIREMENTS.md` §Testability — TEST-02, TEST-03, TEST-04 requirement descriptions
- `.planning/ROADMAP.md` §Phase 13 — Phase goal and dependencies

### Prior Decisions
- `.planning/phases/10-poc-engine-fixes-clock-injection/10-CONTEXT.md` — D-01..D-10: Clock injection design, globalDay counter, captureClock pattern

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `fakeClock` (engine_test.go:15): Basic clock with immediate After() — sufficient for day counter and DayDigest tests
- `captureClock` (engine_test.go:312): Wraps fakeClock, records engine state each After() call — useful for transition verification
- `afterTrackClock` (engine_test.go:831): Records After() call durations — useful for verifying sleep periods
- `blockingClock` pattern (engine_test.go:866): Clock where After() blocks until manually unblocked — essential for stop-signal tests
- `newPoCEngine()` (engine_test.go:30): Creates engine with configurable phase1/gap/phase2 day counts
- `newPoCEngineWithCampaign()` (engine_test.go:487): Same but with campaign steps for Phase 2

### Established Patterns
- Tests run PoC in a goroutine, then inspect engine state after completion or stop
- `eng.Stop()` sends stop signal; tests verify engine exits cleanly
- DayDigest tests use `eng.GetDayDigests()` to inspect state after PoC runs

### Integration Points
- All tests use `internal/engine` package — same package as production code
- Tests call `eng.RunWhatIfPoC()` which delegates to `runPoC()`

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches using established test patterns.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 13-poc-scheduling-tests*
*Context gathered: 2026-04-09*
