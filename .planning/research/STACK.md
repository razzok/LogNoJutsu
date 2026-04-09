# Stack Research

**Domain:** SIEM validation tool — Realistic Attack Simulation (v1.3)
**Researched:** 2026-04-09
**Confidence:** HIGH (core recommendations), MEDIUM (ARP implementation approach)

## Context

This is a subsequent milestone. The core stack (Go 1.26.1, vanilla HTML/JS UI, `gopkg.in/yaml.v3`, embed FS) is
already validated and ships as a single Windows `.exe` with no installer. This document covers ONLY what is
needed for v1.3's three new capability areas:

1. **Real network discovery** — ARP cache read + ICMP ping sweep + TCP port scan on local /24 subnet
2. **Realistic LDAP queries** — actual AD enumeration (users, groups, computers) instead of PowerShell echo stubs
3. **WMI interaction** — query Win32 system information to generate authentic Sysmon/WMI event artifacts

The non-negotiable constraint throughout: **CGO_ENABLED=0 must remain possible**, or the new dependency must
be demonstrably buildable as a single Windows `.exe` without WinPcap/Npcap runtime installation requirements.

---

## Recommended Stack Additions

### Network Discovery

| Technology | Version | Purpose | CGO? | Why |
|------------|---------|---------|------|-----|
| `golang.org/x/net/icmp` | latest (x/net) | ICMP echo requests for host liveness | No | stdlib-adjacent, pure Go, already in go.sum indirectly via x/net; `SetPrivileged(true)` works on Windows 10+ without real elevation |
| `github.com/prometheus-community/pro-bing` | v0.8.0 (Feb 2026) | Higher-level ICMP ping with concurrency | No | Maintained fork of abandoned go-ping/ping; pure Go; Windows-tested; SetPrivileged(true) does NOT require admin on Windows 10/11 |
| `golang.org/x/sys/windows` (iphlpapi.dll) | latest | Read Windows ARP cache via GetIpNetTable | No | Pure syscall via `windows.NewLazySystemDLL("iphlpapi.dll")` — no WinPcap, no CGO, no extra driver needed; already used in gopsutil and other Windows tools |
| stdlib `net` package | Go 1.26.1 | TCP port scan via `net.DialTimeout` | No | Pure Go, no library needed; goroutine-per-port with semaphore channel limits concurrency |

**Do NOT use `google/gopacket`** — it requires WinPcap/Npcap installed at runtime AND CGO for the
pcap bindings. Distributing a standalone `.exe` to clients who do not have Npcap installed will fail silently
or panic. The iphlpapi.dll approach reads the system ARP cache populated by Windows automatically — no raw
packet injection needed.

### LDAP / Active Directory Enumeration

| Technology | Version | Purpose | CGO? | Why |
|------------|---------|---------|------|-----|
| `github.com/go-ldap/ldap/v3` | v3.4.13 (Mar 2026) | LDAP v3 client for AD queries | No | Pure Go (uses `go-asn1-ber` which is also pure Go); zero CGO; MIT license; 30 imports all pure Go; canonical Go LDAP library used by go-windapsearch, HashiCorp, and others |

The library supports all required operations: `DialURL` (ldap:// on port 389), anonymous and authenticated
`Bind`, `SearchWithPaging` for large result sets, and standard LDAP filters for users
(`objectClass=user`), groups (`objectClass=group`), and computers (`objectClass=computer`).
This generates authentic Event ID 4662 (AD object access) artifacts in Windows Security logs.

### WMI Interaction

| Technology | Version | Purpose | CGO? | Why |
|------------|---------|---------|------|-----|
| `github.com/go-ole/go-ole` | v1.3.0 (Aug 2023) | Windows COM/OLE bindings | No | Pure Go bindings via Windows syscalls — explicitly described as "Go bindings for Windows COM using shared libraries instead of cgo"; foundation for all WMI libraries |
| `github.com/yusufpapurcu/wmi` | v1.2.x (active fork) | WQL queries to Windows WMI | Via go-ole syscalls | Actively maintained fork of StackExchange/wmi; go-ole's syscall approach means no gcc/CGO tool chain required; WQL queries generate authentic WMI event artifacts |

**Important:** `go-ole` uses Windows `LoadLibrary`/`GetProcAddress` syscalls, NOT CGO. This means
`CGO_ENABLED=0 go build` still works. The binary remains fully standalone — no DLL installation needed
beyond what Windows ships with (ole32.dll, oleaut32.dll).

---

## Core Technologies (Unchanged)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.26.1 | All backend logic | Single binary, no runtime deps |
| `gopkg.in/yaml.v3` | v3.0.1 | Playbook YAML parsing | Already in go.mod |
| Vanilla HTML/CSS/JS | ES2020 | Web UI | No build step, zero client deps |
| `golang.org/x/sys/windows` | latest | Windows syscalls (audit policy, now ARP) | Pure Go, already used in codebase |

---

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Network scanning | stdlib `net` + `golang.org/x/net/icmp` + iphlpapi.dll | `google/gopacket` | gopacket requires WinPcap/Npcap runtime driver + CGO; breaks standalone distribution |
| ARP discovery | `iphlpapi.dll GetIpNetTable` via syscall | `mdlayher/arp` (RFC 826 implementation) | mdlayher/arp sends raw ARP packets requiring raw socket privilege; reading the OS ARP cache is sufficient for host discovery |
| ICMP ping | `prometheus-community/pro-bing` | `golang.org/x/net/icmp` directly | pro-bing handles Windows quirks (SetPrivileged, timeout goroutines, stats); x/net/icmp is lower-level and requires more boilerplate |
| LDAP | `go-ldap/ldap/v3` | `dlampsi/adc` | adc is a convenience wrapper around go-ldap; go-ldap directly gives full control over LDAP filters needed for varied attack scenarios |
| WMI | `yusufpapurcu/wmi` (go-ole syscall) | `bi-zone/wmi` | bi-zone/wmi also uses go-ole; yusufpapurcu is the official recommended fork of StackExchange/wmi with more recent updates |
| WMI | `yusufpapurcu/wmi` | `microsoft/wmi` | microsoft/wmi is a heavier abstraction; published Dec 2025 but overkill for WQL reads |
| TCP port scan | stdlib `net.DialTimeout` + goroutines | External scanner library | The goroutine/semaphore pattern needs ~20 lines of Go; no library adds value; libraries like anvie/port-scanner are thin wrappers |

---

## Installation — New go.mod Additions

```bash
go get github.com/prometheus-community/pro-bing@v0.8.0
go get github.com/go-ldap/ldap/v3@v3.4.13
go get github.com/go-ole/go-ole@v1.3.0
go get github.com/yusufpapurcu/wmi
```

After adding, verify CGO is not required:

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
```

`golang.org/x/sys/windows` will be pulled in transitively (it already may be, given Go 1.26.1 on Windows).

---

## Integration Points

| New Capability | Existing Package | Integration |
|---------------|-----------------|-------------|
| TCP port scan | `internal/executor` | New technique functions call `net.DialTimeout` with goroutines; results logged via existing `simlog` |
| ICMP host sweep | `internal/executor` | New technique calls pro-bing `NewPinger`; on Windows set `SetPrivileged(true)` — no admin needed |
| ARP cache read | `internal/executor` | New technique calls `iphlpapi.GetIpNetTable` via `windows.NewLazySystemDLL`; returns live MAC→IP table |
| LDAP enumeration | `internal/executor` | New technique calls `ldap.DialURL("ldap://"+domainController+":389")` then `SearchWithPaging`; falls back gracefully if no DC detected |
| WMI queries | `internal/executor` | New technique calls `wmi.Query("SELECT * FROM Win32_Process", ...)` — generates Event ID 4688 / Sysmon 1 artifacts |
| Discovery phase | `internal/playbooks` | New YAML playbook entries reference the new executor functions; existing MITRE mapping fields unchanged |

---

## What NOT to Add

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `google/gopacket` | Requires WinPcap/Npcap runtime + CGO; breaks single-binary distribution | stdlib `net` + `x/net/icmp` + iphlpapi.dll syscall |
| `paleg/libadclient` | CGO-based C++ wrapper; completely breaks standalone exe | `go-ldap/ldap/v3` (pure Go) |
| Any library requiring `CGO_ENABLED=1` | No gcc on target machines; `go build` must work cross-compile from dev box | Only pure-Go or Windows-syscall-only packages |
| Caldera/Atomic Red Team Go agent packages | pkg.go.dev shows `github.com/redcanaryco/atomic-red-team` as a Go module, but it is the YAML test library — not executable Go code; Caldera's Sandcat agent is useful as a reference pattern only | Implement technique execution directly in `internal/executor` following established patterns |
| `github.com/go-ping/ping` | Officially abandoned; fork is pro-bing | `github.com/prometheus-community/pro-bing` |
| `mostlygeek/arp`, `juruen/goarp` | Both abandoned (2015/2018), OSX-centric, parse `arp -a` output | Direct iphlpapi.dll syscall |

---

## Build Verification Matrix

| Scenario | Expected Result |
|----------|----------------|
| `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...` | Succeeds — all new deps are pure Go or Windows syscall |
| Deploy `.exe` to Windows 10 without Npcap installed | Works — no pcap dependency |
| Deploy `.exe` to Windows Server 2016 without AD | LDAP techniques gracefully return "no DC found"; other techniques unaffected |
| Run ICMP sweep as local admin | Works with SetPrivileged(true) |
| Run ICMP sweep as standard user | Works — SetPrivileged(true) on Windows 10/11 does NOT require elevation |

---

## Sources

- [go-ldap/ldap v3 on pkg.go.dev](https://pkg.go.dev/github.com/go-ldap/ldap/v3) — v3.4.13, Mar 1 2026, 0 CGO imports confirmed — HIGH confidence
- [go-ole/go-ole on GitHub](https://github.com/go-ole/go-ole) — "Go bindings for Windows COM using shared libraries instead of cgo" — HIGH confidence
- [prometheus-community/pro-bing on GitHub](https://github.com/prometheus-community/pro-bing) — v0.8.0 Feb 2026, pure Go, Windows-tested — HIGH confidence
- [yusufpapurcu/wmi on pkg.go.dev](https://pkg.go.dev/github.com/yusufpapurcu/wmi) — active StackExchange/wmi fork, Windows COM via go-ole — MEDIUM confidence (CGO status inferred from go-ole's pure syscall nature)
- [gopacket pcap on DeepWiki](https://deepwiki.com/google/gopacket/3.1-pcap-capture) — confirms WinPcap/Npcap requirement and CGO binding — HIGH confidence (confirms exclusion)
- [Microsoft GetIpNetTable docs](https://learn.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getipnettable) — confirms iphlpapi.dll GetIpNetTable function for ARP cache — HIGH confidence
- [gopsutil net_windows.go](https://github.com/shirou/gopsutil/blob/master/net/net_windows.go) — confirms `windows.NewLazySystemDLL("iphlpapi.dll")` pattern in production Go code — HIGH confidence
- [go-windapsearch on GitHub](https://github.com/ropnop/go-windapsearch) — confirms go-ldap as the right library for AD enumeration in Go security tools — HIGH confidence

---

*Stack research for: LogNoJutsu v1.3 — Realistic Attack Simulation*
*Researched: 2026-04-09*
