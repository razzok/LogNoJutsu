# Pitfalls Research

**Domain:** Adding daily execution tracking and timeline visualization to a long-running Go simulation engine with in-memory state
**Researched:** 2026-04-08
**Confidence:** HIGH — all findings derived directly from LogNoJutsu codebase inspection

---

## Critical Pitfalls

### Pitfall 1: Flat Results Slice Makes Per-Day Grouping Retroactively Impossible

**What goes wrong:**
`engine.Status.Results` is a single `[]playbooks.ExecutionResult` slice. Every technique executed across all PoC days and all phases appends to the same slice. There is no day index, no phase tag, and no day-boundary marker in `ExecutionResult`. If you build the daily digest panel by filtering `Results` at render time you either need to store day numbers on every result going forward, or try to reconstruct days from timestamps — which fails when techniques on different days run at the same configured hour.

**Why it happens:**
The flat slice was designed for single-run simulations. When PoC mode was added it reused the same accumulator instead of introducing a per-day structure. Because it "just worked" for storing results, no day discriminator was added to `ExecutionResult`.

**How to avoid:**
Before writing any UI code, add `PoCDay int` and `PoCPhase string` fields to `playbooks.ExecutionResult`. Populate them inside `runTechnique()` when `cfg.PoCMode` is true — the engine already sets `status.PoCDay` and `status.PoCPhase` under the lock, so reading those values in `runTechnique()` before appending the result is safe. Design the grouping key into the data layer, not the rendering layer.

**Warning signs:**
- Daily digest groups by timestamp range and breaks when two PoC days contain techniques at the same time of day
- UI shows 0 results for a day because the group boundary is off by one
- Retroactive backfilling is required after Phase 2 adds results with no day marker

**Phase to address:** Data structure phase — must precede any digest or timeline UI work

---

### Pitfall 2: PoCDay Counter Is Stale During Gap and Phase 2

**What goes wrong:**
In `runPoC()`, `status.PoCDay` is set only inside the Phase 1 loop (`e.status.PoCDay = day`). During the Gap and Phase 2 loops, `PoCDay` is never updated — it retains the last Phase 1 day value for the entire gap and all of Phase 2. The timeline calendar will show the wrong day number throughout a multi-week engagement. This is confirmed by reading `engine.go` lines 378–392 (Gap loop) and 399–430 (Phase 2 loop): neither sets `PoCDay`.

**Why it happens:**
The day counter was introduced alongside Phase 1 but was not extended to the other loops. No tests covered the full Phase1 → Gap → Phase2 lifecycle, so the bug was not caught before shipping.

**How to avoid:**
Introduce a single `globalDay` variable in `runPoC()` that increments on every day loop iteration across all three sub-phases. Set `status.PoCDay = globalDay` in all three loops, not only Phase 1. A consultant should see "Day 8 of 14" whether they are in Phase 1, Gap, or Phase 2.

**Warning signs:**
- `/api/status` returns `poc_day: 5` (the last Phase 1 day) while `poc_phase: "gap"` for the entire gap duration
- Timeline calendar highlights day 5 throughout the gap and all of Phase 2
- A test for the day counter that simulates all three sub-phases would catch this immediately

**Phase to address:** Bug fix phase — earliest possible, before any timeline UI work

---

### Pitfall 3: German Strings in CurrentStep Couple the Timeline UI to Localized Content

**What goes wrong:**
`status.CurrentStep` contains German strings during PoC execution (confirmed in `engine.go`):
- Phase 1: `"PoC Phase 1 — Tag %d/%d — warte bis %02d:00 Uhr"`
- Gap: `"PoC Pause — Tag %d/%d (keine Aktionen)"`
- Phase 2: `"PoC Phase 2 — Tag %d/%d — warte bis %02d:00 Uhr"`

Any timeline or digest UI built before these are translated will either display German text directly or require brittle string-matching against German tokens. If UI is built against English strings and the translation is done after, no coupling problem exists — but the order matters.

**Why it happens:**
The v1.1 English translation pass covered `index.html` but missed `engine.go` PoC step labels. These strings are in Go source, not the SPA, so a grep over HTML files would not catch them.

**How to avoid:**
Fix the three German format strings in `engine.go` before building the timeline UI. This is a 3-line change. Treat it as a prerequisite to UI work, not a parallel task.

**Warning signs:**
- Timeline UI has `if (step.includes("warte bis"))` conditional logic
- Day cards display "Tag 3/5" in German in an otherwise English UI
- Grep for "warte", "Tag %d", "keine Aktionen" in Go sources finds matches after the translation milestone

**Phase to address:** Bug fix phase — string fix must precede timeline UI phase

---

### Pitfall 4: waitOrStop Sleeps for Up to 24 Hours — Scheduling Logic Is Untestable

**What goes wrong:**
`waitOrStop(nextOccurrenceOfHour(hour))` blocks the engine goroutine for up to 24 hours between daily executions. There is no way to test `runPoC()` scheduling logic without either waiting real time or making the clock injectable. The existing test suite has zero tests for `runPoC()`. The stale PoCDay counter bug shipped precisely because no test simulated a full Phase1 → Gap → Phase2 cycle.

**Why it happens:**
`time.After` and `time.Now` are used directly throughout `runPoC()`. The `RunnerFunc` injection pattern (`SetRunner`) was added for technique execution testability but no equivalent abstraction was applied to time and scheduling.

**How to avoid:**
Introduce a `clockFn func() time.Time` and `nextDayFn func(hour int) time.Duration` into `Engine` (mirrors the existing `RunnerFunc` / `QueryFn` injection pattern). In tests, inject a fake clock that returns instantly. This is the only way to write a test that simulates "Phase1 Day 1 runs, Phase1 Day 2 runs, Gap Day 1 waits, Phase2 Day 1 runs" without the test taking days.

**Warning signs:**
- No test covers the Gap-between-phases transition
- No test asserts `status.PoCDay` is correct across all three sub-phases
- `runPoC` is the only substantial engine function with zero test coverage

**Phase to address:** Test coverage phase — clock injection is a prerequisite to any scheduling test

---

### Pitfall 5: In-Memory Log Growth — Full Slice Returned on Every Poll

**What goes wrong:**
`simlog.GetEntries()` copies the entire `[]Entry` slice on every call. The UI polls `/api/logs` every 2–3 seconds. A PoC run with 5 Phase 2 days × 10 techniques × ~20 log entries per technique generates ~1000 entries for Phase 2 alone. Phase 1, verification events, and PoC wait messages push the total to several thousand entries. Every poll copies and serializes the full slice over the loopback socket.

The UI's `if (entries.length === lastLogCount) return` check skips re-rendering but does not skip the network round-trip or the server-side copy. On a resource-constrained Windows machine (which SIEM validation targets commonly are), this degrades noticeably after day 3+. The existing `CONCERNS.md` documents this risk explicitly.

**Why it happens:**
`/api/logs` was designed for short single-day runs. No offset mechanism was added when PoC mode was introduced.

**How to avoid:**
Add `?offset=N` query parameter support to `handleLogs`. The server returns only `entries[N:]` and the UI passes `lastLogCount` as the offset. This is a one-pass change: server adds offset parsing, UI appends `?offset=${lastLogCount}` to the fetch URL. Also cap the in-memory slice at ~2000 entries to bound memory regardless of run duration.

**Warning signs:**
- `/api/logs` response payload exceeds 100 KB after a multi-day run
- Go process memory grows continuously over a week-long run
- Poll latency visibly increases as the simulation ages

**Phase to address:** Performance fix phase — log pagination

---

### Pitfall 6: GetStatus Shallow Copy Breaks with Nested Per-Day Slice

**What goes wrong:**
`GetStatus()` returns `Status` by value. `Status.Results` is a slice, so the returned copy shares the underlying array with the engine's internal slice. This is currently safe because callers only read results and append semantics create a new backing array on growth. But if you add `DayRecords []DayRecord` where each `DayRecord` contains its own `[]ExecutionResult`, you now have a nested slice. Any structural change to an inner slice after `GetStatus()` returns can corrupt the caller's view.

This is a latent bug that activates the moment per-day records are introduced and the UI polls while the engine is appending results.

**Why it happens:**
Go's by-value struct copy is shallow. Slices of slices are not deep-copied. The pattern looks safe for the flat current `Status` struct but breaks with nested collections.

**How to avoid:**
When `DayRecord` is introduced, either (a) deep-copy the `DayRecords` slice in `GetStatus()` using the same `copy()` pattern already in `simlog.GetEntries()`, or (b) design `DayRecord` as append-only (never mutate existing entries, only append new ones) so shallow copies remain valid. Whichever approach, add an explicit comment to `GetStatus()` documenting the copy depth guarantee.

**Warning signs:**
- Race detector fires on `status.DayRecords[n].Techniques` when a polling goroutine and `runTechnique` run concurrently
- Timeline UI shows intermittently missing or duplicated technique entries on days where techniques were still running during the poll

**Phase to address:** Data structure phase — when DayRecord is introduced, not as an afterthought

---

### Pitfall 7: Global simlog Session Is Replaced on Restart — Mid-Engagement Log History Lost

**What goes wrong:**
`simlog.Start()` replaces the global `current *Logger` pointer under `globalMu`. Any in-memory log entries from the previous session are discarded. If a consultant stops and restarts a PoC run mid-engagement (e.g. to adjust timing), all in-memory log entries from elapsed days vanish from the log viewer — even though the underlying `.log` file on disk retains them. The UI's `lastLogCount` cursor is also reset to 0 on the next `Start()`, so the incremental rendering logic re-fetches from the beginning of the new (empty) session rather than from the old offset.

**Why it happens:**
Single-simulation design. `simlog` was built for one session at a time; the logger swap was acceptable when each run was hours long. Multi-day runs accumulate logs across days and any restart resets the viewer.

**How to avoid:**
For PoC mode specifically: when restarting during an active PoC run, `Start()` should append to the existing session's logger rather than creating a new one. The simplest implementation: pass the existing `*Logger` pointer to `Start()` optionally, or expose a `Continue(campaignID string)` variant that skips the pointer swap. The file log can still be shared (the file handle is already open).

**Warning signs:**
- Log viewer shows an empty or nearly-empty state after a PoC stop/restart
- Day 3 log entries missing even though Day 3 ran successfully before a stop
- `lastLogCount` in the UI resets to 0 after restart, causing `entries.length === lastLogCount` to short-circuit and render nothing

**Phase to address:** Bug fix phase — simlog continuity for PoC restarts

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Flat `Results` slice for all PoC days | No data model change | Per-day grouping requires timestamp math that breaks at same-hour boundaries | Never — add day discriminator before digest UI ships |
| Global `simlog.current` pointer | Simple API, no dependency injection | Restart resets log history; concurrent sessions impossible | Only if single-session is a permanent constraint (it is for v1.2) |
| `time.After` directly in `runPoC` | Simplest possible scheduler | Scheduling logic is untestable without real-time waits | Never for correctness-critical code; inject the clock |
| `status.PoCDay` set only in Phase 1 loop | Minimal diff at PoC introduction | UI shows wrong day throughout Gap and Phase 2 — this is the confirmed bug | Never — it is a bug, not a tradeoff |
| `GetEntries()` returns full slice | Simple server handler | O(n) copy + serialize on every 2-second poll; degrades on multi-week runs | Acceptable for runs shorter than 1 day; not for multi-week PoC |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| `Status.Results` returned by `GetStatus()` | Assuming the returned copy is independent of the engine's internal slice | The copy is shallow — slice headers are copied but the backing array is shared until a new append triggers reallocation. Never write to results returned by `GetStatus()`. |
| `simlog.GetEntries()` as an offset cursor | Treating `lastLogCount` as a durable cursor across `simlog.Start()` calls | The count resets to 0 when `Start()` creates a new session. The UI must detect a count decrease (entries < lastLogCount) and reset its own cursor. |
| `nextOccurrenceOfHour(hour)` across system sleep/hibernate | Assuming `time.After` fires at real-world scheduled time after a Windows sleep | `time.After` uses a Go runtime timer; on Windows, timer resolution is ~15ms but hibernation can extend the delay. PoC runs on client machines that hibernate overnight will fire late. No wakelock or retry logic currently exists. |
| `PoCTotalDays` in first status poll | Assuming the field is populated before the first poll | `PoCTotalDays` is set inside `runPoC()` after the goroutine starts. There is a brief window where it is 0. The UI must treat 0 as "not yet known" rather than "zero days total". |
| `poc_phase` values in `Status` | Treating `"phase1"`, `"gap"`, `"phase2"` as an enum | These are plain strings set by assignment with no type constraint. Any typo in a new code path silently produces a mismatch. Define them as typed string constants. |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Full log copy on every poll | `/api/logs` latency increases; process memory rises ~1MB/day on verbose runs | Add `?offset=N` to `handleLogs`; cap in-memory slice at 2000 entries | Day 3+ on a verbose Phase 2 run with 10+ techniques |
| Full `Status.Results` in every `/api/status` response | Status payload grows proportionally to total techniques × days | Add a separate `/api/poc-summary` endpoint for the daily digest, keeping `/api/status` lightweight | Phase 2 with 5 days × 10 techniques = 50 results serialized on every 3-second poll |
| `GetStatus()` called inside the engine goroutine | Re-entrant read lock panic if called from code that already holds `e.mu.RLock` | Never call `GetStatus()` from within `runPoC()` or `runTechnique()`. Use internal fields directly inside the goroutine. | Any convenience call to `GetStatus()` added inside `runPoC()` during refactoring |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| New digest/timeline API routes registered without `authMiddleware` | Daily digest exposes technique execution timing and pass/fail to unauthenticated callers if auth is enabled | Register every new route through `s.authMiddleware(...)`. Make this a mandatory code-review checklist item for new route additions. |
| Writing partial day results to the HTML report file on day boundaries | Partial data looks like a completed report to the consultant; may be used for an interim SIEM presentation | Only invoke `reporter.SaveResults()` on `finish()` or `abort()`, never at day boundaries, unless interim reports are an explicit feature with clear "IN PROGRESS" labeling |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Timeline shows day numbers relative to the current phase, not the entire engagement | Consultant sees "Day 1/5" when Phase 2 starts and thinks the engagement restarted | Show global day number ("Day 8 of 14") throughout; show the phase label separately |
| Daily digest panel shows zero results for today while techniques are in progress | Consultant thinks today failed, stops the simulation prematurely | Populate today's digest entry with in-progress results before the day completes — show "running" state for the current day |
| `NextScheduledRun` displayed as raw RFC3339 string | "2026-04-09T09:00:00+02:00" is unreadable in a business context | Format as "Tomorrow at 09:00" or "In 14h 22m" in the UI; keep RFC3339 in the API for programmatic consumers |
| Future day cells shown as "scheduled" after the simulation is stopped | Consultant restarts a stopped run and sees stale future-day entries from the previous run | Future day cells must derive from live `PoCTotalDays` and current `PoCDay` on every status poll — never cache day state in DOM element data |

---

## "Looks Done But Isn't" Checklist

- [ ] **PoCDay counter fix:** Verify `status.PoCDay` increments monotonically through Phase1 → Gap → Phase2 in a test that exercises all three sub-phases. Testing Phase 1 alone does not validate the fix.
- [ ] **Daily digest panel:** Verify Day 1 results are still visible when viewing on Day 3 — re-rendering the timeline must not discard earlier completed days.
- [ ] **Log viewer after restart:** Stop and restart a PoC run mid-engagement; verify earlier log entries remain visible in the log viewer.
- [ ] **Offset-based log polling:** Verify the UI's `lastLogCount` cursor resets correctly when `simlog.Start()` creates a new session (count decreases signals a reset).
- [ ] **Timeline edge case (GapDays=0):** Verify the calendar renders correctly with no gap section — must not panic or render an empty gap row.
- [ ] **Campaign delay_after in Phase 2:** After the `getTechniquesForCampaign()` fix, verify it actually sleeps between steps. Test by injecting a fake clock and asserting elapsed > 0.
- [ ] **German strings gone:** After the translation fix, grep Go sources for "warte bis", "Tag %d", "keine Aktionen" — zero matches expected.
- [ ] **New API routes auth-gated:** For every new route added, write a test that verifies a request without credentials returns 401 when a password is configured.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Flat Results with no day discriminator — UI built on top | HIGH | Add `PoCDay` to `ExecutionResult`; attempt to back-fill from `PhaseStartTimes` timestamps; update all rendering logic; retest all digest views |
| German strings in timeline UI — built before translation | MEDIUM | Replace German string matches with English constants; update UI conditional logic; retest all day-boundary detection paths |
| waitOrStop untestable — scheduling bug discovered in engagement | HIGH | Refactor clock injection into Engine; rewrite runPoC to accept the injected clock; existing tests unaffected but new test suite requires significant setup |
| In-memory log overflow mid-engagement | LOW | Add entry cap (last 2000) to `simlog.Logger.write()`; file log is unaffected; UI log viewer loses oldest entries but continues correctly |
| PoCDay stuck at Phase1 value — discovered by consultant during a live run | LOW | Hot-fix: add `globalDay` counter to all three loops in `runPoC()`; redeploy binary; consultant restarts run |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Flat Results with no day discriminator | Data structure phase — add `PoCDay`/`PoCPhase` to `ExecutionResult` before any UI work | Test: run simulated PoC with fakeRunner; assert `Results[n].PoCDay` matches the expected global day counter |
| PoCDay stale during Gap/Phase2 | Bug fix phase — monotonic global day counter in `runPoC` | Test: simulate 3+1+2 day PoC; assert `PoCDay` == 1,2,3,4,5,6 sequentially, never resets |
| German CurrentStep strings | Bug fix phase — translate `engine.go` PoC step labels before any timeline UI | Grep: zero matches for German tokens in all `.go` files |
| simlog history lost on restart | Bug fix phase — PoC restart appends to existing logger | Test: call `Start()`, log entries, call `Start()` again in PoC-continue mode, call `GetEntries()` — earlier entries still present |
| In-memory log growth | Performance fix phase — offset param + entry cap | Test: generate 5000 log entries, verify `GetEntries()` returns ≤ 2000; verify `/api/logs?offset=1000` returns 1000 entries |
| waitOrStop untestable | Test coverage phase — clock/duration injection into Engine | Test: simulate full 3-phase PoC in < 100ms with fake clock; assert all phase transitions, day counters, and technique execution order |
| Nested DayRecord shallow copy | Data structure phase — document and enforce copy depth when DayRecord introduced | Race detector test: concurrent `GetStatus()` + `runTechnique()` calls; zero races with `-race` |
| New routes missing auth middleware | Every new route phase | Test: each new endpoint returns 401 without credentials when password is configured |

---

## Sources

- `internal/engine/engine.go` — `runPoC()`, `Status` struct, `runTechnique()`, `waitOrStop()`, German string literals at lines 351, 389, 411
- `internal/simlog/simlog.go` — global logger, `GetEntries()`, `Start()` session replacement at line 69
- `internal/server/server.go` — `handleLogs()`, `handleStatus()`, auth middleware registration pattern
- `internal/server/static/index.html` — `setInterval(pollStatus, 3000)`, `lastLogCount` cursor at line 1121, log viewer rendering
- `.planning/codebase/CONCERNS.md` — documented in-memory growth risk, simlog dual-lock fragility, engine singleton concern
- Direct codebase inspection, 2026-04-08

---
*Pitfalls research for: LogNoJutsu v1.2 PoC Mode Fix & Overhaul*
*Researched: 2026-04-08*
