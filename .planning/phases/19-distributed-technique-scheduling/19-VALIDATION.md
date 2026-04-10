---
phase: 19
slug: distributed-technique-scheduling
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-10
---

# Phase 19 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) — Go 1.26.1 |
| **Config file** | none (standard `go test`) |
| **Quick run command** | `go test ./internal/engine/... -timeout 30s` |
| **Full suite command** | `go test ./... -timeout 60s` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/engine/... -timeout 30s`
- **After every plan wave:** Run `go test ./... -timeout 60s`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 19-01-01 | 01 | 1 | POC-03 | unit | `go test ./internal/engine/... -run TestRandomSlotsInWindow -timeout 30s` | ❌ W0 | ⬜ pending |
| 19-02-01 | 02 | 2 | POC-01 | unit | `go test ./internal/engine/... -run TestPoCPhase1_DistributedSlots -timeout 30s` | ❌ W0 | ⬜ pending |
| 19-03-01 | 03 | 2 | POC-02 | unit | `go test ./internal/engine/... -run TestPoCPhase2_BatchedSlots -timeout 30s` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/engine/poc_schedule_test.go` — `TestRandomSlotsInWindow` stubs for POC-03 (window boundary invariant)
- [ ] `internal/engine/poc_schedule_test.go` — `TestPoCPhase1_DistributedSlots` stubs for POC-01 (one technique per slot, multiple After() calls)
- [ ] `internal/engine/poc_schedule_test.go` — `TestPoCPhase2_BatchedSlots` stubs for POC-02 (batch grouping, multiple slots per day)

*Note: Existing test infrastructure (fakeClock, captureClock) covers framework needs. No new test framework installation required.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| UI schedule preview shows window range | POC-03 | Browser rendering | Open index.html, verify `updatePoCSchedule()` displays `08:00–17:00` format |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
