# LogNoJutsu

## What This Is

LogNoJutsu is a SIEM validation tool that simulates MITRE ATT&CK techniques on Windows, generating the system artifacts (Windows Event Logs, Sysmon events, PowerShell logs) that SIEMs should detect. It is used by security consultants during SIEM onboarding and validation engagements with clients. When a technique runs and the SIEM does not alert, LogNoJutsu helps pinpoint whether the problem is technique execution, log forwarding, parsing, or rule configuration.

The v1.0 release ships with automatic pass/fail verification against the local Windows Event Log, a 57-technique library covering MITRE ATT&CK + Exabeam UEBA + CrowdStrike Falcon + Microsoft Sentinel, and an HTML report with per-technique verification columns for each SIEM platform.

## Core Value

Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

- ✓ Web UI with simulation control (start/stop/status) — v0.1
- ✓ Multi-phase simulation (Preparation → Discovery → Attack) — v0.1
- ✓ Multi-user simulation mode — v0.2
- ✓ WhatIf dry-run mode — v0.3
- ✓ HTML report generation — v0.3
- ✓ Tactic filter (run only specific phases) — v0.3
- ✓ Technique-level logging to .log file — v0.1
- ✓ Events manifest: each technique declares expected Windows Event IDs / log sources — v1.0
- ✓ Verification engine: after execution, query local Event Log and report pass/fail per technique — v1.0
- ✓ Enhanced HTML report showing verification results (expected vs. observed events) — v1.0
- ✓ Test coverage: Go unit and integration tests for engine, HTTP handlers, verification logic — v1.0
- ✓ Code structure: refactor package-level globals to struct, split into packages — v1.0
- ✓ Additional MITRE ATT&CK techniques and Exabeam UEBA scenarios — v1.0
- ✓ CrowdStrike SIEM coverage: detection mappings + Falcon-sensor-specific techniques — v1.0
- ✓ Microsoft Sentinel coverage: detection mappings + Azure AD / Sentinel-specific techniques — v1.0
- ✓ Windows Audit Policy uses locale-independent GUIDs — v1.1
- ✓ Version injected at build time via ldflags — v1.1
- ✓ Full English UI — all German strings in index.html replaced — v1.1
- ✓ Preparation tab: inline error panels replace alert() dialogs — v1.1
- ✓ Dashboard technique count wired to live /api/techniques count — v1.1
- ✓ PoC mode bugs fixed: day counter, German strings, missing log separators — v1.2
- ✓ Clock interface injected for deterministic PoC engine testing — v1.2
- ✓ DayDigest per-day execution tracking with heartbeat and campaign delay — v1.2
- ✓ Daily digest panel and timeline calendar in web UI — v1.2
- ✓ 6 deterministic PoC scheduling tests (day counter, stop-signal, DayDigest lifecycle) — v1.2
- ✓ Safety audit: destructive techniques rewritten, all 58 techniques classified Tier 1/2/3, defer-style cleanup — v1.3
- ✓ Native Go architecture: type:go executor dispatch, technique registry, T1482 LDAP + T1057 WMI native implementations — v1.3

### Active

<!-- Current scope for next milestone. -->

- [ ] Real network discovery (ARP/ICMP + TCP port scan on local /24 subnet)
- [ ] Realistic technique execution (real LDAP queries, authentic tool usage instead of PowerShell echo stubs)
- [ ] Expanded technique repertoire with diverse attack patterns
- [ ] Research-driven implementation based on open-source BAS platforms (Atomic Red Team, Caldera, MITRE tooling)

### Out of Scope

- SIEM API queries at runtime — adds external dependency, breaks standalone deployment model
- Non-Windows platforms — techniques are fundamentally Windows Event Log / Sysmon based
- Destructive attacks or exploitation causing actual damage — techniques must generate real artifacts without causing harm or disruption

## Context

- Inspired by Exabeam's internal "Magneto" tool (PowerShell-based, web UI, configurable timing)
- Single Go binary (~10 MB), no runtime dependencies, no internet required at runtime
- Distributed to clients as a standalone .exe for use on Windows 10/11/Server 2016+
- Requires local admin for Attack phase techniques; normal user for Discovery
- README written in German (target user base), translated to English in v1.0
- **v1.0 shipped 2026-03-26:** 57 techniques (43 base + 5 ATT&CK + 4 UEBA + 3 Falcon + 3 Azure), 38k LOC Go, 17 plans across 7 phases
- **v1.1 shipped 2026-03-26:** Locale-independent audit policy (GUID migration), build-time version injection, full English UI, inline error panels, tactic badge colors — 5 plans across 2 phases, 33 commits
- **v1.2 shipped 2026-04-09:** PoC engine fixes, Clock injection, DayDigest tracking, timeline calendar UI, 6 scheduling tests — 6 plans across 4 phases, 37 commits
- **Codebase packages:** cmd/lognojutsu, internal/{engine,executor,playbooks,preparation,reporter,server,simlog,userstore,verifier}
- **Test coverage:** 26 test functions across engine_test, poc_test, server_test, verifier_test, reporter_test, loader_test; playbooks blocked by Windows Defender quarantine
- **Codebase size:** ~56k LOC Go
- Codebase map available at `.planning/codebase/`

## Constraints

- **Platform**: Windows-only — techniques use Windows Event Log, PowerShell, Sysmon
- **Distribution**: Single binary — no installer, no runtime dependencies
- **Security**: Tool must not perform real attacks — simulate artifacts only
- **Language**: Go — existing codebase, no runtime needed on target machines

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go over PowerShell | Single binary distribution, no PS execution policy issues | ✓ Good |
| Local event log verification (not SIEM API) | Keeps tool standalone; SIEM-agnostic approach | ✓ Validated — consultants can run without SIEM API access |
| Server struct refactor | Replaced package-level globals with Server struct + method receivers | ✓ Good — handlers testable via httptest |
| QueryFn injection for verifier | Injectable func for testability without real PowerShell in tests | ✓ Good — all 3 D-11 verifier tests pass cleanly |
| SIEMCoverage map[string][]string | Extensible data layer supporting multiple SIEM platforms per technique | ✓ Good — reused for both CrowdStrike and Sentinel with zero changes |
| Conditional report columns | CS/Sentinel columns appear only when results carry mappings | ✓ Good — clean report for single-SIEM engagements |
| Race detector skip (no CGO/gcc) | TestEngineRace validates mutex discipline structurally; -race requires CGO | ✓ Acceptable — documented in STATE.md and Phase 2 VALIDATION.md |
| EventSpec format for expected_events | Typed struct (Channel+ID+Description) over plain strings | ✓ Good — enables per-event pass/fail reporting in HTML |
| Clock interface injection | Injectable clock for deterministic PoC testing without real sleeps | ✓ Good — enabled 10 fake-clock tests across poc_test.go |
| DayDigest as separate Engine field | Keep /api/status JSON surface clean; dedicated /api/poc/days endpoint | ✓ Good — clean separation of concerns |
| captureClock pattern | Synchronous state capture on each After() call prevents race conditions in fast fake-clock tests | ✓ Good — reused across Phase 10 and 13 tests |
| Custom JS accordion over details/summary | Programmatic open/close needed for auto-expand and calendar-to-digest linking | ✓ Good — D-04 and D-11 features work cleanly |
| Tier classification (1/2/3) | Consultants need instant visibility into which techniques fire realistic vs stub events | ✓ Good — badges in HTML report + web UI, classification doc for reference |
| Defer-style RunWithCleanup | Named return + defer ensures cleanup fires even on panic | ✓ Good — prevents orphaned artifacts on client machines |
| Custom LogNoJutsu-Test channel for T1070.001 | Generates authentic EID 104 without clearing real Security/Application/System logs | ✓ Good — safe for client machines |
| Native Go technique registry | In-process execution via type:go dispatch — no child process for native techniques | ✓ Good — eliminates shell overhead, enables real library calls (LDAP, WMI) |

## Current Milestone: v1.3 Realistic Attack Simulation

**Goal:** Upgrade techniques from obvious test scenarios to realistic attack simulations that generate authentic SIEM artifacts — without causing damage or disruption.

**Target features:**
- Real network discovery (ARP/ICMP + TCP port scan on local /24 subnet)
- Realistic technique execution (real LDAP queries, authentic tool usage instead of PowerShell echo stubs)
- Expanded technique repertoire with diverse attack patterns
- Research-driven implementation based on open-source BAS platforms (Atomic Red Team, Caldera, MITRE tooling)

## Current State

**Latest shipped:** v1.2 PoC Mode Fix & Overhaul (2026-04-09)

v1.2 delivered reliable PoC/Multiday mode with per-day execution tracking, timeline calendar UI, and deterministic scheduling tests. Consultants can now trust the multi-week engagement tool with clear daily feedback.

**Phase 14 complete:** Safety Audit (2026-04-09) — All 58 techniques classified Tier 1/2/3, destructive techniques (T1070.001, T1490) rewritten for safety, defer-style cleanup guarantees, tier badges in HTML report + web UI.

**Phase 15 complete:** Native Go Architecture (2026-04-09) — Thread-safe technique registry, `type: go` executor dispatch (no child process spawning), T1482 LDAP trust discovery and T1057 WMI process discovery as first native Go techniques. Foundation for realistic technique execution in Phases 17-18.

**Known tech debt (carried forward):**
- `/api/techniques` behind authMiddleware — stat box silent in password-protected deployments
- German strings remain in `reporter.go` htmlTemplate (HTML reports)
- Two audit GUIDs need on-machine validation on non-English Windows
- engine.go:165 — `time.Now()` in Start() status init not clock-injected (cosmetic, doesn't affect scheduling)
- engine.go:634 — Pre-existing German WhatIf string (consider full English localization in future milestone)

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-09 after Phase 15 (Native Go Architecture) completed*
