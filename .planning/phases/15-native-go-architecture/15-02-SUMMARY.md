---
phase: 15-native-go-architecture
plan: "02"
subsystem: native-techniques
tags: [go-native, ldap, wmi, t1482, t1057, technique-upgrade]
dependency_graph:
  requires: ["15-01"]
  provides: ["T1482-native", "T1057-native"]
  affects: ["internal/native", "internal/playbooks", "go.mod"]
tech_stack:
  added:
    - "github.com/go-ldap/ldap/v3 v3.4.13 (promoted to direct dependency)"
    - "github.com/yusufpapurcu/wmi v1.2.4 (promoted to direct dependency)"
  patterns:
    - "init() registration: Go technique file calls native.Register() in init(); no explicit wiring needed"
    - "Explicit WMI class name: pass Win32_Process string to wmi.CreateQuery when struct name differs"
    - "DC auto-discovery chain: LOGONSERVER env var -> DNS SRV -> ErrNoDCReachable sentinel"
    - "NTLM bind with empty credentials: current user's Windows session token used automatically"
key_files:
  created:
    - internal/native/t1482_ldap.go
    - internal/native/t1482_test.go
    - internal/native/t1057_wmi.go
    - internal/native/t1057_test.go
  modified:
    - internal/playbooks/embedded/techniques/T1482_domain_trust_discovery.yaml
    - internal/playbooks/embedded/techniques/T1057_process_discovery.yaml
    - internal/playbooks/loader_test.go
    - go.mod
    - go.sum
decisions:
  - "Pass WMI class name explicitly to wmi.CreateQuery when Go struct name differs from WMI class name (win32Process vs Win32_Process)"
  - "T1057 tier promoted from 3 to 1 — WMI Win32_Process query is a real implementation, not a stub"
  - "T1482 expected_events retained as-is; they will be refined in Phase 18 when discovery techniques are overhauled"
metrics:
  duration: "4 minutes"
  completed: "2026-04-09"
  tasks_completed: 2
  files_modified: 9
---

# Phase 15 Plan 02: T1482 LDAP and T1057 WMI Native Techniques Summary

T1482 Domain Trust Discovery via LDAP with DC auto-discovery and graceful fallback, and T1057 Process Discovery via WMI Win32_Process query — both upgraded from PowerShell stubs to Tier 1 real Go implementations.

## Tasks Completed

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | T1482 LDAP + T1057 WMI implementations with tests (TDD) | cbc2cc5 | t1482_ldap.go, t1482_test.go, t1057_wmi.go, t1057_test.go |
| 2 | Update YAMLs to type:go, loader tests, go mod tidy | c3a7436 | T1482.yaml, T1057.yaml, loader_test.go, go.mod |

## What Was Built

### T1482 — Domain Trust Discovery via LDAP

`internal/native/t1482_ldap.go` implements:
- `discoverDC()` — auto-discovers domain controller via LOGONSERVER env var then DNS SRV (`_ldap._tcp.dc._msdcs.<USERDNSDOMAIN>`)
- `ErrNoDCReachable` — exported sentinel error for graceful fallback path
- `constructBaseDN(domain)` — converts DNS domain to LDAP DN under CN=System
- `runT1482()` — LDAP dial with 5s timeout, NTLM bind (current user token), RootDSE query for defaultNamingContext, trustedDomain object search
- Graceful fallback at every failure point: no DC, dial failure, bind failure each return Success=false with descriptive ErrorOutput
- Registers via `init()`: `Register("T1482", runT1482, nil)`

### T1057 — Process Discovery via WMI

`internal/native/t1057_wmi.go` implements:
- `win32Process` struct binding WMI fields (Name, ProcessId, ParentProcessId, CommandLine)
- `runT1057()` — Win32_Process WMI query via `wmi.Query`, formats output with PID/PPID/name/cmdline, limits to first 50 processes
- Registers via `init()`: `Register("T1057", runT1057, nil)`

### YAML Updates

Both techniques updated to `executor.type: go` with `executor.command: ""`. T1057 tier promoted from 3 to 1. All metadata (expected_events, tags, siem_coverage) preserved.

### Loader Tests Added

`TestT1482ExecutorType` and `TestT1057ExecutorType` verify the YAML executor type changes and T1057 tier promotion.

## Test Results

All tests pass:
- `TestT1482NoDC` — verifies ErrNoDCReachable when no env vars set
- `TestT1482DiscoverDCFromLogonserver` — verifies DC=DC01:389 from LOGONSERVER=\\DC01
- `TestT1482GracefulFallback` — verifies runT1482() returns Success=false with "no domain controller reachable"
- `TestT1057WMI` — verifies Win32_Process query returns Success=true with "processes found"
- `TestT1057WMIHasResults` — verifies output contains at least one PID= line
- Full `go test ./...` green after both tasks

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] WMI struct name mismatch with Win32_Process class**
- **Found during:** Task 1 GREEN phase — TestT1057WMI/TestT1057WMIHasResults failed with "Exception occurred. (Invalid class )"
- **Issue:** `wmi.CreateQuery` derives the WMI class name from the Go struct type name. The struct was named `win32Process` (lowercase) but WMI expects `Win32_Process`. The generated query was `SELECT ... FROM win32Process` which is not a valid WMI class.
- **Fix:** Pass the class name explicitly as a variadic argument: `wmi.CreateQuery(&procs, "", "Win32_Process")` — this is supported by the `wmi.CreateQuery(src interface{}, where string, class ...string)` signature.
- **Files modified:** `internal/native/t1057_wmi.go`
- **Commit:** cbc2cc5

## Known Stubs

None — both techniques are fully implemented real queries (not stubs).

## Self-Check: PASSED

Files exist:
- FOUND: /d/Code/LogNoJutsu/internal/native/t1482_ldap.go
- FOUND: /d/Code/LogNoJutsu/internal/native/t1482_test.go
- FOUND: /d/Code/LogNoJutsu/internal/native/t1057_wmi.go
- FOUND: /d/Code/LogNoJutsu/internal/native/t1057_test.go
- FOUND: /d/Code/LogNoJutsu/internal/playbooks/embedded/techniques/T1482_domain_trust_discovery.yaml (updated)
- FOUND: /d/Code/LogNoJutsu/internal/playbooks/embedded/techniques/T1057_process_discovery.yaml (updated)

Commits exist:
- cbc2cc5: feat(15-02): implement T1482 LDAP trust discovery and T1057 WMI process discovery
- c3a7436: feat(15-02): update T1482 and T1057 YAMLs to type:go and promote to direct deps
