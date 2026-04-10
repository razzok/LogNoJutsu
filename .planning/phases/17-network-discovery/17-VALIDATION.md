---
phase: 17
slug: network-discovery
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-10
---

# Phase 17 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (`testing` package) |
| **Config file** | None — `go test ./...` |
| **Quick run command** | `go test ./internal/native/... -v -run TestT104` |
| **Full suite command** | `go test ./internal/native/... ./internal/playbooks/... -v` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/native/... -v -run TestT104`
- **After every plan wave:** Run `go test ./internal/native/... ./internal/playbooks/... -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 17-01-01 | 01 | 1 | SCAN-01 | unit | `go test ./internal/native/... -run TestT1046` | No — Wave 0 | ⬜ pending |
| 17-01-02 | 01 | 1 | SCAN-01 | unit | `go test ./internal/native/... -run TestT1046Pool` | No — Wave 0 | ⬜ pending |
| 17-01-03 | 01 | 1 | SCAN-01 | unit | `go test ./internal/native/... -run TestT1046ExcludesOwnIP` | No — Wave 0 | ⬜ pending |
| 17-02-01 | 02 | 1 | SCAN-02 | unit | `go test ./internal/native/... -run TestT1018ARP` | No — Wave 0 | ⬜ pending |
| 17-02-02 | 02 | 1 | SCAN-02 | unit | `go test ./internal/native/... -run TestT1018NltestFallback` | No — Wave 0 | ⬜ pending |
| 17-02-03 | 02 | 1 | SCAN-02 | unit | `go test ./internal/native/... -run TestT1018DNS` | No — Wave 0 | ⬜ pending |
| 17-03-01 | 03 | 2 | SCAN-03 | unit | `go test ./internal/playbooks/... -run TestT1046YAML` | No — Wave 0 | ⬜ pending |
| 17-03-02 | 03 | 2 | SCAN-03 | unit | `go test ./internal/playbooks/... -run TestT1018YAML` | No — Wave 0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/native/t1046_scan_test.go` — stubs for SCAN-01 (TCP scan, pool concurrency, own-IP exclusion)
- [ ] `internal/native/t1018_discovery_test.go` — stubs for SCAN-02 (ARP parsing, nltest fallback, DNS reverse lookup)
- [ ] `internal/playbooks/technique_yaml_test.go` — stubs for SCAN-03 (YAML field validation for T1046, T1018)

*Existing Go test infrastructure covers all framework requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| TCP scan produces Sysmon EID 3 events | SCAN-01 | Requires Sysmon installed and Event Log access | Run T1046 technique, check Event Viewer > Sysmon > EID 3 for burst of NetworkConnect from lognojutsu.exe |
| ICMP ping sweep requires admin | SCAN-02 | Requires elevation toggle | Run T1018 as admin vs non-admin, verify ICMP path vs TCP fallback |
| Scan confirmation modal fires | SCAN-03 | UI interaction | Execute T1046/T1018 via UI, verify confirmation modal appears before scan starts |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
