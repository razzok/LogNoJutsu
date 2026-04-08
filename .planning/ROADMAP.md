# Roadmap: LogNoJutsu

## Milestones

- ✅ **v1.0 Verified & Expanded** — Phases 1-7 (shipped 2026-03-26)
- ✅ **v1.1 Bug Fixes & UI Polish** — Phases 8-9 (shipped 2026-03-26)
- 🚧 **v1.2 PoC Mode Fix & Overhaul** — Phases 10-13 (in progress)

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

### 🚧 v1.2 PoC Mode Fix & Overhaul (In Progress)

**Milestone Goal:** Make PoC/Multiday mode work reliably with clear daily feedback — fix bugs that make it appear broken, add per-day execution tracking, and improve schedule visualization so consultants trust the tool during multi-week engagements.

- [ ] **Phase 10: PoC Engine Fixes & Clock Injection** - Fix stale day counter, German strings, missing log separators, and inject a testable clock interface
- [ ] **Phase 11: Daily Tracking Data Layer** - Record DayDigest structs, expose /api/poc/days, apply campaign delay_after during PoC Phase 2
- [ ] **Phase 12: Daily Digest & Timeline Calendar UI** - Render per-day digest panel and horizontal calendar grid from Phase 11 data
- [ ] **Phase 13: runPoC() Scheduling Tests** - Test day counter monotonicity, stop-signal handling, and DayDigest lifecycle using Phase 10 clock injection

## Phase Details

### Phase 10: PoC Engine Fixes & Clock Injection
**Goal**: The PoC engine runs correctly — day counter advances through all three phases, all status strings are English, phase transitions emit log separators, and the clock is injectable for deterministic testing
**Depends on**: Nothing (first phase of milestone; continues from Phase 9)
**Requirements**: POCFIX-01, POCFIX-02, POCFIX-03, TEST-01
**Success Criteria** (what must be TRUE):
  1. Starting a PoC run and watching the log shows English-only CurrentStep strings ("Waiting until", "Day N of M", "No actions") with no German text
  2. The day counter shown in CurrentStep reads the correct global day N across Phase1, Gap, and Phase2 without resetting or skipping
  3. Each phase transition (Phase1 start, Gap start, Phase2 start) produces a visible separator entry in the log viewer
  4. The engine accepts an injectable clock/wait function that tests can supply to eliminate real sleeps
**Plans**: TBD

### Phase 11: Daily Tracking Data Layer
**Goal**: The engine records a complete DayDigest for every PoC day from the moment the run starts, exposes it via a new API endpoint, and applies campaign delay_after between technique steps during Phase 2
**Depends on**: Phase 10
**Requirements**: TRACK-01, TRACK-02, TRACK-03, TRACK-04, CAMP-01
**Success Criteria** (what must be TRUE):
  1. Polling GET /api/poc/days immediately after PoC start returns a full array of pending entries — one per scheduled day — not an empty list
  2. Each DayDigest entry in the API response carries day number, phase label, status, technique results, pass/fail counts, start/end timestamps, and a last-heartbeat timestamp
  3. After a PoC day completes, its DayDigest entry transitions to "complete" with accurate technique results visible in the API response
  4. During PoC Phase 2 execution, the delay_after value from the campaign definition is applied between technique steps
**Plans**: TBD

### Phase 12: Daily Digest & Timeline Calendar UI
**Goal**: Consultants can visually inspect the full PoC schedule and per-day results from the web UI — a digest panel shows technique-level outcomes per day, and a calendar grid shows the complete timeline at a glance
**Depends on**: Phase 11
**Requirements**: DIGEST-01, DIGEST-02, DIGEST-03, CAL-01, CAL-02, CAL-03, CAL-04
**Success Criteria** (what must be TRUE):
  1. Opening the PoC tab during or after a run shows a per-day digest panel listing which techniques ran and their pass/fail results for each day
  2. The current day's digest entry is auto-expanded; all completed and future days are collapsed by default
  3. Each digest day entry shows technique count, pass/fail counts, and the execution time window
  4. A horizontal calendar grid shows every day in the PoC schedule with color coding: green for complete, yellow/accent for current, gray for future, muted for gap days
  5. Each calendar day cell shows a technique count badge, and Phase 1 / Gap / Phase 2 labels are visible above their respective day groups
**Plans**: TBD
**UI hint**: yes

### Phase 13: runPoC() Scheduling Tests
**Goal**: The runPoC() scheduling logic is covered by deterministic tests that validate day counter correctness, stop-signal handling, and DayDigest state transitions — using the injectable clock from Phase 10
**Depends on**: Phase 10, Phase 11
**Requirements**: TEST-02, TEST-03, TEST-04
**Success Criteria** (what must be TRUE):
  1. A test suite runs without real sleeps and asserts that the day counter increments monotonically across Phase1, Gap, and Phase2 transitions without gaps or resets
  2. A test exercises the stop-signal path mid-sleep and asserts that runPoC() exits cleanly without hanging
  3. A test validates the full DayDigest lifecycle — entries begin as "pending" at run start, transition to "active" when a day begins, and reach "complete" with correct counts when the day ends
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
| 10 | PoC Engine Fixes & Clock Injection | v1.2 | 0/? | Not started | - |
| 11 | Daily Tracking Data Layer | v1.2 | 0/? | Not started | - |
| 12 | Daily Digest & Timeline Calendar UI | v1.2 | 0/? | Not started | - |
| 13 | runPoC() Scheduling Tests | v1.2 | 0/? | Not started | - |
