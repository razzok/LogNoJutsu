# Requirements: LogNoJutsu

**Defined:** 2026-03-24
**Core Value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

## v1 Requirements

### Verification

- [x] **VERIF-01**: Each technique declares the Windows Event IDs / log sources it is expected to generate
- [x] **VERIF-02**: After simulation run, tool queries local Windows Event Log for expected events
- [x] **VERIF-03**: Each technique reports pass (events found) or fail (events missing) in the run results
- [x] **VERIF-04**: Verification results are included in the HTML report (expected vs. observed, pass/fail per technique)
- [x] **VERIF-05**: Verification distinguishes "technique did not execute" from "technique executed but event not found"

### Quality

- [x] **QUAL-01**: Package-level globals in server.go refactored to a struct (enables testable handlers)
- [x] **QUAL-02**: Codebase split into logical packages (engine, server, techniques, reporter)
- [x] **QUAL-03**: Unit tests for simulation engine state machine (start, stop, phase transitions)
- [x] **QUAL-04**: Unit tests for HTTP handlers covering key API endpoints
- [x] **QUAL-05**: Unit tests for verification logic (event log querying, pass/fail determination)

### Techniques

- [x] **TECH-01**: At least 5 additional MITRE ATT&CK techniques added (Discovery or Attack phase)
- [x] **TECH-02**: At least 3 additional Exabeam UEBA scenarios added
- [x] **TECH-03**: All existing and new techniques have events manifest entries (VERIF-01 coverage)

### CrowdStrike Coverage

- [x] **CROW-01**: CrowdStrike Falcon detection rule mappings documented per technique in events manifest
- [ ] **CROW-02**: At least 3 techniques that specifically generate Falcon sensor events
- [ ] **CROW-03**: HTML report shows CrowdStrike-specific coverage column when Falcon events are present

### Sentinel Coverage

- [ ] **SENT-01**: Microsoft Sentinel detection rule mappings documented per technique in events manifest
- [ ] **SENT-02**: At least 3 techniques that target Azure AD / Microsoft Defender log sources
- [ ] **SENT-03**: HTML report shows Sentinel-specific coverage column when Azure events are present

## v2 Requirements

### Reporting

- **REPO-01**: MITRE ATT&CK Navigator layer export (JSON) from completed simulation
- **REPO-02**: Coverage score per SIEM platform (% of techniques with verified detections)
- **REPO-03**: PDF export of HTML report

### Distribution

- **DIST-01**: Installer / setup wizard for non-technical users
- **DIST-02**: Auto-update mechanism for new technique versions

### SIEM Integrations

- **SIEM-01**: Optional SIEM API query post-run (Exabeam, Sentinel) to verify rule fired
- **SIEM-02**: IBM QRadar detection mappings
- **SIEM-03**: Splunk detection mappings

## Out of Scope

| Feature | Reason |
|---------|--------|
| Real credential extraction | Tool simulates artifacts only — never performs real attacks |
| Non-Windows platforms | Techniques are fundamentally Windows Event Log / Sysmon based |
| SIEM API queries at runtime | Breaks standalone model; adds network/auth dependencies |
| Mobile or cloud deployment | Windows desktop/server tool by design |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| VERIF-01 | Phase 1 | Complete |
| VERIF-02 | Phase 1 | Complete |
| VERIF-03 | Phase 1 | Complete |
| VERIF-04 | Phase 1 | Complete |
| VERIF-05 | Phase 1 | Complete |
| QUAL-01 | Phase 2 | Complete |
| QUAL-02 | Phase 2 | Complete |
| QUAL-03 | Phase 2 | Complete |
| QUAL-04 | Phase 2 | Pending |
| QUAL-05 | Phase 2 | Complete |
| TECH-01 | Phase 3 | Complete |
| TECH-02 | Phase 3 | Complete |
| TECH-03 | Phase 3 | Complete |
| CROW-01 | Phase 4 | Complete |
| CROW-02 | Phase 4 | Pending |
| CROW-03 | Phase 4 | Pending |
| SENT-01 | Phase 5 | Pending |
| SENT-02 | Phase 5 | Pending |
| SENT-03 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 19 total
- Mapped to phases: 19
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-24*
*Last updated: 2026-03-24 after initial definition*
