---
phase: 1
slug: events-manifest-verification-engine
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-24
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (standard library) |
| **Config file** | none — Wave 0 creates first test files |
| **Quick run command** | `go test ./internal/verifier/... ./internal/playbooks/... -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/verifier/... ./internal/playbooks/... -v`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| EventSpec YAML parse | types | 1 | VERIF-01 | unit | `go test ./internal/playbooks/... -run TestEventSpecParsing` | ✅ | ✅ green |
| queryCount mock | verifier | 1 | VERIF-02 | unit (mock) | `go test ./internal/verifier/... -run TestQueryCountMock` | ✅ | ✅ green |
| determineStatus logic | verifier | 1 | VERIF-03 | unit | `go test ./internal/verifier/... -run TestDetermineStatus` | ✅ | ✅ green |
| HTML verification column | reporter | 2 | VERIF-04 | unit | `go test ./internal/reporter/... -run TestHTMLVerificationColumn` | ✅ | ✅ green |
| NotExecuted vs EventsMissing | verifier | 1 | VERIF-05 | unit | `go test ./internal/verifier/... -run TestNotExecutedVsEventsMissing` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/verifier/verifier_test.go` — stubs for VERIF-02, VERIF-03, VERIF-05
- [x] `internal/playbooks/types_test.go` — stubs for VERIF-01 (EventSpec YAML parsing)
- [x] `internal/reporter/reporter_test.go` — stubs for VERIF-04 (HTML verification column)

*Note: All Wave 0 test files were created during phase implementation. Tests confirmed passing 2026-03-26.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Verification queries real Windows Event Log after technique runs | VERIF-02 (integration) | Requires Windows + real Event Log writes | Run a technique, check simlog for "Verified: pass/fail" entry |
| HTML report verification column renders in browser | VERIF-04 (visual) | Template rendering is unit-tested but layout is visual | Open generated HTML report, verify column appears with ✓/✗ badges |
| 3-second wait correctly captures events | VERIF-02 (timing) | Timing is environment-dependent | Run technique, observe events in local Event Log with timestamp comparison |

---

## Testability Notes

**VERIF-02 injectable queryCount:** The `queryCount` function must accept a mockable dependency (function type or interface) so PowerShell subprocess calls can be replaced in unit tests. The actual subprocess only runs in integration/manual tests on Windows.

**VERIF-01 YAML migration:** All 43 technique YAML files must be updated from `expected_events: ["string"]` to `expected_events: [{event_id: N, channel: "...", description: "..."}]`. The parsing test validates the new schema is correctly deserialized.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved (2026-03-26)
