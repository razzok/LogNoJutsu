# Roadmap: LogNoJutsu

## Milestones

- ✅ **v1.0 Verified & Expanded** — Phases 1-7 (shipped 2026-03-26)
- ✅ **v1.1 Bug Fixes & UI Polish** — Phases 8-9 (shipped 2026-03-26)
- ✅ **v1.2 PoC Mode Fix & Overhaul** — Phases 10-13 (shipped 2026-04-09)
- 🚧 **v1.3 Realistic Attack Simulation** — Phases 14-18 (in progress)

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

### 🚧 v1.3 Realistic Attack Simulation (Phases 14-18)

- [x] Phase 14: Safety Audit (3/3 plans) — completed 2026-04-09
- [x] Phase 15: Native Go Architecture (2/2 plans) — completed 2026-04-09
- [x] Phase 16: Safety Infrastructure (3/3 plans) — completed 2026-04-09
- [ ] Phase 17: Network Discovery (0/2 plans) — Native Go TCP/ICMP network scanning (T1046 subnet scan, T1018 ping sweep/ARP/DC discovery). Depends on: Phase 15. Reqs: SCAN-01, SCAN-02, SCAN-03
  Plans:
  - [x] 17-01-PLAN.md — T1046 TCP/UDP connect scanner with goroutine pool
  - [ ] 17-02-PLAN.md — T1018 ICMP/ARP/nltest/DNS discovery chain
- [ ] Phase 18: Technique Realism Upgrades — Discovery stub upgrades, persistence techniques, defense evasion, C2/exfiltration. Reqs: TECH-01, TECH-02, TECH-03, TECH-04

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
| 16 | Safety Infrastructure | v1.3 | 3/3 | Complete | 2026-04-09 |
| 17 | Network Discovery | v1.3 | 0/2 | Not started | — |
| 18 | Technique Realism Upgrades | v1.3 | 0/0 | Not started | — |
