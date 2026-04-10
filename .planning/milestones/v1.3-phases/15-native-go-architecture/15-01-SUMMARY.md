---
phase: 15-native-go-architecture
plan: "01"
subsystem: executor/native
tags: [native-go, registry, executor, tdd]
dependency_graph:
  requires: []
  provides: [native.Register, native.Lookup, native.LookupCleanup, type-go-dispatch]
  affects: [internal/executor, internal/native]
tech_stack:
  added: [internal/native package]
  patterns: [registry pattern, sync.RWMutex, TDD red-green]
key_files:
  created:
    - internal/native/registry.go
    - internal/native/registry_test.go
    - internal/executor/executor_go_test.go
  modified:
    - internal/executor/executor.go
decisions:
  - "NativeFunc signature is func() (NativeResult, error) — no technique argument needed since Go functions can close over their own state"
  - "type:go RunAs emits a log note and proceeds as current user — Go functions cannot change user context"
  - "RunCleanupOnly handles Go techniques first (early return) — keeps shell cleanup path clean"
metrics:
  duration_seconds: 139
  completed_date: "2026-04-09"
  tasks_completed: 2
  tasks_total: 2
  files_created: 3
  files_modified: 1
requirements: [ARCH-01]
---

# Phase 15 Plan 01: Native Go Registry & Executor Dispatch Summary

**One-liner:** native.Register/Lookup/LookupCleanup registry with sync.RWMutex + type:go dispatch in executor without spawning child processes.

## What Was Built

### Task 1: Native Registry Package (8b96d24)

Created `internal/native/registry.go` with:
- `NativeResult` struct (Output, ErrorOutput, Success)
- `NativeFunc` type: `func() (NativeResult, error)`
- `CleanupFunc` type: `func() error`
- `Register(id, fn, cleanup)` — concurrent-safe write via sync.RWMutex
- `Lookup(id)` — returns NativeFunc or nil
- `LookupCleanup(id)` — returns CleanupFunc or nil

6 unit tests in `registry_test.go`: register/lookup, cleanup lifecycle, overwrite semantics, missing lookup returns nil (no panic).

### Task 2: Executor Go Dispatch (c143361)

Modified `internal/executor/executor.go`:
- Added `lognojutsu/internal/native` import
- In `runInternal()`: type:go dispatch branch before shell execution — calls `native.Lookup(t.ID)`, maps NativeResult to ExecutionResult, handles unregistered case with descriptive error
- RunAs log note emitted when profile.UserType != "current" for type:go techniques
- In `RunWithCleanup()`: if/else-if chain — Go techniques use `native.LookupCleanup()` in defer, shell techniques use existing Cleanup string path
- In `RunCleanupOnly()`: Go cleanup path with early return

6 new tests in `executor_go_test.go`: dispatch, unregistered technique error, function returning error, cleanup fires in RunWithCleanup, nil cleanup when not registered, RunAs log note execution.

## Verification

```
go test ./internal/native/... -v     # 6 tests pass
go test ./internal/executor/... -v   # all tests pass including 6 new TestGo* tests
go test ./...                        # full regression passes (all packages)
go build ./...                       # clean compilation
```

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — all functionality fully wired. No placeholder returns or TODO code paths.

## Self-Check: PASSED

- internal/native/registry.go: FOUND
- internal/native/registry_test.go: FOUND
- internal/executor/executor.go: modified with native import and type:go dispatch
- internal/executor/executor_go_test.go: FOUND
- Commit 8b96d24: FOUND
- Commit c143361: FOUND
