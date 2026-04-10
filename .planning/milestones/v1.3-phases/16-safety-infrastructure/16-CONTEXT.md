# Phase 16: Safety Infrastructure - Context

**Gathered:** 2026-04-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Add runtime safety checks to the technique execution engine: AMSI block detection for PowerShell techniques, admin elevation gating for privileged techniques, and a scan confirmation modal in the web UI for network scanning techniques. No new techniques are added — this phase hardens the execution pipeline before Phases 17 and 18 add realistic and network-scanning techniques.

</domain>

<decisions>
## Implementation Decisions

### AMSI Detection
- **D-01:** Detect AMSI blocks by parsing PowerShell stderr for known AMSI error patterns (`ScriptContainedMaliciousContent`, exit code -196608). No external dependencies.
- **D-02:** AMSI detection applies to PowerShell executor type only — CMD and native Go techniques do not go through AMSI.
- **D-03:** AMSI-blocked techniques are classified with a distinct `amsi_blocked` verification status and moved on. No retry, no bypass. Consultant sees which techniques were blocked and adjusts Defender policy themselves.

### Elevation Gating
- **D-04:** Check admin status once at engine start using Windows token check (`golang.org/x/sys/windows` — check process token for admin group membership). No shell spawning.
- **D-05:** Per-technique runtime gating — when the engine encounters a technique with `elevation_required: true` and the process is not admin, it skips that technique with an `elevation_required` status. The `elevation_required` field already exists in `Technique` struct and YAML files.
- **D-06:** Elevation-skipped techniques count toward the total technique count in reports with a distinct "Elevation Required" status. Consultant sees "45/58 passed, 8 skipped (elevation), 5 failed" — full visibility.

### Scan Confirmation UX
- **D-07:** Scan confirmation appears as a web UI modal before scan techniques execute. Shows target subnet, rate limit, IDS warning, and the list of scan techniques that will run. Scan does not proceed until the consultant clicks "Confirm".
- **D-08:** Confirmation is triggered by a tag-based mechanism — a `requires_confirmation` YAML field (or tag). Any technique with this flag triggers the confirmation flow. Future-proof for other risky techniques beyond network scans.
- **D-09:** The modal displays four pieces of information: (1) auto-detected target /24 subnet, (2) rate limit notice (connections/second), (3) IDS/IPS warning, (4) list of specific scan techniques that will run.

### Verification Statuses
- **D-10:** Add `amsi_blocked` and `elevation_required` to the existing `VerificationStatus` enum in `types.go`. Existing API consumers see new values in the same `verification_status` field.
- **D-11:** HTML report displays new statuses as color-coded badges: "AMSI Blocked" = orange/amber, "Elevation Required" = gray/blue. Follows existing pass=green, fail=red badge pattern.
- **D-12:** Web UI technique list also displays the new status badges, consistent with the HTML report styling.

### Claude's Discretion
- Exact AMSI error string patterns to match (may need to cover multiple Windows/Defender versions)
- Windows token check implementation details (which specific SID/group to check)
- Scan confirmation API endpoint design (how UI sends confirmation back to engine)
- Rate limit default value and whether it's configurable
- Modal styling and layout within existing web UI patterns
- Whether `requires_confirmation` is a bool field or a tags array entry

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Executor & Technique Model
- `internal/executor/executor.go` — Current dispatch logic; AMSI detection goes in the PowerShell result handling path
- `internal/playbooks/types.go` — `Technique` struct (has `ElevationRequired` and `VerificationStatus`); new status values added here

### Engine
- `internal/engine/engine.go` — Simulation loop; admin check at start, per-technique elevation gating, scan confirmation pause point

### Report & UI
- `internal/reporter/reporter.go` — HTML report template; new status badges go here
- `internal/server/static/index.html` — Web UI; scan confirmation modal + status badges
- `internal/server/server.go` — API handlers; scan confirmation endpoint

### Phase 14 Context (tier system)
- `.planning/phases/14-safety-audit/14-CONTEXT.md` — Tier classification and badge patterns (D-03, D-11)

### Phase 15 Context (native Go dispatch)
- `.planning/phases/15-native-go-architecture/15-CONTEXT.md` — Native Go executor dispatch (D-10, D-11, D-12)

### Requirements
- `.planning/REQUIREMENTS.md` Safety Infrastructure — INFRA-01, INFRA-02, INFRA-03

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `VerificationStatus` type with `VerifNotRun`, `VerifPass`, `VerifFail`, `VerifNotExecuted` — extend with `VerifAMSIBlocked`, `VerifElevationRequired`
- `ElevationRequired bool` field already in `Technique` struct and populated in YAML files — just needs runtime enforcement
- `ExcludedTactics`/`IncludedTactics` filtering pattern in engine — elevation skip follows similar conditional logic
- HTML report badge/badge-color pattern (tier badges from Phase 14) — reuse for new status badges
- Web UI modal patterns in `index.html` — existing preparation tab error panels for reference

### Established Patterns
- `VerificationStatus` string constants with typed enum
- Conditional HTML report columns (CS/Sentinel appear only when relevant)
- `simlog.Info()` / `simlog.TechStart()` / `simlog.TechEnd()` logging for technique lifecycle events
- `RunWithCleanup()` defer pattern — elevation check happens before this is called

### Integration Points
- `internal/executor/executor.go` — AMSI detection in PowerShell result path (after `runCommand` returns)
- `internal/engine/engine.go` — Admin check at startup, per-technique elevation skip before `executor.Run()`
- `internal/server/server.go` — New `/api/scan/confirm` endpoint (or similar) for scan confirmation flow
- `internal/server/static/index.html` — Scan confirmation modal + new status badge colors
- `internal/reporter/reporter.go` — Badge rendering for new statuses in HTML template

</code_context>

<specifics>
## Specific Ideas

- A consultant running non-admin should complete a full simulation without cryptic error messages — "Elevation Required" is actionable, "Access Denied" is not
- AMSI detection must distinguish "technique was blocked by Defender" from "technique had a legitimate PowerShell error" — the error patterns should be specific enough to avoid false positives
- The scan confirmation modal should feel like a safety checkpoint, not a nag — it runs once per simulation, not per technique
- `requires_confirmation` as a YAML field future-proofs the pattern for any technique that needs explicit user consent (e.g., if destructive techniques ever need confirmation)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 16-safety-infrastructure*
*Context gathered: 2026-04-09*
