---
phase: 02-code-structure-test-coverage
plan: 01
subsystem: server, engine
tags: [refactor, testability, struct, dependency-injection]
dependency_graph:
  requires: []
  provides: [Server struct with method receivers, RunnerFunc injection on Engine]
  affects: [internal/server/server.go, internal/engine/engine.go, cmd/lognojutsu/main.go]
tech_stack:
  added: []
  patterns: [dependency injection via RunnerFunc (mirrors QueryFn in verifier), Server struct encapsulating HTTP dependencies]
key_files:
  created: []
  modified:
    - internal/server/server.go
    - internal/engine/engine.go
    - cmd/lognojutsu/main.go
decisions:
  - Server struct holds eng/registry/users/cfg — all HTTP handlers are method receivers on Server
  - Start(c Config) remains package-level — main.go call site unchanged
  - writeJSON and writeError remain package-level helpers (no state dependency)
  - RunnerFunc nil-default pattern mirrors verifier QueryFn — no change to New() or production path
metrics:
  duration: "2min"
  completed_date: "2026-03-25"
  tasks_completed: 2
  files_modified: 3
---

# Phase 02 Plan 01: Server Struct Refactor and RunnerFunc Injection Summary

Server.go package-level globals replaced by Server struct with method receivers; Engine gains RunnerFunc injection point using the nil-default QueryFn pattern.

## What Was Built

### Task 1: Fix go vet warning and add RunnerFunc to Engine

- `cmd/lognojutsu/main.go`: Changed `fmt.Println(banner)` to `fmt.Print(banner)` — banner constant already ends with `\n`, so Println was causing a go vet warning about a newline-terminated argument.
- `internal/engine/engine.go`: Added `RunnerFunc` type (mirrors `verifier.QueryFn` pattern per D-06), added `runner RunnerFunc` field to Engine struct, added `SetRunner(fn RunnerFunc)` method, modified `runTechnique` to call the injected runner when non-nil (WhatIf → runner != nil → RunCleanup → RunAs chain).

### Task 2: Refactor server.go — replace globals with Server struct

- Removed `var (eng, registry, users, cfg)` block entirely — zero package-level mutable state remains.
- Added `type Server struct` holding `eng *engine.Engine`, `registry *playbooks.Registry`, `users *userstore.Store`, `cfg Config`.
- `Start(c Config) error` remains package-level and constructs Server internally — main.go call site is unchanged.
- Added `func (s *Server) registerRoutes(mux *http.ServeMux)` method.
- Converted `authMiddleware` and all 15 handlers to `(s *Server)` method receivers.
- `writeJSON` and `writeError` remain package-level helpers (they reference no state).

## Verification Results

- `go vet ./...` — zero warnings (both tasks)
- `go build ./cmd/lognojutsu/` — succeeds without errors
- `go test ./internal/verifier/...` — PASS (cached)
- `go test ./internal/playbooks/...` — PASS (cached)
- `go test ./internal/reporter/...` — PASS (cached)
- `grep -n "^var eng|^var registry|^var users|^var cfg" internal/server/server.go` — returns empty (no globals)

Note: `-race` flag was skipped on Windows — CGO_ENABLED=0 prevents race detector. Tests ran without -race and passed.

## Commits

| Task | Commit | Message |
|------|--------|---------|
| 1    | 739cc9c | feat(02-01): fix go vet warning and add RunnerFunc injection to Engine |
| 2    | 2a0ea7a | refactor(02-01): replace server.go package globals with Server struct and method receivers |

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — no stubs, placeholders, or hardcoded empty values introduced.

## Self-Check: PASSED

- `internal/server/server.go` — exists and contains `type Server struct`, `func Start(c Config) error`, `func (s *Server) registerRoutes`, `func (s *Server) handleStatus`, `func (s *Server) handleStart`, `func (s *Server) handleStop`, `func (s *Server) handleUsers`
- `internal/engine/engine.go` — exists and contains `type RunnerFunc func`, `runner RunnerFunc` field, `func (e *Engine) SetRunner`, `} else if e.runner != nil {`
- `cmd/lognojutsu/main.go` — exists and contains `fmt.Print(banner)`
- Commit 739cc9c — present in git log
- Commit 2a0ea7a — present in git log
