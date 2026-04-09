# Phase 13: PoC Scheduling Tests - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-09
**Phase:** 13-poc-scheduling-tests
**Areas discussed:** Coverage depth, Stop-signal scenarios, Test organization

---

## Coverage Depth

| Option | Description | Selected |
|--------|-------------|----------|
| Requirements only | Fill gaps in TEST-02, TEST-03, TEST-04. Existing tests cover basics. | ✓ |
| Requirements + edge cases | Also add 0-day gap, single-day PoC, large day counts, concurrent stop. | |
| Comprehensive | Full coverage including boundary, race conditions, negative tests. | |

**User's choice:** Requirements only (Recommended)
**Notes:** Existing tests already provide baseline coverage — phase focuses on filling the 3 specific requirement gaps.

---

## Stop-Signal Scenarios

| Option | Description | Selected |
|--------|-------------|----------|
| During day wait | Stop during nextOccurrenceOfHour sleep. Most common real-world scenario. | ✓ |
| Between phase transitions | Stop during Phase1→Gap or Gap→Phase2 boundary. | ✓ |
| During gap days | Stop during gap phase with no techniques, just sleep. | ✓ |
| Immediate after start | Stop before first day executes. Quick cancellation edge case. | ✓ |

**User's choice:** All four scenarios
**Notes:** User wants comprehensive stop-signal coverage across all timing windows.

---

## Test Organization

| Option | Description | Selected |
|--------|-------------|----------|
| Separate poc_test.go | New file internal/engine/poc_test.go. engine_test.go already 900+ lines. | ✓ |
| Add to engine_test.go | Keep everything in one file. Simpler but larger. | |

**User's choice:** Separate poc_test.go (Recommended)
**Notes:** Existing helpers accessible from same package. Keeps PoC scheduling tests grouped.

---

## Claude's Discretion

- Test function naming and subtesting structure
- Which clock wrapper patterns to reuse vs create
- Table-driven vs individual test functions

## Deferred Ideas

None
