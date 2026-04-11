# Phase 19: Distributed Technique Scheduling - Context

**Gathered:** 2026-04-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Rewrite runPoC() so techniques execute at random times throughout the day instead of all firing at one fixed hour. Phase 1 runs one technique at a time; Phase 2 runs in small batches. Both phases distribute execution within a configurable daily time window.

</domain>

<decisions>
## Implementation Decisions

### Time Window Configuration
- **D-01:** Replace `Phase1DailyHour` and `Phase2DailyHour` with start/end hour pairs (`WindowStart`/`WindowEnd`). The existing single-hour fields become the window start/end instead of a single execution time.
- **D-02:** Default window is 08:00-17:00 (business hours) — blends technique execution with normal employee activity on the SIEM.
- **D-03:** Claude's Discretion: Whether Phase 1 and Phase 2 share one window config or have separate `Phase1WindowStart`/`Phase1WindowEnd` and `Phase2WindowStart`/`Phase2WindowEnd` fields. Decide based on what fits the existing `PoCConfig` struct pattern.

### Jitter Algorithm
- **D-04:** Fully random distribution — pick N random times within the window for each day's techniques. No even spacing or slot-based approach. This produces more unpredictable, realistic attacker behavior patterns.
- **D-05:** Claude's Discretion: Whether `NextScheduledRun` in the status API shows the exact next technique execution time or just the window boundaries. Decide based on what fits the existing status polling pattern (`pollStatus()` in index.html).

### Phase 2 Batching
- **D-06:** Claude's Discretion: Batch size approach — fixed random 2-3, or configurable `Phase2BatchSize`. Decide what fits engagement workflow best.
- **D-07:** Techniques within a batch use the existing `DelayBetweenTechniques` setting for short delays between each technique in the burst. This preserves the "burst of attacker activity" detection pattern.
- **D-08:** Claude's Discretion: How campaign `step.DelayAfter` interacts with the new jitter scheduling. Decide what preserves campaign semantics while enabling distribution.

### Phase 1 Execution
- **D-09:** Phase 1 executes one technique at a time at each random slot (not batched). Each day's `techsPerDay` techniques get N random times, one technique per time slot.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### PoC Engine
- `internal/engine/engine.go` lines 508-630 — Current `runPoC()` implementation (Phase 1 discovery loop)
- `internal/engine/engine.go` lines 632-760 — Current `runPoC()` Phase 2 campaign/attack loop
- `internal/engine/engine.go` lines 45-67 — `PoCConfig` struct with current fields
- `internal/engine/engine.go` lines 92-112 — `DayDigest` struct (must continue working)
- `internal/engine/engine.go` lines 898-905 — `delayBetween()` helper

### Tests
- `internal/engine/engine_test.go` — All PoC scheduling tests using fakeClock/captureClock patterns
- `internal/engine/poc_test.go` — Additional PoC tests (day counter, stop signal, DayDigest lifecycle)

### Prior Decisions
- Phase 10: Clock interface injected via `e.clock.Now()` / `e.clock.After()` — all timing must use this
- Phase 11: DayDigest pre-populated at runPoC() start with all days as "pending"
- Phase 13: captureClock/digestCaptureClock patterns for race-free test assertions

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `e.clock` interface: All timing goes through injectable clock — enables deterministic testing
- `e.waitOrStop(duration)`: Blocking wait that respects stop signal — reusable for jitter waits
- `nextOccurrenceOfHour(hour, now)`: Computes wait duration until next occurrence of an hour — will need replacement with window-based scheduling
- `e.delayBetween()`: Short delay between techniques — reuse within Phase 2 batches

### Established Patterns
- `PoCConfig` struct with JSON tags — new fields must follow same pattern
- DayDigest lifecycle: pending → active → complete with pass/fail/heartbeat tracking per technique
- Status updates via `e.mu.Lock()` / `e.mu.Unlock()` — thread-safe status for API polling

### Integration Points
- `PoCConfig` struct fields read from web UI form submission (server.go `handleStart`)
- `index.html` PoC form inputs map to config fields — new window fields need corresponding form inputs
- `/api/status` returns `NextScheduledRun` — polling JS uses this to show countdown
- DayDigest tracking must continue feeding `/api/poc/days` accurately

</code_context>

<specifics>
## Specific Ideas

- The original intent was always to spread techniques throughout the day, not fire them all at once. This is a correction to match the original vision.
- Business hours default (08:00-17:00) chosen because technique execution should blend with normal employee SIEM activity during client engagements.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 19-distributed-technique-scheduling*
*Context gathered: 2026-04-10*
