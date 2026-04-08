---
gsd_state_version: 1.0
milestone: v1.2
milestone_name: PoC Engine Fixes & Clock Injection
status: In Progress
last_updated: "2026-04-08T19:58:27Z"
progress:
  total_phases: 1
  completed_phases: 0
  total_plans: 3
  completed_plans: 1
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-26)

**Core value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

**Current focus:** Planning next milestone

## Current Status

**Milestone:** v1.2 — PoC Engine Fixes & Clock Injection — In Progress (2026-04-08)
**Current phase:** 10 — poc-engine-fixes-clock-injection
**Current plan:** 10-02 (plan 01 complete)

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
| 10 | PoC Engine Fixes & Clock Injection | In Progress | — |

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
*Last session: 2026-04-08 — Stopped at: Completed 10-poc-engine-fixes-clock-injection/10-01-PLAN.md*
