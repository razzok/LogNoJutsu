# Architecture Research

**Domain:** SIEM validation tool — realistic attack simulation integration (v1.3)
**Researched:** 2026-04-09
**Confidence:** HIGH

## Standard Architecture

### Current System Overview

```
UI (browser)
    |
    v HTTP
internal/server
    |
    v
internal/engine  <-- orchestrates phases, user rotation, PoC scheduling
    |
    +---> internal/executor  <-- TODAY: only powershell/cmd shell execution
    |         |
    |         v
    |     OS (powershell.exe / cmd.exe)
    |
    +---> internal/playbooks  <-- YAML registry, technique definitions
    +---> internal/reporter
    +---> internal/simlog
    +---> internal/verifier
```

### Target System Overview (v1.3)

```
UI (browser)
    |
    v HTTP
internal/server
    |
    v
internal/engine  <-- unchanged: orchestrates phases, calls executor dispatch
    |
    v
internal/executor  <-- EXTENDED: dispatch by executor type
    |
    +---> shell executor path (unchanged)
    |     powershell.exe / cmd.exe
    |
    +---> native executor path (NEW)
          internal/native/  <-- new package: Go technique implementations
              registry.go   <-- map[string]NativeFunc, Register()
              scanner.go    <-- TCP port scan, ICMP ping, /24 sweep
              ldap.go       <-- LDAP queries against DC
              (future: dns.go, smb.go, ...)
```

The single-binary constraint is preserved: all native Go code compiles in. No runtime processes are spawned for native techniques. Dependencies added to `go.mod` are compiled in at build time.

### Component Responsibilities

| Component | Responsibility | Change for v1.3 |
|-----------|----------------|-----------------|
| `internal/playbooks/types.go` | Technique struct, Executor type enum | Add `GoFunc string` field to `Executor` struct |
| `internal/executor/executor.go` | Shell command dispatch | Add branch for `executor.type == "go"` that calls native registry |
| `internal/native/` | Library of Go technique implementations | NEW package |
| `internal/engine/engine.go` | Phase orchestration, WhatIf bypass | Unchanged — delegates to executor already |
| YAML technique files | Declare executor type and params | New files use `type: go` + `go_func: scanner_tcp` |

## Recommended Project Structure (delta from current)

```
internal/
├── executor/
│   └── executor.go          # MODIFIED: add dispatchNative() branch
├── native/                  # NEW package
│   ├── registry.go          # NativeFunc type, Registry map, Register()
│   ├── scanner.go           # TCP port scan, ICMP host sweep
│   ├── ldap.go              # LDAP DC query techniques
│   └── native_test.go       # Unit tests for each func
internal/playbooks/
│   ├── types.go             # MODIFIED: add GoFunc to Executor struct
│   └── embedded/techniques/
│       ├── T1046_network_scan.yaml   # MODIFIED: type: go
│       ├── T1018_remote_discovery.yaml  # NEW: LDAP user/computer enum
│       └── ...              # additional new technique YAML files
go.mod                       # MODIFIED: add go-ldap/ldap/v3, golang.org/x/net
```

### Structure Rationale

- **`internal/native/` as a separate package:** Keeps `executor.go` focused on dispatch. Native implementations have no dependency on `simlog` or `userstore` — they accept `map[string]string` args and return `(string, error)`. Tests are isolated from OS shell requirements.
- **Registry map over interface:** Caldera uses a full `Executor` interface with `Run()`, `CheckIfAvailable()`, `UpdateBinary()`, `DownloadPayloadToMemory()`. For LogNoJutsu that level of abstraction is premature. A `map[string]NativeFunc` matches the existing YAML `input_args` pattern, keeps dispatch trivial, and is sufficient for the projected scale (under 20 native techniques).
- **No changes to engine.go:** The engine already calls `executor.Run()` / `executor.RunAs()`. These functions absorb the native dispatch internally. The engine never sees the difference between shell and native execution.

## Architectural Patterns

### Pattern 1: Native Func Registry (recommended)

**What:** A package-level map keyed by string name, populated by `init()` in each `native/` source file. Executor dispatch switches on `t.Executor.Type == "go"` and looks up `t.Executor.GoFunc` in the registry.

**When to use:** When native techniques are self-contained with no shared mutable state, and the number is moderate (under 20).

**Trade-offs:** Simple to add new techniques — one `Register()` call in an `init()` function. No interface contract means type errors surface at call time via the map lookup, not compile time. Acceptable at this scale; a formal interface can be introduced later if needed.

**Example:**

```go
// internal/native/registry.go
package native

// NativeFunc is the signature all native technique implementations must satisfy.
// args maps input_args keys from the YAML definition to their string values.
// Returns human-readable output and any execution error.
type NativeFunc func(args map[string]string) (string, error)

var Registry = map[string]NativeFunc{}

func Register(name string, fn NativeFunc) {
    Registry[name] = fn
}
```

```go
// internal/native/scanner.go
package native

import "net"
import "fmt"
import "strconv"
import "sync"
import "time"

func init() {
    Register("scanner_tcp", runTCPScan)
    Register("scanner_icmp", runICMPSweep)
}

func runTCPScan(args map[string]string) (string, error) {
    // net.DialTimeout worker pool — zero external deps
    // reads args["subnet"], args["ports"], args["timeout_ms"], args["workers"]
    // returns formatted scan output string
}
```

```go
// internal/executor/executor.go — added dispatch branch (inside runInternal)
import "lognojutsu/internal/native"

// ... existing runInternal function, before final return:

switch strings.ToLower(t.Executor.Type) {
case "go":
    fn, ok := native.Registry[t.Executor.GoFunc]
    if !ok {
        out, errOut, err = "", "", fmt.Errorf("unknown native func: %q", t.Executor.GoFunc)
    } else {
        output, nativeErr := fn(t.InputArgs)
        out, errOut, err = output, "", nativeErr
    }
default:
    // existing: runCommand / runCommandAs paths
}
```

```yaml
# T1046_network_scan.yaml — updated executor block
executor:
  type: go
  command: ""      # unused for go type; keep for schema compatibility
  go_func: scanner_tcp
input_args:
  subnet: "auto"   # "auto" = detect local /24 at runtime
  ports: "22,80,135,139,389,443,445,1433,3389,5985,8443"
  timeout_ms: "500"
  workers: "50"
```

### Pattern 2: Atomic Red Team Executor Model (reference — not adopted)

Atomic Red Team defines executors as `powershell`, `bash`, `sh`, `command_prompt`, `manual` only. The official schema has no native/binary executor type. The `go-atomicredteam` project (activeshadow) is a Go runner for those YAML definitions — it still shells out to the named interpreter.

**Why not adopted:** LogNoJutsu already follows this model for shell-based techniques. Adding a `"go"` type is a controlled local extension that does not break YAML compatibility — all existing techniques remain unchanged.

### Pattern 3: Caldera Native Executor Model (reference — partial adoption)

Caldera's `gocat` agent defines a formal executor interface:

```go
type Executor interface {
    Run(command string, timeout int, info InstructionInfo) ([]byte, string, string, time.Time)
    String() string
    CheckIfAvailable() bool
    UpdateBinary(newBinary string)
    DownloadPayloadToMemory(payloadName string) bool
}
```

Implementations register into `var Executors = map[string]Executor{}`. The "native" gocat extension provides pure Go TTPs (network recon, process inspection) compiled into the agent binary — no subprocess.

**What to adopt:** The registry-map pattern and the "native as compiled-in Go code" model.

**What not to adopt:** The full interface. `CheckIfAvailable`, `UpdateBinary`, `DownloadPayloadToMemory` are C2 agent concepts irrelevant to a local simulation tool. The simpler `NativeFunc = func(map[string]string) (string, error)` is sufficient.

## Data Flow

### Native Technique Execution (new path)

```
engine.run()
    |
    v
executor.runInternal(t, profile, password)
    |
    +-- t.Executor.Type == "go" ?
    |       |
    |       v
    |   native.Registry[t.Executor.GoFunc](t.InputArgs)
    |       |
    |       v
    |   (output string, error)   <-- no OS subprocess spawned
    |
    +-- else: runCommand() / runCommandAs()  <-- existing path unchanged
    |
    v
ExecutionResult{Output, Success, EndTime, ...}
    |
    v (unchanged downstream)
engine appends to status.Results
simlog.TechOutput(...)
verifier checks EventLog
reporter.SaveResults()
```

### YAML Schema Extension

One field added to the existing `Executor` struct in `playbooks/types.go`:

```go
type Executor struct {
    Type    string `yaml:"type"    json:"type"`
    Command string `yaml:"command" json:"command"`
    GoFunc  string `yaml:"go_func,omitempty" json:"go_func,omitempty"`
}
```

`GoFunc` is read only when `Type == "go"`. All existing YAML files that omit `go_func` are unaffected — the field defaults to empty string and the existing shell path is taken.

### WhatIf Mode Compatibility

The engine's WhatIf bypass fires before `executor.runInternal()` is called. Native techniques receive the same synthetic `ExecutionResult` with placeholder output. No change needed in the WhatIf path.

### Input Args for Native Techniques

`InputArgs map[string]string` already exists on `Technique` and is currently unused by the executor (shell commands embed their parameters inline). For native techniques, `InputArgs` is passed directly to `NativeFunc`. Defaults (e.g., `subnet: "auto"` meaning detect local /24) are resolved inside the NativeFunc, not in the engine or executor.

## Scaling Considerations

LogNoJutsu is single-user, local-execution. Network scaling is not relevant. The relevant scale axis is: how many native techniques before the registry pattern needs restructuring?

| Scale | Approach |
|-------|----------|
| 1-10 native funcs | Single `scanner.go` + `ldap.go`, inline `init()` registration |
| 10-30 native funcs | Split by domain: `scanner.go`, `ldap.go`, `dns.go`, `smb.go` |
| 30+ native funcs | Sub-packages per tactic with blank-import registration via `import _ "lognojutsu/internal/native/scanner"` |

At v1.3 scope (TCP scan + ICMP sweep + LDAP recon), the 1-10 range applies. One `internal/native/` package with two or three files is the right size.

## Anti-Patterns

### Anti-Pattern 1: New Executor Package Per Technique

**What people do:** Create `internal/executor/native_scanner.go`, `internal/executor/native_ldap.go`, etc., each with exported functions called directly from executor.go.

**Why it's wrong:** `executor.go` becomes a routing table with direct function calls to dozens of files. Adding a technique requires editing both the implementation file and `executor.go`. The `native/` boundary is lost — techniques accumulate in the executor package.

**Do this instead:** One `internal/native/` package with `init()` registration per file. `executor.go` imports `native` once and the registry grows without touching `executor.go`.

### Anti-Pattern 2: Native Func Calling Back Into simlog or verifier

**What people do:** Have `runTCPScan()` directly call `simlog.TechOutput()` to emit structured log entries mid-scan, or call `verifier.Check()` to verify events inline.

**Why it's wrong:** Creates an import cycle risk (`native` imports `simlog`, `simlog` has no dependency on `native` — but if simlog ever imports playbooks it could close a cycle). More importantly it breaks the executor contract: callers of `NativeFunc` expect `(string, error)` and handle logging themselves.

**Do this instead:** Return all output as a formatted string from `NativeFunc`. The executor layer passes that string to `simlog.TechOutput()` exactly as it does for shell techniques today. Native funcs stay pure — no side effects outside their return value.

### Anti-Pattern 3: Converting Existing PowerShell Techniques to Go Unnecessarily

**What people do:** Rewrite T1059 (PowerShell execution) in native Go because "native is better."

**Why it's wrong:** PowerShell techniques exist specifically to generate PowerShell event logs (EID 4103/4104 ScriptBlock logging, EID 4688 process creation). A native Go replacement would not generate those events — defeating the SIEM validation purpose. The SIEM needs to see PowerShell events to validate its PowerShell detection rules.

**Do this instead:** Use `type: go` only when the technique's detection signal comes from network/OS artifacts that Go can generate natively — Sysmon EID 3 network connections, Security EID 5156 WFP filter hits, Security EID 4624 logon events from LDAP bind. PowerShell techniques stay as PowerShell.

### Anti-Pattern 4: PowerShell LDAP Queries via Get-ADUser

**What people do:** Write PowerShell that calls `Get-ADUser` or accesses `([adsi]"LDAP://...")` to simulate LDAP reconnaissance.

**Why it's wrong:** `Get-ADUser` requires the RSAT ActiveDirectory module, which is not installed on most client machines by default. The command fails silently or with confusing errors during client engagements.

**Do this instead:** Use `github.com/go-ldap/ldap/v3` with anonymous bind (or authenticated bind if credentials are configured) to enumerate rootDSE and basic AD objects. This works on any domain-joined or domain-routable machine with port 389 accessible. It generates the same Sysmon EID 3 network connection events and Security EID 5156 WFP filter events that the SIEM should detect.

### Anti-Pattern 5: ICMP Sweep Requiring Raw Socket Mode

**What people do:** Use `golang.org/x/net/icmp` in privileged raw socket mode (`ListenPacket("ip4:icmp", ...)`) for host discovery.

**Why it's wrong:** Raw socket mode requires elevated privileges and has known Windows issues (`x/net/icmp` CPU-hogging and packet-capture problems on Windows reported in golang/go#33117 and golang/go#38427).

**Do this instead:** Use `go-ping` library (or `x/net/icmp` in datagram/UDP mode with `SetPrivileged(false)`) which works on Windows 10/11 without requiring administrator privileges. Alternatively, fall back to TCP connect probes on port 445 for host-alive detection — this is reliable, generates Sysmon EID 3 events, and requires only standard user permissions.

## Integration Points

### New Dependencies (go.mod additions)

| Library | Purpose | Confidence |
|---------|---------|------------|
| `github.com/go-ldap/ldap/v3` | LDAP v3 client — anonymous + authenticated bind, search with paging, rootDSE enumeration | HIGH — actively maintained, used in production AD tooling (go-windapsearch), pure Go, no CGO |
| `golang.org/x/net` | ICMP datagram mode for host sweep (optional — TCP fallback may suffice) | MEDIUM — known Windows issues in raw socket mode; datagram mode is safer; validate at runtime |

TCP port scan uses only `net.DialTimeout` from the stdlib — zero new deps.

### Changes to Existing Packages

| Package | Change | Scope |
|---------|--------|-------|
| `internal/playbooks/types.go` | Add `GoFunc string` to `Executor` struct | One field, one line |
| `internal/executor/executor.go` | Add `case "go":` dispatch in `runInternal()` | One switch branch, ~10 lines, one new import |
| `go.mod` | Add `go-ldap/ldap/v3` (and optionally `golang.org/x/net`) | Build-only change; single binary preserved |

### Packages Requiring No Changes

`internal/engine/engine.go`, `internal/server/server.go`, `internal/simlog/simlog.go`, `internal/reporter/reporter.go`, `internal/verifier/`, `internal/userstore/`, `internal/preparation/` — all unchanged.

### YAML Technique Files

New technique files use `type: go`. Existing files are untouched. The loader's `//go:embed embedded` directive picks up all `.yaml` files at build time — no code changes to the loader.

## Build Order (Dependency-aware)

Each step has all its dependencies available before it begins.

1. **`internal/playbooks/types.go`** — Extend `Executor` struct with `GoFunc string omitempty`. Zero deps. This is a pure additive change.
2. **`internal/native/registry.go`** — Define `NativeFunc` type and `Registry` map with `Register()`. No imports from any internal package.
3. **`internal/native/scanner.go`** — TCP + ICMP implementations. Depends on: stdlib `net`, optionally `golang.org/x/net/icmp`. Register via `init()`.
4. **`internal/native/ldap.go`** — LDAP query implementations. Depends on: `github.com/go-ldap/ldap/v3`. Register via `init()`.
5. **`internal/native/native_test.go`** — Unit tests for each NativeFunc using real network (localhost) or mock. No dependency on other internal packages.
6. **`internal/executor/executor.go`** — Add native dispatch branch. Add `import "lognojutsu/internal/native"`. One switch case in `runInternal()`.
7. **`internal/playbooks/embedded/techniques/`** — New and modified YAML files. No code change; picked up by existing embed. Existing techniques with `type: powershell` remain unchanged.
8. **Integration test** — If needed: verify engine runs a native technique end-to-end in WhatIf mode (avoids needing an AD environment).

## Sources

- Atomic Red Team YAML Schema: https://github.com/redcanaryco/atomic-red-team/wiki/YAML-Schema
- go-atomicredteam (Go runner for Atomic Red Team): https://github.com/activeshadow/go-atomicredteam
- Caldera gocat executor interface (verified source): https://raw.githubusercontent.com/mitre/gocat/master/execute/execute.go
- Caldera Sandcat plugin details: https://caldera.readthedocs.io/en/latest/plugins/sandcat/Sandcat-Details.html
- go-windapsearch (Go LDAP AD enumeration reference implementation): https://github.com/ropnop/go-windapsearch
- go-ldap/ldap/v3 package: https://pkg.go.dev/github.com/go-ldap/ldap/v3
- golang.org/x/net/icmp package: https://pkg.go.dev/golang.org/x/net/icmp
- go-ping (ICMP on Windows without raw socket): https://github.com/go-ping/ping
- golang/go#33117 — x/net/icmp Windows CPU issue: https://github.com/golang/go/issues/33117
- golang/go#38427 — x/net/icmp Windows packet capture issue: https://github.com/golang/go/issues/38427

---
*Architecture research for: LogNoJutsu v1.3 realistic attack simulation integration*
*Researched: 2026-04-09*
