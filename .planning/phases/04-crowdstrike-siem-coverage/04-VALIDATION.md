---
phase: 4
slug: crowdstrike-siem-coverage
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard testing (`testing` package) |
| **Config file** | None — `go test ./...` convention |
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
| 4-01-01 | 01 | 0 | CROW-01 | unit | `go test ./internal/playbooks/... -run TestSIEMCoverage` | ❌ Wave 0 | ⬜ pending |
| 4-01-02 | 01 | 0 | CROW-02 | unit | `go test ./internal/playbooks/... -run TestFalconTechniques` | ❌ Wave 0 | ⬜ pending |
| 4-01-03 | 01 | 0 | CROW-02 | unit | `go test ./internal/playbooks/... -run TestNewTechniqueCount` | ✅ needs bump | ⬜ pending |
| 4-01-04 | 01 | 0 | CROW-03 | unit | `go test ./internal/reporter/... -run TestHTMLCrowdStrikeColumn` | ❌ Wave 0 | ⬜ pending |
| 4-02-01 | 02 | 1 | CROW-01 | unit | `go test ./internal/playbooks/... -run TestSIEMCoverage` | ❌ Wave 0 | ⬜ pending |
| 4-03-01 | 03 | 2 | CROW-02 | unit | `go test ./internal/playbooks/... -run TestFalconTechniques` | ❌ Wave 0 | ⬜ pending |
| 4-03-02 | 03 | 2 | CROW-02 | unit | `go test ./internal/playbooks/... -run TestNewTechniqueCount` | ✅ needs bump | ⬜ pending |
| 4-04-01 | 04 | 3 | CROW-03 | unit | `go test ./internal/reporter/... -run TestHTMLCrowdStrikeColumn` | ❌ Wave 0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/playbooks/loader_test.go` — add `TestSIEMCoverage` (verify `map[string][]string` parses from YAML with `siem_coverage` key)
- [ ] `internal/playbooks/loader_test.go` — add `TestFalconTechniques` (3 FALCON_ files exist, have non-empty `siem_coverage.crowdstrike`, have `expected_events`)
- [ ] `internal/playbooks/loader_test.go` — bump `TestNewTechniqueCount` threshold from 48 to 51 (adds 3 FALCON_ files)
- [ ] `internal/reporter/reporter_test.go` — add `TestHTMLCrowdStrikeColumn` (column present when SIEMCoverage populated, absent when all entries have empty crowdstrike slice)

*No new framework install needed — `testing` stdlib already in use across all packages.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Falcon alert names match actual Falcon console | CROW-01 | No Falcon sandbox available in dev | Reviewer reads mapped names and confirms against Falcon documentation |
| CrowdStrike docs section in README readable | CROW-01 | Human readability check | Read the new German README section for CrowdStrike prerequisites |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
