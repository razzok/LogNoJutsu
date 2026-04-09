# Phase 11: Daily Tracking Backend & Campaign Delay — Research

**Researched:** 2026-04-09
**Domain:** Go engine internals — struct extension, mutex-guarded state, HTTP endpoint, injectable clock
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-01:** DayDigest carries counts only — day number, phase label, status enum (pending/active/complete), technique count, pass count, fail count, start timestamp, end timestamp, last heartbeat timestamp. No per-technique detail (that data lives in `engine.status.Results`).

**D-02:** All DayDigest entries are pre-populated as "pending" at the top of `runPoC()` before the first day loop begins. The full schedule is visible from the very first `/api/poc/days` poll.

**D-03:** DayDigest slice lives as a `[]DayDigest` field on the Engine struct (not inside Status). Guarded by the existing `sync.RWMutex`. Returned via a dedicated getter method.

**D-04:** `GET /api/poc/days` returns the full `[]DayDigest` array. No per-day lookup endpoint needed — max ~30 entries for any PoC run.

**D-05:** Endpoint is behind `authMiddleware`, consistent with `/api/status`, `/api/logs`, and all other data endpoints.

**D-06:** When no PoC is running (or engine is idle), return HTTP 200 with empty JSON array `[]`. Phase 12 UI handles the empty state.

**D-07:** `CampaignStep.DelayAfter` is treated as seconds. A value of 300 means a 5-minute delay.

**D-08:** Delay uses the injectable clock via `e.clock.After(time.Duration(step.DelayAfter) * time.Second)`. Tests with fakeClock skip delays instantly.

**D-09:** Delay is interruptible — use existing `waitOrStop(d)` which selects on `e.clock.After` AND `stopCh`. User can stop the engine mid-delay.

**D-10:** LastHeartbeat is updated at key events only: day start, each technique execution, day completion. No background timer goroutine needed.

**D-11:** Heartbeat semantics: proves the engine was alive during the execution window. If timestamp is stale (e.g., >10min for an hourly schedule), something may be wrong.

### Claude's Discretion

- Internal naming of the DayDigest status enum values (e.g., `DayPending`, `DayActive`, `DayComplete`)
- Whether to add a `GetDayDigests()` method or expose the slice through `GetStatus()` with a separate field
- Handler function naming for the new endpoint (`handlePoCDays`, `handleDays`, etc.)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TRACK-01 | Engine records a DayDigest struct per PoC day containing: day number, phase, status, techniques executed, pass/fail counts, start/end timestamps | DayDigest struct design; pre-population loop in `runPoC()`; mutex-write points at day start, technique complete, day end |
| TRACK-02 | DayDigest entries are pre-populated as "pending" at runPoC() start so the full schedule is visible from first poll | Pre-population loop using same phase boundary arithmetic already in `runPoC()`; phaseForDay helper |
| TRACK-03 | GET /api/poc/days endpoint returns the DayDigest array (behind authMiddleware) | `handleStatus` pattern exactly replicates; `GetDayDigests()` getter; empty-slice normalisation |
| TRACK-04 | DayDigest includes a "last heartbeat" timestamp proving the engine was alive during each day's execution window | `LastHeartbeat` field on DayDigest; updated at day-start, per-technique, day-end using `e.clock.Now()` |
| CAMP-01 | Campaign delay_after field is applied between technique steps during PoC Phase 2 execution | `CampaignStep.DelayAfter` already parsed; `getTechniquesForCampaign()` discards steps today — must switch to iterating `campaign.Steps` directly in Phase 2 loop; `waitOrStop(d)` reuse |
</phase_requirements>

---

## Summary

Phase 11 is a pure backend extension of `internal/engine/engine.go` and `internal/server/server.go`. No new packages, no new files, no schema changes to YAML playbooks. The work consists of four tightly coupled changes: (1) define a `DayDigest` struct and a `DayStatus` typed-string enum in `engine.go`; (2) add a `[]DayDigest` field to the `Engine` struct, pre-populate it in `runPoC()`, and update individual entries at day start / technique completion / day end; (3) add a `GetDayDigests()` getter method and register a `handlePoCDays` handler at `GET /api/poc/days`; and (4) apply `CampaignStep.DelayAfter` during Phase 2 by iterating `campaign.Steps` directly instead of using the `getTechniquesForCampaign()` helper (which discards step metadata).

The existing `fakeClock`, `waitOrStop`, `sync.RWMutex`, and `captureClock` test infrastructure fully supports all new code paths without modification. All test assertions will be in `engine_test.go` in the `engine` package (white-box, unexported field access allowed).

**Primary recommendation:** Implement in a single commit touching `engine.go` and `server.go`. The DayDigest pre-population and mutation are localized to `runPoC()`. The campaign delay change requires a targeted refactor of the Phase 2 inner loop only — the existing `getTechniquesForCampaign()` helper remains correct for normal (non-PoC) mode.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `sync` | Go 1.26.1 | `sync.RWMutex` guards DayDigest slice | Already in use for `Status`; same pattern |
| Go stdlib `time` | Go 1.26.1 | Timestamps, `time.Duration` for delays | All timestamps via `e.clock.Now()` for testability |
| Go stdlib `encoding/json` | Go 1.26.1 | JSON serialisation of `[]DayDigest` | Same as all other API responses |
| Go stdlib `net/http` | Go 1.26.1 | New `/api/poc/days` handler | Identical pattern to `handleStatus` |

No new dependencies. The module has only one external dependency (`gopkg.in/yaml.v3`) and this phase does not require adding any others.

**Installation:** None required.

---

## Architecture Patterns

### Pattern 1: Typed-string status enum (project convention)

All domain enumerations use `type X string` + `const` block. For `DayStatus`:

```go
// Source: CONVENTIONS.md §Typed string constants for enumerations
type DayStatus string

const (
    DayPending  DayStatus = "pending"
    DayActive   DayStatus = "active"
    DayComplete DayStatus = "complete"
)
```

### Pattern 2: DayDigest struct definition

Mirrors the existing `Status` struct style (json tags aligned, `omitempty` on optional fields). The struct is purely counts — no per-technique slice.

```go
// DayDigest records per-day execution summary for a PoC run.
type DayDigest struct {
    Day           int       `json:"day"`
    Phase         string    `json:"phase"`            // "phase1", "gap", "phase2"
    Status        DayStatus `json:"status"`
    TechniqueCount int      `json:"technique_count"`
    PassCount     int       `json:"pass_count"`
    FailCount     int       `json:"fail_count"`
    StartTime     string    `json:"start_time,omitempty"`
    EndTime       string    `json:"end_time,omitempty"`
    LastHeartbeat string    `json:"last_heartbeat,omitempty"`
}
```

### Pattern 3: Engine struct field addition

Add `dayDigests []DayDigest` to the `Engine` struct alongside `status`, guarded by the same `e.mu` mutex (per D-03). Also add `dayDigests` reset in `Start()` alongside `executedTechniques = nil`.

```go
// internal/engine/engine.go — Engine struct (lines 92-105)
type Engine struct {
    mu                 sync.RWMutex
    status             Status
    dayDigests         []DayDigest   // ADD — per-day tracking, guarded by mu
    registry           *playbooks.Registry
    // ... (existing fields unchanged)
}
```

Reset in `Start()`:

```go
e.dayDigests = nil   // reset at same point as e.executedTechniques = nil
```

### Pattern 4: Pre-population loop in runPoC()

Insert immediately after `e.status.PoCTotalDays = totalDays` (line 337), before Phase 1 begins. Uses the same phase boundary arithmetic already computed from `cfg.*DurationDays`.

```go
// Pre-populate all days as pending so /api/poc/days returns full schedule from first poll (D-02)
digests := make([]DayDigest, totalDays)
globalIdx := 0
for i := 0; i < cfg.Phase1DurationDays; i++ {
    digests[globalIdx] = DayDigest{Day: globalIdx + 1, Phase: "phase1", Status: DayPending, TechniqueCount: techsPerDay}
    globalIdx++
}
for i := 0; i < cfg.GapDays; i++ {
    digests[globalIdx] = DayDigest{Day: globalIdx + 1, Phase: "gap", Status: DayPending}
    globalIdx++
}
for i := 0; i < cfg.Phase2DurationDays; i++ {
    digests[globalIdx] = DayDigest{Day: globalIdx + 1, Phase: "phase2", Status: DayPending}
    globalIdx++
}
e.mu.Lock()
e.dayDigests = digests
e.mu.Unlock()
```

**Note:** `techsPerDay` is computed above the pre-population loop (line 346 in current engine.go), so it is available. Phase 2 technique count for gap days is 0 and for phase2 is not statically known (campaign length), so leave `TechniqueCount` as 0 at pre-population for those — it gets filled at day-start when actual techniques are resolved.

### Pattern 5: Day lifecycle mutations

At each key event, acquire `e.mu.Lock()`, update `e.dayDigests[globalDay-1]`, release. This mirrors the existing pattern for `e.status.PoCDay` updates throughout `runPoC()`.

**Day start (all phases):**
```go
e.mu.Lock()
e.dayDigests[globalDay-1].Status = DayActive
e.dayDigests[globalDay-1].StartTime = e.clock.Now().Format(time.RFC3339)
e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
e.mu.Unlock()
```

**After each technique execution (Phase 1 and Phase 2 inner loops):**
```go
// After e.runTechnique(t) returns
e.mu.Lock()
e.dayDigests[globalDay-1].PassCount++    // or FailCount++ based on last result
e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
e.mu.Unlock()
```

**Determining pass/fail:** `runTechnique()` appends to `e.status.Results`. After it returns, the last result is `e.status.Results[len(e.status.Results)-1]`. However, reading `e.status.Results` requires holding the lock. A cleaner approach: check the result returned by reading the last element of `e.status.Results` under read lock, or restructure so `runTechnique` returns the result (minor refactor). **Simpler pattern:** `runTechnique` appends under `e.mu.Lock()` — so immediately after `e.runTechnique(t)`, acquire RLock, read `e.status.Results[len-1].Success`, release, then acquire Lock for dayDigest update.

**Day complete:**
```go
e.mu.Lock()
e.dayDigests[globalDay-1].Status = DayComplete
e.dayDigests[globalDay-1].EndTime = e.clock.Now().Format(time.RFC3339)
e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
e.mu.Unlock()
```

### Pattern 6: GetDayDigests() getter

Exact mirror of `GetStatus()`. Returns a copy to avoid external mutation:

```go
// GetDayDigests returns a copy of the per-day digest slice.
func (e *Engine) GetDayDigests() []DayDigest {
    e.mu.RLock()
    defer e.mu.RUnlock()
    if len(e.dayDigests) == 0 {
        return []DayDigest{}    // never return nil — JSON must be [] not null
    }
    result := make([]DayDigest, len(e.dayDigests))
    copy(result, e.dayDigests)
    return result
}
```

### Pattern 7: handlePoCDays in server.go

Exact mirror of `handleStatus`. Registers under the PoC data section separator:

```go
// in registerRoutes():
mux.HandleFunc("/api/poc/days", s.authMiddleware(s.handlePoCDays))

// handler:
func (s *Server) handlePoCDays(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, s.eng.GetDayDigests())
}
```

### Pattern 8: Campaign delay_after in Phase 2

**The key problem:** `getTechniquesForCampaign()` currently resolves a `[]*Technique` slice, discarding `CampaignStep` metadata (including `DelayAfter`). To apply `DelayAfter`, the Phase 2 loop must iterate over `campaign.Steps` directly and apply the delay after each step.

Replace the current Phase 2 inner loop (which calls `getTechniquesForPhase()` → `getTechniquesForCampaign()`) with an inline loop over `campaign.Steps`:

```go
// Phase 2 day inner loop — replaces the current attackTechs loop
if e.cfg.CampaignID != "" {
    campaign, ok := e.registry.Campaigns[e.cfg.CampaignID]
    if ok {
        for _, step := range campaign.Steps {
            if e.isStopped() { e.abort(); return }
            t, exists := e.registry.Techniques[step.TechniqueID]
            if !exists { continue }
            e.runTechnique(t)
            // heartbeat update here
            if step.DelayAfter > 0 {
                if !e.waitOrStop(time.Duration(step.DelayAfter) * time.Second) {
                    e.abort()
                    return
                }
            }
        }
    }
} else {
    // non-campaign Phase 2 — use existing logic
    attackTechs := e.getTechniquesForPhase()
    for _, t := range attackTechs {
        if e.isStopped() { e.abort(); return }
        e.runTechnique(t)
    }
}
```

The existing `getTechniquesForPhase()` / `getTechniquesForCampaign()` helpers remain unchanged — they are still used in non-PoC `run()`.

### Anti-Patterns to Avoid

- **Storing DayDigest inside Status:** D-03 explicitly rejects this. Keeps the Status JSON surface clean and avoids bloating every `/api/status` poll with digest data.
- **Returning nil instead of empty slice from GetDayDigests():** JSON would encode as `null` instead of `[]`, breaking Phase 12 UI expectations (D-06).
- **Forgetting to reset dayDigests in Start():** If a second PoC run starts, stale digests from the previous run would persist.
- **Using time.Now() instead of e.clock.Now() for timestamps:** All times in runPoC scope must go through the injectable clock so tests remain deterministic.
- **Adding a background heartbeat goroutine:** D-10 explicitly forbids this. Heartbeat updates are event-driven only.
- **Calling waitOrStop for DelayAfter == 0:** D-09 says skip entirely. Avoids a channel select with zero duration (which would fire on the timer immediately but wastes a context switch).
- **Locking e.mu for the entire day execution:** runTechnique() acquires the lock internally. Lock only the minimal critical sections (pre-pop, per-event updates).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Interruptible delay | Custom goroutine + channel | `waitOrStop(d)` | Already handles stopCh correctly; reinventing breaks cancellation contract |
| Clock abstraction | Real time.After | `e.clock.After(d)` | Testing with real clock causes real sleep durations; fakeClock skips instantly |
| JSON empty-slice guarantee | `if digests == nil { return [] }` in every caller | Return `[]DayDigest{}` from getter | Centralise the nil-to-empty normalisation once |
| Per-technique result tracking | Second results slice | Read from `e.status.Results[len-1]` after runTechnique | Data already captured there under lock |

---

## Common Pitfalls

### Pitfall 1: Pass/fail counting requires reading Results after runTechnique

**What goes wrong:** `runTechnique()` does not return a value — it appends to `e.status.Results` internally under a lock. Attempting to count pass/fail by inspecting anything before the lock is released causes a data race or stale read.

**Why it happens:** The method signature is `func (e *Engine) runTechnique(t *playbooks.Technique)` — void return.

**How to avoid:** After `e.runTechnique(t)` returns, acquire `e.mu.RLock()`, read `e.status.Results[len(e.status.Results)-1].Success`, release, then acquire `e.mu.Lock()` for dayDigest update.

**Warning signs:** `-race` test flag reporting a concurrent write/read on `status.Results`.

### Pitfall 2: Phase 2 campaign step loop loses access to step metadata

**What goes wrong:** Using `getTechniquesForCampaign()` returns `[]*Technique` only — `DelayAfter` is lost. No panic, no compile error; delay is simply never applied.

**Why it happens:** The helper was written before `DelayAfter` needed to be consumed.

**How to avoid:** In the PoC Phase 2 day loop, inline the `campaign.Steps` iteration directly (as shown in Pattern 8). The existing helper remains correct for non-PoC mode.

**Warning signs:** Test asserting `DelayAfter > 0` causes a delay would never fail — fakeClock's After() would need to be called once per step.

### Pitfall 3: dayDigests index is off-by-one

**What goes wrong:** `globalDay` is incremented at the top of each day loop (before the day executes). So the correct index into `dayDigests` is `globalDay - 1`.

**Why it happens:** `globalDay` starts at 0 and is incremented immediately (`globalDay++` is the first line of each day loop body). After increment it equals the 1-based day number.

**How to avoid:** Always index `e.dayDigests[globalDay-1]`. Verify in pre-population: `digests[globalIdx]` where `globalIdx` starts at 0 and increments after each entry — matches `globalDay-1` at runtime.

**Warning signs:** Index-out-of-bounds panic at day 1 (if index used is `globalDay` not `globalDay-1`) or stale data at wrong position.

### Pitfall 4: dayDigests not reset between PoC runs

**What goes wrong:** If the engine is stopped and restarted for a second PoC run, the old `dayDigests` slice persists. The new pre-population appends to the old slice if `append` is used, or silently overwrites if direct assignment is used correctly.

**Why it happens:** `Start()` resets `executedTechniques = nil` but does not know about new fields unless explicitly added.

**How to avoid:** Add `e.dayDigests = nil` in `Start()` at the same point as `e.executedTechniques = nil`.

**Warning signs:** Second run showing stale day counts from the first run when querying `/api/poc/days`.

### Pitfall 5: Gap days update heartbeat but have no techniques

**What goes wrong:** Gap days have zero techniques, so the per-technique heartbeat path is never triggered. If StartTime and LastHeartbeat are only set when techniques run, gap days will have empty timestamps even though the engine slept through them.

**Why it happens:** The gap loop only calls `waitOrStop(d)` — no technique execution.

**How to avoid:** Set StartTime and LastHeartbeat at gap day-start (same pattern as other phases). Set EndTime and Status=DayComplete immediately after `waitOrStop` returns for gap days (no technique loop needed).

---

## Code Examples

### Retrieving last result success flag safely

```go
// After e.runTechnique(t) returns, results slice has been updated under lock.
e.mu.RLock()
lastResult := e.status.Results[len(e.status.Results)-1]
e.mu.RUnlock()

e.mu.Lock()
if lastResult.Success {
    e.dayDigests[globalDay-1].PassCount++
} else {
    e.dayDigests[globalDay-1].FailCount++
}
e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
e.mu.Unlock()
```

### waitOrStop reuse for campaign delay

```go
// Source: engine.go:678-685 (waitOrStop implementation)
if step.DelayAfter > 0 {
    if !e.waitOrStop(time.Duration(step.DelayAfter) * time.Second) {
        e.abort()
        return
    }
}
```

### Empty-slice guarantee in getter

```go
func (e *Engine) GetDayDigests() []DayDigest {
    e.mu.RLock()
    defer e.mu.RUnlock()
    if len(e.dayDigests) == 0 {
        return []DayDigest{}
    }
    out := make([]DayDigest, len(e.dayDigests))
    copy(out, e.dayDigests)
    return out
}
```

### Registering the route (server.go)

```go
// In registerRoutes(), alongside /api/status:
mux.HandleFunc("/api/poc/days", s.authMiddleware(s.handlePoCDays))
```

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing stdlib (go test) — Go 1.26.1 |
| Config file | None — standard go test discovery |
| Quick run command | `go test ./internal/engine/... -run TestDayDigest -v` |
| Full suite command | `go test ./...` |

All existing tests pass (verified: `go test ./...` green baseline as of 2026-04-09).

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TRACK-01 | DayDigest populated with correct counts after PoC run | unit | `go test ./internal/engine/... -run TestDayDigest_Counts -v` | No — Wave 0 |
| TRACK-02 | DayDigest pre-populated as pending before first day executes | unit | `go test ./internal/engine/... -run TestDayDigest_PrePopulated -v` | No — Wave 0 |
| TRACK-03 | GET /api/poc/days returns [] when idle, full array during/after PoC | unit | `go test ./internal/server/... -run TestHandlePoCDays -v` | No — Wave 0 |
| TRACK-04 | LastHeartbeat updates at day start, per-technique, day end | unit | `go test ./internal/engine/... -run TestDayDigest_Heartbeat -v` | No — Wave 0 |
| CAMP-01 | DelayAfter > 0 triggers waitOrStop; DelayAfter == 0 skips delay | unit | `go test ./internal/engine/... -run TestCampaignDelayAfter -v` | No — Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/engine/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/engine/engine_test.go` — add `TestDayDigest_Counts`, `TestDayDigest_PrePopulated`, `TestDayDigest_Heartbeat`, `TestCampaignDelayAfter` functions (append to existing file — do NOT create a new test file; project convention is one file per package)
- [ ] `internal/server/server_test.go` — check if `TestHandlePoCDays` pattern is supported; existing server test file must be read before adding

---

## Environment Availability

Step 2.6: SKIPPED — this phase is pure Go source changes to existing files. No external tools, services, databases, or CLI utilities required beyond Go 1.26.1 (confirmed present: `go version go1.26.1 windows/amd64`).

---

## Runtime State Inventory

Step 2.5: NOT APPLICABLE — this is a greenfield addition of new structs and fields, not a rename/refactor/migration phase. No stored data, live service config, OS-registered state, secrets, or build artifacts reference DayDigest (it does not yet exist).

---

## Open Questions

1. **Pass/fail count for WhatIf mode**
   - What we know: WhatIf records a synthetic result with `Success: true` always.
   - What's unclear: Should WhatIf runs still increment `PassCount`? The data is synthetic.
   - Recommendation: Yes — increment PassCount. WhatIf mode is a preview; the digest should reflect what would happen, not be empty.

2. **TechniqueCount for Phase 2 at pre-population time**
   - What we know: Phase 2 technique count depends on campaign step count, which is statically knowable at pre-population time (just `len(campaign.Steps)`).
   - What's unclear: The pre-population loop runs before `getTechniquesForCampaign()` is called.
   - Recommendation: At pre-population, look up `e.registry.Campaigns[cfg.CampaignID]` and set `TechniqueCount = len(campaign.Steps)` for Phase 2 days. If no campaign, use `len(e.registry.GetTechniquesByPhase("attack"))`. This makes the digest immediately useful.

3. **Server test file pattern for handlePoCDays**
   - What we know: `internal/server/server_test.go` exists but was not read in this research pass.
   - What's unclear: Whether existing test helpers (httptest.NewRecorder, etc.) are already set up.
   - Recommendation: Read `server_test.go` before writing the Wave 0 server test. The planner should include this as an explicit pre-task read.

---

## Sources

### Primary (HIGH confidence)
- `internal/engine/engine.go` (full read) — all existing patterns, struct layouts, runPoC() implementation, waitOrStop, mutex usage, fakeClock infrastructure
- `internal/server/server.go` (full read) — handler patterns, registerRoutes, authMiddleware, writeJSON/writeError helpers
- `internal/engine/engine_test.go` (full read) — fakeClock, captureClock, newPoCEngine, test helper patterns
- `internal/playbooks/types.go` (full read) — CampaignStep.DelayAfter field confirmed present and typed as `int`
- `.planning/codebase/CONVENTIONS.md` (full read) — naming, struct tags, enum pattern, mutex pattern
- `.planning/codebase/ARCHITECTURE.md` (full read) — engine as state machine, module responsibilities
- `go test ./...` — confirmed green baseline for all packages

### Secondary (MEDIUM confidence)
- `.planning/phases/11-daily-tracking-backend-campaign-delay/11-CONTEXT.md` — locked decisions D-01 through D-11

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies; all patterns verified in existing source
- Architecture patterns: HIGH — all code patterns extracted directly from existing engine.go; no speculation
- Pitfalls: HIGH — identified from direct code inspection (void runTechnique return, globalDay off-by-one, reset omission)
- Campaign delay: HIGH — DelayAfter field confirmed in types.go; waitOrStop confirmed reusable; refactor scope identified precisely

**Research date:** 2026-04-09
**Valid until:** 2026-05-09 (stable — no external library versions involved)
