---
phase: 16-safety-infrastructure
plan: 02
subsystem: engine, server, ui
tags: [scan-confirmation, safety, modal, api, engine-pause]

# Dependency graph
requires:
  - phase: 16-safety-infrastructure
    plan: 01
    provides: RequiresConfirmation bool field on Technique struct
  - phase: 13-poc-scheduling-tests
    provides: engine test patterns (RunnerFunc injection, testRegistry helpers)
provides:
  - ScanInfo struct (TargetSubnet, RateLimitNote, IDSWarning, Techniques) in engine
  - scanConfirmCh/scanCancelCh/scanConfirmMu/scanPendingInfo fields on Engine struct
  - runScanConfirmation() method: pre-flight scan gate in run() and runPoC()
  - ConfirmScan(), CancelScan(), GetScanPending() exported Engine methods
  - detectLocalSubnet() using net.Interfaces() for /24 auto-detection
  - /api/scan/pending (204/200+JSON), /api/scan/confirm (POST 200/409), /api/scan/cancel (POST 200/409)
  - scanConfirmModal in index.html with 4 info items, Confirm/Cancel buttons
  - pollScanPending() JS polling loop wired to simulation start/stop
affects: [engine, server, ui, 16-03]

# Tech tracking
tech-stack:
  added: [net package (stdlib) for subnet detection, github.com/go-ldap/ldap/v3, github.com/yusufpapurcu/wmi (parallel agent deps)]
  patterns:
    - "Channel-based engine pause/resume: scanConfirmCh and scanCancelCh as gate signals"
    - "runScanConfirmation() helper isolates pre-flight logic from run()/runPoC() loops"
    - "Separate mutex (scanConfirmMu) guards scan state independently from status mutex"
    - "Local channel copies before select prevents race when ConfirmScan() nils the channel"

key-files:
  created:
    - internal/engine/engine_scan_confirm_test.go
    - internal/server/server_scan_test.go
  modified:
    - internal/engine/engine.go
    - internal/server/server.go
    - internal/server/static/index.html
    - internal/playbooks/types.go
    - go.mod
    - go.sum

key-decisions:
  - "runScanConfirmation() captures channel pointers locally before select — prevents race when ConfirmScan() nils scanConfirmCh while select is evaluating"
  - "Pre-flight runs once per simulation in both run() and runPoC() — WhatIf mode skips gate entirely"
  - "abort() called inside runScanConfirmation() on cancel/stop — consistent with run()/runPoC() abort pattern"
  - "scanPendingInfo cleared before abort() call — prevents stale info appearing after cancel"
  - "Modal uses api() helper for fetch calls — consistent with existing auth pattern; no separate authHeaders needed"
  - "stopScanPoll() + hideScanModal() on done/aborted — cleans up polling when simulation ends regardless of confirmation state"

# Metrics
duration: ~6min
completed: 2026-04-09
---

# Phase 16 Plan 02: Scan Confirmation Flow Summary

**Channel-based engine pause/resume with API endpoints and web UI modal — consultant must explicitly confirm before network scanning techniques execute**

## Performance

- **Duration:** ~6 min
- **Started:** 2026-04-09T17:08:00Z
- **Completed:** 2026-04-09T17:14:00Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments

- Added `ScanInfo` struct to engine package with TargetSubnet, RateLimitNote, IDSWarning, Techniques fields
- Added scan confirmation fields to Engine struct: `scanConfirmCh`, `scanCancelCh`, `scanConfirmMu`, `scanPendingInfo`
- Implemented `runScanConfirmation()` — collects `RequiresConfirmation=true` techniques, blocks on channel select
- Pre-flight scan gate inserted in `run()` and `runPoC()` before discovery phase; WhatIf mode skips gate entirely
- `detectLocalSubnet()` uses `net.Interfaces()` to auto-detect the first non-loopback IPv4 /24 subnet
- `ConfirmScan()`, `CancelScan()`, `GetScanPending()` exported methods with mutex-guarded channel operations
- Scan state reset in `Start()` alongside stopCh reset — clean state on each simulation start
- Three new API endpoints: `GET /api/scan/pending` (204/200+JSON), `POST /api/scan/confirm` (200/409), `POST /api/scan/cancel` (200/409)
- Scan confirmation modal in index.html: 4 info items, Cancel Scan / Confirm Scan buttons
- JS polling loop (`pollScanPending()`, 1s interval) wired to `startSimulation()`, `quickStart()`, and done/aborted completion handler
- Modal not dismissible by overlay click or Escape — deliberate safety checkpoint
- 11 unit tests across engine_scan_confirm_test.go and server_scan_test.go — all pass
- TDD RED/GREEN cycle followed per plan specification

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: Failing tests for scan confirmation** — `a603237` (test)
2. **Task 1 GREEN: Engine pause/resume and API endpoints** — `5f03e81` (feat)
3. **Task 2: Scan confirmation modal in web UI** — `5c91db8` (feat)

## Files Created/Modified

- `internal/engine/engine.go` — ScanInfo struct, scan fields on Engine, ConfirmScan/CancelScan/GetScanPending methods, runScanConfirmation(), detectLocalSubnet(), scan pre-flight in run()/runPoC()
- `internal/server/server.go` — Three new scan API routes + handleScanPending/handleScanConfirm/handleScanCancel handlers
- `internal/server/static/index.html` — scanConfirmModal HTML, JS scan modal functions, poll wiring to start/stop
- `internal/playbooks/types.go` — Added `Tier int` field to Technique and ExecutionResult (parallel agent dependency fix)
- `internal/engine/engine_scan_confirm_test.go` — 6 engine tests (blocks, cancel, no-block, whatif, pending, no-pending)
- `internal/server/server_scan_test.go` — 5 server tests (pending 204, confirm 409, confirm 405, cancel 409, cancel 405)
- `go.mod`, `go.sum` — Added go-ldap/v3, wmi, crypto modules (parallel agent dependency fix)

## Decisions Made

- `runScanConfirmation()` captures channel pointers locally before the select statement — prevents race when `ConfirmScan()` nils `scanConfirmCh` while select is in-flight
- Pre-flight runs once in both `run()` and `runPoC()` before discovery — single gate per simulation regardless of mode
- WhatIf mode skips the pre-flight gate entirely — no modal for preview-only runs
- Modal uses existing `api()` helper for fetch calls — consistent with established auth pattern
- `stopScanPoll()` + `hideScanModal()` called on simulation done/aborted — prevents stale modal after completion

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Missing `Tier` field on Technique and ExecutionResult**
- **Found during:** Task 1 (compile error during test setup)
- **Issue:** Parallel agent added `t.Tier` and `res.Tier` references to executor.go and reporter.go but didn't add the `Tier int` field to playbooks/types.go
- **Fix:** Added `Tier int` to both `Technique` struct and `ExecutionResult` struct in types.go
- **Files modified:** `internal/playbooks/types.go`
- **Commit:** `5f03e81`

**2. [Rule 3 - Blocking] Missing go module dependencies**
- **Found during:** Task 1 (build failure: go-ldap, wmi packages not in go.mod)
- **Issue:** Parallel agent added `internal/native` package with LDAP and WMI imports but didn't run `go get` for the required modules
- **Fix:** Ran `go get github.com/go-ldap/ldap/v3` and `go get github.com/yusufpapurcu/wmi`
- **Files modified:** `go.mod`, `go.sum`
- **Commit:** `5f03e81`

## Known Stubs

None — scan confirmation flow is fully wired end-to-end.

## User Setup Required

None — no external configuration required.

## Self-Check: PASSED

- FOUND: `internal/engine/engine_scan_confirm_test.go`
- FOUND: `internal/server/server_scan_test.go`
- FOUND: `.planning/phases/16-safety-infrastructure/16-02-SUMMARY.md`
- FOUND commit a603237: test(16-02): add failing tests
- FOUND commit 5f03e81: feat(16-02): implement engine/API
- FOUND commit 5c91db8: feat(16-02): add scan confirmation modal UI
