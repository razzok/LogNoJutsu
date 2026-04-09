# Phase 12: Daily Digest & Timeline Calendar UI - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-09
**Phase:** 12-daily-digest-timeline-calendar-ui
**Areas discussed:** Panel placement & layout, Daily digest presentation, Timeline calendar design, Polling & live updates

---

## Panel Placement & Layout

| Option | Description | Selected |
|--------|-------------|----------|
| Below PoC info panel | Add both panels directly below existing pocInfoPanel on Dashboard tab. Visible when PoC running/has data, hidden when idle. | ✓ |
| New PoC tab | Create dedicated 'PoC Progress' tab alongside Dashboard, Techniques, Log Viewer. | |
| Collapsible section in Dashboard | Collapsible 'PoC Schedule' section to reduce visual noise. | |

**User's choice:** Below PoC info panel
**Notes:** Calendar above digest. Both on Dashboard tab.

### Follow-up: Visibility after completion

| Option | Description | Selected |
|--------|-------------|----------|
| Persist after completion | Panels stay visible after PoC finishes for result review. Hidden only when idle. | ✓ |
| Only while running | Panels disappear when PoC completes. | |
| You decide | Claude picks. | |

**User's choice:** Persist after completion

---

## Daily Digest Presentation

### Display style

| Option | Description | Selected |
|--------|-------------|----------|
| Collapsible accordion | Each day is a row that expands/collapses. Collapsed: day number, phase badge, status, pass/fail counts, time window. | ✓ |
| Flat list (no expand) | Simple list rows with all info visible. No interaction. | |
| Cards | Each day as a small card with phase color border. | |

**User's choice:** Collapsible accordion

### Future days visibility

| Option | Description | Selected |
|--------|-------------|----------|
| Show pending days (collapsed) | All days visible in list. Pending days show day number, phase, 'Pending' status. | |
| Hide pending days | Only show active + completed days. Calendar shows full schedule. | ✓ |
| You decide | Claude picks. | |

**User's choice:** Hide pending days

### Sort order

| Option | Description | Selected |
|--------|-------------|----------|
| Newest first | Active/latest day at top, oldest at bottom. | ✓ |
| Chronological | Day 1 at top, latest at bottom. | |

**User's choice:** Newest first

---

## Timeline Calendar Design

### Layout style

| Option | Description | Selected |
|--------|-------------|----------|
| Horizontal day strip | Single horizontal row of day cells grouped by phase. Compact, fits above digest. | ✓ |
| Calendar grid (7-col weeks) | Traditional 7-column calendar grid. Good for long PoCs. | |
| Segmented progress bar | Horizontal bar divided into segments per day. More visual/compact. | |

**User's choice:** Horizontal day strip

### Cell content

| Option | Description | Selected |
|--------|-------------|----------|
| Day number + technique count tooltip | Cell shows day number with color background. Hover shows technique count and pass/fail. | ✓ |
| Day number + inline badge | Cell shows day number with small technique count badge inside. | |
| Day number only | Minimal cells with just day number and color. | |

**User's choice:** Day number + technique count tooltip

### Click interaction

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, click to navigate | Clicking completed/active day cell scrolls to and expands that day in digest. | ✓ |
| No interaction | Calendar is display-only. | |
| You decide | Claude picks. | |

**User's choice:** Yes, click to navigate

---

## Polling & Live Updates

### Polling approach

| Option | Description | Selected |
|--------|-------------|----------|
| Piggyback on existing pollStatus | Add /api/poc/days fetch inside existing pollStatus() function (2-3s interval). | ✓ |
| Separate polling interval | Independent setInterval at slower rate (5-10s). | |
| You decide | Claude picks. | |

**User's choice:** Piggyback on existing pollStatus

### Animations

| Option | Description | Selected |
|--------|-------------|----------|
| Subtle CSS transitions | Background color fades on status changes. Pure CSS transition property. | ✓ |
| No animation | Status changes instant. Consistent with existing UI. | |
| You decide | Claude picks. | |

**User's choice:** Subtle CSS transitions

---

## Claude's Discretion

- CSS class naming for new panels and day cells
- Whether to use `<details>/<summary>` or custom JS for accordion
- Exact tooltip format
- Panel visibility detection approach
- Day cell wrapping behavior

## Deferred Ideas

None — discussion stayed within phase scope.
