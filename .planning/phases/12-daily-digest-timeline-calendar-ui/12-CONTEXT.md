# Phase 12: Daily Digest & Timeline Calendar UI - Context

**Gathered:** 2026-04-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Add two new UI panels to the Dashboard tab for PoC schedule visualization: a timeline calendar (horizontal day strip showing the full schedule) and a daily digest (collapsible accordion of per-day execution summaries). Both panels consume `GET /api/poc/days` (DayDigest array from Phase 11). No backend changes.

</domain>

<decisions>
## Implementation Decisions

### Panel Placement & Layout
- **D-01:** Both panels live on the Dashboard tab, directly below the existing `#pocInfoPanel`. Timeline calendar above, daily digest below.
- **D-02:** Panels are visible whenever DayDigest data exists (PoC running OR completed). Hidden only when idle (empty array from `/api/poc/days`). This differs from `#pocInfoPanel` which hides on completion — digest/calendar persist so users can review results after the run finishes.

### Daily Digest (DIGEST-01, DIGEST-02, DIGEST-03)
- **D-03:** Collapsible accordion style. Each day is a row with expand/collapse toggle. Collapsed view: day number, phase badge, status label, pass/fail counts, time window. Expanded view: technique count, start time, last heartbeat.
- **D-04:** Current (active) day auto-expands. Completed days default to collapsed. (DIGEST-02)
- **D-05:** Only active + completed days shown in the digest list. Pending (future) days are hidden — the timeline calendar above already shows the full schedule.
- **D-06:** Sort order is newest-first (active/latest day at top, oldest at bottom). Most relevant info visible without scrolling.
- **D-07:** Each day entry shows: technique count, pass count, fail count, and execution time window (start → end timestamps). (DIGEST-03)

### Timeline Calendar (CAL-01, CAL-02, CAL-03, CAL-04)
- **D-08:** Horizontal day strip layout — single row of day cells grouped by phase with phase labels above each group. (CAL-01, CAL-04)
- **D-09:** Day cells color-coded by status: green (complete), yellow/accent (active), gray (pending/future), muted (gap days). Reuse existing phase color variables: `--accent` (blue for phase labels), `--red`, `--muted`. (CAL-02)
- **D-10:** Each cell shows day number with color-coded background. Technique count and pass/fail shown in hover tooltip (title attribute). Keeps cells compact. (CAL-03)
- **D-11:** Clicking a completed/active day cell scrolls to and expands that day in the digest panel below. Pending days are non-interactive (not in digest). Links calendar and digest as a cohesive unit.

### Polling & Live Updates
- **D-12:** Piggyback on existing `pollStatus()` loop (2-3s interval). Add a second `fetch('/api/poc/days')` call inside the existing function. Only fetches when PoC is running or has data.
- **D-13:** Subtle CSS transitions on status changes — background-color fades when day transitions (gray→yellow on active, yellow→green on complete). Pure CSS `transition` property, no JS animation library.

### Claude's Discretion
- CSS class naming for new panels and day cells
- Whether to use `<details>/<summary>` or custom JS for accordion expand/collapse
- Exact tooltip format for day cells
- How to detect "has data" for panel visibility (check array length, or check if any day is non-pending)
- Whether to wrap day cells at a certain count or keep a single scrollable row

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope
- `.planning/ROADMAP.md` section "Phase 12" — requirements DIGEST-01..03, CAL-01..04
- `.planning/REQUIREMENTS.md` sections "UI — Daily Digest" and "UI — Timeline Calendar" — detailed requirement descriptions

### Upstream dependency (Phase 11)
- `.planning/phases/11-daily-tracking-backend-campaign-delay/11-CONTEXT.md` — DayDigest struct decisions (D-01..D-11), API contract
- `.planning/phases/11-daily-tracking-backend-campaign-delay/11-01-SUMMARY.md` — DayDigest implementation details
- `.planning/phases/11-daily-tracking-backend-campaign-delay/11-02-SUMMARY.md` — /api/poc/days endpoint details

### UI patterns and conventions
- `.planning/codebase/CONVENTIONS.md` — naming conventions, code style
- `.planning/codebase/STACK.md` — single-file vanilla HTML/CSS/JS frontend, no framework
- `.planning/phases/09-ui-polish/09-CONTEXT.md` — UI polish decisions, dark theme, phase badge colors

### Primary file to modify
- `internal/server/static/index.html` — single-file UI with embedded CSS/JS
- Existing PoC panel: `#pocInfoPanel` (lines 225-242) — placement anchor for new panels
- Existing polling: `pollStatus()` function (line 683) — where to add /api/poc/days fetch
- Existing phase badge CSS: `.phase-poc_phase1`, `.phase-poc_gap`, `.phase-poc_phase2` (lines 40-42)
- Existing color vars: `--accent`, `--red`, `--muted`, `--orange`, `--green` (CSS custom properties)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- Phase badge CSS classes (`.phase-poc_phase1`, `.phase-poc_gap`, `.phase-poc_phase2`) — reuse for day cell phase coloring
- `writeJSON` / `api()` JS helper — existing fetch wrapper for API calls
- `pollStatus()` function — polling loop to piggyback on
- CSS custom properties (`--bg3`, `--border`, `--accent`, `--muted`, etc.) — dark theme vars for consistent styling
- `.poc-schedule` CSS class — existing PoC schedule styling pattern to follow

### Established Patterns
- All UI state is vanilla JS DOM manipulation (getElementById, textContent, style.display)
- No component framework — direct HTML + JS event handlers
- Existing panels use `display:none`/`block` for visibility toggling
- Stats use `.stat-box` class with `.val` and `.lbl` children
- Existing timeline uses `<ul class="timeline">` with `<li>` rows

### Integration Points
- New panels insert after `#pocInfoPanel` in the Dashboard tab HTML
- `pollStatus()` calls `updateDashboard(s)` — new function(s) needed for day digest/calendar updates
- `GET /api/poc/days` returns `[]DayDigest` with fields: day, phase, status, technique_count, pass_count, fail_count, start_time, end_time, last_heartbeat

</code_context>

<specifics>
## Specific Ideas

- Calendar day cells: small squares (~32x32px) with day number centered, background color per status
- Phase labels above day groups: "Phase 1", "Gap", "Phase 2" in smaller text
- Accordion uses chevron icon (▶/▼) for expand/collapse indicator
- Active day in digest should have a subtle pulsing border or accent highlight to draw attention
- Gap days in calendar use the existing `--muted` color (gray) from Phase 9 badge styling

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 12-daily-digest-timeline-calendar-ui*
*Context gathered: 2026-04-09*
