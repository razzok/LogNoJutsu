# Phase 17: Network Discovery - Context

**Gathered:** 2026-04-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Native Go network scanning to replace the Tier 3 T1046 PowerShell stub and add new T1018 Remote System Discovery. Both techniques use `type: go` executor, generate authentic Sysmon EID 3 artifacts, and integrate with the Phase 16 scan confirmation flow. Satisfies SCAN-01, SCAN-02, SCAN-03.

</domain>

<decisions>
## Implementation Decisions

### Scan Scope & Targeting
- **D-01:** Auto-detect /24 subnet only — reuse `detectLocalSubnet()` from `engine.go`. No configurable CIDR. Matches the scan confirmation modal which already displays this subnet.
- **D-02:** Exclude the host's own IP from scan results — self-scan noise isn't useful for SIEM detection.
- **D-03:** Rate limit at 50 connections/second — matches the Phase 16 scan confirmation modal text ("Rate limit: 50 connections/second"). Hard-coded, not configurable.

### Port & Protocol Selection
- **D-04:** Default port set: ~20 common service ports (21,22,23,25,53,80,135,139,389,443,445,1433,3306,3389,5985,8080,8443). Same as current stub — covers what SIEMs care about.
- **D-05:** TCP connect scan as primary method. Additionally scan 3 UDP ports: DNS(53), SNMP(161), NTP(123) for protocol variety.

### T1018 Discovery Methods
- **D-06:** T1018 is a single technique file (`type: go`) that runs all four discovery methods sequentially: (1) ICMP ping sweep of /24, (2) ARP table dump, (3) nltest DC discovery, (4) DNS reverse lookups on responding IPs.
- **D-07:** Single YAML file, not split into sub-techniques. One confirmation prompt covers the whole discovery chain.

### Sysmon Artifact Quality
- **D-08:** T1046 TCP scanner uses a goroutine pool (10-20 concurrent workers) making `net.DialTimeout()` calls. Creates burst of simultaneous TCP connections from the `lognojutsu.exe` PID — same signature as real port scanners. Sysmon sees multiple concurrent NetworkConnect (EID 3) events from one process.
- **D-09:** T1018 ICMP ping sweep uses `net.Dial("ip4:icmp")` when running as admin. Falls back to TCP connect on ports 445/135 as host-alive check when not admin. Both generate Sysmon EID 3 events.

### Scan Confirmation Integration
- **D-10:** Both T1046 and T1018 YAML files set `requires_confirmation: true`. The Phase 16 scan confirmation modal fires before either technique executes.
- **D-11:** T1046 is upgraded from Tier 3 → Tier 1 (real scanner). T1018 is new, created as Tier 1.

### Claude's Discretion
- Connection timeout values for TCP dial (200-500ms range)
- UDP scan implementation details (Go `net.Dial("udp4",...)` with timeout)
- ARP table reading approach (parse `arp -a` output or use Go syscalls)
- nltest wrapper implementation (exec nltest.exe or native Go alternative)
- DNS reverse lookup implementation details
- Worker pool size within the 10-20 range
- Output formatting of scan results (tables, lists, etc.)
- Test structure and mocking strategy

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase 15 Context (native Go executor pattern)
- `.planning/phases/15-native-go-architecture/15-CONTEXT.md` — Register/Lookup pattern (D-01 to D-12), NativeFunc signature, executor dispatch, cleanup pattern

### Phase 16 Context (scan confirmation + safety)
- `.planning/phases/16-safety-infrastructure/16-CONTEXT.md` — Scan confirmation flow (D-07 to D-09), RequiresConfirmation field, detectLocalSubnet()

### Existing Code
- `internal/native/registry.go` — NativeFunc/CleanupFunc signatures, Register/Lookup API
- `internal/native/t1482_ldap.go` — Reference implementation for Go technique (build tags, DC discovery, graceful fallback)
- `internal/engine/engine.go:detectLocalSubnet()` — Reuse for scan target detection
- `internal/engine/engine.go:runScanConfirmation()` — Confirmation flow that T1046/T1018 will trigger
- `internal/playbooks/embedded/techniques/T1046_network_scan.yaml` — Current Tier 3 stub to replace

### Requirements
- `.planning/REQUIREMENTS.md` — SCAN-01, SCAN-02, SCAN-03

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `native.Register()` pattern — T1046 and T1018 register as Go techniques in `init()`
- `detectLocalSubnet()` in engine.go — returns /24 CIDR string, reuse for scan target
- `NativeResult{Output, ErrorOutput, Success}` — scan results map to this struct
- `RequiresConfirmation bool` field in Technique struct — already wired to confirmation modal
- `ScanInfo` struct — already has TargetSubnet, RateLimitNote, IDSWarning, Techniques fields

### Established Patterns
- Build tag `//go:build windows` for Windows-specific code (ICMP, ARP)
- Permissive stub on non-Windows (`//go:build !windows`) returning graceful fallback
- `init()` registration — each technique file self-registers
- Error wrapping with `fmt.Errorf("...: %w", err)`

### Integration Points
- `internal/native/` — Add `t1046_scan.go` and `t1018_discovery.go`
- `internal/playbooks/embedded/techniques/T1046_network_scan.yaml` — Rewrite: change `type: powershell` → `type: go`, set `tier: 1`, add `requires_confirmation: true`
- New `internal/playbooks/embedded/techniques/T1018_remote_system_discovery.yaml` — New technique file
- Engine scan confirmation flow already handles `RequiresConfirmation` techniques

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches within the decisions above.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 17-network-discovery*
*Context gathered: 2026-04-10*
