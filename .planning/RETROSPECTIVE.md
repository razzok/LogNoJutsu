# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.0 — Verified & Expanded

**Shipped:** 2026-03-26
**Phases:** 7 | **Plans:** 17 | **Commits:** 126

### What Was Built
- Events manifest + verification engine: EventSpec struct, injectable QueryFn verifier, pass/fail per technique in HTML report
- Code quality: Server struct refactor (no more package globals), 14 unit tests across 5 packages, `go test ./...` green
- Technique library: 57 total (43 base + 5 ATT&CK Collection/C2 + 4 Exabeam UEBA + 3 Falcon + 3 Azure)
- Multi-SIEM coverage: CrowdStrike Falcon detection names + Microsoft Sentinel analytic rule names in SIEMCoverage map, conditional HTML columns per SIEM
- Documentation completeness: VALIDATION.md Nyquist compliance for all phases, requirements traceability corrected, README translated to English

### What Worked
- GSD phased execution: each phase delivered a discrete, verifiable increment; no phase required rollback
- QueryFn injection pattern: made verifier 100% testable without real PowerShell — zero flaky tests
- SIEMCoverage map[string][]string: single data model extension served both CrowdStrike and Sentinel phases identically — no structural change needed for Phase 5 after Phase 4 established the pattern
- Milestone audit before closing: identified 4 tech debt items and 2 unexecuted phases; Phases 6+7 resolved them cleanly
- Nyquist validation as a final phase: forced sign-off of all test coverage, surfaced function name mismatches (TestQueryCount → TestQueryCountMock, TestHasSentinel → TestHTMLSentinelColumn) that would have been silent debt

### What Was Inefficient
- VALIDATION.md files were created in draft state during planning and never promoted until Phase 7 — should be promoted at phase completion, not deferred
- Windows Defender quarantine of playbooks.test.exe discovered mid-execution with no prior mitigation — the exclusion path should be documented in CLAUDE.md upfront
- tacticColor funcMap entries for new tactics (command-and-control, ueba-scenario) missed during Phase 3/4/5 — cosmetic but now deferred to v2.0; review template needed
- README stayed in German until a late quick task; translation should be part of each phase's docs task, not a catch-up

### Patterns Established
- **SIEMCoverage pattern:** `map[string][]string` keyed by SIEM name, propagated on Technique and ExecutionResult, conditional HTML column when any result has mappings — reusable for any future SIEM
- **Conditional report column pattern:** `HasX bool` field computed at render time, column header and cells absent when `HasX` is false for all results
- **QueryFn injection:** all external I/O goes through injectable funcs — keeps tests hermetic without mocks library
- **Wave 0 test pre-commitment:** required test files named before implementation prevents function name drift; missed in v1.0, enforced via VALIDATION.md sign-off

### Key Lessons
1. **Sign off VALIDATION.md at phase close, not at milestone close.** Deferring all 5 to Phase 7 created unnecessary batch work and required re-running tests to confirm all names.
2. **Add Windows Defender exclusion note to CLAUDE.md at project start** for any Go project on Windows — playbooks.test.exe will always be quarantined without it.
3. **Extend tacticColor/badge maps at the same time you add a new tactic** — don't let cosmetic gaps accumulate; one-line fix becomes a milestone-level debt item.
4. **SIEMCoverage map pattern is highly reusable** — future SIEM integrations (Splunk ES, QRadar, etc.) require no structural changes, only new YAML keys and a new report column.

### Cost Observations
- Model mix: ~100% sonnet (all executor/verifier agents ran on claude-sonnet-4-6)
- Sessions: ~10 (one per phase execution + planning sessions)
- Notable: Parallel worktree execution kept orchestrator context under 15%; VALIDATION.md promotion (Phase 7) completed in ~5 min on sonnet with no rework

---

## Milestone: v1.2 — PoC Mode Fix & Overhaul

**Shipped:** 2026-04-09
**Phases:** 4 | **Plans:** 6 | **Commits:** 37

### What Was Built
- Clock interface injected into Engine for deterministic testing — fakeClock eliminates real sleeps in all PoC tests
- DayDigest per-day execution tracking with pending pre-population, lifecycle mutations (pending→active→complete), heartbeat timestamps, and campaign DelayAfter support
- GET /api/poc/days endpoint behind authMiddleware for daily digest data consumption
- Timeline calendar strip (horizontal phase-grouped day grid with color-coded status) and daily digest accordion (auto-expand current day, collapsed completed days)
- 6 deterministic PoC scheduling tests: day counter monotonicity, 4 stop-signal scenarios, DayDigest lifecycle transitions

### What Worked
- **Clock injection pattern:** Enabled 10+ fake-clock tests across two phases (10 and 13) with zero flakiness — the captureClock wrapper solved race conditions between fast fake timers and polling goroutines
- **Phase dependency chain:** P10→P11→P12 built cleanly on each other — Clock interface enabled DayDigest testing, DayDigest enabled API endpoint, API endpoint enabled UI polling
- **UI-SPEC design contract:** Phase 12 UI-SPEC caught typography and spacing issues before implementation, preventing rework
- **Milestone audit:** Caught all tech debt items (4 minor) and confirmed 19/19 requirements satisfied before closing — no surprises

### What Was Inefficient
- VALIDATION.md files still created in draft state and never promoted — same v1.0 lesson not yet enforced
- Phase 10 one-liner field missing from SUMMARY.md frontmatter — minor but broke automated extraction
- v1.1 retrospective section was skipped entirely — milestone went straight from v1.0 to v1.2 in RETROSPECTIVE.md

### Patterns Established
- **captureClock pattern:** Wrap fakeClock to synchronously capture engine state on each After() call — prevents race conditions in fast-timer tests. Variants: dayCaptureClock (PoCDay+Phase), digestCaptureClock (full []DayDigest snapshot), stopOnNthClock (configurable block-at-N)
- **DayDigest pre-population:** Pre-populate all days as "pending" at runPoC() start so schedule is visible from first poll — eliminates progressive discovery UX problem
- **hasDayData persistence:** UI panels persist visibility after PoC completion using a flag set once data arrives — not gated on running status

### Key Lessons
1. **captureClock is the go-to pattern for fake-clock testing** — direct polling of engine status after fake timer fires is inherently racy; synchronous capture eliminates it
2. **Pre-populate schedules at start, not progressively** — users need the full picture from the first poll, not gradual revelation
3. **VALIDATION.md promotion at phase close still not enforced** — third milestone with this lesson; needs a structural fix (hook or checklist gate)
4. **Don't skip milestone retrospectives** — v1.1 was missed entirely, losing 2 phases of process observations

### Cost Observations
- Model mix: ~90% sonnet (executors/verifiers), ~10% opus (orchestration, UI-SPEC)
- Sessions: ~6 (planning + execution per phase)
- Notable: Phase 12 UI-SPEC checker caught 3 issues pre-implementation; 37 commits across 2 days is efficient for 4 phases

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Phases | Plans | Key Change |
|-----------|--------|-------|------------|
| v1.0 | 7 | 17 | First GSD milestone — established SIEMCoverage pattern, Nyquist sign-off, QueryFn injection |
| v1.1 | 2 | 5 | Bug fixes & UI polish — GUID audit policy, version injection, inline error panels |
| v1.2 | 4 | 6 | PoC overhaul — Clock injection, DayDigest tracking, timeline calendar UI, scheduling tests |

### Cumulative Quality

| Milestone | Test Functions | Packages Tested | Notes |
|-----------|---------------|-----------------|-------|
| v1.0 | 14 | 5/9 | playbooks blocked by Defender; executor/preparation/simlog/userstore have no testable logic |
| v1.1 | 20 | 5/9 | +6 tests from backend correctness and UI polish phases |
| v1.2 | 26 | 5/9 | +6 PoC scheduling tests via fake clock injection (poc_test.go) |

### Top Lessons (Verified Across Milestones)

1. Sign off VALIDATION.md at phase close — not as a batch at milestone close (v1.0, v1.2 — still not enforced)
2. Document environment constraints (Defender, no CGO) in CLAUDE.md before first phase
3. captureClock pattern eliminates race conditions in fake-timer tests (v1.2 — validated across 10+ tests)
4. Pre-populate schedules at start for immediate visibility (v1.2)
5. Don't skip milestone retrospectives — process observations get lost (v1.2)
