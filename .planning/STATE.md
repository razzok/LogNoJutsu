---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: Executing Phase 05
last_updated: "2026-03-25T20:33:37.616Z"
progress:
  total_phases: 5
  completed_phases: 5
  total_plans: 15
  completed_plans: 15
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-24)

**Core value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

**Current focus:** Phase 05 — microsoft-sentinel-coverage

## Current Status

**Milestone:** 1 — Verified & Expanded
**Active phase:** 02-code-structure-test-coverage
**Last action:** Completed 02-code-structure-test-coverage/02-03-PLAN.md
**Next step:** Phase 02 complete — all 3 plans done
**Last session:** 2026-03-25T20:33:37.610Z

## Phase Progress

| Phase | Title | Status |
|-------|-------|--------|
| 1 | Events Manifest & Verification Engine | In Progress |
| 2 | Code Structure & Test Coverage | Pending |
| 3 | Additional Techniques | Pending |
| 4 | CrowdStrike SIEM Coverage | Pending |
| 5 | Microsoft Sentinel Coverage | Pending |

## Decisions

- QueryFn injection pattern for verifier testability without real PowerShell (Phase 01, Plan 01)
- VerifNotRun used for both WhatIf mode and no-expected-events case (Phase 01, Plan 01)
- Default 3s wait before event log query to allow OS flush (Phase 01, Plan 01)
- [Phase 01]: Used verifStr funcMap helper for typed string comparison in HTML template
- Removed non-queryable entries (proxy/firewall logs) from expected_events during YAML migration (Phase 01, Plan 02)
- T1490 retains contains field on bcdedit 4688 entry to distinguish from vssadmin 4688 (Phase 01, Plan 02)
- [Phase 02]: Server struct holds eng/registry/users/cfg — all HTTP handlers are method receivers, Start remains package-level, main.go unchanged
- [Phase 02]: RunnerFunc nil-default pattern in Engine mirrors verifier QueryFn — no change to New() or production path
- [Phase 02-code-structure-test-coverage]: Package engine (not engine_test) used for engine tests to access unexported filterByTactics
- [Phase 02-code-structure-test-coverage]: D-11 naming convention for verifier tests: TestVerifier_pass/fail/notRun_WhatIf as thin wrapper tests
- [Phase 02-03]: package server (white-box) test enables direct Server struct instantiation without exported constructor
- [Phase 02-03]: TestHandleStatus_running uses slow RunnerFunc + 50ms sleep to verify phase transitions without real execution
- [Phase 02-03]: Race detector requires CGO/gcc absent on this Windows dev machine — tests pass without -race flag
- [Phase 03-additional-techniques]: Used plain string format for expected_events ([]string in types.go) consistent with all 43 existing techniques
- [Phase 03-additional-techniques]: C2 techniques use .invalid TLD (RFC 2606) and 127.0.0.1 loopback only — no real outbound C2 traffic
- [Phase 03-additional-techniques]: TestNewTechniqueCount committed in RED state (partially failing) because plan 03-01's ATT&CK files are Wave 1 parallel — test passes at wave-end when all 9 files are present
- [Phase 03-additional-techniques]: UEBA-LATERAL-NEW-ASSET covers network-based first-time asset access (SMB+RDP probe), distinct from UEBA-LATERAL-CHAIN which covers enumeration speed burst
- [Phase 03-additional-techniques]: ATT&CK sections inserted at end of Phase 2: Attack block, UEBA sections at end of UEBA-Szenarien — preserves section organization
- [Phase 04-crowdstrike-siem-coverage]: Single result.SIEMCoverage = t.SIEMCoverage line after if/else block in runTechnique covers all four code paths with one line
- [Phase 04-crowdstrike-siem-coverage]: Discovery/UEBA techniques excluded from siem_coverage — benign enumeration does not trigger Falcon prevention policies
- [Phase 04-crowdstrike-siem-coverage]: FALCON_ technique YAML commands use single-line Add-Type signatures (not here-strings) to avoid YAML @' parse errors
- [Phase 04-crowdstrike-siem-coverage]: FALCON_lsass_access uses GrantedAccess 0x0410 vs T1003.001's 0x1010 — different access mask triggers distinct Falcon ML detection path
- [Phase 04-crowdstrike-siem-coverage]: CSS classes for CrowdStrike (cs-badge/cs-na/cs-list) rendered inside {{if .HasCrowdStrike}} conditional — ensures absent HTML has zero vendor-specific markup
- [Phase 04-crowdstrike-siem-coverage]: siemCoverage funcMap helper used for nil-safe map access in templates, avoiding nil map panic when SIEMCoverage is nil on ExecutionResult
- [Phase 05-microsoft-sentinel-coverage]: T1003.006 uses verified GitHub name 'Non Domain Controller Active Directory Replication' (not shorthand 'Potential DCSync Attack')
- [Phase 05-microsoft-sentinel-coverage]: MEDIUM confidence Sentinel rule names (T1059.001, T1136.001) annotated with inline YAML comments for portal verification
- [Phase 05-microsoft-sentinel-coverage]: siem_coverage.sentinel added as sibling key to crowdstrike using existing map[string][]string design — no schema changes
- [Phase 05-microsoft-sentinel-coverage]: AZURE_ldap_recon uses Anomalous LDAP Activity with MEDIUM confidence — community hunting query only, no built-in Sentinel rule for EID 1644
- [Phase 05-microsoft-sentinel-coverage]: AZURE_dcsync uses LDAP ACL enumeration of DS-Replication GUIDs (not repadmin) to trigger EID 4662 Sentinel rule pattern
- [Phase 05-microsoft-sentinel-coverage]: Sentinel CSS classes (ms-badge/ms-na/ms-list) use #0078D4 as badge background — mirrors cs-badge pattern with vendor-accurate Microsoft blue
- [Phase 05-microsoft-sentinel-coverage]: Column order: Verification | CrowdStrike | Sentinel | Benutzer — Sentinel inserted between CrowdStrike and Benutzer

## Performance Metrics

| Phase | Plan | Duration | Tasks | Files |
|-------|------|----------|-------|-------|
| 01-events-manifest-verification-engine | 01 | 15min | 2 | 5 |
| 01-events-manifest-verification-engine | 02 | 20min | 1 | 43 |
| Phase 02-code-structure-test-coverage P01 | 2min | 2 tasks | 3 files |
| Phase 02-code-structure-test-coverage P02 | 5min | 2 tasks | 2 files |
| Phase 02-code-structure-test-coverage P03 | 3min | 2 tasks | 1 file |
| Phase 03-additional-techniques P01 | 3min | 2 tasks | 5 files |
| Phase 03-additional-techniques P02 | 10min | 2 tasks | 5 files |
| Phase 03-additional-techniques P03 | 3min | 2 tasks | 1 files |
| Phase 04-crowdstrike-siem-coverage P01 | 3min | 2 tasks | 13 files |
| Phase 04-crowdstrike-siem-coverage P02 | 4min | 2 tasks | 4 files |
| Phase 04-crowdstrike-siem-coverage P03 | 15min | 2 tasks | 2 files |
| Phase 05-microsoft-sentinel-coverage P01 | 10min | 1 tasks | 6 files |
| Phase 05-microsoft-sentinel-coverage P02 | 4min | 1 tasks | 4 files |
| Phase 05-microsoft-sentinel-coverage P03 | 10min | 2 tasks | 3 files |

## Codebase Map

Available at `.planning/codebase/` — generated 2026-03-24.
Key finding: zero test files; package-level globals in server.go block clean testing.

---
*Initialized: 2026-03-24*
*Updated: 2026-03-24 after plan 01-01 completion*
