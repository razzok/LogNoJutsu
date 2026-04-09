# Phase 11: Daily Tracking Backend & Campaign Delay - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-09
**Phase:** 11-daily-tracking-backend-campaign-delay
**Areas discussed:** DayDigest struct shape, API design, Campaign delay_after behavior, Heartbeat mechanism

---

## DayDigest Struct Shape

| Option | Description | Selected |
|--------|-------------|----------|
| Counts only | Day number, phase, status, technique count, pass/fail counts, timestamps | ✓ |
| Full technique list | Include array of technique IDs + results per day | |
| Hybrid | Counts + technique IDs without full results | |

**User's choice:** Counts only
**Notes:** Lightweight — Phase 12 UI just needs counts for badges and summaries. Full technique data already in engine.status.Results.

| Option | Description | Selected |
|--------|-------------|----------|
| At runPoC() start | Create all entries as 'pending' before first loop | ✓ |
| Lazily per section | Create entries at each section start | |

**User's choice:** At runPoC() start
**Notes:** Full schedule visible from first poll.

| Option | Description | Selected |
|--------|-------------|----------|
| Slice on Engine struct | []DayDigest field, guarded by RWMutex | ✓ |
| Embedded in Status struct | Serializes with /api/status automatically | |

**User's choice:** Slice on Engine struct
**Notes:** Consistent with Status pattern. Avoids bloating /api/status JSON payload.

---

## API Design (/api/poc/days)

| Option | Description | Selected |
|--------|-------------|----------|
| Full array only | GET /api/poc/days returns all entries | ✓ |
| Both: array + single | Array endpoint + per-day /api/poc/days/{n} | |

**User's choice:** Full array only
**Notes:** Max ~30 entries for any PoC run — small payload.

| Option | Description | Selected |
|--------|-------------|----------|
| Behind authMiddleware | Consistent with all other data endpoints | ✓ |
| Public (no auth) | Like /api/info | |

**User's choice:** Behind authMiddleware

| Option | Description | Selected |
|--------|-------------|----------|
| Empty array [] | HTTP 200 with [] when idle | ✓ |
| 404 or error | Return error when not in PoC mode | |

**User's choice:** Empty array []
**Notes:** Phase 12 UI handles empty state display.

---

## Campaign delay_after Behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Seconds | DelayAfter int treated as seconds | ✓ |
| Minutes | Treat as minutes | |

**User's choice:** Seconds

| Option | Description | Selected |
|--------|-------------|----------|
| e.clock.After | Injectable clock, fakeClock skips instantly | ✓ |
| Real time.After only | Always real time even in tests | |

**User's choice:** e.clock.After

| Option | Description | Selected |
|--------|-------------|----------|
| waitOrStop (interruptible) | Select on clock.After AND stopCh | ✓ |
| Non-interruptible sleep | Block for full duration | |

**User's choice:** waitOrStop (interruptible)
**Notes:** Consistent pattern with all other waits in runPoC().

---

## Heartbeat Mechanism (TRACK-04)

| Option | Description | Selected |
|--------|-------------|----------|
| At key events only | Update at day start, technique execution, day completion | ✓ |
| Periodic timer | Background goroutine updates every N seconds | |
| On every status poll | Update when /api/poc/days is called | |

**User's choice:** At key events only
**Notes:** No background timer complexity. Events happen frequently enough to prove liveness.

| Option | Description | Selected |
|--------|-------------|----------|
| Engine was alive during execution window | Staleness indicates potential issue | ✓ |
| Engine is currently active | More real-time | |

**User's choice:** Engine was alive during execution window
**Notes:** Engine sleeps for hours between daily runs — "currently active" would be misleading during wait periods.

---

## Claude's Discretion

- DayDigest status enum naming
- Getter method design (GetDayDigests vs extending GetStatus)
- Handler function naming for new endpoint

## Deferred Ideas

None — discussion stayed within phase scope.
