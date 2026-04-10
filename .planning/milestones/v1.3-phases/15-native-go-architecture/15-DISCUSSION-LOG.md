# Phase 15: Native Go Architecture - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-09
**Phase:** 15-native-go-architecture
**Areas discussed:** Go technique registry design, LDAP integration approach, WMI integration approach, Executor dispatch changes

---

## Go Technique Registry Design

### Registration Pattern

| Option | Description | Selected |
|--------|-------------|----------|
| Init-time registration | Each Go technique file calls Register() in init(). Registry builds automatically at startup. | ✓ |
| Explicit map in registry.go | Single file maps technique IDs to Go functions manually. More visible but two-place updates. | |

**User's choice:** Init-time registration
**Notes:** Matches how playbooks.Registry works. Adding a technique = adding a .go file.

### Function Signature

| Option | Description | Selected |
|--------|-------------|----------|
| Structured NativeResult | Return struct with Output, ErrorOutput, Success. Maps cleanly to ExecutionResult. | ✓ |
| Simple (string, error) | Return output string + error. Simpler but less expressive. | |

**User's choice:** Structured NativeResult
**Notes:** error = infrastructure failure; !Success = technique ran but didn't find expected data.

### Cleanup Model

| Option | Description | Selected |
|--------|-------------|----------|
| Go cleanup function | Register(id, fn, cleanup). Executor's defer calls it. No shell needed. | ✓ |
| YAML cleanup field only | Go techniques use YAML cleanup: with shell commands. | |
| Both available | Go cleanup takes priority; falls back to YAML cleanup. | |

**User's choice:** Go cleanup function
**Notes:** Cleanup can be nil for read-only techniques.

### First Technique

| Option | Description | Selected |
|--------|-------------|----------|
| T1482 domain trust (real) | Real technique using go-ldap. Validates full pipeline end-to-end. | ✓ |
| Minimal hello-world | Stub to prove architecture. Real techniques in Phase 18. | |
| Both | Hello-world first, then T1482. | |

**User's choice:** T1482 domain trust
**Notes:** Satisfies success criteria 2 and 3 simultaneously.

---

## LDAP Integration Approach

### DC Discovery

| Option | Description | Selected |
|--------|-------------|----------|
| Windows API DNS query | LOGONSERVER env var → DNS SRV lookup → graceful fallback. | ✓ |
| YAML input_args config | DC address specified in technique YAML. Requires per-engagement config. | |

**User's choice:** Windows API DNS query
**Notes:** Works on domain-joined machines without configuration.

### Authentication

| Option | Description | Selected |
|--------|-------------|----------|
| Current user context | NTLM bind with empty credentials. No credential passing needed. | ✓ |
| Explicit credentials from userstore | Pass credentials from userstore profile. | |

**User's choice:** Current user context
**Notes:** Matches how a real attacker doing discovery would operate.

### Query Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Trust objects only | trustedDomain objects from CN=System. Focused on T1482. | ✓ |
| Broad AD enumeration | Users, groups, OUs alongside trusts. Overlaps Phase 18. | |

**User's choice:** Trust objects only
**Notes:** Does not enumerate users/groups — that's T1069/T1087 in Phase 18.

---

## WMI Integration Approach

### Target WMI Class

| Option | Description | Selected |
|--------|-------------|----------|
| Win32_Process | Process listing maps to T1057. Returns PID, name, command line, owner. | ✓ |
| Win32_ComputerSystem | System info maps to T1082. Simpler query. | |
| Both | Two WMI techniques for thorough validation. | |

**User's choice:** Win32_Process
**Notes:** Validates go-ole/wmi pipeline with a familiar WMI query.

### Library Choice

| Option | Description | Selected |
|--------|-------------|----------|
| go-wmi high-level | github.com/yusufpapurcu/wmi. Struct-based binding, wraps go-ole. No CGO. | ✓ |
| go-ole directly | Lower-level COM automation. More control but more boilerplate. | |

**User's choice:** go-wmi high-level
**Notes:** Clean Query() API with struct auto-binding.

---

## Executor Dispatch Changes

### Dispatch Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| New case in runInternal | Add 'go' branch that calls native.Lookup(t.ID). Minimal change. | ✓ |
| Separate RunNative function | New exported function alongside Run/RunAs/RunWithCleanup. | |

**User's choice:** New case in runInternal
**Notes:** Maps NativeResult to ExecutionResult the same way shell output does.

### RunAs Support

| Option | Description | Selected |
|--------|-------------|----------|
| No RunAs for Go techniques | Always run as current process user. Log note when RunAs configured. | ✓ |
| Token-based impersonation | Windows LogonUser + ImpersonateLoggedOnUser via syscall. | |

**User's choice:** No RunAs for Go techniques
**Notes:** Token APIs are complex and low value for discovery techniques.

### Cleanup Dispatch

| Option | Description | Selected |
|--------|-------------|----------|
| Native cleanup via registry | LookupCleanup(t.ID) first, defer calls Go function. | ✓ |

**User's choice:** Native cleanup via registry
**Notes:** Integrates with existing RunWithCleanup defer pattern.

---

## Claude's Discretion

- Package layout within internal/native/
- Error message wording for fallback scenarios
- Timeout values for LDAP and WMI connections
- Output formatting of technique results
- context.Context with timeout for native function calls

## Deferred Ideas

None — discussion stayed within phase scope
