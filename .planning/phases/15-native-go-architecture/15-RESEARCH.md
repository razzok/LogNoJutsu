# Phase 15: Native Go Architecture - Research

**Researched:** 2026-04-09
**Domain:** Go executor extension, init-time registration pattern, go-ldap/v3, go-wmi
**Confidence:** HIGH

## Summary

Phase 15 adds a `type: go` executor to the existing dispatch switch in `executor.go`, creates a new `internal/native/` package with an init-time registration registry, and delivers two real techniques (T1482 via LDAP, T1057 via WMI). The architecture closely mirrors the existing playbooks registry — a map populated via `init()` side effects, looked up at dispatch time.

Both third-party libraries are already integrated into `go.mod` and `go.sum` via `go get` during research (go-ldap/v3 v3.4.13, yusufpapurcu/wmi v1.2.4). The existing `go build ./...` and `go test ./...` pass cleanly after these additions. CGO is currently disabled (`CGO_ENABLED=0`) but go-ole uses Windows shared-library syscalls rather than CGO — WMI tests compile and run without enabling CGO. The -race detector remains unsupported on this machine (known debt from Phase 7).

**Primary recommendation:** Build `internal/native/` with a sync.RWMutex-protected map, register techniques in `init()`, add a `"go"` branch to `runInternal()` before the existing shell dispatch, and map `NativeResult` fields directly to `ExecutionResult`.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Go Technique Registry
- **D-01:** Init-time registration pattern — each Go technique file calls `native.Register()` in its `init()`. Adding a technique = adding a `.go` file. Mirrors how `playbooks.Registry` auto-populates.
- **D-02:** `NativeFunc` signature returns a structured `NativeResult` (Output, ErrorOutput, Success) plus an `error`. Error = infrastructure failure (can't connect); `!Success` = technique ran but didn't find expected data.
- **D-03:** Go cleanup functions registered alongside the technique via `Register(id, fn, cleanup)`. Cleanup can be nil for read-only techniques. Executor's defer pattern calls the Go cleanup — no shell needed.
- **D-04:** First real Go technique is T1482 (Domain Trust Discovery via LDAP). Validates full pipeline: YAML `type: go` → executor dispatch → native registry → LDAP query → result. Satisfies success criteria 2 and 3.

#### LDAP Integration
- **D-05:** DC auto-discovery: check `LOGONSERVER` env var first, then DNS SRV lookup (`_ldap._tcp.dc._msdcs.DOMAIN`). Graceful fallback with `ErrNoDCReachable` when no DC found.
- **D-06:** Authentication via current user context — NTLM bind with empty credentials (`go-ldap NTLMBind("", "")`). No credential passing needed.
- **D-07:** T1482 queries `trustedDomain` objects from `CN=System,DC=...` only — returns trust name, direction, type. Does not enumerate users/groups.

#### WMI Integration
- **D-08:** Use go-wmi high-level API (`github.com/yusufpapurcu/wmi`) — struct-based result binding. Pure Go via go-ole shared-library syscalls, no CGO.
- **D-09:** Initial WMI technique targets `Win32_Process` (T1057). Returns PID, name, command line, parent PID.

#### Executor Dispatch
- **D-10:** New `go` case in `runInternal()` — calls `native.Lookup(t.ID)`, runs it directly, maps `NativeResult` to `ExecutionResult`.
- **D-11:** No RunAs support for Go techniques — always run as current process user. Log note emitted when RunAs is configured but technique is `type: go`.
- **D-12:** Cleanup for `type: go` in `RunWithCleanup()` checks `native.LookupCleanup(t.ID)` first. If a Go cleanup function is registered, defer calls it. YAML `cleanup` field is not used for Go techniques.

### Claude's Discretion
- Package layout within `internal/native/` (flat vs `techniques/` subdirectory)
- Error message wording for graceful fallback scenarios
- Timeout values for LDAP and WMI connections
- Output formatting of technique results (tables, key-value, etc.)
- Whether to add a `context.Context` with timeout to native function calls

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ARCH-01 | Native Go executor (type: go) added to executor with internal/native/ registry for Go-implemented techniques | init-time registration pattern with sync.RWMutex map; `"go"` branch in `runInternal()` switch |
| ARCH-02 | LDAP enumeration implemented via go-ldap/v3 with graceful fallback when no DC is reachable | go-ldap/v3 v3.4.13 verified API; DialURL timeout; ErrNoDCReachable sentinel error; LOGONSERVER + DNS SRV discovery chain |
| ARCH-03 | WMI queries implemented via go-ole/wmi (pure Go, no CGO) for native technique execution | yusufpapurcu/wmi v1.2.4 verified; depends on go-ole v1.2.6 (shared-lib syscalls, no CGO); Win32_Process struct binding confirmed |
</phase_requirements>

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/go-ldap/ldap/v3` | v3.4.13 | LDAP v3 client — connect, bind, search | Official go-ldap project; NTLM bind built-in; verified published Mar 1, 2026 |
| `github.com/yusufpapurcu/wmi` | v1.2.4 | WMI WQL query via struct binding | Active fork of StackExchange/wmi; struct-based API; published Jan 28, 2024 |
| `github.com/go-ole/go-ole` | v1.2.6 | Windows COM/OLE bindings (indirect dep) | Required by yusufpapurcu/wmi; uses Windows shared-library syscalls — no CGO |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/Azure/go-ntlmssp` | v0.1.0 | NTLM SSP auth (indirect dep) | Pulled by go-ldap for NTLM bind; no direct use needed |
| `golang.org/x/crypto` | v0.48.0 | Crypto primitives (indirect dep) | Pulled by go-ldap |
| `net` (stdlib) | — | DNS SRV lookup for DC discovery | `net.LookupSRV("ldap", "tcp", domain)` |
| `sync` (stdlib) | — | Registry concurrency | RWMutex on the native registry map |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `yusufpapurcu/wmi` | `bi-zone/wmi` or `drtimf/wmi` | yusufpapurcu is the recommended fork of StackExchange; wider adoption |
| DNS SRV fallback | hardcoded LOGONSERVER only | DNS SRV is the real attacker discovery path; more realistic for T1482 |
| Flat `internal/native/` | `internal/native/techniques/` subdirectory | Either works; flat is simpler for small technique count in this phase |

**Installation (already added to go.mod during research):**
```bash
go get github.com/go-ldap/ldap/v3@v3.4.13
go get github.com/yusufpapurcu/wmi@v1.2.4
```

**Current go.mod state:** Both libraries are in `go.mod` and `go.sum`. `go build ./...` and `go test ./...` pass cleanly.

---

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── native/
│   ├── registry.go        # Register(), Lookup(), LookupCleanup(), NativeFunc type, NativeResult type
│   ├── t1482_ldap.go      # T1482 Domain Trust Discovery via LDAP — init() registers
│   └── t1057_wmi.go       # T1057 Process Discovery via WMI — init() registers
internal/
├── executor/
│   └── executor.go        # Add "go" case in runInternal(); cleanup in RunWithCleanup()
internal/
├── playbooks/
│   └── embedded/techniques/
│       ├── T1482_domain_trust_discovery.yaml   # executor.type: go, executor.command: ""
│       └── T1057_process_discovery.yaml        # executor.type: go, executor.command: ""
```

### Pattern 1: Init-Time Registration (mirrors database/sql driver pattern)
**What:** Each technique file registers itself in `init()` by calling `native.Register()`. The registry is a package-level map. Callers import the package with a blank import to trigger side effects.
**When to use:** Any new Go technique file. Adding a technique = adding one `.go` file with an `init()`.

```go
// internal/native/registry.go
// Source: decision D-01; mirrors database/sql driver pattern

package native

import "sync"

// NativeResult is the structured return from a Go technique.
type NativeResult struct {
    Output      string
    ErrorOutput string
    Success     bool
}

// NativeFunc is the signature every Go technique must implement.
type NativeFunc func() (NativeResult, error)

// CleanupFunc is the optional cleanup a Go technique can register.
type CleanupFunc func() error

type entry struct {
    fn      NativeFunc
    cleanup CleanupFunc
}

var (
    mu       sync.RWMutex
    registry = make(map[string]entry)
)

// Register adds a technique to the native registry.
// cleanup may be nil for read-only techniques.
func Register(id string, fn NativeFunc, cleanup CleanupFunc) {
    mu.Lock()
    defer mu.Unlock()
    registry[id] = entry{fn: fn, cleanup: cleanup}
}

// Lookup returns the NativeFunc for a technique ID, or nil if not found.
func Lookup(id string) NativeFunc {
    mu.RLock()
    defer mu.RUnlock()
    e, ok := registry[id]
    if !ok {
        return nil
    }
    return e.fn
}

// LookupCleanup returns the CleanupFunc for a technique ID, or nil.
func LookupCleanup(id string) CleanupFunc {
    mu.RLock()
    defer mu.RUnlock()
    e, ok := registry[id]
    if !ok {
        return nil
    }
    return e.cleanup
}
```

### Pattern 2: Executor Dispatch for `type: go`
**What:** Add a `"go"` branch in `runInternal()` before the existing shell dispatch. Map `NativeResult` fields to `ExecutionResult`.
**When to use:** Only in `runInternal()`. Keep the existing shell path untouched.

```go
// internal/executor/executor.go — addition to runInternal()
// Source: decision D-10, D-11

import "lognojutsu/internal/native"

// In runInternal(), after logging TechStart/TechCommand:
if strings.ToLower(t.Executor.Type) == "go" {
    if profile != nil && profile.UserType != userstore.UserTypeCurrent {
        simlog.Info(fmt.Sprintf("[%s] RunAs not supported for type:go — executing as current user", t.ID))
    }
    fn := native.Lookup(t.ID)
    if fn == nil {
        result.ErrorOutput = fmt.Sprintf("no native function registered for technique %s", t.ID)
        result.Success = false
    } else {
        nr, nErr := fn()
        result.Output = nr.Output
        result.ErrorOutput = nr.ErrorOutput
        result.Success = nr.Success
        if nErr != nil {
            result.ErrorOutput = nErr.Error() + "\n" + result.ErrorOutput
            result.Success = false
        }
    }
    result.EndTime = time.Now().Format(time.RFC3339)
    simlog.TechOutput(t.ID, result.Output, result.ErrorOutput)
    simlog.TechEnd(t.ID, result.Success, time.Since(start))
    return result
}
```

### Pattern 3: RunWithCleanup for `type: go`
**What:** Check `native.LookupCleanup(t.ID)` before existing shell cleanup. If Go cleanup is registered, defer it instead.
**When to use:** Only in `RunWithCleanup()`. YAML `cleanup` field is ignored for `type: go`.

```go
// internal/executor/executor.go — RunWithCleanup addition
// Source: decision D-12

func RunWithCleanup(t *playbooks.Technique, profile *userstore.UserProfile, password string) (result playbooks.ExecutionResult) {
    if strings.ToLower(t.Executor.Type) == "go" {
        // Use native cleanup if registered; YAML cleanup field is ignored for type:go
        if cleanFn := native.LookupCleanup(t.ID); cleanFn != nil {
            defer func() {
                simlog.TechCleanup(t.ID, "(go native cleanup)", false)
                cleanErr := cleanFn()
                result.CleanupRun = true
                simlog.TechCleanup(t.ID, "(go cleanup completed)", cleanErr == nil)
            }()
        }
    } else if strings.TrimSpace(t.Cleanup) != "" {
        // Existing shell cleanup path — unchanged
        defer func() { ... }()
    }
    result = runInternal(t, profile, password)
    return result
}
```

### Pattern 4: LDAP DC Discovery Chain
**What:** Check LOGONSERVER env var → DNS SRV lookup → return ErrNoDCReachable sentinel.
**When to use:** T1482 init before LDAP dial. Any future technique needing a DC.

```go
// internal/native/t1482_ldap.go — DC discovery
// Source: decision D-05; net.LookupSRV standard library

import (
    "errors"
    "fmt"
    "net"
    "os"
    "strings"
    "time"

    "github.com/go-ldap/ldap/v3"
)

var ErrNoDCReachable = errors.New("no domain controller reachable")

func discoverDC() (string, error) {
    // 1. LOGONSERVER env var (fast path on domain-joined machine)
    if ls := os.Getenv("LOGONSERVER"); ls != "" {
        dc := strings.TrimPrefix(ls, `\\`)
        return dc + ":389", nil
    }

    // 2. DNS SRV lookup: _ldap._tcp.dc._msdcs.<domain>
    domain := os.Getenv("USERDNSDOMAIN")
    if domain != "" {
        _, addrs, err := net.LookupSRV("ldap", "tcp", "dc._msdcs."+domain)
        if err == nil && len(addrs) > 0 {
            return fmt.Sprintf("%s:%d", strings.TrimSuffix(addrs[0].Target, "."), addrs[0].Port), nil
        }
    }

    return "", ErrNoDCReachable
}
```

### Pattern 5: WMI Win32_Process Query
**What:** Define a struct matching WMI field names. Use `wmi.Query()` for struct-bound results.
**When to use:** T1057 and any future WMI-based technique.

```go
// internal/native/t1057_wmi.go
// Source: yusufpapurcu/wmi v1.2.4 docs (pkg.go.dev)

import "github.com/yusufpapurcu/wmi"

type win32Process struct {
    Name            string
    ProcessId       uint32
    ParentProcessId uint32
    CommandLine     string
}

func queryProcesses() ([]win32Process, error) {
    var procs []win32Process
    q := wmi.CreateQuery(&procs, "")
    return procs, wmi.Query(q, &procs)
}
```

### Pattern 6: YAML Update for `type: go`
**What:** Set `executor.type: go` and `executor.command: ""` in technique YAML. The command field is ignored by the Go dispatch path.

```yaml
# T1482_domain_trust_discovery.yaml (updated)
executor:
  type: go
  command: ""
```

### Anti-Patterns to Avoid
- **Shell command in a `type: go` technique:** If executor.type is `go`, the YAML command field must be empty. Planner must not leave PowerShell content there.
- **Importing `internal/native` in techniques:** Technique files should only call `native.Register()` in `init()` — they must NOT import other technique files.
- **Calling `native.Lookup()` at package init time:** Registry is only safe to read after all `init()` calls complete. Lookup must only happen at dispatch time.
- **Blank import in executor.go:** `internal/native` package must be imported with a regular named import (`"lognojutsu/internal/native"`), NOT a blank import. The registry methods are called directly — only the technique files need no explicit usage of their own package.
- **Passing NativeResult.ErrorOutput as an error:** `NativeResult.ErrorOutput` is a string (like stderr). Infrastructure failures use the returned `error`. Do not conflate them.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| LDAP connection & bind | Custom TCP + ASN.1 parser | `go-ldap/v3 DialURL + NTLMBind` | LDAP ASN.1/BER encoding is complex; NTLM handshake has many edge cases |
| WMI COM dispatch | Manual syscall to `WbemLocator` | `yusufpapurcu/wmi Query()` | COM dispatch table, BSTR marshaling, and variant type handling are error-prone |
| NTLM authentication | SSPI syscalls | Built into go-ldap via `go-ntlmssp` | NTLM challenge-response has complex crypto; already pulled as indirect dep |
| DNS SRV parsing | String parsing of nslookup output | `net.LookupSRV` (stdlib) | Handles priority/weight sorting per RFC 2782 automatically |

**Key insight:** The LDAP and WMI stacks have significant Windows-specific plumbing (BER encoding, COM dispatch, NTLM challenge handshake). The chosen libraries encapsulate all of it behind clean Go APIs.

---

## Common Pitfalls

### Pitfall 1: Blank Import Causes "imported and not used" on technique files
**What goes wrong:** Technique files in `internal/native/` have no exported symbols used by the executor. If the executor imports them with `_`, the Go compiler is fine. But the registry itself needs a named import from executor.
**Why it happens:** The technique files self-register via `init()` — they only need to be in the same package (`package native`) to have their `init()` called when the package is first used.
**How to avoid:** All technique files are in `package native`. The executor does `import "lognojutsu/internal/native"` — this is enough to trigger all `init()` calls in the package. No blank import needed.
**Warning signs:** Compiler error "imported and not used" if you try blank-importing native from executor.

### Pitfall 2: CGO_ENABLED=0 + go-ole
**What goes wrong:** `go-ole` uses `syscall.LoadDLL` — pure Windows DLL loading, NOT CGO. If a developer sets `CGO_ENABLED=0` and tries to build, WMI still works because go-ole is not CGO. However, some documentation implies CGO involvement.
**Why it happens:** go-ole README says "shared libraries instead of cgo" but this is non-obvious.
**How to avoid:** Document this explicitly. Current machine has `CGO_ENABLED=0` and the build passes with go-ole. The `-race` flag requires CGO (known debt from Phase 7) — WMI tests must not use `-race`.
**Warning signs:** If WMI fails with "The procedure entry point..." error, it's a Windows DLL issue, not a CGO issue.

### Pitfall 3: WMI `CommandLine` Can Be Empty
**What goes wrong:** `Win32_Process.CommandLine` is empty for system processes (kernel, SYSTEM processes). A test asserting `len(CommandLine) > 0` for all results will fail.
**Why it happens:** The OS doesn't expose command lines for some privileged processes.
**How to avoid:** Filter with `WHERE CommandLine IS NOT NULL` in WQL, or test only that at least one non-empty CommandLine exists in results. Use `wmi.CreateQuery(&dst, "WHERE CommandLine IS NOT NULL")` or accept empty strings.
**Warning signs:** Test passes on developer machine but fails on minimal Windows Server with minimal processes.

### Pitfall 4: `NTLMBind("", "")` Requires Domain-Joined Machine
**What goes wrong:** `NTLMBind` with empty credentials sends the current user's Kerberos/NTLM token. On a non-domain-joined machine, this returns `LDAP_INVALID_CREDENTIALS` (code 49).
**Why it happens:** The decision (D-06) correctly calls for empty credentials, which is appropriate for domain-joined machines only.
**How to avoid:** T1482 graceful fallback must handle LDAP error code 49 as a "no domain available" condition — not a crash. The `discoverDC()` function returning `ErrNoDCReachable` handles the pre-connect case; an additional check after bind handles the post-connect invalid-credentials case.
**Warning signs:** Error message "LDAP Result Code 49 "Invalid Credentials"" on developer machine not joined to domain.

### Pitfall 5: DialURL Timeout Must Be Explicit
**What goes wrong:** `ldap.DialURL("ldap://dc:389")` with no timeout will hang if the DC is unreachable (firewall drop vs TCP RST). This causes the technique to block indefinitely.
**Why it happens:** The default `net.Dialer` has no timeout.
**How to avoid:** Use `ldap.DialURL(url, ldap.DialWithDialer(&net.Dialer{Timeout: 3*time.Second}))`. The 3-second value is Claude's discretion (see constraints above).
**Warning signs:** Technique hangs, no output, no error logged.

### Pitfall 6: T1482 YAML Now Has `type: go` — Test Coverage Gap
**What goes wrong:** The existing `TestWriteArtifactsHaveCleanup` test does not include T1482 (read-only). But `TestExpectedEvents` still requires `expected_events` in the updated YAML. If the YAML update drops `expected_events`, that test fails.
**Why it happens:** The YAML rewrite focuses on the executor change and may accidentally trim other fields.
**How to avoid:** T1482 updated YAML must retain `expected_events`, `tier`, `tags`, and other metadata. Only `executor.type` and `executor.command` change.

---

## Code Examples

Verified patterns from official sources and live testing:

### go-ldap/v3: Full T1482 Connection Pattern
```go
// Source: pkg.go.dev/github.com/go-ldap/ldap/v3 (version v3.4.13, published 2026-03-01)

import (
    "fmt"
    "net"
    "time"
    "github.com/go-ldap/ldap/v3"
)

func runLDAPTrustQuery(dcAddr string) (string, error) {
    // Dial with explicit timeout to avoid hanging on unreachable DC
    conn, err := ldap.DialURL("ldap://"+dcAddr,
        ldap.DialWithDialer(&net.Dialer{Timeout: 3 * time.Second}),
    )
    if err != nil {
        return "", fmt.Errorf("ldap dial: %w", err)
    }
    defer conn.Close()

    // NTLM bind with current user context (empty credentials = use current token)
    if err := conn.NTLMBind("", "", ""); err != nil {
        return "", fmt.Errorf("ldap ntlm bind: %w", err)
    }

    // Search for trustedDomain objects
    req := ldap.NewSearchRequest(
        baseDN,                              // e.g. "CN=System,DC=corp,DC=example,DC=com"
        ldap.ScopeWholeSubtree,
        ldap.NeverDerefAliases,
        0, 10, false,
        "(objectClass=trustedDomain)",
        []string{"name", "trustType", "trustDirection"},
        nil,
    )
    sr, err := conn.Search(req)
    if err != nil {
        return "", fmt.Errorf("ldap search: %w", err)
    }
    // Build output string from sr.Entries
    var sb strings.Builder
    for _, entry := range sr.Entries {
        fmt.Fprintf(&sb, "Trust: %s (type=%s direction=%s)\n",
            entry.GetAttributeValue("name"),
            entry.GetAttributeValue("trustType"),
            entry.GetAttributeValue("trustDirection"),
        )
    }
    return sb.String(), nil
}
```

**Note on NTLMBind signature:** The actual signature is `NTLMBind(domain, username, password string) error`. For current-user token binding, pass empty strings for all three parameters.

### yusufpapurcu/wmi: Win32_Process Query
```go
// Source: pkg.go.dev/github.com/yusufpapurcu/wmi@v1.2.4

import "github.com/yusufpapurcu/wmi"

type win32Process struct {
    Name            string
    ProcessId       uint32
    ParentProcessId uint32
    CommandLine     string
}

func queryProcesses() ([]win32Process, error) {
    var procs []win32Process
    // CreateQuery derives "SELECT * FROM Win32_Process WHERE CommandLine IS NOT NULL"
    q := wmi.CreateQuery(&procs, "WHERE CommandLine IS NOT NULL")
    if err := wmi.Query(q, &procs); err != nil {
        return nil, fmt.Errorf("wmi Win32_Process query: %w", err)
    }
    return procs, nil
}
```

### Executor Dispatch — Minimal `runInternal` Change
```go
// Insertion point: internal/executor/executor.go, in runInternal()
// Add this block BEFORE the existing profile/shell dispatch

switch strings.ToLower(t.Executor.Type) {
case "go":
    if profile != nil && profile.UserType != userstore.UserTypeCurrent {
        simlog.Info(fmt.Sprintf("[%s] type:go does not support RunAs — executing as current user", t.ID))
    }
    fn := native.Lookup(t.ID)
    if fn == nil {
        result.ErrorOutput = fmt.Sprintf("no native Go function registered for %s", t.ID)
        result.Success = false
    } else {
        nr, nErr := fn()
        result.Output, result.ErrorOutput, result.Success = nr.Output, nr.ErrorOutput, nr.Success
        if nErr != nil {
            result.ErrorOutput = nErr.Error() + "\n" + result.ErrorOutput
            result.Success = false
        }
    }
    result.EndTime = time.Now().Format(time.RFC3339)
    simlog.TechOutput(t.ID, result.Output, result.ErrorOutput)
    simlog.TechEnd(t.ID, result.Success, time.Since(start))
    return result
}
// ... existing profile/shell dispatch continues below
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| PowerShell stub for T1482 | Native Go LDAP via go-ldap/v3 | Phase 15 | Real EID 4662 on DC; Tier 3 → Tier 1 |
| PowerShell stub for T1057 | Native Go WMI via yusufpapurcu/wmi | Phase 15 | Real process data; Tier 3 → Tier 1 |
| All shell executor types | Shell + new `type: go` | Phase 15 | Enables zero-subprocess technique execution |

**Deprecated/outdated:**
- `github.com/StackExchange/wmi`: Original WMI package, now recommends users move to `yusufpapurcu/wmi`. Do not use.
- `gopkg.in/ldap.v3`: Legacy import path for go-ldap. Use `github.com/go-ldap/ldap/v3` instead.

---

## Open Questions

1. **Base DN construction for T1482 LDAP search**
   - What we know: The base DN for trustedDomain objects is `CN=System,DC=<components>`. The DC components come from `USERDNSDOMAIN` env var (e.g., `corp.example.com` → `DC=corp,DC=example,DC=com`).
   - What's unclear: Whether to construct base DN from `USERDNSDOMAIN` string splitting or query RootDSE first.
   - Recommendation: Query RootDSE's `defaultNamingContext` attribute after connecting — the authoritative source. Fallback: split `USERDNSDOMAIN`. Both are implementable; RootDSE is more reliable.

2. **T1057 WMI test on machines without WMI access**
   - What we know: `yusufpapurcu/wmi` only works on Windows. The test machine is Windows/amd64.
   - What's unclear: Whether CI or a future Linux build would need WMI tests to compile-but-skip.
   - Recommendation: Use `//go:build windows` build tag in `t1057_wmi.go` and its test file. The test can check `runtime.GOOS == "windows"` and `t.Skip()` otherwise.

3. **`NTLMBind` parameter order on go-ldap v3.4.13**
   - What we know: Documented signature is `NTLMBind(domain, username, password string) error`.
   - What's unclear: Whether passing all-empty strings uses current user's cached Kerberos token or fails.
   - Recommendation: If `NTLMBind("", "", "")` fails on domain-joined machine, try `NTLMUnauthenticatedBind("", "")` which is explicitly designed for anonymous/current-context bind. Both are available in v3.4.13.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All compilation | ✓ | go1.26.1 windows/amd64 | — |
| `github.com/go-ldap/ldap/v3` | ARCH-02 / T1482 | ✓ | v3.4.13 (in go.mod) | — |
| `github.com/yusufpapurcu/wmi` | ARCH-03 / T1057 | ✓ | v1.2.4 (in go.mod) | — |
| `github.com/go-ole/go-ole` | yusufpapurcu/wmi dep | ✓ | v1.2.6 (indirect in go.mod) | — |
| Active Directory / Domain Controller | T1482 LDAP | ✗ (dev machine) | — | Graceful fallback via ErrNoDCReachable (required by ARCH-02) |
| CGO toolchain (gcc) | -race flag | ✗ | — | Omit -race (known debt, Phase 7) |

**Missing dependencies with no fallback:**
- None blocking compilation or test execution.

**Missing dependencies with fallback:**
- Domain Controller: T1482 must log a graceful fallback message when `discoverDC()` returns `ErrNoDCReachable`. This is explicitly required by ARCH-02 and success criterion 3.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none (go test ./... from repo root) |
| Quick run command | `go test ./internal/native/... ./internal/executor/... ./internal/playbooks/...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| ARCH-01 | Registry registers and looks up a Go function by technique ID | unit | `go test ./internal/native/... -run TestRegisterLookup` | ❌ Wave 0 |
| ARCH-01 | `type: go` technique executes without spawning a child process | unit | `go test ./internal/executor/... -run TestGoDispatch` | ❌ Wave 0 |
| ARCH-01 | `type: go` unregistered technique returns error result (not panic) | unit | `go test ./internal/executor/... -run TestGoDispatchUnregistered` | ❌ Wave 0 |
| ARCH-02 | T1482 Go function returns NativeResult with graceful fallback when no DC | unit | `go test ./internal/native/... -run TestT1482NoDC` | ❌ Wave 0 |
| ARCH-02 | T1482 YAML loads with executor.type=go | unit | `go test ./internal/playbooks/... -run TestT1482ExecutorType` | ❌ Wave 0 |
| ARCH-03 | T1057 WMI query returns at least one Win32_Process result | unit | `go test ./internal/native/... -run TestT1057WMI` | ❌ Wave 0 |
| ARCH-03 | T1057 YAML loads with executor.type=go | unit | `go test ./internal/playbooks/... -run TestT1057ExecutorType` | ❌ Wave 0 |
| All | Existing test suite still passes after changes | regression | `go test ./...` | ✅ |

### Sampling Rate
- **Per task commit:** `go test ./internal/native/... ./internal/executor/... ./internal/playbooks/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/native/registry_test.go` — covers ARCH-01 register/lookup/cleanup
- [ ] `internal/executor/executor_go_test.go` — covers ARCH-01 dispatch, unregistered case, RunAs log note
- [ ] `internal/native/t1482_test.go` — covers ARCH-02 no-DC fallback
- [ ] `internal/native/t1057_test.go` — covers ARCH-03 Win32_Process result (Windows build tag)
- [ ] `internal/playbooks/loader_test.go` additions — verify T1482 and T1057 now have `executor.type == "go"`

---

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/github.com/go-ldap/ldap/v3` — DialURL, NTLMBind, NewSearchRequest, Error types; version v3.4.13 published 2026-03-01
- `pkg.go.dev/github.com/yusufpapurcu/wmi@v1.2.4` — Query, CreateQuery, Win32_Process struct binding; version v1.2.4 published 2024-01-28
- Live `go get` execution — both packages resolved into `go.mod`; `go build ./...` and `go test ./...` confirmed passing
- `go.mod` (project) — current dependency state with all transitive dependencies verified
- Go stdlib `net.LookupSRV` — standard DC discovery mechanism documented at go.dev/src/net/lookup.go

### Secondary (MEDIUM confidence)
- `github.com/go-ole/go-ole` README — "shared libraries instead of cgo" claim; v1.3.0 latest release (2023-08-15); indirect dep of wmi
- `pkg.go.dev/github.com/go-ldap/ldap/v3` — NTLMBind signature and NTLMUnauthenticatedBind as fallback
- WebSearch + pkg.go.dev — yusufpapurcu/wmi is the recommended successor to StackExchange/wmi

### Tertiary (LOW confidence)
- WMI `CommandLine` null behavior on system processes — inferred from Windows WMI documentation patterns; not directly tested on this machine yet

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — both libraries `go get`-resolved and `go build` verified live
- Architecture: HIGH — dispatcher pattern mirrors existing code; init registration mirrors database/sql
- LDAP pitfalls: HIGH — NTLMBind signature verified from live pkg.go.dev; timeout pitfall from go-ldap docs
- WMI pitfalls: MEDIUM — CommandLine null behavior inferred from WMI conventions, not live-tested
- DC discovery: HIGH — net.LookupSRV is stdlib, LOGONSERVER is standard Windows env var

**Research date:** 2026-04-09
**Valid until:** 2026-05-09 (stable libraries; go-ldap/v3 and wmi are mature)
