# Architecture Research

**Domain:** Go single-binary SIEM validation tool — v1.2 PoC Mode Fix & Overhaul
**Researched:** 2026-04-08
**Confidence:** HIGH (derived entirely from direct codebase inspection)

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                       cmd/lognojutsu                             │
│  main.go — entry point, server.Start()                           │
├─────────────────────────────────────────────────────────────────┤
│                      internal/server                             │
│  Server struct — HTTP routing, all handlers                      │
│  GET /api/status   →  engine.GetStatus()  (polls every 2-3s)    │
│  GET /api/logs     →  simlog.GetEntries()                        │
│  GET /api/poc/days →  engine.GetDayDigests()  [NEW]              │
├─────────────────────────────────────────────────────────────────┤
│                      internal/engine                             │
│  Engine struct — mutex-protected Status                          │
│  run()  →  runPoC()  →  3-loop (Phase1 / Gap / Phase2)           │
│  runTechnique()  →  results appended to Status.Results[]         │
│  NEW: DayDigests []DayDigest  (per-day tracking)                 │
├──────────────────────────────────┬──────────────────────────────┤
│        internal/simlog           │    internal/playbooks        │
│  global Logger, in-memory        │  Technique, ExecutionResult  │
│  []Entry + file write            │  Campaign, CampaignStep      │
│  TypePhase entries for separator │  VerificationStatus          │
└──────────────────────────────────┴──────────────────────────────┘
```

## Current State: What Exists and What Is Broken

### What exists in runPoC() today

- Three nested loops: Phase1 days, Gap days, Phase2 days
- Each loop iteration calls `nextOccurrenceOfHour()`, waits, then calls `runTechnique()` for each selected technique
- `e.status.PoCDay` is set inside the Phase1 loop but **never updated** in the Gap or Phase2 loops — it freezes at the last Phase1 day number
- `simlog.Phase()` is called once at the top of Phase1 and Phase2 (`e.setPhase()`) but not called at the start of each individual day — no per-day separators in the log
- `CurrentStep` strings in Phase1 and Phase2 contain German text ("warte bis", "Tag")
- All technique results go into the flat `Status.Results[]` array with no day attribution
- `Status` has no data structure tracking which techniques executed on which day

### What the UI does today

- Polls `GET /api/status` every 2-3 seconds
- Reads `s.poc_day`, `s.poc_total_days`, `s.poc_phase`, `s.next_scheduled_run` from the status payload
- Renders a four-box summary panel (PoC Day / Total Days / Countdown / Current Phase)
- Has no concept of a "daily digest" — there is no per-day technique list
- Has no timeline calendar — `pocSchedulePreview` shows a static pre-computed date range based on form inputs, not live execution state

---

## New Data Structures Required

### DayDigest (new, in engine package)

This is the central new type. It records everything that happened on a single simulated day.

```go
// DayDigest records execution results for a single PoC simulation day.
type DayDigest struct {
    DayNumber     int                         `json:"day_number"`
    PoCPhase      string                      `json:"poc_phase"`   // "phase1", "gap", "phase2"
    ScheduledFor  string                      `json:"scheduled_for"`  // RFC3339 — when it was scheduled
    ExecutedAt    string                      `json:"executed_at,omitempty"` // RFC3339 — when execution started
    CompletedAt   string                      `json:"completed_at,omitempty"` // RFC3339 — when last technique finished
    Status        DayStatus                   `json:"status"`      // "pending", "waiting", "running", "done", "skipped"
    Results       []playbooks.ExecutionResult `json:"results"`     // techniques that ran this day (empty for gap days)
    TechCount     int                         `json:"tech_count"`  // convenience: len(Results)
    SuccessCount  int                         `json:"success_count"`
    FailCount     int                         `json:"fail_count"`
}

type DayStatus string

const (
    DayPending DayStatus = "pending"  // scheduled, not yet started waiting
    DayWaiting DayStatus = "waiting"  // inside nextOccurrenceOfHour wait
    DayRunning DayStatus = "running"  // techniques executing now
    DayDone    DayStatus = "done"     // day complete
    DaySkipped DayStatus = "skipped"  // gap day — no actions, but logged
)
```

**Rationale for embedding `[]playbooks.ExecutionResult` in each DayDigest:** The results already exist in `Status.Results[]`. Rather than a secondary index (e.g., a `map[int][]int` of day → result indices), embedding copies per-day results directly in the digest. This makes the `/api/poc/days` response self-contained and avoids UI-side join logic. The flat `Status.Results[]` is preserved unchanged for backward compatibility with the existing results table.

### Engine struct additions

```go
type Engine struct {
    // ... existing fields unchanged ...

    // NEW: per-day tracking for PoC mode
    dayDigests []DayDigest  // protected by e.mu
}
```

### Status struct additions

Two new fields on the existing Status struct support the UI countdown and day tracking:

```go
type Status struct {
    // ... all existing fields unchanged ...

    // v1.2 additions
    PoCDayAbsolute int `json:"poc_day_absolute,omitempty"` // monotonic counter across all phases (1..totalDays)
}
```

`PoCDay` already exists but is broken (frozen after Phase1). The fix is to update it in all three loops, not add a second field. `PoCDayAbsolute` tracks position across the full multi-phase sequence (day 1 of Phase2 = day Phase1Days+GapDays+1 in absolute terms) — useful for the timeline calendar.

---

## Integration Points: New vs Modified

| Item | New / Modified | Location | What Changes |
|------|---------------|----------|--------------|
| `DayDigest` struct | NEW | `internal/engine/engine.go` | New type, ~25 lines |
| `DayStatus` type + constants | NEW | `internal/engine/engine.go` | New type, 5 lines |
| `Engine.dayDigests []DayDigest` | Modified | `internal/engine/engine.go` | Add field to existing struct |
| `Status.PoCDayAbsolute int` | NEW | `internal/engine/engine.go` | Add to existing Status struct |
| `Engine.GetDayDigests()` | NEW | `internal/engine/engine.go` | Public method returning `[]DayDigest` under RLock |
| `runPoC()` — Phase1 loop | Modified | `internal/engine/engine.go` | Fix PoCDay, add DayDigest lifecycle calls, fix CurrentStep language |
| `runPoC()` — Gap loop | Modified | `internal/engine/engine.go` | Add DayDigest (status=skipped), fix PoCDay/PoCDayAbsolute |
| `runPoC()` — Phase2 loop | Modified | `internal/engine/engine.go` | Fix PoCDay, add DayDigest lifecycle, add delay_after support |
| `simlog.PoCDay()` | NEW | `internal/simlog/simlog.go` | New function writing TypePoCDay entry with separator |
| `TypePoCDay` entry type | NEW | `internal/simlog/simlog.go` | New EntryType constant |
| `GET /api/poc/days` handler | NEW | `internal/server/server.go` | ~10 lines, calls `eng.GetDayDigests()` |
| Route registration | Modified | `internal/server/server.go` | One line in `registerRoutes` |
| `pocInfoPanel` in HTML | Modified | `internal/server/static/index.html` | Add daily digest table beneath existing four-box summary |
| Timeline calendar section | NEW | `internal/server/static/index.html` | New HTML section + JS render function |
| `updateStatus()` JS | Modified | `internal/server/static/index.html` | Poll `/api/poc/days` when `isPocRunning`, call `renderCalendar()` |

---

## Recommended Data Flow

### Per-day execution tracking

```
runPoC() Phase1 loop — day N starts:
    e.mu.Lock()
    digest := DayDigest{
        DayNumber:    N,
        PoCPhase:     "phase1",
        ScheduledFor: nextRun.Format(time.RFC3339),
        Status:       DayWaiting,
    }
    e.dayDigests = append(e.dayDigests, digest)
    e.status.PoCDay = N
    e.status.PoCDayAbsolute = N   // absolute across all phases
    e.mu.Unlock()

    waitOrStop(d)  // waits for nextOccurrenceOfHour

    e.mu.Lock()
    e.dayDigests[last].Status = DayRunning
    e.dayDigests[last].ExecutedAt = time.Now().Format(time.RFC3339)
    e.mu.Unlock()

    simlog.PoCDay(N, cfg.Phase1DurationDays, "phase1")  // new separator in log

    for each technique:
        e.runTechnique(t)
        result = e.status.Results[last]  // just appended by runTechnique
        e.mu.Lock()
        e.dayDigests[last].Results = append(e.dayDigests[last].Results, result)
        e.mu.Unlock()

    e.mu.Lock()
    e.dayDigests[last].Status = DayDone
    e.dayDigests[last].CompletedAt = time.Now().Format(time.RFC3339)
    e.dayDigests[last].TechCount = len(e.dayDigests[last].Results)
    e.dayDigests[last].SuccessCount = countSucceeded(e.dayDigests[last].Results)
    e.dayDigests[last].FailCount = e.dayDigests[last].TechCount - e.dayDigests[last].SuccessCount
    e.mu.Unlock()
```

Gap days follow the same pattern but set `Status = DaySkipped` immediately (no wait, no techniques). They still get a DayDigest entry so the timeline calendar can render them.

### UI polling flow

```
setInterval(async () => {
    const s = await api('/api/status')   // existing — every 2-3s
    updateStatus(s)                      // existing handler

    if (isPocRunning(s.phase)) {
        const days = await api('/api/poc/days')   // NEW endpoint
        renderDigestTable(days)          // NEW: per-day summary rows
        renderCalendar(days, s)          // NEW: day-by-day timeline
    }
}, 2500)
```

The two fetches in the interval are fine for a local-only tool with no network latency. They do not need to be combined into one endpoint — keeping `/api/status` unchanged avoids breaking the existing results table.

### GET /api/poc/days response shape

```json
[
  {
    "day_number": 1,
    "poc_phase": "phase1",
    "scheduled_for": "2026-04-09T09:00:00+02:00",
    "executed_at": "2026-04-09T09:00:01+02:00",
    "completed_at": "2026-04-09T09:03:47+02:00",
    "status": "done",
    "results": [ { "technique_id": "T1059", ... } ],
    "tech_count": 3,
    "success_count": 3,
    "fail_count": 0
  },
  {
    "day_number": 2,
    "poc_phase": "phase1",
    "scheduled_for": "2026-04-10T09:00:00+02:00",
    "status": "waiting",
    "results": [],
    "tech_count": 0,
    "success_count": 0,
    "fail_count": 0
  }
]
```

---

## Architectural Patterns

### Pattern 1: Append-only digest slice with mutex protection

**What:** Append new `DayDigest` entries to `Engine.dayDigests` as each day begins. Mutate the last element in-place as the day progresses (status transitions, results accumulation). Never remove entries.

**When to use:** This is the right approach here because:
- The slice is append-only over the lifetime of a simulation run (bounded by totalDays, max ~60 entries)
- The last element is the only one that needs mutation
- Finding the last element is O(1) using `len(e.dayDigests)-1`
- `GetDayDigests()` makes a full copy under RLock, so the caller cannot race with ongoing appends

**Trade-offs:** Copying the full slice on every poll (every 2.5s) is negligible at 60 entries max. The copy protects the UI from seeing a partially-written entry.

**Code pattern:**

```go
// Start a day
e.mu.Lock()
e.dayDigests = append(e.dayDigests, DayDigest{
    DayNumber: day, PoCPhase: "phase1", Status: DayWaiting,
    ScheduledFor: nextRun.Format(time.RFC3339),
})
e.mu.Unlock()

// Mutate the last element
e.mu.Lock()
last := len(e.dayDigests) - 1
e.dayDigests[last].Status = DayRunning
e.mu.Unlock()
```

### Pattern 2: Result duplication between flat Results[] and per-day digest

**What:** After `runTechnique()` appends to `Status.Results`, copy the result into the current day's digest.

**When to use:** This is necessary because `Status.Results` is the existing flat array consumed by the results table, and it must remain unchanged. The digest needs its own copy of the results for the `/api/poc/days` response.

**Trade-offs:** Memory doubles for PoC mode results. At 60 techniques/day × 30 days = 1800 results, each result is a small struct (~500 bytes JSON), total ~900 KB — acceptable for a local tool.

**Do not:** Attempt to reference results by index into `Status.Results`. Index-based references break if the slice is ever reallocated or if non-PoC runs interleave (they cannot with the current engine design, but it creates fragile coupling).

### Pattern 3: Pre-allocate all future DayDigest entries at runPoC() start

**What:** At the start of `runPoC()`, populate `e.dayDigests` with one `DayDigest{Status: DayPending}` entry per scheduled day, using pre-computed `ScheduledFor` timestamps.

**When to use:** This allows the timeline calendar to render the full schedule immediately — including future days — from the first poll. The UI does not need to infer the schedule from the Config; it reads actual planned dates from the digest.

**How:** Compute each day's scheduled time at the start of `runPoC()` using the same `nextOccurrenceOfHour` logic but iterated forward, then mutate entries from `DayPending` → `DayWaiting` → `DayRunning` → `DayDone` as execution proceeds.

**Trade-offs:** Pre-computation assumes the clock moves forward at 1 day/day (true for real runs). For accelerated testing (fake time injection), this needs the `nextOccurrenceOfHour` function to be injectable — which is already the pattern the engine uses for `RunnerFunc`.

---

## Anti-Patterns

### Anti-Pattern 1: Adding day number as a field to ExecutionResult

**What people do:** Add a `PoCDay int` field to `playbooks.ExecutionResult` so every result knows which day it belongs to.

**Why it's wrong:** `ExecutionResult` is used in both PoC mode and Quick mode. Adding a PoC-specific field to a shared type pollutes the common struct. It also couples the playbooks package to engine scheduling concepts. The Quick mode HTML report already consumes `ExecutionResult` directly — a PoC-specific field appears in reports as noise.

**Do this instead:** Keep `DayDigest.Results []playbooks.ExecutionResult` as the attribution layer. The existing `ExecutionResult` type stays clean and the coupling lives in the engine package where scheduling is already managed.

### Anti-Pattern 2: Storing day digests in simlog

**What people do:** Encode per-day summaries as structured `TypePoCDay` log entries in the simlog `[]Entry` and reconstruct digests from the log on the read path.

**Why it's wrong:** The log is write-only structured text, not a queryable store. Reconstructing digests requires a full scan of all entries on every `/api/poc/days` poll. It also couples the simlog package to engine scheduling structures.

**Do this instead:** Keep `Engine.dayDigests` as the authoritative structured store. Add `simlog.PoCDay()` only for human-readable log file output (separators + summary line) — not as the data source for the digest API.

### Anti-Pattern 3: Separate `/api/poc/status` endpoint combining status + digests

**What people do:** Replace `/api/status` with a new combined endpoint that returns both the existing `Status` fields and the new `[]DayDigest` in one payload.

**Why it's wrong:** The existing `/api/status` contract is stable and consumed by the existing results table, phase badge, countdown, and stop button. Merging everything into one response requires the UI to handle a large combined struct. It also makes the test setup for `handleStatus` more complex.

**Do this instead:** Add a separate `/api/poc/days` endpoint. The UI fetches both when in PoC mode. The endpoints remain independently testable and `handleStatus` stays unchanged.

### Anti-Pattern 4: Computing DayDigest.ScheduledFor from Config in the UI

**What people do:** Have the timeline calendar JavaScript reconstruct the schedule from the Config fields (p1 days, p1 hour, gap days, etc.) as the existing `updatePoCSchedule()` function already does.

**Why it's wrong:** The UI-computed schedule diverges from reality when execution is delayed (e.g., machine sleeps, engine starts mid-day). The timeline would show projected dates that don't match when techniques actually ran.

**Do this instead:** Compute `ScheduledFor` in Go at `runPoC()` start and surface it through `DayDigest`. The UI reads actual planned times, not inferred ones. The existing `updatePoCSchedule()` pre-run preview remains (it's fine for before the simulation starts), but once running, the calendar renders from `/api/poc/days`.

---

## Bug Fixes That Must Precede Feature Work

These are pre-requisites. Building the digest on top of broken loop logic will produce wrong output.

### Bug 1: Stale PoCDay counter

**Root cause:** `e.status.PoCDay = day` is only set in the Phase1 loop (line 349). The Gap and Phase2 loops never update it.

**Fix:** In the Gap loop, set `PoCDay = Phase1DurationDays + gapDay`. In the Phase2 loop, set `PoCDay = Phase1DurationDays + GapDays + day`.

**Where:** `runPoC()`, Gap loop and Phase2 loop, immediately before the `waitOrStop` call.

### Bug 2: Missing simlog.Phase() separators per day

**Root cause:** `e.setPhase(PhasePoCPhase1)` is called once before the Phase1 loop — this writes one Phase separator for the entire Phase1 block. There are no per-day separators, so the log file is hard to read during long multi-day runs.

**Fix:** Add `simlog.PoCDay(day, totalDaysInPhase, phaseName)` at the start of each day loop iteration, after the wait returns. This writes a readable day header to the log.

**New simlog function:**

```go
// PoCDay logs the start of a simulation day in PoC mode.
func PoCDay(day, total int, phase string) {
    globalMu.Lock()
    defer globalMu.Unlock()
    if current == nil {
        return
    }
    separator := strings.Repeat("─", 60)
    current.write(TypePoCDay, "", separator, "")
    current.write(TypePoCDay, "", fmt.Sprintf("▶ POC DAY %d/%d — %s — %s",
        day, total, strings.ToUpper(phase), time.Now().Format("2006-01-02 15:04")), "")
    current.write(TypePoCDay, "", separator, "")
}
```

### Bug 3: German strings in CurrentStep

**Root cause:** `CurrentStep` is set to German strings in both the Phase1 and Phase2 loops (lines 351 and 411).

**Fix:** Replace with English equivalents. The strings must be in English because `CurrentStep` is rendered in the UI.

**Phase1:** `"PoC Phase 1 — Tag %d/%d — warte bis %02d:00 Uhr"` → `"PoC Phase 1 — Day %d/%d — waiting until %02d:00"`

**Phase2:** `"PoC Phase 2 — Tag %d/%d — warte bis %02d:00 Uhr"` → `"PoC Phase 2 — Day %d/%d — waiting until %02d:00"`

**Gap:** `"PoC Pause — Tag %d/%d (keine Aktionen)"` → `"PoC Gap — Day %d/%d (no actions)"`

---

## Build Order (Dependency-aware)

The work must be sequenced to avoid building digest features on top of buggy loops.

**Phase A — Bug fixes (no new data structures)**

1. Fix PoCDay counter in Gap and Phase2 loops
2. Fix German CurrentStep strings in all three loops
3. Add `simlog.PoCDay()` function and `TypePoCDay` constant
4. Call `simlog.PoCDay()` at start of each day iteration
5. Wire `delay_after` from CampaignStep in Phase2 (currently not applied)
6. Write `engine_test.go` tests for the scheduling logic (before adding digest)

**Phase B — DayDigest structure (no UI changes)**

7. Add `DayDigest`, `DayStatus` types to `engine.go`
8. Add `Engine.dayDigests` field
9. Add `Status.PoCDayAbsolute` field
10. Pre-populate pending digests at `runPoC()` start
11. Add lifecycle mutations in Phase1/Gap/Phase2 loops
12. Add `Engine.GetDayDigests()` public method
13. Add `GET /api/poc/days` handler and route

**Phase C — UI (depends on Phase B endpoint existing)**

14. Daily digest table in `pocInfoPanel` (renders from `/api/poc/days` data)
15. Timeline calendar section (renders from `/api/poc/days` data)
16. Update `updateStatus()` polling to fetch `/api/poc/days` when PoC is running

---

## Component Boundary Summary

| Component | Owns | Communicates With |
|-----------|------|-------------------|
| `engine.go` | `DayDigest` type, `Engine.dayDigests`, per-day lifecycle | `simlog` (write-only calls), `playbooks.ExecutionResult` (read) |
| `simlog.go` | Log file, in-memory `[]Entry`, `TypePoCDay` | None (package is write-only from engine's perspective) |
| `server.go` | HTTP routing, `/api/poc/days` handler | `engine.GetDayDigests()` (new), `engine.GetStatus()` (existing) |
| `index.html` | Digest table render, calendar render, status polling | `/api/status` (existing), `/api/poc/days` (new) |

---

## Sources

- Direct inspection: `internal/engine/engine.go` — full `runPoC()` function, `Status` struct, `Engine` struct
- Direct inspection: `internal/simlog/simlog.go` — full Logger, all write functions
- Direct inspection: `internal/server/server.go` — all routes, `handleStatus`, `handleLogs`
- Direct inspection: `internal/server/static/index.html` — PoC panel HTML, `updateStatus()`, `updatePoCSchedule()`
- Direct inspection: `internal/playbooks/types.go` — `ExecutionResult`, `Technique`, `CampaignStep`
- Direct inspection: `.planning/PROJECT.md` — v1.2 scope and known bugs

---
*Architecture research for: LogNoJutsu v1.2 PoC Mode Fix & Overhaul*
*Researched: 2026-04-08*
