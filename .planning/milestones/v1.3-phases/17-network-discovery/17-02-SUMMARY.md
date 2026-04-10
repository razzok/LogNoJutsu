---
phase: 17-network-discovery
plan: 02
subsystem: native
tags: [go, native, t1018, discovery, icmp, arp, nltest, dns, sysmon]

# Dependency graph
requires:
  - phase: 17-01
    provides: subnetHosts(), localSubnet(), localIPAddrs() helpers; T1046 scanner as pattern reference

provides:
  - T1018 Remote System Discovery native Go implementation (ICMP/ARP/nltest/DNS chain)
  - parseARPTable() for structured ARP output parsing
  - icmpChecksum() RFC 1071 implementation
  - tcpAliveCheck() non-admin fallback (ports 445/135)
  - nltestDCDiscovery() with graceful fallback on non-domain machines
  - dnsReverseLookup() on union of ICMP+ARP alive hosts
  - T1018_remote_system_discovery.yaml with type: go, tier: 1, requires_confirmation: true

affects:
  - internal/executor (dispatches T1018 via native registry)
  - internal/playbooks (YAML loaded into technique library)
  - reporter (T1018 appears in HTML report)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "TDD RED/GREEN: test file committed before implementation, tests confirmed failing before writing code"
    - "Four-method discovery chain: ICMP sweep -> ARP dump -> nltest DC -> DNS PTR in runT1018()"
    - "Admin/non-admin bifurcation: net.ListenPacket error signals fallback to TCP 445/135"
    - "strings.Fields ARP parsing: field[0] IP validation via net.ParseIP, field[1] MAC"
    - "Union set for DNS targets: ICMP alive hosts + ARP IP entries, deduped via seen map"

key-files:
  created:
    - internal/native/t1018_discovery.go
    - internal/native/t1018_discovery_other.go
    - internal/native/t1018_discovery_test.go
    - internal/playbooks/embedded/techniques/T1018_remote_system_discovery.yaml
  modified: []

key-decisions:
  - "ICMP fallback to TCP 445/135: net.ListenPacket(\"ip4:icmp\") error signals non-admin; no separate privilege check needed"
  - "strings.Fields ARP parsing: handles varied whitespace from arp -a output; net.ParseIP guards header lines"
  - "Union set for DNS lookup: merge ICMP alive hosts and ARP IP entries with seen-map dedup"
  - "nltest graceful fallback: distinguish 'not found' (exec error) from 'non-domain exit' (non-zero exit code)"

patterns-established:
  - "Four-method discovery chain pattern: ICMP -> ARP -> nltest -> DNS with section headers"
  - "Admin-or-fallback ICMP pattern: attempt raw socket, delegate to TCP alive check on error"

requirements-completed: [SCAN-02, SCAN-03]

# Metrics
duration: 35min
completed: 2026-04-10
---

# Phase 17 Plan 02: T1018 Remote System Discovery Summary

**T1018 four-method discovery chain in native Go: ICMP ping sweep with RFC 1071 checksum, ARP table parsing, nltest DC discovery, and DNS reverse lookups — all with graceful fallbacks.**

## Performance

- **Duration:** ~35 min
- **Started:** 2026-04-10T00:00:00Z
- **Completed:** 2026-04-10T00:35:00Z
- **Tasks:** 2 completed
- **Files created:** 4

## Accomplishments

- Implemented T1018 as a native Go technique registering in the same package as T1046 and reusing its helpers (subnetHosts, localSubnet, localIPAddrs) without duplication
- ICMP ping sweep uses hand-crafted echo request with RFC 1071 checksum; falls back to TCP 445/135 alive check when raw sockets fail (non-admin path per D-09)
- ARP table dump parses `arp -a` output using strings.Fields with net.ParseIP guard to skip header lines
- nltest.exe DC discovery distinguishes "binary not found" from "non-zero exit" (not domain-joined) for accurate graceful fallback messaging
- DNS reverse lookups run on union of ICMP alive hosts and ARP-discovered IPs via seen-map dedup
- All 7 unit tests pass (TestParseARPTable x3, TestNltestDCDiscovery, TestDnsReverseLookup, TestIcmpChecksum, TestT1018Registered)
- YAML file created with type: go, tier: 1, requires_confirmation: true, elevation_required: false

## Task Commits

Each task committed atomically via TDD:

1. **TDD RED - T1018 failing tests** - `c0c2e02` (test)
2. **T1018 implementation + non-Windows stub** - `66a0ed7` (feat)
3. **T1018 YAML technique file** - `b488960` (feat)

## Files Created/Modified

- `internal/native/t1018_discovery.go` - Windows-only T1018 implementation: icmpPingSweep, tcpAliveCheck, parseARPTable, arpTableDump, nltestDCDiscovery, dnsReverseLookup, runT1018, init()
- `internal/native/t1018_discovery_other.go` - Non-Windows permissive stub with !windows build tag
- `internal/native/t1018_discovery_test.go` - 7 unit tests: ARP parsing (3), nltest fallback, DNS lookup, ICMP checksum, registry check
- `internal/playbooks/embedded/techniques/T1018_remote_system_discovery.yaml` - Tier 1 Go technique with requires_confirmation, Sysmon EID 3 + Security EID 4688 expected events

## Decisions Made

- **ICMP fallback via net.ListenPacket error**: On non-admin, ListenPacket("ip4:icmp") returns an error — this is used directly as the signal to call tcpAliveCheck() without any separate privilege detection
- **strings.Fields for ARP parsing**: Handles any whitespace layout from arp -a; net.ParseIP(fields[0]) == nil skips header lines cleanly
- **nltest fallback strategy**: Check exec.Error for "executable file not found" vs. non-zero exit for "not domain-joined" — different messages for different root causes
- **DNS targets union**: Combine ICMP alive hosts and ARP IP entries with a seen map to avoid duplicate lookups

## Deviations from Plan

None — plan executed exactly as written. All acceptance criteria met. All tests pass.

## Known Stubs

None — T1018 is fully wired to native registry and YAML executor. No placeholder data.

## Self-Check: PASSED

Files created:
- internal/native/t1018_discovery.go — FOUND
- internal/native/t1018_discovery_other.go — FOUND
- internal/native/t1018_discovery_test.go — FOUND
- internal/playbooks/embedded/techniques/T1018_remote_system_discovery.yaml — FOUND

Commits verified:
- c0c2e02 (TDD RED tests) — FOUND
- 66a0ed7 (implementation) — FOUND
- b488960 (YAML file) — FOUND

Build: `go build ./...` — PASS
Tests: `go test ./internal/native/...` — 26 PASS, 0 FAIL
Tests: `go test ./internal/playbooks/...` — 14 PASS, 0 FAIL
