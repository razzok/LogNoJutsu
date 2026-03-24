# Roadmap: LogNoJutsu

**Created:** 2026-03-24
**Core Value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute.

## Milestone 1 — Verified & Expanded

Make LogNoJutsu trustworthy for professional client delivery: add verification so silent failures are caught, improve code quality so changes are safe, and expand technique and SIEM coverage.

---

### Phase 1 — Events Manifest & Verification Engine

**Goal:** Each technique declares which Event IDs it should produce. After a run, the tool queries the local Windows Event Log and reports pass/fail per technique.

**Requirements:** VERIF-01, VERIF-02, VERIF-03, VERIF-04, VERIF-05

**Key deliverables:**
- `ExpectedEvents` field added to each `Technique` struct (Event IDs + log channel)
- Post-run verification function: queries `windows/system32/winevt` via PowerShell or Win32 API
- Pass/fail status stored in run results
- HTML report updated with verification column (expected events, found events, status)
- All existing techniques populated with their expected Event IDs

**Success criteria:**
- Running a technique and checking "did it produce Event ID 4688?" works end-to-end
- HTML report shows green/red per technique with event details
- A technique that fails to execute shows "not verified" distinct from "events not found"

---

### Phase 2 — Code Structure & Test Coverage

**Goal:** Refactor global state so handlers are testable. Add unit tests for engine, handlers, and verification logic. Split into packages.

**Requirements:** QUAL-01, QUAL-02, QUAL-03, QUAL-04, QUAL-05

**Key deliverables:**
- `AppState` struct replacing package-level globals in server.go
- Code split into packages: `engine/`, `server/`, `techniques/`, `reporter/`, `verifier/`
- `engine_test.go`: state machine tests (start, stop, phase transitions, concurrent access)
- `server_test.go`: HTTP handler tests using httptest
- `verifier_test.go`: unit tests for event log query and pass/fail logic

**Success criteria:**
- `go test ./...` passes with no race conditions (`-race` flag)
- HTTP handlers testable without global state setup
- Engine state transitions covered by tests

---

### Phase 3 — Additional Techniques

**Goal:** Expand technique library with at least 5 new ATT&CK techniques and 3 new Exabeam UEBA scenarios. All new techniques include events manifest entries.

**Requirements:** TECH-01, TECH-02, TECH-03

**Key deliverables:**
- 5+ new MITRE ATT&CK techniques (gaps from current library — T1003, T1059 variants, T1547, etc.)
- 3+ Exabeam UEBA scenarios (unusual process execution, lateral movement indicators, etc.)
- All new techniques populated with `ExpectedEvents` from Phase 1 schema
- README updated with new techniques

**Success criteria:**
- Technique count meaningfully increased
- All new techniques produce verifiable events (Phase 1 verification shows green)
- No regressions in existing techniques

---

### Phase 4 — CrowdStrike SIEM Coverage

**Goal:** Add CrowdStrike Falcon-specific detection mappings and techniques that generate Falcon sensor events. HTML report shows CrowdStrike coverage column.

**Requirements:** CROW-01, CROW-02, CROW-03

**Key deliverables:**
- Research: Falcon sensor event IDs and detection rule names per ATT&CK technique
- `SIEMCoverage` map in technique struct: `{"crowdstrike": ["DetectionName"], "exabeam": [...]}`
- 3+ techniques targeting Falcon sensor logs (EDR process events, network events)
- HTML report: CrowdStrike coverage column showing mapped detection rules
- Documentation section: CrowdStrike-specific setup requirements

**Success criteria:**
- Running LogNoJutsu on a Falcon-protected host generates detectable Falcon events
- Report shows which techniques have CrowdStrike detection mappings
- Client can use report to validate CrowdStrike rule coverage

---

### Phase 5 — Microsoft Sentinel Coverage

**Goal:** Add Sentinel / Azure AD detection mappings and techniques targeting Microsoft Defender / Azure log sources. HTML report shows Sentinel coverage column.

**Requirements:** SENT-01, SENT-02, SENT-03

**Key deliverables:**
- Research: Azure AD / Microsoft Defender / Sentinel analytic rule names per ATT&CK technique
- Sentinel entries in `SIEMCoverage` map (from Phase 4 schema)
- 3+ techniques targeting Azure-specific log sources (AAD sign-in logs, Defender AV events, MDE alerts)
- HTML report: Sentinel coverage column
- Documentation section: Sentinel-specific prerequisites

**Success criteria:**
- Techniques that trigger Azure AD or Defender events are mapped and verified
- Report shows Sentinel detection coverage alongside Exabeam and CrowdStrike
- Client can identify Sentinel rule gaps from the report

---

## Phase Index

| Phase | Title | Requirements | Status |
|-------|-------|-------------|--------|
| 1 | Events Manifest & Verification Engine | VERIF-01–05 | Pending |
| 2 | Code Structure & Test Coverage | QUAL-01–05 | Pending |
| 3 | Additional Techniques | TECH-01–03 | Pending |
| 4 | CrowdStrike SIEM Coverage | CROW-01–03 | Pending |
| 5 | Microsoft Sentinel Coverage | SENT-01–03 | Pending |

---
*Roadmap created: 2026-03-24*
*Last updated: 2026-03-24 after initial definition*
