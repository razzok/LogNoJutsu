# Phase 4: CrowdStrike SIEM Coverage - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Add a `siem_coverage` field to the Technique struct and YAML format that maps SIEM platform names to lists of detection rule names. Populate CrowdStrike mappings for existing techniques where applicable. Add 3 new Falcon-targeted technique YAML files (FALCON_ prefix) that specifically generate events triggering Falcon sensor detections. Extend the HTML report with a conditional CrowdStrike coverage column. Add CrowdStrike prerequisites/setup documentation. This phase does NOT modify the verification engine, change existing expected_events, or add Sentinel mappings (Phase 5).

</domain>

<decisions>
## Implementation Decisions

### SIEMCoverage Data Model
- **D-01:** Add `SIEMCoverage map[string][]string` field to the `Technique` struct with YAML key `siem_coverage` and `omitempty`. The map key is the SIEM platform name (`crowdstrike`, `sentinel`), value is a list of detection rule names. This single field handles both Phase 4 (CrowdStrike) and Phase 5 (Sentinel) — no struct change needed in Phase 5.
- **D-02:** Techniques with no SIEM mappings omit the `siem_coverage` field entirely (`omitempty`). Not all 52 existing techniques will have CrowdStrike mappings — only add where there is a genuine, accurate mapping. Keep YAML files clean.

### Falcon-Specific Techniques
- **D-03:** Create 3 new YAML technique files (prefix `FALCON_`) rather than modifying existing techniques. Targets:
  1. **FALCON_process_injection** — CreateRemoteThread / VirtualAllocEx pattern; triggers Falcon's process injection behavioral detection
  2. **FALCON_lsass_access** — LSASS memory access via OpenProcess/ReadProcessMemory; triggers Falcon's credential theft detection
  3. **FALCON_lateral_movement_psexec** — PsExec-style / WMI remote execution pattern; triggers Falcon's lateral movement detection
- **D-04:** Each FALCON_ technique includes `expected_events` (structured EventSpec entries) for the Windows/Sysmon events generated, AND `siem_coverage.crowdstrike` with the official Falcon alert names those behaviors trigger.

### HTML Report Column
- **D-05:** CrowdStrike column shows: green **CS** badge + list of detection rule names when `siem_coverage.crowdstrike` is populated for that technique. Grey **N/A** cell when no CrowdStrike mappings exist for that technique.
- **D-06:** The CrowdStrike column is **conditional** — it only renders when at least one technique in the results has `siem_coverage.crowdstrike` populated. This matches CROW-03 wording ("when Falcon events are present") and keeps the report clean on non-CrowdStrike environments. Consistent with how the Sentinel column will work in Phase 5.
- **D-07:** Column style follows the Phase 1 badge pattern — the CS badge should visually match the existing pass/fail badge style already used in the verification column.

### Detection Rule Naming
- **D-08:** Use **official Falcon alert names** as they appear in the CrowdStrike Falcon console UI (e.g., `"Malicious PowerShell Interpreter"`, `"Suspicious LSASS Access"`, `"Lateral Movement - PsExec"`). These are the exact strings consultants will see in the client's Falcon dashboard — makes the report directly actionable.

### Claude's Discretion
- Specific official Falcon alert names to map to each existing technique — research actual Falcon detection names
- Which existing techniques (out of 52) merit CrowdStrike mappings — only map where Falcon genuinely fires
- Exact PowerShell/cmd commands in FALCON_ techniques — must generate the expected Windows/Sysmon events safely (no real attacks), following the multi-variant pattern from Phase 3 (D-06)
- Where to add CrowdStrike documentation in the README — extend existing German README following Phase 3's documentation pattern

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Planning
- `.planning/REQUIREMENTS.md` — CROW-01, CROW-02, CROW-03 are the requirements this phase must satisfy
- `.planning/ROADMAP.md` — Phase 4 success criteria (4 items) define the acceptance bar

### Codebase — Core Types
- `internal/playbooks/types.go` — Current Technique struct; `SIEMCoverage map[string][]string` field to be added here
- `internal/reporter/reporter.go` — HTML template and htmlData struct; CrowdStrike column requires new conditional rendering logic

### Codebase — Technique Format
- `internal/playbooks/embedded/techniques/T1059_001_powershell.yaml` — canonical example of a technique that should receive CrowdStrike mappings
- `internal/playbooks/embedded/techniques/T1003_001_lsass.yaml` — canonical example for FALCON_lsass_access reference
- `internal/playbooks/embedded/techniques/T1021_002_smb_shares.yaml` — canonical example for FALCON_lateral_movement reference

### Prior Phase Context
- `.planning/phases/01-events-manifest-verification-engine/01-CONTEXT.md` — D-04/D-05: HTML inline column pattern and badge style that CrowdStrike column must match
- `.planning/phases/03-additional-techniques/03-CONTEXT.md` — D-06/D-07: Multi-variant execution style and safe C2/execution simulation rules for new FALCON_ techniques

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `Technique` struct in `internal/playbooks/types.go`: adding `SIEMCoverage map[string][]string` follows the same pattern as existing `InputArgs map[string]string` and `NistControls []string` fields — use `yaml:"siem_coverage,omitempty" json:"siem_coverage,omitempty"`
- Existing badge HTML in the reporter template — CrowdStrike CS badge should reuse the same CSS class pattern as verification pass/fail badges
- `internal/playbooks/loader.go` auto-embed — FALCON_ technique YAMLs placed in `internal/playbooks/embedded/techniques/` are picked up automatically at build time

### Established Patterns
- Struct tags: `yaml:"field" json:"field"` (visually aligned) — follow for SIEMCoverage field
- YAML omitempty: `NistControls []string yaml:"nist_controls,omitempty"` — SIEMCoverage should follow the same omitempty pattern
- Technique YAML structure: id, name, description, tactic, technique_id, platform, phase, elevation_required, expected_events[], tags[], executor — FALCON_ files follow this schema exactly
- Multi-variant execution: 3–5 commands per technique to maximize SIEM signal diversity (Phase 3 D-06)

### Integration Points
- `internal/reporter/reporter.go` `saveHTML()` / `htmlData` struct: add a `HasCrowdStrike bool` flag and pass `SIEMCoverage` data to template for conditional column rendering
- HTML template loop over results: add `{{if $.HasCrowdStrike}}` conditional column header and per-row cell rendering

</code_context>

<specifics>
## Specific Ideas

- The `siem_coverage` YAML key name is lowercase/underscore consistent with all other YAML fields (`technique_id`, `elevation_required`, `nist_controls`, etc.)
- The FALCON_ technique files should have `tactic: "crowdstrike-falcon"` or use the closest matching MITRE tactic — check how the engine filters by tactic to ensure FALCON_ techniques run in the correct phase
- For the HTML conditional column: compute `HasCrowdStrike` in `SaveResults()` by scanning results for any non-empty `SIEMCoverage["crowdstrike"]` slice — same pattern used for `WhatIf bool` flag
- The "N/A" grey cell for techniques without CrowdStrike mappings should use muted/grey styling, distinct from the verification "not_run" state

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 04-crowdstrike-siem-coverage*
*Context gathered: 2026-03-25*
