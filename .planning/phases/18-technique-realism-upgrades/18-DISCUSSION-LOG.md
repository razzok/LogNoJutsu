# Phase 18: Technique Realism Upgrades - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-10
**Phase:** 18-technique-realism-upgrades
**Areas discussed:** Discovery stub upgrades, Persistence techniques, Defense evasion, C2 & exfiltration

---

## Discovery Stub Upgrades (TECH-01)

### Q1: What does upgrading discovery stubs mean?

| Option | Description | Selected |
|--------|-------------|----------|
| Re-audit tiers only | Review T1069/T1082/T1083/T1135 — they run real tools and generate real events. Likely just need tier bump to 1 or 2 and expected_events enrichment. Minimal code changes. | ✓ |
| Convert to native Go | Rewrite these 4 techniques as type:go using Go syscalls/WMI/registry APIs instead of PowerShell. | |
| Both: re-audit + selective Go conversion | Bump tiers for techniques already realistic via PowerShell, convert specific high-value ones to native Go. | |

**User's choice:** Re-audit tiers only
**Notes:** Techniques already execute real Windows commands — the Tier 3 classification is inaccurate.

### Q2: Should re-audit also update expected_events?

| Option | Description | Selected |
|--------|-------------|----------|
| Tier bump + expected_events review | Update tier from 3→1/2 AND review expected_events for accuracy. | ✓ |
| Tier bump only | Just change tier field. Expected_events already look reasonable. | |
| You decide | Claude's discretion. | |

**User's choice:** Tier bump + expected_events review
**Notes:** Makes SIEM validation more precise.

### Q3: Are T1057 and T1482 already satisfied?

| Option | Description | Selected |
|--------|-------------|----------|
| Already satisfied | T1057 and T1482 are Tier 1 native Go techniques from Phase 15. No further work needed. | ✓ |
| Review expected_events only | Keep execution as-is but double-check expected_events. | |

**User's choice:** Already satisfied

---

## Persistence Techniques (TECH-02)

### Q1: Are existing persistence techniques sufficient?

| Option | Description | Selected |
|--------|-------------|----------|
| Already satisfied | All 4 named persistence mechanisms exist at Tier 1 with cleanup. | ✓ |
| Audit cleanup reliability | Verify cleanup runs via RunWithCleanup defer pattern. | |
| Add more persistence techniques | Add additional persistence mechanisms beyond the 4 listed. | |

**User's choice:** Already satisfied
**Notes:** T1053.005, T1547.001, T1197, T1543.003 all Tier 1 with cleanup commands.

---

## Defense Evasion (TECH-03)

### Q1: Are existing defense evasion techniques sufficient?

| Option | Description | Selected |
|--------|-------------|----------|
| Already satisfied | All 4 named techniques exist. T1574.002 Tier 2 is appropriate for safety. | ✓ |
| Review T1574.002 tier | Investigate if DLL sideloading can be upgraded to Tier 1. | |
| Add more evasion techniques | Add beyond the 4 listed. | |

**User's choice:** Already satisfied
**Notes:** T1027 (Tier 1), T1036.005 (Tier 1), T1218.011 (Tier 1), T1574.002 (Tier 2) all exist.

---

## C2 & Exfiltration (TECH-04)

### Q1: Are existing C2/exfil techniques sufficient?

| Option | Description | Selected |
|--------|-------------|----------|
| Already satisfied | All 4 C2/exfil techniques exist at Tier 2. Loopback patterns are inherently Tier 2. | ✓ |
| Review for Tier upgrade | Investigate if any can be upgraded. | |
| Add more C2/exfil techniques | Add beyond the 4 listed. | |

**User's choice:** Already satisfied
**Notes:** Tier 2 is correct — real protocol but simulated target, matching Out of Scope constraints.

---

## Claude's Discretion

- Exact tier assignment for each of the 4 discovery techniques (1 vs 2)
- Expected_events enrichment details
- Whether to update technique descriptions

## Deferred Ideas

None — discussion stayed within phase scope.
