---
phase: 17-network-discovery
plan: "01"
subsystem: native-go-executor
tags: [t1046, network-discovery, goroutine-pool, tcp-scanner, sysmon]
dependency_graph:
  requires: [internal/native/registry.go, internal/engine/engine.go:detectLocalSubnet]
  provides: [T1046 native Go scanner, t1046_scan.go, t1046_scan_test.go]
  affects: [internal/native/, internal/playbooks/embedded/techniques/T1046_network_scan.yaml]
tech_stack:
  added: []
  patterns: [goroutine-pool-with-semaphore, build-tag-windows/!windows, init-registration, subnetHosts-exclusion]
key_files:
  created:
    - internal/native/t1046_scan.go
    - internal/native/t1046_scan_other.go
    - internal/native/t1046_scan_test.go
  modified:
    - internal/playbooks/embedded/techniques/T1046_network_scan.yaml
decisions:
  - "localSubnet() duplicated from engine.detectLocalSubnet() to avoid import cycle between internal/engine and internal/native"
  - "scanWorkers=15 as midpoint of D-08 range (10-20) for bounded concurrency"
  - "dialTimeout=300ms per research recommendation; covers LAN hosts, avoids excessive delay"
  - "UDP results are best-effort — Windows Firewall suppresses ICMP port-unreachable; noted in output"
metrics:
  duration: "148s"
  completed: "2026-04-10"
  tasks_completed: 2
  files_created: 3
  files_modified: 1
---

# Phase 17 Plan 01: T1046 Network Service Discovery Scanner Summary

**One-liner:** Native Go TCP/UDP connect scanner with 15-worker goroutine pool replacing PowerShell Tier 3 stub, generating Sysmon EID 3 burst from lognojutsu.exe PID across auto-detected /24 subnet.

## What Was Built

T1046 Network Service Discovery implemented as a native Go technique registered in the native executor registry. The scanner auto-detects the local /24 subnet, excludes own IPs, and scans 17 TCP ports + 3 UDP ports using a bounded goroutine pool (semaphore pattern). The T1046 YAML was rewritten from a PowerShell Tier 3 stub to a Go Tier 1 executor with `requires_confirmation: true`.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Implement T1046 TCP/UDP scanner with goroutine pool | 25b7736 | internal/native/t1046_scan.go, t1046_scan_other.go, t1046_scan_test.go |
| 2 | Rewrite T1046 YAML from PowerShell stub to Go executor | 5132008 | internal/playbooks/embedded/techniques/T1046_network_scan.yaml |

## Key Implementation Details

**t1046_scan.go (Windows build tag):**
- `tcpPorts`: 17 ports per D-04 (21,22,23,25,53,80,135,139,389,443,445,1433,3306,3389,5985,8080,8443)
- `udpPorts`: 3 ports per D-05 (53,161,123)
- `scanWorkers = 15`, `dialTimeout = 300ms`
- `localSubnet()`: duplicates `detectLocalSubnet()` from engine.go to avoid import cycle
- `localIPAddrs()`: collects all local IPv4s for self-exclusion
- `subnetHosts(cidr, excludeIPs...)`: generates .1-.254, excludes network/broadcast/own IPs
- `scanTCP()`: semaphore `make(chan struct{}, scanWorkers)` + `sync.WaitGroup` + `net.DialTimeout("tcp4",...)`
- `scanUDP()`: `net.Dial("udp4")` + null-byte probe + read with deadline
- `runT1046()`: orchestrates subnet detection → host enumeration → TCP scan → UDP scan → formatted output

**t1046_scan_test.go (8 unit tests):**
- TestSubnetHosts_ValidCIDR, TestSubnetHosts_ExcludesOwnIP, TestSubnetHosts_ExcludesNetworkAndBroadcast, TestSubnetHosts_InvalidCIDR
- TestLocalIPAddrs, TestTCPPortList, TestUDPPortList, TestT1046Registered

**T1046_network_scan.yaml changes:**
- `tier: 3` → `tier: 1`
- `executor.type: powershell` → `executor.type: go`
- `executor.command:` cleared (PowerShell runspace code removed)
- `requires_confirmation: true` added
- `expected_events` updated: removed EID 4104 (PowerShell ScriptBlock), kept EID 3 (Sysmon) + EID 5156 (WFP)

## Verification Results

```
go build ./...                          PASS (full project)
go test ./internal/native/... -run TestSubnetHosts    PASS (4 tests)
go test ./internal/native/... -run TestT1046          PASS (1 test)
go test ./internal/native/... -run TestLocalIPAddrs   PASS (1 test)
go test ./internal/native/... -run TestTCPPortList    PASS (1 test)
go test ./internal/native/... -run TestUDPPortList    PASS (1 test)
go test ./internal/playbooks/...                      PASS (14 tests)
grep "type: go" T1046_network_scan.yaml               FOUND (1 match)
grep "tier: 1" T1046_network_scan.yaml                FOUND (1 match)
```

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None — T1046 is a real scanner. UDP results note Windows Firewall suppression in output (intentional, documented as best-effort behavior).

## Self-Check: PASSED

Files exist:
- internal/native/t1046_scan.go: FOUND
- internal/native/t1046_scan_other.go: FOUND
- internal/native/t1046_scan_test.go: FOUND
- internal/playbooks/embedded/techniques/T1046_network_scan.yaml: FOUND (modified)

Commits exist:
- 25b7736: feat(17-01): implement T1046 TCP/UDP scanner with goroutine pool
- 5132008: feat(17-01): rewrite T1046 YAML from PowerShell stub to Go executor
