# Phase 5: Microsoft Sentinel Coverage - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Add Microsoft Sentinel detection rule mappings to existing techniques using the already-present `SIEMCoverage["sentinel"]` map key. Create 3 new `AZURE_` technique YAML files targeting Kerberos/LDAP attack patterns that generate Windows Security Events with known Sentinel analytic rule coverage. Extend the HTML report with a conditional Sentinel coverage column (blue MS badge). Add Sentinel-specific prerequisites/setup documentation to the German README. This phase does NOT modify the verification engine, change existing expected_events, or add any other SIEM platform mappings.

</domain>

<decisions>
## Implementation Decisions

### SIEMCoverage Data Model
- **D-01 (locked from Phase 4):** `SIEMCoverage map[string][]string` is already on the `Technique` struct with YAML key `siem_coverage`. Phase 5 requires zero struct changes — just populate `siem_coverage.sentinel` in YAML files. The `sentinel` map key is already supported by the existing data model.
- **D-02 (locked from Phase 4):** Only map where Sentinel genuinely fires. Techniques with no Sentinel mappings omit `siem_coverage.sentinel` entirely. Keep YAML files clean.

### Azure-Specific Techniques
- **D-03:** Create 3 new YAML technique files with `AZURE_` prefix targeting Kerberos/LDAP attack patterns. These run locally on Windows, generate Windows Security Events, and are picked up by Sentinel's AMA/MMA agents — zero Azure connectivity required at runtime. Targets:
  1. **AZURE_kerberoasting** — Request RC4-encrypted Kerberos service tickets (T1558.003); generates Windows Security Event 4769. Triggers Sentinel "Potential Kerberoasting" analytic rule.
  2. **AZURE_ldap_recon** — LDAP directory enumeration (T1087.002 / T1018); generates Windows Event 1644 (expensive LDAP query). Triggers Sentinel "LDAP Query Reconnaissance" analytic rule.
  3. **AZURE_dcsync** — Simulate DCSync directory replication access (T1003.006); generates Windows Security Event 4662 (DS-Replication-Get-Changes). Triggers Sentinel "Potential DCSync Attack" analytic rule.
- **D-04:** Each AZURE_ technique includes `expected_events` entries for the Windows Security Events generated AND `siem_coverage.sentinel` with the official Sentinel analytic rule names.

### HTML Report Column
- **D-05 (mirrors Phase 4 D-05/D-06):** Sentinel column shows: blue **MS** badge (#0078D4) + list of analytic rule names when `siem_coverage.sentinel` is populated. Grey **N/A** cell when no Sentinel mappings exist for that technique.
- **D-06 (mirrors Phase 4 D-06):** The Sentinel column is **conditional** — renders only when at least one technique in the results has `siem_coverage.sentinel` populated. Matches SENT-03 wording ("when Azure events are present") and keeps the report clean on non-Sentinel environments.
- **D-07:** Badge text is `MS`, background color `#0078D4` (Microsoft blue), white text. CSS classes follow the `cs-badge`/`cs-na`/`cs-list` naming pattern → use `ms-badge`/`ms-na`/`ms-list`. Rendered inside `{{if .HasSentinel}}` conditional — absent HTML has zero vendor-specific markup.

### Detection Rule Naming
- **D-08 (mirrors Phase 4 D-08):** Use **official Microsoft Sentinel analytic rule names** as they appear in the Sentinel portal (e.g., `"Potential Kerberoasting"`, `"LDAP Query Reconnaissance"`, `"Potential DCSync Attack"`). These are the exact strings consultants will see in the client's Sentinel workspace — makes the report directly actionable.

### Claude's Discretion
- Specific official Sentinel analytic rule names to map to each existing technique — research actual Sentinel built-in rule names
- Which existing techniques (out of 52) merit Sentinel mappings — only map where Sentinel genuinely fires
- Exact PowerShell/cmd commands in AZURE_ techniques — must generate the expected Windows Security Events safely (no real attacks), following the multi-variant pattern from Phase 3
- `HasSentinel bool` flag on `htmlData` struct — compute in `SaveResults()` by scanning for any non-empty `SIEMCoverage["sentinel"]` slice, same pattern as `HasCrowdStrike`
- Where to add Sentinel documentation in the German README — extend existing README following Phase 4's documentation pattern

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Planning
- `.planning/REQUIREMENTS.md` — SENT-01, SENT-02, SENT-03 are the requirements this phase must satisfy
- `.planning/ROADMAP.md` — Phase 5 success criteria (4 items) define the acceptance bar

### Codebase — Core Types
- `internal/playbooks/types.go` — Current Technique struct; `SIEMCoverage map[string][]string` already present — no changes needed
- `internal/reporter/reporter.go` — HTML template and htmlData struct; Sentinel column requires `HasSentinel bool` flag and `ms-badge`/`ms-na`/`ms-list` CSS classes

### Codebase — Phase 4 Patterns (mirror for Sentinel)
- `internal/playbooks/embedded/techniques/FALCON_process_injection.yaml` — canonical SIEM-targeted technique format with `siem_coverage` + `expected_events`
- `internal/playbooks/embedded/techniques/FALCON_lsass_access.yaml` — canonical SIEM-targeted technique format
- `internal/reporter/reporter.go` — existing `{{if .HasCrowdStrike}}` conditional column block to mirror for `{{if .HasSentinel}}`

### Prior Phase Context
- `.planning/phases/04-crowdstrike-siem-coverage/04-CONTEXT.md` — D-01 through D-08: all data model, column rendering, and naming decisions that Phase 5 mirrors exactly
- `.planning/phases/01-events-manifest-verification-engine/01-CONTEXT.md` — HTML inline column pattern and badge style reference

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `SIEMCoverage map[string][]string` on `Technique` struct — `sentinel` key already supported, YAML omitempty already set
- `siemCoverage` funcMap helper in reporter — nil-safe map access for `SIEMCoverage["sentinel"]` lookup, reuse directly
- `HasCrowdStrike bool` pattern in `htmlData` + `SaveResults()` — duplicate as `HasSentinel bool` with identical logic but `"sentinel"` key
- AZURE_ technique YAMLs in `internal/playbooks/embedded/techniques/` — picked up automatically by loader at build time (same as FALCON_ files)

### Established Patterns
- SIEM-targeted technique YAML schema: `siem_coverage: sentinel: ["Rule Name 1", "Rule Name 2"]`
- `{{if $.HasSentinel}}` conditional block in HTML template mirrors `{{if $.HasCrowdStrike}}` — add after existing CrowdStrike column
- Multi-variant execution: 3–5 commands per technique for SIEM signal diversity (Phase 3 D-06)

### Integration Points
- `internal/reporter/reporter.go` `saveHTML()` / `htmlData` struct: add `HasSentinel bool`, pass to template
- HTML template table header: add `{{if $.HasSentinel}}<th>Microsoft Sentinel</th>{{end}}` after CrowdStrike header
- HTML template table row: add `{{if $.HasSentinel}}<td>...</td>{{end}}` cell with ms-badge/ms-na styling

</code_context>

<specifics>
## Specific Ideas

- The `sentinel` YAML map key follows the same lowercase convention as `crowdstrike` — `siem_coverage: sentinel: [...]`
- AZURE_kerberoasting can use `Add-Type` P/Invoke to request service tickets with `KerbRetrieveEncodedTicketMessage` — generates genuine 4769 events without real credential theft
- AZURE_ldap_recon can use `[System.DirectoryServices.DirectorySearcher]` with broad filter `(objectClass=*)` and `SizeLimit` to trigger expensive query Event 1644 on domain controllers
- AZURE_dcsync simulates `DS-Replication-Get-Changes` access using directory service object enumeration to generate 4662 events — no real replication, just the access pattern
- The `ms-badge` CSS should use `background-color: #0078D4` (Microsoft blue), `color: white` — visually distinct from the green `cs-badge`
- Column ordering: Verification | CrowdStrike | Sentinel (left to right, chronological by phase)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 05-microsoft-sentinel-coverage*
*Context gathered: 2026-03-25*
