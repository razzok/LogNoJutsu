# Phase 14: Safety Audit - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-09
**Phase:** 14-safety-audit
**Areas discussed:** Tier classification, Destructive technique fixes, Cleanup reliability, Audit output format

---

## Tier Classification

### Where should the tier label live?

| Option | Description | Selected |
|--------|-------------|----------|
| YAML field per technique | Add `tier: 1\|2\|3` to each YAML. Classification travels with definition. | ✓ |
| External manifest file | Single file maps technique IDs to tiers. Keeps YAMLs unchanged. | |
| Both | YAML field as source of truth, plus generated summary document. | |

**User's choice:** YAML field per technique
**Notes:** None

### How do you define the tier boundaries?

| Option | Description | Selected |
|--------|-------------|----------|
| By event realism | Tier 1 = real events. Tier 2 = some real events + shortcuts. Tier 3 = echo/stub. | ✓ |
| By risk level | Tier 1 = safe. Tier 2 = writes but cleans up. Tier 3 = potentially destructive. | |
| By both axes | Two dimensions: realism AND risk. | |

**User's choice:** By event realism
**Notes:** None

### Should the tier field be visible in the HTML report and/or web UI?

| Option | Description | Selected |
|--------|-------------|----------|
| HTML report only | Tier column/badge in report. Web UI unchanged. | |
| Both report and UI | Tier in HTML report AND as badge in web UI technique list. | ✓ |
| Neither — data only | Tier exists in YAML only. Display is future work. | |

**User's choice:** Both report and UI
**Notes:** None

### Should the tier assignment produce a rationale per technique?

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — short rationale per technique | 1-line rationale per technique in classification document. | ✓ |
| No — just the tier number | Tier definitions are clear enough without per-technique rationale. | |

**User's choice:** Yes — short rationale per technique
**Notes:** None

---

## Destructive Technique Fixes

### What's the strategy for making destructive techniques safe?

| Option | Description | Selected |
|--------|-------------|----------|
| Scope-limit the real actions | Keep real tools but reduce blast radius. T1490 targets only LNJ artifacts. T1070.001 clears only test log. | ✓ |
| Dry-run with real tool paths | Run real binaries with non-destructive flags. Generates EID 4688 but not destructive events. | |
| Keep destructive but warn | Leave as-is. Add warning/confirmation UI. Phase 16 handles UX. | |

**User's choice:** Scope-limit the real actions
**Notes:** None

### For T1070.001 (log clearing), what's acceptable?

| Option | Description | Selected |
|--------|-------------|----------|
| Clear only a custom test log | Create and clear LogNoJutsu-specific event log channel. Generates EID 104 without wiping real logs. | ✓ |
| Clear only Application log | Application is lowest-impact. Security/System stay intact. | |
| Keep full clearing with pre-export | Export logs to .evtx backup before clearing, restore after. | |

**User's choice:** Clear only a custom test log
**Notes:** None

### For T1490 (inhibit recovery): scope?

| Option | Description | Selected |
|--------|-------------|----------|
| bcdedit + registry only, skip VSS delete | Keep reversible steps (bcdedit, registry). Skip vssadmin/wmic shadow delete. | ✓ |
| All steps but with LNJ-only targets | Create test shadow copy first, delete only that one. | |
| Echo-only simulation | Replace all commands with Write-Host. Becomes Tier 3. | |

**User's choice:** bcdedit + registry only, skip VSS delete
**Notes:** None

### For T1546.003 (WMI persistence): safe enough?

| Option | Description | Selected |
|--------|-------------|----------|
| Safe as-is, just verify cleanup | Harmless trigger, benign action, cleanup removes all CIM objects. | ✓ |
| Add a timeout/auto-expire | Make WMI filter expire after 5 minutes. | |

**User's choice:** Safe as-is, just verify cleanup
**Notes:** None

---

## Cleanup Reliability

### How should cleanup be guaranteed on interrupt/failure?

| Option | Description | Selected |
|--------|-------------|----------|
| Defer-style in executor | Wrap execution so cleanup runs in defer/finally within RunWithCleanup. | ✓ |
| Cleanup registry in engine | Engine tracks techniques and calls RunCleanupOnly on Stop/SIGINT. | |
| Both — belt and suspenders | Executor defer + engine-level safety net. | |

**User's choice:** Defer-style in executor
**Notes:** None

### What about techniques with empty cleanup that write artifacts?

| Option | Description | Selected |
|--------|-------------|----------|
| Audit and add cleanup where needed | Review all 58. Add cleanup for writers with empty cleanup. Read-only stays empty. | ✓ |
| Require cleanup for all techniques | Every technique gets non-empty cleanup, even if just a no-op message. | |
| Flag missing cleanup as tech debt | Document gaps but don't fix. Focus on destructive three only. | |

**User's choice:** Audit and add cleanup where needed
**Notes:** None

---

## Audit Output Format

### What format should the classification document be?

| Option | Description | Selected |
|--------|-------------|----------|
| Markdown table in docs/ | docs/TECHNIQUE-CLASSIFICATION.md with table columns. Human-readable, version-controlled. | ✓ |
| YAML manifest | classification-manifest.yaml. Machine-queryable. | |
| Generated from YAML fields | No separate doc — generate from `tier` fields via script. | |

**User's choice:** Markdown table in docs/
**Notes:** None

### Who is the primary consumer?

| Option | Description | Selected |
|--------|-------------|----------|
| Security consultant | Person running LogNoJutsu at client site. Needs quick realism assessment. | ✓ |
| Developer maintaining techniques | Person adding/modifying techniques. Needs classification criteria. | |
| Both equally | Serves both audiences with criteria section + quick-reference table. | |

**User's choice:** Security consultant
**Notes:** None

---

## Claude's Discretion

- Order of technique auditing
- Exact wording of per-technique rationales
- Whether to add `writes_artifacts` YAML field or derive from cleanup presence
- Custom event log channel naming for T1070.001

## Deferred Ideas

None — discussion stayed within phase scope
