# Phase 11: Daily Tracking Backend & Campaign Delay - Context

**Gathered:** 2026-04-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Add per-day execution tracking to the PoC engine via a DayDigest struct, expose it through a new `/api/poc/days` endpoint, and apply the existing `CampaignStep.DelayAfter` field during PoC Phase 2 campaign execution. No UI changes (Phase 12 handles display).

</domain>

<decisions>
## Implementation Decisions

### DayDigest Struct (TRACK-01, TRACK-02)
- **D-01:** DayDigest carries counts only — day number, phase label, status enum (pending/active/complete), technique count, pass count, fail count, start timestamp, end timestamp, last heartbeat timestamp. No per-technique detail (that data lives in `engine.status.Results`).
- **D-02:** All DayDigest entries are pre-populated as "pending" at the top of `runPoC()` before the first day loop begins. The full schedule is visible from the very first `/api/poc/days` poll.
- **D-03:** DayDigest slice lives as a `[]DayDigest` field on the Engine struct (not inside Status). Guarded by the existing `sync.RWMutex`. Returned via a dedicated getter method.

### API Endpoint (TRACK-03)
- **D-04:** `GET /api/poc/days` returns the full `[]DayDigest` array. No per-day lookup endpoint needed — max ~30 entries for any PoC run.
- **D-05:** Endpoint is behind `authMiddleware`, consistent with `/api/status`, `/api/logs`, and all other data endpoints.
- **D-06:** When no PoC is running (or engine is idle), return HTTP 200 with empty JSON array `[]`. Phase 12 UI handles the empty state.

### Campaign delay_after (CAMP-01)
- **D-07:** `CampaignStep.DelayAfter` is treated as seconds. A value of 300 means a 5-minute delay.
- **D-08:** Delay uses the injectable clock via `e.clock.After(time.Duration(step.DelayAfter) * time.Second)`. Tests with fakeClock skip delays instantly.
- **D-09:** Delay is interruptible — use existing `waitOrStop(d)` which selects on `e.clock.After` AND `stopCh`. User can stop the engine mid-delay.

### Heartbeat (TRACK-04)
- **D-10:** LastHeartbeat is updated at key events only: day start, each technique execution, day completion. No background timer goroutine needed.
- **D-11:** Heartbeat semantics: proves the engine was alive during the execution window. If timestamp is stale (e.g., >10min for an hourly schedule), something may be wrong. Phase 12 UI can show a staleness indicator.

### Claude's Discretion
- Internal naming of the DayDigest status enum values (e.g., `DayPending`, `DayActive`, `DayComplete`)
- Whether to add a `GetDayDigests()` method or expose the slice through `GetStatus()` with a separate field
- Handler function naming for the new endpoint (`handlePoCDays`, `handleDays`, etc.)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope
- `.planning/ROADMAP.md` §Phase 11 — requirements TRACK-01..04, CAMP-01
- `.planning/REQUIREMENTS.md` §Daily Tracking, §Campaign Execution — detailed requirement descriptions

### Primary files to modify
- `internal/engine/engine.go` — DayDigest struct, pre-population in `runPoC()`, heartbeat updates, delay_after application
- `internal/server/server.go` — new `/api/poc/days` route registration and handler

### Established patterns
- `.planning/codebase/CONVENTIONS.md` — naming conventions, error handling, code style
- `.planning/codebase/ARCHITECTURE.md` §Engine as a state machine — phase transitions, mutex usage
- `.planning/phases/10-poc-engine-fixes-clock-injection/10-CONTEXT.md` — Clock interface decisions (D-01..D-04)

### Existing code to reference
- `internal/engine/engine.go:69-89` — Status struct (existing PoC fields: PoCDay, PoCTotalDays, PoCPhase)
- `internal/engine/engine.go:92-106` — Engine struct (where DayDigest slice goes)
- `internal/engine/engine.go:675-683` — waitOrStop() (reuse for delay_after)
- `internal/playbooks/types.go:65-69` — CampaignStep struct with DelayAfter field
- `internal/server/server.go:78-98` — existing API route registrations (pattern for new endpoint)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `waitOrStop(d)` — already handles interruptible clock-based waits with stopCh. Reuse directly for delay_after.
- `CampaignStep.DelayAfter int` — field already exists in playbooks/types.go, just unused in engine. No schema changes needed.
- `globalDay` counter — already increments monotonically; DayDigest can index by `globalDay-1` into the slice.
- `e.clock.Now()` — use for all heartbeat timestamps and DayDigest start/end times.

### Established Patterns
- `sync.RWMutex` guards all Status writes — extend to DayDigest slice writes
- `handleStatus` in server.go — pattern for new `handlePoCDays`: acquire read lock, marshal to JSON, return
- `authMiddleware` wrapping — all data endpoints use this pattern
- `engine.GetStatus()` — pattern for a `GetDayDigests()` getter that acquires RLock

### Integration Points
- Phase 12 UI will poll `/api/poc/days` to render digest panel and calendar
- Phase 13 tests will assert DayDigest lifecycle transitions (pending -> active -> complete)
- `e.status.PoCDay` and `globalDay` already track position; DayDigest adds per-day detail

</code_context>

<specifics>
## Specific Ideas

- Pre-population loop: `for i := 0; i < totalDays; i++ { digests[i] = DayDigest{Day: i+1, Phase: phaseForDay(i), Status: DayPending} }`
- Phase label for each day derived from the Phase1/Gap/Phase2 day ranges already calculated in runPoC()
- DelayAfter of 0 means no delay — skip the waitOrStop call entirely

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 11-daily-tracking-backend-campaign-delay*
*Context gathered: 2026-04-09*
