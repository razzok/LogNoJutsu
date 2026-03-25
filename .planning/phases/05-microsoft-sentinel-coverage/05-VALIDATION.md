---
phase: 5
slug: microsoft-sentinel-coverage
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (`testing` stdlib) |
| **Config file** | none — `go test ./...` from project root |
| **Quick run command** | `go test ./internal/playbooks/... ./internal/reporter/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/playbooks/... ./internal/reporter/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 5-01-01 | 01 | 1 | SENT-01 | unit | `go test ./internal/playbooks/... -run TestSentinelCoverage` | ❌ Wave 0 | ⬜ pending |
| 5-02-01 | 02 | 1 | SENT-02 | unit | `go test ./internal/playbooks/... -run TestAZURETechniqueCount` | ❌ Wave 0 | ⬜ pending |
| 5-03-01 | 03 | 1 | SENT-03 | unit | `go test ./internal/reporter/... -run TestHasSentinel` | ❌ Wave 0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/playbooks/playbooks_test.go` (or existing test file) — add `TestSentinelCoverage` and `TestAZURETechniqueCount` stubs (check if Phase 3/4 added technique count tests to extend)
- [ ] `internal/reporter/reporter_test.go` — add `TestHasSentinel` covering `HasSentinel=true` path (results with sentinel data) and `HasSentinel=false` path (no sentinel data)

*Check existing test files first — Phase 3 added `TestNewTechniqueCount`, Phase 4 may have added `HasCrowdStrike` reporter tests. Extend rather than duplicate.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| HTML report renders MS badge with correct blue color (#0078D4) | SENT-03 | Visual CSS check | Build binary, run simulation with AZURE_ technique, open HTML report, verify blue MS badge appears in Sentinel column |
| Sentinel column absent when no AZURE_ techniques in results | SENT-03 | Conditional rendering | Run simulation without AZURE_ techniques, verify Sentinel column does not appear in HTML output |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
