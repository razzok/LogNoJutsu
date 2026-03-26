---
phase: 5
slug: microsoft-sentinel-coverage
status: complete
nyquist_compliant: true
wave_0_complete: true
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
| 5-01-01 | 01 | 1 | SENT-01 | unit | `go test ./internal/playbooks/... -run TestSentinelCoverage` | ✅ | ✅ green |
| 5-02-01 | 02 | 1 | SENT-02 | unit | `go test ./internal/playbooks/... -run TestAzureTechniques` | ✅ | ✅ green |
| 5-03-01 | 03 | 1 | SENT-03 | unit | `go test ./internal/reporter/... -run TestHTMLSentinelColumn` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/playbooks/loader_test.go` — TestSentinelCoverage and TestAzureTechniques (originally planned as TestAZURETechniqueCount)
- [x] `internal/reporter/reporter_test.go` — TestHTMLSentinelColumn (originally planned as TestHasSentinel) covering HasSentinel=true and HasSentinel=false paths

*All Wave 0 test functions were created during Phase 5 implementation. Function names differ from original plan (TestAzureTechniques instead of TestAZURETechniqueCount, TestHTMLSentinelColumn instead of TestHasSentinel) but coverage is equivalent. Tests confirmed passing 2026-03-26.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| HTML report renders MS badge with correct blue color (#0078D4) | SENT-03 | Visual CSS check | Build binary, run simulation with AZURE_ technique, open HTML report, verify blue MS badge appears in Sentinel column |
| Sentinel column absent when no AZURE_ techniques in results | SENT-03 | Conditional rendering | Run simulation without AZURE_ techniques, verify Sentinel column does not appear in HTML output |

# Note: go test ./internal/playbooks/... may be blocked by Windows Defender quarantining playbooks.test.exe. Add %LOCALAPPDATA%\Temp\go-build* exclusion before running. Tests passed at implementation time; structural coverage is present.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 10s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved (2026-03-26)
