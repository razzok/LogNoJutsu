---
phase: 15
slug: native-go-architecture
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-09
---

# Phase 15 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | none (go test ./... from repo root) |
| **Quick run command** | `go test ./internal/native/... ./internal/executor/... ./internal/playbooks/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/native/... ./internal/executor/... ./internal/playbooks/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 15-01-01 | 01 | 1 | ARCH-01 | unit | `go test ./internal/native/... -run TestRegisterLookup` | ❌ W0 | ⬜ pending |
| 15-01-02 | 01 | 1 | ARCH-01 | unit | `go test ./internal/executor/... -run TestGoDispatch` | ❌ W0 | ⬜ pending |
| 15-01-03 | 01 | 1 | ARCH-01 | unit | `go test ./internal/executor/... -run TestGoDispatchUnregistered` | ❌ W0 | ⬜ pending |
| 15-02-01 | 02 | 2 | ARCH-02 | unit | `go test ./internal/native/... -run TestT1482NoDC` | ❌ W0 | ⬜ pending |
| 15-02-02 | 02 | 2 | ARCH-02 | unit | `go test ./internal/playbooks/... -run TestT1482ExecutorType` | ❌ W0 | ⬜ pending |
| 15-03-01 | 03 | 2 | ARCH-03 | unit | `go test ./internal/native/... -run TestT1057WMI` | ❌ W0 | ⬜ pending |
| 15-03-02 | 03 | 2 | ARCH-03 | unit | `go test ./internal/playbooks/... -run TestT1057ExecutorType` | ❌ W0 | ⬜ pending |
| 15-XX-XX | all | all | All | regression | `go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/native/registry_test.go` — covers ARCH-01 register/lookup/cleanup
- [ ] `internal/executor/executor_go_test.go` — covers ARCH-01 dispatch, unregistered case, RunAs log note
- [ ] `internal/native/t1482_test.go` — covers ARCH-02 no-DC fallback
- [ ] `internal/native/t1057_test.go` — covers ARCH-03 Win32_Process result (Windows build tag)
- [ ] `internal/playbooks/loader_test.go` additions — verify T1482 and T1057 now have `executor.type == "go"`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| LDAP query against real DC | ARCH-02 | Requires domain-joined machine with reachable DC | Run T1482 on domain-joined test machine, verify trust output |
| WMI process listing | ARCH-03 | Requires Windows runtime with WMI service | Run T1057 on Windows, verify process list in output |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
