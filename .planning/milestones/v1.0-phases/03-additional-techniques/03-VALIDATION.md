---
phase: 3
slug: additional-techniques
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-25
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (`go test ./...`) |
| **Config file** | None — standard Go test runner |
| **Quick run command** | `go test ./internal/playbooks/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/playbooks/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 3-01-01 | 01 | 1 | TECH-01 | unit | `go test ./internal/playbooks/... -count=1` | ✅ existing | ✅ green |
| 3-01-02 | 01 | 1 | TECH-01 | unit | `go test ./internal/playbooks/... -count=1` | ✅ existing | ✅ green |
| 3-01-03 | 01 | 1 | TECH-01 | unit | `go test ./internal/playbooks/... -count=1` | ✅ existing | ✅ green |
| 3-01-04 | 01 | 1 | TECH-03 | unit | `go test ./internal/playbooks/... -run TestExpectedEvents -count=1` | ✅ | ✅ green |
| 3-02-01 | 02 | 1 | TECH-02 | unit | `go test ./internal/playbooks/... -count=1` | ✅ existing | ✅ green |
| 3-02-02 | 02 | 1 | TECH-02 | unit | `go test ./internal/playbooks/... -count=1` | ✅ existing | ✅ green |
| 3-03-01 | 03 | 2 | TECH-01 | manual | Build binary, verify new entries in Web UI Playbooks tab | manual | ✅ green |
| 3-03-02 | 03 | 2 | TECH-01/02/03 | unit | `go test ./... -count=1` | ✅ existing | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Add `TestExpectedEvents` to `internal/playbooks/` test file — asserts every loaded technique has `len(ExpectedEvents) > 0`, satisfying TECH-03 in a machine-verifiable way.

*TestExpectedEvents was created during Phase 3 plan 02 implementation in loader_test.go. Confirmed passing 2026-03-26.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| New techniques appear in Web UI Playbooks tab | TECH-01, TECH-02 | UI display requires running binary | Build `go build ./...`, launch `lognojutsu.exe`, open http://localhost:8080, check Playbooks tab shows 9 new entries |
| README technique table shows new rows | TECH-01 | File content check | `grep "T1005\|T1560\|T1071\|T1119\|UEBA" README.md` should return 9+ matches |

# Note: go test ./internal/playbooks/... may be blocked by Windows Defender quarantining playbooks.test.exe. Add %LOCALAPPDATA%\Temp\go-build* exclusion before running. Tests passed at implementation time; structural coverage is present.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved (2026-03-26)
