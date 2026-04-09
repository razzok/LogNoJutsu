# Requirements: LogNoJutsu

**Defined:** 2026-04-08
**Core Value:** Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

## v1.2 Requirements

Requirements for PoC Mode Fix & Overhaul. Each maps to roadmap phases.

### Bug Fixes

- [x] **POCFIX-01**: PoC day counter updates correctly across all three phases (Phase1, Gap, Phase2) — showing global day N of total
- [x] **POCFIX-02**: All CurrentStep strings in runPoC() display in English (no German "Tag", "warte bis", "keine Aktionen")
- [x] **POCFIX-03**: Phase transitions in runPoC() produce simlog.Phase() separator entries visible in log viewer

### Daily Tracking

- [x] **TRACK-01**: Engine records a DayDigest struct per PoC day containing: day number, phase, status, techniques executed, pass/fail counts, start/end timestamps
- [x] **TRACK-02**: DayDigest entries are pre-populated as "pending" at runPoC() start so the full schedule is visible from first poll
- [x] **TRACK-03**: GET /api/poc/days endpoint returns the DayDigest array (behind authMiddleware)
- [x] **TRACK-04**: DayDigest includes a "last heartbeat" timestamp proving the engine was alive during each day's execution window

### UI — Daily Digest

- [x] **DIGEST-01**: User can see a per-day summary panel showing which techniques ran and their results
- [x] **DIGEST-02**: Current day auto-expands; completed days are collapsed by default
- [x] **DIGEST-03**: Each day entry shows technique count, pass/fail counts, and execution time window

### UI — Timeline Calendar

- [x] **CAL-01**: User can see a horizontal day-by-day grid showing the full PoC schedule
- [x] **CAL-02**: Days are color-coded: green (complete), yellow/accent (current), gray (future), muted (gap)
- [x] **CAL-03**: Each day cell shows technique count badge
- [x] **CAL-04**: Phase labels (Phase 1 / Gap / Phase 2) are visible above day groups

### Testability

- [x] **TEST-01**: Engine accepts injectable clock/wait function for deterministic runPoC() testing
- [ ] **TEST-02**: Tests validate monotonic day counter across Phase1→Gap→Phase2 transitions
- [ ] **TEST-03**: Tests validate stop-signal handling during PoC sleep periods
- [ ] **TEST-04**: Tests validate DayDigest lifecycle (pending→active→complete)

### Campaign Execution

- [x] **CAMP-01**: Campaign delay_after field is applied between technique steps during PoC Phase 2 execution

## Future Requirements

### PoC Reporting

- **REPORT-01**: HTML report includes per-day breakdown section for PoC runs
- **REPORT-02**: DayDigest data persists across binary restart (disk serialization)

### Log Performance

- **LOG-01**: Log entries paginated with offset/limit on GET /api/logs to prevent multi-week memory growth

## Out of Scope

| Feature | Reason |
|---------|--------|
| Per-day SIEM verification results | Requires SIEM API integration — deferred to future milestone |
| Real-time WebSocket updates | SSE/WebSocket adds complexity; 2-3s polling sufficient for daily-execution tool |
| Parallel PoC simulations | Engine is singleton by design; multi-simulation not needed for current use case |
| PoC state persistence to disk | In-memory state acceptable for v1.2; restart resumes from scratch |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| POCFIX-01 | Phase 10 | Complete |
| POCFIX-02 | Phase 10 | Complete |
| POCFIX-03 | Phase 10 | Complete |
| TEST-01 | Phase 10 | Complete |
| TRACK-01 | Phase 11 | Complete |
| TRACK-02 | Phase 11 | Complete |
| TRACK-03 | Phase 11 | Complete |
| TRACK-04 | Phase 11 | Complete |
| CAMP-01 | Phase 11 | Complete |
| DIGEST-01 | Phase 12 | Complete |
| DIGEST-02 | Phase 12 | Complete |
| DIGEST-03 | Phase 12 | Complete |
| CAL-01 | Phase 12 | Complete |
| CAL-02 | Phase 12 | Complete |
| CAL-03 | Phase 12 | Complete |
| CAL-04 | Phase 12 | Complete |
| TEST-02 | Phase 13 | Pending |
| TEST-03 | Phase 13 | Pending |
| TEST-04 | Phase 13 | Pending |

**Coverage:**
- v1.2 requirements: 19 total
- Mapped to phases: 19
- Unmapped: 0 ✓

---
*Requirements defined: 2026-04-08*
*Last updated: 2026-04-08 — traceability filled after roadmap creation*
