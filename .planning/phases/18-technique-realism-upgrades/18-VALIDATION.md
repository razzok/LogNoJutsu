---
phase: 18
slug: technique-realism-upgrades
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-10
---

# Phase 18 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing test infrastructure |
| **Quick run command** | `go test ./internal/playbooks/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/playbooks/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 18-01-01 | 01 | 1 | TECH-01 | unit | `go test ./internal/playbooks/...` | ✅ | ⬜ pending |
| 18-01-02 | 01 | 1 | TECH-01 | unit | `go test ./internal/playbooks/...` | ✅ | ⬜ pending |
| 18-02-01 | 02 | 1 | TECH-02, TECH-03, TECH-04 | unit | `go test ./internal/playbooks/...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| EID 5140/5145 audit policy caveat accuracy | TECH-01 | Requires Windows audit policy state | Enable Object Access auditing, run T1135, verify EID 5140 appears |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
