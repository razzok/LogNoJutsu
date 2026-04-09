---
gsd_state_version: 1.0
milestone: v1.2
milestone_name: PoC Mode Fix & Overhaul
current_phase: 11
current_plan: 2
status: Executing Phase 11
last_updated: "2026-04-09T08:00:00.000Z"
progress:
  total_phases: 4
  completed_phases: 1
  total_plans: 4
  completed_plans: 4
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-26)

**Core value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

**Current focus:** Phase 11 — daily-tracking-backend-campaign-delay

## Current Status

**Milestone:** v1.2 — PoC Mode Fix & Overhaul — In Progress (2026-04-08)
**Current phase:** 11
**Current plan:** 1

## Phase Progress

| Phase | Title | Status | Completed |
|-------|-------|--------|-----------|
| 1 | Events Manifest & Verification Engine | Complete | 2026-03-24 |
| 2 | Code Structure & Test Coverage | Complete | 2026-03-25 |
| 3 | Additional Techniques | Complete | 2026-03-25 |
| 4 | CrowdStrike SIEM Coverage | Complete | 2026-03-25 |
| 5 | Microsoft Sentinel Coverage | Complete | 2026-03-25 |
| 6 | Documentation Consistency | Complete | 2026-03-26 |
| 7 | Nyquist Validation | Complete | 2026-03-26 |
| 8 | Backend Correctness | Complete | 2026-03-26 |
| 9 | UI Polish | Complete | 2026-03-26 |
| 10 | PoC Engine Fixes & Clock Injection | Complete | 2026-04-08 |
| 11 | Daily Tracking Backend & Campaign Delay | In Progress | — |

## Decisions (v1.0)

See `.planning/milestones/v1.0-ROADMAP.md` for full decision log from all phases.

Key decisions carried forward:

- QueryFn injection pattern for verifier testability without real PowerShell
- SIEMCoverage map[string][]string — extensible multi-SIEM data model, reuse for future platforms
- Server struct with method receivers — all handlers testable via httptest
- Race detector requires CGO/gcc — tests pass without -race, documented in VALIDATION.md

## Decisions (v1.1)

- Audit policy commands use GUID format to avoid locale-dependent subcategory name parsing
- /api/info endpoint returns build-time version injected via ldflags (VER-03)
- confirm() for deleteUser preserved — destructive action safeguard, not a disruptive alert
- Inline feedback: showInlineSuccess auto-dismisses after 5s; showInlineError persists until next action

## Decisions (Phase 10)

- Clock interface injected into Engine struct via unexported `clock` field; `realClock{}` default in `New()`; `e.clock.Now()`/`e.clock.After()` replaces `time.Now()`/`time.After()` in `runPoC`, `waitOrStop`, `setPhase`
- `globalDay` counter increments monotonically across Phase1/Gap/Phase2 loops; `nextOccurrenceOfHour` accepts `now` parameter (pure function); `simlog.Phase()` at call sites not inside `setPhase()`
- Clock interface defined inline in engine.go (not clock.go) — single file, minimal surface area
- captureClock pattern for reliable state capture: when fake clock fires too fast for polling goroutines, embed fakeClock in a wrapper that reads engine status synchronously on each After() call

## Decisions (Phase 11)

- DayDigest stored as separate Engine field (not inside Status) to keep /api/status JSON surface clean
- Phase 2 inlines campaign.Steps iteration to preserve DelayAfter metadata discarded by getTechniquesForCampaign()
- TechniqueCount pre-populated from len(campaign.Steps) or len(registry.GetTechniquesByPhase("attack")) at runPoC() start
- GetDayDigests() returns []DayDigest{} (never nil) for safe JSON encoding as [] not null
- handlePoCDays delegates entirely to engine — no nil-check needed; route placed in Simulation API section of registerRoutes()

## Roadmap Evolution

- Phase 10 completed: PoC Engine Fixes & Clock Injection (2026-04-08)
- Phase 11 added: Daily Tracking Backend & Campaign Delay (TRACK-01..04, CAMP-01)
- Phase 12 added: Daily Digest & Timeline Calendar UI (DIGEST-01..03, CAL-01..04)
- Phase 13 added: PoC Scheduling Tests (TEST-02..04)

## Known Technical Debt

- Windows Defender quarantine of playbooks.test.exe — add `%LOCALAPPDATA%\Temp\go-build*` exclusion before running playbooks tests
- HTML report tactic colors (UI-04): funcMap fix shipped (09-01), but could not be visually verified during UAT — no report was available at verification time

## Quick Tasks Completed

| # | Description | Date | Commit |
|---|-------------|------|--------|
| 260325-v44 | Translate README.md from German to English | 2026-03-25 | 4ac348a |

---
*Initialized: 2026-03-24*
*v1.0 complete: 2026-03-26*
*v1.1 complete: 2026-03-26*
*Last session: 2026-04-09 — Stopped at: Completed 11-daily-tracking-backend-campaign-delay/11-02-PLAN.md*
