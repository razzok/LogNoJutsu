# Requirements: LogNoJutsu

**Defined:** 2026-04-10
**Core Value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

## v1.4 Requirements

Requirements for milestone v1.4: PoC Technique Distribution.

### PoC Scheduling

- [x] **POC-01**: Phase 1 techniques execute one at a time at random intervals throughout the day instead of all at the scheduled hour
- [x] **POC-02**: Phase 2 techniques execute in small batches (2-3) at random intervals throughout the day instead of all at the scheduled hour
- [x] **POC-03**: Random jitter is bounded within a configurable daily time window (e.g., start hour to end hour)
- [ ] **POC-04**: Existing PoC scheduling tests updated to validate distributed execution and DayDigest accuracy

## Future Requirements

Deferred to future releases.

### Campaign Chains
- **CHAIN-01**: Network scan feeds host list into credential access which feeds lateral movement as a narrative sequence
- **CHAIN-02**: Technique parameterization via YAML (subnet, scan_timeout_ms, port_list) configurable per engagement

### Additional Techniques
- **ADD-01**: T1021.006 WinRM loopback lateral movement simulation
- **ADD-02**: Collection techniques with real file enumeration (T1005, T1119, T1560.001)
- **ADD-03**: Multi-host LDAP enumeration across discovered AD users

## Out of Scope

| Feature | Reason |
|---------|--------|
| Real exploitation (CVE-based attacks) | Causes actual damage; SIEM detects the event artifacts, not the vulnerability |
| Cross-subnet scanning | Generates noise on production segments not authorized for testing |
| Credential extraction (displaying actual hashes) | Data handling liability during client engagements |
| Real lateral movement to remote hosts | Requires authenticated access to remote systems; unauthorized access risk |
| C2 beacon to external infrastructure | Violates offline/standalone deployment requirement |
| UI changes for technique distribution | Scheduling is engine-internal; existing DayDigest UI already shows per-day progress |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| POC-01 | Phase 19 | Complete |
| POC-02 | Phase 19 | Complete |
| POC-03 | Phase 19 | Complete |
| POC-04 | Phase 20 | Pending |

**Coverage:**
- v1.4 requirements: 4 total
- Mapped to phases: 4
- Unmapped: 0

---
*Requirements defined: 2026-04-10*
*Last updated: 2026-04-10 after roadmap creation — all 4 requirements mapped*
