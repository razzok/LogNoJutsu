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

## Milestone: v1.3 — Realistic Attack Simulation

**Shipped:** 2026-04-10
**Phases:** 7 (08-09, 14-18) | **Plans:** 16

### What Was Built
- Safety audit: all 58 techniques classified Tier 1/2/3, destructive techniques (T1070.001, T1490) rewritten for client safety, defer-style cleanup guarantees
- Native Go architecture: type:go executor dispatch, thread-safe technique registry, T1482 LDAP trust discovery, T1057 WMI process discovery
- Safety infrastructure: AMSI block detection, elevation gating with graceful skip, scan confirmation modal with subnet/IDS warning
- Network discovery: T1046 TCP/UDP /24 subnet scanner (15-worker goroutine pool), T1018 four-method remote system discovery (ICMP/ARP/nltest/DNS)
- Technique realism: 4 discovery techniques reclassified (T1069/T1082/T1083 to Tier 1, T1135 to Tier 2), expected_events enriched with Sysmon entries

### What Worked
- **Safety-first design:** Phases 14 and 16 establishing safety gates before network scanning (Phase 17) meant the dangerous techniques were already safe when real scanning was added
- **Native Go executor registry pattern:** `Register/Lookup/LookupCleanup` in init() functions gave zero-overhead dispatch and automatic cleanup for all native techniques
- **Phase 18 re-audit approach:** Using phase to formally verify and close existing requirements (TECH-02/03/04) rather than building new code saved significant effort
- **Integration checker at milestone audit:** Caught the tier-field-not-propagated gap in elevation-skip results that no individual phase verification would have found
- **UAT sessions:** Retesting phase 08 blocked tests after Go build was fixed cleared all verification debt in one session

### What Was Inefficient
- **Phases 08-09 carried from v1.1:** These phases were completed during v1.1 timeframe but never archived — they rode along in v1.3 adding noise to milestone scope
- **REQUIREMENTS.md merge conflict markers:** Phase 16 execution left unresolved git conflict markers in REQUIREMENTS.md that persisted until milestone audit caught them
- **SAFE-02 traceability row stale text:** "In progress" text in traceability table while checkbox showed [x] — cosmetic inconsistency that should be caught by plan completion automation
- **Some SUMMARY.md one-liner fields empty or malformed:** Automated extraction returned "One-liner:" or empty for several summaries — template compliance needs enforcement

### Patterns Established
- **Tier classification pattern:** `Tier int` on Technique struct, propagated to ExecutionResult, rendered as badges in HTML report and web UI — reusable for any technique quality dimension
- **Scan confirmation gate:** `RequiresConfirmation` flag + channel-based pause/resume + web UI modal polling — reusable for any technique needing user consent before execution
- **Elevation gating:** `checkIsElevated()` at engine start with platform build tags (engine_windows.go / engine_other.go) — clean cross-platform admin detection
- **Native technique init() registration:** Each Go technique file registers itself via init() — no central manifest needed, new techniques are just new files

### Key Lessons
1. **Archive milestone phases promptly** — phases 08-09 should have been archived with v1.1, not carried into v1.3
2. **Git merge conflicts in .planning/ files need automated detection** — merge conflicts in REQUIREMENTS.md survived unnoticed through multiple phases
3. **Milestone audit integration checker is high-value** — it catches cross-phase wiring gaps that individual phase verifications miss (tier field in elevation-skip results)
4. **SUMMARY.md template compliance matters for automation** — empty one-liner fields break milestone stats extraction
5. **Human UAT sessions should be batched at milestone boundary** — testing phase 08 and 14 together in one session was efficient

### Cost Observations
- Model mix: ~85% sonnet (executors/verifiers), ~15% opus (orchestration, planning)
- Sessions: ~8 (planning + execution across phases, plus UAT and milestone completion)
- Notable: Phase 18 was a single-plan, single-wave execution that completed in under 10 minutes — lightweight phases for audit/verification work are efficient

---

## Milestone: v1.4 — PoC Technique Distribution

**Shipped:** 2026-04-11
**Phases:** 2 | **Plans:** 4 | **Commits:** 14

### What Was Built
- Distributed technique scheduling: `randomSlotsInWindow()` distributes techniques across configurable time windows with random jitter
- Phase 1 fires one technique per random slot, Phase 2 fires batches of 2-3 per slot
- UI window start/end inputs (08:00-17:00 default) replacing single hour inputs
- DayDigest accuracy tests for multi-technique Phase 1 days and Phase 2 step-count verification
- All existing scheduling tests documented with distributed scheduling correctness comments

### What Worked
- **Small milestone scope:** 2 phases, 4 plans — lean scope with clear goal made execution fast (under 12 hours wall clock)
- **Phase 19→20 dependency chain:** Engine rewrite first, test coverage second — clean separation of concerns
- **Wave 0 stubs:** Nyquist compliance stubs provided clear verify targets before implementation; stubs were un-skipped when real tests landed
- **Worktree isolation:** Executor agent merged master into worktree to acquire Phase 19 code — no conflict with orchestrator

### What Was Inefficient
- **Phase 20 executor needed to merge master:** Worktree was created from a base that didn't have Phase 19's code — required manual merge inside the agent. Worktree creation should be from HEAD, not from a stale base
- **ROADMAP.md plan counts stale:** Progress table showed "0/3 Not started" for Phase 19 even after completion — gsd-tools `phase complete` didn't update the progress table rows
- **Test comment D-03 tag instead of POC-04:** Cosmetic traceability inconsistency in test comments — POC-04 should have been referenced directly

### Patterns Established
- **randomSlotsInWindow helper:** Single function distributes N items across a time window — reusable for any future scheduling needs
- **afterCountClock test wrapper:** Counts After() calls on top of fakeClock — enables assertions about how many scheduling slots fired without inspecting internal state

### Key Lessons
1. **Small milestones are efficient** — 2 phases with clear scope executed and verified in a single session
2. **Worktree agents may need explicit merge of recent work** — when phases are sequential, the worktree base may not include the prior phase's commits
3. **Progress table rows should be updated by phase completion tooling** — manual table maintenance creates stale data

### Cost Observations
- Model mix: ~85% sonnet (executor/verifier agents), ~15% opus (orchestration)
- Sessions: 1 (single session for both phases + audit + completion)
- Notable: Smallest milestone yet — 2 phases, 14 commits, ~12 hours. Single-session milestone completion is achievable for focused scopes

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Phases | Plans | Key Change |
|-----------|--------|-------|------------|
| v1.0 | 7 | 17 | First GSD milestone — established SIEMCoverage pattern, Nyquist sign-off, QueryFn injection |
| v1.1 | 2 | 5 | Bug fixes & UI polish — GUID audit policy, version injection, inline error panels |
| v1.2 | 4 | 6 | PoC overhaul — Clock injection, DayDigest tracking, timeline calendar UI, scheduling tests |
| v1.3 | 7 | 16 | Realistic attack simulation — tier classification, native Go executor, safety infrastructure, network scanning |
| v1.4 | 2 | 4 | Distributed technique scheduling — randomSlotsInWindow, window config UI, DayDigest accuracy tests |

### Cumulative Quality

| Milestone | Test Functions | Packages Tested | Notes |
|-----------|---------------|-----------------|-------|
| v1.0 | 14 | 5/9 | playbooks blocked by Defender; executor/preparation/simlog/userstore have no testable logic |
| v1.1 | 20 | 5/9 | +6 tests from backend correctness and UI polish phases |
| v1.2 | 26 | 5/9 | +6 PoC scheduling tests via fake clock injection (poc_test.go) |
| v1.3 | 40+ | 7/10 | +14 tests: registry, executor dispatch, native techniques (T1046/T1018/T1057/T1482), playbook loader tier/events/cleanup |
| v1.4 | 42+ | 7/10 | +2 DayDigest accuracy tests (distributed counts, Phase 2 step count) in poc_schedule_test.go |

### Top Lessons (Verified Across Milestones)

1. Sign off VALIDATION.md at phase close — not as a batch at milestone close (v1.0, v1.2 — still not enforced)
2. Document environment constraints (Defender, no CGO) in CLAUDE.md before first phase
3. captureClock pattern eliminates race conditions in fake-timer tests (v1.2 — validated across 10+ tests)
4. Pre-populate schedules at start for immediate visibility (v1.2)
5. Don't skip milestone retrospectives — process observations get lost (v1.2)
6. Milestone audit integration checker catches cross-phase wiring gaps (v1.3 — found tier-field propagation gap)
7. Archive milestone phases promptly — don't carry completed phases into future milestones (v1.3)
8. Batch human UAT sessions at milestone boundary for efficiency (v1.3)
