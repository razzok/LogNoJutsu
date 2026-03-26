# Phase 2: Code Structure & Test Coverage — Research

**Researched:** 2026-03-25
**Domain:** Go refactoring, dependency injection, HTTP handler testing with net/http/httptest, data-race detection
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Server Struct Refactor**
- D-01: Define `type Server struct` in `internal/server/server.go` holding `eng *engine.Engine`, `registry *playbooks.Registry`, `users *userstore.Store`, `cfg Config` — replacing current package-level globals.
- D-02: Convert all handler functions to method receivers on `Server`: `func (s *Server) handleStatus(...)`, `func (s *Server) handleStart(...)`, etc.
- D-03: `server.Start(cfg Config)` stays as a package-level function — it constructs the `Server` internally and calls a private `listenAndServe()` method. `main.go` is unchanged.
- D-04: Route registration moves from the package-level `Start()` directly into a `Server` method (e.g., `s.registerRoutes(mux)`), removing the dependency on package-level state.

**Executor Abstraction for Testing**
- D-05: Add a `RunnerFunc` type to `internal/engine/engine.go` — a function type with the same signature as the internal execution call. The `Engine` struct gains a `runner RunnerFunc` field; `nil` means use the real executor (production path unchanged).
- D-06: This pattern is intentionally consistent with the `QueryFn` injection already used in the verifier package (Phase 1). Tests inject a fake runner that returns synthetic `ExecutionResult` values without shelling out.
- D-07: A `SetRunner(fn RunnerFunc)` method (or field set directly in tests within the same package) provides the injection point. No changes to the public `New()` constructor signature.

**Test Coverage Scope**
- D-08: Target: critical paths + `-race` detector. `go test ./... -race` must pass with no failures.
- D-09: Engine tests in `internal/engine/engine_test.go`:
  - `TestEngineStart_transitionsToDiscovery`
  - `TestEngineStop_abortsRun`
  - `TestFilterByTactics` (table-driven, 4 cases)
  - `TestEngineRace`
- D-10: Handler tests in `internal/server/server_test.go` (using `net/http/httptest`):
  - `TestHandleStatus_idle`
  - `TestHandleStatus_running`
  - `TestHandleStart_validConfig`
  - `TestHandleStop`
  - `TestHandleTechniques`
  - `TestAuthMiddleware_rejectsWrongPassword`
- D-11: Verifier tests in `internal/verifier/verifier_test.go`:
  - `TestVerifier_pass`
  - `TestVerifier_fail`
  - `TestVerifier_notRun_WhatIf`
- D-12: No additional test dependencies — stdlib `testing` + `net/http/httptest` only.

**Package Structure**
- D-13: Keep the `playbooks` package name as-is. Renaming is deferred.
- D-14: The package split for QUAL-02 is already satisfied by the existing `internal/` structure. Main work is the server.go refactor (QUAL-01), not new package creation.

### Claude's Discretion
- Exact field order and visibility in `Server` struct (e.g., whether `cfg` is embedded vs named field)
- Whether `RunnerFunc` signature captures user profile / password args or uses a simpler `func(*Technique) ExecutionResult` — match what engine actually needs to inject
- Whether engine tests live in the same package (`package engine`) or a separate test package (`package engine_test`) — use `package engine` if internal access is needed for `SetRunner`
- Table-driven test case data for `TestFilterByTactics`

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| QUAL-01 | Package-level globals in server.go refactored to a struct (enables testable handlers) | D-01 through D-04 define exact refactor shape; current globals identified in §Standard Stack |
| QUAL-02 | Codebase split into logical packages (engine, server, techniques, reporter) | D-14: already satisfied by existing `internal/` structure — confirmed by codebase scan |
| QUAL-03 | Unit tests for simulation engine state machine (start, stop, phase transitions) | D-09 specifies exact test names; RunnerFunc injection (D-05/D-06) is the enabler |
| QUAL-04 | Unit tests for HTTP handlers covering key API endpoints | D-10 specifies exact test names; httptest.NewRecorder is the mechanism |
| QUAL-05 | Unit tests for verification logic (event log querying, pass/fail determination) | Already partially complete — verifier_test.go exists and passes; D-11 names remaining gaps |
</phase_requirements>

---

## Summary

Phase 2 is a pure refactoring + test-writing phase. No new features, no new packages. The codebase analysis shows the package structure (`internal/engine`, `internal/server`, etc.) already satisfies QUAL-02. The real work is two tasks: (1) converting `server.go` from package-level globals to a `Server` struct so handlers become method receivers testable with `httptest.NewRecorder`, and (2) adding test files for engine and server packages — the verifier and playbooks tests already exist from Phase 1.

**The single most important pre-coding insight:** `internal/verifier/verifier_test.go` and `internal/playbooks/types_test.go` already exist and pass. The `internal/reporter/reporter_test.go` also passes. This means `go test ./... -race` will pass for those packages immediately. The only packages that need new test files are `internal/engine` and `internal/server`. There is also one pre-existing `go vet` failure in `cmd/lognojutsu/main.go` (redundant `\n` in `fmt.Println(banner)`) that blocks `go test ./... -race` from a clean run — this must be fixed as Wave 0 work.

**Primary recommendation:** Fix the `fmt.Println` vet error first, implement the `Server` struct refactor second (unblocks handler tests), then write engine tests with `RunnerFunc` injection.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `testing` | stdlib (go1.26.1) | Test runner, `t.Run`, `t.Helper`, `t.TempDir` | Standard Go test package — no external dependency |
| `net/http/httptest` | stdlib (go1.26.1) | `NewRecorder()`, `NewRequest()` for handler tests | Standard HTTP testing tool — D-12 mandates stdlib only |
| `sync` | stdlib | `RWMutex` for `Server` struct if needed | Already used in `Engine` and `Store` — same pattern |

### No New Dependencies
The go.mod currently has only one require: `gopkg.in/yaml.v3 v3.0.1`. This phase adds zero new dependencies — all tools (`testing`, `httptest`, `sync`) are stdlib. This is a hard constraint from D-12.

**Version verification:** Go 1.26.1 confirmed on machine (`go version go1.26.1 windows/amd64`). No minimum version concerns.

---

## Architecture Patterns

### Recommended Project Structure (after phase)
```
internal/
├── engine/
│   ├── engine.go          # Engine struct + RunnerFunc field added
│   └── engine_test.go     # NEW: 4 tests (D-09)
├── server/
│   ├── server.go          # Server struct refactor (D-01 to D-04)
│   └── server_test.go     # NEW: 6 tests (D-10)
├── verifier/
│   ├── verifier.go        # UNCHANGED
│   └── verifier_test.go   # EXISTS — passes; may add D-11 named tests
├── playbooks/
│   ├── types.go           # UNCHANGED
│   ├── types_test.go      # EXISTS — passes
│   └── loader.go          # UNCHANGED
└── reporter/
    ├── reporter.go        # UNCHANGED
    └── reporter_test.go   # EXISTS — passes
```

### Pattern 1: Server Struct with Method Receivers

**What:** Replace 4 package-level globals (`eng`, `registry`, `users`, `cfg`) with a struct that holds the same fields as named fields. All handler functions become method receivers.

**Before (current state):**
```go
// server.go — package-level globals
var (
    eng      *engine.Engine
    registry *playbooks.Registry
    users    *userstore.Store
    cfg      Config
)

func handleStatus(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, eng.GetStatus())
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if cfg.Password != "" { ... }
        next(w, r)
    }
}
```

**After (D-01 through D-04):**
```go
// server.go — Server struct
type Server struct {
    mu       sync.RWMutex    // only if Server itself needs mutex protection
    eng      *engine.Engine
    registry *playbooks.Registry
    users    *userstore.Store
    cfg      Config
}

func Start(c Config) error {
    s := &Server{}
    s.cfg = c
    // ... load registry, users, engine ...
    s.eng = engine.New(s.registry, s.users)
    mux := http.NewServeMux()
    s.registerRoutes(mux)
    addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
    return http.ListenAndServe(addr, mux)
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
    staticFS, _ := fs.Sub(staticFiles, "static")
    mux.Handle("/", http.FileServer(http.FS(staticFS)))
    mux.HandleFunc("/api/status", s.authMiddleware(s.handleStatus))
    // ... all 15 routes
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, s.eng.GetStatus())
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if s.cfg.Password != "" { ... }
        next(w, r)
    }
}
```

**Why:** `Server` struct makes the `eng`, `registry`, `users`, and `cfg` dependencies explicit. Tests can construct a `Server` with mock/real values and call handlers directly via `httptest.NewRecorder` — no global mutation.

### Pattern 2: httptest.NewRecorder for Handler Tests

**What:** Use `net/http/httptest` to test HTTP handlers without starting a real server.

```go
// server_test.go
func newTestServer(t *testing.T) *Server {
    t.Helper()
    reg, err := playbooks.LoadEmbedded()
    if err != nil {
        t.Fatalf("load registry: %v", err)
    }
    us, _ := userstore.Load()   // empty store OK for most tests
    return &Server{
        cfg:      Config{Password: ""},
        registry: reg,
        users:    us,
        eng:      engine.New(reg, us),
    }
}

func TestHandleStatus_idle(t *testing.T) {
    s := newTestServer(t)
    rec := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
    s.handleStatus(rec, req)
    if rec.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", rec.Code)
    }
    var status engine.Status
    if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
        t.Fatalf("decode: %v", err)
    }
    if status.Phase != engine.PhaseIdle {
        t.Errorf("expected idle phase, got %q", status.Phase)
    }
}

func TestAuthMiddleware_rejectsWrongPassword(t *testing.T) {
    s := newTestServer(t)
    s.cfg.Password = "secret"
    rec := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
    req.SetBasicAuth("", "wrong")
    handler := s.authMiddleware(s.handleStatus)
    handler(rec, req)
    if rec.Code != http.StatusUnauthorized {
        t.Errorf("expected 401, got %d", rec.Code)
    }
}
```

**Key insight:** `httptest.NewRecorder()` captures the HTTP response without a network. `httptest.NewRequest()` creates an `*http.Request` suitable for direct handler invocation. Both are stdlib.

### Pattern 3: RunnerFunc Injection for Engine Tests

**What:** Add a `RunnerFunc` type and a `runner` field to `Engine`. When `runner` is nil, use real executor. When non-nil, call `runner` instead. Mirror the `QueryFn` pattern already in `verifier.go`.

```go
// engine.go additions

// RunnerFunc abstracts technique execution for testability.
// Receives technique, user profile (nil = current user), and password.
// Returns an ExecutionResult without shelling out.
type RunnerFunc func(t *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult

// In Engine struct:
type Engine struct {
    mu     sync.RWMutex
    // ... existing fields ...
    runner RunnerFunc  // nil = use real executor
}

// SetRunner injects a fake runner for tests. Call before Engine.Start().
func (e *Engine) SetRunner(fn RunnerFunc) {
    e.runner = fn
}

// In runTechnique(), replace executor calls:
func (e *Engine) runTechnique(t *playbooks.Technique) {
    profile, password := e.pickUser()
    // ...
    var result playbooks.ExecutionResult
    if e.cfg.WhatIf {
        // ... existing WhatIf path unchanged ...
    } else if e.runner != nil {
        result = e.runner(t, profile, password)
    } else if e.cfg.RunCleanup {
        result = executor.RunWithCleanup(t, profile, password)
    } else {
        result = executor.RunAs(t, profile, password)
        // ...
    }
    // ... rest unchanged ...
}
```

**Engine test pattern:**
```go
// engine_test.go — package engine (internal access needed for SetRunner)
package engine

func makeRegistry() *playbooks.Registry {
    reg, err := playbooks.LoadEmbedded()
    if err != nil {
        panic(err)
    }
    return reg
}

func fakeRunner(t *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
    return playbooks.ExecutionResult{
        TechniqueID:        t.ID,
        TechniqueName:      t.Name,
        Success:            true,
        VerificationStatus: playbooks.VerifNotRun,
    }
}

func TestEngineStart_transitionsToDiscovery(t *testing.T) {
    reg := makeRegistry()
    us, _ := userstore.Load()
    e := New(reg, us)
    e.SetRunner(fakeRunner)
    cfg := Config{WhatIf: false}
    if err := e.Start(cfg); err != nil {
        t.Fatalf("Start: %v", err)
    }
    // Give goroutine a moment to transition
    time.Sleep(50 * time.Millisecond)
    status := e.GetStatus()
    // Phase should be discovery or done (fast with fake runner)
    if status.Phase == PhaseIdle {
        t.Errorf("expected non-idle phase after Start, got idle")
    }
}

func TestEngineStop_abortsRun(t *testing.T) {
    reg := makeRegistry()
    us, _ := userstore.Load()
    e := New(reg, us)
    e.SetRunner(func(t *playbooks.Technique, p *userstore.UserProfile, pw string) playbooks.ExecutionResult {
        time.Sleep(500 * time.Millisecond) // slow runner to ensure stop races it
        return playbooks.ExecutionResult{TechniqueID: t.ID, Success: true}
    })
    if err := e.Start(Config{}); err != nil {
        t.Fatalf("Start: %v", err)
    }
    time.Sleep(10 * time.Millisecond)
    e.Stop()
    time.Sleep(600 * time.Millisecond) // wait for abort path
    status := e.GetStatus()
    if status.Phase != PhaseAborted && status.Phase != PhaseDone {
        t.Errorf("expected aborted/done after Stop, got %q", status.Phase)
    }
}

func TestFilterByTactics(t *testing.T) {
    techniques := []*playbooks.Technique{
        {ID: "T1", Tactic: "discovery"},
        {ID: "T2", Tactic: "execution"},
        {ID: "T3", Tactic: "persistence"},
    }
    cases := []struct {
        name     string
        included []string
        excluded []string
        wantIDs  []string
    }{
        {"no filters", nil, nil, []string{"T1", "T2", "T3"}},
        {"include only", []string{"discovery"}, nil, []string{"T1"}},
        {"exclude only", nil, []string{"execution"}, []string{"T1", "T3"}},
        {"both", []string{"discovery", "execution"}, []string{"execution"}, []string{"T1"}},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            e := &Engine{cfg: Config{
                IncludedTactics: tc.included,
                ExcludedTactics: tc.excluded,
            }}
            got := e.filterByTactics(techniques)
            if len(got) != len(tc.wantIDs) {
                t.Fatalf("got %d techniques, want %d", len(got), len(tc.wantIDs))
            }
            for i, id := range tc.wantIDs {
                if got[i].ID != id {
                    t.Errorf("[%d] got %q, want %q", i, got[i].ID, id)
                }
            }
        })
    }
}

func TestEngineRace(t *testing.T) {
    reg := makeRegistry()
    us, _ := userstore.Load()
    e := New(reg, us)
    e.SetRunner(fakeRunner)
    var wg sync.WaitGroup
    wg.Add(2)
    go func() {
        defer wg.Done()
        _ = e.Start(Config{})
    }()
    go func() {
        defer wg.Done()
        time.Sleep(5 * time.Millisecond)
        e.Stop()
    }()
    wg.Wait()
    time.Sleep(100 * time.Millisecond)
}
```

**Why `package engine` not `package engine_test`:** `filterByTactics` and `SetRunner` are on the `Engine` struct. If `SetRunner` is unexported (field direct assignment), tests must be in `package engine`. If exported, either package works. D-07 is ambiguous — use `package engine` to keep flexibility.

### Pattern 4: Verifier Test Naming Alignment

**What:** Phase 1 already wrote verifier tests. D-11 names three specific tests (`TestVerifier_pass`, `TestVerifier_fail`, `TestVerifier_notRun_WhatIf`). Existing tests cover pass and fail under different names (`TestDetermineStatus`, `TestNotExecutedVsEventsMissing`). The plan should add the three D-11 named tests or confirm the existing coverage is sufficient.

**Current verifier test coverage (already passing):**
- `TestDetermineStatus` — covers pass/fail/notRun (maps to D-11's pass + fail + notRun)
- `TestNotExecutedVsEventsMissing` — covers notExecuted vs fail
- `TestVerifyAllFound` — additional pass coverage
- `TestQueryCountMock` — argument pass-through verification

**Recommendation:** The existing tests satisfy D-11's intent. Add thin wrapper tests with D-11 names to satisfy naming requirement, or confirm that existing tests satisfy the requirement and document the mapping. Either is acceptable.

### Anti-Patterns to Avoid

- **Parallel test mutation of package globals:** The old `var eng *engine.Engine` pattern means tests that mutate globals will race if run with `t.Parallel()`. The `Server` struct refactor eliminates this — each test constructs its own `Server`. Do not add `t.Parallel()` to engine tests that share a mutable `Engine` instance.
- **Using `http.ListenAndServe` in tests:** Never call `server.Start()` from tests. Construct `Server` directly and call method handlers.
- **Sleeping in tests to wait for goroutines:** The engine goroutine pattern requires small sleeps in `TestEngineStart_*` tests. Keep sleep durations to the minimum needed (10–50ms for transitions, 100ms for cleanup). Use `t.Cleanup` to call `e.Stop()` and wait, avoiding leaks.
- **Calling `simlog.Start()` in tests without cleanup:** `engine.Start()` calls `simlog.Start()` which opens a log file in the working directory. Tests that call `e.Start()` will write log files to the test directory. Use `t.TempDir()` via os.Chdir if this is undesirable, or accept the side-effect since it's a test artifact.
- **Circular import risk:** `server_test.go` in `package server` (or `package server_test`) may import `engine`. This is fine — `server` already imports `engine`. Do not import `server` from within `engine` tests.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTTP test recording | Custom response writer mock | `httptest.NewRecorder()` | stdlib, implements full `http.ResponseWriter` including headers and status code |
| HTTP test request construction | `http.Request{...}` literal | `httptest.NewRequest()` | Sets up Body as `io.NopCloser`, URL parsing, and context correctly |
| Fake executor | Custom interface + struct | `RunnerFunc` function type (D-05/D-06) | Function type injection is idiomatic Go; mirrors `QueryFn` pattern already in codebase |
| Test fixture registry | Hand-written in-memory `Registry` | `playbooks.LoadEmbedded()` | Embedded FS is available in tests within the same module; real data exercises real parsing |

**Key insight:** Go's stdlib `net/http/httptest` is mature and complete — it handles headers, status codes, body buffering, and trailer headers. No mocking framework needed.

---

## Common Pitfalls

### Pitfall 1: `go vet` Failure Blocks `go test ./...`

**What goes wrong:** `cmd/lognojutsu/main.go:28` has `fmt.Println(banner)` where `banner` is a string that ends with `\n`. Go 1.26 `go vet` flags this as a redundant newline. This causes `go test ./...` to report `[build failed]` for the `cmd/lognojutsu` package and the overall exit code is non-zero even though all other packages pass.

**Why it happens:** `go test ./...` runs `go vet` on every package before testing. The `fmt.Println` function appends its own newline — a trailing `\n` in the argument is redundant.

**How to avoid:** Fix in Wave 0 before writing any tests. Change `fmt.Println(banner)` to `fmt.Print(banner)` (the banner const already ends with `\n`).

**Warning signs:** `go test ./...` output shows `# lognojutsu/cmd/lognojutsu` with a vet error before any test output.

### Pitfall 2: Server Struct — authMiddleware Closure Captures cfg

**What goes wrong:** In the current code, `authMiddleware` captures the package-level `cfg` variable by reference in a closure. After refactoring to `(s *Server)`, the closure must capture `s` (or `s.cfg.Password`) — not a local copy of `cfg` at middleware creation time.

**Why it happens:** Go closures capture variables, not values. If middleware is created once in `registerRoutes` and `s.cfg.Password` changes after creation, the closure should see the current value. Since `Config` is set once at startup and never mutated, this is safe — but the closure must reference `s.cfg.Password`, not a local `password := s.cfg.Password` captured at registration time.

**How to avoid:** Write `if s.cfg.Password != ""` inside the closure body, not as a captured local. This is already how the current code works (reads `cfg.Password` at call time).

### Pitfall 3: Engine Tests — runTechnique Calls simlog.Info Without a Session

**What goes wrong:** `engine.Start()` calls `simlog.Start(campaignID)` which opens a log file. If tests call `e.Start()`, simlog state is mutated. If tests do NOT call `e.Start()` (e.g., for `TestFilterByTactics` which calls `filterByTactics` directly), no simlog issue exists. For `TestEngineStart_*` tests that call `e.Start()`, the goroutine will call `simlog.Info(...)` which is safe even without a session (simlog guards with `if current == nil { return }`).

**Why it happens:** `simlog` uses a package-level singleton. Multiple test functions that call `e.Start()` concurrently would interleave simlog sessions.

**How to avoid:** Do not run engine integration tests with `t.Parallel()`. The `-race` detector test (`TestEngineRace`) should also avoid parallel execution with other engine tests.

### Pitfall 4: `httptest.NewRecorder` Status Code Default

**What goes wrong:** `httptest.NewRecorder()` initializes `Code` to `200` (not `0`). If a handler never calls `w.WriteHeader()`, the recorder reports 200 even if the handler only called `w.Write()`. This is correct HTTP behaviour but can mask missing `WriteHeader` calls in test assertions.

**Why it happens:** Standard HTTP behaviour — implicit 200 if `WriteHeader` not called before first write.

**How to avoid:** When testing error paths, assert that the status code is the error code (e.g., 401, 400, 409). For success paths, explicitly check for 200 to confirm the handler completed normally. The `writeError` helper always calls `w.WriteHeader(code)` before writing the body, so error paths are safe.

### Pitfall 5: `handleStart` Shadows `cfg` Variable Name

**What goes wrong:** `internal/server/server.go:handleStart` has:
```go
var cfg engine.Config
if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
```
This local `cfg` shadows the package-level `cfg Config` (server config). After the refactor, `s.cfg` is the server config and the local `cfg` is the engine config — the shadowing is eliminated. But the test must POST a valid JSON `engine.Config` body to avoid a 400 error.

**How to avoid:** In `TestHandleStart_validConfig`, construct a minimal valid `engine.Config` JSON body:
```go
body := strings.NewReader(`{"whatif":true}`)
req := httptest.NewRequest(http.MethodPost, "/api/start", body)
```
WhatIf mode is ideal for tests — engine starts without shelling out.

### Pitfall 6: Race in `Engine.isStopped()` — stopCh Read After Close

**What goes wrong:** `isStopped()` does a non-blocking receive on `stopCh`. `Stop()` does a non-blocking send. After a test calls `e.Stop()` and the goroutine aborts, the `stopCh` channel has been consumed. A second call to `e.Start()` reinitialises `stopCh` (`e.stopCh = make(chan struct{}, 1)`). If a test races a second `Start` with an in-progress abort, the old goroutine may read from the new channel.

**Why it happens:** The stop channel is recreated on each `Start()` call under lock, but the goroutine closure holds a reference to `e` (the engine), not the channel directly — it calls `e.isStopped()` which reads `e.stopCh` at call time. This is safe as written because `isStopped` reads `e.stopCh` (the current channel), but tests should not call `Start()` twice on the same engine without waiting for the first run to complete.

**How to avoid:** In `TestEngineRace`, only call `Start()` once. Use a `sync.WaitGroup` to ensure goroutines complete before the test returns. If the `-race` test wants concurrent Start+Stop, ensure Stop is called before a second Start.

---

## Code Examples

### Handler Test Skeleton

```go
// internal/server/server_test.go
package server

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "lognojutsu/internal/engine"
    "lognojutsu/internal/playbooks"
    "lognojutsu/internal/userstore"
)

func newTestServer(t *testing.T) *Server {
    t.Helper()
    reg, err := playbooks.LoadEmbedded()
    if err != nil {
        t.Fatalf("load registry: %v", err)
    }
    us, _ := userstore.Load()
    eng := engine.New(reg, us)
    return &Server{cfg: Config{}, eng: eng, registry: reg, users: us}
}

func TestHandleStatus_idle(t *testing.T) {
    s := newTestServer(t)
    rec := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
    s.handleStatus(rec, req)
    if rec.Code != http.StatusOK {
        t.Errorf("status: want 200, got %d", rec.Code)
    }
    var st engine.Status
    if err := json.NewDecoder(rec.Body).Decode(&st); err != nil {
        t.Fatalf("decode: %v", err)
    }
    if st.Phase != engine.PhaseIdle {
        t.Errorf("phase: want idle, got %q", st.Phase)
    }
}
```

### Engine RunnerFunc Injection

```go
// internal/engine/engine.go — additions only

// RunnerFunc abstracts OS execution for testability.
type RunnerFunc func(t *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult

type Engine struct {
    mu     sync.RWMutex
    // ... existing fields unchanged ...
    runner RunnerFunc
}

func (e *Engine) SetRunner(fn RunnerFunc) {
    e.runner = fn
}

// In runTechnique — replace the executor dispatch block:
} else if e.runner != nil {
    result = e.runner(t, profile, password)
} else if e.cfg.RunCleanup {
    result = executor.RunWithCleanup(t, profile, password)
} else {
```

### filterByTactics Table Test

```go
// internal/engine/engine_test.go
package engine

func TestFilterByTactics(t *testing.T) {
    techs := []*playbooks.Technique{
        {ID: "A", Tactic: "discovery"},
        {ID: "B", Tactic: "execution"},
        {ID: "C", Tactic: "persistence"},
    }
    cases := []struct {
        name     string
        included []string
        excluded []string
        want     int
        wantIDs  []string
    }{
        {"no filters", nil, nil, 3, []string{"A", "B", "C"}},
        {"include only discovery", []string{"discovery"}, nil, 1, []string{"A"}},
        {"exclude execution", nil, []string{"execution"}, 2, []string{"A", "C"}},
        {"include discovery+execution, exclude execution", []string{"discovery", "execution"}, []string{"execution"}, 1, []string{"A"}},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            e := &Engine{cfg: Config{
                IncludedTactics: tc.included,
                ExcludedTactics: tc.excluded,
            }}
            got := e.filterByTactics(techs)
            if len(got) != tc.want {
                t.Fatalf("got %d, want %d", len(got), tc.want)
            }
            for i, id := range tc.wantIDs {
                if got[i].ID != id {
                    t.Errorf("[%d] got %q, want %q", i, got[i].ID, id)
                }
            }
        })
    }
}
```

---

## State of the Art

| Old Approach | Current Approach | Notes |
|--------------|------------------|-------|
| Package-level globals for HTTP state | `Server` struct with method receivers | Standard Go HTTP server pattern since Go 1.0 |
| No tests | `testing` + `httptest` (stdlib only) | No testify needed for this codebase size |
| Interface extraction for mocking | Function type injection (`RunnerFunc`) | Preferred in idiomatic Go for single-method "interfaces" |

---

## Open Questions

1. **`TestHandleStart_validConfig` — does it wait for the engine goroutine?**
   - What we know: `handleStart` calls `eng.Start()` which spawns a goroutine and returns immediately. The handler returns 200 `{"status":"started"}`.
   - What's unclear: The test should assert the HTTP response only (200 + JSON body). The goroutine running in background will call `simlog.Start()` and write to disk. This is acceptable for tests.
   - Recommendation: Assert response code and body only. Do not assert engine phase (it races). Call `eng.Stop()` in `t.Cleanup` to prevent goroutine leaks between test functions.

2. **`authMiddleware` — method or closure field on `Server`?**
   - What we know: D-02 says all handlers become method receivers. `authMiddleware` is a middleware wrapper, not a direct handler.
   - What's unclear: Whether `(s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc` is a method or a field.
   - Recommendation: Method receiver `(s *Server) authMiddleware(...)` is cleanest. Used as `s.authMiddleware(s.handleStatus)` in `registerRoutes`.

3. **`writeJSON` / `writeError` — methods or package-level helpers?**
   - What we know: CONTEXT.md D-code context says "Claude's discretion."
   - Recommendation: Keep as package-level functions. They take `http.ResponseWriter` as input and have no dependency on `Server` state — no reason to make them methods.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| Go toolchain | All compilation + testing | Yes | go1.26.1 | — |
| `go test -race` | D-08, QUAL-03/04/05 | Yes | Included in go1.26.1 | — |
| `net/http/httptest` | Handler tests (D-10) | Yes | stdlib | — |
| `gopkg.in/yaml.v3` | playbooks package | Yes | v3.0.1 (go.sum) | — |

No missing dependencies. All test tooling is stdlib or already present in go.mod.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | stdlib `testing` — go1.26.1 |
| Config file | none (no go test flags needed beyond `-race`) |
| Quick run command | `go test ./internal/...` |
| Full suite command | `go test ./... -race` |

### Phase Requirements — Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| QUAL-01 | Server struct — no globals | unit (compile-time verification) | `go build ./...` + `go test ./internal/server/...` | Wave 0 |
| QUAL-02 | Packages exist as `internal/X` | structural (directory check) | `go build ./internal/...` | Already satisfied |
| QUAL-03 | Engine state machine transitions | unit | `go test ./internal/engine/... -race` | Wave 0 |
| QUAL-04 | HTTP handler responses | unit | `go test ./internal/server/... -race` | Wave 0 |
| QUAL-05 | Verifier pass/fail/notRun | unit | `go test ./internal/verifier/... -race` | Exists — passes |

### Sampling Rate
- **Per task commit:** `go test ./internal/...`
- **Per wave merge:** `go test ./... -race`
- **Phase gate:** `go test ./... -race` green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/engine/engine_test.go` — covers QUAL-03 (D-09 tests)
- [ ] `internal/server/server_test.go` — covers QUAL-04 (D-10 tests); depends on Server struct refactor
- [ ] Fix `cmd/lognojutsu/main.go:28` — `fmt.Println(banner)` redundant newline (vet error blocks `go test ./...`)

*(Existing: `internal/verifier/verifier_test.go`, `internal/playbooks/types_test.go`, `internal/reporter/reporter_test.go` all pass)*

---

## Project Constraints (from CLAUDE.md)

CLAUDE.md does not exist in this repository. No project-level directives to apply beyond what is stated in CONTEXT.md and the codebase conventions documented in `.planning/codebase/CONVENTIONS.md`.

Applicable conventions (from CONVENTIONS.md — treat as binding):
- Receiver: `(s *Server)` — short 1-2 letter matching type initial
- Constructor: `func New(...) *Type` pattern — `Start()` is an exception by design (D-03)
- Section separators: `// ── Section ────────────────────────────────────────────` in large files
- Error variable naming: first error is `err`; subsequent use descriptive names (`readErr`, `parseErr`)
- Imports: stdlib first, blank line, then internal
- No `TODO`/`FIXME` markers
- `go vet` / `gofmt` compliance required (no linter config but standard tooling assumed)

---

## Sources

### Primary (HIGH confidence)
- Source: direct code reading of `internal/server/server.go` — all globals, handlers, and routes confirmed
- Source: direct code reading of `internal/engine/engine.go` — Engine struct, RunnerFunc injection point, filterByTactics
- Source: direct code reading of `internal/verifier/verifier.go` — QueryFn pattern (D-06 mirror)
- Source: direct code reading of `internal/verifier/verifier_test.go` — confirms Phase 1 already wrote verifier tests
- Source: `go test ./...` output — confirms playbooks, reporter, verifier pass; engine, server have no tests; cmd has vet error
- Source: `go version` — go1.26.1 confirmed
- Source: `.planning/codebase/TESTING.md` — zero-test baseline confirmed; httptest pattern documented
- Source: `.planning/codebase/CONVENTIONS.md` — naming, error, import conventions

### Secondary (MEDIUM confidence)
- `net/http/httptest` package documentation (stdlib, stable since Go 1.0) — patterns are well-established

### Tertiary (LOW confidence)
- None

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — stdlib only, confirmed by go.mod and go version
- Architecture patterns: HIGH — derived from direct code reading of source files
- Pitfalls: HIGH — identified from actual code (vet error confirmed by `go test` run, shadowing from code reading)
- Test patterns: HIGH — mirrors existing verifier_test.go patterns already in codebase

**Research date:** 2026-03-25
**Valid until:** 2026-06-25 (Go stdlib patterns are stable; no external dependencies)
