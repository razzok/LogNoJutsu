# Phase 9: UI Polish - Context

**Gathered:** 2026-03-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Web UI changes to `internal/server/static/index.html` and `internal/reporter/reporter.go`:
- Version badge fetches live version from `/api/info` instead of hardcoded `v0.1.0`
- All German text in the UI replaced with English equivalents — complete UI, no exceptions
- Preparation tab error feedback replaced: inline styled error panels instead of `alert()` dialogs
- Dashboard gains a "Techniques Available" stat box loaded from `/api/techniques`
- Tactic badge colors for `command-and-control` and `ueba-scenario` added to `tacticColor` funcMap in reporter.go

No new capabilities — every item is a fix or polish to existing visible UI.

</domain>

<decisions>
## Implementation Decisions

### Error Panels (UI-02)

- **D-01:** Error panels appear **inline below the failing step row** — a styled div expands under the step when it fails, showing the full message. Stays visible until the next run.
- **D-02:** **ALL `alert()` calls are replaced** — not just the prep step. 7 `alert()` calls exist across runPrepStep, user management, and campaign launch. All replaced with inline feedback. No `alert()` remains anywhere in the UI.

### German → English (UI-01)

- **D-03:** The entire UI must be in English — **no German strings anywhere**, including:
  - Static HTML labels and descriptions (Scheduler tab, PoC mode config, step labels, tooltips)
  - JavaScript-generated strings in the PoC schedule renderer (phase names, date/time display)
  - Button labels ("Alle auswählen" → "Select All", "Alle abwählen" → "Deselect All")
- **D-04:** PoC schedule time format: `X days • daily H:00` (replaces `X Tage • tägl. H:00 Uhr`)
- **D-05:** Phase names in JS schedule: "Phase 1: Discovery", "Gap (no actions)", "Phase 2: Attack"

### Dashboard Technique Count (UI-03)

- **D-06:** A new stat box labeled **"Techniques Available"** is added as the **first stat box** in the Dashboard grid — before "Techniques Run", "Succeeded", and "Failed". Order: Available → Run → Succeeded → Failed.
- **D-07:** The count is loaded from `/api/techniques` on page load (same endpoint already called in `loadScheduler` and `loadCampaigns`). Display the array length.

### Version Badge (VER-03)

- **D-08:** On page load, fetch `GET /api/info` and update `.version-badge` text with the returned version. If the fetch fails, leave the badge as-is (graceful degradation — no broken state).

### Tactic Colors (UI-04)

- **D-09:** Add `command-and-control` and `ueba-scenario` to the `tacticColor` funcMap in `internal/reporter/reporter.go`. Colors should follow the existing palette:
  - `command-and-control` → `#f85149` (red — offensive/attack aligned)
  - `ueba-scenario` → `#bc8cff` (purple — matches existing UEBA `.tag-ueba` CSS class in web UI)
  - Claude's Discretion: exact color values can be adjusted to best fit the existing color scheme.

### Claude's Discretion

- Error panel styling (border color, padding, icon) — follow existing `.status-fail` / `--red` pattern
- Which non-prep `alert()` calls become toast notifications vs. inline messages — Claude decides based on context (e.g., user management success → brief inline message or toast; campaign error → inline in the relevant card)
- Exact English phrasing for PoC mode descriptions — natural English, no literal translation needed

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §v1.1 — VER-03, UI-01, UI-02, UI-03, UI-04 are the five pending requirements for this phase

### Codebase
- `internal/server/static/index.html` — Entire web UI (single file, vanilla JS, no build step)
- `internal/reporter/reporter.go` — `tacticColor` funcMap at line ~204 (add missing tactic entries)

No external specs — requirements fully captured in decisions above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `.version-badge` CSS class (line 20): already styled — just update the text content
- `.status-fail` / `.status-ok` / `.status-running` classes (lines 100-102): use for error panel color
- `api()` helper function: used for all fetch calls — use for `/api/info` and `/api/techniques`
- `.stat-box` CSS class (lines 142-148): add new "Techniques Available" stat box using this class

### Established Patterns
- All API calls use `async function` + `await api(url, method, body)` pattern
- CSS custom properties (`--red`, `--green`, `--accent`, etc.) used for all colors — follow this
- Inline styles used for one-off colors in generated HTML (PoC schedule rows)
- `stat-total` class colors the `.val` with `--accent` (blue) — reuse for library count

### Integration Points
- `/api/techniques` already called in `loadScheduler()` and `loadCampaigns()` — can call again on init or share the result
- `/api/info` is the new endpoint (Phase 8 VER-02) — call once on DOMContentLoaded
- `runPrepStep()` at line 1053: the alert is at line 1058 — replace with inline DOM injection below the step row
- 7 `alert()` calls: lines 814, 861, 1058, 1214, 1223, 1231, 1238

</code_context>

<specifics>
## Specific Ideas

- User confirmed: "UI should be completely in English" — no German strings anywhere, including dynamically generated ones
- Error panels: inline below the failing step row, stays visible until next run
- Library stat: "Techniques Available" as first stat box, loaded from `/api/techniques`
- PoC schedule format: `X days • daily H:00`

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 09-ui-polish*
*Context gathered: 2026-03-26*
