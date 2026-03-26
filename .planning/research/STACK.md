# Stack Research

**Domain:** Go CLI/Web tool — build-time versioning and Windows system command locale fix
**Researched:** 2026-03-26
**Confidence:** HIGH

## Context

This is a targeted research file for the v1.1 milestone. The project is an existing Go 1.26.1
single-binary tool (`module lognojutsu`). No new runtime dependencies are being added.
The two features under research are:

1. **Build-time version injection** via Go `ldflags` — replaces the hardcoded `v0.1.0` string in
   `cmd/lognojutsu/main.go` (banner) and `internal/server/static/index.html` (version badge).
2. **Locale-independent auditpol subcategory lookup** — replaces English subcategory names in
   `internal/preparation/preparation.go` with fixed Windows GUIDs so the step works on German and
   other non-English Windows installations.

---

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go `linker -X` flag | stdlib (any Go 1.5+) | Inject a `string` variable value at link time | Built into the Go toolchain — zero dependency, no external tool needed. The `-ldflags "-X pkg.Var=value"` syntax is stable since Go 1.5 and unchanged through Go 1.26. |
| `auditpol.exe /subcategory:{GUID}` | Windows built-in | Address audit subcategories without locale-sensitive names | Microsoft's own protocol spec [MS-GPAC] defines fixed GUIDs for every subcategory. The GUID form is ignored by locale; only the GUID is used when the policy is applied. |

### Supporting Libraries

No new dependencies. Both features are implemented with existing stdlib (`os/exec`, `fmt`) and
build tooling already present in the project.

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `gopkg.in/yaml.v3` | v3.0.1 | Already present — playbook loading | Not relevant to v1.1 features |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `go build -ldflags` | Inject version at build time | No wrapper tool needed; raw `go build` command is sufficient |
| PowerShell `$env:GIT_TAG` / `git describe` | Supply version value to build command | Standard pattern: `go build -ldflags "-X main.Version=$(git describe --tags --always)"` |

---

## Feature 1: Build-Time Version Injection

### Mechanism

Go's linker supports `-X importpath.name=value` to set a package-level `string` variable before
the binary is written. Requirements:

- The target variable must be a `var` (not `const`) at package scope.
- The type must be `string`.
- The value cannot be the result of a function call.

### Required Change — Go Source

Declare a package-level variable in `cmd/lognojutsu/main.go`. The current hardcoded string inside
the `banner` const must become a `var` that the linker can overwrite:

```go
// Declare at package level — linker target for -ldflags -X
var Version = "dev"
```

Then reference `Version` inside the banner string instead of the literal `v0.1.0`.

Because the banner uses a raw string literal (backtick), concatenation is the cleanest approach:

```go
const bannerTemplate = `
 ...ascii art...
  SIEM Validation & ATT&CK Simulation Tool  `

var Version = "dev"

var banner = bannerTemplate + Version + "\n"
```

### Required Change — HTML

The version badge at `index.html` line 167 is a static string served from disk. To make it
dynamic, the server must inject the version into the template at startup rather than serving raw
HTML. Two options:

| Option | Mechanism | Complexity |
|--------|-----------|------------|
| `html/template` with `{{.Version}}` | Pass `Version` var into template execute | Low — existing server already uses `html/template` for the report |
| String replacement at serve time | `strings.ReplaceAll` on HTML bytes | Minimal — works without restructuring static file serving |

Recommendation: use the `html/template` approach if `index.html` is already executed as a
template. If it is served as a static file via `http.FileServer`, a simple
`strings.ReplaceAll(html, "v0.1.0", Version)` applied once at server start is the lowest-friction
path and avoids restructuring the static file handler.

### Build Command Pattern

```bash
# Development build (Version defaults to "dev")
go build ./cmd/lognojutsu/

# Release build (version supplied externally)
go build -ldflags "-X main.Version=v1.1.0" -o lognojutsu.exe ./cmd/lognojutsu/

# With git tag
go build -ldflags "-X main.Version=$(git describe --tags --always)" -o lognojutsu.exe ./cmd/lognojutsu/
```

On Windows PowerShell (the distribution target):

```powershell
$ver = git describe --tags --always
go build -ldflags "-X main.Version=$ver" -o lognojutsu.exe ./cmd/lognojutsu/
```

The `-X` path must match the module path. With `module lognojutsu` and the variable in
`cmd/lognojutsu/main.go`, the fully-qualified path is `main.Version` (the `main` package path in
a Go binary is always `main`, not the module path).

---

## Feature 2: Locale-Independent auditpol GUIDs

### Problem

`auditpol.exe /set /subcategory:Logon` works only when the OS locale is English. On German
Windows, the subcategory name is localised (e.g. `Anmeldung`) and `auditpol` rejects the English
string with an error. The current code passes English names directly, causing all 12 subcategory
calls to fail silently (failures are collected but the step reports partial failure rather than
panicking).

### Solution

Replace every subcategory name string with its fixed Windows GUID. The GUID form is fully
locale-independent — confirmed by Microsoft's [MS-GPAC] protocol specification which states:
"The Subcategory field is for user reference only and is ignored when the advanced audit policy is
applied."

Syntax accepted by `auditpol.exe`:

```
auditpol /set /subcategory:{0CCE9215-69AE-11D9-BED3-505054503030} /success:enable /failure:enable
```

The GUID must be wrapped in braces `{}` within the argument string.

### GUID Reference Table

All GUIDs sourced from Microsoft [MS-GPAC] §2.6.1
(`learn.microsoft.com/en-us/openspecs/windows_protocols/ms-gpac/77878370-0712-47cd-997d-b07053429f6d`).
These values are fixed by the Windows protocol and do not change between OS versions.

| Current Subcategory Name | GUID | Events |
|--------------------------|------|--------|
| `Logon` | `{0CCE9215-69AE-11D9-BED3-505054503030}` | 4624, 4625, 4634 |
| `Logoff` | `{0CCE9216-69AE-11D9-BED3-505054503030}` | Logoff events |
| `Account Lockout` | `{0CCE9217-69AE-11D9-BED3-505054503030}` | 4740 |
| `Process Creation` | `{0CCE922B-69AE-11D9-BED3-505054503030}` | 4688 |
| `Audit Policy Change` | `{0CCE922F-69AE-11D9-BED3-505054503030}` | 4719 |
| `Security Group Management` | `{0CCE9237-69AE-11D9-BED3-505054503030}` | 4728, 4732 |
| `User Account Management` | `{0CCE9235-69AE-11D9-BED3-505054503030}` | 4720 |
| `Sensitive Privilege Use` | `{0CCE9228-69AE-11D9-BED3-505054503030}` | 4673, 4674 |
| `Special Logon` | `{0CCE921B-69AE-11D9-BED3-505054503030}` | 4672 |
| `Other Object Access Events` | `{0CCE9227-69AE-11D9-BED3-505054503030}` | Task Scheduler / COM+ |
| `Scheduled Task` | `{0CCE9227-69AE-11D9-BED3-505054503030}` | 4698 — same GUID as Other Object Access |
| `Filtering Platform Connection` | `{0CCE9226-69AE-11D9-BED3-505054503030}` | Network connections |

Note on `Scheduled Task`: Windows does not have a separate "Scheduled Task" subcategory GUID.
Scheduled task events (4698) fall under **Other Object Access Events**
(`{0CCE9227-69AE-11D9-BED3-505054503030}`). The current code lists both "Other Object Access
Events" and "Scheduled Task" as separate entries — after the GUID migration these should be
deduplicated to a single entry using `{0CCE9227-69AE-11D9-BED3-505054503030}`.

### Required Change — preparation.go

Replace the `subcategory` string field with the GUID string. The `exec.Command` call does not
change shape — only the value passed to `/subcategory:` changes:

```go
// Before
exec.Command("auditpol.exe", "/set", "/subcategory:Logon", "/success:enable", "/failure:enable")

// After
exec.Command("auditpol.exe", "/set", "/subcategory:{0CCE9215-69AE-11D9-BED3-505054503030}", "/success:enable", "/failure:enable")
```

No new imports, no new dependencies. The description field in the `policies` slice becomes the
primary human-readable label in error messages, which is already the pattern in the existing code.

---

## Alternatives Considered

| Recommended | Alternative | Why Not |
|-------------|-------------|---------|
| `go build -ldflags -X main.Version=...` | `govvv` wrapper tool | govvv is an unmaintained third-party wrapper (last commit 2019). Raw ldflags achieve the same result with zero dependency. |
| `go build -ldflags -X main.Version=...` | Embed version in a file read at runtime | Adds file I/O dependency at startup; breaks single-binary distribution model. |
| GUID-based auditpol call | PowerShell `Set-PolicyFileEntry` or `EvotecIT/AuditPolicy` module | Adds PowerShell module dependency; the project already uses `auditpol.exe` exec which works fine — only the argument needs changing. |
| GUID-based auditpol call | `auditpol /list /subcategory:* /v` to resolve names dynamically | Adds runtime discovery overhead; GUIDs are protocol constants and will not change. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `govvv` | Unmaintained since 2019; not needed when `go build -ldflags` is available | `go build -ldflags "-X main.Version=..."` |
| English subcategory names in auditpol | Fail silently on non-English Windows installations | GUID form (`{0CCE9215-...}`) |
| `const Version` | The linker `-X` flag cannot overwrite constants, only `var` declarations | `var Version = "dev"` |
| Module-path prefix for `-X` (e.g., `-X lognojutsu/cmd/lognojutsu/main.Version`) | The `main` package path in a compiled Go binary is always `main`, not the module path | `-X main.Version=...` |

---

## Version Compatibility

| Component | Compatibility Note |
|-----------|--------------------|
| `go build -ldflags -X` | Stable since Go 1.5. No version-specific concerns for Go 1.26.1. |
| auditpol GUID syntax | Supported on Windows Vista and later (all targets: Win10/11/Server 2016+). |
| `{GUID}` brace format | Required by auditpol — omitting braces causes "The parameter is incorrect" error. |

---

## Sources

- [Microsoft Learn: auditpol set](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/auditpol-set) — confirmed GUID syntax `/subcategory:{guid}` is valid; HIGH confidence
- [MS-GPAC §2.6.1: Subcategory and SubcategoryGUID](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-gpac/77878370-0712-47cd-997d-b07053429f6d) — authoritative GUID table for all 50 Windows audit subcategories; HIGH confidence
- [DigitalOcean: Using ldflags to Set Version Information for Go Applications](https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications) — ldflags -X pattern and requirements (var not const); MEDIUM confidence (verified against Go stdlib docs)
- [Leapcell: Advanced Go Linker Usage](https://leapcell.io/blog/advanced-go-linker-usage-injecting-version-info-and-build-configurations) — multiple variable injection syntax; MEDIUM confidence
- [community.windows issue #14](https://github.com/ansible-collections/community.windows/issues/14) — real-world confirmation that auditpol subcategory names fail on non-English Windows; MEDIUM confidence

---

*Stack research for: LogNoJutsu v1.1 — build-time versioning and locale-independent auditpol*
*Researched: 2026-03-26*
