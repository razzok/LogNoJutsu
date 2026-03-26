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

## Cross-Milestone Trends

### Process Evolution

| Milestone | Phases | Plans | Key Change |
|-----------|--------|-------|------------|
| v1.0 | 7 | 17 | First GSD milestone — established SIEMCoverage pattern, Nyquist sign-off, QueryFn injection |

### Cumulative Quality

| Milestone | Test Functions | Packages Tested | Notes |
|-----------|---------------|-----------------|-------|
| v1.0 | 14 | 5/9 | playbooks blocked by Defender; executor/preparation/simlog/userstore have no testable logic |

### Top Lessons (Verified Across Milestones)

1. Sign off VALIDATION.md at phase close — not as a batch at milestone close
2. Document environment constraints (Defender, no CGO) in CLAUDE.md before first phase
