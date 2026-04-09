---
phase: 14-safety-audit
plan: 03
subsystem: techniques, reporter, ui
tags: [yaml, playbooks, safety, tier-classification, html-report, web-ui]

# Dependency graph
requires:
  - phase: 14-safety-audit plan 01
    provides: Tier int field on Technique struct and ExecutionResult struct
  - phase: 14-safety-audit plan 02
    provides: destructive technique rewrites complete so tier assignments are accurate

provides:
  - tier: N field in all 58 technique YAML files (1=Realistic, 2=Partial, 3=Stub)
  - docs/TECHNIQUE-CLASSIFICATION.md with rationale for every technique
  - HTML report Tier column (conditional HasTier, T1/T2/T3 badges)
  - Web UI technique table Tier column (colored badges)

affects:
  - internal/playbooks/embedded/techniques/*.yaml (all 58 files)
  - docs/TECHNIQUE-CLASSIFICATION.md (new)
  - internal/reporter/reporter.go
  - internal/server/static/index.html

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Conditional HTML column pattern: HasTier bool computed from results, guards both CSS and table column (mirrors HasCrowdStrike/HasSentinel)"
    - "Tier classification criteria: Tier 1=real tools/APIs, Tier 2=simulation shortcuts, Tier 3=enumeration/stub"

key-files:
  created:
    - docs/TECHNIQUE-CLASSIFICATION.md
  modified:
    - internal/playbooks/embedded/techniques/*.yaml (58 files — tier: N added)
    - internal/reporter/reporter.go
    - internal/server/static/index.html

key-decisions:
  - "30 techniques classified Tier 1 (real attacker tools/APIs), 15 Tier 2 (simulation shortcuts), 13 Tier 3 (enumeration/stub)"
  - "AZURE_kerberoasting classified Tier 1: real System.IdentityModel TGS requests trigger real EID 4769 per SPN on DC"
  - "AZURE_dcsync classified Tier 2: LDAP ACL queries trigger EID 4662 but no actual replication tool (mimikatz/impacket)"
  - "T1070.001 classified Tier 2: custom LogNoJutsu-Test channel approach generates real EID 104 but not clearing real Security/System logs"
  - "HasTier computed from results slice (at least one result.Tier > 0) — matches existing HasCrowdStrike/HasSentinel pattern"
  - "Tier column in web UI placed between Phase and Elevation columns per UI-SPEC.md D-03"

requirements-completed: [SAFE-02, SAFE-03]

# Metrics
duration: 35min
completed: 2026-04-09
---

# Phase 14 Plan 03: Safety Audit — Tier Classification Summary

**Tier 1|2|3 added to all 58 technique YAMLs; consultant-facing classification document created; T1/T2/T3 badges in HTML report and web UI technique table**

## Performance

- **Duration:** ~35 min
- **Started:** 2026-04-09
- **Completed:** 2026-04-09
- **Tasks:** 2
- **Files modified:** 61 (58 YAMLs + reporter.go + index.html + classification doc)

## Accomplishments

- Added `tier: N` field (after `elevation_required:`) to all 58 technique YAML files — 30 Tier 1, 15 Tier 2, 13 Tier 3
- Created `docs/TECHNIQUE-CLASSIFICATION.md` with 58-row table: ID, Name, Tier, Rationale, Has Cleanup, Writes Artifacts
- TestTierClassified passes: all 58 techniques have tier 1, 2, or 3 (was intentionally failing as scaffold gate)
- TestWriteArtifactsHaveCleanup still passes: all 26 write-artifact techniques have cleanup
- Added `HasTier bool` to `htmlData` struct and compute block in `saveHTML()`
- Added `tier1-badge`/`tier2-badge`/`tier3-badge` CSS + conditional Tier column to HTML report template
- Added `.tier-badge .tier1/.tier2/.tier3` CSS classes and Tier column to web UI technique table in index.html
- All existing tests (reporter, server, engine, playbooks, verifier) pass

## Task Commits

1. **Task 1: Add tier field to all 58 technique YAMLs and create classification document** - `4014f1b` (feat)
2. **Task 2: Add tier badges to HTML report and web UI technique table** - `39c9d0e` (feat)

## Files Created/Modified

- `internal/playbooks/embedded/techniques/*.yaml` — 58 files each have `tier: N` added after `elevation_required:` field
- `docs/TECHNIQUE-CLASSIFICATION.md` — new consultant-facing reference with tier rationale for all 58 techniques
- `internal/reporter/reporter.go` — HasTier field, hasTier computation, tier CSS classes, conditional Tier column in HTML template
- `internal/server/static/index.html` — tier-badge CSS classes, Tier column header (8 columns), T1/T2/T3 badge rendering in JS template

## Tier Distribution

| Tier | Count | Description |
|------|-------|-------------|
| Tier 1 (Realistic) | 30 | Real tools/APIs generating authentic SIEM events |
| Tier 2 (Partial) | 15 | Some real events, simulation shortcuts |
| Tier 3 (Stub) | 13 | Enumeration/query-only, not attacker-realistic |

## Deviations from Plan

None — plan executed exactly as written. The test required running from the worktree directory (`D:/Code/LogNoJutsu/.claude/worktrees/agent-a396e550`) not the main repo directory; this is expected worktree behavior.

**Note:** The worktree was behind master by 9 commits (Phase 14 Plan 01/02 changes). Resolved by `git merge master` at the start of execution (fast-forward merge, no conflicts).

## Known Stubs

None. All tier fields are real values (1, 2, or 3). The classification document rationales are substantive. Both HTML report and web UI handle tier=0 (unclassified) as em dash fallback.

---
*Phase: 14-safety-audit*
*Completed: 2026-04-09*
