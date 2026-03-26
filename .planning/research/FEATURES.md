# Feature Research

**Domain:** Security tool admin UI — polish/bug-fix milestone for a Go SIEM validation tool
**Researched:** 2026-03-26
**Confidence:** HIGH (current UI code fully reviewed; v1.1 requirements directly from PROJECT.md)

---

## Context

This research covers UI polish improvements for v1.1, not new product features. The core engine
(57 techniques, verification engine, HTML report) shipped in v1.0. The six items in scope are:

1. Windows Audit Policy using locale-independent GUIDs (fixes non-English Windows)
2. Dynamic build-time version via ldflags (replaces hardcoded `v0.1.0`)
3. Preparation tab: clear, actionable error messages (not raw exit codes)
4. UI labels, placeholder text, and technique counts updated throughout
5. Layout and spacing polish across all tabs
6. Dashboard technique count reflecting 57-technique library

Each item is categorised below as Table Stakes, Differentiator, or Anti-Feature to inform roadmap
phase ordering and implementation strategy.

---

## Feature Landscape

### Table Stakes (Users Expect These)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Actionable error messages in Preparation tab | When a prep step fails, consultants need to know *why* (no admin rights, service already running, policy already set, exit code meaning) not just "✗ Failed". Showing raw exit codes is universally considered broken UX. | LOW | Current code: `alert('Step failed:\n' + r.message)` in `runPrepStep()`. `r.message` may already carry the Go-side error string; UI just needs to surface it clearly, add per-step inline error panel, remove the browser `alert()` anti-pattern |
| Version badge reflects actual build | A tool distributed as a .exe to clients must show a real version so clients and consultants can identify which release is installed. Hardcoded `v0.1.0` on line 167 is incorrect for every release after the first. | LOW | Go ldflags pattern: `go build -ldflags "-X main.version=$(git describe --tags)"`. The Go server needs to expose version via `/api/status` or a dedicated `/api/version` endpoint; JS reads it on init and updates the badge |
| UI language consistency | The current UI mixes English and German in the same tab (e.g., "Configure & Run" tab has German labels "Wartezeit vor Phase 1: Discovery" alongside English labels). A consultant presenting to a German client or an English client needs one consistent language, or a clear language choice. | MEDIUM | ~15 German-only strings identified in the Scheduler tab and PoC panel (lines 358–535). Dashboard PoC panel has German strings: "Gesamttage", "Nächste Ausführung", "Aktuelle Phase" (lines 229–239). Decision needed: go all-English or add language toggle (see Anti-Features) |
| Technique count accuracy on Dashboard | The "Expected SIEM Events" table and any stat that implies library size must not mislead. Dashboard Quick Start button says "All Techniques" with no count; technique count is 57 but the table shows 10 hardcoded events. | LOW | The JS already pulls live technique count via `/api/techniques` and populates `diag-disc-count`/`diag-atk-count` on the Scheduler tab. The same data should populate a stat box on Dashboard. The hardcoded SIEM Events table is illustrative — add a note or count it accurately |
| Preparation steps show current status on load | Consultants re-open the tool on subsequent sessions. Prep step statuses show `—` until re-run, even if the system is already configured. Status should reflect system state, not just last-run state. | MEDIUM | Requires a `/api/prepare/status` GET endpoint to check current state of each prep step (e.g., query registry key for PS logging, run `auditpol /get` for audit policy, check if Sysmon service exists). Higher complexity than other items — may be deferred to v1.2 |
| No browser `alert()` for error feedback | Browser `alert()` blocks execution, cannot be styled, looks unprofessional, and is universally avoided in polished tooling. Current code uses `alert()` in 5 places: simulation error, user CRUD, and prep step failure. | LOW | Replace all `alert()` with inline notification component using existing `.alert` CSS classes. The CSS already has `alert-info`, `alert-warn`, `alert-success` — add `alert-error` (red) variant and a toast/banner display function |

### Differentiators (Competitive Advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Prep step inline error panel with remediation hint | Instead of "✗ Failed", show a collapsible panel under the failed step with: (a) exact error output, (b) a one-line human-readable cause, (c) a remediation hint specific to the failure type (e.g., "Run as Administrator", "Sysmon binary not found — check internet access", "auditpol returned exit code 5 — insufficient privileges"). This goes beyond table-stakes error surfacing to make the tool self-documenting during consultant onboarding. | MEDIUM | Go-side: preparation package should return structured error types with Code, Message, Remediation fields. JS-side: render the remediation hint in a styled hint block beneath the step. Pattern from NN/g error guidelines: "every error should suggest a next step" |
| Version badge with build date and commit | Show `v1.1.0 · 2026-03-26 · abc1234` in the header badge (or in a tooltip). Clients can screenshot it for support tickets. Security tools that distribute as binaries without build metadata are harder to audit. | LOW | ldflags can carry `version`, `buildDate`, `commit`. Server exposes all three. Badge tooltip shows full build info. This differentiates from the bare `v0.1.0` string a client currently sees |
| Tactic badge colour completeness | Known cosmetic gap from PROJECT.md: `command-and-control` and `ueba-scenario` tactic badges render grey. Correct colour mapping makes the technique table readable at a glance for consultants explaining coverage to clients. | LOW | Add two entries to `tacticColor` funcMap in Go template. Fix is one-line per missing tactic |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Language toggle (EN/DE switch) | UI is bilingual; some consultants work in German contexts, others English | Doubles string maintenance burden; JS i18n in a single-file SPA requires either a build step or a large inline translation map; deferred language state complicates tests | Pick one language for each string and apply consistently. The tool's README is in German but the UI should be English for broadest adoption and testability. Migrate all German UI strings to English in v1.1 |
| Persist prep step results to disk | Users want status to survive page refresh | Requires a state file or DB for prep results; prep state can go stale if system changes outside the tool; adds complexity disproportionate to a polish milestone | Surface current system state on load (via check endpoint, noted above as Medium complexity) — do not persist |
| Animated / toast notification system | More polished than `alert()` or banner | A full toast system (position, stacking, auto-dismiss timers) is 50+ lines of new JS for marginal gain in an internal tool | Use a single reusable `showNotification(message, type)` function that renders a dismissible banner at the top of the active card. Simple, reusable, consistent with existing `.alert` CSS |
| Light mode toggle | Some users prefer light mode | Requires a full parallel colour token set. The dark theme (`--bg: #0d1117`) is a deliberate product identity choice for a security tool. Adding light mode doubles CSS maintenance. | Not in scope for v1.1. The CSS uses variables consistently so it can be added later with a single token swap if demanded |
| Real-time prep step progress bar | Looks impressive | Prep steps run for 2–15 seconds max; a progress bar adds complexity with no real user value at that timescale | Spinner + "Running…" text (already implemented as `.status-running`) is sufficient |

---

## Feature Dependencies

```
[Actionable error messages]
    └──requires──> [Structured error response from Go preparation package]
                       └──requires──> [PrepResult.Remediation field in API response]

[Dynamic version badge]
    └──requires──> [ldflags wired in build command / Makefile]
    └──requires──> [/api/status or /api/version exposes version string]
    └──requires──> [JS reads version on init and sets badge text]

[Remove alert() anti-pattern]
    └──requires──> [showNotification() helper function]
    └──enhances──> [Actionable error messages] (same display path)

[UI language consistency]
    ──independent──> all other features (string changes only)

[Technique count on Dashboard]
    └──requires──> [/api/techniques already working — it is]
    ──enhances──> [Dashboard accuracy]

[Tactic badge colour fix]
    ──independent──> (one-line Go template change)
```

### Dependency Notes

- **Actionable errors require structured Go response:** The JS currently uses `r.message` as a tooltip (`title` attribute) and calls `alert()`. To show inline error panels with remediation hints, the Go handler must return a richer object. This is a backend change (preparation package) that gates the UI improvement.
- **Version badge requires build pipeline change:** The ldflags approach is well-established (HIGH confidence, multiple official sources). The Makefile or build script must be updated first; the UI change is trivial once the server exposes the value.
- **alert() removal enhances error messaging:** Both features write to the same notification display path. Implement `showNotification()` once; use it for both error message improvements and alert() replacements.
- **Language consistency is independent:** Pure string changes in HTML. Can be done in any order, no API changes needed.

---

## MVP Definition

### v1.1 Launch With (all six items from PROJECT.md are in scope)

- [x] **Dynamic version via ldflags** — prevents clients from running with a permanently stale version badge; low complexity, high trust signal
- [x] **Actionable prep error messages** — current raw exit code display is the most visible UX failure; fixing it requires structured Go response + inline UI panel
- [x] **Remove alert() calls** — universally expected in professional tooling; blocker for polished feel
- [x] **UI language consistency (migrate German strings to English)** — ~15 strings in Scheduler/PoC/Dashboard tabs; low complexity, pure string edits
- [x] **Dashboard technique count accuracy** — stat box showing correct 57-technique count; trivially reads existing API data
- [x] **Tactic badge colour fix** — one-line Go fix; eliminates grey badge cosmetic gap noted in PROJECT.md

### Add After Validation (v1.2 candidates)

- [ ] **Prep step status on page load (system check endpoint)** — `GET /api/prepare/status` checks registry/auditpol/Sysmon on load; Medium complexity; deferred because it requires Windows-side read calls and careful error handling for partial configs
- [ ] **Version badge with build date + commit** — expand ldflags to include date and git hash; Low complexity once ldflags is wired; deferred to keep v1.1 scope tight

### Future Consideration (v2+)

- [ ] **Language toggle (EN/DE)** — requires i18n infrastructure; out of scope until user research confirms need
- [ ] **Light mode** — CSS tokens are in place but doubling the colour set is a design task; defer
- [ ] **Persistent prep state across sessions** — requires state storage strategy; architectural decision for a later milestone

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Dynamic version badge (ldflags) | HIGH — every client engagement starts with "what version is this?" | LOW — build flag + 1 API field + 1 JS line | P1 |
| Actionable prep error messages | HIGH — current UX is broken (exit codes, browser alert) | MEDIUM — Go struct change + inline UI panel | P1 |
| Remove alert() anti-pattern | MEDIUM — polish; currently functional but jarring | LOW — replace 5 call sites with banner function | P1 |
| UI language consistency | MEDIUM — bilingual UI undermines professional impression | LOW — string replacements, no logic changes | P1 |
| Dashboard technique count accuracy | MEDIUM — stat box shows 0 until simulation runs | LOW — read existing API data on load | P1 |
| Tactic badge colour fix | LOW — grey badges are cosmetic, data is correct | LOW — one Go template line per missing tactic | P2 |
| Prep status check on load | HIGH — consultants re-run the tool across sessions | MEDIUM-HIGH — Windows registry/service queries | P3 (v1.2) |

**Priority key:**
- P1: Must have for v1.1 launch
- P2: Should have, fits in v1.1 if schedule allows
- P3: Nice to have, future consideration

---

## Competitor Feature Analysis

| Feature | Exabeam Magneto (inspiration) | Typical SIEM validation tools | LogNoJutsu v1.1 approach |
|---------|------------------------------|-------------------------------|--------------------------|
| Version display | Internal tool, version rarely surfaced to end users | Often absent or buried in About page | Badge in header, every page — correct version via ldflags |
| Error messages in setup steps | PowerShell output shown raw in transcript | Raw CLI output or no UI at all | Structured messages with remediation hints |
| Language | English only | English only | English (migrate v1.1); German README stays |
| Preparation status | No persistent prep UI — run scripts manually | No UI — manual script execution | Visual prep tab with per-step status (v1.0 shipped); system-check on load deferred to v1.2 |
| UI consistency | Dark PowerShell-hosted web UI, minimal polish | Minimal or no web UI | CSS variables in place; apply spacing/label consistency pass |

---

## Sources

- Current UI code: `internal/server/static/index.html` (1321 lines, reviewed in full)
- Project requirements: `.planning/PROJECT.md` — v1.1 Active requirements section
- [Using ldflags to Set Version Information for Go Applications — DigitalOcean](https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications) — HIGH confidence (official DigitalOcean tutorial, stable Go pattern)
- [Error-Message Guidelines — Nielsen Norman Group](https://www.nngroup.com/articles/error-message-guidelines/) — HIGH confidence (authoritative UX research; guideline: every error should suggest a next step)
- [Dark Mode UI Best Practices 2025 — LogRocket](https://blog.logrocket.com/ux-design/dark-mode-ui-design-best-practices-and-examples/) — MEDIUM confidence (verified against multiple sources)
- [Best Practices for UI Error Handling — DevX](https://www.devx.com/web-ui/9-best-practices-and-examples-for-effective-error-handling-in-ui-design/) — MEDIUM confidence (corroborates NN/g findings)

---

*Feature research for: LogNoJutsu v1.1 Bug Fixes & UI Polish*
*Researched: 2026-03-26*
