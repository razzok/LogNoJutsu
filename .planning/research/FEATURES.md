# Feature Research

**Domain:** Multi-day simulation dashboard / daily execution feedback for SIEM validation tool
**Researched:** 2026-04-08
**Confidence:** HIGH (analysis from codebase, not external research)

---

## Context: What Already Exists vs What Is Being Added

This milestone is a targeted overhaul, not greenfield. Features below are evaluated against the
existing PoC mode implementation. The existing system has:

- `runPoC()` in `engine.go`: 3-phase loop (Phase1 Discovery → Gap → Phase2 Attack), hour-based scheduling via `nextOccurrenceOfHour()`
- `Status` struct PoC fields: `PoCDay`, `PoCTotalDays`, `PoCPhase`, `NextScheduledRun`
- Dashboard panel (`pocInfoPanel`): day counter, total days, countdown to next run, current phase description
- Schedule preview (`pocSchedulePreview`): pre-run static date range table (Phase 1 start/end, Gap, Phase 2 start/end)
- Known bugs: `PoCDay` only updated in Phase1 loop (stale during Gap and Phase2); German `CurrentStep` strings hardcoded; no `simlog.Phase()` calls in `runPoC()`
- Campaign YAML has `step.DelayAfter` but `getTechniquesForCampaign()` discards it

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features a consultant using PoC mode will assume exist. Their absence makes the tool feel broken
or untrustworthy during a multi-week client engagement.

| Feature | Why Expected | Complexity | Depends On |
|---------|--------------|------------|------------|
| Bug: stale PoCDay counter during Gap and Phase2 | Day counter shows "1" throughout Gap and all of Phase2. Consultant cannot tell which day of the engagement they are on. PoC mode appears stuck. | LOW | `engine.go` `runPoC()` — add `PoCDay` update to Gap and Phase2 loops |
| Bug: German CurrentStep strings | `"PoC Phase 1 — Tag %d/%d — warte bis %02d:00 Uhr"` — product was fully translated to English in v1.1; these strings survived. Dashboard shows German during every PoC run. | LOW | `engine.go` `runPoC()` — translate 3 format strings |
| Bug: missing simlog.Phase() calls in runPoC() | Phase transitions in PoC mode don't write the separator lines to the log that normal mode writes. The `.log` file has no visual phase boundaries, making it hard to audit or share with clients. | LOW | `engine.go` `runPoC()`, `simlog.Phase()` already exists |
| Daily digest panel: per-day execution summary | After each day completes, consultant must see which techniques ran, at what time, and whether they succeeded or failed. Without this, the only feedback is the day counter advancing. Over a 17-day engagement this is the primary progress signal. | MEDIUM | New `DailyDigest []DailyDigest` slice in `Status`; engine appends a record after each day's technique loop |
| Timeline calendar: visual schedule with completion state | During a 17-day engagement, a consultant presenting to a client needs to see at a glance which days are done, which is current, and which remain — not just a static pre-run date preview. | MEDIUM | Requires `DailyDigest` data from backend; JS renders day strip in index.html |

### Differentiators (Competitive Advantage)

Features that make PoC mode distinctly more trustworthy than a consultant manually tracking
execution in a spreadsheet.

| Feature | Value Proposition | Complexity | Depends On |
|---------|-------------------|------------|------------|
| Campaign `delay_after` support in Phase2 | Campaign YAML steps declare `delay_after` seconds for realistic timing between techniques. `getTechniquesForCampaign()` discards this. Honoring it makes Phase2 simulations more realistic, which matters for SIEM correlation rule timing windows. | LOW-MEDIUM | `engine.go` — `getTechniquesForCampaign()` must return steps (not just techniques), Phase2 loop must call `waitOrStop(step.DelayAfter)` |
| Test coverage for runPoC() scheduling logic | `runPoC()` has never had unit tests despite being the most complex scheduling path. The injectable `RunnerFunc` already exists. Tests validate day counting, phase transitions, and stop-signal handling without real time.Sleep. | MEDIUM | `engine_test.go`, `RunnerFunc` (already in engine.go) |
| Technique counts in schedule preview (pre-run) | Static preview shows dates only. Showing "3 techniques/day" for Phase1 and "N techniques (campaign)" for Phase2 makes the preview actionable before starting. | LOW | JS `updatePoCSchedule()` only — no backend change |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Pause/Resume PoC mid-engagement | Seems useful if machine reboots or consultant interrupts the run | Go's goroutine/channel model has no clean pause primitive. Stop cancels the run entirely and resets state. Resume-from-day-N requires persisting state to disk, adds a persistence layer, and creates failure modes around state corruption | Stop + restart; the timeline calendar shows which days completed; add start-from-day-N config field only if consultants consistently request it (v2+) |
| Real-time technique output in daily digest | Showing raw PowerShell output per technique in the daily summary | Output can be kilobytes per technique; rendering all of it in the digest panel makes the UI unusable. The log viewer (`/api/logs`) already handles this. | Link each digest entry to the log viewer; no raw output in digest |
| Persistent state across binary restarts | Consultant restarts lognojutsu.exe mid-engagement and wants it to continue | Requires a state file on disk, migration logic, and version-aware deserialization — scope explosion for marginal use case | The `.log` file (already written to disk) provides a full audit trail. Restart PoC mode from scratch. |
| Per-day HTML report | Separate HTML report for each day's execution | Reporter always generates a full-run report at finish. Per-day reports require calling `reporter.SaveResults()` mid-run with partial results, which changes its semantics. | The existing HTML report covers all executed techniques across all PoC days. |
| SIEM alert correlation in digest | Show whether SIEM fired an alert for each technique that day | Requires SIEM API access — explicitly out of scope in PROJECT.md; breaks standalone deployment model | Local event log verification (already per-technique in `ExecutionResult`) is the proxy. Show `VerificationStatus` in the digest. |

---

## Feature Dependencies

```
[Bug: stale PoCDay]
    └──must-fix-before──> [Daily Digest Panel]
                              └──provides-data-for──> [Timeline Calendar]

[Bug: German CurrentStep strings]
    └──standalone fix (no dependencies)

[Bug: missing simlog.Phase() calls]
    └──standalone fix (no dependencies)

[DailyDigest struct added to Status]
    └──required-by──> [Daily Digest Panel UI]
    └──required-by──> [Timeline Calendar UI]

[Campaign delay_after in Phase2]
    └──structural change: getTechniquesForCampaign() return type
    └──standalone — does not block other features

[runPoC() test coverage]
    └──best-added-after──> [Bug fixes] (tests should validate correct behavior, not bugs)
    └──uses──> [RunnerFunc injection] (already exists)
```

### Dependency Notes

- **Daily digest requires PoCDay bug fix first.** The digest records which day each bucket of results belongs to. A stale counter mislabels entries.
- **Timeline calendar requires daily digest.** The calendar needs completed-day data (techniques run, success count) to show anything beyond the static date preview that already exists.
- **delay_after is self-contained but structurally significant.** `getTechniquesForCampaign()` returns `[]*playbooks.Technique`, discarding `CampaignStep.DelayAfter`. The fix either changes the return type (affects callers) or adds a sibling helper returning `[]CampaignStep` for Phase2 use only.
- **Tests should come after bug fixes.** Writing tests that document currently-buggy behavior creates misleading test suite state.

---

## MVP Definition

### v1.2 Launch With (required for milestone)

- [ ] Bug fix: stale `PoCDay` counter during Gap and Phase2
- [ ] Bug fix: German `CurrentStep` strings in `runPoC()`
- [ ] Bug fix: missing `simlog.Phase()` calls in `runPoC()`
- [ ] `DailyDigest` data structure added to `Status` (Go backend)
- [ ] Engine appends `DailyDigest` entry after each completed day (Phase1, Gap, Phase2 loops)
- [ ] Daily digest panel in UI: scrollable list of completed days, technique names, success/fail indicators
- [ ] Timeline calendar: day strip showing completed (check), current (pulsing), gap (silent), future (dimmed)
- [ ] Test coverage for `runPoC()`: day sequencing, phase transitions, stop signal

### Add After Validation (v1.3 candidates)

- [ ] Campaign `delay_after` support in Phase2 — structural change; ships cleanly in its own phase
- [ ] Technique count labels in schedule preview (pre-run) — low effort, cosmetic improvement

### Future Consideration (v2+)

- [ ] Start-from-day-N config field — only if consultants consistently interrupt and restart engagements
- [ ] Per-day simlog markers with embedded day number — already partially covered by `simlog.Info()`

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Bug: stale PoCDay | HIGH — affects every PoC run, day counter is the primary status signal | LOW | P1 |
| Bug: German strings | HIGH — product regression from v1.1 English translation | LOW | P1 |
| Bug: missing Phase() calls | MEDIUM — log readability for client deliverables | LOW | P1 |
| DailyDigest backend struct | HIGH — unblocks digest panel and timeline calendar | MEDIUM | P1 |
| Daily digest panel (UI) | HIGH — only feedback mechanism during a 17-day engagement | MEDIUM | P1 |
| Timeline calendar (UI) | HIGH — visual orientation; consultant presenting to client | MEDIUM | P1 |
| runPoC() test coverage | HIGH — prevents regressions in scheduling logic | MEDIUM | P1 |
| delay_after in Phase2 | MEDIUM — simulation realism improvement | MEDIUM | P2 |
| Technique counts in schedule preview | LOW — cosmetic, pre-run only | LOW | P3 |

**Priority key:**
- P1: Required for v1.2 milestone
- P2: Should have, ship in v1.3
- P3: Nice to have, future consideration

---

## Implementation Notes

### DailyDigest struct design

The engine appends a record after each completed day. Suggested shape:

```go
// DailyDigest records the result of one PoC simulation day.
type DailyDigest struct {
    Day          int                `json:"day"`           // 1-based within the phase
    Phase        string             `json:"phase"`         // "phase1", "gap", "phase2"
    Date         string             `json:"date"`          // RFC3339 of execution time
    TechCount    int                `json:"tech_count"`
    SuccessCount int                `json:"success_count"`
    Techniques   []DailyTechEntry   `json:"techniques"`
}

type DailyTechEntry struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Success bool   `json:"success"`
    Time    string `json:"time"`   // RFC3339
}
```

`Status` gains: `DailyDigests []DailyDigest json:"daily_digests,omitempty"`

Gap days append a zero-technique entry so the timeline calendar can render them as a distinct
"silent" state rather than treating silence as a missing entry.

### Timeline calendar design

A horizontal strip of day cells rendered from two sources: the pre-computed schedule (same
logic as `updatePoCSchedule()` in JS) overlaid with `status.daily_digests` for completion state.

Cell states:
- **Completed Phase1/Phase2 day**: date label + technique count + green check
- **Current day (waiting)**: date label + pulsing accent indicator
- **Gap day completed**: date label + "gap" (muted)
- **Future day**: date label, dimmed

Implemented in vanilla JS; no new CSS primitives needed beyond what exists.

### delay_after structural fix

`getTechniquesForCampaign()` must be supplemented (or its return type changed) to preserve
`CampaignStep.DelayAfter`. The Phase2 loop in `runPoC()` iterates over steps and calls
`waitOrStop(time.Duration(step.DelayAfter) * time.Second)` after each `runTechnique()` call.
The `CampaignStep` struct already carries `DelayAfter int` — the data is there, just discarded.

---

## Sources

- Codebase analysis: `internal/engine/engine.go` (runPoC full implementation reviewed)
- Codebase analysis: `internal/simlog/simlog.go` (entry types, Phase() function)
- Codebase analysis: `internal/server/static/index.html` (pocInfoPanel, pocSchedulePreview, updatePoCSchedule())
- Codebase analysis: `internal/playbooks/types.go` (CampaignStep.DelayAfter confirmed present)
- PROJECT.md v1.2 Active requirements
- STRUCTURE.md codebase map

---

*Feature research for: LogNoJutsu v1.2 PoC Mode Fix & Overhaul*
*Researched: 2026-04-08*
