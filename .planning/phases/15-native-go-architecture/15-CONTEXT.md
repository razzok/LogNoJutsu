# Phase 15: Native Go Architecture - Context

**Gathered:** 2026-04-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Add a native Go executor type (`type: go`) to the technique execution pipeline, create an `internal/native/` registry for Go-implemented techniques, and integrate go-ldap and go-wmi libraries. The phase delivers the architecture and two real techniques (T1482 via LDAP, T1057 via WMI) that prove it works end-to-end. No shell stubs — both techniques perform real queries.

</domain>

<decisions>
## Implementation Decisions

### Go Technique Registry
- **D-01:** Init-time registration pattern — each Go technique file calls `native.Register()` in its `init()`. Adding a technique = adding a `.go` file. Mirrors how `playbooks.Registry` auto-populates.
- **D-02:** `NativeFunc` signature returns a structured `NativeResult` (Output, ErrorOutput, Success) plus an `error`. Error = infrastructure failure (can't connect); `!Success` = technique ran but didn't find expected data.
- **D-03:** Go cleanup functions registered alongside the technique via `Register(id, fn, cleanup)`. Cleanup can be nil for read-only techniques. Executor's defer pattern calls the Go cleanup — no shell needed.
- **D-04:** First real Go technique is T1482 (Domain Trust Discovery via LDAP). Validates full pipeline: YAML `type: go` → executor dispatch → native registry → LDAP query → result. Satisfies success criteria 2 and 3.

### LDAP Integration
- **D-05:** DC auto-discovery: check `LOGONSERVER` env var first, then DNS SRV lookup (`_ldap._tcp.dc._msdcs.DOMAIN`). Graceful fallback with `ErrNoDCReachable` when no DC found.
- **D-06:** Authentication via current user context — NTLM bind with empty credentials (go-ldap `NTLMBind("", "")`). No credential passing needed. Matches real attacker discovery behavior.
- **D-07:** T1482 queries `trustedDomain` objects from `CN=System,DC=...` only — returns trust name, direction, type. Does not enumerate users/groups (that's T1069/T1087 in Phase 18).

### WMI Integration
- **D-08:** Use go-wmi high-level API (`github.com/yusufpapurcu/wmi`) — struct-based result binding, wraps go-ole internally. Pure Go, no CGO.
- **D-09:** Initial WMI technique targets `Win32_Process` (T1057 Process Discovery). Returns PID, name, command line, parent PID. Validates the go-ole/wmi pipeline.

### Executor Dispatch
- **D-10:** New `go` case in `runInternal()` — calls `native.Lookup(t.ID)` to find the registered Go function, runs it directly, maps `NativeResult` to `ExecutionResult`. Minimal change to existing flow.
- **D-11:** No RunAs support for Go techniques — they always run as the current process user. User impersonation via Windows token APIs is complex and low value for discovery techniques. A log note is emitted when RunAs is configured but technique is `type: go`.
- **D-12:** Cleanup for `type: go` in `RunWithCleanup()` checks `native.LookupCleanup(t.ID)` first. If a Go cleanup function is registered, defer calls it. The YAML `cleanup` field is not used for Go techniques.

### Claude's Discretion
- Package layout within `internal/native/` (e.g., flat vs `techniques/` subdirectory)
- Error message wording for graceful fallback scenarios
- Timeout values for LDAP and WMI connections
- Output formatting of technique results (tables, key-value, etc.)
- Whether to add a `context.Context` with timeout to native function calls

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Executor & Technique Model
- `internal/executor/executor.go` — Current dispatch logic (`runInternal`, `runCommand`, `RunWithCleanup`); `type: go` case added here
- `internal/playbooks/types.go` — `Technique`, `Executor`, `ExecutionResult` structs; `Executor.Type` field determines dispatch

### Playbook Registry
- `internal/playbooks/loader.go` — Embedded YAML loading; reference for how the existing registry pattern works

### Existing Techniques (will be converted to type: go)
- `internal/playbooks/embedded/techniques/T1482_domain_trust_discovery.yaml` — Current T1482 (PowerShell stub) to be replaced with `type: go`
- `internal/playbooks/embedded/techniques/T1057_process_discovery.yaml` — Current T1057 (PowerShell stub) to be replaced with `type: go`

### Requirements
- `.planning/REQUIREMENTS.md` §Architecture — ARCH-01, ARCH-02, ARCH-03

### Phase 14 Context (cleanup patterns)
- `.planning/phases/14-safety-audit/14-CONTEXT.md` — Defer-style cleanup guarantee (D-09, D-10) that Go techniques must honor

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `RunWithCleanup()` defer pattern in `executor.go` — Go cleanup functions integrate into this same pattern
- `ExecutionResult` struct in `types.go` — `NativeResult` maps directly to Output/ErrorOutput/Success fields
- `simlog.TechStart/TechEnd/TechCleanup` — Same logging calls used for Go techniques
- `Executor.Type` field already supports switch dispatch — adding `"go"` is a natural extension

### Established Patterns
- Init-time registration: `playbooks.Registry` loaded once at startup via `LoadEmbedded()` — native registry follows same pattern
- Struct tags aligned columns: `yaml:"field" json:"field"` — new types follow this convention
- Error wrapping with `fmt.Errorf("...: %w", err)` at each boundary
- Typed string constants for enumerations (`Phase`, `VerificationStatus`) — use for native error types if needed

### Integration Points
- `internal/executor/executor.go:runInternal()` — Add `type: go` branch before shell dispatch
- `internal/executor/executor.go:RunWithCleanup()` — Add native cleanup lookup before existing shell cleanup defer
- `go.mod` — Add `github.com/go-ldap/ldap/v3` and `github.com/yusufpapurcu/wmi` dependencies
- `internal/native/` — New package (does not exist yet)

</code_context>

<specifics>
## Specific Ideas

- T1482 YAML should have `executor.type: go` and `executor.command: ""` (command field ignored for Go techniques)
- DC discovery fallback chain: LOGONSERVER env var → DNS SRV → graceful error message (not crash)
- go-wmi struct binding means Win32_Process fields map directly to a Go struct — clean and type-safe
- Both new Go techniques should be Tier 1 (real queries, real SIEM artifacts) — upgraded from their current Tier 3 stubs

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 15-native-go-architecture*
*Context gathered: 2026-04-09*
