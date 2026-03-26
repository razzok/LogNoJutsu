# Project Research Summary

**Project:** LogNoJutsu v1.1 — Bug Fixes & UI Polish
**Domain:** Go single-binary SIEM validation tool — build-time versioning and Windows locale fix
**Researched:** 2026-03-26
**Confidence:** HIGH

## Executive Summary

LogNoJutsu v1.1 is a polish-and-correctness milestone on an existing, fully-shipped Go tool. The v1.0 engine (57 ATT&CK techniques, verification engine, HTML report) is complete. This milestone addresses two correctness bugs and four UI polish items — all contained within three files (`cmd/lognojutsu/main.go`, `internal/server/server.go`, `internal/preparation/preparation.go`, and `internal/server/static/index.html`). No new packages, no new dependencies, and no architectural changes are required.

The recommended implementation approach is two parallel tracks: a backend track (GUID migration in `preparation.go` + version plumbing through `main.go` and `server.go`) and a frontend track (UI label cleanup, error surfacing, badge update, technique count in `index.html`). The backend track has a hard dependency constraint — the `/api/info` endpoint must exist before the JS badge fetch can work — but all other items are independent. Build the backend first, then do the HTML in one coherent pass.

The key risks are both in the ldflags version injection: Go silently ignores `-X` when the target is a `const` (not a `var`), and Windows shell quoting differences between cmd.exe and PowerShell can corrupt the injected string. Both are well-documented failure modes with simple preventions. The auditpol GUID migration carries a medium-confidence note: one GUID (`Audit Policy Change`) differs between the STACK.md and PITFALLS.md sources — this must be verified with `auditpol /list /subcategory:* /v` on a test machine before shipping.

## Key Findings

### Recommended Stack

No new dependencies are introduced in v1.1. The two technical mechanisms are both built into existing tooling: Go's linker `-ldflags "-X main.version=..."` (stable since Go 1.5, confirmed for Go 1.26.1) injects the version string at build time, and `auditpol.exe /subcategory:{GUID}` is a Windows built-in that accepts locale-independent GUIDs sourced from Microsoft's MS-GPAC protocol specification.

**Core technologies:**
- `go build -ldflags "-X main.version=..."`: build-time version injection — zero dependency, built into Go toolchain, stable since Go 1.5
- `auditpol.exe /subcategory:{GUID}`: locale-independent audit policy configuration — Windows built-in, GUIDs are protocol constants from MS-GPAC §2.6.1
- `GET /api/info` (new endpoint in `server.go`): bridges Go version var to the embedded static HTML badge — necessary because `embed.FS` content cannot be modified by ldflags

### Expected Features

All six v1.1 items are in scope and fully defined. Four are P1 must-haves, one is P2 (cosmetic, low effort), and one is deferred to v1.2.

**Must have (table stakes):**
- Dynamic version badge via ldflags — every client engagement starts with "what version is this?"; hardcoded `v0.1.0` is wrong after every release
- Actionable prep error messages — current raw exit code display and `alert()` calls are visibly broken UX
- Remove `alert()` anti-pattern — 5 call sites; replace with inline `showNotification()` banner using existing CSS classes
- UI language consistency — ~15 German strings in Scheduler/PoC/Dashboard tabs; migrate to English; pure string changes, no logic involved
- Dashboard technique count accuracy — stat box must read live count from existing `/api/techniques`; currently shows 0 or stale data

**Should have (competitive):**
- Tactic badge colour fix for `command-and-control` and `ueba-scenario` — one-line Go template change per missing tactic; eliminates grey badge gap noted in PROJECT.md

**Defer (v1.2+):**
- Prep step status on page load (`GET /api/prepare/status`) — requires Windows registry/auditpol/service queries; medium-high complexity; deferred to keep v1.1 scope tight
- Version badge with build date and commit hash — expand ldflags to carry date and git hash; low complexity, but deferred for scope discipline
- Language toggle, light mode, persistent prep state — anti-features for this milestone; defer to v2+

### Architecture Approach

The architecture is unchanged for v1.1. The existing single-binary Go tool uses `http.FileServer` to serve an embedded `index.html` as a static file — there is no Go template engine in the static file path. This constraint is important: the version string cannot be injected directly into the embedded HTML via ldflags or template execution. The correct pattern is a new `/api/info` JSON endpoint that the JS badge fetches on `DOMContentLoaded`. All other features are straightforward in-place edits.

**Major components touched:**
1. `cmd/lognojutsu/main.go` — declare `var version = "dev"` (replacing const-embedded string); pass into `server.Config`
2. `internal/server/server.go` — add `Version` field to `Config` and `Server`; add `handleInfo` handler and `GET /api/info` route (~10 lines total)
3. `internal/preparation/preparation.go` — replace 12 English subcategory name strings with GUIDs inline in the existing struct slice; no struct shape change
4. `internal/server/static/index.html` — `loadInfo()` JS fetch for badge; prep error `<div>` elements per step; `setPrepStatus()` signature update; remove `alert()` calls; German string migration; technique count from API

**Build order constraint:** `preparation.go` GUID change and `server.go`/`main.go` version plumbing are independent and can be done first. The `index.html` version badge depends on `/api/info` existing. All other HTML changes are independent.

### Critical Pitfalls

1. **Targeting a `const` with ldflags `-X` silently does nothing** — Go issue #20649 confirms the linker ignores `-X` on constants without error. The `banner` in `main.go` is currently a `const`. Fix: declare `var version = "dev"` separately and interpolate it into the banner with `fmt.Sprintf`. Prevention: verify with `go tool nm ./lognojutsu.exe | grep version` that the symbol exists.

2. **Embedded HTML badge is unreachable by ldflags** — `embed.FS` snapshots files at compile time; no ldflags mechanism can modify embedded blob content. The badge span in `index.html` will always show `v0.1.0` unless driven by a JS fetch to `/api/info`. Prevention: never attempt to put the version into static HTML directly; always use the API endpoint approach.

3. **Windows shell quoting corrupts the injected version string** — cmd.exe ignores single quotes; PowerShell 5.x handles quote nesting inconsistently (Go issue #43179). The binary can end up with `version = "'1.1.0'"` or an empty string. Prevention: use `go build -ldflags "-X main.version=1.1.0"` with no interior quotes for values without spaces; document the exact PowerShell form in the Makefile or build script.

4. **GUID discrepancy for "Audit Policy Change" between research files** — STACK.md lists `{0CCE922F-...}` but PITFALLS.md lists `{0CCE9223-...}` for this subcategory. These differ. The correct GUID must be verified against `auditpol /list /subcategory:* /v` on a real Windows machine before shipping. Using the wrong GUID silently fails with "The parameter is incorrect."

5. **GUID-only error messages are unreadable** — after migrating from English names to GUIDs, the `failures` accumulator will produce `{0CCE9215-...}: exit status 1` which is actionable for no one. Prevention: keep the existing `description` field in the struct slice and format errors as `"Logon (4624/4625/4634) [{guid}]: exit status 1"`.

## Implications for Roadmap

Based on research, the work decomposes naturally into two phases that can each be PR'd and verified independently.

### Phase 1: Backend Correctness

**Rationale:** The auditpol GUID migration is a pure Go change with no UI dependencies; it is the highest-value correctness fix and can be built and manually tested on a Windows machine independently of everything else. The version plumbing (`main.go` + `server.go`) is also pure Go and should be done here so the `/api/info` endpoint exists before Phase 2 touches HTML.

**Delivers:** A binary that works correctly on non-English Windows installations; a `/api/info` endpoint; a banner that prints the build-injected version; a build command pattern documented for release builds.

**Addresses:** Locale-independent auditpol (correctness bug); dynamic version badge backend (table stakes).

**Avoids:** Pitfall 1 (`const` vs `var`), Pitfall 2 (shell quoting — nail down the Windows build command here), Pitfall 3 (GUID stability — verify before merging), Pitfall 4 (human-readable error labels — keep description field).

### Phase 2: UI Polish

**Rationale:** All HTML changes are grouped together because they share the same file (`index.html`) and the same verification workflow — load the tool in a browser and smoke-test all 6 tabs. Grouping avoids multiple rounds of "rebuild binary, re-test all tabs." This phase has a soft dependency on Phase 1 for the version badge (needs `/api/info`), but all other HTML items are fully independent.

**Delivers:** A polished English-only UI; a live version badge; actionable prep error messages displayed inline (no `alert()`); correct technique count on Dashboard; correct tactic badge colours.

**Addresses:** Remove `alert()` anti-pattern; actionable prep error messages; UI language consistency; technique count accuracy; tactic badge colour fix.

**Avoids:** Pitfall 5 (CSS regression — smoke-test all 6 tabs after any CSS change); UX pitfall of partial auditpol failure showing only `Success: false`.

### Phase Ordering Rationale

- Phase 1 before Phase 2 because the `/api/info` endpoint must exist before the JS badge fetch can work. All other Phase 2 items are independent, but doing Go changes first and HTML changes second means Phase 2 can do a single integrated browser test rather than testing with a stub API.
- The GUID verification step (run `auditpol /list /subcategory:* /v` and resolve the `Audit Policy Change` GUID discrepancy) must gate Phase 1 merge — this is the only open research gap and it has a concrete, fast resolution path.
- No phase requires deeper research during planning. Both phases operate on well-understood, existing code patterns with documented failure modes.

### Research Flags

Phases with standard patterns (skip research-phase):
- **Phase 1:** Go ldflags and auditpol GUID patterns are fully documented with official sources. The only open item is GUID verification on a real Windows machine — this is a 5-minute check, not a research task.
- **Phase 2:** Single-file SPA edits against existing CSS and JS patterns. No novel patterns introduced.

No phases require a `/gsd:research-phase` invocation during planning.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Go ldflags from official Go stdlib docs; auditpol GUID syntax from Microsoft Learn and MS-GPAC spec. No new dependencies means no version-compatibility unknowns. |
| Features | HIGH | Research based on direct inspection of the full 1321-line `index.html` and PROJECT.md requirements. All 6 features are concretely scoped with identified line numbers. |
| Architecture | HIGH | Derived entirely from direct codebase inspection. Constraint around `embed.FS` / `http.FileServer` is confirmed from code, not assumption. Integration point table is exact. |
| Pitfalls | HIGH (ldflags/GUID), MEDIUM (GUID table) | ldflags pitfalls sourced from Go issue tracker (authoritative). GUID table has one discrepancy between STACK.md and PITFALLS.md for "Audit Policy Change" — must verify before shipping. |

**Overall confidence:** HIGH

### Gaps to Address

- **"Audit Policy Change" GUID discrepancy:** STACK.md says `{0CCE922F-69AE-11D9-BED3-505054503030}`; PITFALLS.md says `{0CCE9223-69AE-11D9-BED3-505054503030}`. Run `auditpol /list /subcategory:* /v` on any English Windows machine and grep for "Audit Policy Change" to get the canonical GUID. Block Phase 1 merge on this verification.
- **Sysmon version and download URL:** The Sysmon download path in `preparation.go` is not covered in this research. If the Sysmon download URL has changed between v1.0 and v1.1, the InstallSysmon step will fail. Spot-check the existing URL during Phase 1 implementation.
- **Scheduled Task GUID disagreement:** STACK.md lists `{0CCE9227-...}` (same as "Other Object Access Events") for Scheduled Task, while ARCHITECTURE.md and PITFALLS.md list `{0CCE9232-...}` as a distinct GUID. PITFALLS.md's separate GUID is more likely correct. Verify with `auditpol /list /subcategory:* /v`.

## Sources

### Primary (HIGH confidence)
- Microsoft Learn — auditpol set: confirmed `/subcategory:{guid}` syntax is valid
- MS-GPAC §2.6.1 (Microsoft openspecs) — authoritative GUID table for Windows audit subcategories
- Go issue #20649, #47072 — confirmed `const` is silently ignored by ldflags `-X`
- Direct codebase inspection: `cmd/lognojutsu/main.go`, `internal/server/server.go`, `internal/preparation/preparation.go`, `internal/server/static/index.html`

### Secondary (MEDIUM confidence)
- DigitalOcean — "Using ldflags to Set Version Information for Go Applications" — ldflags `-X` pattern and var requirements; verified against Go stdlib docs
- Go issue #16743, #43179 — Windows shell quoting edge cases for ldflags values
- Nielsen Norman Group — error message guidelines (every error should suggest a next step)
- Ansible community.windows issue #14 — real-world confirmation that auditpol subcategory names fail on non-English Windows

### Tertiary (LOW confidence)
- None — no findings from single-source or inferred-only sources drive implementation decisions

---
*Research completed: 2026-03-26*
*Ready for roadmap: yes*
