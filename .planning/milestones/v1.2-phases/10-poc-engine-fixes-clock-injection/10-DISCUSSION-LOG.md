# Phase 10: PoC Engine Fixes & Clock Injection - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-08
**Phase:** 10-poc-engine-fixes-clock-injection
**Areas discussed:** Clock injection design, Day counter fix, Log separators, German string cleanup

---

## Clock Injection Design

### Q1: How should the clock be injected into the Engine?

| Option | Description | Selected |
|--------|-------------|----------|
| Clock interface | Define a Clock interface with Now() and After(d) methods. Engine gets a Clock field; production uses realClock, tests supply a fakeClock. | ✓ |
| Function fields on Engine | Add NowFn and AfterFn fields directly on Engine struct, like QueryFn pattern in verifier. | |
| Function fields on Config | Add NowFn/AfterFn to engine.Config struct. Keeps Engine constructor unchanged. | |

**User's choice:** Clock interface
**Notes:** Clean, testable, matches Go idioms. Previewed the interface shape and approved.

### Q2: Where should the Clock interface be defined?

| Option | Description | Selected |
|--------|-------------|----------|
| In engine package | Define Clock, realClock in engine package. Only engine needs it currently. | ✓ |
| Own package internal/clock | Separate package for reuse. More structure but premature. | |

**User's choice:** In engine package
**Notes:** No other package needs it yet — keep it local.

---

## Day Counter Fix

### Q1: How should the global day counter work across Phase1 → Gap → Phase2?

| Option | Description | Selected |
|--------|-------------|----------|
| Running counter | Single globalDay variable starts at 1, increments at each day boundary. PoCDay always shows globalDay. | ✓ |
| Computed from section offsets | Calculate globalDay as offset + localDay. Same result, computed instead of accumulated. | |

**User's choice:** Running counter
**Notes:** Previewed the code pattern and approved. Simple and correct.

---

## Log Separators

### Q1: Where should simlog.Phase() separators appear in the PoC flow?

| Option | Description | Selected |
|--------|-------------|----------|
| At each phase start | simlog.Phase() right after each setPhase() call. Three separators total. | ✓ |
| Inside setPhase() | Move simlog.Phase() into setPhase() so it always emits on transition. Would also affect normal run(). | |
| At phase + daily boundaries | Separators at each phase start AND each new day. More granular but noisy. | |

**User's choice:** At each phase start
**Notes:** Previewed the three call sites. Labels: "PoC Phase 1: Discovery", "PoC Gap", "PoC Phase 2: Attack".

---

## German String Cleanup

### Q1: What English format should CurrentStep strings use?

| Option | Description | Selected |
|--------|-------------|----------|
| Descriptive | Full phrases: "PoC Phase 1 — Day N of M — Waiting until HH:00" | ✓ |
| Compact | Short: "Phase1 DN/M — next HH:00" | |

**User's choice:** Descriptive
**Notes:** Previewed all three string formats and approved the descriptive style.

---

## Claude's Discretion

- Whether Clock lives in engine.go or a separate clock.go
- Whether to use functional options or direct field assignment for test clock injection
- Internal variable naming for the running day counter

## Deferred Ideas

None — discussion stayed within phase scope.
