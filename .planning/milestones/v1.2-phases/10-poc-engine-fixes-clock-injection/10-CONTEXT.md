# Phase 10: PoC Engine Fixes & Clock Injection - Context

**Gathered:** 2026-04-08
**Status:** Ready for planning

<domain>
## Phase Boundary

Fix four bugs in the PoC engine (`runPoC()` in `internal/engine/engine.go`) and inject a testable clock interface:
1. Stale day counter that resets per-section instead of tracking global day
2. German CurrentStep strings that should be English
3. Missing `simlog.Phase()` separator entries at phase transitions
4. Hard-coded `time.Now()` / `time.After()` preventing deterministic testing

No new features, no UI changes, no API changes. Pure engine correctness and testability.

</domain>

<decisions>
## Implementation Decisions

### Clock Injection (TEST-01)
- **D-01:** Define a `Clock` interface in the engine package with two methods: `Now() time.Time` and `After(d time.Duration) <-chan time.Time`
- **D-02:** Provide a `realClock` struct implementing the interface with standard `time.Now()` and `time.After()` calls
- **D-03:** Add an unexported `clock Clock` field to the `Engine` struct. Default to `realClock{}` in `engine.New()`. Tests supply a fake clock via a functional option or direct field assignment.
- **D-04:** Replace all direct `time.Now()` and `time.After()` calls in `runPoC()` and `waitOrStop()` with `e.clock.Now()` and `e.clock.After()`

### Day Counter Fix (POCFIX-01)
- **D-05:** Introduce a running `globalDay` variable initialized to 0 at the top of `runPoC()`. Increment at the start of each day iteration across all three sections (Phase1, Gap, Phase2).
- **D-06:** Set `e.status.PoCDay = globalDay` at each increment point — replaces the current Phase1-only assignment at line 348
- **D-07:** CurrentStep format strings use `globalDay` and `totalDays` for the "Day N of M" display, not the section-local loop index

### German String Cleanup (POCFIX-02)
- **D-08:** Replace all three German CurrentStep strings with English equivalents:
  - Phase1 (line 351): `"PoC Phase 1 — Day %d of %d — Waiting until %02d:00"` (globalDay, totalDays, hour)
  - Gap (line 389): `"PoC Gap — Day %d of %d (no actions)"` (globalDay, totalDays)
  - Phase2 (line 411): `"PoC Phase 2 — Day %d of %d — Waiting until %02d:00"` (globalDay, totalDays, hour)

### Log Separators (POCFIX-03)
- **D-09:** Add `simlog.Phase()` calls immediately after each `setPhase()` call in `runPoC()`:
  - After `setPhase(PhasePoCPhase1)`: `simlog.Phase("PoC Phase 1: Discovery")`
  - After `setPhase(PhasePoCGap)`: `simlog.Phase("PoC Gap")`
  - After `setPhase(PhasePoCPhase2)`: `simlog.Phase("PoC Phase 2: Attack")`
- **D-10:** Do NOT move `simlog.Phase()` into `setPhase()` — keep it explicit at call sites to avoid duplicating existing calls in normal `run()`

### Claude's Discretion
- Whether to define Clock in `engine.go` or a separate `clock.go` file within the engine package
- Whether to use functional options pattern (`WithClock(c Clock)`) or direct struct field assignment for injecting test clocks
- Internal variable naming for the running day counter (`globalDay`, `dayNum`, etc.)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope
- `.planning/ROADMAP.md` §Phase 10 — requirements POCFIX-01, POCFIX-02, POCFIX-03, TEST-01 and success criteria
- `.planning/REQUIREMENTS.md` §Bug Fixes, §Testability — detailed requirement descriptions

### Primary file to modify
- `internal/engine/engine.go` — `runPoC()` at line 313, `waitOrStop()` at line 657, `setPhase()` at line 650, `Status` struct at line 68

### Established patterns
- `.planning/codebase/CONVENTIONS.md` — naming conventions, error handling, code style
- `.planning/codebase/ARCHITECTURE.md` §Engine as a state machine — phase transitions, stopCh pattern, mutex usage

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `waitOrStop()` at engine.go:657 — currently uses `time.After()` directly; needs to use `e.clock.After()` after injection
- `nextOccurrenceOfHour()` — uses `time.Now()` directly; needs to accept a `now` parameter or use `e.clock.Now()`
- `setPhase()` at engine.go:650 — already handles mutex locking for phase transitions; simlog.Phase() calls go right after it
- `stopCh` channel pattern — remains unchanged; `waitOrStop` keeps its select but sources the timer from Clock interface

### Established Patterns
- `QueryFn` injection in verifier — same dependency injection philosophy applies to Clock
- `engine.New(registry, users)` constructor — extend with optional Clock parameter
- Typed string constants for Phase enum — `PhasePoCPhase1`, `PhasePoCGap`, `PhasePoCPhase2` already defined
- `sync.RWMutex` guards all status writes — maintain this for PoCDay updates

### Integration Points
- `e.status.PoCDay` (line 85) — JSON field `poc_day` consumed by UI polling via `/api/status`
- `e.status.CurrentStep` (line 73) — JSON field `current_step` shown in UI status display
- `simlog.Phase()` entries appear in `/api/logs` response and log viewer UI
- Phase 13 tests will depend on the Clock interface defined here

</code_context>

<specifics>
## Specific Ideas

- Clock interface keeps it minimal — just `Now()` and `After()`, no `Sleep()` or `Timer` abstractions
- Running counter is simplest correct approach — no offset math, just `globalDay++` at each day boundary
- Separator labels include descriptive text ("PoC Phase 1: Discovery") not just phase enum values
- English strings match the "descriptive" style: "Waiting until HH:00" not abbreviated forms

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 10-poc-engine-fixes-clock-injection*
*Context gathered: 2026-04-08*
