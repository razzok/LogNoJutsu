# Testing

**Analysis Date:** 2026-03-24

## Test Coverage

**Current state: zero test files exist.**

A search of the entire repository for `*_test.go` files returns no results. There are no unit tests, integration tests, or benchmark tests of any kind. The project has never had automated tests written for it.

**Files with zero test coverage:**
- `cmd/lognojutsu/main.go` — entry point, flag parsing, startup
- `internal/engine/engine.go` — core simulation lifecycle, phase transitions, user rotation, PoC scheduling
- `internal/executor/executor.go` — command execution, RunAs, cleanup logic
- `internal/playbooks/loader.go` — YAML loading, registry construction, tactic deduplication
- `internal/playbooks/types.go` — data structures (no logic to test directly)
- `internal/preparation/preparation.go` — PowerShell/auditpol invocation, Sysmon install
- `internal/reporter/reporter.go` — JSON/HTML report generation, tactic stats, template rendering
- `internal/server/server.go` — HTTP routing, auth middleware, all API handlers
- `internal/simlog/simlog.go` — structured logging, in-memory entry store, file output
- `internal/userstore/userstore.go` — profile CRUD, DPAPI encryption, user discovery

## Test Types

**Unit tests:** Not present.

**Integration tests:** Not present.

**End-to-end tests:** Not present.

**Benchmark tests:** Not present.

**Manual testing** is the only current verification method — the binary is run against a live Windows environment and the UI is exercised interactively.

## Frameworks and Tools

**Test runner:** None configured.

**go.mod declares no test dependencies:**
```
module lognojutsu

go 1.26.1

require gopkg.in/yaml.v3 v3.0.1 // indirect
```

No `testify`, `gomock`, `httptest` (stdlib), or any other test library is present.

**Build/lint tooling:** No `Makefile`, no `.golangci.yml`, no `Taskfile`, no CI pipeline configuration (`.github/`, `.gitlab-ci.yml`, etc.) exists in the repository.

## Testing Patterns

Because no tests exist, this section documents **patterns that should be used** when tests are written, based on the codebase's existing structure.

**Standard library `testing` package** is the natural fit — no third-party dependency required.

**Table-driven tests** suit the pure functions that do exist:

```go
// Example pattern for registry/loader functions
func TestGetAllTactics(t *testing.T) {
    cases := []struct {
        name     string
        input    []*Technique
        expected []string
    }{
        {"empty", nil, nil},
        {"deduplicates", [...], [...]},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) { ... })
    }
}
```

**`net/http/httptest`** (stdlib) should be used for handler tests — no external dependency needed:
```go
rec := httptest.NewRecorder()
req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
handleStatus(rec, req)
```

**Interface extraction for executor mocking:** `executor.RunAs` and `executor.Run` are currently plain functions, not methods on an interface. To unit-test `engine.Engine` without executing real OS commands, an `Executor` interface would need to be introduced.

**`embed.FS` testing:** `playbooks.LoadEmbedded()` can be tested directly in a `_test.go` file within the `playbooks` package — the embedded FS is available to test binaries in the same package.

## Gaps and Weaknesses

**Critical gaps (high risk):**

1. **Engine state machine — untested**
   - `internal/engine/engine.go` contains the most complex logic in the project: phase transitions, concurrent access via `sync.RWMutex`, stop-channel signalling, PoC multi-day scheduling, user rotation, tactic filtering
   - Race conditions in `runTechnique` / `pickUser` / `setPhase` could corrupt `Status` — no `-race` tests have ever been run

2. **HTTP API handlers — untested**
   - All handlers in `internal/server/server.go` rely on global variables (`eng`, `registry`, `users`, `cfg`)
   - Method-checking logic, JSON decode error paths, auth middleware, and response shapes are unverified
   - `handleUsersDiscover` silently concatenates partial errors; `handleReport` reads a path from engine state — both have failure branches never exercised by tests

3. **Playbook loader — testable but untested**
   - `LoadEmbedded()` has straightforward error paths (malformed YAML, missing fields) that are easy to test but currently not covered
   - `GetAllTactics()` uses a manual bubble sort — correctness is untested

4. **Reporter HTML template — untested**
   - `saveHTML()` in `internal/reporter/reporter.go` constructs a `text/template` inline; template parse errors or data-binding panics would only be caught at runtime
   - The `truncate`, `fmtTime`, and `tacticColor` template functions have no unit tests

5. **Userstore CRUD and DPAPI fallback — untested**
   - `Store.Add()` has a silent fallback to `"PLAIN:"` prefix when DPAPI encryption fails — this security-relevant fallback path is never tested
   - `DecryptPassword()` `PLAIN:` prefix stripping is correct but untested
   - File permission (`0600`) on the persisted JSON is set but never asserted

6. **`filterByTactics` logic — testable, untested**
   - `internal/engine/engine.go:filterByTactics()` handles four cases (no filters, include-only, exclude-only, both) — pure function with no side effects, straightforward to table-test but currently untested

7. **No `-race` detector runs**
   - The engine spawns goroutines (`go e.run()`) and uses mutex-protected shared state. Without `go test -race`, data races are undetectable. This is the highest-risk gap given the concurrent architecture.

**Low-risk but missing:**
- `simlog` functions (`Start`, `Stop`, `Phase`, `TechStart`, etc.) are easy to unit test (check `GetEntries()` after each call) but are untested
- `preparation.go` helper `fileExists()` is trivially testable
- `reporter.fmtTime()` is a pure function with straightforward edge cases (invalid RFC3339 input returns the original string)

**Structural impediment to testing:**
- `server.go` uses package-level global variables (`var eng *engine.Engine`, `var registry *playbooks.Registry`, etc.) rather than injecting dependencies through a struct. This makes handler tests require global state mutation, which is fragile and not goroutine-safe in parallel test runs. Refactoring to a `Server` struct with method receivers would unblock clean handler testing.
