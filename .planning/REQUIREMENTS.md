# Requirements: LogNoJutsu

**Defined:** 2026-04-09
**Core Value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

## v1.3 Requirements

Requirements for milestone v1.3: Realistic Attack Simulation.

### Safety & Audit

- [x] **SAFE-01**: All potentially destructive techniques (T1490, T1070.001, T1546.003) are audited and fixed to prevent damage on client machines
- [x] **SAFE-02**: All 57 existing techniques are classified as Tier 1 (realistic) / Tier 2 (partial) / Tier 3 (stub) with documented assessment
- [x] **SAFE-03**: All persistence and write techniques have verified cleanup commands that execute regardless of technique body success/failure

### Network Discovery

- [x] **SCAN-01**: T1046 scans the host's auto-detected /24 subnet via TCP connect scan (not just loopback/gateway)
- [ ] **SCAN-02**: T1018 includes ICMP ping sweep, ARP table enumeration, and nltest DC discovery
- [x] **SCAN-03**: Network scanning is implemented as native Go (type: go executor) generating Sysmon EID 3 artifacts

### Technique Realism Upgrades

- [ ] **TECH-01**: Discovery stub techniques upgraded to real tool execution (T1057, T1069, T1082, T1083, T1135, T1482)
- [ ] **TECH-02**: Persistence techniques added with mandatory cleanup (scheduled tasks, registry run keys, BITS jobs, service creation)
- [ ] **TECH-03**: Defense evasion techniques added (encoded commands, masquerading, rundll32 LOLBin, DLL sideloading)
- [ ] **TECH-04**: C2 and exfiltration techniques added using loopback/internal-only patterns (HTTP C2, DNS C2, data encoding, exfil-over-alt-protocol)

### Safety Infrastructure

- [x] **INFRA-01**: AMSI-blocked technique failures are classified separately from execution errors in verification results
- [x] **INFRA-02**: Admin vs non-admin execution is detected; techniques requiring elevation are skipped gracefully with clear status
- [x] **INFRA-03**: Network scan has target range confirmation, rate limiting, and IDS warning displayed before execution

### Architecture

- [x] **ARCH-01**: Native Go executor (type: go) added to executor with internal/native/ registry for Go-implemented techniques
- [x] **ARCH-02**: LDAP enumeration implemented via go-ldap/v3 with graceful fallback when no DC is reachable
- [x] **ARCH-03**: WMI queries implemented via go-ole/wmi (pure Go, no CGO) for native technique execution

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
| Ransomware on user directories | Unrecoverable risk if cleanup fails; T1486 scoped to %TEMP% only |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| SAFE-01 | Phase 14 | Complete |
| SAFE-02 | Phase 14 | In progress (14-01: Tier field added; 14-03: YAML annotation + classification doc) |
| SAFE-03 | Phase 14 | Complete |
| SCAN-01 | Phase 17 | Complete |
| SCAN-02 | Phase 17 | Pending |
| SCAN-03 | Phase 17 | Complete |
| TECH-01 | Phase 18 | Pending |
| TECH-02 | Phase 18 | Pending |
| TECH-03 | Phase 18 | Pending |
| TECH-04 | Phase 18 | Pending |
| INFRA-01 | Phase 16 | Complete |
| INFRA-02 | Phase 16 | Complete |
| INFRA-03 | Phase 16 | Complete |
| ARCH-01 | Phase 15 | Complete |
| ARCH-02 | Phase 15 | Complete |
| ARCH-03 | Phase 15 | Complete |

**Coverage:**
- v1.3 requirements: 16 total
- Mapped to phases: 16
- Unmapped: 0

---
*Requirements defined: 2026-04-09*
*Last updated: 2026-04-09 — traceability complete (Phases 14-18)*
