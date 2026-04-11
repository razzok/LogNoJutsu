# Phase 19: Distributed Technique Scheduling - Research

**Researched:** 2026-04-10
**Domain:** Go scheduling logic — runPoC() rewrite, random jitter, window-based execution
**Confidence:** HIGH (full codebase read, no external library dependencies)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Replace `Phase1DailyHour` and `Phase2DailyHour` with start/end hour pairs (`WindowStart`/`WindowEnd`). The existing single-hour fields become the window start/end instead of a single execution time.
- **D-02:** Default window is 08:00–17:00 (business hours).
- **D-04:** Fully random distribution — pick N random times within the window for each day's techniques. No even spacing or slot-based approach.
- **D-07:** Techniques within a Phase 2 batch use the existing `DelayBetweenTechniques` setting for short delays between each technique in the burst.
- **D-09:** Phase 1 executes one technique at a time at each random slot (not batched).

### Claude's Discretion
- **D-03:** Whether Phase 1 and Phase 2 share one window config or have separate `Phase1WindowStart`/`Phase1WindowEnd` and `Phase2WindowStart`/`Phase2WindowEnd` fields. Decide based on what fits the existing `PoCConfig` struct pattern.
- **D-05:** Whether `NextScheduledRun` in the status API shows the exact next technique execution time or just the window boundaries. Decide based on what fits the existing status polling pattern.
- **D-06:** Batch size approach — fixed random 2–3, or configurable `Phase2BatchSize`. Decide what fits engagement workflow best.
- **D-08:** How campaign `step.DelayAfter` interacts with the new jitter scheduling.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| POC-01 | Phase 1 techniques execute one at a time at random intervals throughout the day instead of all at the scheduled hour | Replace `nextOccurrenceOfHour()` with window-based random slot generator; run one technique per slot |
| POC-02 | Phase 2 techniques execute in small batches (2–3) at random intervals throughout the day instead of all at the scheduled hour | Generate N/batchSize random time slots per day; fire `delayBetween()` between techniques within each batch |
| POC-03 | Random jitter is bounded within a configurable daily time window (e.g., start hour to end hour) | New `WindowStart`/`WindowEnd` (or per-phase pairs) in `PoCConfig`; jitter generator enforces bounds |
</phase_requirements>

---

## Summary

Phase 19 rewrites `runPoC()` in `internal/engine/engine.go`. The current implementation fires all techniques at a fixed hour using `nextOccurrenceOfHour()`. The new approach generates N random `time.Time` values within a configurable window (e.g., 08:00–17:00), sorts them, and fires one technique (Phase 1) or a small batch of techniques (Phase 2) at each slot. All timing must continue going through `e.clock.After()` so that existing `fakeClock`-based tests stay deterministic.

The change touches: `PoCConfig` struct (new fields), `runPoC()` (Phase 1 and Phase 2 inner loops), the `nextOccurrenceOfHour()` helper (to be replaced or supplemented), and `index.html` + `updatePoCSchedule()` (form inputs and schedule preview). DayDigest `TechniqueCount` must remain pre-populated at the correct value regardless of how execution is distributed — the count reflects what is scheduled, not what has run.

**Primary recommendation:** Extract a `randomSlotsInWindow(n int, windowStart, windowEnd int, date time.Time, rng *rand.Rand) []time.Duration` helper. For each day, call it once to get all wait durations, then loop over slots, calling `waitOrStop()` for each inter-slot gap and executing technique(s) at each slot. Use `math/rand` (already imported in engine.go) with `rand.New(rand.NewSource(time.Now().UnixNano()))` seeded once at runPoC() start so that tests can override via the `Clock` interface (or inject a fixed seed through a new helper).

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `math/rand` | stdlib (Go 1.26) | Random slot generation | Already imported in engine.go; used for `rand.Shuffle()` today |
| `sort` | stdlib | Sort random time slots into ascending order within a day | No third-party dependency needed |
| `time` | stdlib | Duration arithmetic, window boundary computation | All timing goes through `e.clock` interface |

### No New External Dependencies
This phase is pure internal scheduling logic. No new `go get` packages are needed. All required primitives (`rand`, `sort`, `time`) are already in the import list or stdlib.

---

## Architecture Patterns

### Existing Pattern: Clock Interface
```go
// Source: internal/engine/engine.go line 156-160
type Clock interface {
    Now() time.Time
    After(d time.Duration) <-chan time.Time
}
```
All wait calls go through `e.waitOrStop(d)`, which calls `e.clock.After(d)`. This is non-negotiable — any random jitter waits must use `e.waitOrStop()`, never `time.Sleep()`.

### Existing Pattern: fakeClock in Tests
`fakeClock.After(d)` advances `f.now` by `d` and fires immediately. This means even if `runPoC()` calls `waitOrStop()` ten times per day, the fake clock fires all of them instantly — tests remain fast. The new distributed logic (multiple `waitOrStop()` calls per day instead of one) is automatically compatible.

### Recommended: randomSlotsInWindow Helper
```go
// New helper — all timing in absolute time.Time, returned as sorted []time.Duration
// durations relative to "now" at day start.
func randomSlotsInWindow(n, windowStart, windowEnd int, dayStart time.Time, src rand.Source) []time.Duration {
    if windowEnd <= windowStart {
        windowEnd = windowStart + 1 // guard: minimum 1-hour window
    }
    windowSecs := (windowEnd - windowStart) * 3600
    rng := rand.New(src)
    offsets := make([]int, n)
    for i := range offsets {
        offsets[i] = rng.Intn(windowSecs)
    }
    sort.Ints(offsets)
    base := time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(),
        windowStart, 0, 0, 0, dayStart.Location())
    now := dayStart
    durations := make([]time.Duration, n)
    for i, off := range offsets {
        slotTime := base.Add(time.Duration(off) * time.Second)
        if slotTime.Before(now) {
            slotTime = slotTime.Add(24 * time.Hour) // already passed: defer to tomorrow
        }
        durations[i] = slotTime.Sub(now)
        now = slotTime
    }
    return durations
}
```
This function returns the **inter-slot wait durations** (not absolute times). The caller does:
```go
slots := randomSlotsInWindow(techsPerDay, cfg.Phase1WindowStart, cfg.Phase1WindowEnd, e.clock.Now(), randSrc)
for i, slot := range slots {
    if !e.waitOrStop(slot) { e.abort(); return }
    // run technique i
}
```

### Recommended Phase 2 Batching Strategy (D-06 Discretion)
Use **fixed random batch size (2–3)** rather than a configurable `Phase2BatchSize`. Rationale: engagement workflow value comes from unpredictability. A configurable field adds UI surface for no practical gain at this stage. Implementation:
```go
batchSize := 2 + rand.Intn(2) // 2 or 3 per batch
```
Number of batches per day = `ceil(len(techniques) / batchSize)`. Generate one random slot per batch. Within each batch, techniques fire sequentially with `delayBetween()` between them (D-07).

### Recommended: Separate Window Fields per Phase (D-03 Discretion)
Use **separate `Phase1WindowStart`/`Phase1WindowEnd`/`Phase2WindowStart`/`Phase2WindowEnd`** rather than a single shared window. Rationale: Phase 1 (discovery) and Phase 2 (attack) occur on different days, so they cannot conflict. Having separate fields lets operators set, e.g., discovery at 09:00–12:00 and attacks at 13:00–17:00 during client engagements. This follows the existing `Phase1DailyHour` / `Phase2DailyHour` naming pattern exactly. Four new int fields with JSON tags.

### Recommended: NextScheduledRun Shows Next Slot Time (D-05 Discretion)
Set `NextScheduledRun` to the absolute time of the **next technique slot** (first slot of next wait or current slot in progress). This preserves the countdown UX in `pollStatus()` — the JS already does `new Date(s.next_scheduled_run) - new Date()` to display `Nh Nm`. Showing window boundaries instead would break the countdown semantics.

### Recommended: DelayAfter Interaction (D-08 Discretion)
Campaign `step.DelayAfter` continues to apply between techniques **within a batch** (in addition to `delayBetween()`). The batch is a burst — `DelayAfter` already represents intra-burst delay. Inter-batch delays are handled by the random slot waits. Do not apply `DelayAfter` between batches.

### PoCConfig Struct Changes
Replace:
```go
Phase1DailyHour int `json:"phase1_daily_hour"` // 0-23: hour at which to run each day
Phase2DailyHour int `json:"phase2_daily_hour"` // 0-23: hour at which to run each day
```
With:
```go
Phase1WindowStart int `json:"phase1_window_start"` // 0-23: window start hour (default 8)
Phase1WindowEnd   int `json:"phase1_window_end"`   // 0-23: window end hour (default 17)
Phase2WindowStart int `json:"phase2_window_start"` // 0-23: window start hour (default 8)
Phase2WindowEnd   int `json:"phase2_window_end"`   // 0-23: window end hour (default 17)
```

### Log Message Update
The startup log `simlog.Info(fmt.Sprintf("[PoC] Multi-day simulation: Phase1=%dd (%d techs/day @ %02d:00) ..."))` references `cfg.Phase1DailyHour` and `cfg.Phase2DailyHour`. Update to show window ranges: `@ %02d:00–%02d:00`.

### Anti-Patterns to Avoid
- **Do not use `time.Sleep()` anywhere in runPoC()**: All waits must go through `e.waitOrStop()` → `e.clock.After()`.
- **Do not generate slots all at once at runPoC() start**: Generate each day's slots at the top of that day's loop iteration, using `e.clock.Now()` as the anchor. This ensures "already-passed" detection is relative to the actual current time at day start.
- **Do not change DayDigest.TechniqueCount**: It stays pre-populated from the total technique count for the day. It does not need to equal the number of slots — it reflects scheduling intent.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Random int in range | Custom LCG or modulo-bias hack | `rand.Intn(n)` from `math/rand` | Already imported, correct distribution |
| Sorting slot offsets | Bubble sort | `sort.Ints()` | Stdlib, zero cost |
| Seeding random source | Global `rand.Seed()` (deprecated Go 1.20+) | `rand.New(rand.NewSource(seed))` | Avoids global state, testable |

---

## Common Pitfalls

### Pitfall 1: Slot Already Passed When Day Starts
**What goes wrong:** If the engine starts at 14:00 and `WindowStart` is 08:00, the first computed slots for "today" are in the past. All waits would be negative durations.
**Why it happens:** The random offset is relative to `WindowStart` (08:00), but `e.clock.Now()` is already past that.
**How to avoid:** After computing slot times, if a slot is before `e.clock.Now()`, add 24 hours to defer it to the next calendar day — OR enforce that the first day's window check skips to the next day if the window has already closed. The safest approach: always compute slots for the next calendar day (add 1 day to today's window anchor).
**Warning signs:** `waitOrStop(d)` called with a negative duration causes `e.clock.After()` to fire immediately with zero advancement — techniques would all run instantly.

### Pitfall 2: rand.Rand Not Thread-Safe
**What goes wrong:** `math/rand` global functions use a mutex but a `*rand.Rand` instance does not. If the engine uses the same `*rand.Rand` from multiple goroutines (it does not today, but worth noting), races occur.
**How to avoid:** Create one `rand.New(rand.NewSource(...))` per `runPoC()` call and use it only within that function (single goroutine).

### Pitfall 3: Tests Counting After() Calls Break
**What goes wrong:** Existing `captureClock`, `dayCaptureClock`, `digestCaptureClock`, and `stopOnNthClock` count or trigger on every `After()` call. Phase 19 multiplies the number of `After()` calls per day (from 1 to N, where N = techsPerDay or batchCount). Tests that use `blockAt=2` in `stopOnNthClock` to mean "block at day 2 wait" will now block earlier or later.
**Why it happens:** The `blockAt` index was calibrated against 1 `After()` call per day. With 3 techniques per day, day 1 alone consumes 3 `After()` calls.
**How to avoid:** Phase 20 is responsible for updating the tests (per REQUIREMENTS.md traceability). Phase 19 must NOT update tests — it only ships the engine change. Tests will fail after Phase 19 until Phase 20 fixes them. This is expected and documented in STATE.md: "Phase 20 must update them without breaking the captureClock/fakeClock infrastructure."

### Pitfall 4: updatePoCSchedule() Shows Wrong Preview
**What goes wrong:** The UI schedule preview (`updatePoCSchedule()` in index.html) still shows `daily ${p1Hour}:00` after Phase 19. The form now sends `phase1_window_start` / `phase1_window_end` but the preview JS still reads `pocP1Hour`.
**How to avoid:** Replace `pocP1Hour` and `pocP2Hour` form inputs with `pocP1WindowStart`/`pocP1WindowEnd` and `pocP2WindowStart`/`pocP2WindowEnd` inputs. Update `updatePoCSchedule()` to preview a range (e.g., `08:00–17:00`).

### Pitfall 5: Zero-width or Inverted Window
**What goes wrong:** `WindowStart >= WindowEnd` produces zero or negative `windowSecs`, causing `rand.Intn(0)` panic.
**How to avoid:** Guard in `randomSlotsInWindow`: `if windowEnd <= windowStart { windowEnd = windowStart + 1 }`. Also validate in `Start()` or at the top of `runPoC()`.

### Pitfall 6: DayDigest TechniqueCount Misreported for Batched Phase 2
**What goes wrong:** If `phase2TechCount` is set to `len(c.Steps)` but techniques are now split into batches, the pre-population code might be tempted to set count = number of batches instead of number of techniques.
**How to avoid:** Keep `TechniqueCount = len(c.Steps)` (total techniques). `PassCount` and `FailCount` increment per technique, not per batch. This is unchanged.

---

## Code Examples

### Current Day-Wait Pattern (to be replaced)
```go
// Source: internal/engine/engine.go line 580
d := nextOccurrenceOfHour(cfg.Phase1DailyHour, e.clock.Now())
nextRun := e.clock.Now().Add(d)
// ... set status ...
if !e.waitOrStop(d) { e.abort(); return }
// run all techsPerDay techniques back-to-back
```

### New Day-Loop Pattern (Phase 1)
```go
// After setting DayActive status:
randSrc := rand.NewSource(time.Now().UnixNano()) // seeded once per runPoC() call
slots := randomSlotsInWindow(techsPerDay, cfg.Phase1WindowStart, cfg.Phase1WindowEnd, e.clock.Now(), randSrc)
for i, slotDelay := range slots {
    // Update NextScheduledRun before each wait
    e.mu.Lock()
    e.status.NextScheduledRun = e.clock.Now().Add(slotDelay).Format(time.RFC3339)
    e.status.CurrentStep = fmt.Sprintf("PoC Phase 1 — Day %d of %d — Slot %d/%d", globalDay, totalDays, i+1, techsPerDay)
    e.mu.Unlock()

    if !e.waitOrStop(slotDelay) { e.abort(); return }

    t := discoveryTechs[(start+i)%len(discoveryTechs)]
    e.runTechnique(t)
    // update pass/fail counts + heartbeat
}
```

### New Day-Loop Pattern (Phase 2, Campaign, Batched)
```go
steps := campaign.Steps
batchSize := 2 + rng.Intn(2) // 2 or 3
numBatches := (len(steps) + batchSize - 1) / batchSize
slots := randomSlotsInWindow(numBatches, cfg.Phase2WindowStart, cfg.Phase2WindowEnd, e.clock.Now(), randSrc)
for b, slotDelay := range slots {
    if !e.waitOrStop(slotDelay) { e.abort(); return }
    batchStart := b * batchSize
    batchEnd := batchStart + batchSize
    if batchEnd > len(steps) { batchEnd = len(steps) }
    for _, step := range steps[batchStart:batchEnd] {
        if e.isStopped() { e.abort(); return }
        t, exists := e.registry.Techniques[step.TechniqueID]
        if !exists { continue }
        e.runTechnique(t)
        // update pass/fail counts + heartbeat
        if step.DelayAfter > 0 {
            if !e.waitOrStop(time.Duration(step.DelayAfter) * time.Second) { e.abort(); return }
        }
        e.delayBetween()
    }
}
```

### Index.html Config Payload (new fields)
```javascript
// Replace phase1_daily_hour / phase2_daily_hour with:
phase1_window_start: parseInt(document.getElementById('pocP1WindowStart').value) || 8,
phase1_window_end:   parseInt(document.getElementById('pocP1WindowEnd').value)   || 17,
phase2_window_start: parseInt(document.getElementById('pocP2WindowStart').value) || 8,
phase2_window_end:   parseInt(document.getElementById('pocP2WindowEnd').value)   || 17,
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| All techniques at one fixed hour | Random slots within configurable window | Phase 19 | More realistic attacker behavior; better UEBA baseline blending |
| Single `Phase1DailyHour` int | `Phase1WindowStart` + `Phase1WindowEnd` pair | Phase 19 | Breaking config change — existing form inputs replaced |

---

## Open Questions

1. **Seed reproducibility for debugging**
   - What we know: `rand.New(rand.NewSource(time.Now().UnixNano()))` produces a different schedule every run.
   - What's unclear: If an engagement needs to reproduce a specific slot schedule (e.g., for audit logs), there is no seed replay mechanism.
   - Recommendation: Out of scope for Phase 19. If needed in future, add optional `PoCRandSeed int64` config field. For now, UnixNano seed is sufficient.

2. **Gap days still use Phase2DailyHour for their wait anchor**
   - What we know: Current gap logic calls `nextOccurrenceOfHour(cfg.Phase2DailyHour, ...)` as an arbitrary daily heartbeat.
   - What's unclear: With no daily hour field, gap days have no anchor. They could use `Phase2WindowStart` or just wait 24 hours.
   - Recommendation: Gap days wait exactly 24 hours from when the previous day completed (simple `e.waitOrStop(24 * time.Hour)`). Gap days have no techniques, so there is no need for a window.

---

## Environment Availability

Step 2.6: SKIPPED — this phase is a pure code change within an existing Go project. No external tools, services, databases, or CLI utilities beyond the project's own `go` toolchain are needed.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) — Go 1.26.1 |
| Config file | none (standard `go test`) |
| Quick run command | `go test ./internal/engine/... -timeout 30s` |
| Full suite command | `go test ./... -timeout 60s` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| POC-01 | Phase 1 executes one technique per random slot, not all at fixed hour | unit | `go test ./internal/engine/... -run TestPoCPhase1_DistributedSlots -timeout 30s` | ❌ Wave 0 |
| POC-02 | Phase 2 executes in batches at random slots | unit | `go test ./internal/engine/... -run TestPoCPhase2_BatchedSlots -timeout 30s` | ❌ Wave 0 |
| POC-03 | Random slots fall within configured window bounds | unit | `go test ./internal/engine/... -run TestRandomSlotsInWindow -timeout 30s` | ❌ Wave 0 |

**Note:** POC-04 (existing test updates) is Phase 20's responsibility. Phase 19 ships the engine change; Phase 20 repairs the existing tests. After Phase 19, some existing tests that use `stopOnNthClock` or exact After()-call counts WILL fail — this is expected and documented in STATE.md.

### Sampling Rate
- **Per task commit:** `go test ./internal/engine/... -timeout 30s`
- **Per wave merge:** `go test ./... -timeout 60s`
- **Phase gate:** Full suite green (with POC-04 deferred to Phase 20 — existing After()-count tests may fail until Phase 20)

### Wave 0 Gaps
- [ ] New test function `TestRandomSlotsInWindow` in `poc_test.go` or a new `poc_schedule_test.go` — covers POC-03 (window boundary invariant)
- [ ] New test function `TestPoCPhase1_DistributedSlots` — covers POC-01 (verifies multiple After() calls per day, one technique per slot)
- [ ] New test function `TestPoCPhase2_BatchedSlots` — covers POC-02 (verifies batch grouping and multiple slots per day)

---

## Sources

### Primary (HIGH confidence)
- `internal/engine/engine.go` — Full read: `PoCConfig` struct (lines 45–67), `runPoC()` (lines 508–762), `nextOccurrenceOfHour()` (lines 498–506), `waitOrStop()` (lines 1009–1014), `delayBetween()` (lines 898–903), `Clock` interface (lines 156–160)
- `internal/engine/engine_test.go` — Full read: `fakeClock`, `captureClock`, `dayCaptureClock`, `digestCaptureClock`, `stopOnNthClock`, all PoC test functions
- `internal/engine/poc_test.go` — Full read: `TestPoCDayCounter_Monotonic`, `TestDayDigest_PendingActiveComplete`, `TestPoCStop_DuringDayWait`, `TestPoCStop_BetweenPhaseTransitions`
- `internal/server/static/index.html` — Form inputs (`pocP1Hour`, `pocP2Hour`), `updatePoCSchedule()`, `pollStatus()` countdown logic
- `internal/server/server.go` — `handleStart()`: JSON decode of `engine.Config` directly — no field name mapping layer

### Secondary (MEDIUM confidence)
- Go standard library documentation for `math/rand` — `rand.New()`, `rand.NewSource()`, `rand.Intn()` usage pattern (verified against Go 1.26 which is installed)

---

## Metadata

**Confidence breakdown:**
- Implementation approach: HIGH — full source read, no guesswork
- Standard stack: HIGH — no new dependencies, all stdlib
- Architecture: HIGH — patterns derived directly from existing code
- Pitfalls: HIGH — derived from concrete code analysis (After() count impact on existing tests is a confirmed structural consequence)
- Test gaps: HIGH — confirmed by running `go test ./internal/engine/... -list ".*"` and cross-referencing against requirements

**Research date:** 2026-04-10
**Valid until:** Until Phase 20 begins (stable codebase, no external dependencies to expire)
