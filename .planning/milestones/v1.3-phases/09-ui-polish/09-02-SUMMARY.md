---
phase: 09-ui-polish
plan: "02"
subsystem: web-ui
tags: [ui, version-badge, dashboard, api-integration]
dependency_graph:
  requires: []
  provides: [version-badge-live, dashboard-technique-count]
  affects: [internal/server/static/index.html]
tech_stack:
  added: []
  patterns: [api()-helper-fetch, async-await, graceful-degradation]
key_files:
  created: []
  modified:
    - internal/server/static/index.html
decisions:
  - "Call /api/techniques directly on init (not shared with loadScheduler) per D-07 — avoids lazy-tab dependency"
  - "Use bare init block (not DOMContentLoaded) to match existing pattern and avoid ordering issues"
metrics:
  duration_minutes: 5
  completed: "2026-03-26T20:32:00Z"
  tasks_completed: 1
  files_changed: 1
---

# Phase 9 Plan 2: Version Badge Wire-up + Techniques Available Stat Box Summary

Live version badge via `/api/info` fetch on load; Dashboard "Techniques Available" stat box wired to `/api/techniques` array length.

## Tasks Completed

| # | Name | Commit | Files |
|---|------|--------|-------|
| 1 | Add version badge fetch + technique count stat box + init calls | befe33e | internal/server/static/index.html |

## What Was Built

Two new async functions added to `internal/server/static/index.html`:

1. **`loadVersionBadge()`** — Fetches `GET /api/info` on page load and sets `.version-badge` text to `r.version`. If fetch fails, badge remains as-is (`v0.1.0` fallback preserved via HTML).

2. **`loadTechniqueCount()`** — Fetches `GET /api/techniques` on page load, reads `techniques.length`, and sets `document.getElementById('dashAvailable').textContent`. If fetch fails, the em-dash placeholder remains.

A new stat box `id="dashAvailable"` with label "Techniques Available" was inserted as the **first** box in the Dashboard `.stat-grid`, before Current Phase / Techniques Run / Succeeded / Failed — satisfying the order required by D-06.

Both functions are called from the bare `// Init` block (not from a `DOMContentLoaded` listener), matching the existing pattern.

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — both functions wire to live API endpoints. The em-dash `&mdash;` initial value in `dashAvailable` is an intentional placeholder until the API response arrives; it is not a stub.

## Verification

- `go build ./...` — exits 0
- `go vet ./...` — exits 0
- All acceptance criteria confirmed via grep:
  - `id="dashAvailable"` present
  - `Techniques Available` label present
  - `async function loadVersionBadge()` present
  - `api('/api/info')` present
  - `.version-badge` selector present
  - `async function loadTechniqueCount()` present
  - `api('/api/techniques')` present
  - `loadVersionBadge();` in init block
  - `loadTechniqueCount();` in init block
  - No `addEventListener('DOMContentLoaded'` introduced
  - `dashAvailable` appears before `dashTotal` in DOM order

## Self-Check: PASSED
