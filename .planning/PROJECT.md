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

### Active

<!-- Current scope for next milestone. -->

- ✓ Windows Audit Policy uses locale-independent GUIDs — fixes failure on non-English (e.g. German) Windows — v1.1 (Validated in Phase 08: backend-correctness)
- ✓ Version injected at build time via ldflags — replaces hardcoded v0.1.0 in HTML badge and Go banner — v1.1 (Validated in Phase 08: backend-correctness)
- ✓ Outdated/inconsistent UI labels and placeholder text updated throughout — v1.1 (Validated in Phase 09: ui-polish)
- ✓ Preparation tab surfaces clear, actionable error messages (not raw exit codes) — v1.1 (Validated in Phase 09: ui-polish)
- ✓ Web UI fully in English — all German strings replaced — v1.1 (Validated in Phase 09: ui-polish)
- ✓ Dashboard technique count reflects current 57-technique library — v1.1 (Validated in Phase 09: ui-polish)

### Out of Scope

- SIEM API queries at runtime — adds external dependency, breaks standalone deployment model
- Non-Windows platforms — techniques are fundamentally Windows Event Log / Sysmon based
- Real credential extraction or actual exploitation — tool simulates artifacts only, never real attacks

## Context

- Inspired by Exabeam's internal "Magneto" tool (PowerShell-based, web UI, configurable timing)
- Single Go binary (~10 MB), no runtime dependencies, no internet required at runtime
- Distributed to clients as a standalone .exe for use on Windows 10/11/Server 2016+
- Requires local admin for Attack phase techniques; normal user for Discovery
- README written in German (target user base), translated to English in v1.0
- **v1.0 shipped 2026-03-26:** 57 techniques (43 base + 5 ATT&CK + 4 UEBA + 3 Falcon + 3 Azure), 38k LOC Go, 17 plans across 7 phases
- **Codebase packages:** cmd/lognojutsu, internal/{engine,executor,playbooks,preparation,reporter,server,simlog,userstore,verifier}
- **Test coverage:** 14 test functions across engine_test, server_test, verifier_test, reporter_test, loader_test; playbooks blocked by Windows Defender quarantine
- **Tactic badge colors:** `command-and-control` → red (#f85149), `ueba-scenario` → purple (#bc8cff) — fixed in Phase 09
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

## Current Milestone: v1.1 Bug Fixes & UI Polish

**Goal:** Fix locale-dependent Windows Audit Policy failure and modernise the Web UI with dynamic versioning and visual polish.

**Target features:**
- Windows Audit Policy: locale-independent GUID-based subcategory lookup (fixes German/non-English Windows)
- Dynamic build-time version via ldflags (replaces hardcoded v0.1.0)
- Preparation tab: clear, actionable error messages surfaced to user
- UI labels, placeholder text, and technique counts updated throughout
- Layout and spacing polish across all tabs

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
*Last updated: 2026-03-26 — Phase 09 complete: UI fully in English, inline error panels, live version badge, technique count stat box, tactic badge colors fixed. Milestone v1.1 complete.*
