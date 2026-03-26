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
- [x] **CROW-02**: At least 3 techniques that specifically generate Falcon sensor events
- [x] **CROW-03**: HTML report shows CrowdStrike-specific coverage column when Falcon events are present

### Sentinel Coverage

- [x] **SENT-01**: Microsoft Sentinel detection rule mappings documented per technique in events manifest
- [x] **SENT-02**: At least 3 techniques that target Azure AD / Microsoft Defender log sources
- [x] **SENT-03**: HTML report shows Sentinel-specific coverage column when Azure events are present

## v1.1 Requirements

### Bug Fixes

- [ ] **BUG-01**: Windows Audit Policy subcategories use locale-independent GUIDs instead of English names — fixes failure on non-English (German) Windows installations
- [ ] **BUG-02**: Preparation step failure messages include human-readable subcategory description alongside raw error detail (not just "exit status 87")

### Version

- [x] **VER-01**: Version string declared as injectable Go `var` in main.go, overridable at build time via `-ldflags "-X main.version=v1.1.0"` — defaults to `"dev"` when built without ldflags
- [x] **VER-02**: Server exposes `GET /api/info` endpoint returning `{"version": "..."}` JSON
- [ ] **VER-03**: Web UI version badge fetches version from `/api/info` on page load — no more hardcoded `v0.1.0`

### UI Polish

- [ ] **UI-01**: All German strings in Web UI replaced with English equivalents (Scheduler tab, PoC mode configuration)
- [ ] **UI-02**: Preparation tab uses inline styled error panels instead of browser `alert()` for step failures
- [ ] **UI-03**: Dashboard displays total technique library count loaded from `/api/techniques` (currently 57)
- [ ] **UI-04**: Tactic badges render correct colours for `command-and-control` and `ueba-scenario` tactics (currently grey due to missing funcMap entries)

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
| QUAL-04 | Phase 2 | Complete |
| QUAL-05 | Phase 2 | Complete |
| TECH-01 | Phase 3 | Complete |
| TECH-02 | Phase 3 | Complete |
| TECH-03 | Phase 3 | Complete |
| CROW-01 | Phase 4 | Complete |
| CROW-02 | Phase 4 | Complete |
| CROW-03 | Phase 4 | Complete |
| SENT-01 | Phase 5 | Complete |
| SENT-02 | Phase 5 | Complete |
| SENT-03 | Phase 5 | Complete |
| BUG-01 | Phase 8 | Pending |
| BUG-02 | Phase 8 | Pending |
| VER-01 | Phase 8 | Complete |
| VER-02 | Phase 8 | Complete |
| VER-03 | Phase 9 | Pending |
| UI-01 | Phase 9 | Pending |
| UI-02 | Phase 9 | Pending |
| UI-03 | Phase 9 | Pending |
| UI-04 | Phase 9 | Pending |

**Coverage:**
- v1 requirements: 19 total — all mapped ✓
- v1.1 requirements: 9 total — all mapped ✓ (Phase 8: 4, Phase 9: 5)

---
*Requirements defined: 2026-03-24*
*Last updated: 2026-03-26 — v1.1 traceability added (BUG-01/02 → Phase 8, VER-01/02 → Phase 8, VER-03/UI-01/02/03/04 → Phase 9)*
