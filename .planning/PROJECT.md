# LogNoJutsu

## What This Is

LogNoJutsu is a SIEM validation tool that simulates MITRE ATT&CK techniques on Windows, generating the system artifacts (Windows Event Logs, Sysmon events, PowerShell logs) that SIEMs should detect. It is used by security consultants during SIEM onboarding and validation engagements with clients. When a technique runs and the SIEM does not alert, LogNoJutsu helps pinpoint whether the problem is technique execution, log forwarding, parsing, or rule configuration.

The tool ships 59 techniques across MITRE ATT&CK, Exabeam UEBA, CrowdStrike Falcon, and Microsoft Sentinel — with automatic pass/fail verification against the local Windows Event Log, tier-classified realism (29 Tier 1 realistic, 19 Tier 2 partial, 10 Tier 3 stub), native Go execution for network scanning, and safety infrastructure (AMSI detection, elevation gating, scan confirmation).

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
- ✓ Safety infrastructure: AMSI detection, elevation gating, scan confirmation modal with IDS warning — v1.3
- ✓ Network discovery: T1046 TCP/UDP /24 subnet scanner + T1018 ICMP/ARP/nltest/DNS discovery chain — v1.3
- ✓ Technique realism: 4 discovery techniques reclassified, tier distribution finalized at 29/19/10, all TECH requirements closed — v1.3
- ✓ Technique execution distributed across the day with random jitter (not all at scheduled hour) — v1.4
- ✓ Phase 2 batching: 2-3 techniques per slot with jittered delays between batches — v1.4
- ✓ Scheduling test coverage: distributed scheduling correctness documented and tested — v1.4
- ✓ Credentials never exposed on the PowerShell command line — passed via `$env:` + stdin — v1.5
- ✓ EventSpec.Contains enforced by verifier — precise per-technique message-content verification — v1.5
- ✓ Password-protected UI fully functional — static UI Basic-auth challenge propagates creds to all API calls — v1.5
- ✓ Deterministic technique/report ordering — sorted by ID, no map-iteration randomness — v1.5
- ✓ HTML report fully English (DE→EN) — consistent with English-UI decision — v1.5

### Active

<!-- Current scope — to be defined in next milestone. -->

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
- **v1.3 shipped 2026-04-10:** Safety audit + tier classification, native Go executor + LDAP/WMI, AMSI/elevation/scan safety, network discovery, technique realism upgrades — 16 plans across 7 phases (08-09, 14-18)
- **v1.4 shipped 2026-04-11:** Distributed technique scheduling — randomSlotsInWindow() distributes techniques across configurable time windows, Phase 1 one-at-a-time, Phase 2 in batches of 2-3, DayDigest accuracy tests — 4 plans across 2 phases, 14 commits
- **v1.5 in review 2026-09-20 (PR #1, branch `v1.5-polish-hardening`):** Security & polish pass after full-codebase read — credentials via `$env:`+stdin (no command-line leak into Event 4688 / ScriptBlock logs), EventSpec.Contains enforcement, password-protected UI fix (static Basic-auth challenge), deterministic ordering, DE→EN report, tech-debt cleanup (Tier propagation, clock consistency, dead-code removal, go mod tidy, go vet clean). Research in `.planning/v1.5-IMPROVEMENTS-RESEARCH.md`. Single commit; build+vet+tests green.
- **Codebase packages:** cmd/lognojutsu, internal/{engine,executor,native,playbooks,preparation,reporter,server,simlog,userstore,verifier}
- **Test coverage:** 40+ test functions across engine_test, poc_test, server_test, verifier_test, reporter_test, loader_test, registry_test, executor_test, native technique tests
- **Codebase size:** ~8.3k LOC Go (production code)
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
| randomSlotsInWindow helper | Central function for distributing slots across a time window with jitter | ✓ Good — reused for Phase 1 (single) and Phase 2 (batched) scheduling |
| Window config over single hour | Four PoCConfig fields (Phase1/2 WindowStart/End) replace Phase1/2DailyHour | ✓ Good — consultants can constrain scheduling to business hours |
| Secrets via `$env:` + stdin, not `-Command` argv | Passwords/DPAPI blobs on the command line are captured by Event 4688 + ScriptBlock logs and readable by other processes | ✓ Good — removes credential leak and fragile escaping (v1.5) |
| Static UI Basic-auth challenge | SPA sends no auth header; challenging the static page makes the browser cache creds for all same-origin /api/* fetches | ✓ Good — fixes totally-broken password mode with no UI password field (v1.5) |
| QueryFn gains `contains` param | Thread EventSpec.Contains through to Get-WinEvent message filter | ✓ Good — precise per-technique verification, backward compatible (v1.5) |
| Sort GetTechniquesByPhase by ID | Map iteration made run/report order nondeterministic | ✓ Good — reproducible runs (v1.5) |

## Current State

**Latest shipped:** v1.4 PoC Technique Distribution (2026-04-11)
**In review:** v1.5 Polish & Hardening (2026-09-20, PR #1)

v1.5 is a security/polish pass done after a full-codebase read (research in `.planning/v1.5-IMPROVEMENTS-RESEARCH.md`). It removes the plaintext-credential leak onto the PowerShell command line (secrets now via `$env:` + stdin), enforces `EventSpec.Contains` in the verifier, fixes the previously-broken password-protected UI (static Basic-auth challenge), makes technique/report ordering deterministic, translates the HTML report to English, and clears most carried-forward tech debt. Build, `go vet`, and all tests green; verified end-to-end against the built binary. Pending merge of PR #1.

**Known tech debt (carried forward):**
- Two audit GUIDs need on-machine validation on non-English Windows
- T1070.001 absent from TestWriteArtifactsHaveCleanup (test coverage gap, behavior correct)
- `userstore` and `simlog` packages have no test files (DPAPI round-trip, store CRUD untested)
- `DefaultQueryFn` reports query errors as "0 events" (no distinct verify-error status) — see v1.5 research Tier 3

**Resolved in v1.5 (was carried-forward debt):**
- ~~`/api/techniques` silent in password mode~~ — fixed by static Basic-auth challenge (whole UI was broken, not just the count)
- ~~German strings in reporter.go / engine.go WhatIf~~ — translated to English
- ~~`time.Now()` in Start() not clock-injected~~ — now `e.clock.Now()`
- ~~Tier not propagated in elevation-skip/WhatIf results~~ — now propagated
- ~~go-ldap/v3, wmi marked // indirect~~ — `go mod tidy`

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
*Last updated: 2026-09-20 for v1.5 (in review, PR #1)*
