# Phase 3: Additional Techniques - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-25
**Phase:** 03-additional-techniques
**Areas discussed:** ATT&CK technique selection, UEBA scenario themes, Execution fidelity, README documentation

---

## ATT&CK Technique Selection

| Option | Description | Selected |
|--------|-------------|----------|
| Fill coverage gaps | Target missing tactics: Collection (T1005, T1560), C2 (T1071, T1095), Initial Access artifacts (T1566) | ✓ |
| Double down on Discovery | More Discovery techniques — T1201, T1518, T1120 | |
| Credential Access depth | T1558_001 Golden Ticket, T1556, T1528 | |

**User's choice:** Fill coverage gaps
**Count:** Exactly 5 techniques (minimum to satisfy success criteria)
**Focus areas selected:** Collection (T1005, T1560) + Command & Control (T1071, T1095)
**Notes:** Fifth technique left to Claude's discretion from remaining gaps.

---

## UEBA Scenario Themes

| Option | Description | Selected |
|--------|-------------|----------|
| Data staging + exfiltration | Large-volume copy → exfil chain → Exabeam data exfiltration use case | ✓ |
| Account takeover chain | Failed logins → success → new device + unusual hour → Exabeam account compromise | ✓ |
| Privilege escalation chain | Normal user runs admin tools, token impersonation → Exabeam abnormal privilege | ✓ |
| Lateral movement + new asset | First-time access to new host via SMB/RDP → Exabeam new asset access | ✓ |

**User's choice:** All 4 scenarios
**Notes:** Exceeds the 3-scenario minimum — user wants comprehensive UEBA coverage.

---

## Execution Fidelity

| Option | Description | Selected |
|--------|-------------|----------|
| Match existing style | Multi-variant like T1057 — 3–5 commands/LOLBin variants per technique | ✓ |
| Simpler for C2/Collection | Single commands for harder-to-simulate tactics | |
| Minimal — just generate events | Shortest path to expected events | |

**User's choice:** Match existing style
**Notes:** All new techniques should use 3–5 commands for SIEM signal diversity.

---

## README Documentation

| Option | Description | Selected |
|--------|-------------|----------|
| Extend existing technique table | Add rows to existing table — consistent with current structure | ✓ |
| New section per tactic group | New grouped sections for Collection, C2, UEBA | |
| Minimal — just list names | Brief mention only | |

**User's choice:** Extend existing technique table
**Notes:** No structural changes to README — just new rows.

---

## Claude's Discretion

- Selection of the 5th ATT&CK technique
- Specific T-ID sub-technique variants for C2 (T1071.001, T1071.004, T1095)
- Exact PowerShell commands for safe C2/Collection simulation
- German README table column values

## Deferred Ideas

- T1566 Initial Access artifacts (phishing simulation) — future technique expansion
- T1059_005 VBScript, T1106 Native API — future Execution coverage
- T1558_001 Golden Ticket, T1556 — future Credential Access expansion
