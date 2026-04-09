---
phase: 12-daily-digest-timeline-calendar-ui
plan: "01"
subsystem: frontend/dashboard
tags: [ui, dashboard, calendar, digest, polling, poc]
dependency_graph:
  requires: [Phase 11 /api/poc/days endpoint]
  provides: [dayCalendarPanel, dayDigestPanel, updateDayPanels, pollStatus day integration]
  affects: [internal/server/static/index.html]
tech_stack:
  added: []
  patterns: [innerHTML render pattern (existing), custom JS accordion (.open class toggle), pollStatus piggyback fetch]
key_files:
  created: []
  modified:
    - internal/server/static/index.html
decisions:
  - Custom JS accordion (classList.toggle open) over details/summary — required for programmatic auto-expand (D-04) and calendar-to-digest link (D-11)
  - Panels inserted as .card siblings inside Simulation Status card, below pocInfoPanel — per plan spec D-01
  - hasDayData flag tracks whether days were previously received so panels persist after PoC ends
  - pollStatus() redeclares pocPhases locally — simpler than module-level extraction, matches existing function-local patterns
metrics:
  duration: "~15 minutes"
  completed: "2026-04-09"
  tasks_completed: 1
  tasks_total: 2
  files_modified: 1
---

# Phase 12 Plan 01: Timeline Calendar and Daily Digest Panels Summary

**One-liner:** Horizontal phase-grouped day strip calendar and collapsible daily digest accordion wired to `/api/poc/days` polling inside `pollStatus()`.

## What Was Built

Task 1 added two new panels to the Dashboard tab in `internal/server/static/index.html`:

**Timeline Calendar Panel (`#dayCalendarPanel`):** A horizontal day strip grouped by PoC phase (Phase 1 / Gap / Phase 2). Each day cell is color-coded by status — accent blue for active, green for complete, muted gray for pending, translucent gray for gap days. Cells carry `title` tooltips with technique count and pass/fail. Active and complete cells are clickable and call `focusDayInDigest(dayNum)` to scroll and expand the corresponding digest row.

**Daily Digest Panel (`#dayDigestPanel`):** A collapsible accordion listing active and complete days in newest-first order. Each row shows a phase badge, status label, pass/fail counts, and time window when collapsed. Expanded rows show technique count, start time, and last heartbeat. The active day auto-expands on each render cycle. Expanded state is preserved across polling re-renders via a `wasOpen` Set collected before innerHTML rebuild.

**Polling integration:** `pollStatus()` fetches `/api/poc/days` when `isPocActive || hasDayData`. The `hasDayData` flag persists after PoC completion so panels remain visible.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | c0ec5ba | feat(12-01): add timeline calendar and daily digest panels to dashboard |

## Acceptance Criteria Verification

All acceptance criteria from Task 1 met:

- `id="dayCalendarPanel"` — present
- `id="dayDigestPanel"` — present
- `class="day-strip"` — present
- `class="day-cell` — present
- `class="digest-row` — present
- `class="digest-header"` — present
- `class="digest-body"` — present
- `function renderDayCalendar(days)` — present
- `function renderDayDigest(days)` — present
- `function updateDayPanels(days)` — present
- `function toggleDigestRow(dayNum)` — present
- `function openDigestRow(dayNum)` — present
- `function focusDayInDigest(dayNum)` — present
- `let hasDayData = false` — present
- `api('/api/poc/days')` — present
- `day-group-label` — present
- `scrollIntoView` — present
- `go test ./internal/server/... -v -run TestHandlePoCDays` — PASS
- `go test ./...` — PASS (all packages)

## Task 2: Pending Human Verification

Task 2 is a `checkpoint:human-verify` requiring visual inspection of both panels in a live WhatIf PoC run. Binary has been built (`lognojutsu.exe` in repo root). Awaiting user approval.

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — all data is fetched from the live `/api/poc/days` API. No hardcoded placeholder values flow to rendering.

## Self-Check: PASSED

- `internal/server/static/index.html` — modified and staged (commit c0ec5ba)
- Commit c0ec5ba — confirmed present in git log
- All 17 acceptance criteria — verified via grep
- All Go tests — PASS
