# Roadmap: LogNoJutsu

## Milestones

- ✅ **v1.0 Verified & Expanded** — Phases 1-7 (shipped 2026-03-26)
- ✅ **v1.1 Bug Fixes & UI Polish** — Phases 8-9 (shipped 2026-03-26)
- ✅ **v1.2 PoC Mode Fix & Overhaul** — Phases 10-13 (shipped 2026-04-09)
- **v1.3 Realistic Attack Simulation** — Phases 14-18 (in progress)

## Phases

<details>
<summary>✅ v1.0 Verified & Expanded (Phases 1-7) — SHIPPED 2026-03-26</summary>

- [x] Phase 1: Events Manifest & Verification Engine (3/3 plans) — completed 2026-03-24
- [x] Phase 2: Code Structure & Test Coverage (3/3 plans) — completed 2026-03-25
- [x] Phase 3: Additional Techniques (3/3 plans) — completed 2026-03-25
- [x] Phase 4: CrowdStrike SIEM Coverage (3/3 plans) — completed 2026-03-25
- [x] Phase 5: Microsoft Sentinel Coverage (3/3 plans) — completed 2026-03-25
- [x] Phase 6: Documentation Consistency (1/1 plan) — completed 2026-03-26
- [x] Phase 7: Nyquist Validation (1/1 plan) — completed 2026-03-26

Full details: `.planning/milestones/v1.0-ROADMAP.md`

</details>

<details>
<summary>✅ v1.1 Bug Fixes & UI Polish (Phases 8-9) — SHIPPED 2026-03-26</summary>

- [x] Phase 8: Backend Correctness (2/2 plans) — completed 2026-03-26
- [x] Phase 9: UI Polish (3/3 plans) — completed 2026-03-26

Full details: `.planning/milestones/v1.1-ROADMAP.md`

</details>

<details>
<summary>✅ v1.2 PoC Mode Fix & Overhaul (Phases 10-13) — SHIPPED 2026-04-09</summary>

- [x] Phase 10: PoC Engine Fixes & Clock Injection (2/2 plans) — completed 2026-04-08
- [x] Phase 11: Daily Tracking Backend & Campaign Delay (2/2 plans) — completed 2026-04-09
- [x] Phase 12: Daily Digest & Timeline Calendar UI (1/1 plan) — completed 2026-04-09
- [x] Phase 13: PoC Scheduling Tests (1/1 plan) — completed 2026-04-09

Full details: `.planning/milestones/v1.2-ROADMAP.md`

</details>

### v1.3 Realistic Attack Simulation

- [x] **Phase 14: Safety Audit** — All existing techniques audited, classified, and made safe before any upgrades (1/3 plans complete) (completed 2026-04-09)
- [ ] **Phase 15: Native Go Architecture** — Go executor and native Go libraries in place to support realistic technique execution
- [ ] **Phase 16: Safety Infrastructure** — AMSI classification, elevation detection, and scan safeguards protect technique execution
- [ ] **Phase 17: Network Discovery** — Real ARP/ICMP/TCP scanning of the local /24 subnet with user-facing safeguards
- [ ] **Phase 18: Technique Realism** — Discovery stubs replaced with real tool execution; persistence, defense evasion, and C2 technique sets added

## Phase Details

### Phase 14: Safety Audit
**Goal**: All existing techniques are safe to run on client machines, classified by realism tier, and have verified cleanup paths
**Depends on**: Nothing (must run before any technique changes)
**Requirements**: SAFE-01, SAFE-02, SAFE-03
**Success Criteria** (what must be TRUE):
  1. Destructive techniques (T1490, T1070.001, T1546.003) execute without causing file loss, log deletion, or persistent registry damage on a test machine
  2. All 57 techniques carry a Tier 1/2/3 label visible in their YAML (or classification manifest), so a consultant can instantly know which fire realistic events
  3. Every technique that writes to disk, registry, or scheduled tasks has a cleanup command that runs even when the technique body fails or is interrupted
  4. The tier classification document exists and covers all 57 techniques with a rationale for each assignment
**Plans**: 3 plans (3 complete)
Plans:
- [x] 14-01-PLAN.md — Tier field, executor defer cleanup, test scaffolds (completed 2026-04-09)
- [x] 14-02-PLAN.md — Destructive technique rewrites and cleanup audit (completed 2026-04-09)
- [x] 14-03-PLAN.md — Tier YAML classification, classification doc, UI/report badges

### Phase 15: Native Go Architecture
**Goal**: The executor layer supports native Go techniques and the two Go libraries needed for realistic AD and WMI queries are integrated
**Depends on**: Phase 14
**Requirements**: ARCH-01, ARCH-02, ARCH-03
**Success Criteria** (what must be TRUE):
  1. A technique YAML with `type: go` executes its registered Go function through the executor without spawning a child process
  2. The internal/native registry exists with at least one Go-implemented technique that compiles and runs in the test suite
  3. LDAP enumeration runs against a reachable DC and returns results; when no DC is reachable it logs a graceful fallback message instead of crashing
  4. WMI queries execute via pure Go (no CGO) and return process or system data that appears in the technique result log
**Plans**: 2 plans
Plans:
- [ ] 15-01-PLAN.md — Native registry package and type:go executor dispatch
- [ ] 15-02-PLAN.md — T1482 LDAP and T1057 WMI techniques with YAML updates

### Phase 16: Safety Infrastructure
**Goal**: The engine detects AMSI blocks and missing elevation, and network scans require explicit user acknowledgment before running
**Depends on**: Phase 14
**Requirements**: INFRA-01, INFRA-02, INFRA-03
**Success Criteria** (what must be TRUE):
  1. When a PowerShell technique is blocked by AMSI, the verification result shows "AMSI Blocked" as a distinct status — not "Failed" or "Error"
  2. When a technique requiring admin is run by a non-admin user, the engine skips it with a visible "Elevation required" status rather than producing a misleading failure
  3. Before a network scan starts, the UI displays the target subnet, a rate-limit notice, and an IDS warning; the scan does not proceed until the user confirms
  4. A consultant running non-admin can complete a full simulation without cryptic error messages for elevation-gated techniques
**Plans**: 3 plans
Plans:
- [ ] 14-01-PLAN.md — Tier field, executor defer cleanup, test scaffolds
- [ ] 14-02-PLAN.md — Destructive technique rewrites and cleanup audit
- [ ] 14-03-PLAN.md — Tier YAML classification, classification doc, UI/report badges
**UI hint**: yes

### Phase 17: Network Discovery
**Goal**: T1018 and T1046 perform real network reconnaissance on the host's local /24 subnet, generating authentic Sysmon artifacts
**Depends on**: Phase 15, Phase 16
**Requirements**: SCAN-01, SCAN-02, SCAN-03
**Success Criteria** (what must be TRUE):
  1. T1046 auto-detects the host's primary network interface subnet and performs a TCP connect scan across the /24 range — not just loopback or a hardcoded gateway
  2. T1018 combines ICMP ping sweep, ARP table enumeration, and nltest DC discovery into a single technique execution that logs all three discovery methods
  3. Both scan techniques are implemented as native Go (`type: go`) and generate Sysmon EID 3 (Network Connection) events observable in Event Viewer after execution
  4. The scan completes within a reasonable time on a typical /24 (rate limiting active) and does not cause network errors that fail the technique verification
**Plans**: 3 plans
Plans:
- [ ] 14-01-PLAN.md — Tier field, executor defer cleanup, test scaffolds
- [ ] 14-02-PLAN.md — Destructive technique rewrites and cleanup audit
- [ ] 14-03-PLAN.md — Tier YAML classification, classification doc, UI/report badges

### Phase 18: Technique Realism
**Goal**: Stub discovery techniques are replaced with real tool execution, and the technique library gains persistence, defense evasion, and C2 categories
**Depends on**: Phase 15, Phase 16
**Requirements**: TECH-01, TECH-02, TECH-03, TECH-04
**Success Criteria** (what must be TRUE):
  1. T1057, T1069, T1082, T1083, T1135, and T1482 no longer use PowerShell echo stubs — each invokes the real Windows tool or API and produces the expected event IDs in the Windows Event Log
  2. At least one persistence technique (scheduled task, registry run key, BITS job, or service creation) executes, generates its expected event, and then cleans up completely — leaving no artifact on the test machine after the simulation ends
  3. At least one defense evasion technique (encoded command, masquerading, rundll32, or DLL sideloading) executes and produces a detectable Sysmon or Security event
  4. At least one C2/exfiltration technique using loopback or internal-only patterns (HTTP C2, DNS C2, data encoding, or exfil-over-alt-protocol) executes and generates the expected network or process event
  5. All new techniques pass the verification engine (expected vs. observed event IDs) and appear correctly in the HTML report
**Plans**: 3 plans
Plans:
- [ ] 14-01-PLAN.md — Tier field, executor defer cleanup, test scaffolds
- [ ] 14-02-PLAN.md — Destructive technique rewrites and cleanup audit
- [ ] 14-03-PLAN.md — Tier YAML classification, classification doc, UI/report badges

## Progress

| Phase | Title | Milestone | Plans Complete | Status | Completed |
|-------|-------|-----------|----------------|--------|-----------|
| 1 | Events Manifest & Verification Engine | v1.0 | 3/3 | Complete | 2026-03-24 |
| 2 | Code Structure & Test Coverage | v1.0 | 3/3 | Complete | 2026-03-25 |
| 3 | Additional Techniques | v1.0 | 3/3 | Complete | 2026-03-25 |
| 4 | CrowdStrike SIEM Coverage | v1.0 | 3/3 | Complete | 2026-03-25 |
| 5 | Microsoft Sentinel Coverage | v1.0 | 3/3 | Complete | 2026-03-25 |
| 6 | Documentation Consistency | v1.0 | 1/1 | Complete | 2026-03-26 |
| 7 | Nyquist Validation | v1.0 | 1/1 | Complete | 2026-03-26 |
| 8 | Backend Correctness | v1.1 | 2/2 | Complete | 2026-03-26 |
| 9 | UI Polish | v1.1 | 3/3 | Complete | 2026-03-26 |
| 10 | PoC Engine Fixes & Clock Injection | v1.2 | 2/2 | Complete | 2026-04-08 |
| 11 | Daily Tracking Backend & Campaign Delay | v1.2 | 2/2 | Complete | 2026-04-09 |
| 12 | Daily Digest & Timeline Calendar UI | v1.2 | 1/1 | Complete | 2026-04-09 |
| 13 | PoC Scheduling Tests | v1.2 | 1/1 | Complete | 2026-04-09 |
| 14 | Safety Audit | v1.3 | 3/3 | Complete | 2026-04-09 |
| 15 | Native Go Architecture | v1.3 | 0/2 | Planned | - |
| 16 | Safety Infrastructure | v1.3 | 0/? | Not started | - |
| 17 | Network Discovery | v1.3 | 0/? | Not started | - |
| 18 | Technique Realism | v1.3 | 0/? | Not started | - |
