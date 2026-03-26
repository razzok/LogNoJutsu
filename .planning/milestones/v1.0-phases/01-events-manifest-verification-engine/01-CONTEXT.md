# Phase 1: Events Manifest & Verification Engine - Context

**Gathered:** 2026-03-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Add a structured events manifest to each technique and a verification engine that, after each technique executes, waits briefly, queries the local Windows Event Log, and records pass/fail per technique. Verification results appear inline in the HTML report alongside existing execution results. This phase does NOT query SIEM APIs, add new techniques, or refactor package structure.

</domain>

<decisions>
## Implementation Decisions

### Event ID Data Model
- **D-01:** Replace the current free-text `ExpectedEvents []string` with a structured `EventSpec` type containing at minimum: `event_id` (int), `channel` (string), `description` (string). This is the canonical format for both the YAML technique files and querying.
- **D-02:** Whether to add optional keyword filter criteria (e.g., `contains` field for message matching) is left to Claude's discretion — balance precision vs complexity.
- **D-03:** All 39+ existing technique YAML files must be updated to use the new `EventSpec` format, replacing the existing free-text strings.

### Report Presentation
- **D-04:** Verification results appear as an **inline column in the existing per-technique results table** in the HTML report (not a separate section).
- **D-05:** The verification column shows: a pass/fail badge (green/red) + a compact list of each expected event ID with its individual status (✓ EID 10 Sysmon/Operational, ✗ EID 4656 Security).

### Verification Timing
- **D-06:** Verification runs **after each technique**, immediately following execution, with a **configurable wait delay** (default: 3 seconds) to allow Windows Event Log writes to settle before querying.
- **D-07:** Verification result is stored in the `ExecutionResult` for that technique and appears in real-time in the simulation logs.

### "Not Executed" vs "Events Missing" Detection
- **D-08:** Use the existing `ExecutionResult.Success` flag as the gate: if `Success=false` → mark as **"Not Executed"** (tool-side failure); if `Success=true` but no matching events found → mark as **"Events Missing"** (SIEM-side gap). No additional execution detection logic needed.

### Claude's Discretion
- Query mechanism for Windows Event Log (PowerShell `Get-WinEvent`, Go Win32 API via `golang.org/x/sys`, or subprocess — choose based on existing executor patterns)
- Whether `EventSpec` includes an optional `contains` keyword filter for message content matching
- Time window for event log search (how far back to look — recommend 30-60s from technique start)
- New fields to add to `ExecutionResult` to carry verification status
- WhatIf mode behavior (skip verification since techniques don't execute)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Planning
- `.planning/REQUIREMENTS.md` — VERIF-01 through VERIF-05 are the requirements this phase must satisfy
- `.planning/ROADMAP.md` — Phase 1 success criteria (5 items) define the acceptance bar

### Codebase
- `.planning/codebase/STRUCTURE.md` — Package layout, entry points, dependency graph
- `.planning/codebase/CONVENTIONS.md` — Naming conventions, struct tags (yaml+json aligned), error handling patterns
- `.planning/codebase/CONCERNS.md` — Known technical debt and fragile areas to avoid touching unnecessarily

### Key Source Files (must read before touching)
- `internal/playbooks/types.go` — Current `Technique` struct with `ExpectedEvents []string` and `ExecutionResult` struct (both need new fields)
- `internal/engine/engine.go` — Where techniques are executed; verification must hook in here after each technique runs
- `internal/reporter/reporter.go` — HTML template and `htmlData` struct; verification column requires new template data
- `internal/executor/executor.go` — Execution patterns; verification query approach should follow executor's PowerShell/cmd style

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `ExecutionResult` struct in `internal/playbooks/types.go`: already carries `Success`, `StartTime`, `EndTime` — verification should add `VerificationStatus` (enum: not_run, pass, fail, not_executed) and `VerifiedEvents []VerifiedEvent`
- `Technique.ExpectedEvents []string`: already YAML/JSON tagged and serialized — changing to `[]EventSpec` is a clean field replacement
- `executor.go` `runPS()` function: PowerShell execution pattern already established — querying Event Log via PowerShell follows the same pattern

### Established Patterns
- Struct tags: `yaml:"field" json:"field"` (visually aligned with spaces) — follow this for all new fields
- Typed string constants: `PhaseIdle Phase = "idle"` style — use for VerificationStatus enum
- Error handling: sequential `err` / `readErr` / `parseErr` variable names
- Section separator comments: `// ── Section ──────────────────────────────────────────` style in large files

### Integration Points
- `engine.go` `runInternal()` / `runPoC()`: technique execution loops — verification hook goes here after each `executor.RunWithCleanup()` call
- `reporter.go` `saveHTML()` / `htmlData` struct: HTML template data structure — add `VerificationStatus` and `VerifiedEvents` to the result data passed to template
- `index.html` live log stream: if verification results are emitted as simlog entries, they'll appear in the real-time UI automatically

</code_context>

<specifics>
## Specific Ideas

- The 3-second default wait is a starting point — make it a const or config field so it can be adjusted without code changes
- The pass/fail badge in the HTML report should visually match the existing success/failure badge style already used for technique execution results
- Existing free-text `ExpectedEvents` strings are already descriptive (e.g., `"Sysmon 10 (ProcessAccess - TargetImage: lsass.exe, GrantedAccess: 0x1010)"`) — the `description` field of the new `EventSpec` can preserve this text for human readability

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 01-events-manifest-verification-engine*
*Context gathered: 2026-03-24*
