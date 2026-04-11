# Phase 20: Scheduling Test Coverage - Research

**Researched:** 2026-04-11
**Domain:** Go test engineering — distributed scheduling assertions, DayDigest accuracy, test modernization
**Confidence:** HIGH

## Summary

Phase 20 is a pure test-engineering phase. No production code changes. The Phase 19 engine (`runPoC()`) ships 3 tests in `poc_schedule_test.go` that already satisfy success criteria 2 and 3 (distributed Phase 1 slots, Phase 2 batching). This phase's job is: (1) add a DayDigest accuracy test under distributed scheduling (criterion 4, gap D-03), and (2) audit and fix all existing tests for hidden assumptions about the old fire-all-at-once behavior (criteria 1, D-06).

All 35 existing engine tests pass today (`go test ./internal/engine/... 0.928s`). The test infrastructure is mature: `fakeClock`, `afterCountClock`, `digestCaptureClock`, `stopOnNthClock`, `pocConfig()`, `newPoCEngineWithCampaign()`, `waitForPhase()` are all available and proven. Every new or updated test must use structural assertions (After() call counts, DayDigest field values, phase transitions) — not exact timing.

**Primary recommendation:** Add one new test (`TestDayDigest_DistributedCounts`) with 3+ Phase 1 techs/day to close the DayDigest gap. Audit the six `stopOnNthClock`-based tests to confirm `blockAt` values are still correct under distributed scheduling (they use 1 tech/day so they should be, but the assumption must be verified and documented in comments).

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Phase 19 already delivered 3 tests in `poc_schedule_test.go` covering success criteria 2-3 (distributed Phase 1 slots, Phase 2 batching). Phase 20 focuses on DayDigest accuracy (criterion 4) and auditing existing tests for hidden assumptions.
- **D-02:** Comprehensive overhaul — review every existing test for implicit assumptions about the old fire-all-at-once behavior, not just add new tests.
- **D-03:** Count verification per slot: test that Phase 1 with multiple techs/day shows correct TechniqueCount and PassCount+FailCount after distributed execution. Test that Phase 2 campaign shows total technique count (not batch count).
- **D-04:** No need for full lifecycle-per-slot testing (heartbeat updates between individual slots). Count accuracy is sufficient.
- **D-05:** Structural assertions only — assert on observable behavior (After() call count, technique execution count, DayDigest values) rather than exact slot timing. No rand.Source injection into runPoC(). This matches the Phase 19 test pattern and is resilient to algorithm changes.
- **D-06:** Audit and fix implicit assumptions in existing tests. Review each test's `blockAt` values, After() count assumptions, and phase transition timing. Update any that pass by coincidence rather than correctness. Document which tests now exercise distributed code paths.
- **D-07:** Do NOT replace tests entirely — update in place. Lower risk, preserves test history.

### Claude's Discretion
- Edge case selection for new tests (window boundaries, single-technique days, etc.)
- Whether to add sub-tests or keep flat test functions
- Test naming conventions for new distributed-aware tests

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| POC-04 | Existing PoC scheduling tests updated to validate distributed execution and DayDigest accuracy | All findings below directly enable implementation |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go `testing` | stdlib | Test runner, assertions, subtests | Project already uses exclusively |
| Go `sync` | stdlib | Mutex-guarded state in clock wrappers | Already used in all clock types |
| Go `math/rand` | stdlib | Fixed-seed sources for `randomSlotsInWindow` unit tests | Already used in `TestRandomSlotsInWindow` |
| Go `time` | stdlib | `time.Date()` anchors, `time.Duration` | Already used in all tests |

No new dependencies. This is a test-only phase against an existing package.

**Test run command (verified working):**
```bash
go test ./internal/engine/... -count=1 -timeout 30s
```

**Race detector run:**
```bash
go test ./internal/engine/... -count=1 -race -timeout 60s
```

## Architecture Patterns

### Existing Test Infrastructure (canonical — reuse, do not recreate)

```
internal/engine/
├── engine_test.go       — fakeClock, captureClock, afterTrackClock, blockingClock,
│                          newPoCEngine, newPoCEngineWithCampaign, pocConfig,
│                          waitForPhase, fakeRunner, testRegistry, testTechnique
├── poc_test.go          — dayCaptureClock, digestCaptureClock, stopOnNthClock,
│                          newStopOnNthEngine
└── poc_schedule_test.go — afterCountClock (Phase 19)
                           TestRandomSlotsInWindow
                           TestPoCPhase1_DistributedSlots
                           TestPoCPhase2_BatchedSlots
```

### Pattern 1: afterCountClock — Count After() calls to assert distributed slots

Already defined in `poc_schedule_test.go`. Reuse directly for any new distributed assertion.

```go
// Source: internal/engine/poc_schedule_test.go
type afterCountClock struct {
    mu    sync.Mutex
    inner *fakeClock
    count int
}
func (c *afterCountClock) Now() time.Time { return c.inner.Now() }
func (c *afterCountClock) After(d time.Duration) <-chan time.Time {
    c.mu.Lock(); c.count++; c.mu.Unlock()
    return c.inner.After(d)
}
func (c *afterCountClock) getCount() int {
    c.mu.Lock(); defer c.mu.Unlock(); return c.count
}
```

**When to use:** Assert minimum After() call count. Phase 1 with N techs/day = N After() per day. Phase 2 with M steps in batches of 2-3 = ceil(M/batchSize) After() calls.

### Pattern 2: digestCaptureClock — Snapshot DayDigest mid-run

Already defined in `poc_test.go`. Reuse for mid-execution DayDigest state assertions.

```go
// Source: internal/engine/poc_test.go
type digestCaptureClock struct {
    fakeClock
    eng       *Engine
    snapshots [][]DayDigest
    mu        sync.Mutex
}
func (c *digestCaptureClock) After(d time.Duration) <-chan time.Time {
    digests := c.eng.GetDayDigests()
    if len(digests) > 0 {
        snap := make([]DayDigest, len(digests))
        copy(snap, digests)
        c.mu.Lock()
        c.snapshots = append(c.snapshots, snap)
        c.mu.Unlock()
    }
    return c.fakeClock.After(d)
}
```

**When to use:** Verify DayDigest fields evolve correctly during a run (not just at completion).

### Pattern 3: pocConfig() helper — consistent Config construction

```go
// Source: internal/engine/engine_test.go
func pocConfig(phase1Days, gapDays, phase2Days int, campaignID string) Config {
    return Config{
        PoCMode:           true,
        Phase1DurationDays: phase1Days,
        Phase1TechsPerDay:  1,         // ← defaults to 1; override inline for multi-tech tests
        Phase1WindowStart: 8, Phase1WindowEnd: 17,
        GapDays:           gapDays,
        Phase2DurationDays: phase2Days,
        Phase2WindowStart: 9, Phase2WindowEnd: 18,
        CampaignID:        campaignID,
    }
}
```

**When to use:** Base for all new PoC tests. For multi-tech Phase 1 tests, construct `Config` inline (not via `pocConfig()`) because `Phase1TechsPerDay` is hardcoded to 1 in the helper.

### Pattern 4: stopOnNthClock — Precise After() blocking

```go
// Source: internal/engine/poc_test.go
type stopOnNthClock struct {
    fakeClock
    blockAt   int
    blockCh   chan time.Time
    callCount int
    mu        sync.Mutex
}
func (c *stopOnNthClock) After(d time.Duration) <-chan time.Time {
    c.mu.Lock(); c.callCount++; n := c.callCount; c.mu.Unlock()
    if n >= c.blockAt { return c.blockCh }
    return c.fakeClock.After(d)
}
```

**Critical for Phase 20:** The `blockAt` values in `newStopOnNthEngine` and the four `TestPoCStop_*` tests were calibrated for the old fire-all-at-once engine. Under distributed scheduling:
- Phase 1 with **1 tech/day** still produces **1 After() per day** (because N=1 slot = 1 After() call). All four stop tests use 1 tech/day via `pocConfig()`, so their `blockAt` values are structurally correct.
- However, this coincidence must be **verified explicitly** and **documented in comments** so future developers understand the dependency.

### Anti-Patterns to Avoid

- **Asserting exact After() counts:** `got == 2` instead of `got >= 2`. batchSize is `2 + rng.Intn(2)` (2 or 3), so the batch count for N steps varies. Assert `>= minBatches`.
- **Assuming Phase1TechsPerDay=1 in new distributed tests:** The whole point of DayDigest accuracy testing is to exercise multi-technique days. Explicitly set 3+ techs/day.
- **Replacing tests instead of updating in place:** D-07 forbids full replacement. Audit + comment + targeted fix.
- **Using rand.Source injection into runPoC():** D-05 forbids this. Tests must be resilient to slot timing changes.
- **Testing exact PassCount/FailCount sequences:** The alternating runner's call-count resets between engine instances but accumulates within a run. Verify `PassCount + FailCount == TechniqueCount` not the individual split.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| After() counting | New clock wrapper | `afterCountClock` in poc_schedule_test.go | Already proven, mutex-safe |
| DayDigest snapshotting | New capture mechanism | `digestCaptureClock` in poc_test.go | Already proven, mutex-safe |
| Phase completion polling | Custom wait loop | `waitForPhase(eng, target, timeout)` | Already handles 10ms polling + deadline |
| Campaign fixture with N steps | New registry builder | Extend existing `testRegistry()` inline | Simple map assignment |
| Alternating pass/fail runner | New runner func | Inline closure (see `newPoCEngineWithCampaign`) | Straightforward, no abstraction needed |

## Existing Test Audit: After() Call Accounting

This is the core work of D-06. Every test with a `blockAt` value or After() assertion must be verified.

### Tests with blockAt values (stopOnNthClock-based)

| Test | blockAt | Phase1TechsPerDay | After() per Phase1 day | Assessment |
|------|---------|-------------------|------------------------|------------|
| `TestPoCStop_DuringDayWait` | 2 | 1 (via pocConfig) | 1 | Correct — day 1 fires (call 1), day 2 blocks (call 2) |
| `TestPoCStop_BetweenPhaseTransitions` | 2 | 1 (via pocConfig) | 1 | Correct — Phase1 day 1 fires (call 1), gap day 1 blocks (call 2) |
| `TestPoCStop_DuringGapDays` | 2 | 1 (via pocConfig) | 1 | Correct — Phase1 day 1 fires (call 1), gap day 1 blocks (call 2) |
| `TestPoCStop_ImmediateAfterStart` | 1 | 1 (via pocConfig) | 1 | Correct — first slot wait blocks (call 1) |

**Verdict:** All four pass because they use 1 tech/day. Under distributed scheduling, 1 tech = 1 slot = 1 After() — identical to the old behavior for N=1. The tests pass by correctness, not coincidence. However, they lack explanatory comments that document this assumption. Phase 20 must add comments.

### Tests with After() count assertions

| Test | Assertion | Techs | Expected min After() | Assessment |
|------|-----------|-------|----------------------|------------|
| `TestPoCPhase1_DistributedSlots` | `got >= 2` | 2 (Phase1TechsPerDay=2) | 2 | Correct — already uses `>=` |
| `TestPoCPhase2_BatchedSlots` | `got >= 2` | 5 steps, batches 2-3 | 2 | Correct — already uses `>=` |

### Tests relying on single-technique-per-day behavior

| Test | Implicit assumption | Risk |
|------|---------------------|------|
| `TestDayDigest_PrePopulated` | TechniqueCount=1 for Phase1 days | None — sets Phase1TechsPerDay=1 explicitly via pocConfig |
| `TestDayDigest_Counts` | PassCount+FailCount=1 per Phase1 day | None — same |
| `TestDayDigest_Lifecycle` | Complete status per day | None — lifecycle not count-dependent |
| `TestPoCDayCounter` | totalDays arithmetic only | None — no slot-count dependency |
| `TestPoCClockInjection` | Clock advances (not zero) | None — structural only |

**Coverage gap (D-03):** No existing test runs Phase 1 with 3+ techs/day and verifies TechniqueCount accuracy. This is the gap that the new test must close.

## Common Pitfalls

### Pitfall 1: batchSize is non-deterministic
**What goes wrong:** Phase 2 uses `batchSize := 2 + rng.Intn(2)` (2 or 3). Asserting `got == 2` for a 5-step campaign will fail ~50% of the time (when batches are size 2 → 3 batches, not 2).
**Why it happens:** The rand source is seeded from clock.Now().UnixNano() — different each run.
**How to avoid:** Always assert `got >= minBatches` where `minBatches = ceil(steps / maxBatchSize)`. For 5 steps: `>= 2` (ceil(5/3)=2).

### Pitfall 2: fakeClock.After() mutates the clock
**What goes wrong:** `fakeClock.After(d)` advances `fc.now` by `d`. When wrapping `fakeClock` in `afterCountClock`, the inner clock advances. Assertions about `fc.Now()` after the run reflect accumulated slot durations, not wall time.
**Why it happens:** Design choice in `fakeClock` — time advances deterministically.
**How to avoid:** Don't assert absolute clock time. Assert relative outcomes (PassCount, FailCount, TechniqueCount). This is already the established pattern.

### Pitfall 3: TechniqueCount vs PassCount+FailCount mismatch
**What goes wrong:** `TechniqueCount` is pre-populated from `Phase1TechsPerDay` at runPoC() start. `PassCount+FailCount` accumulates as techniques run. A test that verifies only `TechniqueCount` misses the case where fewer techniques actually ran.
**Why it happens:** Pre-population and execution are separate code paths.
**How to avoid:** New DayDigest accuracy test must assert BOTH `TechniqueCount == N` AND `PassCount+FailCount == N` for Phase 1 days.

### Pitfall 4: alternating runner state
**What goes wrong:** The alternating runner in `newPoCEngineWithCampaign` closes over `callCount` which resets per engine construction. If a test constructs multiple engines or campaigns, call-count expectations break.
**Why it happens:** Closure captures the variable at construction time.
**How to avoid:** For multi-technique Phase 1 tests, use `fakeRunner(0)` (always succeeds) and verify `PassCount == N`, `FailCount == 0` as a simpler invariant. Or construct the alternating runner explicitly with documented expectations.

### Pitfall 5: Race detector failures on unguarded clock wrappers
**What goes wrong:** Engine goroutine and test goroutine both call the clock concurrently. If a new clock wrapper accesses shared state without mutex, `-race` flag catches it.
**Why it happens:** `runPoC()` runs in a goroutine; test reads snapshot state from main goroutine.
**How to avoid:** Follow the pattern in `afterCountClock` — `sync.Mutex` protecting all shared fields. Read state only after `waitForPhase` confirms engine completion.

## Code Examples

### New DayDigest Accuracy Test (structural sketch)

```go
// TestDayDigest_DistributedCounts verifies TechniqueCount and PassCount+FailCount
// are accurate when Phase 1 runs 3 techniques per day across distributed slots (D-03).
func TestDayDigest_DistributedCounts(t *testing.T) {
    reg := testRegistry(
        testTechnique("T1087", "discovery", "discovery"),
        testTechnique("T1059", "discovery", "execution"),
        testTechnique("T1078", "attack", "persistence"),
        // 3+ techniques so Phase1TechsPerDay=3 is satisfiable
    )

    fc := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
    eng := New(reg, nil)
    eng.clock = fc
    eng.runner = fakeRunner(0) // all pass; simplifies PassCount assertion

    cfg := Config{
        PoCMode:            true,
        Phase1DurationDays: 2,
        Phase1TechsPerDay:  3,  // 3 slots/day = 3 After() per day
        Phase1WindowStart:  0,
        Phase1WindowEnd:    23,
        GapDays:            0,
        Phase2DurationDays: 0,
    }
    if err := eng.Start(cfg); err != nil {
        t.Fatalf("Start: %v", err)
    }
    if !waitForPhase(eng, PhaseDone, 5*time.Second) {
        t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
    }

    digests := eng.GetDayDigests()
    if len(digests) != 2 {
        t.Fatalf("expected 2 DayDigest entries, got %d", len(digests))
    }
    for i, d := range digests {
        if d.TechniqueCount != 3 {
            t.Errorf("day %d: TechniqueCount=%d, want 3", i+1, d.TechniqueCount)
        }
        if d.PassCount+d.FailCount != 3 {
            t.Errorf("day %d: PassCount+FailCount=%d, want 3", i+1, d.PassCount+d.FailCount)
        }
        if d.Status != DayComplete {
            t.Errorf("day %d: Status=%q, want DayComplete", i+1, d.Status)
        }
    }
}
```

### Documenting blockAt correctness (comment pattern for existing tests)

```go
// TestPoCStop_DuringDayWait verifies stop signal during scheduling wait aborts the engine.
// Under distributed scheduling, Phase 1 with 1 tech/day produces exactly 1 After() per day
// (1 slot = 1 After() call). blockAt=2 therefore means: day 1 slot fires (call 1), day 2
// slot blocks (call 2). This is correct and not coincidental — single-technique days have
// a 1:1 slot-to-After() correspondence with the old fire-all-at-once behavior.
func TestPoCStop_DuringDayWait(t *testing.T) {
```

### Asserting After() count for multi-tech Phase 1

```go
// Phase 1 with N techs/day = N After() calls per day (one per slot).
// With 3 techs/day and 2 days: minimum 6 After() calls.
got := cc.getCount()
if got < 6 {
    t.Errorf("expected >= 6 After() calls for 3 distributed slots x 2 days, got %d", got)
}
```

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` (stdlib) |
| Config file | None — `go test` directly |
| Quick run command | `go test ./internal/engine/... -count=1 -timeout 30s` |
| Full suite command | `go test ./internal/engine/... -count=1 -race -timeout 60s` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| POC-04 | All existing tests pass with distributed runPoC() | regression | `go test ./internal/engine/... -count=1` | Yes — 35 tests currently passing |
| POC-04 | Phase 1 not all dispatched at single clock tick | unit | same | Yes — `TestPoCPhase1_DistributedSlots` |
| POC-04 | Phase 2 dispatched in groups of 2-3 | unit | same | Yes — `TestPoCPhase2_BatchedSlots` |
| POC-04 | DayDigest TechniqueCount accurate under distributed scheduling | unit | same | No — `TestDayDigest_DistributedCounts` needs creation |

### Sampling Rate
- **Per task commit:** `go test ./internal/engine/... -count=1 -timeout 30s`
- **Per wave merge:** `go test ./internal/engine/... -count=1 -race -timeout 60s`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/engine/poc_schedule_test.go` — add `TestDayDigest_DistributedCounts` (covers POC-04 criterion 4)

*(All other gaps: existing test infrastructure is complete — no new files, framework, or fixtures needed)*

## Environment Availability

Step 2.6: All tooling is the Go standard library. No external dependencies.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | `go test` | Yes | (project already compiles) | — |

## Sources

### Primary (HIGH confidence)
- Direct code inspection: `internal/engine/engine.go` — `runPoC()`, `randomSlotsInWindow()`, `DayDigest` struct
- Direct code inspection: `internal/engine/poc_test.go` — all 6 test functions, all clock wrappers
- Direct code inspection: `internal/engine/poc_schedule_test.go` — 3 Phase 19 tests, `afterCountClock`
- Direct code inspection: `internal/engine/engine_test.go` — 17 tests, `pocConfig()`, `newPoCEngineWithCampaign()`
- Live test run: `go test ./internal/engine/... -count=1` — all 35 tests pass

### Secondary (MEDIUM confidence)
- N/A — all findings are from direct source inspection

## Metadata

**Confidence breakdown:**
- Existing test audit: HIGH — ran all 35 tests, read all source
- New test design: HIGH — patterns well-established in codebase
- blockAt analysis: HIGH — traced After() call sequence in runPoC() for 1-tech/day case
- DayDigest gap: HIGH — confirmed by reading TestDayDigest_Counts (Phase1TechsPerDay=1)

**Research date:** 2026-04-11
**Valid until:** Stable — no external dependencies; valid until engine.go changes
