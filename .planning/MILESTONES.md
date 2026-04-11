# Milestones

## v1.4 PoC Technique Distribution (Shipped: 2026-04-11)

**Phases completed:** 2 phases, 4 plans, 6 tasks

**Key accomplishments:**

- Three t.Skip stub test functions in poc_schedule_test.go provide automated verify targets for distributed technique scheduling (POC-01/02/03)
- runPoC() rewritten: Phase 1 distributes one technique per random slot, Phase 2 batches 2-3 techniques across window-bounded random slots using randomSlotsInWindow helper
- PoC form updated: Phase 1 and Phase 2 now show window start/end hour inputs (08:00-17:00 default) instead of single hour inputs; config payload and schedule preview updated to match

---

## v1.3 Realistic Attack Simulation (Shipped: 2026-04-10)

**Phases completed:** 7 phases, 16 plans, 25 tasks

**Key accomplishments:**

- GUID-based auditpol calls replacing 12 locale-dependent English names with 11 stable Microsoft GUIDs, plus human-readable error messages using description field
- Build-time version injection via ldflags and public GET /api/info endpoint returning `{"version":"..."}` without authentication
- 1. [Rule 1 - Bug] TDD RED phase not achievable due to CSS false positives
- Full English UI with zero German strings and all 7 alert() dialogs replaced by inline styled error/success panels
- T1070.001 rewritten to use a safe LogNoJutsu-Test custom channel (no real log destruction); T1490 scope-limited by removing vssadmin/wmic/wbadmin; 7 unverified techniques audited and confirmed read-only or self-cleaning
- One-liner:
- 1. [Rule 1 - Bug] WMI struct name mismatch with Win32_Process class
- Channel-based engine pause/resume with API endpoints and web UI modal — consultant must explicitly confirm before network scanning techniques execute
- One-liner:
- T1018 four-method discovery chain in native Go: ICMP ping sweep with RFC 1071 checksum, ARP table parsing, nltest DC discovery, and DNS reverse lookups — all with graceful fallbacks.

---

## v1.2 PoC Mode Fix & Overhaul (Shipped: 2026-04-09)

**Phases completed:** 4 phases, 6 plans, 11 tasks

**Key accomplishments:**

- Four deterministic PoC engine tests using fakeClock and captureClock — validating day counter monotonicity, English-only CurrentStep strings, simlog.Phase separators, and fake clock eliminating real sleeps
- Per-day execution digest with pending pre-population, lifecycle mutations, heartbeat tracking, and interruptible campaign DelayAfter — all backed by 9 new unit tests using existing fakeClock/captureClock infrastructure.
- GET /api/poc/days endpoint wired to engine.GetDayDigests() behind authMiddleware, returning [] when idle and full DayDigest array during a PoC run — with two tests covering idle response and auth enforcement.
- Horizontal phase-grouped day strip calendar and collapsible daily digest accordion wired to `/api/poc/days` polling inside `pollStatus()`.
- 6 deterministic PoC scheduling tests using fake clock injection covering monotonic day counter (TEST-02), stop-signal handling in 4 scenarios (TEST-03), and DayDigest pending->active->complete lifecycle (TEST-04)

---

## v1.0 Verified & Expanded (Shipped: 2026-03-26)

**Phases completed:** 7 phases, 17 plans, 28 tasks

**Key accomplishments:**

- EventSpec struct, VerificationStatus typed constants, VerifiedEvent type, and injectable QueryFn-based verifier package wired into engine post-execution loop via PowerShell Get-WinEvent
- 4 engine tests (phase transitions, stop, tactic filter, race) + 3 D-11 named verifier tests using RunnerFunc and QueryFn injection — all passing
- 6 HTTP handler tests via httptest.NewRecorder against injected Server struct — no global state, covers all D-10 endpoints
- 4 Exabeam UEBA scenario YAMLs (data-staging, account-takeover, priv-esc, lateral-new-asset) plus LoadEmbedded loader tests validating TECH-02/TECH-03 coverage
- SIEMCoverage map[string][]string data layer added to Technique and ExecutionResult structs, propagated through the engine, and 10 existing technique YAMLs populated with official CrowdStrike Falcon prevention policy names
- 3 CrowdStrike Falcon-targeted technique YAML files (process injection, LSASS credential dump, PsExec lateral movement) with authentic Sysmon events and official Falcon detection name mappings
- Conditional CrowdStrike coverage column added to HTML report: CS badge + detection rule names for mapped techniques, grey N/A for unmapped, column entirely absent when no results carry CrowdStrike mappings.
- Conditional Microsoft Sentinel column (blue MS badge, #0078D4) added to HTML report using HasSentinel bool pattern mirroring Phase 4 CrowdStrike implementation; README documents AZURE_ techniques and AMA prerequisites in German.

---
