---
phase: 11
slug: daily-tracking-backend-campaign-delay
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-09
---

# Phase 11 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — standard Go test toolchain |
| **Quick run command** | `go test ./internal/engine/ ./internal/server/` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/engine/ ./internal/server/`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 11-01-01 | 01 | 1 | TRACK-01 | unit | `go test ./internal/engine/ -run TestDayDigest` | ✅ | ⬜ pending |
| 11-01-02 | 01 | 1 | TRACK-02 | unit | `go test ./internal/engine/ -run TestDayDigestPrePopulate` | ❌ W0 | ⬜ pending |
| 11-01-03 | 01 | 1 | TRACK-04 | unit | `go test ./internal/engine/ -run TestHeartbeat` | ❌ W0 | ⬜ pending |
| 11-02-01 | 02 | 1 | TRACK-03 | unit | `go test ./internal/server/ -run TestHandlePoCDays` | ❌ W0 | ⬜ pending |
| 11-03-01 | 03 | 2 | CAMP-01 | unit | `go test ./internal/engine/ -run TestDelayAfter` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/engine/engine_test.go` — test stubs for DayDigest lifecycle, pre-population, heartbeat, delay_after
- [ ] `internal/server/server_test.go` — test stub for `/api/poc/days` handler

*Existing test infrastructure (fakeClock, waitOrStop) covers all phase requirements.*

---

## Manual-Only Verifications

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
