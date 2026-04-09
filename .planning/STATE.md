---
gsd_state_version: 1.0
milestone: v1.3
milestone_name: Realistic Attack Simulation
current_phase: 14
current_plan: 01 complete
status: In progress
last_updated: "2026-04-09"
progress:
  total_phases: 5
  completed_phases: 0
  total_plans: 0
  completed_plans: 1
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-09)

**Core value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

**Current focus:** v1.3 Realistic Attack Simulation — Phase 14: Safety Audit

## Current Status

**Milestone:** v1.3 — Realistic Attack Simulation
**Current phase:** 14 — Safety Audit
**Current plan:** 01 complete — Plan 02 next
**Last activity:** 2026-04-09 — 14-01 complete: Tier field, defer cleanup, test scaffolds

## Milestone Progress

```
v1.3 [                    ] 0% — 0/5 phases complete
```

## Phase Progress

| Phase | Title | Milestone | Status | Completed |
|-------|-------|-----------|--------|-----------|
| 1 | Events Manifest & Verification Engine | v1.0 | Complete | 2026-03-24 |
| 2 | Code Structure & Test Coverage | v1.0 | Complete | 2026-03-25 |
| 3 | Additional Techniques | v1.0 | Complete | 2026-03-25 |
| 4 | CrowdStrike SIEM Coverage | v1.0 | Complete | 2026-03-25 |
| 5 | Microsoft Sentinel Coverage | v1.0 | Complete | 2026-03-25 |
| 6 | Documentation Consistency | v1.0 | Complete | 2026-03-26 |
| 7 | Nyquist Validation | v1.0 | Complete | 2026-03-26 |
| 8 | Backend Correctness | v1.1 | Complete | 2026-03-26 |
| 9 | UI Polish | v1.1 | Complete | 2026-03-26 |
| 10 | PoC Engine Fixes & Clock Injection | v1.2 | Complete | 2026-04-08 |
| 11 | Daily Tracking Backend & Campaign Delay | v1.2 | Complete | 2026-04-09 |
| 12 | Daily Digest & Timeline Calendar UI | v1.2 | Complete | 2026-04-09 |
| 13 | PoC Scheduling Tests | v1.2 | Complete | 2026-04-09 |
| 14 | Safety Audit | v1.3 | Not started | - |
| 15 | Native Go Architecture | v1.3 | Not started | - |
| 16 | Safety Infrastructure | v1.3 | Not started | - |
| 17 | Network Discovery | v1.3 | Not started | - |
| 18 | Technique Realism | v1.3 | Not started | - |

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

## Decisions (Phase 12)

- Custom JS accordion (classList.toggle open) over details/summary — programmatic open/close required for D-04 auto-expand and D-11 calendar-to-digest link
- hasDayData flag persists panel visibility after PoC completion — panels not gated on isPocRunning
- pollStatus() redeclares pocPhases locally — simpler than module-level; matches existing self-contained function patterns

## Decisions (Phase 13)

- dayCaptureClock captures PoCDay+PoCPhase per After() call — avoids race conditions from polling in fast fake-clock tests
- digestCaptureClock snapshots full []DayDigest slice per After() call for lifecycle transition observation
- stopOnNthClock generalizes blockingClock with configurable block-at-N: fires immediately for calls < N, blocks on call >= N
- No production code changes — all 6 tests verify existing runPoC() behavior via clock injection only

## Decisions (Phase 14 Plan 01)

- Tier field added to Technique struct between NistControls and SIEMCoverage, matching existing aligned struct tag convention
- Tier added to ExecutionResult with json:"tier,omitempty" so HTML report can access it from results
- TestTierClassified scaffold intentionally fails — it is the gate Plan 02 must satisfy by adding tier: N to all 58 YAMLs
- T1059.001 and T1550.002 excluded from writeArtifacts map: no persistent artifacts / inline self-cleanup respectively
- T1070.001 excluded from writeArtifacts in Plan 01 — cleanup added in Plan 02 (custom log channel rewrite per D-06)
- Named return variable 'result' in RunWithCleanup is required for defer closure to write CleanupRun = true to caller's copy

## Roadmap Evolution

- Phase 10 completed: PoC Engine Fixes & Clock Injection (2026-04-08)
- Phase 11 added: Daily Tracking Backend & Campaign Delay (TRACK-01..04, CAMP-01)
- Phase 12 added: Daily Digest & Timeline Calendar UI (DIGEST-01..03, CAL-01..04)
- Phase 13 added: PoC Scheduling Tests (TEST-02..04)
- v1.3 roadmap defined (2026-04-09): Phases 14-18, 16 requirements

## Known Technical Debt

- Windows Defender quarantine of playbooks.test.exe — add `%LOCALAPPDATA%\Temp\go-build*` exclusion before running playbooks tests
- HTML report tactic colors (UI-04): funcMap fix shipped (09-01), but could not be visually verified during UAT — no report was available at verification time
- `/api/techniques` behind authMiddleware — stat box silent in password-protected deployments
- German strings remain in `reporter.go` htmlTemplate (HTML reports)
- Two audit GUIDs need on-machine validation on non-English Windows
- engine.go:165 — `time.Now()` in Start() status init not clock-injected (cosmetic, doesn't affect scheduling)
- engine.go:634 — Pre-existing German WhatIf string (consider full English localization in future milestone)

## Quick Tasks Completed

| # | Description | Date | Commit |
|---|-------------|------|--------|
| 260325-v44 | Translate README.md from German to English | 2026-03-25 | 4ac348a |

---
*Initialized: 2026-03-24*
*v1.0 complete: 2026-03-26*
*v1.1 complete: 2026-03-26*
*v1.2 complete: 2026-04-09*
*v1.3 roadmap defined: 2026-04-09*
*Last session: 2026-04-09 — Completed 14-01-PLAN.md (Tier field, defer cleanup, test scaffolds)*
