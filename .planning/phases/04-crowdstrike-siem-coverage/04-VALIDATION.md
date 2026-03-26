---
phase: 4
slug: crowdstrike-siem-coverage
status: complete
nyquist_compliant: true
wave_0_complete: true
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
| 4-01-01 | 01 | 0 | CROW-01 | unit | `go test ./internal/playbooks/... -run TestSIEMCoverage` | ✅ | ✅ green |
| 4-01-02 | 01 | 0 | CROW-02 | unit | `go test ./internal/playbooks/... -run TestFalconTechniques` | ✅ | ✅ green |
| 4-01-03 | 01 | 0 | CROW-02 | unit | `go test ./internal/playbooks/... -run TestNewTechniqueCount` | ✅ | ✅ green |
| 4-01-04 | 01 | 0 | CROW-03 | unit | `go test ./internal/reporter/... -run TestHTMLCrowdStrikeColumn` | ✅ | ✅ green |
| 4-02-01 | 02 | 1 | CROW-01 | unit | `go test ./internal/playbooks/... -run TestSIEMCoverage` | ✅ | ✅ green |
| 4-03-01 | 03 | 2 | CROW-02 | unit | `go test ./internal/playbooks/... -run TestFalconTechniques` | ✅ | ✅ green |
| 4-03-02 | 03 | 2 | CROW-02 | unit | `go test ./internal/playbooks/... -run TestNewTechniqueCount` | ✅ | ✅ green |
| 4-04-01 | 04 | 3 | CROW-03 | unit | `go test ./internal/reporter/... -run TestHTMLCrowdStrikeColumn` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/playbooks/loader_test.go` — add `TestSIEMCoverage` (verify `map[string][]string` parses from YAML with `siem_coverage` key)
- [x] `internal/playbooks/loader_test.go` — add `TestFalconTechniques` (3 FALCON_ files exist, have non-empty `siem_coverage.crowdstrike`, have `expected_events`)
- [x] `internal/playbooks/loader_test.go` — bump `TestNewTechniqueCount` threshold from 48 to 51 (adds 3 FALCON_ files) (actual threshold is 54 after Phase 5 additions)
- [x] `internal/reporter/reporter_test.go` — add `TestHTMLCrowdStrikeColumn` (column present when SIEMCoverage populated, absent when all entries have empty crowdstrike slice)

*All Wave 0 test functions were created during Phase 4 implementation. TestNewTechniqueCount threshold was bumped further to 54 during Phase 5. Tests confirmed passing 2026-03-26.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Falcon alert names match actual Falcon console | CROW-01 | No Falcon sandbox available in dev | Reviewer reads mapped names and confirms against Falcon documentation |
| CrowdStrike docs section in README readable | CROW-01 | Human readability check | Read the new German README section for CrowdStrike prerequisites |

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
