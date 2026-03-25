# Phase 2: Code Structure & Test Coverage - Context

**Gathered:** 2026-03-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Refactor `server.go` package-level globals into a testable `Server` struct with method receivers. Add unit tests for engine state machine, HTTP handlers, and verification logic. Ensure `go test ./... -race` passes. This phase does NOT rename packages, add new techniques, or change HTTP API behavior.

</domain>

<decisions>
## Implementation Decisions

### Server Struct Refactor
- **D-01:** Define `type Server struct` in `internal/server/server.go` holding `eng *engine.Engine`, `registry *playbooks.Registry`, `users *userstore.Store`, `cfg Config` — replacing current package-level globals.
- **D-02:** Convert all handler functions to method receivers on `Server`: `func (s *Server) handleStatus(...)`, `func (s *Server) handleStart(...)`, etc.
- **D-03:** `server.Start(cfg Config)` stays as a package-level function — it constructs the `Server` internally and calls a private `listenAndServe()` method. `main.go` is unchanged.
- **D-04:** Route registration moves from the package-level `Start()` directly into a `Server` method (e.g., `s.registerRoutes(mux)`), removing the dependency on package-level state.

### Executor Abstraction for Testing
- **D-05:** Add a `RunnerFunc` type to `internal/engine/engine.go` — a function type with the same signature as the internal execution call. The `Engine` struct gains a `runner RunnerFunc` field; `nil` means use the real executor (production path unchanged).
- **D-06:** This pattern is intentionally consistent with the `QueryFn` injection already used in the verifier package (Phase 1). Tests inject a fake runner that returns synthetic `ExecutionResult` values without shelling out.
- **D-07:** A `SetRunner(fn RunnerFunc)` method (or field set directly in tests within the same package) provides the injection point. No changes to the public `New()` constructor signature.

### Test Coverage Scope
- **D-08:** Target: critical paths + `-race` detector. `go test ./... -race` must pass with no failures.
- **D-09:** Engine tests (in `internal/engine/engine_test.go`):
  - `TestEngineStart_transitionsToDiscovery` — verify phase transitions on Start
  - `TestEngineStop_abortsRun` — verify stop channel cooperative cancellation
  - `TestFilterByTactics` — table-driven, 4 cases (no filters, include-only, exclude-only, both)
  - `TestEngineRace` — run start+stop concurrently under `-race` to catch data races
- **D-10:** Handler tests (in `internal/server/server_test.go`, using `net/http/httptest`):
  - `TestHandleStatus_idle` — returns correct JSON when engine is idle
  - `TestHandleStatus_running` — returns correct phase when simulation active
  - `TestHandleStart_validConfig` — valid POST body triggers engine start
  - `TestHandleStop` — POST to /api/stop calls engine stop
  - `TestHandleTechniques` — GET returns technique list from registry
  - `TestAuthMiddleware_rejectsWrongPassword` — 401 on bad credentials
- **D-11:** Verifier tests (in `internal/verifier/verifier_test.go` or alongside verifier package):
  - `TestVerifier_pass` — matching event found → VerifPass status
  - `TestVerifier_fail` — no matching event → VerifFail status
  - `TestVerifier_notRun_WhatIf` — WhatIf mode → VerifNotRun status
- **D-12:** No additional test dependencies — stdlib `testing` + `net/http/httptest` only.

### Package Structure
- **D-13:** Keep the `playbooks` package name as-is. The ROADMAP mentions "techniques" as a target package name but renaming cascades through 6+ import sites with no behavioral benefit. The discrepancy is noted here and can be addressed in a future cleanup phase.
- **D-14:** The package split required by QUAL-02 is already largely satisfied by the existing `internal/` structure (engine/, server/, playbooks/, reporter/, verifier/). The main work is the server.go refactor (QUAL-01), not new package creation.

### Claude's Discretion
- Exact field order and visibility in `Server` struct (e.g., whether `cfg` is embedded vs named field)
- Whether `RunnerFunc` signature captures user profile / password args or uses a simpler `func(*Technique) ExecutionResult` — match what engine actually needs to inject
- Whether engine tests live in the same package (`package engine`) or a separate test package (`package engine_test`) — use `package engine` if internal access is needed for `SetRunner`
- Table-driven test case data for `TestFilterByTactics`

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Planning
- `.planning/REQUIREMENTS.md` — QUAL-01 through QUAL-05 are the requirements this phase satisfies
- `.planning/ROADMAP.md` — Phase 2 success criteria (5 items) define the acceptance bar

### Codebase
- `.planning/codebase/ARCHITECTURE.md` — Current package dependency graph (must not introduce circular imports)
- `.planning/codebase/STRUCTURE.md` — Package layout, entry points, all existing handler registrations
- `.planning/codebase/CONVENTIONS.md` — Naming conventions, error handling patterns, mutex usage
- `.planning/codebase/TESTING.md` — Zero-test baseline, identified gaps, recommended patterns, structural impediments (read entire file before planning)

### Key Source Files (must read before touching)
- `internal/server/server.go` — All globals to be refactored, all handlers to convert, route registration
- `internal/engine/engine.go` — Engine struct, Start/Stop/GetStatus, runInternal, RunnerFunc injection point
- `internal/verifier/verifier.go` — QueryFn pattern already in use — RunnerFunc must be consistent

### Prior Phase Context
- `.planning/phases/01-events-manifest-verification-engine/01-CONTEXT.md` — QueryFn injection decision (D-05 in Phase 1) that this phase's RunnerFunc pattern mirrors

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `sync.RWMutex` embedded as first field in `Engine` and `Store` — `Server` struct should follow the same pattern if it needs mutex protection
- `writeJSON` / `writeError` helpers already exist in server.go — keep them as package-level helpers or move to Server methods (Claude's discretion)
- `net/http/httptest` is stdlib — no new imports needed for handler tests

### Established Patterns
- Constructor: `func New(...) *Type` — use `engine.New(registry, users)` call site in `server.Start()`
- Method receivers: short 1-2 letter matching type initial — `(s *Server)` for Server methods
- Section separator comments: `// ── Section ──────────────────────────────────────────` in large files
- Phase 1 QueryFn injection: `type QueryFn func(...) ([]VerifiedEvent, error)` on verifier struct — mirror this for RunnerFunc on Engine

### Integration Points
- `cmd/lognojutsu/main.go` calls `server.Start(cfg)` — this call signature MUST NOT change
- `engine.New(registry, users)` is called inside `server.Start()` — stays there post-refactor
- All 15 API routes registered in `server.go` `Start()` — move to `s.registerRoutes(mux)` method

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 02-code-structure-test-coverage*
*Context gathered: 2026-03-24*
