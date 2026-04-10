# Phase 14: Safety Audit - Context

**Gathered:** 2026-04-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Audit all 58 existing techniques for safety, classify each by realism tier, fix destructive techniques, and ensure verified cleanup paths. No new techniques are added — this is purely an audit and remediation of the existing library.

</domain>

<decisions>
## Implementation Decisions

### Tier Classification
- **D-01:** Add a `tier: 1|2|3` field to each technique YAML file. This is the source of truth for classification.
- **D-02:** Tier boundaries are defined by **event realism**: Tier 1 = generates the real Windows events a SIEM would see in an actual attack. Tier 2 = generates some real events but uses simulation shortcuts. Tier 3 = echo/stub that only proves the technique runs.
- **D-03:** Tier is visible in **both** the HTML report and web UI (badge/label in technique list).
- **D-04:** Each technique gets a 1-line rationale explaining its tier assignment. Rationales live in the classification document (docs/TECHNIQUE-CLASSIFICATION.md).

### Destructive Technique Fixes
- **D-05:** Strategy is **scope-limiting real actions** — keep real tool invocations but reduce blast radius so no permanent damage occurs while still generating the target event IDs.
- **D-06:** **T1070.001 (Clear Logs):** Replace full log clearing with creating and clearing a LogNoJutsu-specific custom event log channel. Generates EID 104 without wiping Security/Application/System logs.
- **D-07:** **T1490 (Inhibit Recovery):** Keep bcdedit recoveryenabled and registry disable steps (easily reversible). Skip vssadmin/wmic shadow delete entirely. Still generates EID 4688 for bcdedit.exe and Sysmon EID 13 for registry changes.
- **D-08:** **T1546.003 (WMI Persistence):** Safe as-is — harmless trigger (uptime check), benign action (whoami to temp file), cleanup removes all CIM objects. Just needs cleanup reliability guarantee from D-10.

### Cleanup Reliability
- **D-09:** Cleanup guarantee via **defer-style pattern in executor** — wrap technique execution in RunWithCleanup so cleanup runs even if the technique body panics or context is cancelled. Minimal change to existing code structure.
- **D-10:** **Audit all 58 techniques** for missing cleanup. If a technique writes to disk, registry, or scheduled tasks and has empty cleanup, add the appropriate cleanup command. Read-only/discovery techniques stay with empty cleanup (legitimate).

### Audit Output Format
- **D-11:** Classification document is a **Markdown table in docs/TECHNIQUE-CLASSIFICATION.md** — columns: Technique ID, Name, Tier, Rationale, Has Cleanup, Writes Artifacts. Human-readable, version-controlled.
- **D-12:** Primary audience is the **security consultant** running LogNoJutsu at a client site. They need to quickly know which techniques are realistic, which are stubs, and which need admin review.

### Claude's Discretion
- Order of technique auditing (can batch by tactic, phase, or alphabetical)
- Exact wording of per-technique rationales
- Whether to add a `writes_artifacts: bool` field to YAML or derive it from cleanup presence
- Custom event log channel naming convention for T1070.001

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Technique Structure
- `internal/playbooks/types.go` — Technique struct definition; `tier` field must be added here
- `internal/playbooks/loader.go` — YAML loading; must handle new `tier` field

### Executor & Cleanup
- `internal/executor/executor.go` — RunWithCleanup() and RunCleanupOnly() functions; defer-style cleanup changes go here

### Destructive Techniques (must be rewritten)
- `internal/playbooks/embedded/techniques/T1490_inhibit_recovery.yaml` — Currently deletes VSS shadows; must be scope-limited
- `internal/playbooks/embedded/techniques/T1070_001_clear_logs.yaml` — Currently clears all logs with no cleanup; must use custom log channel
- `internal/playbooks/embedded/techniques/T1546_003_wmi_event_subscription.yaml` — Safe as-is; verify cleanup path

### Report & UI
- `internal/reporter/reporter.go` — HTML report template; tier column/badge must be added
- `internal/server/static/index.html` — Web UI; tier badge in technique list

### Requirements
- `.planning/REQUIREMENTS.md` §Safety & Audit — SAFE-01, SAFE-02, SAFE-03

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `Technique` struct in `types.go` already has YAML+JSON struct tags pattern — adding `tier` field follows established convention
- `RunWithCleanup()` in `executor.go` already handles cleanup flow — defer pattern is a focused modification
- `RunCleanupOnly()` exists for abort scenarios — can be used as safety net alongside defer
- HTML report template in `reporter.go` already has conditional columns (SIEM coverage) — tier column follows same pattern
- Web UI technique list already displays technique metadata — adding tier badge is incremental

### Established Patterns
- YAML struct tags with aligned columns: `yaml:"tier" json:"tier"`
- Conditional report columns: CS/Sentinel columns appear only when results carry mappings
- `SIEMCoverage map[string][]string` pattern for extensible per-technique metadata
- `ErrorAction SilentlyContinue` / `-ErrorAction Ignore` in PowerShell cleanup commands

### Integration Points
- `internal/playbooks/embedded/techniques/` — all 58 YAML files need `tier:` field added
- `internal/reporter/reporter.go` htmlTemplate — tier column in report table
- `internal/server/static/index.html` — tier badge in technique listing
- `docs/TECHNIQUE-CLASSIFICATION.md` — new file, consultant-facing reference

</code_context>

<specifics>
## Specific Ideas

- T1070.001 should create a custom "LogNoJutsu-Test" event log, write test entries, then clear it — this generates EID 104 authentically without destroying real logs
- T1490 keeps bcdedit and registry steps because they're easily reversible in cleanup, but shadow deletion (irreversible) is removed entirely
- The tier classification document should be scannable — a consultant should find their technique in under 10 seconds

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 14-safety-audit*
*Context gathered: 2026-04-09*
