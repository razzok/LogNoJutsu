---
phase: 14
slug: safety-audit
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-09
---

# Phase 14 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go test infrastructure |
| **Quick run command** | `go test ./internal/playbooks/... ./internal/executor/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/playbooks/... ./internal/executor/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 14-01-01 | 01 | 1 | SAFE-02 | unit | `go test ./internal/playbooks/ -run TestTierField` | ❌ W0 | ⬜ pending |
| 14-01-02 | 01 | 1 | SAFE-02 | unit | `go test ./internal/playbooks/ -run TestTierLoading` | ❌ W0 | ⬜ pending |
| 14-02-01 | 02 | 1 | SAFE-03 | unit | `go test ./internal/executor/ -run TestRunWithCleanup` | ❌ W0 | ⬜ pending |
| 14-03-01 | 03 | 2 | SAFE-01 | integration | `go test ./internal/playbooks/ -run TestDestructiveTechniques` | ❌ W0 | ⬜ pending |
| 14-04-01 | 04 | 3 | SAFE-02 | manual | visual inspection of HTML report and web UI | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Test stubs for tier field loading and validation
- [ ] Test stubs for RunWithCleanup defer pattern
- [ ] Test stubs for destructive technique safety verification

*Existing Go test infrastructure covers framework needs — only test files need creation.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Tier badge visible in HTML report | SAFE-02 | Visual rendering check | Open report.html, verify tier column shows 1/2/3 for each technique |
| Tier badge visible in web UI | SAFE-02 | Visual rendering check | Run `go run ./cmd/lognojutsu serve`, verify tier badges in technique list |
| Classification doc scannable | SAFE-02 | UX/readability judgment | Open docs/TECHNIQUE-CLASSIFICATION.md, find any technique in <10s |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
