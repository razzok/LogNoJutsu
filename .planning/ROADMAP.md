# Roadmap: LogNoJutsu

## Milestones

- ✅ **v1.0 Verified & Expanded** — Phases 1-7 (shipped 2026-03-26)
- 🚧 **v1.1 Bug Fixes & UI Polish** — Phases 8-9 (in progress)

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

### 🚧 v1.1 Bug Fixes & UI Polish (In Progress)

**Milestone Goal:** Fix locale-dependent Windows Audit Policy failure and modernise the Web UI with dynamic versioning and visual polish.

- [ ] **Phase 8: Backend Correctness** — GUID-based audit policy + build-time version plumbing
- [ ] **Phase 9: UI Polish** — Version badge, error panels, language, technique count, tactic colours

## Phase Details

### Phase 8: Backend Correctness
**Goal**: The Go binary works correctly on non-English Windows and exposes a version endpoint
**Depends on**: Phase 7 (v1.0 complete)
**Requirements**: BUG-01, BUG-02, VER-01, VER-02
**Success Criteria** (what must be TRUE):
  1. Running preparation on a German Windows machine succeeds without "The parameter is incorrect" failures from auditpol
  2. When a preparation step fails, the error message includes the human-readable subcategory name alongside the raw exit code
  3. Building with `-ldflags "-X main.version=v1.1.0"` produces a binary whose banner prints `v1.1.0`; building without ldflags prints `dev`
  4. `GET /api/info` returns `{"version":"v1.1.0"}` (or `"dev"`) matching the injected build value
**Plans:** 2 plans
Plans:
- [ ] 08-01-PLAN.md — GUID-based audit policy migration + error message fix (BUG-01, BUG-02)
- [ ] 08-02-PLAN.md — Build-time version injection + /api/info endpoint (VER-01, VER-02)

### Phase 9: UI Polish
**Goal**: The Web UI displays accurate, English-only content with a live version badge and inline error feedback
**Depends on**: Phase 8
**Requirements**: VER-03, UI-01, UI-02, UI-03, UI-04
**Success Criteria** (what must be TRUE):
  1. The version badge in the Web UI shows the build-injected version fetched from `/api/info` on page load — not the hardcoded `v0.1.0`
  2. All visible text in the Scheduler tab and PoC mode configuration is in English — no German strings remain in the UI
  3. When a preparation step fails, an inline styled error panel appears in the Preparation tab — no browser `alert()` dialogs fire
  4. The Dashboard technique count stat box displays the live count loaded from `/api/techniques` (currently 57)
  5. Tactic badges for `command-and-control` and `ueba-scenario` render with correct colours — not grey
**Plans**: TBD
**UI hint**: yes

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
| 8 | Backend Correctness | v1.1 | 0/2 | Planning complete | - |
| 9 | UI Polish | v1.1 | 0/? | Not started | - |
