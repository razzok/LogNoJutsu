# Roadmap: LogNoJutsu

## Milestones

- ✅ **v1.0 Verified & Expanded** — Phases 1-7 (shipped 2026-03-26)
- ✅ **v1.1 Bug Fixes & UI Polish** — Phases 8-9 (shipped 2026-03-26)
- ✅ **v1.2 PoC Mode Fix & Overhaul** — Phases 10-13 (shipped 2026-04-09)
- ✅ **v1.3 Realistic Attack Simulation** — Phases 14-18 (shipped 2026-04-10)
- 🚧 **v1.4 PoC Technique Distribution** — Phases 19-20 (in progress)

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

<details>
<summary>✅ v1.3 Realistic Attack Simulation (Phases 14-18) — SHIPPED 2026-04-10</summary>

- [x] Phase 14: Safety Audit (3/3 plans) — completed 2026-04-09
- [x] Phase 15: Native Go Architecture (2/2 plans) — completed 2026-04-09
- [x] Phase 16: Safety Infrastructure (3/3 plans) — completed 2026-04-09
- [x] Phase 17: Network Discovery (2/2 plans) — completed 2026-04-10
- [x] Phase 18: Technique Realism Upgrades (1/1 plan) — completed 2026-04-10

Full details: `.planning/milestones/v1.3-ROADMAP.md`

</details>

### v1.4 PoC Technique Distribution (In Progress)

**Milestone Goal:** Spread technique execution across the day with random jitter — Phase 1 one-at-a-time, Phase 2 in small batches — instead of firing all techniques at the scheduled hour.

- [x] **Phase 19: Distributed Technique Scheduling** - Rewrite runPoC() to spread techniques across the day with jitter for both Phase 1 (single) and Phase 2 (batched) execution (completed 2026-04-10)
- [ ] **Phase 20: Scheduling Test Coverage** - Update existing PoC scheduling tests to validate distributed execution behavior and DayDigest accuracy under the new logic

## Phase Details

### Phase 19: Distributed Technique Scheduling
**Goal**: runPoC() distributes techniques across each campaign day with random jitter instead of firing all techniques at the scheduled hour
**Depends on**: Phase 18
**Requirements**: POC-01, POC-02, POC-03
**Success Criteria** (what must be TRUE):
  1. Phase 1 techniques execute one at a time at randomly spaced intervals throughout the day, not all at once at Phase1DailyHour
  2. Phase 2 techniques execute in batches of 2-3 at randomly spaced intervals throughout the day, not all at once at Phase2DailyHour
  3. All random intervals fall within a configurable daily time window (start hour to end hour), not beyond the window boundaries
  4. Each day's DayDigest TechniqueCount still reflects the full set of techniques scheduled for that day
**Plans**: 3 plans
Plans:
- [x] 19-00-PLAN.md — Wave 0: test stubs for poc_schedule_test.go (Nyquist compliance)
- [x] 19-01-PLAN.md — Engine rewrite: PoCConfig window fields, randomSlotsInWindow helper, distributed runPoC() loops
- [x] 19-02-PLAN.md — UI update: window start/end form inputs and schedule preview

### Phase 20: Scheduling Test Coverage
**Goal**: Existing and new PoC scheduling tests validate distributed execution behavior with deterministic fake-clock control
**Depends on**: Phase 19
**Requirements**: POC-04
**Success Criteria** (what must be TRUE):
  1. All existing poc_test.go scheduling tests pass with the rewritten runPoC() logic (no regressions)
  2. At least one test verifies that Phase 1 techniques are not all dispatched at a single clock tick
  3. At least one test verifies that Phase 2 techniques are dispatched in groups of 2-3
  4. DayDigest tracking remains accurate under distributed scheduling (TechniqueCount matches dispatched count)
**Plans**: TBD

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
| 15 | Native Go Architecture | v1.3 | 2/2 | Complete | 2026-04-09 |
| 16 | Safety Infrastructure | v1.3 | 3/3 | Complete | 2026-04-09 |
| 17 | Network Discovery | v1.3 | 2/2 | Complete | 2026-04-10 |
| 18 | Technique Realism Upgrades | v1.3 | 1/1 | Complete | 2026-04-10 |
| 19 | Distributed Technique Scheduling | v1.4 | 0/3 | Not started | - |
| 20 | Scheduling Test Coverage | v1.4 | 0/? | Not started | - |
