---
phase: 10-poc-engine-fixes-clock-injection
plan: 01
subsystem: engine
tags: [go, clock-injection, poc-mode, testability, engine]

# Dependency graph
requires:
  - phase: 09-ui-polish
    provides: complete v1.1 codebase baseline
provides:
  - Clock interface with Now() and After() in engine package
  - realClock production implementation defaulting in Engine.New()
  - globalDay counter for monotonic day tracking across PoC sections
  - English-only CurrentStep strings in runPoC()
  - simlog.Phase() separators at each PoC phase transition
affects: [13-poc-engine-tests]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Clock interface injection — same pattern as QueryFn in verifier; unexported field in struct, realClock{} default in constructor"

key-files:
  created: []
  modified:
    - internal/engine/engine.go

key-decisions:
  - "Clock interface defined inline in engine.go (not clock.go) — single file, minimal surface"
  - "nextOccurrenceOfHour accepts now time.Time parameter — pure function, no hidden state"
  - "globalDay++ at top of each section loop — monotonic counter, no offset math"
  - "simlog.Phase() at call sites not inside setPhase() — avoids duplication with normal run() calls"

patterns-established:
  - "Clock injection: type Clock interface { Now() / After() }, realClock{} default in constructor, e.clock.* at all use sites"

requirements-completed: [POCFIX-01, POCFIX-02, POCFIX-03, TEST-01]

# Metrics
duration: 3min
completed: 2026-04-08
---

# Phase 10 Plan 01: PoC Engine Fixes & Clock Injection Summary

**Injectable Clock interface (Now/After) wired into Engine struct, globalDay monotonic counter, English CurrentStep strings, and simlog.Phase separators after each PoC phase transition**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-04-08T19:55:50Z
- **Completed:** 2026-04-08T19:58:27Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments

- Added `Clock` interface and `realClock` production implementation to engine package; wired into `Engine` struct with `clock: realClock{}` default in constructor
- Replaced all `time.After(d)` and `time.Now()` calls in `waitOrStop`, `nextOccurrenceOfHour`, `setPhase`, and `runPoC` with `e.clock.After()` / `e.clock.Now()`
- Fixed stale day counter: `globalDay` increments monotonically across Phase1, Gap, and Phase2 loops; `e.status.PoCDay` now reflects global day not section-local index
- Replaced three German `CurrentStep` format strings with English equivalents using `globalDay`/`totalDays`
- Added `simlog.Phase()` separator calls after each `setPhase()` in `runPoC()` — three separators for Discovery, Gap, and Attack phases

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Clock interface and wire into Engine struct** - `7bbe16a` (feat)
2. **Task 2: Fix day counter, German strings, and add log separators** - `d0eaf02` (fix)

**Plan metadata:** (docs commit follows)

## Files Created/Modified

- `internal/engine/engine.go` — Clock interface + realClock, clock field in Engine, globalDay counter, English strings, simlog.Phase separators

## Decisions Made

- Clock interface defined inline in `engine.go` rather than a separate file — single file modification, minimal surface area
- `nextOccurrenceOfHour` changed to pure function with `now time.Time` parameter rather than calling `e.clock.Now()` internally — keeps it a pure function consistent with its existing design
- `globalDay` counter chosen over offset math — simplest correct approach with no edge cases at section boundaries
- `simlog.Phase()` kept at call sites in `runPoC()` not moved into `setPhase()` — preserves existing normal `run()` behavior which calls `simlog.Phase("discovery")` / `simlog.Phase("attack")` directly

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `Clock` interface ready for Phase 13 deterministic scheduling tests via direct struct field assignment (`engine.clock = &fakeClock{}`)
- All four PoC engine bugs resolved — `runPoC()` is now correct for production use on day 1+
- `go build ./...` and `go vet ./internal/engine/...` pass cleanly

---
*Phase: 10-poc-engine-fixes-clock-injection*
*Completed: 2026-04-08*
