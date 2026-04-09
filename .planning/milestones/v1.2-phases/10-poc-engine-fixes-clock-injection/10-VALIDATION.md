---
phase: 10
slug: poc-engine-fixes-clock-injection
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-08
---

# Phase 10 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — standard Go test tooling |
| **Quick run command** | `go test ./internal/engine/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/engine/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 10-01-01 | 01 | 1 | TEST-01 | unit | `go test ./internal/engine/... -run TestClock` | ❌ W0 | ⬜ pending |
| 10-01-02 | 01 | 1 | POCFIX-01 | unit | `go test ./internal/engine/... -run TestPoCDay` | ❌ W0 | ⬜ pending |
| 10-01-03 | 01 | 1 | POCFIX-02 | grep | `grep -c "Tag\|warte\|keine" internal/engine/engine.go` | ✅ | ⬜ pending |
| 10-01-04 | 01 | 1 | POCFIX-03 | grep | `grep -c "simlog.Phase" internal/engine/engine.go` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing test infrastructure covers framework setup (go test built-in)
- Phase 13 will add comprehensive runPoC() scheduling tests using the Clock interface from this phase

*Existing infrastructure covers all phase requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| CurrentStep shows English in live UI | POCFIX-02 | Requires running PoC mode and observing web UI | Start a PoC run via web UI, verify CurrentStep in status response shows English text |
| Log viewer shows phase separators | POCFIX-03 | Requires visual inspection of log viewer | Start PoC run, check /api/logs response for phase separator entries |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
