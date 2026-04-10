# Phase 16: Safety Infrastructure - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-09
**Phase:** 16-safety-infrastructure
**Areas discussed:** AMSI detection, Elevation gating, Scan confirmation UX, Verification statuses

---

## AMSI Detection

| Option | Description | Selected |
|--------|-------------|----------|
| Parse stderr for AMSI patterns | Check PowerShell stderr/exit code for known AMSI error strings. Zero dependencies. | ✓ |
| Pre-flight AMSI probe | Test a harmless AMSI trigger string before running techniques. | |
| You decide | Claude picks the best approach. | |

**User's choice:** Parse stderr for AMSI patterns
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| Classify and move on | Mark as "AMSI Blocked" status, no retry. | ✓ |
| Offer optional bypass | Retry with AMSI bypass (e.g., amsi.dll patch). | |
| You decide | Claude picks based on safety philosophy. | |

**User's choice:** Classify and move on
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| PowerShell only | AMSI primarily blocks PowerShell scripts. CMD and Go don't go through AMSI. | ✓ |
| PowerShell + CMD | CMD techniques could theoretically trigger AMSI via script block logging. | |

**User's choice:** PowerShell only
**Notes:** None

---

## Elevation Gating

| Option | Description | Selected |
|--------|-------------|----------|
| Per-technique at runtime | Check admin status once at engine start, skip each elevation_required technique individually. | ✓ |
| Pre-flight batch check | Before simulation starts, list all elevation-gated techniques and warn. | |
| Both | Pre-flight warning + per-technique skip at runtime. | |

**User's choice:** Per-technique at runtime
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| Windows token check | Use golang.org/x/sys/windows to check process token for admin group membership. | ✓ |
| Shell-based check | Run "net session" or similar and check exit code. | |
| You decide | Claude picks the cleanest approach. | |

**User's choice:** Windows token check
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| Count with distinct status | Show in report as "Elevation Required" with full technique count visibility. | ✓ |
| Exclude from totals | Only count techniques that actually ran. | |

**User's choice:** Count with distinct status
**Notes:** None

---

## Scan Confirmation UX

| Option | Description | Selected |
|--------|-------------|----------|
| Web UI modal | Modal dialog in the web UI before scan techniques execute. | ✓ |
| API-level gate | Scan techniques require a separate API call with confirmation token. | |
| Engine-level prompt | Engine pauses and returns pending status; UI polls and shows confirmation. | |

**User's choice:** Web UI modal
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| Tag-based: requires_confirmation | Add a YAML field; any technique with this flag triggers confirmation. Future-proof. | ✓ |
| Hardcoded technique IDs | Only T1046 and T1018 trigger confirmation. Simple but inflexible. | |
| Tactic-based | All "discovery" techniques with network activity trigger confirmation. | |

**User's choice:** Tag-based: requires_confirmation
**Notes:** None

Modal information (multi-select):
- ✓ Target subnet (auto-detected)
- ✓ Rate limit notice
- ✓ IDS warning
- ✓ Technique list

---

## Verification Statuses

| Option | Description | Selected |
|--------|-------------|----------|
| Color-coded badges | Same badge style as tier labels. AMSI Blocked = orange/amber, Elevation Required = gray/blue. | ✓ |
| Separate status column | Add a new "Skip Reason" column that only appears when techniques were skipped. | |
| You decide | Claude picks based on existing report patterns. | |

**User's choice:** Color-coded badges
**Notes:** None

| Option | Description | Selected |
|--------|-------------|----------|
| Extend existing fields | Add "amsi_blocked" and "elevation_required" to the VerificationStatus enum. | ✓ |
| Separate skip_reason field | Add a new field alongside verification_status. | |

**User's choice:** Extend existing fields
**Notes:** None

---

## Claude's Discretion

- Exact AMSI error string patterns to match
- Windows token check implementation details
- Scan confirmation API endpoint design
- Rate limit default value and configurability
- Modal styling and layout
- Whether requires_confirmation is a bool field or tags array entry

## Deferred Ideas

None — discussion stayed within phase scope
