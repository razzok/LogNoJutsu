---
phase: 20
slug: scheduling-test-coverage
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-11
---

# Phase 20 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — standard Go testing |
| **Quick run command** | `go test ./internal/engine/... -run TestPoC -count=1` |
| **Full suite command** | `go test ./internal/engine/... -count=1` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/engine/... -run TestPoC -count=1`
- **After every plan wave:** Run `go test ./internal/engine/... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 2 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 20-01-01 | 01 | 1 | POC-04 | unit | `go test ./internal/engine/... -run TestPoCStop -count=1` | ✅ | ⬜ pending |
| 20-01-02 | 01 | 1 | POC-04 | unit | `go test ./internal/engine/... -run TestDayDigest -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Existing infrastructure covers all phase requirements.*

---

## Manual-Only Verifications

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 2s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
