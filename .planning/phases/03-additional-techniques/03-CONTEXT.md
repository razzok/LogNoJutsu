# Phase 3: Additional Techniques - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Expand the technique library with exactly 5 new MITRE ATT&CK techniques (targeting Collection and Command & Control tactic gaps) and 4 new Exabeam UEBA scenarios. All new techniques must have `expected_events` populated in the structured EventSpec format established in Phase 1. README updated with new rows in the existing technique table. This phase does NOT change the execution engine, verification logic, HTML report structure, or SIEM coverage columns (those are Phases 4-5).

</domain>

<decisions>
## Implementation Decisions

### ATT&CK Technique Selection
- **D-01:** Add exactly 5 new MITRE ATT&CK techniques — minimum to satisfy TECH-01 success criteria.
- **D-02:** Focus on filling tactic coverage gaps: **Collection** (T1005 Data from Local System, T1560 Archive Collected Data) and **Command & Control** (T1071 Application Layer Protocol, T1095 Non-Application Layer Protocol). Fifth technique: Claude's discretion — choose from remaining gap areas (Initial Access artifacts or additional C2/Collection variant).
- **D-03:** All 5 new techniques must include `expected_events` with structured EventSpec entries (event_id, channel, description) — same format as all existing techniques.

### UEBA Scenario Themes
- **D-04:** Add 4 new Exabeam UEBA scenarios (exceeds the 3-scenario minimum):
  1. **Data staging + exfiltration** — user copies large volumes to staging dir then exfils; triggers Exabeam data exfiltration use case
  2. **Account takeover chain** — failed logins → success → new device + unusual hour; triggers Exabeam account compromise chain
  3. **Privilege escalation chain** — normal user runs admin tools, token impersonation; triggers Exabeam abnormal privilege use case
  4. **Lateral movement + new asset** — first-time access to new internal host via SMB/RDP; triggers Exabeam new asset access + lateral movement use case
- **D-05:** Each UEBA scenario follows the existing chain YAML pattern (tactic: ueba-scenario, phase: attack/discovery) and includes `expected_events` per the EventSpec format.

### Execution Fidelity
- **D-06:** New techniques follow the multi-variant deep execution style established by existing techniques (e.g., T1057 uses 5+ commands with multiple LOLBin variants). Each technique should use 3–5 commands/methods to maximize SIEM signal diversity and demonstrate multiple detection opportunities.
- **D-07:** For C2 simulation (T1071, T1095): use safe, realistic beacon simulation (DNS lookups to known benign domains, HTTP GET to localhost/loopback) — must NOT generate actual outbound C2 traffic. Generate the expected Windows/Sysmon events without real exfiltration.

### README Documentation
- **D-08:** Add new technique rows to the **existing technique table** in the German README — do not restructure the README or add new sections. Keep consistent with current table format.

### Claude's Discretion
- Selection of the 5th ATT&CK technique (from remaining gap areas — Initial Access artifacts, additional C2 variant, or another tactic with no current coverage)
- Specific T-IDs for C2 techniques (T1071.001 Web Protocols, T1071.004 DNS, T1095 Non-App Layer are all candidates — pick the most SIEM-detectable variants)
- Exact commands used in Collection techniques (T1005, T1560) — must simulate file enumeration/staging artifacts safely, no real sensitive data copied
- German README table column values for new techniques — match existing row format exactly

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Planning
- `.planning/REQUIREMENTS.md` — TECH-01, TECH-02, TECH-03 are the requirements this phase must satisfy
- `.planning/ROADMAP.md` — Phase 3 success criteria (4 items) define the acceptance bar

### Codebase — Technique Format
- `internal/playbooks/embedded/techniques/T1057_process_discovery.yaml` — canonical example of deep multi-variant technique with full expected_events
- `internal/playbooks/embedded/techniques/UEBA_offhours_activity.yaml` — canonical example of UEBA scenario YAML format
- `internal/playbooks/embedded/techniques/UEBA_credential_spray_chain.yaml` — canonical example of UEBA chain scenario
- `internal/playbooks/types.go` — Technique struct with EventSpec type definition (event_id, channel, description fields)

### Prior Phase Context
- `.planning/phases/01-events-manifest-verification-engine/01-CONTEXT.md` — D-01 through D-03 define the EventSpec format all new techniques must follow

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- All 40 existing YAML technique files as templates — copy structure, replace content. Most relevant: T1057 (multi-variant Discovery), T1041 (Exfiltration), T1021_002 (SMB lateral movement), UEBA chain files.
- `internal/playbooks/loader.go` — auto-discovers and loads all YAML files from embedded/techniques/ — no registration required, just add the YAML file.
- `internal/playbooks/embedded/` — embed directive picks up new files automatically at build time.

### Established Patterns
- YAML structure: id, name, description, tactic, technique_id, platform, phase, elevation_required, expected_events[], tags[], executor.type/command, cleanup
- Tags in use: discovery, no-prereqs, ueba-baseline, ueba, exabeam, lateral-movement, credential-access, persistence, defense-evasion, impact, execution
- Phase values: "preparation", "discovery", "attack" — new Collection/C2 techniques likely phase: "attack"
- elevation_required: false for discovery; true for techniques needing local admin (token impersonation, service installs)

### Integration Points
- Drop new YAML files into `internal/playbooks/embedded/techniques/` — they are automatically embedded and loaded
- No engine, server, or reporter changes required for this phase
- README.md at repo root — German language, extend existing technique table

</code_context>

<specifics>
## Specific Ideas

- C2 simulation must stay safe: simulate DNS/HTTP beaconing using loopback addresses or known-benign public DNS — the goal is to generate Sysmon network events (EID 3) and DNS query events (EID 22), not actual C2 traffic
- T1560 (Archive Collected Data) pairs well with T1005 (Data from Local System) — implement them as a natural sequence: enumerate → stage → archive
- UEBA "account takeover chain" should generate EID 4625 (failed logon) then EID 4624 (success) events in sequence — the chain pattern is what triggers Exabeam's correlation use case
- The 4 new UEBA scenarios bring the total UEBA count to 7 (3 existing + 4 new), providing good Exabeam scenario coverage

</specifics>

<deferred>
## Deferred Ideas

- Initial Access artifacts (T1566 phishing simulation) — mentioned during selection but not prioritized; consider for a future technique expansion phase
- More Execution variants (T1059_005 VBScript, T1106 Native API) — deferred to future phase
- T1558_001 Golden Ticket, T1556 modify auth process — deferred to future Credential Access expansion

</deferred>

---

*Phase: 03-additional-techniques*
*Context gathered: 2026-03-25*
