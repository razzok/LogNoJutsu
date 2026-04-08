# Stack Research

**Domain:** SIEM validation tool — PoC mode fix and overhaul (v1.2)
**Researched:** 2026-04-08
**Confidence:** HIGH

## Context

This is a subsequent milestone. The core stack (Go backend, vanilla HTML/JS UI, `gopkg.in/yaml.v3`, embed FS) is already validated and shipped. This document covers ONLY what is needed for the three new capabilities:

1. **Daily digest tracking** — per-day summary of which techniques ran, when, success/failure
2. **Timeline calendar visualization** — visual day-by-day schedule of completed/current/future days
3. **Improved PoC mode feedback** — correct log separators, English `CurrentStep` strings, stale day counter fix

No new runtime Go dependencies are needed. No JavaScript libraries are needed.

---

## Recommended Stack

### Core Technologies (unchanged)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.26.1 (go.mod) | Backend, scheduling, engine | Single binary, no runtime deps on target |
| `gopkg.in/yaml.v3` | v3.0.1 | Playbook YAML parsing | Already in go.mod, unchanged |
| Vanilla HTML/CSS/JS | ES2020 | Single-page UI | No build step, embedded via `//go:embed`, zero client deps |

### New Go Data Structures (no new packages)

All new Go data lives in the existing `engine` package and flows through the existing `/api/status` endpoint via `engine.Status`. The standard library's `encoding/json` and `time` packages handle everything.

**`DayDigest` struct** — the key new type for daily tracking:

```go
// DayDigest records what executed on a single PoC day.
// Added to engine package; serialized through existing engine.Status.
type DayDigest struct {
    DayNumber   int                         `json:"day_number"`   // 1-based across all phases
    PoCPhase    string                      `json:"poc_phase"`    // "phase1", "gap", "phase2"
    Date        string                      `json:"date"`         // "2026-04-08" local date
    TechCount   int                         `json:"tech_count"`
    Succeeded   int                         `json:"succeeded"`
    Failed      int                         `json:"failed"`
    Results     []playbooks.ExecutionResult `json:"results"`
    StartedAt   string                      `json:"started_at"`   // RFC3339
    CompletedAt string                      `json:"completed_at"` // RFC3339, empty if day active
}
```

Add `DayDigests []DayDigest` to `engine.Status`. The field serializes for free via existing `/api/status` — no new API endpoint needed. The field is `omitempty`-safe; it is nil in non-PoC runs.

**`simlog.DayStart()` function** — follows the existing `simlog.Phase()` pattern. Uses the existing `TypePhase` entry type for log file visual consistency. No new `EntryType` constant needed.

### New UI Patterns (vanilla JS — no library)

**Timeline calendar** is a horizontal row of day-cells rendered from `status.day_digests[]`. Uses CSS Grid `grid-template-columns: repeat(N, 1fr)` — the same layout technique already used for `poc-grid` and `stat-grid` in the current `index.html`. Up to 30 cells for a typical 7+3+7 PoC; no external calendar widget is warranted.

**Daily digest accordion** uses the existing `result-item` / `result-header` / `result-body` expand/collapse pattern already in the Results tab. Each `DayDigest` is one accordion row; clicking it expands the per-technique results list for that day.

**Countdown timer** already exists (`pocCountdown` in `updateDashboard()`). No change needed.

---

## Supporting Libraries

No new libraries. All needs are met by existing patterns:

| Capability | Implementation | Why No Library |
|------------|---------------|----------------|
| Day-grid calendar | CSS Grid + `Array.from({length:N})` | 30 cells max, no drag/drop, no time-of-day axis |
| Per-day data storage | `[]DayDigest` slice in engine.Status | stdlib `encoding/json` serializes it through existing `/api/status` |
| Digest accordion | CSS `display:none` toggle (existing result-item pattern) | Already ships in index.html |
| Day log separators | New `simlog.DayStart()` wrapping existing `Phase()` logic | Stays in simlog package, TypePhase entry type |
| Countdown | `setInterval` + `Date` arithmetic (already in `updateDashboard`) | No change needed |

---

## Development Tools (unchanged)

| Tool | Purpose | Notes |
|------|---------|-------|
| `go test ./...` | Unit + integration tests | Add engine_test.go cases for `nextOccurrenceOfHour`, `runPoC` scheduling, and `DayDigest` accumulation |
| `go build -ldflags` | Version injection | Existing pattern, no change |
| Browser DevTools | UI iteration | No bundler or hot-reload used |

---

## Installation

No new packages. `go.mod` stays at:

```
require gopkg.in/yaml.v3 v3.0.1 // indirect
```

For the UI: no `npm install`. All new CSS and JS stays inline in `index.html`.

---

## Alternatives Considered

| Recommended | Alternative | Why Not |
|-------------|-------------|---------|
| `[]DayDigest` in `engine.Status` via `/api/status` poll | New `/api/poc/digest` endpoint | Extra endpoint, handler, test — the existing 2-3s poll already carries all status; no latency benefit for day-level data |
| CSS Grid calendar in vanilla JS | FullCalendar.js, tui-calendar | External JS library breaks single-binary distribution constraint; 200KB+ bundle for a 30-cell grid |
| `simlog.DayStart()` reusing `TypePhase` | New `TypePoCDay` entry type + new CSS class | New entry type requires new CSS `.log-POC_DAY` in index.html; `TypePhase` visual is already correct for day separators |
| Inline `index.html` CSS/JS | Separate static `.css`/`.js` files | Current pattern embeds everything in one file; splitting adds `//go:embed` complexity at no benefit for this scale |
| `DayNumber` counted across all phases (1-based absolute) | Per-phase day counter reset to 1 each phase | Absolute day number maps directly to calendar cell position; avoids ambiguity between "Phase 1 Day 3" vs "Phase 2 Day 3" |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| FullCalendar, tui-calendar, or other JS calendar libs | Breaks single-binary zero-dependency model; .exe must be fully self-contained | CSS Grid + vanilla JS following existing `poc-grid` pattern |
| Chart.js or similar for pass/fail visualization | Same distribution constraint; pass/fail ratio as `3/5` text with colored spans is sufficient for a security tool | Inline `<span>` with existing CSS vars (`--green`, `--red`) |
| New Go packages for scheduling or time math | `nextOccurrenceOfHour()` and `time.Duration` arithmetic already cover all scheduling needs | Extend existing engine logic |
| Persistent storage (SQLite, GORM, files) for digest data | Digest data is only meaningful for the current active run; in-memory is consistent with all other status data | `[]DayDigest` slice on the Engine struct |
| Separate `/api/poc/digest` endpoint | Adds handler code, tests, and complexity — status poll already batches all state | Add `DayDigests` to existing `engine.Status` struct |

---

## Integration Points

| New Component | Integrates With | How |
|--------------|----------------|-----|
| `DayDigest` struct | `engine.Status` | New field `DayDigests []DayDigest`; zero-cost/nil when non-PoC |
| Digest accumulation | `runPoC()` Phase 1 and Phase 2 day loops | After each day's techniques complete, build and append `DayDigest`; gap days append a digest with `TechCount:0` |
| `simlog.DayStart()` | `runPoC()` day loops | Called at start of each day's execution block, parallel to `simlog.Phase()` in standard run mode |
| Calendar UI | `updateDashboard(s)` in JS | Reads `s.day_digests` from existing status poll; renders below `pocInfoPanel` when `s.day_digests` has entries |
| English `CurrentStep` fix | `runPoC()` lines ~351, ~389, ~411 | Direct string replacement — no structural change needed |
| Stale `PoCDay` counter fix | Gap loop in `runPoC()` | `e.status.PoCDay` must track absolute day number across all phases; current code resets implicitly during Gap phase (bug: it is never written in Gap loop, so it stays at the last Phase 1 value) |

---

## Version Compatibility

| Component | Current Version | Constraint |
|-----------|----------------|------------|
| Go | 1.26.1 | No change needed |
| `gopkg.in/yaml.v3` | v3.0.1 | No change needed |
| Target Windows | 10/11/Server 2016+ | CSS Grid supported in all target browsers (Edge/Chrome on Windows) |

---

## Sources

- Codebase read: `internal/engine/engine.go`, `internal/simlog/simlog.go`, `internal/server/server.go`, `internal/server/static/index.html`, `internal/playbooks/types.go`, `go.mod` — HIGH confidence (primary source, direct inspection)
- [CSS Grid calendar layout — js-craft.io](https://www.js-craft.io/blog/four-lines-of-css-to-make-a-calendar-layout-with-css-grid/) — confirms CSS Grid is idiomatic for day-grid layout without libraries; MEDIUM confidence
- [Vanilla JS calendar with CSS Grid — DEV Community](https://dev.to/amitgupta15/create-a-responsive-calendar-with-vanilla-javascript-and-css-grid-35ih) — confirms feasibility of no-library approach for day-indexed grids; MEDIUM confidence

---

*Stack research for: LogNoJutsu v1.2 — PoC Mode Fix & Overhaul*
*Researched: 2026-04-08*
