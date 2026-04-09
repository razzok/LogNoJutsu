# Phase 12: Daily Digest & Timeline Calendar UI - Research

**Researched:** 2026-04-09
**Domain:** Vanilla HTML/CSS/JS single-file UI — accordion, calendar strip, polling integration
**Confidence:** HIGH

## Summary

Phase 12 is a pure frontend change to `internal/server/static/index.html`. No Go code changes are needed — the backend API (`GET /api/poc/days`) was fully delivered in Phase 11. The work is entirely CSS and JavaScript: two new panels inserted after `#pocInfoPanel` on the Dashboard tab, a second `fetch` call piggybacked inside the existing `pollStatus()` loop, and DOM manipulation functions following established patterns in the file.

The codebase has no frontend build step, no framework, and no external JS libraries. All patterns are vanilla DOM manipulation with CSS custom properties. This phase follows the identical pattern as previous UI phases (Phase 9 was the most recent): find the insertion point in the single HTML file, add CSS classes at the top, add HTML markup in the Dashboard tab, and add JS functions that mirror existing patterns like `renderTimeline()` and `updateDashboard()`.

The only non-trivial design question is accordion behavior: whether to use native `<details>/<summary>` or custom JS toggle. Both approaches work, but custom JS (mirroring the existing `.result-item`/`.result-body.open` accordion pattern already in the file) is the better choice for this project — it matches the existing pattern, allows programmatic expand/collapse control needed for auto-expand and click-from-calendar behavior, and keeps behavior consistent.

**Primary recommendation:** One plan, one task — insert CSS + HTML + JS for both panels into index.html, piggybacking on `pollStatus()`. Mirror the existing `.result-item` accordion pattern. No Go files touched.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Panel Placement & Layout**
- D-01: Both panels live on the Dashboard tab, directly below the existing `#pocInfoPanel`. Timeline calendar above, daily digest below.
- D-02: Panels are visible whenever DayDigest data exists (PoC running OR completed). Hidden only when idle (empty array from `/api/poc/days`). This differs from `#pocInfoPanel` which hides on completion — digest/calendar persist so users can review results after the run finishes.

**Daily Digest (DIGEST-01, DIGEST-02, DIGEST-03)**
- D-03: Collapsible accordion style. Each day is a row with expand/collapse toggle. Collapsed view: day number, phase badge, status label, pass/fail counts, time window. Expanded view: technique count, start time, last heartbeat.
- D-04: Current (active) day auto-expands. Completed days default to collapsed. (DIGEST-02)
- D-05: Only active + completed days shown in the digest list. Pending (future) days are hidden — the timeline calendar above already shows the full schedule.
- D-06: Sort order is newest-first (active/latest day at top, oldest at bottom). Most relevant info visible without scrolling.
- D-07: Each day entry shows: technique count, pass count, fail count, and execution time window (start → end timestamps). (DIGEST-03)

**Timeline Calendar (CAL-01, CAL-02, CAL-03, CAL-04)**
- D-08: Horizontal day strip layout — single row of day cells grouped by phase with phase labels above each group. (CAL-01, CAL-04)
- D-09: Day cells color-coded by status: green (complete), yellow/accent (active), gray (pending/future), muted (gap days). Reuse existing phase color variables: `--accent` (blue for phase labels), `--red`, `--muted`. (CAL-02)
- D-10: Each cell shows day number with color-coded background. Technique count and pass/fail shown in hover tooltip (title attribute). Keeps cells compact. (CAL-03)
- D-11: Clicking a completed/active day cell scrolls to and expands that day in the digest panel below. Pending days are non-interactive (not in digest). Links calendar and digest as a cohesive unit.

**Polling & Live Updates**
- D-12: Piggyback on existing `pollStatus()` loop (2-3s interval). Add a second `fetch('/api/poc/days')` call inside the existing function. Only fetches when PoC is running or has data.
- D-13: Subtle CSS transitions on status changes — background-color fades when day transitions (gray→yellow on active, yellow→green on complete). Pure CSS `transition` property, no JS animation library.

### Claude's Discretion
- CSS class naming for new panels and day cells
- Whether to use `<details>/<summary>` or custom JS for accordion expand/collapse
- Exact tooltip format for day cells
- How to detect "has data" for panel visibility (check array length, or check if any day is non-pending)
- Whether to wrap day cells at a certain count or keep a single scrollable row

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DIGEST-01 | User can see a per-day summary panel showing which techniques ran and their results | Accordion panel in Dashboard tab, consuming DayDigest array |
| DIGEST-02 | Current day auto-expands; completed days are collapsed by default | JS auto-expand logic on active day, collapsed default for complete |
| DIGEST-03 | Each day entry shows technique count, pass/fail counts, and execution time window | DayDigest fields: technique_count, pass_count, fail_count, start_time, end_time |
| CAL-01 | User can see a horizontal day-by-day grid showing the full PoC schedule | Horizontal strip of day cells, one per DayDigest entry |
| CAL-02 | Days are color-coded: green (complete), yellow/accent (current), gray (future), muted (gap) | CSS status→color mapping using existing CSS vars |
| CAL-03 | Each day cell shows technique count badge | Day number + technique_count in cell, or in title tooltip per D-10 |
| CAL-04 | Phase labels (Phase 1 / Gap / Phase 2) are visible above day groups | Phase group headers above cell clusters |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Vanilla JS | ES2020 (browser) | DOM manipulation, fetch, event handling | No build step; all prior phases use this |
| CSS custom properties | Browser native | Theming, color vars | Established in this file (`--accent`, `--green`, etc.) |
| HTML5 | Browser native | Markup structure | Single-file UI pattern |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| CSS `transition` property | Browser native | Background-color fade on status change | For D-13 (subtle status transitions) |
| `title` attribute | HTML native | Tooltip for day cell details | For D-10 (technique count tooltip) |
| `scrollIntoView()` | Browser native JS | Scroll digest panel to clicked day | For D-11 (calendar → digest linking) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom JS accordion | `<details>/<summary>` | Native details requires no JS for toggle, but does not support programmatic `open` attribute manipulation from a separate click handler (calendar→digest link requires `element.open = true` + scroll, which does work with details but feels inconsistent with existing `.result-body.open` pattern) |
| `title` tooltip | Custom JS tooltip | title is zero-code, accessible, always works — sufficient for compact cells per D-10 |
| Single scrollable row | CSS `flex-wrap` | For typical PoC schedules (10–30 days), a single row fits. Overflow-x scroll on the container handles edge cases without wrap complexity |

**Installation:** No new packages — this is pure HTML/CSS/JS in `internal/server/static/index.html`.

## Architecture Patterns

### Recommended Project Structure
This phase touches exactly one file:
```
internal/server/static/
└── index.html    # All CSS (lines 8-158), HTML (lines 160-650), JS (lines 651+)
```

### Pattern 1: CSS-first accordion (existing pattern in file)
**What:** `.result-item` accordion used for execution results — header row is always visible, body toggles `display:none`/`block` via `.open` class added/removed in JS.
**When to use:** Any collapsible row where programmatic open/close is needed (required here for D-04 auto-expand and D-11 calendar click).
**Example:**
```css
/* Existing pattern — lines 88-94 of index.html */
.result-item { border: 1px solid var(--border); border-radius: 6px; margin-bottom: 8px; overflow: hidden; }
.result-header { display: flex; align-items: center; gap: 10px; padding: 10px 14px; background: var(--bg3); cursor: pointer; }
.result-body { padding: 12px 14px; background: var(--bg2); display: none; }
.result-body.open { display: block; }
```
```javascript
// Toggle pattern used throughout index.html
header.onclick = () => body.classList.toggle('open');
```

### Pattern 2: Panel visibility toggle (existing pattern)
**What:** Panels use `style.display = 'block'/'none'` controlled by `updateDashboard()`.
**When to use:** Any panel that appears/disappears based on API state.
**Example:**
```javascript
// Existing: pocInfoPanel — lines 715-734
const pocPanel = document.getElementById('pocInfoPanel');
if (isPocRunning) {
  pocPanel.style.display = 'block';
} else {
  pocPanel.style.display = 'none';
}
```
New panels follow the same pattern but with different trigger condition (D-02): show when `days.length > 0`, not just while running.

### Pattern 3: Polling function extension (existing pattern)
**What:** `pollStatus()` is the single polling loop at 2-3s intervals. All live data fetches are added here.
**When to use:** Any data that needs live refresh during PoC execution.
**Example:**
```javascript
// Existing pollStatus — line 683
async function pollStatus() {
  const s = await api('/api/status');
  if (s.error) return;
  updateDashboard(s);
  // New: add fetch for /api/poc/days here
}
```

### Pattern 4: DOM element generation via innerHTML
**What:** `renderTimeline()` builds list HTML via `.map().join('')` and sets `innerHTML`. Used throughout index.html for dynamic content.
**When to use:** Any list/grid that needs full re-render on data change.
**Example:**
```javascript
// Existing renderTimeline — lines 803-812
function renderTimeline(results) {
  const items = results.slice().reverse().map(r => `<li>...</li>`).join('');
  document.getElementById('timeline').innerHTML = `<ul class="timeline">${items}</ul>`;
}
```

### Pattern 5: Phase badge reuse
**What:** `.phase-poc_phase1`, `.phase-poc_gap`, `.phase-poc_phase2` CSS classes exist (lines 40-42) and encode phase identity as background+border color. Reusable for day cell phase tinting.
**When to use:** Anywhere phase identity needs visual encoding.

### Recommended CSS additions (new classes)
```css
/* Timeline calendar */
.day-strip        { display:flex; gap:4px; overflow-x:auto; padding-bottom:4px; }
.day-group        { display:flex; flex-direction:column; gap:4px; }
.day-group-label  { font-size:10px; color:var(--muted); text-transform:uppercase; letter-spacing:.5px; }
.day-cells        { display:flex; gap:4px; }
.day-cell         { width:32px; height:32px; border-radius:4px; display:flex; align-items:center;
                    justify-content:center; font-size:11px; font-weight:600; cursor:default;
                    border:1px solid var(--border); transition:background-color .4s ease; }
.day-cell.active  { background:rgba(88,166,255,0.3); color:var(--accent); border-color:var(--accent); cursor:pointer; }
.day-cell.complete{ background:rgba(63,185,80,0.25); color:var(--green); border-color:var(--green); cursor:pointer; }
.day-cell.pending { background:var(--bg3); color:var(--muted); }
.day-cell.gap     { background:rgba(139,148,158,0.1); color:var(--muted); }

/* Daily digest accordion */
.digest-row       { border:1px solid var(--border); border-radius:6px; margin-bottom:6px; overflow:hidden; }
.digest-row.active-day { border-color:var(--accent); }
.digest-header    { display:flex; align-items:center; gap:10px; padding:10px 14px; background:var(--bg3); cursor:pointer; }
.digest-chevron   { font-size:10px; transition:transform .2s; }
.digest-chevron.open { transform:rotate(90deg); }
.digest-body      { padding:12px 14px; background:var(--bg2); display:none; }
.digest-body.open { display:block; }
```

### Anti-Patterns to Avoid
- **Separate `<details>/<summary>` for digest rows:** The existing `.result-item` accordion pattern is the established convention. Using `<details>` would require mirroring open state in JS for the calendar-click link, creating dual state.
- **Fetching `/api/poc/days` in a separate `setInterval`:** All polling is consolidated in `pollStatus()`. A second interval creates timing skew and divergent state.
- **Storing days in a global JS array without guarding empty:** The API returns `[]` when idle. Always check `days.length > 0` before showing panels and building HTML.
- **Hardcoding pixel sizes that differ from existing cells:** Existing `.stat-box` padding is 14px; day cells at 32x32px match the CONTEXT.md spec and feel native at the existing font-size-14px base.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Status → color mapping | Custom switch/if chain | CSS classes (`.day-cell.active`, `.day-cell.complete`, etc.) | CSS handles all state colors; add class name = add color — no JS color logic |
| Tooltip display | Custom tooltip overlay | `title` attribute | Zero JS, accessible, always visible on hover — sufficient per D-10 |
| Smooth status transitions | JS animation or requestAnimationFrame | CSS `transition: background-color .4s ease` | Pure CSS, no event loop pressure, matches D-13 spec |
| Scroll-to-element | Manual `scrollTop` arithmetic | `element.scrollIntoView({ behavior: 'smooth', block: 'nearest' })` | Browser native, handles all cases |

**Key insight:** The existing file already has all the primitive patterns needed. Phase 12 is composition of existing patterns, not invention of new ones.

## Common Pitfalls

### Pitfall 1: Stale innerHTML wipes event listeners
**What goes wrong:** If `renderDayDigest()` rebuilds the entire accordion HTML on every poll cycle (like `renderTimeline()` does for the execution list), any open/closed accordion state is lost every 2-3 seconds — the user's expanded rows snap closed.
**Why it happens:** `innerHTML` replacement destroys all existing DOM nodes and creates fresh ones; expanded state is not persisted anywhere.
**How to avoid:** On re-render, read which rows are currently expanded (`document.querySelectorAll('.digest-body.open')`), rebuild HTML, then re-apply open state to the same day numbers. OR diff the data and only update changed rows. The simplest safe approach: before rebuilding innerHTML, collect all expanded day numbers; after setting innerHTML, re-open those rows.
**Warning signs:** Accordion collapses on its own during a PoC run every few seconds.

### Pitfall 2: Panel persists after PoC reset without data guard
**What goes wrong:** If panels are shown whenever `days.length > 0` but hidden on `days.length === 0`, and the engine resets `dayDigests` to `[]` on a new run start, the panels flicker. More critically: if the hide logic is only in `updateDayPanels(days)` and that function is only called when the API returns, there's a window where old panel state shows stale data.
**Why it happens:** `/api/poc/days` and `/api/status` are fetched independently; status may show `idle` before the days array empties.
**How to avoid:** Drive panel visibility from `days.length > 0` exclusively (the days array is the source of truth per D-02). Clear panels when days is empty.

### Pitfall 3: `pocInfoPanel` and new panels have different hide conditions
**What goes wrong:** `pocInfoPanel` is hidden when PoC is not running (phase not in pocPhases). New digest/calendar panels must stay visible after completion (D-02). If the implementer copies the pocInfoPanel visibility condition, panels will disappear after PoC completes.
**Why it happens:** D-02 explicitly differs from pocInfoPanel behavior — easy to miss.
**How to avoid:** New panels use `days.length > 0` as show condition, NOT `isPocRunning`. Document this difference explicitly in code comments.

### Pitfall 4: Calendar cell click silently fails if digest row not yet in DOM
**What goes wrong:** Clicking a calendar cell for day N calls `scrollIntoView` and opens the accordion for that day, but if the digest panel hasn't been built yet (or is hidden), the element doesn't exist.
**Why it happens:** D-11 says clicking a completed/active cell scrolls and expands the digest row. But if data has just arrived and DOM hasn't updated, `getElementById('digest-day-' + n)` returns null.
**How to avoid:** In the calendar click handler, guard: `const el = document.getElementById('digest-day-' + n); if (!el) return;`. This is a no-op on the first frame.

### Pitfall 5: D-05 (hide pending days from digest) vs D-02 (show panels when data exists)
**What goes wrong:** If ALL days are pending (first poll after PoC starts), `days.length > 0` is true, panels show, but the digest list is empty (all pending = all hidden per D-05). This is correct per spec — calendar shows the full schedule while digest shows nothing yet. But it looks odd to a reviewer.
**Why it happens:** The two panels have different data filters: calendar shows ALL days, digest shows only active/complete.
**How to avoid:** This behavior is correct per spec. Add a comment in the digest render function. The calendar will immediately show the full schedule (all gray/pending) which is the intended "preview" UX.

### Pitfall 6: Phase label grouping logic
**What goes wrong:** Phase labels (Phase 1 / Gap / Phase 2) need to appear once per group, not once per cell. If grouping logic scans the days array incorrectly, labels repeat or appear mid-group.
**Why it happens:** The DayDigest `phase` field is `"phase1"`, `"gap"`, or `"phase2"` — a string, not an index. Groups are runs of consecutive same-phase days.
**How to avoid:** Group by phase using a reduce or consecutive-run scan:
```javascript
// Build phase groups: [{phase, days: [...]}, ...]
const groups = [];
days.forEach(d => {
  if (!groups.length || groups[groups.length-1].phase !== d.phase) {
    groups.push({ phase: d.phase, days: [] });
  }
  groups[groups.length-1].days.push(d);
});
```

## Code Examples

Verified patterns from the actual codebase:

### Polling extension point (line 683-691 of index.html)
```javascript
// Source: internal/server/static/index.html line 683
async function pollStatus() {
  const s = await api('/api/status');
  if (s.error) return;
  updateDashboard(s);
  // ADD HERE: fetch /api/poc/days and call updateDayPanels(days)
  // Per D-12: only fetch when poc is running or when we already have data
  if (isPocPhaseActive(s.phase) || hasDayData) {
    const days = await api('/api/poc/days');
    if (!days.error) updateDayPanels(days);
  }
}
```

### Phase badge reuse (lines 40-42 of index.html)
```css
/* Source: internal/server/static/index.html */
.phase-poc_phase1 { background: rgba(88,166,255,0.15); color: var(--accent); border: 1px solid var(--accent); }
.phase-poc_gap    { background: rgba(139,148,158,0.15); color: var(--muted); border: 1px solid var(--muted); }
.phase-poc_phase2 { background: rgba(248,81,73,0.15); color: var(--red); border: 1px solid var(--red); }
```

### DayDigest JSON shape (from engine.go — exact field names for JS)
```json
{
  "day": 1,
  "phase": "phase1",
  "status": "active",
  "technique_count": 5,
  "pass_count": 3,
  "fail_count": 1,
  "start_time": "2026-04-09T09:00:00Z",
  "end_time": "",
  "last_heartbeat": "2026-04-09T09:05:00Z"
}
```

### Accordion open/close (mirroring .result-item pattern)
```javascript
// Source: internal/server/static/index.html — pattern from result items
function toggleDigestRow(dayNum) {
  const body = document.getElementById('digest-body-' + dayNum);
  const chevron = document.getElementById('digest-chevron-' + dayNum);
  body.classList.toggle('open');
  chevron.classList.toggle('open');
}
function openDigestRow(dayNum) {
  const body = document.getElementById('digest-body-' + dayNum);
  const chevron = document.getElementById('digest-chevron-' + dayNum);
  body.classList.add('open');
  chevron.classList.add('open');
}
```

### Calendar → digest scroll + expand (D-11)
```javascript
function focusDayInDigest(dayNum) {
  const el = document.getElementById('digest-day-' + dayNum);
  if (!el) return;
  openDigestRow(dayNum);
  el.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}
```

### Timestamp formatting (matching existing renderTimeline pattern)
```javascript
// Source: internal/server/static/index.html line 806
// Existing: (r.start_time||'').replace('T',' ').substring(0,19)
// For digest: format start→end window
function fmtTime(ts) { return ts ? ts.replace('T',' ').substring(0,19) : '—'; }
function fmtWindow(start, end) { return `${fmtTime(start)} → ${fmtTime(end)}`; }
```

### Status label mapping
```javascript
const STATUS_LABEL = { active: '● Active', complete: '✓ Done', pending: '○ Pending' };
const STATUS_COLOR = { active: 'var(--accent)', complete: 'var(--green)', pending: 'var(--muted)' };
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Poll only /api/status | Poll /api/status + /api/poc/days | Phase 12 | Digest and calendar get fresh data every 2-3s |
| pocInfoPanel hides on completion | Digest/calendar persist after completion | Phase 12 (per D-02) | Users can review results after run |

**Deprecated/outdated:**
- None for this phase — Phase 11 just shipped the backend; Phase 12 is the first consumer.

## Open Questions

1. **Polling gate condition for `/api/poc/days`**
   - What we know: D-12 says "only fetch when PoC is running or has data"
   - What's unclear: How to detect "has data" on subsequent renders — needs a module-level JS variable `let hasDayData = false` set to `true` when the first non-empty array is received, reset to `false` when `/api/poc/days` returns `[]`.
   - Recommendation: Declare `let hasDayData = false;` alongside the existing `let pollInterval = null;` module-level variables. Set it in `updateDayPanels()`.

2. **Active-day pulse animation**
   - What we know: CONTEXT.md specifics section mentions "subtle pulsing border or accent highlight" for active day in digest
   - What's unclear: Whether to use CSS `@keyframes` animation or just a static accent border
   - Recommendation: Use static accent border via `.digest-row.active-day { border-color: var(--accent); }` — avoids animation complexity and is consistent with the "subtle" intent. A CSS `@keyframes` pulse is optional.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Build + test | Yes | go1.26.1 windows/amd64 | — |
| Browser (target) | Rendering | N/A — runtime only | Modern Chrome/Edge | — |

**Missing dependencies with no fallback:** None — this is a pure file edit.

**Missing dependencies with fallback:** None.

## Validation Architecture

> `workflow.nyquist_validation` not set to false in .planning/config.json — section included.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` package (stdlib) |
| Config file | none — `go test ./...` convention |
| Quick run command | `go test ./internal/server/... -v -run TestHandlePoCDays` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DIGEST-01 | Digest panel renders from DayDigest data | manual-only | visual inspection at localhost:8080 | N/A |
| DIGEST-02 | Active day auto-expands; complete days collapsed | manual-only | visual inspection | N/A |
| DIGEST-03 | Day entry shows count + pass/fail + time window | manual-only | visual inspection | N/A |
| CAL-01 | Horizontal day grid visible | manual-only | visual inspection | N/A |
| CAL-02 | Day cells color-coded by status | manual-only | visual inspection | N/A |
| CAL-03 | Technique count visible in cell/tooltip | manual-only | visual inspection | N/A |
| CAL-04 | Phase labels above groups | manual-only | visual inspection | N/A |

**All requirements are manual-only** because this phase is entirely frontend UI rendering. There is no DOM testing infrastructure (no jsdom, no Playwright, no headless browser). Backend tests from Phase 11 (`TestHandlePoCDays_idle`, `TestHandlePoCDays_auth`) already cover the API layer and remain green — they are the regression guard for the data source.

**Recommended verification approach:** The planner should include a manual checklist task — start the binary with a short PoC schedule (e.g. 2-day Phase 1, 1-day Gap, 2-day Phase 2 using WhatIf mode), open the Dashboard, and verify each requirement visually.

### Sampling Rate
- **Per task commit:** `go test ./internal/server/... -v -run TestHandlePoCDays` (confirm API still green)
- **Per wave merge:** `go test ./...` (full suite)
- **Phase gate:** Full suite green + manual visual verification before `/gsd:verify-work`

### Wave 0 Gaps
None — existing test infrastructure covers the API layer. No new test files needed for Phase 12 (UI-only requirements are manual-only).

## Sources

### Primary (HIGH confidence)
- `internal/server/static/index.html` — Direct read of existing CSS classes, HTML structure, JS patterns
- `internal/engine/engine.go` lines 100-111 — DayDigest struct, exact JSON field names
- `.planning/phases/12-daily-digest-timeline-calendar-ui/12-CONTEXT.md` — Locked decisions D-01 through D-13
- `.planning/phases/11-daily-tracking-backend-campaign-delay/11-01-SUMMARY.md` — DayDigest implementation details
- `.planning/phases/11-daily-tracking-backend-campaign-delay/11-02-SUMMARY.md` — /api/poc/days endpoint details

### Secondary (MEDIUM confidence)
- `.planning/codebase/CONVENTIONS.md` — Code style, naming patterns (Go-side; JS conventions inferred from index.html)
- `.planning/phases/09-ui-polish/09-CONTEXT.md` — UI polish decisions, dark theme patterns

### Tertiary (LOW confidence)
- None — all findings are from direct code reads, not web search.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new libraries; entire phase is vanilla DOM manipulation of an existing file
- Architecture patterns: HIGH — patterns extracted directly from the target file
- Pitfalls: HIGH — derived from reading existing code and understanding D-02/D-05 distinction
- DayDigest API contract: HIGH — struct read directly from engine.go; endpoint confirmed in Phase 11 summaries

**Research date:** 2026-04-09
**Valid until:** This research is valid indefinitely — there are no external library versions to expire. It is only invalidated if index.html structure changes significantly before Phase 12 executes.
