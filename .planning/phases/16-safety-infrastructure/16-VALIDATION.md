---
phase: 16
slug: safety-infrastructure
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-09
---

# Phase 16 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (`testing` package) |
| **Config file** | none — standard Go test discovery |
| **Quick run command** | `go test ./internal/engine/ ./internal/executor/ -run "TestAMSI|TestElevation|TestScanConfirm" -timeout 30s` |
| **Full suite command** | `go test ./... -timeout 120s` |
| **Estimated runtime** | ~30 seconds (quick), ~120 seconds (full) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/engine/ ./internal/executor/ -run "TestAMSI|TestElevation|TestScanConfirm" -timeout 30s`
- **After every plan wave:** Run `go test ./... -timeout 120s`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 16-01-01 | 01 | 1 | INFRA-01 | unit | `go test ./internal/executor/ -run TestAMSI -v` | ❌ W0 | ⬜ pending |
| 16-01-02 | 01 | 1 | INFRA-01 | unit | `go test ./internal/executor/ -run TestAMSI_ExecutorTypes -v` | ❌ W0 | ⬜ pending |
| 16-01-03 | 01 | 1 | INFRA-02 | unit | `go test ./internal/engine/ -run TestElevationSkip -v` | ❌ W0 | ⬜ pending |
| 16-01-04 | 01 | 1 | INFRA-02 | unit | `go test ./internal/engine/ -run TestElevationRun -v` | ❌ W0 | ⬜ pending |
| 16-02-01 | 02 | 2 | INFRA-03 | unit | `go test ./internal/engine/ -run TestScanConfirm -v` | ❌ W0 | ⬜ pending |
| 16-02-02 | 02 | 2 | INFRA-03 | unit | `go test ./internal/server/ -run TestScanConfirmAPI -v` | ❌ W0 | ⬜ pending |
| 16-02-03 | 02 | 2 | INFRA-03 | unit | `go test ./internal/server/ -run TestScanPending -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/executor/executor_amsi_test.go` — stubs for INFRA-01 (AMSI detection with fake stderr)
- [ ] `internal/engine/engine_elevation_test.go` — stubs for INFRA-02 (engine with injected `isAdmin` flag)
- [ ] `internal/engine/engine_scan_confirm_test.go` — stubs for INFRA-03 (channel-based confirmation)
- [ ] `internal/server/server_scan_test.go` — stubs for INFRA-03 API endpoints

*Note: The existing `RunnerFunc` injection pattern in engine.go enables testing elevation skip without executing real techniques. The `isAdmin bool` field must be exported or exposed via a test-only setter.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Real AMSI block with Defender enabled | INFRA-01 | Requires Windows Defender active — cannot mock system-wide AMSI | Run binary on Windows with Defender, execute a known-flagged PowerShell technique, verify "AMSI Blocked" status |
| Real `IsElevated()` check | INFRA-02 | Requires running binary as admin vs non-admin | Run binary as non-admin — verify elevation-required techniques are skipped. Run as admin — verify they execute. |
| Scan confirmation modal UX | INFRA-03 | Browser interaction test | Start a simulation with T1046, verify modal appears with subnet/IDS warning, confirm and verify scan proceeds |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
