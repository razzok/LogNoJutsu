# Phase 10: PoC Engine Fixes & Clock Injection - Research

**Researched:** 2026-04-08
**Domain:** Go engine refactor — clock dependency injection, day counter arithmetic, string localisation, structured logging
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Clock Injection (TEST-01)**
- D-01: Define a `Clock` interface in the engine package with two methods: `Now() time.Time` and `After(d time.Duration) <-chan time.Time`
- D-02: Provide a `realClock` struct implementing the interface with standard `time.Now()` and `time.After()` calls
- D-03: Add an unexported `clock Clock` field to the `Engine` struct. Default to `realClock{}` in `engine.New()`. Tests supply a fake clock via a functional option or direct field assignment.
- D-04: Replace all direct `time.Now()` and `time.After()` calls in `runPoC()` and `waitOrStop()` with `e.clock.Now()` and `e.clock.After()`

**Day Counter Fix (POCFIX-01)**
- D-05: Introduce a running `globalDay` variable initialized to 0 at the top of `runPoC()`. Increment at the start of each day iteration across all three sections (Phase1, Gap, Phase2).
- D-06: Set `e.status.PoCDay = globalDay` at each increment point — replaces the current Phase1-only assignment at line 348
- D-07: CurrentStep format strings use `globalDay` and `totalDays` for the "Day N of M" display, not the section-local loop index

**German String Cleanup (POCFIX-02)**
- D-08: Replace all three German CurrentStep strings with English equivalents:
  - Phase1 (line 351): `"PoC Phase 1 — Day %d of %d — Waiting until %02d:00"` (globalDay, totalDays, hour)
  - Gap (line 389): `"PoC Gap — Day %d of %d (no actions)"` (globalDay, totalDays)
  - Phase2 (line 411): `"PoC Phase 2 — Day %d of %d — Waiting until %02d:00"` (globalDay, totalDays, hour)

**Log Separators (POCFIX-03)**
- D-09: Add `simlog.Phase()` calls immediately after each `setPhase()` call in `runPoC()`:
  - After `setPhase(PhasePoCPhase1)`: `simlog.Phase("PoC Phase 1: Discovery")`
  - After `setPhase(PhasePoCGap)`: `simlog.Phase("PoC Gap")`
  - After `setPhase(PhasePoCPhase2)`: `simlog.Phase("PoC Phase 2: Attack")`
- D-10: Do NOT move `simlog.Phase()` into `setPhase()` — keep it explicit at call sites

### Claude's Discretion
- Whether to define Clock in `engine.go` or a separate `clock.go` file within the engine package
- Whether to use functional options pattern (`WithClock(c Clock)`) or direct struct field assignment for injecting test clocks
- Internal variable naming for the running day counter (`globalDay`, `dayNum`, etc.)

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| POCFIX-01 | PoC day counter updates correctly across all three phases (Phase1, Gap, Phase2) — showing global day N of total | Single `globalDay` int initialized to 0, incremented once per loop iteration across all three sections before setting `e.status.PoCDay` |
| POCFIX-02 | All CurrentStep strings in runPoC() display in English (no German "Tag", "warte bis", "keine Aktionen") | Three targeted `fmt.Sprintf` string replacements at lines 351, 389, 411 in engine.go |
| POCFIX-03 | Phase transitions in runPoC() produce simlog.Phase() separator entries visible in log viewer | Three `simlog.Phase(label)` calls placed immediately after the three `setPhase()` calls in runPoC() |
| TEST-01 | Engine accepts injectable clock/wait function for deterministic runPoC() testing | `Clock` interface + `realClock` struct + `clock Clock` field on `Engine`; replace direct `time.Now()`/`time.After()` calls in runPoC() and waitOrStop() |
</phase_requirements>

---

## Summary

Phase 10 is a surgical refactor of `internal/engine/engine.go`. All four changes target `runPoC()` and its helper functions — nothing in the API layer, no UI changes, no new packages required. The code surface is well-understood: all bugs are visible in the existing source, all fix locations are known, and the patterns to follow (QueryFn injection, RunnerFunc injection, stop-channel, mutex guards) are already established in the codebase.

The most architecturally significant change is the Clock interface (TEST-01). It follows the exact same dependency injection philosophy already used by `RunnerFunc` and `QueryFn` in the verifier. Inserting it correctly requires updating the `Engine` struct, `engine.New()`, `waitOrStop()`, `nextOccurrenceOfHour()`, and the three `time.Now()` calls inside `runPoC()`. The interface itself is minimal — two methods — with zero external dependencies.

The remaining three fixes (POCFIX-01, POCFIX-02, POCFIX-03) are string and variable substitutions with no architectural impact. All correctness is in the arithmetic (globalDay must increment before the status write, totalDays must be used rather than section-local counts) and the placement of `simlog.Phase()` calls (after `setPhase()`, never inside it).

**Primary recommendation:** Implement in a single plan covering all four fixes together — they all live in the same file and the clock injection touches the same lines as the day counter and string fixes.

---

## Standard Stack

No new dependencies. This phase modifies existing Go stdlib-only code.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `time` | go 1.26.1 | `time.Time`, `time.Duration`, `time.After` | Already used throughout engine.go |
| Go stdlib `sync` | go 1.26.1 | `sync.RWMutex` for status writes | Existing pattern in Engine struct |
| Go stdlib `fmt` | go 1.26.1 | `fmt.Sprintf` for CurrentStep strings | Existing pattern |

### No New Dependencies
The `Clock` interface and `realClock` implementation use only stdlib `time`. No external packages needed.

**Installation:** Nothing to install.

---

## Architecture Patterns

### Recommended File Structure

The Clock interface can live in `engine.go` or a new `clock.go`. Given that the engine package currently has one source file (`engine.go`) and the convention is "one file per package in most packages" (CONVENTIONS.md), placing the interface in `engine.go` is consistent. A separate `clock.go` is acceptable if the implementer prefers cleaner file organisation, since the `playbooks` package already demonstrates a two-file split.

```
internal/engine/
├── engine.go          # Engine struct, Clock interface, realClock, all methods
└── engine_test.go     # Existing tests + new Clock-based tests (Wave 0 gap)
```

Alternative (discretion):
```
internal/engine/
├── engine.go          # Engine struct and all methods
├── clock.go           # Clock interface + realClock struct
└── engine_test.go
```

### Pattern 1: Clock Interface (Dependency Injection)

**What:** Define a minimal `Clock` interface; inject a fake in tests; default to `realClock{}` in production.

**When to use:** Any function that calls `time.Now()` or `time.After()` and must be testable without real wall-clock sleeps.

**Example (follows RunnerFunc/QueryFn pattern already in codebase):**
```go
// Clock abstracts time operations for deterministic testing.
type Clock interface {
    Now() time.Time
    After(d time.Duration) <-chan time.Time
}

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
```

**Injection in New():**
```go
func New(registry *playbooks.Registry, users *userstore.Store) *Engine {
    return &Engine{
        registry: registry,
        users:    users,
        clock:    realClock{},
        // ... existing fields unchanged
    }
}
```

**Test injection (direct field access — package `engine` tests are in-package):**
```go
eng := New(reg, testUsers())
eng.clock = fakeClock{...}  // direct field assignment; tests are in package engine
```

Since `engine_test.go` uses `package engine` (not `package engine_test`), direct field access to `eng.clock` is legal without any exported setter.

### Pattern 2: Global Day Counter

**What:** A single `globalDay int` declared at the top of `runPoC()`. Every loop body across all three sections does `globalDay++` as its first statement, before updating status.

**When to use:** Whenever multiple loop ranges must contribute to a single monotonically increasing count.

**Example:**
```go
func (e *Engine) runPoC() {
    cfg := e.cfg
    totalDays := cfg.Phase1DurationDays + cfg.GapDays + cfg.Phase2DurationDays
    globalDay := 0

    // Phase 1
    e.setPhase(PhasePoCPhase1)
    simlog.Phase("PoC Phase 1: Discovery")
    for day := 1; day <= cfg.Phase1DurationDays; day++ {
        globalDay++
        e.mu.Lock()
        e.status.PoCDay = globalDay
        e.status.CurrentStep = fmt.Sprintf("PoC Phase 1 — Day %d of %d — Waiting until %02d:00",
            globalDay, totalDays, cfg.Phase1DailyHour)
        e.mu.Unlock()
        // ...
    }

    // Gap
    if cfg.GapDays > 0 {
        e.setPhase(PhasePoCGap)
        simlog.Phase("PoC Gap")
        for day := 1; day <= cfg.GapDays; day++ {
            globalDay++
            e.mu.Lock()
            e.status.PoCDay = globalDay
            e.status.CurrentStep = fmt.Sprintf("PoC Gap — Day %d of %d (no actions)",
                globalDay, totalDays)
            e.mu.Unlock()
            // ...
        }
    }

    // Phase 2
    e.setPhase(PhasePoCPhase2)
    simlog.Phase("PoC Phase 2: Attack")
    for day := 1; day <= cfg.Phase2DurationDays; day++ {
        globalDay++
        e.mu.Lock()
        e.status.PoCDay = globalDay
        e.status.CurrentStep = fmt.Sprintf("PoC Phase 2 — Day %d of %d — Waiting until %02d:00",
            globalDay, totalDays, cfg.Phase2DailyHour)
        e.mu.Unlock()
        // ...
    }
}
```

### Pattern 3: nextOccurrenceOfHour with Clock

**What:** `nextOccurrenceOfHour(hour int)` currently calls `time.Now()` directly. After clock injection, it needs access to the clock. Two options:

1. Make it a method: `func (e *Engine) nextOccurrenceOfHour(hour int) time.Duration` — uses `e.clock.Now()` internally.
2. Accept a `now time.Time` parameter: `func nextOccurrenceOfHour(hour int, now time.Time) time.Duration` — caller passes `e.clock.Now()`.

Option 2 is the simpler refactor. The function becomes a pure function with no receiver dependency, easier to unit test in isolation.

### Pattern 4: waitOrStop with Clock

**What:** `waitOrStop(d time.Duration)` currently uses `time.After(d)` directly. After injection, it uses `e.clock.After(d)`.

**Example:**
```go
func (e *Engine) waitOrStop(d time.Duration) bool {
    select {
    case <-e.clock.After(d):
        return true
    case <-e.stopCh:
        return false
    }
}
```

The channel-based `select` structure is unchanged — only the timer source changes.

### Pattern 5: setPhase + simlog.Phase placement

**What:** `simlog.Phase()` must be called at each call site immediately after `setPhase()`, not inside `setPhase()` itself (D-10).

**Why:** `setPhase()` is also called from the normal `run()` path, which has its own `simlog.Phase()` calls. Moving it inside would duplicate entries on the normal path.

**Example:**
```go
e.setPhase(PhasePoCPhase1)
simlog.Phase("PoC Phase 1: Discovery")
```

### Anti-Patterns to Avoid

- **Adding `simlog.Phase()` inside `setPhase()`:** Would produce double entries for normal simulation runs. D-10 explicitly forbids this.
- **Using section-local loop variable `day` in `CurrentStep`:** The section counter resets to 1 for each section. Must use `globalDay`.
- **Using section duration (e.g., `cfg.Phase1DurationDays`) as the denominator in CurrentStep:** Must use `totalDays` for "Day N of M".
- **Forgetting to update `e.status.PoCDay` in the Gap section:** The original code only sets `PoCDay` in Phase1. The fix must set it in all three sections.
- **Calling `time.Sleep()` directly in verification flow:** `runTechnique()` contains a `time.Sleep(waitSecs * time.Second)` for verification. This is NOT in scope for Phase 10 — leave it unchanged. Clock injection only targets `runPoC()` and `waitOrStop()`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Fake clock for tests | Custom time-mocking library | Simple `fakeClock` struct implementing the `Clock` interface | Two-method interface needs no framework; a struct with `nowFn` and `afterFn` fields is sufficient |
| Monotonic day counting | Complex offset arithmetic | Plain `globalDay++` counter | Running counter with no math is correct and trivially verifiable |

**Key insight:** Both the clock interface and the day counter are simpler than they look. The temptation is to over-engineer — a two-method interface and an integer increment are the complete solution.

---

## Common Pitfalls

### Pitfall 1: GapDays == 0 — globalDay Stays Correct

**What goes wrong:** When `cfg.GapDays == 0`, the gap loop body never executes. `globalDay` correctly skips those increments. Phase2 then starts at `Phase1DurationDays + 1`. This is correct behaviour but easy to break if someone moves `globalDay++` outside the loop body.

**How to avoid:** Keep `globalDay++` as the first statement inside each loop body, not before the loop.

### Pitfall 2: PoCDay vs. section-local `day` confusion

**What goes wrong:** The original code uses the per-section `day` variable (1-indexed within that section) to set `PoCDay`. A Phase2 day 1 would show as "Day 1" even though it's actually day 8 (if Phase1 was 7 days). The fix requires `globalDay` everywhere `day` was used for status display.

**Warning signs:** In the log, "Day 1 of N" appears at the start of Phase2 — the counter has reset.

### Pitfall 3: nextOccurrenceOfHour still uses time.Now() after refactor

**What goes wrong:** If `nextOccurrenceOfHour` is not updated alongside `waitOrStop`, the fake clock controls the timer channel but real-time is still used to compute the sleep duration. In tests this could produce unexpected sleep durations.

**How to avoid:** Update `nextOccurrenceOfHour` to accept a `now time.Time` parameter (or convert to a receiver method), and call it with `e.clock.Now()`.

### Pitfall 4: Missing mutex around PoCDay update in Gap section

**What goes wrong:** The Gap section currently acquires the mutex for `PoCPhase` and `NextScheduledRun` but not `PoCDay`. If `PoCDay` update is added without a mutex, the race detector catches it. The test suite runs the race detector.

**How to avoid:** All writes to `e.status` fields must be under `e.mu.Lock()`.

### Pitfall 5: setPhase already acquires the mutex

**What goes wrong:** `setPhase()` acquires `e.mu.Lock()` internally. Calling `setPhase()` from within a block that already holds `e.mu.Lock()` would deadlock.

**How to avoid:** `simlog.Phase()` calls go AFTER `setPhase()` returns, never inside a mutex block. The current call sites in `runPoC()` do not hold the mutex when calling `setPhase()` — preserve this.

---

## Code Examples

### Verified: Current buggy strings (engine.go lines 351, 389, 411)

```go
// Line 351 — Phase 1 (CURRENT — German, wrong denominator)
e.status.CurrentStep = fmt.Sprintf("PoC Phase 1 — Tag %d/%d — warte bis %02d:00 Uhr", day, cfg.Phase1DurationDays, cfg.Phase1DailyHour)

// Line 389 — Gap (CURRENT — German, wrong denominator)
e.status.CurrentStep = fmt.Sprintf("PoC Pause — Tag %d/%d (keine Aktionen)", day, cfg.GapDays)

// Line 411 — Phase 2 (CURRENT — German, wrong denominator)
e.status.CurrentStep = fmt.Sprintf("PoC Phase 2 — Tag %d/%d — warte bis %02d:00 Uhr", day, cfg.Phase2DurationDays, cfg.Phase2DailyHour)
```

### Verified: Current waitOrStop (engine.go line 657)

```go
func (e *Engine) waitOrStop(d time.Duration) bool {
    select {
    case <-time.After(d):   // <-- must become e.clock.After(d)
        return true
    case <-e.stopCh:
        return false
    }
}
```

### Verified: Current nextOccurrenceOfHour (engine.go line 304)

```go
func nextOccurrenceOfHour(hour int) time.Duration {
    now := time.Now()   // <-- must accept now as parameter or use e.clock.Now()
    next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
    if !next.After(now) {
        next = next.Add(24 * time.Hour)
    }
    return next.Sub(now)
}
```

### Verified: Existing RunnerFunc injection pattern (engine.go line 114) — Clock follows same shape

```go
// RunnerFunc abstracts technique execution for testability.
// Mirrors the QueryFn pattern in the verifier package.
type RunnerFunc func(t *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult
```

### Verified: Existing engine_test.go is package-internal (in-package access)

```go
package engine  // line 1 of engine_test.go — direct struct field access is legal
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Direct `time.Now()` / `time.After()` | `e.clock.Now()` / `e.clock.After()` | Phase 10 | Enables deterministic tests without real sleeps |
| Section-local `day` variable for PoCDay | `globalDay` running counter | Phase 10 | Day counter monotonically increases across all three phases |
| German CurrentStep strings | English equivalents | Phase 10 | UI displays correct language |
| No log separator at phase transitions | `simlog.Phase()` call after each `setPhase()` | Phase 10 | Log viewer shows visible phase boundaries |

---

## Open Questions

1. **Functional option vs. direct field assignment for Clock injection**
   - What we know: `engine_test.go` is in `package engine` (in-package), so `eng.clock = fakeClock{}` is legal without any exported setter.
   - What's unclear: Whether a `WithClock(c Clock)` functional option is desired for future external test packages (e.g., Phase 13 tests).
   - Recommendation: Direct field assignment is sufficient for Phase 10 since all tests are in-package. A functional option can be added in Phase 13 if needed. Both are within Claude's discretion.

2. **nextOccurrenceOfHour refactor style**
   - What we know: The function is only called from `runPoC()` — two call sites.
   - What's unclear: Method receiver vs. parameter injection.
   - Recommendation: Accept `now time.Time` as a parameter. This keeps it a pure function, requires no receiver, and is simpler to test in isolation.

---

## Environment Availability

Step 2.6: All dependencies are stdlib Go — no external tools, services, or CLIs beyond what's already verified as present.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Build + test | Yes | go 1.26.1 | — |
| `go test ./internal/engine/...` | Validation | Yes | verified above | — |

No missing dependencies.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` |
| Config file | none (no go test flags file; run directly) |
| Quick run command | `go test ./internal/engine/... -timeout 30s` |
| Full suite command | `go test ./... -timeout 60s` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| POCFIX-01 | globalDay increments monotonically across Phase1→Gap→Phase2 | unit | `go test ./internal/engine/... -run TestPoCDayCounter -timeout 30s` | Wave 0 gap |
| POCFIX-02 | CurrentStep strings contain no German text | unit | `go test ./internal/engine/... -run TestPoCCurrentStepStrings -timeout 30s` | Wave 0 gap |
| POCFIX-03 | simlog.Phase() entries appear after each phase transition | unit | `go test ./internal/engine/... -run TestPoCPhaseLogSeparators -timeout 30s` | Wave 0 gap |
| TEST-01 | Clock interface accepted; fake clock eliminates real sleeps | unit | `go test ./internal/engine/... -run TestPoCClockInjection -timeout 30s` | Wave 0 gap |

### Sampling Rate

- **Per task commit:** `go test ./internal/engine/... -timeout 30s`
- **Per wave merge:** `go test ./... -timeout 60s`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/engine/engine_test.go` — add `TestPoCDayCounter`, `TestPoCCurrentStepStrings`, `TestPoCPhaseLogSeparators`, `TestPoCClockInjection` — covers POCFIX-01, POCFIX-02, POCFIX-03, TEST-01

The existing test file exists and compiles. New test functions are appended to it. No new file creation required; no framework install needed.

**Fake clock design for tests:** Since `runPoC()` calls `waitOrStop()` which blocks on a channel, the fake clock's `After()` must return a channel that fires immediately (or on demand). A minimal fake:

```go
type fakeClock struct {
    now time.Time
}

func (f *fakeClock) Now() time.Time { return f.now }

func (f *fakeClock) After(d time.Duration) <-chan time.Time {
    ch := make(chan time.Time, 1)
    ch <- f.now.Add(d)   // fires immediately — no real sleep
    return ch
}
```

---

## Sources

### Primary (HIGH confidence)
- `internal/engine/engine.go` — direct source inspection, all bug locations verified at lines 304, 338-432, 650-663
- `internal/engine/engine_test.go` — confirmed in-package (`package engine`), test helpers verified
- `.planning/codebase/CONVENTIONS.md` — naming conventions, mutex patterns, constructor convention
- `.planning/codebase/ARCHITECTURE.md` — Engine state machine, stopCh pattern, RunnerFunc injection precedent
- `go.mod` — Go version 1.26.1, no external dependencies

### Secondary (MEDIUM confidence)
- `.planning/phases/10-poc-engine-fixes-clock-injection/10-CONTEXT.md` — user decisions D-01 through D-10, canonical fix locations

### Tertiary (LOW confidence)
- None

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — stdlib only, no new packages
- Architecture: HIGH — all patterns verified in existing source
- Pitfalls: HIGH — bugs are visible in source; fix locations confirmed
- Test design: HIGH — existing test infrastructure confirmed running; fake clock pattern is standard Go

**Research date:** 2026-04-08
**Valid until:** 2026-05-08 (stable codebase — no fast-moving dependencies)
