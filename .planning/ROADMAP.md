# Roadmap: LogNoJutsu

## Milestones

- ✅ **v1.0 Verified & Expanded** — Phases 1-7 (shipped 2026-03-26)
- ✅ **v1.1 Bug Fixes & UI Polish** — Phases 8-9 (shipped 2026-03-26)
- 🔄 **v1.2 PoC Mode Fix & Overhaul** — Phases 10-13 (in progress 2026-04-08)

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

### 🔄 v1.2 PoC Mode Fix & Overhaul (Phases 10-13)

- [x] Phase 10: PoC Engine Fixes & Clock Injection (2/2 plans) — completed 2026-04-08
- [x] Phase 11: Daily Tracking Backend & Campaign Delay (2/2 plans) — completed 2026-04-09
- [x] Phase 12: Daily Digest & Timeline Calendar UI (completed 2026-04-09)
- [ ] Phase 13: PoC Scheduling Tests

### Phase 10: PoC Engine Fixes & Clock Injection

**Goal:** Fix all four PoC engine bugs (day counter, German strings, log separators) and inject Clock interface for testability.
**Requirements:** POCFIX-01, POCFIX-02, POCFIX-03, TEST-01
**Depends on:** None
**Plans:** 2 plans

Plans:
- [x] 10-01: PoC engine fixes & Clock interface — completed 2026-04-08
- [x] 10-02: PoC engine validation tests — completed 2026-04-08

### Phase 11: Daily Tracking Backend & Campaign Delay

**Goal:** Add DayDigest struct to engine for per-day PoC tracking, expose via API, and apply campaign delay_after during Phase 2 execution.
**Requirements:** TRACK-01, TRACK-02, TRACK-03, TRACK-04, CAMP-01
**Depends on:** Phase 10
**Plans:** 2/2 plans complete

Plans:
- [x] 11-01-PLAN.md — DayDigest struct, pre-population, lifecycle mutations, heartbeat, campaign delay, getter + tests
- [x] 11-02-PLAN.md — GET /api/poc/days endpoint and server tests

### Phase 12: Daily Digest & Timeline Calendar UI

**Goal:** Add daily digest panel and timeline calendar to the web UI for PoC schedule visualization.
**Requirements:** DIGEST-01, DIGEST-02, DIGEST-03, CAL-01, CAL-02, CAL-03, CAL-04
**Depends on:** Phase 11
**Plans:** 1/1 plans complete

Plans:
- [x] 12-01-PLAN.md — Timeline calendar strip + daily digest accordion + polling integration

### Phase 13: PoC Scheduling Tests

**Goal:** Write deterministic tests for runPoC() scheduling logic using the fake clock — day counter transitions, stop-signal handling, DayDigest lifecycle.
**Requirements:** TEST-02, TEST-03, TEST-04
**Depends on:** Phase 11
**Plans:** 1 plan

Plans:
- [ ] 13-01-PLAN.md — Day counter monotonicity, stop-signal handling (4 scenarios), DayDigest lifecycle transitions

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
| 12 | Daily Digest & Timeline Calendar UI | v1.2 | 0/1 | Not started | — |
| 13 | PoC Scheduling Tests | v1.2 | 0/1 | Not started | — |
