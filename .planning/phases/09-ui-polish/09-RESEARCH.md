# Phase 9: UI Polish - Research

**Researched:** 2026-03-26
**Domain:** Vanilla JS / HTML UI patching, Go template funcMap
**Confidence:** HIGH

## Summary

Phase 9 is a contained UI-only polish phase with no new backend capabilities. Every
change is a targeted fix to `internal/server/static/index.html` (single-file,
no build step, vanilla JS) and one Go template helper in
`internal/reporter/reporter.go`. The codebase has been fully audited — exact line
numbers and affected strings are documented below so the planner can break this into
precise, non-overlapping tasks.

All decisions are locked in CONTEXT.md. There are no architectural choices to make:
patterns already exist in the file (CSS custom properties, `api()` helper, `stat-box`
class, `.alert` classes) and every fix simply extends existing patterns.

**Primary recommendation:** Three independent tasks (one per file location) are the
right split — (1) reporter.go tactic colors, (2) index.html structural HTML
additions (stat box + version badge wire-up), (3) index.html German → English text +
all `alert()` replacements.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-01:** Error panels appear **inline below the failing step row** — a styled div
expands under the step when it fails, showing the full message. Stays visible until
the next run.

**D-02:** **ALL `alert()` calls are replaced** — not just the prep step. 7 `alert()`
calls exist across runPrepStep, user management, and campaign launch. All replaced
with inline feedback. No `alert()` remains anywhere in the UI.

**D-03:** The entire UI must be in English — **no German strings anywhere**,
including:
- Static HTML labels and descriptions (Scheduler tab, PoC mode config, step labels,
  tooltips)
- JavaScript-generated strings in the PoC schedule renderer (phase names, date/time
  display)
- Button labels ("Alle auswählen" → "Select All", "Alle abwählen" → "Deselect All")

**D-04:** PoC schedule time format: `X days • daily H:00` (replaces
`X Tage • tägl. H:00 Uhr`)

**D-05:** Phase names in JS schedule: "Phase 1: Discovery", "Gap (no actions)",
"Phase 2: Attack"

**D-06:** A new stat box labeled **"Techniques Available"** is added as the **first
stat box** in the Dashboard grid — before "Techniques Run", "Succeeded", and
"Failed". Order: Available → Run → Succeeded → Failed.

**D-07:** The count is loaded from `/api/techniques` on page load (same endpoint
already called in `loadScheduler` and `loadCampaigns`). Display the array length.

**D-08:** On page load, fetch `GET /api/info` and update `.version-badge` text with
the returned version. If the fetch fails, leave the badge as-is (graceful degradation
— no broken state).

**D-09:** Add `command-and-control` and `ueba-scenario` to the `tacticColor` funcMap
in `internal/reporter/reporter.go`:
- `command-and-control` → `#f85149` (red)
- `ueba-scenario` → `#bc8cff` (purple)
- Exact values can be adjusted to best fit the existing color scheme.

### Claude's Discretion

- Error panel styling (border color, padding, icon) — follow existing `.status-fail`
  / `--red` pattern
- Which non-prep `alert()` calls become toast notifications vs. inline messages —
  Claude decides based on context (e.g., user management success → brief inline
  message or toast; campaign error → inline in the relevant card)
- Exact English phrasing for PoC mode descriptions — natural English, no literal
  translation needed

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| VER-03 | Web UI version badge fetches version from `/api/info` on page load — no more hardcoded `v0.1.0` | `/api/info` endpoint verified live (Phase 8 VER-02 complete). `.version-badge` class at line 20. Init sequence at lines 1315–1318. |
| UI-01 | All German strings in Web UI replaced with English equivalents | 40+ German strings catalogued below across static HTML and JS. All in single file. |
| UI-02 | Preparation tab uses inline styled error panels instead of browser `alert()` for step failures | 7 `alert()` call sites catalogued: lines 814, 861, 1058, 1214, 1223, 1231, 1238. `.prep-step` DOM structure documented. `.alert-warn` CSS class available. |
| UI-03 | Dashboard displays total technique library count loaded from `/api/techniques` | `.stat-box`/`.stat-total` CSS classes available. `api()` helper established. Insertion point: line 195 (before existing first stat box). |
| UI-04 | Tactic badges render correct colours for `command-and-control` and `ueba-scenario` tactics | `tacticColor` funcMap at reporter.go line 204. Existing palette verified — gap entries confirmed missing. |
</phase_requirements>

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Vanilla JS | ES2020 (no build) | All UI logic | Established in project — no build step, no npm |
| Go text/template | stdlib | HTML report funcMap | Already in use for tactic colors |
| CSS Custom Properties | CSS3 | Color theming | Entire UI uses `--red`, `--green`, `--accent`, etc. |

No new dependencies. All changes use existing patterns in the codebase.

**Installation:** None required.

---

## Architecture Patterns

### Pattern 1: API Call Pattern (verified from existing code)

All API calls in index.html use the `api()` helper. New calls for `/api/info` and
`/api/techniques` (dashboard stat) MUST follow this pattern.

```javascript
// Source: index.html lines 809-816 (quickStart), 918 (loadCampaignSelect)
async function loadVersionBadge() {
  try {
    const r = await api('/api/info');
    if (r && r.version) {
      document.querySelector('.version-badge').textContent = r.version;
    }
    // If fetch fails, badge stays as-is (D-08: graceful degradation)
  } catch (_) {}
}
```

### Pattern 2: Stat Box HTML (verified from existing code)

The Dashboard stat grid at line 190 uses `.stat-grid` + `.stat-box`. The new
"Techniques Available" box (D-06) uses the same structure as the existing
`stat-total` box. It must be inserted **before** the current first stat box
(line 195).

```html
<!-- Source: index.html lines 141-148 (CSS), lines 195-206 (existing stat boxes) -->
<!-- Insert BEFORE line 195 -->
<div class="stat-box stat-total">
  <div class="val" id="dashAvailable">—</div>
  <div class="lbl">Techniques Available</div>
</div>
```

Populate in init (lines 1315-1318 region):
```javascript
// Call once on startup alongside pollStatus()
async function loadTechniqueCount() {
  const techniques = await api('/api/techniques');
  if (Array.isArray(techniques)) {
    document.getElementById('dashAvailable').textContent = techniques.length;
  }
}
```

### Pattern 3: Inline Error Panel (replacing alert())

The `.prep-step` elements have this DOM structure (lines 290-323):
```
.prep-step
  .prep-step-info
    h3
    p
  div (flex row with status + button)
```

The inline error panel appears after the `.prep-step` div using DOM insertion.
Stays visible until next run (cleared at start of `runPrepStep`).

```javascript
// Source: index.html lines 1053-1065 (runPrepStep + setPrepStatus)
// Error panel uses existing CSS classes
function showPrepError(step, message) {
  // Remove any existing error panel for this step
  const existing = document.getElementById('prep-error-' + step);
  if (existing) existing.remove();

  if (!message) return;

  const panel = document.createElement('div');
  panel.id = 'prep-error-' + step;
  panel.className = 'alert alert-warn';  // or custom styled div with --red
  panel.style.marginTop = '6px';
  panel.textContent = message;

  // Insert after the .prep-step element
  const stepEl = document.querySelector(`[onclick*="runPrepStep('${step}')"]`)
    .closest('.prep-step');
  stepEl.insertAdjacentElement('afterend', panel);
}
```

### Pattern 4: Non-Prep alert() Replacements

The 7 `alert()` sites fall into two categories:

**Simulation start errors (lines 814, 861) — inline in relevant card:**
These are in `quickStart()` and `startSimulation()`. Inline message injected into
the scheduler card area (there is an existing `.alert.alert-info` div at line 530
that can be repurposed, or a new error div injected near the Start button).

**User management (lines 1214, 1223, 1231, 1237, 1238) — inline near the form:**
- Validation error (line 1214): inline below the username input field
- API errors (lines 1223, 1231, 1237): inline in the user card area
- Success/failure toast (line 1238): brief inline status message near the test button

The Users page already has a card structure — inject a `<div id="userFeedback">` at
the top of the user card, show/hide with content as operations complete.

### Pattern 5: tacticColor funcMap Extension (reporter.go)

The `tacticColor` funcMap is at reporter.go lines 204-221. Add two entries to the
existing `colors` map:

```go
// Source: internal/reporter/reporter.go lines 204-221
// Add inside the colors map:
"command-and-control": "#f85149",  // red — attack phase (D-09)
"ueba-scenario":       "#bc8cff",  // purple — matches .tag-ueba in web UI (D-09)
```

The `--purple` CSS variable in index.html is `#bc8cff` (line 12), confirming
`ueba-scenario` → `#bc8cff` is palette-consistent. The `--red` is `#f85149`,
confirming `command-and-control` → `#f85149` is palette-consistent.

### Anti-Patterns to Avoid

- **New JS libraries for inline messages:** Do not introduce any toast library.
  The existing `.alert` CSS classes cover all visual needs.
- **DOMContentLoaded wrapper:** The init at lines 1315-1318 is bare script (no
  event listener). New init calls (version badge, technique count) must be added
  to this same bare init block, not wrapped in an event listener (which fires
  after the bare script, causing ordering issues if any code depends on sequence).
- **Shared `/api/techniques` result:** D-07 says the count is loaded from
  `/api/techniques` — call it directly in the init sequence. Do not try to share
  the array from `loadScheduler()` or `loadCampaigns()` (both are called lazily
  on tab switch, not on init).

---

## German String Inventory (Complete Audit)

All German strings found in `internal/server/static/index.html` by category:

### Static HTML Labels

| Line | German | English Replacement |
|------|--------|---------------------|
| 226 | `PoC Tag` | `PoC Day` |
| 230 | `Gesamttage` | `Total Days` |
| 234 | `Nächste Ausführung` | `Next Execution` |
| 237 | `Aktuelle Phase` | `Current Phase` |
| 352 | `All Techniques (alle Discovery + alle Attack-Techniken)` | `All Techniques (all Discovery + all Attack techniques)` |
| 358 | `Simulations-Modus` | `Simulation Mode` |
| 362 | `Einmalige Simulation (Minuten)` | `One-time simulation (minutes)` |
| 364 | `Mehrtägige Simulation für 4-Wochen PoC` | `Multi-day simulation for 4-week PoC` |
| 390 | `T1 — Wartezeit vor Phase 1: Discovery (Sekunden)` | `T1 — Wait before Phase 1: Discovery (seconds)` |
| 392 | `Zeit zwischen „Start" und dem ersten Discovery-Befehl. Für reale Tests: 600–1800s (10–30 Min.).` | `Time between "Start" and the first Discovery command. For real tests: 600–1800s (10–30 min.).` |
| 395 | `T2 — Pause zwischen Phase 1 und Phase 2 (Sekunden)` | `T2 — Pause between Phase 1 and Phase 2 (seconds)` |
| 397 | `Zeit nach Ende von Phase 1 bis Start der Attack-Phase. UEBA-Tests: 60–300s.` | `Time after end of Phase 1 until Attack phase starts. UEBA tests: 60–300s.` |
| 405 | `ℹ️ Im PoC-Modus läuft LogNoJutsu über Wochen. Phase 1 baut täglich ein paar Discovery-Events auf (Exabeam UEBA-Baseline), dann folgt eine Pause, dann Phase 2 mit täglichen vollen Angriffskampagnen.` | `ℹ️ In PoC mode, LogNoJutsu runs over weeks. Phase 1 builds a few Discovery events each day (Exabeam UEBA baseline), then a gap period, then Phase 2 with daily full attack campaigns.` |
| 413 | `Dauer (Tage)` (Phase 1 panel) | `Duration (days)` |
| 415 | `Empfehlung: 7–14 Tage. Exabeam braucht 2–3 Wochen für UEBA-Baseline.` | `Recommendation: 7–14 days. Exabeam needs 2–3 weeks to build a UEBA baseline.` |
| 418 | `Techniken pro Tag` | `Techniques per day` |
| 423 | `Uhrzeit (Stunde, 0–23)` (Phase 1) | `Time (hour, 0–23)` |
| 425 | `Tageszeit für die tägliche Ausführung. Empfehlung: Bürozeit (8–10 Uhr).` | `Time of day for daily execution. Recommendation: business hours (8–10).` |
| 432 | Gap panel label `Tage` | `Days` |
| 442 | `Dauer (Tage)` (Phase 2 panel) | `Duration (days)` |
| 444 | `Empfehlung: 7–14 Tage. Täglich eine volle Kampagne.` | `Recommendation: 7–14 days. One full campaign per day.` |
| 447 | `Kampagne (täglich)` | `Campaign (daily)` |
| 454 | `Uhrzeit (Stunde, 0–23)` (Phase 2) | `Time (hour, 0–23)` |
| 463 | `📅 Simulationsplan (ab heute)` | `📅 Simulation Schedule (from today)` |
| 472 | `Cleanup nach jeder Technik ausführen (angelegte Artefakte sofort entfernen)` | `Run cleanup after each technique (remove created artifacts immediately)` |
| 474 | `Deaktivieren = alle Artefakte bleiben bis Simulationsende bestehen (End-of-Sim-Cleanup). Sinnvoll um Persistence-Erkennungen über Zeit zu testen.` | `Disable = all artifacts remain until end of simulation (end-of-sim cleanup). Useful for testing persistence detections over time.` |
| 482 | `⚠ WhatIf-Modus (Vorschau — keine echte Ausführung)` | `⚠ WhatIf Mode (Preview — no real execution)` |
| 484 | `Zeigt welche Techniken ausgeführt würden, ohne Befehle wirklich auszuführen. Ideal für Planung, Demos oder erste Überprüfung.` | `Shows which techniques would run without actually executing commands. Ideal for planning, demos, or initial review.` |
| 494 | `🎯 Taktik-Filter (optional)` | `🎯 Tactic Filter (optional)` |
| 495 | `Nur ausgewählte Taktiken ausführen. Keine Auswahl = alle Taktiken. Ausschließen: abgewählte Taktiken werden übersprungen.` | `Run only selected tactics. No selection = all tactics. Exclusions: deselected tactics are skipped.` |
| 500 | `Alle auswählen` (button) | `Select All` |
| 501 | `Alle abwählen` (button) | `Deselect All` |
| 511 | `None — als aktueller Benutzer ausführen` | `None — run as current user` |
| 512 | `Sequential — Profile der Reihe nach durchlaufen` | `Sequential — cycle through profiles in order` |
| 513 | `Random — zufälliges Profil pro Technik` | `Random — random profile per technique` |
| 515 | `Techniken unter anderem Benutzerkontext ausführen → erzeugt Event 4648 (explicit credentials) für Exabeam UEBA.` | `Run techniques under a different user context → generates Event 4648 (explicit credentials) for Exabeam UEBA.` |
| 520 | `Benutzerprofile auswählen (Ctrl/Cmd für Mehrfachauswahl)` | `Select user profiles (Ctrl/Cmd for multi-select)` |
| 531 | `ℹ️ Die Simulation startet erst nach Klick auf Start. Beim Öffnen des Tools läuft nichts automatisch.` | `ℹ️ The simulation only starts after clicking Start. Nothing runs automatically when the tool opens.` |
| 535 | `▶ Simulation starten` (button) | `▶ Start Simulation` |

### JavaScript-Generated Strings (updatePoCSchedule, pollStatus)

| Line | German | English Replacement |
|------|--------|---------------------|
| 1014 | `d.toLocaleDateString('de-DE', ...)` locale | Change to `'en-GB'` or `'en-US'` for English weekday/month names |
| 1019 | `Phase 1 Discovery` in row label | `Phase 1: Discovery` (D-05) |
| 1019 | `${p1Days} Tage • tägl. ${p1Hour}:00 Uhr` | `${p1Days} days • daily ${p1Hour}:00` (D-04) |
| 1025 | `Pause` in row label | `Gap (no actions)` (D-05) |
| 1025 | `${gapDays} Tage • keine Aktionen` | `${gapDays} days • no actions` (D-04) |
| 1031 | `Phase 2 Attack` in row label | `Phase 2: Attack` (D-05) |
| 1031 | `${p2Days} Tage • tägl. ${p2Hour}:00 Uhr` | `${p2Days} days • daily ${p2Hour}:00` (D-04) |
| 1035 | `Ende` in row label | `End` |
| 1035 | `Gesamt: ${totalDays} Tage` | `Total: ${totalDays} days` |
| 714 | `phaseDesc` object values: `'🔍 Phase 1: Discovery-Events aufbauen'` etc. | `'🔍 Phase 1: Building discovery events'`, `'⏸ Gap (no actions)'`, `'⚔️ Phase 2: Attack campaigns'` |
| 723 | `'Jetzt aktiv'` | `'Active now'` |

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Toast notifications | Custom animation system | `.alert` CSS classes (already exist) | Sufficient, consistent with existing UI |
| Error state management | Global error state object | DOM id per step (`prep-error-${step}`) | Simple, already scoped by step id |
| Localization | i18n library | Direct string replacement | Single-file, no dynamic locale switching needed |
| Color palette lookup | New color map | Extend existing `tacticColor` map | One-line addition, same function |

---

## Common Pitfalls

### Pitfall 1: Init Sequence — Bare Script vs. DOMContentLoaded

**What goes wrong:** Adding `document.addEventListener('DOMContentLoaded', ...)` for
new init calls when the existing init is bare script (lines 1315-1318). Both work,
but mixing them can cause subtle ordering issues if any init function references DOM
elements populated by the bare script init.

**Why it happens:** Reflex to use `DOMContentLoaded` for "safety."

**How to avoid:** Add `loadVersionBadge()` and `loadTechniqueCount()` calls directly
in the existing bare init block (lines 1315-1318), alongside the existing
`pollStatus()` and `loadUsers()` calls.

**Warning signs:** If you see `document.addEventListener` introduced for phase 9 init
logic, this is the pitfall in action.

### Pitfall 2: Prep-Step Error Panel — Selector Brittleness

**What goes wrong:** Using `querySelector('[onclick*="runPrepStep"]')` to find the
containing `.prep-step` element — fragile, breaks if button markup changes.

**How to avoid:** Pass the element reference directly to `runPrepStep(step, btnEl)`,
or add `id="prep-row-powershell"` etc. to each `.prep-step` div in HTML. The latter
is cleaner and uses a stable id.

**Recommended approach:** Add `id="prep-row-${step}"` to each `.prep-step` wrapper
div in the HTML, then in JS: `document.getElementById('prep-row-' + step)`.

### Pitfall 3: Stat Box Order — Prepending vs. Inserting

**What goes wrong:** Using `insertBefore` against the wrong reference node, or
appending to `.stat-grid` (which puts it last, not first as D-06 requires).

**How to avoid:** Target the first child of `#stat-grid` explicitly:
```javascript
// Or do it in static HTML — just add the new stat-box HTML as the first child
```
The cleanest solution is to add the new stat box directly in the static HTML
(the value is populated dynamically, but the element can be present from load).

### Pitfall 4: `/api/techniques` Called Twice on Init

**What goes wrong:** Calling `/api/techniques` once for the dashboard count AND once
again when `loadScheduler()` is called on tab switch — network overhead with 57
techniques, though small.

**Why it matters:** Not a correctness bug but worth noting. The count is loaded once
on init; the full array is re-fetched lazily when needed. This is the right approach
(D-07 says "loaded from `/api/techniques` on page load" — one call, not shared).

### Pitfall 5: German Locale in Date Formatter

**What goes wrong:** Changing `'de-DE'` locale in `updatePoCSchedule()` line 1014
to `'en-US'` produces "Mon, Mar 25" style — acceptable. Forgetting this change means
date strings in the schedule preview stay German even after all other strings are
translated.

**How to avoid:** The `fmt` function at line 1014 must change locale:
`d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })`

---

## Code Examples

### Version Badge Wire-up

```javascript
// Source: index.html existing init block (lines 1315-1318)
// Add loadVersionBadge() to init:
async function loadVersionBadge() {
  try {
    const r = await api('/api/info');
    if (r && r.version) {
      document.querySelector('.version-badge').textContent = r.version;
    }
  } catch (_) {}
}
```

### Technique Count Stat Box

```javascript
// Source: index.html — new function, called in init block
async function loadTechniqueCount() {
  const techniques = await api('/api/techniques');
  if (Array.isArray(techniques)) {
    const el = document.getElementById('dashAvailable');
    if (el) el.textContent = techniques.length;
  }
}
```

### tacticColor Map Extension

```go
// Source: internal/reporter/reporter.go lines 204-221
// Inside the colors map literal, add:
"command-and-control": "#f85149",
"ueba-scenario":       "#bc8cff",
```

### Prep Step Error Panel

```javascript
// Source: replaces line 1058 (alert call) in runPrepStep()
// Full revised runPrepStep:
async function runPrepStep(step) {
  // Clear any previous error panel
  const prev = document.getElementById('prep-error-' + step);
  if (prev) prev.remove();

  setPrepStatus(step, 'running', '⏳...');
  const r = await api('/api/prepare/step', 'POST', { step });
  setPrepStatus(step, r.success ? 'ok' : 'fail', r.success ? '✓ OK' : '✗ Failed');
  document.getElementById('status-' + step).title = r.message || '';

  if (!r.success && r.message) {
    const panel = document.createElement('div');
    panel.id = 'prep-error-' + step;
    // Uses existing .alert-warn styling (line 154)
    panel.style.cssText = 'margin-top:6px;padding:10px 14px;border-radius:6px;background:rgba(248,81,73,0.1);border:1px solid rgba(248,81,73,0.3);color:var(--red);font-size:13px;';
    panel.textContent = '✗ ' + r.message;
    document.getElementById('prep-row-' + step).insertAdjacentElement('afterend', panel);
  }
}
```

Requires adding `id="prep-row-powershell"`, `id="prep-row-auditpol"`,
`id="prep-row-sysmon"` to the three `.prep-step` wrapper divs in HTML.

---

## Environment Availability

Step 2.6: SKIPPED (no external dependencies — pure HTML/JS/Go changes in existing
codebase with no new tools or services required).

---

## Validation Architecture

`workflow.nyquist_validation` is not set to `false` in config.json (key absent),
so this section is included.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none (standard `go test ./...`) |
| Quick run command | `go test ./internal/reporter/... -v -run TestTacticColor` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VER-03 | Version badge text updated on page load from `/api/info` | manual-only | n/a — browser DOM verification | — |
| UI-01 | No German strings in rendered HTML or JS-generated content | manual-only | n/a — visual/string inspection | — |
| UI-02 | Inline error panel shown on prep step failure; no `alert()` fires | manual-only | n/a — requires UI interaction | — |
| UI-03 | Dashboard "Techniques Available" count matches `/api/techniques` array length | manual-only | n/a — browser DOM verification | — |
| UI-04 | `tacticColor("command-and-control")` returns `#f85149`; `tacticColor("ueba-scenario")` returns `#bc8cff` | unit | `go test ./internal/reporter/... -run TestTacticColor` | ❌ Wave 0 |

**Note on manual-only items:** UI-01 through UI-03 are DOM/visual changes in a
single vanilla JS HTML file with no build output. There is no automated test
infrastructure for the web UI. These are verified by running the server and
inspecting the browser. This is consistent with how the rest of the UI has been
validated in prior phases.

### Sampling Rate
- **Per task commit:** `go build ./... && go vet ./...`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Full suite green + manual browser check of all 5 requirements
  before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/reporter/reporter_test.go` — add `TestTacticColor` unit test covering
  the two new entries (`command-and-control`, `ueba-scenario`) and the existing
  fallback (unknown tactic → `#8b949e`)

*(All other phase requirements are manual-only — no new test files needed for them)*

---

## Open Questions

1. **`alert()` at line 1238 — success vs failure message for `testUser`**
   - What we know: `alert(r.success ? '✓ ${r.message}' : '✗ ${r.message}')` — used
     to show both success AND failure of user credential test
   - What's unclear: Where to put inline success message in the Users tab — no
     obvious anchor element was found in the visible HTML
   - Recommendation: Claude's Discretion (per CONTEXT.md) — add a
     `<div id="userFeedback">` at the top of the users card, show/hide with
     appropriate message, auto-clear after 5 seconds for success messages

2. **`confirm()` at line 1229 — `deleteUser`**
   - What we know: `if (!confirm('Delete user profile "${name}"?')) return;` — this
     is a browser `confirm()` dialog, not an `alert()`. CONTEXT.md D-02 says "all
     `alert()` calls replaced" but does not mention `confirm()`.
   - Recommendation: Leave `confirm()` as-is — it is not an `alert()` and serves a
     legitimate UX purpose (destructive action confirmation). Only replace if user
     explicitly requests it.

---

## Sources

### Primary (HIGH confidence)

- Direct file read: `internal/server/static/index.html` — full audit of German
  strings, alert() calls, DOM structure, CSS classes, init sequence
- Direct file read: `internal/reporter/reporter.go` lines 180-236 — tacticColor
  funcMap, exact line locations, existing color palette
- Direct file read: `.planning/phases/09-ui-polish/09-CONTEXT.md` — all locked
  decisions, reusable assets, integration points
- Direct file read: `.planning/REQUIREMENTS.md` — requirement definitions for
  VER-03, UI-01, UI-02, UI-03, UI-04

### Secondary (MEDIUM confidence)

None — all findings are from direct codebase inspection.

### Tertiary (LOW confidence)

None.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — single file, no build, zero dependencies
- Architecture: HIGH — all patterns verified from existing code line numbers
- Pitfalls: HIGH — all pitfalls derived from direct code inspection
- German string inventory: HIGH — exhaustive grep + manual line-by-line audit

**Research date:** 2026-03-26
**Valid until:** This research describes static code — valid indefinitely unless
index.html is modified before planning begins.
