# Phase 8: Backend Correctness - Context

**Gathered:** 2026-03-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Go-only changes to three files: replace locale-dependent auditpol subcategory names with stable GUIDs (`preparation.go`), improve failure message content (`preparation.go`), inject version at build time via ldflags (`main.go`), and add a public `/api/info` endpoint (`server.go`).

No HTML or JavaScript changes in this phase — those belong to Phase 9.

</domain>

<decisions>
## Implementation Decisions

### BUG-01: Audit Policy GUID Migration
- **D-01:** Replace all 12 English subcategory name strings in `ConfigureAuditPolicy()` with their stable, locale-independent GUIDs (e.g. `{0CCE9215-69AE-11D9-BED3-505054503030}`). The auditpol call format becomes `/subcategory:{GUID}` instead of `/subcategory:Logon`.
- **D-02:** Keep the `description` field in the policy struct unchanged — it remains the human-readable label used in error messages and success output.
- **D-03 (blocker):** Two GUIDs are disputed between research files — "Audit Policy Change" and "Scheduled Task". Use the most widely cited Microsoft values as a starting point but add a `// VERIFY: auditpol /list /subcategory:* /v` comment on those two entries. These must be confirmed on a Windows machine before merging Phase 8.

### BUG-02: Failure Message Format
- **D-04:** When a subcategory fails, the error message uses the human-readable `description` (not the raw GUID) plus the exit code. Format: `"<description>: failed (exit status N)"`. Example: `"Logon/Logoff events (4624, 4625, 4634): failed (exit status 87)"`.
- **D-05:** The `Result.Message` string format for partial failure remains: `"Partial failure: <entry1>; <entry2>; ..."` — only the individual entry format changes.

### VER-01: Build-time Version Injection
- **D-06:** `const banner` in `main.go` becomes a `var version = "dev"` at package level. The banner itself can stay as a const but the version line uses `fmt.Sprintf` to embed the version at runtime.
- **D-07:** Linker injection target: `-ldflags "-X main.version=v1.1.0"`. The package is `main`, not the full module path.
- **D-08:** Default (no ldflags): version prints as `dev`. This is intentional — local dev builds self-identify as non-release.

### VER-02: `/api/info` Endpoint
- **D-09:** Add `Version string` to `server.Config` struct. Pass `version` var from `main.go` into `server.Config{Version: version, ...}` before calling `server.Start()`.
- **D-10:** Register `GET /api/info` WITHOUT the `authMiddleware` wrapper — version is not sensitive and the badge must load before the user enters a password.
- **D-11:** Response shape: `{"version":"v1.1.0"}` — version string only. No technique count or extra metadata (Phase 9 can use `/api/techniques` for the Dashboard count separately).

### Claude's Discretion
- Internal implementation of the version banner (whether to use `fmt.Sprintf` inline or a helper function) — Claude decides.
- Whether `Server` stores `Version` on `Config` or as a direct field — Config approach is consistent with existing pattern.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope
- `.planning/ROADMAP.md` §Phase 8 — requirements BUG-01, BUG-02, VER-01, VER-02 and success criteria

### Files to modify
- `internal/preparation/preparation.go` — `ConfigureAuditPolicy()` function, `Result` struct
- `cmd/lognojutsu/main.go` — `const banner`, `main()` function, `server.Config{}` instantiation
- `internal/server/server.go` — `Config` struct, `registerRoutes()`, new `handleInfo` handler

### Research
- `.planning/research/STACK.md` — GUID reference table for all 12 subcategories
- `.planning/research/PITFALLS.md` — const trap warning, disputed GUID list, Windows quoting notes

No external specs — requirements fully captured in decisions above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `writeJSON(w, v)` helper in `server.go:114` — use for `/api/info` response, consistent with all other handlers
- `s.authMiddleware(handler)` wrapping pattern in `registerRoutes()` — intentionally NOT used for `/api/info`
- `Result{Step, Success, Message}` struct in `preparation.go:13` — no new fields needed; only message content changes

### Established Patterns
- All API handlers are methods on `*Server` — `handleInfo` follows the same pattern
- `Config` struct passes startup configuration into `Server` — extend with `Version string` field
- Error messages in existing code: `fmt.Sprintf("Failed: %v | %s", err, out)` pattern — Phase 8 normalises audit policy errors to a cleaner format

### Integration Points
- `main.go:30` — `server.Config{Host, Port, Password}` literal → add `Version: version`
- `server.go:71` — `registerRoutes()` → add `mux.HandleFunc("/api/info", s.handleInfo)` (no auth wrapper)
- `preparation.go:84-93` — the `for _, p := range policies` loop → change `/subcategory:`+p.subcategory to `/subcategory:`+p.subcategory (GUID value) and update error format

</code_context>

<specifics>
## Specific Ideas

- The failure error should read naturally: "Logon/Logoff events (4624, 4625, 4634): failed (exit status 87)" — description first, technical detail second.
- Two GUIDs need on-machine verification before shipping — mark them clearly in code comments so the reviewer knows to check them.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 08-backend-correctness*
*Context gathered: 2026-03-26*
