# Milestones

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
