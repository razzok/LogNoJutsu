---
gsd_state_version: 1.0
milestone: v1.3
milestone_name: Realistic Attack Simulation
current_phase: 17
current_plan: Not started
status: Milestone complete
last_updated: "2026-04-10T17:37:54.702Z"
progress:
  total_phases: 6
  completed_phases: 6
  total_plans: 15
  completed_plans: 15
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-09)

**Core value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

**Current focus:** Phase 17 — network-discovery

## Current Status

**Milestone:** v1.2 — PoC Mode Fix & Overhaul — Shipped (2026-04-09)
**Current phase:** 17
**Current plan:** Not started

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
| 11 | Daily Tracking Backend & Campaign Delay | Complete | 2026-04-09 |
| 12 | Daily Digest & Timeline Calendar UI | Complete | 2026-04-09 |
| 13 | PoC Scheduling Tests | Complete | 2026-04-09 |

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

## Decisions (Phase 16 Plan 01)

- AMSI detection only fires for powershell/psh executor types; returns early before verifier to avoid misleading Fail status
- checkIsElevated() split into engine_windows.go (real Windows token API) and engine_other.go (permissive stub) — platform build tags avoid cross-compilation issues
- isAdmin set once at Start() not per-technique — admin status doesn't change mid-run; SetAdmin() test helper mirrors SetRunner() injection pattern

## Decisions (Phase 17 Plan 01)

- localSubnet() duplicated from engine.detectLocalSubnet() to avoid import cycle between internal/engine and internal/native
- scanWorkers=15 as midpoint of D-08 range (10-20) for bounded goroutine concurrency via semaphore channel
- dialTimeout=300ms per research recommendation — covers LAN hosts without excessive delay
- UDP results are best-effort — Windows Firewall suppresses ICMP port-unreachable; noted in output

## Decisions (Phase 17 Plan 02)

- T1018 ICMP fallback: net.ListenPacket("ip4:icmp") error signals non-admin; delegates to tcpAliveCheck() on ports 445/135 — no separate privilege check needed
- strings.Fields ARP parsing: handles varied whitespace from arp -a; net.ParseIP(fields[0]) == nil skips header lines
- nltest fallback: distinguish "binary not found" (exec error) from "non-domain exit" (non-zero exit code) for accurate error messages
- DNS target union: merge ICMP alive hosts and ARP IP entries with seen-map dedup to avoid duplicate lookups

## Decisions (Phase 16 Plan 03)

- AMSI stat boxes shown conditionally (gt 0) — consistent with HasCrowdStrike/HasSentinel pattern
- Elevation-skipped rows use inline opacity:0.6 via JS rowOpacity variable — simpler than CSS class for single property
- verifHtml computed before template literal in loadResults() for clean conditional badge rendering

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
*v1.2 complete: 2026-04-09*
*Last session: 2026-04-10 — Completed 17-02-PLAN.md*
