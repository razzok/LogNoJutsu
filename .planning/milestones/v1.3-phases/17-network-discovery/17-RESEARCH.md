# Phase 17: Network Discovery - Research

**Researched:** 2026-04-10
**Domain:** Native Go network scanning — TCP connect scan (T1046), ICMP/ARP/DC discovery (T1018)
**Confidence:** HIGH

## Summary

Phase 17 replaces the existing T1046 PowerShell stub with a native Go TCP connect scanner and adds a new T1018 Remote System Discovery technique that chains ICMP ping sweep, ARP table dump, nltest DC discovery, and DNS reverse lookups. Both techniques use the established `type: go` executor pattern from Phase 15. The codebase already has the full infrastructure: native registry, executor dispatch, scan confirmation modal, and `detectLocalSubnet()`. This phase is purely additive — two new `.go` files in `internal/native/`, a YAML rewrite for T1046, and a new YAML file for T1018.

All decisions in CONTEXT.md are concrete and locked. There are no library choices to make — the Go standard library (`net`, `os/exec`) covers everything required. The only discretion areas are timeout values, worker pool size, ARP parsing approach, and output formatting.

**Primary recommendation:** Use `net.DialTimeout("tcp4", ...)` with a goroutine pool in a semaphore pattern (channel-based, 15 workers) for T1046, and `os/exec.Command("arp", "-a")` output parsing for ARP table in T1018. ICMP requires a raw socket (`net.ListenPacket("ip4:icmp", ...)`) and administrator privileges — fall back to TCP 445/135 alive-checks when not admin, consistent with D-09.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Scan Scope & Targeting**
- D-01: Auto-detect /24 subnet only — reuse `detectLocalSubnet()` from `engine.go`. No configurable CIDR.
- D-02: Exclude the host's own IP from scan results.
- D-03: Rate limit at 50 connections/second — hard-coded.

**Port & Protocol Selection**
- D-04: Default port set: 21,22,23,25,53,80,135,139,389,443,445,1433,3306,3389,5985,8080,8443
- D-05: TCP connect scan primary. Additionally scan 3 UDP ports: DNS(53), SNMP(161), NTP(123).

**T1018 Discovery Methods**
- D-06: T1018 runs all four discovery methods sequentially: (1) ICMP ping sweep of /24, (2) ARP table dump, (3) nltest DC discovery, (4) DNS reverse lookups on responding IPs.
- D-07: Single YAML file, not split into sub-techniques.

**Sysmon Artifact Quality**
- D-08: T1046 uses goroutine pool (10-20 concurrent workers) with `net.DialTimeout()`. Creates burst of simultaneous TCP connections from `lognojutsu.exe` PID.
- D-09: T1018 ICMP ping sweep uses `net.Dial("ip4:icmp")` when admin. Falls back to TCP 445/135 when not admin.

**Scan Confirmation Integration**
- D-10: Both T1046 and T1018 set `requires_confirmation: true`.
- D-11: T1046 upgraded Tier 3 → Tier 1. T1018 created as Tier 1.

### Claude's Discretion
- Connection timeout values for TCP dial (200-500ms range)
- UDP scan implementation details
- ARP table reading approach (parse `arp -a` output or use Go syscalls)
- nltest wrapper implementation (exec nltest.exe or native Go alternative)
- DNS reverse lookup implementation details
- Worker pool size within the 10-20 range
- Output formatting of scan results

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SCAN-01 | T1046 scans the host's auto-detected /24 subnet via TCP connect scan (not just loopback/gateway) | `detectLocalSubnet()` returns /24 CIDR; goroutine pool dials all 17 ports on all 254 hosts |
| SCAN-02 | T1018 includes ICMP ping sweep, ARP table enumeration, and nltest DC discovery | All three implemented in `runT1018()`; DNS reverse lookup fourth method per D-06 |
| SCAN-03 | Network scanning implemented as native Go (`type: go` executor) generating Sysmon EID 3 artifacts | `net.DialTimeout()` TCP connections from lognojutsu.exe PID trigger Sysmon EID 3 events |
</phase_requirements>

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `net` (stdlib) | Go 1.26 | TCP/UDP dial, ICMP raw socket, DNS lookup | No external dep; `net.DialTimeout` is the correct Go TCP connect pattern |
| `os/exec` (stdlib) | Go 1.26 | Run `arp -a`, `nltest /dsgetdc:` | Shell-out is the established pattern in this codebase (T1482 uses exec pattern too) |
| `sync` (stdlib) | Go 1.26 | WaitGroup for goroutine pool | Standard concurrency primitive; already used in codebase |
| `strings`, `fmt` (stdlib) | Go 1.26 | Output formatting, string parsing | Already in every native technique file |

No new `go.mod` dependencies are required. Everything needed is in the Go standard library.

### No New External Dependencies

The existing `go.mod` already has `golang.org/x/sys v0.41.0` (used by wmi/windows packages). ICMP via raw socket works with `net.ListenPacket("ip4:icmp", "0.0.0.0")` — no external package needed. Windows raw socket privilege requirement (must be admin) is addressed by the D-09 fallback.

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `net.DialTimeout` (TCP) | `golang.org/x/net/ipv4` | Overkill — raw packet crafting not needed for TCP connect scan |
| Parse `arp -a` stdout | Windows `iphlpapi.dll` via `x/sys/windows` | More reliable but much more complex; `arp -a` output is stable on all Windows versions and matches existing codebase style |
| `os/exec("nltest", ...)` | Pure Go LDAP DC discovery (already in T1482) | T1482's DC discovery is already available; nltest provides different artifact (4688 process creation) which is what T1018 needs for SIEM realism |

---

## Architecture Patterns

### Recommended Project Structure

```
internal/native/
├── registry.go              # Existing — NativeFunc/CleanupFunc/Register/Lookup
├── t1482_ldap.go            # Existing — reference Go technique implementation
├── t1057_wmi.go             # Existing — reference Go technique implementation
├── t1046_scan.go            # New — TCP connect scanner (windows build tag)
├── t1046_scan_other.go      # New — permissive stub for non-windows
├── t1018_discovery.go       # New — ICMP/ARP/nltest/DNS discovery (windows build tag)
└── t1018_discovery_other.go # New — permissive stub for non-windows

internal/playbooks/embedded/techniques/
├── T1046_network_scan.yaml             # Rewrite: type: go, tier: 1, requires_confirmation: true
└── T1018_remote_system_discovery.yaml  # New — type: go, tier: 1, requires_confirmation: true
```

### Pattern 1: Goroutine Pool with Semaphore (T1046)

**What:** Channel-based semaphore limits simultaneous TCP connections to control the burst rate. WaitGroup ensures all goroutines complete before returning.

**When to use:** Any scan that dials N targets with bounded concurrency — directly matches D-08 (10-20 workers creating burst Sysmon EID 3 signature).

```go
// Source: Go standard library patterns / existing codebase conventions
func scanSubnet(subnet string, ports []int, workerCount int) []scanResult {
    // Parse subnet to get all 254 host IPs (exclude host and broadcast)
    // Build work channel of (ip, port) pairs
    sem := make(chan struct{}, workerCount) // semaphore limits concurrency
    var wg sync.WaitGroup
    results := make(chan scanResult, len(hosts)*len(ports))

    for _, host := range hosts {
        for _, port := range ports {
            wg.Add(1)
            go func(h string, p int) {
                defer wg.Done()
                sem <- struct{}{}        // acquire slot
                defer func() { <-sem }() // release slot
                addr := fmt.Sprintf("%s:%d", h, p)
                conn, err := net.DialTimeout("tcp4", addr, 300*time.Millisecond)
                if err == nil {
                    conn.Close()
                    results <- scanResult{Host: h, Port: p, Open: true}
                }
            }(host, port)
        }
    }
    wg.Wait()
    close(results)
    // collect results
}
```

**Worker count recommendation:** 15 (midpoint of D-08's 10-20 range). Produces clear Sysmon EID 3 burst without overwhelming non-production networks.

**Timeout recommendation:** 300ms per connection (midpoint of D-03's 200-500ms range). Fast enough to scan /24 × 17 ports in under 2 minutes at 15 workers.

### Pattern 2: ICMP Raw Socket with Admin Fallback (T1018)

**What:** Try `net.ListenPacket("ip4:icmp", "0.0.0.0")` — succeeds when process is admin (raw socket requires elevated privileges on Windows). On failure, fall back to TCP 445/135 alive-check per D-09.

```go
// Source: Go net package documentation
func icmpPingSweep(hosts []string) []string {
    conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
    if err != nil {
        // Not admin — fall back to TCP port check (D-09)
        return tcpAliveCheck(hosts, []int{445, 135}, 300*time.Millisecond)
    }
    defer conn.Close()
    // Send ICMP echo request to each host; collect replies
    // Each Dial/Listen generates Sysmon EID 3
}
```

**ICMP packet construction:** Use `golang.org/x/net/icmp` OR hand-craft the 8-byte ICMP echo request header manually (type=8, code=0, checksum, id, seq). Since `golang.org/x/net` is not in go.mod, hand-craft to avoid adding a dependency.

**ICMP echo request bytes (hand-crafted):**
```go
// Type=8 (Echo Request), Code=0, Checksum (computed), ID, Seq
msg := []byte{8, 0, 0, 0, byte(id >> 8), byte(id), 0, 1}
checksum := icmpChecksum(msg)
msg[2] = byte(checksum >> 8)
msg[3] = byte(checksum)
```

### Pattern 3: ARP Table Parsing (T1018)

**What:** Execute `arp -a` and parse the output. Output format on Windows is stable: lines contain `<IP>   <MAC>   <Type>`.

**Approach:** `os/exec.Command("arp", "-a")` — consistent with how this codebase already uses exec for external tools. Parse with `strings.Fields()` to extract IP and MAC columns.

```go
out, err := exec.Command("arp", "-a").Output()
// Parse lines: skip headers (no IP-format token in field[0])
// ip, mac := fields[0], fields[1]
```

**Why not syscall:** `iphlpapi.GetIpNetTable` requires CGO or `x/sys/windows` interop for struct marshaling. The `arp -a` approach is simpler, produces the same data, and is consistent with the codebase's approach to tool discovery (T1482's discoverDC also uses env vars before doing complex queries).

### Pattern 4: nltest DC Discovery (T1018)

**What:** Execute `nltest.exe /dsgetdc:<domain>` — generates Windows Event 4688 (process creation), exactly the artifact SIEMs monitor for DC discovery reconnaissance.

```go
domain := os.Getenv("USERDNSDOMAIN")
if domain == "" {
    domain = "." // query default domain
}
out, err := exec.Command("nltest.exe", "/dsgetdc:"+domain).CombinedOutput()
// Parse output for "DC:" line
```

**Graceful fallback:** If `nltest.exe` is not found (non-domain joined) or returns error, record "nltest: DC not found (non-domain environment)" in output and continue to next discovery method. Same pattern as T1482's `ErrNoDCReachable`.

### Pattern 5: DNS Reverse Lookup (T1018)

**What:** Call `net.LookupAddr(ip)` for each IP that responded to ICMP/TCP alive check. Returns PTR records. Standard Go stdlib, no exec needed.

```go
for _, ip := range aliveHosts {
    names, err := net.LookupAddr(ip)
    if err == nil && len(names) > 0 {
        fmt.Fprintf(&sb, "  PTR: %s -> %s\n", ip, strings.Join(names, ", "))
    }
}
```

**Timeout note:** `net.LookupAddr` uses the system DNS resolver. On Windows this honors the system DNS timeout. No explicit timeout needed for a small set of hosts.

### Pattern 6: Subnet Host Enumeration

**What:** Parse the /24 CIDR string from `detectLocalSubnet()` to generate the 254 host addresses, excluding the own IP (D-02).

```go
func subnetHosts(cidr string, excludeIP string) ([]string, error) {
    ip, ipnet, err := net.ParseCIDR(cidr)
    // iterate ipnet from x.x.x.1 to x.x.x.254
    // skip if ip == excludeIP
    // skip network address (host bits all 0) and broadcast (host bits all 1)
}
```

**Own IP detection:** `net.Interfaces()` → collect all local IPv4 addresses → exclude any host in the scan list that matches.

### YAML File Structure

**T1046 rewrite** (key changes only):
```yaml
id: T1046
tier: 1
requires_confirmation: true
executor:
  type: go
  command: ""
```

**T1018 new file:**
```yaml
id: T1018
name: Remote System Discovery
tactic: discovery
technique_id: T1018
platform: windows
phase: discovery
elevation_required: false
tier: 1
requires_confirmation: true
executor:
  type: go
  command: ""
expected_events:
  - event_id: 3
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "NetworkConnect - ICMP/TCP alive checks from lognojutsu.exe"
  - event_id: 4688
    channel: "Security"
    description: "Process Creation - nltest.exe DC discovery"
cleanup: ""
```

### Anti-Patterns to Avoid

- **Scanning localhost only:** The whole point of SCAN-01 is subnet-wide scanning. `detectLocalSubnet()` must be called — do not hard-code `127.0.0.1`.
- **Unbounded goroutine spawning:** Launching a goroutine per (host × port) without a semaphore creates 254 × 17 = 4318 goroutines. Use the channel semaphore pattern.
- **Single-threaded scan:** Sequential `net.DialTimeout` across 4318 combinations takes 20+ minutes at 300ms timeout. The goroutine pool is required for realistic burst behavior.
- **Blocking indefinitely on ICMP:** Raw socket operations can block. Set a deadline with `conn.SetDeadline(time.Now().Add(timeout))` before the read loop.
- **Treating UDP scan success/failure as equivalent to TCP:** UDP `net.Dial` always succeeds (UDP is connectionless). You send a packet and wait for a response; closed ports return ICMP port unreachable. This is inherently unreliable without raw socket access. Keep UDP output low-confidence in the results.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Subnet IP list generation | Custom bit-math parser | `net.ParseCIDR()` + increment loop | stdlib handles edge cases (broadcast, network addr) |
| Connection pooling | Custom goroutine lifecycle | Channel semaphore + WaitGroup | Proven pattern; simpler than sync.Pool for this use case |
| DNS PTR lookup | Custom DNS packet builder | `net.LookupAddr()` | stdlib respects system DNS; returns PTR records directly |
| CIDR membership check | Bit masking | `ipnet.Contains(ip)` | stdlib method; correct and tested |

**Key insight:** This phase touches no complex network protocol internals. TCP connect scan, ARP parsing, and DNS PTR are all one-liners in Go stdlib. The complexity is in orchestration (goroutine pool, result collection, graceful fallback) not in protocol implementation.

---

## Common Pitfalls

### Pitfall 1: Raw ICMP Requires Elevated Privileges on Windows
**What goes wrong:** `net.ListenPacket("ip4:icmp", "0.0.0.0")` returns "operation not permitted" when not running as Administrator. This is a Windows raw socket restriction.
**Why it happens:** Windows restricts raw socket access to privileged processes (same restriction as Nmap requiring admin on Windows).
**How to avoid:** Check the error from `net.ListenPacket` — if non-nil, switch to TCP 445/135 alive-check fallback per D-09. This is already specified in the decisions.
**Warning signs:** Error contains "An attempt was made to access a socket in a way forbidden by its access permissions" or "operation not permitted".

### Pitfall 2: UDP Scan Produces No Results on Windows Without Admin
**What goes wrong:** UDP port scan via `net.Dial("udp4", ...)` then sending a payload and reading response — Windows firewall silently drops UDP ICMP port unreachable responses for most ports.
**Why it happens:** Windows Firewall blocks inbound ICMP port unreachable messages by default.
**How to avoid:** Report UDP scan results as best-effort. Emit a note in the output: "UDP scan results may be incomplete (Windows Firewall may suppress ICMP port unreachable responses)". Do not fail the technique if UDP yields zero results.
**Warning signs:** UDP scan returns empty even on reachable hosts with known open UDP ports.

### Pitfall 3: detectLocalSubnet() Returns "unknown" on Some Configs
**What goes wrong:** If the host has no non-loopback IPv4 interface, `detectLocalSubnet()` returns the string `"unknown"`. Passing `"unknown"` to `net.ParseCIDR()` returns an error.
**Why it happens:** Documented in engine.go — the function returns `"unknown"` on error.
**How to avoid:** Check the return value before proceeding. If `detectLocalSubnet()` returns `"unknown"`, return `NativeResult{Success: false, ErrorOutput: "cannot detect local subnet"}`.
**Warning signs:** `net.ParseCIDR()` fails with "invalid CIDR address: unknown".

### Pitfall 4: ARP Table Output Format Varies by Windows Version
**What goes wrong:** `arp -a` output has slightly different column spacing across Windows versions.
**Why it happens:** Microsoft changed spacing but not format between versions.
**How to avoid:** Use `strings.Fields(line)` (splits on any whitespace) rather than fixed-column parsing. The IP is always field[0] and MAC is field[1] on data lines. Skip lines where field[0] doesn't parse as an IP.
**Warning signs:** Test on Windows 10 vs Windows Server — some lines have 3 fields, some 4 (type column may or may not appear).

### Pitfall 5: nltest.exe Location Not in PATH on All Machines
**What goes wrong:** `exec.Command("nltest.exe", ...)` fails with "executable file not found in $PATH".
**Why it happens:** `nltest.exe` is in `%WINDIR%\System32` which is normally in PATH, but may not be on some hardened or minimal Windows Server installations.
**How to avoid:** Try `nltest.exe` first; if exec fails (not just non-zero exit), try the full path `C:\Windows\System32\nltest.exe`. If both fail, log "nltest not available" and continue.
**Warning signs:** `exec.LookPath("nltest.exe")` returns error.

### Pitfall 6: Scan Takes Too Long on Sparse /24 Networks
**What goes wrong:** With 254 hosts × 17 ports × 300ms timeout, a /24 with all hosts down takes up to (254 × 17 / 15 workers) × 300ms ≈ 86 seconds.
**Why it happens:** All connections time out at 300ms but the queue is long.
**How to avoid:** This is acceptable for a security test tool (the user confirmed the scan). Document the expected duration in the output: emit a "Scanning 192.168.1.0/24, estimated 60-90 seconds..." line before starting.
**Warning signs:** Engine hangs — if the scan is taking longer than 5 minutes, something is wrong with the goroutine pool.

---

## Code Examples

Verified patterns from existing codebase:

### Registration Pattern (from t1482_ldap.go, t1057_wmi.go)
```go
//go:build windows

package native

func runT1046() (NativeResult, error) {
    // ... implementation ...
}

func init() {
    Register("T1046", runT1046, nil)
}
```

### Non-Windows Stub Pattern (from Phase 15 conventions)
```go
//go:build !windows

package native

import "fmt"

func init() {
    Register("T1046", func() (NativeResult, error) {
        return NativeResult{
            Success:     false,
            ErrorOutput: "T1046 network scan requires Windows",
        }, fmt.Errorf("T1046: not supported on this platform")
    }, nil)
}
```

### Subnet Parsing (Go stdlib)
```go
// net.ParseCIDR returns the network; use ip.Mask to get the base
_, ipnet, err := net.ParseCIDR("192.168.1.0/24")
// Iterate:
for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
    hosts = append(hosts, ip.String())
}
// Remove network address (last octet 0) and broadcast (last octet 255)
```

### detectLocalSubnet Reuse
```go
// detectLocalSubnet is defined in internal/engine/engine.go (package engine).
// The T1046/T1018 native functions live in internal/native/ (different package).
// Solution: duplicate the subnet detection logic in the native package OR
// extract detectLocalSubnet to an internal/netutil package.
// Recommendation: duplicate the 20-line function — avoids import cycle between
// internal/engine and internal/native. The function is simple and stable.
```

**IMPORTANT:** `detectLocalSubnet()` is defined in `internal/engine/engine.go` (package `engine`). The native techniques are in `internal/native/` (package `native`). These packages must not create an import cycle. The recommended approach is to **duplicate the subnet detection logic** directly in the scan implementation file. The function is 20 lines and stable — duplication is preferable to introducing an import cycle or a new shared package.

---

## Runtime State Inventory

Step 2.5: SKIPPED — this is a greenfield additive phase (new files, YAML rewrite). No rename or migration involved. No stored data, live service config, or OS-registered state contains "T1046" or "T1018" as a key.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `nltest.exe` | T1018 DC discovery | Expected ✓ | Windows built-in (`System32`) | Log "nltest not available", continue |
| `arp.exe` | T1018 ARP table | Expected ✓ | Windows built-in (`System32`) | Log "arp not available", continue |
| Go 1.26 (module) | Build | ✓ | 1.26.1 (go.mod) | — |
| Raw socket (ICMP) | T1018 ICMP sweep | Admin-only | N/A — OS feature | TCP 445/135 fallback (D-09) |

**Missing dependencies with no fallback:** None.

**Missing dependencies with fallback:** Raw ICMP socket (requires admin) — fallback is TCP connect alive-check per D-09.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (`testing` package) |
| Config file | None — `go test ./...` |
| Quick run command | `go test ./internal/native/... -v -run TestT104` |
| Full suite command | `go test ./internal/native/... -v` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SCAN-01 | T1046 scans /24 subnet via TCP connect (not loopback only) | unit | `go test ./internal/native/... -run TestT1046` | No — Wave 0 |
| SCAN-01 | Goroutine pool limits concurrency | unit | `go test ./internal/native/... -run TestT1046Pool` | No — Wave 0 |
| SCAN-01 | Excludes host's own IP from results | unit | `go test ./internal/native/... -run TestT1046ExcludesOwnIP` | No — Wave 0 |
| SCAN-02 | T1018 ARP table parsing returns valid entries | unit | `go test ./internal/native/... -run TestT1018ARP` | No — Wave 0 |
| SCAN-02 | T1018 nltest graceful fallback when not domain-joined | unit | `go test ./internal/native/... -run TestT1018NltestFallback` | No — Wave 0 |
| SCAN-02 | T1018 DNS reverse lookup for responding hosts | unit | `go test ./internal/native/... -run TestT1018DNS` | No — Wave 0 |
| SCAN-03 | T1046 registered as `type: go` in YAML | unit | `go test ./internal/playbooks/... -run TestT1046YAML` | No — Wave 0 |
| SCAN-03 | T1018 registered as `type: go` in YAML | unit | `go test ./internal/playbooks/... -run TestT1018YAML` | No — Wave 0 |

**Note on test scope:** Tests for SCAN-01 TCP scanning against a real subnet are integration-level and would make real network connections. They should be gated with `testing.Short()` skip or a separate build tag. The unit tests focus on: subnet IP enumeration logic, own-IP exclusion, ARP output parsing, nltest error handling, and YAML field correctness.

**Note on ICMP tests:** Raw socket ICMP tests require admin privileges on the CI machine. Tests for the ICMP path should verify the fallback behavior (non-admin path) which is always testable without elevation.

### Sampling Rate
- **Per task commit:** `go test ./internal/native/... -run TestT104 -v`
- **Per wave merge:** `go test ./internal/native/... ./internal/playbooks/... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/native/t1046_scan_test.go` — covers SCAN-01 (subnet enumeration, own-IP exclusion, pool behavior)
- [ ] `internal/native/t1018_discovery_test.go` — covers SCAN-02 (ARP parsing, nltest fallback, DNS lookup)
- [ ] `internal/playbooks/loader_t1046_test.go` or equivalent YAML field test — covers SCAN-03 (`type: go`, `tier: 1`, `requires_confirmation: true` for both techniques)

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| T1046 scans only loopback/gateway (Tier 3 stub) | T1046 scans full /24 via TCP connect goroutine pool (Tier 1) | Phase 17 | SIEM receives realistic EID 3 burst from production PID |
| T1018 did not exist | T1018 chains ICMP/ARP/nltest/DNS discovery | Phase 17 | New technique covers most common real attacker discovery tradecraft |
| PowerShell runspaces for concurrency | Native Go goroutine pool | Phase 15 + 17 | No powershell.exe process — Sysmon EID 3 events attributed to lognojutsu.exe PID |

**Deprecated/outdated:**
- T1046 `type: powershell` executor: replaced by `type: go`. The PowerShell runspace code in the YAML is fully replaced.

---

## Open Questions

1. **Import cycle risk between `internal/engine` and `internal/native`**
   - What we know: `detectLocalSubnet()` is in package `engine`; the native techniques need it
   - What's unclear: Whether there is already an import of `internal/native` in `internal/engine` (there likely is, via executor → native)
   - Recommendation: Duplicate the 20-line `detectLocalSubnet` function in `t1046_scan.go` as a private `localSubnet()` helper. Avoids any import cycle risk.

2. **UDP scan reliability on Windows**
   - What we know: UDP connect-and-send on Windows doesn't generate ICMP port unreachable reliably due to Windows Firewall
   - What's unclear: Whether the D-05 UDP scan requirement is intended to generate Sysmon events or just try the ports
   - Recommendation: Implement UDP dial with a 300ms read timeout, record attempted ports in output regardless of response. Add a note that UDP results are best-effort. This satisfies the "protocol variety" intent without over-promising.

---

## Sources

### Primary (HIGH confidence)
- Go `net` package stdlib — `net.DialTimeout`, `net.ListenPacket`, `net.ParseCIDR`, `net.LookupAddr` — verified against go1.26 documentation
- `internal/native/t1482_ldap.go` — build tag pattern, init() registration, graceful fallback, error wrapping
- `internal/native/t1057_wmi.go` — registration pattern, output formatting
- `internal/engine/engine.go:detectLocalSubnet()` — function signature and return value ("unknown" on error)
- `internal/playbooks/types.go` — `RequiresConfirmation bool` field confirmed present
- `internal/executor/executor.go` — `type: go` dispatch confirmed wired
- `go.mod` — confirmed no new dependencies needed

### Secondary (MEDIUM confidence)
- Windows raw socket ICMP privilege requirement — verified by multiple Windows networking references; consistent with behavior described in D-09
- `arp -a` output format stability — verified by examining Windows documentation; `strings.Fields` approach handles spacing variance

### Tertiary (LOW confidence)
- UDP ICMP port unreachable suppression behavior — based on known Windows Firewall defaults; actual behavior may vary by client machine configuration

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — stdlib only, no new dependencies, all verified against existing code
- Architecture: HIGH — patterns directly mirror existing t1482/t1057 implementations in the same package
- Pitfalls: HIGH for Pitfalls 1/2/3/5; MEDIUM for 4/6 (based on Windows behavior knowledge)
- Test map: HIGH — follows existing native test patterns exactly

**Research date:** 2026-04-10
**Valid until:** 2026-05-10 (stdlib-only, highly stable)
