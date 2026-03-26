# Pitfalls Research

**Domain:** Go CLI/web tool — build-time version injection, Windows audit policy GUID migration, single-file SPA polish
**Researched:** 2026-03-26
**Confidence:** HIGH (ldflags, GUID stability), MEDIUM (SPA regression patterns)

---

## Critical Pitfalls

### Pitfall 1: Targeting a `const` with ldflags -X (silently does nothing)

**What goes wrong:**
`go build -ldflags "-X main.Version=1.1.0"` runs without error but the binary still prints the old value. The build succeeds, tests pass, the badge still reads `v0.1.0`.

In `cmd/lognojutsu/main.go` the banner string is a `const`:

```go
const banner = `...  v0.1.0`
```

`-X` cannot rewrite a `const`. Go issue #20649 confirms the linker silently ignores `-X` when the symbol is a constant — it does not emit a warning or build error.

**Why it happens:**
Developers assume `-X` works like a preprocessor macro that replaces any string. It only rewrites **package-level `var` declarations of type `string`** that are either uninitialized or initialized to a literal string. A `const` is baked into the read-only data segment at compile time; the linker has no slot to patch.

**How to avoid:**
Split the version out of the `const` banner into a separate `var`:

```go
var version = "dev"   // injected at build time

const bannerTemplate = `
 ...banner art...
  SIEM Validation & ATT&CK Simulation Tool  %s
`
// In main(): fmt.Printf(bannerTemplate, version)
```

Then inject with: `go build -ldflags "-X main.version=1.1.0"`.

The HTML badge `<span class="version-badge">v0.1.0</span>` is in `internal/server/static/index.html` which is embedded via `//go:embed static`. That file cannot be patched by ldflags at all — it requires either a Go template served dynamically, or a `/api/version` endpoint that the JS badge reads on load. The simpler approach for this project is a `/api/version` endpoint.

**Warning signs:**
- `strings.Contains(binaryOutput, "v0.1.0")` is still true after a build with `-X`.
- `go tool nm ./lognojutsu.exe | grep version` shows no symbol named `version` — the symbol doesn't exist because it was declared as `const`.

**Phase to address:** Phase that implements ldflags version injection.

---

### Pitfall 2: Shell quoting of ldflags breaks on Windows cmd.exe and PowerShell differently

**What goes wrong:**
The standard Unix form `go build -ldflags "-X 'main.version=1.1.0'"` fails silently or inserts literal single-quotes into the version string on Windows cmd.exe and on older PowerShell. The binary ends up with `version = "'1.1.0'"` or an empty string.

**Why it happens:**
cmd.exe does not interpret single quotes as string delimiters. PowerShell 5.x (default on Windows 10/11) handles quote nesting inconsistently when invoking external processes. Go issue #16743 documents that `-X` values with spaces require special handling, and issue #43179 shows that flag values ending with `.` break the parser through PowerShell.

**How to avoid:**
Use double-quote syntax and escape the inner double quotes explicitly when calling from cmd.exe:

```bat
REM cmd.exe — escape inner quotes with backslash
go build -ldflags "-X main.version=1.1.0" .
REM (works as long as the version string contains no spaces)
```

For PowerShell:
```powershell
# PowerShell — use single quotes for the outer wrapper
go build -ldflags '-X main.version=1.1.0' .
```

Keep version strings free of spaces and special shell characters. Use semver format (`1.1.0`, not `v1.1.0 (beta)`) as the injected value. The `v` prefix can be a Go-side default: `var version = "v" + "dev"` combined with injecting just the numeric part, or inject the full `v1.1.0` string with no spaces.

A Makefile or `go generate` script that constructs the ldflags string is the safest cross-platform approach; hard-code the quoting style for the target shell once rather than re-discovering it per developer.

**Warning signs:**
- Version badge or banner shows `'1.1.0'` (with quote chars) or is empty.
- `go build` exits 0 but `.\lognojutsu.exe --version` prints the fallback `dev`.

**Phase to address:** Phase that implements ldflags version injection.

---

### Pitfall 3: auditpol GUID stability assumption — new subcategories introduced in recent Windows builds

**What goes wrong:**
The assumption "all GUIDs are stable" is mostly true for classic subcategories, but Microsoft has introduced new subcategories in specific Windows feature updates (e.g., "Access Rights" was added in Windows 10 20H2 / Server 2022 Build 20348.502, August 2022 CU). If the Go code references a GUID that does not exist on the target Windows version, `auditpol /set /subcategory:{GUID}` returns exit code 1 with the message "The parameter is incorrect."

For the twelve subcategories currently used in `ConfigureAuditPolicy`, all are classic Vista-era subcategories and their GUIDs have been stable since Windows Vista / Server 2008. The risk is low for the current set — but the failure mode is the same silent partial failure already in the existing code's `failures` slice.

**Why it happens:**
GUIDs for audit subcategories are assigned at OS design time and shipped in `C:\Windows\System32\adtschema.dll`. Older subcategory GUIDs do not change when Windows is updated. New subcategories are additive — they get new GUIDs. A GUID used on Windows 11 24H2 for a new subcategory will not exist on Windows Server 2016.

**How to avoid:**
Use only the well-known Vista-era GUIDs for the twelve subcategories already in the codebase. The complete mapping for those twelve is:

| Subcategory (English name) | GUID |
|---|---|
| Logon | `{0CCE9215-69AE-11D9-BED3-505054503030}` |
| Logoff | `{0CCE9216-69AE-11D9-BED3-505054503030}` |
| Account Lockout | `{0CCE9217-69AE-11D9-BED3-505054503030}` |
| Process Creation | `{0CCE922B-69AE-11D9-BED3-505054503030}` |
| Audit Policy Change | `{0CCE9223-69AE-11D9-BED3-505054503030}` |
| Security Group Management | `{0CCE9237-69AE-11D9-BED3-505054503030}` |
| User Account Management | `{0CCE9235-69AE-11D9-BED3-505054503030}` |
| Sensitive Privilege Use | `{0CCE9228-69AE-11D9-BED3-505054503030}` |
| Special Logon | `{0CCE921B-69AE-11D9-BED3-505054503030}` |
| Other Object Access Events | `{0CCE9227-69AE-11D9-BED3-505054503030}` |
| Scheduled Task | `{0CCE9232-69AE-11D9-BED3-505054503030}` |
| Filtering Platform Connection | `{0CCE9226-69AE-11D9-BED3-505054503030}` |

**Verify these GUIDs on an English Windows machine before shipping**: `auditpol /list /subcategory:* /v`

**Warning signs:**
- `auditpol /set /subcategory:{GUID}` exits non-zero with "The parameter is incorrect" — the GUID does not exist on that Windows version.
- GUIDs that deviate from the `0CCE9xxx-69AE-11D9-BED3-505054503030` pattern for these classic subcategories indicate a lookup error.

**Phase to address:** Phase that migrates `ConfigureAuditPolicy` from name-based to GUID-based calls.

---

### Pitfall 4: GUID-based auditpol calls still silently partially fail without surfacing which GUID failed

**What goes wrong:**
The existing `ConfigureAuditPolicy` already accumulates failures into a `failures` slice and returns `Success: false` with a message. However, the message currently contains the English subcategory name (e.g., `"Logon: exit status 1 (...)"`). After the GUID migration, the error message will show a raw GUID string, making the Preparation tab error unreadable to the consultant running the tool on a client machine.

**Why it happens:**
The GUID-only error message is worse for operators than the English name. The fix that removes locale-dependent names must not also remove human-readable context from error output.

**How to avoid:**
Maintain a parallel human-readable label alongside each GUID in the struct. The existing struct already has a `description` field — keep it:

```go
policies := []struct {
    guid        string
    description string
}{
    {"{0CCE9215-69AE-11D9-BED3-505054503030}", "Logon (4624/4625/4634)"},
    ...
}
// On failure: fmt.Sprintf("%s [%s]: %v", p.description, p.guid, err)
```

The `auditpol /set` command is passed the GUID for locale independence. The error message shown in the UI uses `description` for readability.

**Warning signs:**
- Preparation tab shows error text like `{0CCE9215-69AE-11D9-BED3-505054503030}: exit status 1` with no readable label.

**Phase to address:** Phase that migrates `ConfigureAuditPolicy` from name-based to GUID-based calls.

---

### Pitfall 5: Hardcoded version badge in embedded HTML is not reachable by ldflags

**What goes wrong:**
`internal/server/static/index.html` contains `<span class="version-badge">v0.1.0</span>` as a static string. The file is embedded into the binary via `//go:embed static` at compile time. `go build -ldflags "-X ..."` has no mechanism to modify the content of embedded files. The badge will always show `v0.1.0` regardless of the linker flag.

**Why it happens:**
`embed.FS` snapshots files at compile time into the binary's read-only data section. ldflags -X operates on Go symbol table entries for package-level string vars, not on arbitrary byte sequences inside embedded blobs.

**How to avoid:**
Do not attempt to put the version string into the static HTML directly. Instead:

1. Expose a `/api/version` HTTP endpoint in `server.go` that returns `{"version": version}` where `version` is the Go `var` injected via ldflags.
2. In `index.html`, replace the static span with a JavaScript fetch on `DOMContentLoaded`:

```javascript
fetch('/api/version').then(r => r.json()).then(d => {
  document.querySelector('.version-badge').textContent = d.version;
});
```

This approach is already consistent with how the dashboard polls `/api/status` — the HTML is static, live data comes from the API.

**Warning signs:**
- After a build with `-ldflags "-X main.version=1.1.0"`, the Go banner prints `1.1.0` but the browser badge still shows `v0.1.0`.
- No `/api/version` endpoint exists in the mux registration in `server.go`.

**Phase to address:** Phase that implements ldflags version injection and badge update.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Leave `const banner` and duplicate the version as a `var` only for the badge | Minimal diff | Two sources of truth for version string; banner and badge can drift | Never — single `var version` that feeds both is trivial |
| Use English subcategory names in auditpol for subset of "safe" policies | Less code change | Breaks on any non-English Windows; the bug this milestone exists to fix | Never for this milestone |
| Return all auditpol GUID failures as a single concatenated string | Simpler error accumulation | UI error message is unreadable on partial failure | Acceptable only if descriptions are preserved alongside GUIDs |
| Patch the version string directly into `index.html` at build time via `sed` / `go generate` | No API endpoint needed | Requires build tooling; breaks `go build` reproducibility; platform-dependent | Acceptable as a temporary workaround only if API endpoint is deferred |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| `auditpol.exe` via `exec.Command` | Pass GUID with surrounding braces inside the flag value string: `/subcategory:{GUID}` — forgetting the braces causes "The parameter is incorrect" | Use `/subcategory:{0CCE9215-69AE-11D9-BED3-505054503030}` including the `{` and `}` chars, as shown in Microsoft's own documentation examples |
| `go build -ldflags` on Windows | Use Unix-style single-quote wrapping which cmd.exe ignores | Use `-ldflags "-X main.version=1.1.0"` with no interior quotes when the value has no spaces; use a Makefile variable to avoid per-developer shell confusion |
| `embed.FS` + HTML served at runtime | Attempt to template-process `embed.FS` files as Go `html/template` at serve time | Serve via `http.FileServer(http.FS(staticFiles))` for static assets; add a dedicated API endpoint for dynamic values like version |
| Preparation tab error surfacing | Swallow the raw `cmd.CombinedOutput()` exit code and show only "Failed" | Capture and trim the auditpol stderr/stdout and include it with the description and GUID in the `Result.Message` field |

---

## Performance Traps

Not applicable at this project's scale. The tool runs locally on a single Windows machine; audit policy configuration is a one-time setup operation. No performance considerations apply to the v1.1 scope.

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Exposing a `/api/version` endpoint with no auth when password protection is enabled | Leaks tool version to unauthenticated users if the server is bound to `0.0.0.0` | The version endpoint carries no sensitive data; this is an acceptable, intentional information disclosure. Document it, do not add auth to this endpoint. |
| Injecting a version string containing shell metacharacters via ldflags | If the version string is later interpolated into a shell command (e.g., an error message passed to PowerShell), it could inject commands | Keep version strings to semver format only: `v1.1.0`. Validate in the API handler before returning. |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Badge shows `v0.1.0` after upgrade | Consultant on client site cannot confirm which binary version they are running; undermines trust | Dynamic badge via `/api/version` fetch; version var injected at build time |
| Preparation tab shows raw GUID on auditpol failure | `{0CCE9215-69AE-11D9-BED3-505054503030}: exit status 1` is unactionable for a security consultant | Always pair GUID with human-readable subcategory description in error output |
| Partial auditpol failure reported as total failure | User sees `Success: false` and believes all audit policies failed, stops troubleshooting | Current code already accumulates failures into a list; surface the count: "2 of 12 subcategories failed: ..." |
| Technique count on dashboard hardcoded or stale | Dashboard shows a number that contradicts the playbook count shown in the Playbooks tab | Drive the displayed count from the live `registry.Techniques` length via `/api/status` or a dedicated stat endpoint; never hardcode it in HTML |

---

## "Looks Done But Isn't" Checklist

- [ ] **ldflags injection:** Banner in terminal prints new version AND badge in browser shows new version — both must change. Verify both after each build.
- [ ] **GUID migration:** Run `ConfigureAuditPolicy` on a non-English Windows VM (German, French) and confirm all 12 subcategories succeed. English-only testing misses the bug this feature exists to fix.
- [ ] **GUID migration:** Run on Windows Server 2016 (oldest supported) to confirm no GUID is absent on that version.
- [ ] **Partial failure reporting:** Deliberately break one GUID (change one digit) and confirm the UI shows which subcategory failed, not just `Success: false`.
- [ ] **UI polish:** After `index.html` edits, verify all six tab pages render correctly (Dashboard, Preparation, Playbooks, Configure & Run, Results, Log Viewer) — CSS changes in the `:root` block affect every page simultaneously.
- [ ] **CSS variable change:** If any `--bg`, `--accent`, or `--border` token is adjusted, verify the `.version-badge`, `.phase-badge`, `.stat-box`, `.card`, and `.running-indicator` components all still render as intended — all inherit from those tokens.
- [ ] **JavaScript polling continuity:** After any change to `index.html`, confirm the `/api/status` polling loop still fires on the dashboard and updates phase/count stats in real time.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| ldflags targeting a `const` (version not injected) | LOW | Declare `var version = "dev"`, rebuild, re-ship binary |
| Shell quoting inserts literal quotes into version string | LOW | Adjust Makefile quoting, rebuild |
| Wrong GUID causes auditpol failure on client machine | MEDIUM | Consultant must re-run Preparation after a patched binary is provided; client audit policy is left partially configured until then |
| CSS regression breaks a tab after UI polish | LOW | Revert the specific CSS rule in `index.html`; because the file is embedded, a full binary rebuild and re-deployment is required — no hot-patch possible |
| Version badge never updates because `/api/version` endpoint was forgotten | LOW | Add endpoint, rebuild |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| `const` banner not injectable via ldflags | Phase: ldflags version injection | `go build -ldflags "-X main.version=1.1.0" && ./lognojutsu.exe` — banner must print `1.1.0` |
| Windows shell quoting breaks ldflags | Phase: ldflags version injection | Build from both cmd.exe and PowerShell; assert version string has no extraneous quote chars |
| Embedded HTML badge unreachable by ldflags | Phase: ldflags version injection + badge update | Browser badge must match Go banner version after build |
| auditpol partial failure loses human-readable label | Phase: GUID migration | Trigger deliberate failure; inspect `Result.Message` for subcategory name alongside GUID |
| GUID absent on older/non-English Windows | Phase: GUID migration | Test on German Windows VM and Windows Server 2016 before merging |
| CSS regression after UI polish | Phase: UI polish | Manual smoke-test all 6 tab pages after each CSS change |
| Technique count stale in dashboard | Phase: UI polish / labels | Load Playbooks tab (57 techniques); confirm Dashboard count matches |

---

## Sources

- [Go issue #20649 — go build should error if ldflags -X pointed at const](https://github.com/golang/go/issues/20649)
- [Go issue #16743 — ldflags -X does not allow values with spaces](https://github.com/golang/go/issues/16743)
- [Go issue #47072 — -X on a constant, type, or func doesn't report an error](https://github.com/golang/go/issues/47072)
- [Go issue #64246 — can't override variable with ldflags when initial value is a func](https://github.com/golang/go/issues/64246)
- [DigitalOcean — Using ldflags to Set Version Information for Go Applications](https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications)
- [blog.kowalczyk.info — Embedding build number in Go executable](https://blog.kowalczyk.info/article/vEja/embedding-build-number-in-go-executable.html)
- [Microsoft Learn — auditpol set](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/auditpol-set)
- [Microsoft Learn — Event 4719 (System audit policy was changed)](https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-10/security/threat-protection/auditing/event-4719) — source for subcategory GUID table
- [Ansible community.windows issue #14 — win_audit_policy_system does not work in non-English environment](https://github.com/ansible-collections/community.windows/issues/14)
- [Go issue #43179 — flags ending with "=." not correctly parsed via PowerShell](https://github.com/golang/go/issues/43179)

---
*Pitfalls research for: LogNoJutsu v1.1 — ldflags version injection, auditpol GUID migration, SPA UI polish*
*Researched: 2026-03-26*
