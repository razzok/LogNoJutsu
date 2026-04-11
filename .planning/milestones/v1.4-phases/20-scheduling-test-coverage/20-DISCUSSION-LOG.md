# Phase 20: Scheduling Test Coverage - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-11
**Phase:** 20-scheduling-test-coverage
**Areas discussed:** Coverage gap assessment, DayDigest accuracy depth, Deterministic seeding, Existing test modernization

---

## Coverage Gap Assessment

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal gap closure | Add 1-2 tests for DayDigest count accuracy under distributed slots. Quick phase. | |
| Comprehensive test overhaul | Review every existing test for implicit fire-all-at-once assumptions. Add edge case tests. Thorough but larger. | ✓ |
| Skip phase entirely | Phase 19 tests cover the success criteria. Mark POC-04 complete and close milestone. | |

**User's choice:** Comprehensive test overhaul
**Notes:** User wants thorough review of all existing tests despite Phase 19 already covering most success criteria.

---

## DayDigest Accuracy Depth

| Option | Description | Selected |
|--------|-------------|----------|
| Count verification per slot | Test TechniqueCount and PassCount+FailCount with multiple techs/day. Test Phase 2 shows technique count not batch count. | ✓ |
| Full lifecycle per slot | Above plus status transitions and heartbeat updates between individual slots. | |

**User's choice:** Count verification per slot
**Notes:** Count accuracy is sufficient — no need for per-slot lifecycle tracking.

---

## Deterministic Seeding

| Option | Description | Selected |
|--------|-------------|----------|
| Structural assertions | Assert on observable behavior (After() count, execution count, DayDigest values). Resilient to algorithm changes. | ✓ |
| Inject fixed rand source | Add testable seam for rand.Source injection. More precise but couples tests to implementation. | |
| Both approaches | Structural for most, fixed seed for one timing-validation test. | |

**User's choice:** Structural assertions
**Notes:** Matches existing Phase 19 test pattern. No rand.Source injection into runPoC().

---

## Existing Test Modernization

| Option | Description | Selected |
|--------|-------------|----------|
| Audit and fix implicit assumptions | Review blockAt values, After() count assumptions, phase transition timing. Update false-pass tests. | ✓ |
| Leave existing, add new | Don't touch existing tests. Add new dedicated distributed tests only. | |
| Replace fragile tests entirely | Rewrite stop-signal and day-counter tests from scratch. Most thorough but highest effort. | |

**User's choice:** Audit and fix implicit assumptions
**Notes:** Update in place, don't replace. Document which tests exercise distributed code paths.

---

## Claude's Discretion

- Edge case selection for new tests
- Sub-test vs flat test function organization
- Test naming conventions

## Deferred Ideas

None — discussion stayed within phase scope
