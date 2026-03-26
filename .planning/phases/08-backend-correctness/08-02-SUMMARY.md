---
phase: 08-backend-correctness
plan: 02
subsystem: api
tags: [go, version, ldflags, http-endpoint, testing]

# Dependency graph
requires:
  - phase: 08-backend-correctness plan 01
    provides: auditpol GUID migration and preparation fixes
provides:
  - Build-time version injection via ldflags (-X main.version)
  - Public GET /api/info endpoint returning {"version":"..."} without auth
  - Three new server tests validating /api/info behavior
affects: [09-ui-polish, version-badge, frontend-fetch]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "ldflags version injection: var version = \"dev\" in main package, overridable via -ldflags \"-X main.version=v1.1.0\""
    - "Public endpoint pattern: registered in registerRoutes without authMiddleware wrapper, sets CORS and Content-Type headers directly"

key-files:
  created: []
  modified:
    - cmd/lognojutsu/main.go
    - internal/server/server.go
    - internal/server/server_test.go

key-decisions:
  - "var version = \"dev\" replaces const banner — package-level var enables ldflags injection at link time"
  - "/api/info registered without authMiddleware — version is not sensitive and Phase 9 badge must load before login"
  - "handleInfo sets headers directly (not via middleware) — mirrors D-11 decision from context"
  - "testServer helper updated with Version: \"test-v0.0.0\" — keeps existing tests valid while enabling version-specific tests"

patterns-established:
  - "Public route pattern: mux.HandleFunc without authMiddleware wrapper for non-sensitive endpoints"
  - "Version injection: package-level var in main, passed into Config struct, accessible via s.cfg.Version"

requirements-completed: [VER-01, VER-02]

# Metrics
duration: 15min
completed: 2026-03-26
---

# Phase 8 Plan 02: Version Injection and /api/info Endpoint Summary

**Build-time version injection via ldflags and public GET /api/info endpoint returning `{"version":"..."}` without authentication**

## Performance

- **Duration:** 15 min
- **Started:** 2026-03-26T16:40:00Z
- **Completed:** 2026-03-26T16:55:00Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Replaced `const banner` with `var version = "dev"` + `const bannerArt` in main.go — ldflags target is `main.version`
- Added `Version string` to `server.Config` and plumbed from main.go into server startup
- Registered `GET /api/info` without authMiddleware; `handleInfo` returns `{"version":"..."}` JSON with CORS headers
- Three new tests: `TestHandleInfo_returnsVersion`, `TestHandleInfo_noAuthRequired`, `TestRegisterRoutes_infoNoAuth` — all pass

## Task Commits

Each task was committed atomically:

1. **Task 1: Add version var to main.go and Version to server.Config** - `7a39e9d` (feat)
2. **Task 2: Add tests for /api/info endpoint** - `6e375f9` (test)

**Plan metadata:** (docs commit — created below)

## Files Created/Modified
- `cmd/lognojutsu/main.go` - var version="dev", bannerArt const, fmt.Printf banner, Config.Version plumbed
- `internal/server/server.go` - Version in Config struct, /api/info route, handleInfo method
- `internal/server/server_test.go` - testServer gets Version field, three new TestHandleInfo/TestRegisterRoutes tests

## Decisions Made
- Chose `fmt.Printf("%s%s\n", bannerArt, version)` pattern — clean inline injection without helper function
- handleInfo sets headers directly (not through middleware) — consistent with D-11 decision; middleware is not called
- testServer updated with `Version: "test-v0.0.0"` to ensure existing tests remain unaffected

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- VER-01 and VER-02 complete — Phase 9 can fetch `/api/info` for the version badge before user authentication
- ldflags injection verified: `go build -ldflags "-X main.version=v1.1.0" ./cmd/lognojutsu/` compiles cleanly
- All existing tests continue to pass (9/9 server tests, no regressions)

---
*Phase: 08-backend-correctness*
*Completed: 2026-03-26*
