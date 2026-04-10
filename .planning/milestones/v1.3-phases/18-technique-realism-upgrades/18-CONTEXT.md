# Phase 18: Technique Realism Upgrades - Context

**Gathered:** 2026-04-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Re-audit discovery stub techniques (T1069, T1082, T1083, T1135) to upgrade tier classifications and review expected_events accuracy. Verify that persistence, defense evasion, and C2/exfiltration requirements (TECH-02, TECH-03, TECH-04) are already satisfied by existing Tier 1/2 techniques. No new techniques are added — this is a tier re-audit and expected_events enrichment pass.

</domain>

<decisions>
## Implementation Decisions

### Discovery Stub Upgrades (TECH-01)
- **D-01:** T1057 (Process Discovery) and T1482 (Domain Trust Discovery) are already Tier 1 native Go techniques from Phase 15. No further work needed — TECH-01 is satisfied for these two.
- **D-02:** T1069 (Group Discovery), T1082 (System Info), T1083 (File Discovery), T1135 (Network Share Discovery) are classified Tier 3 but already execute real Windows commands (net localgroup, systeminfo, Get-ChildItem, net view, etc.) generating authentic EID 4688/4104/5140 events. Upgrade strategy is **re-audit tiers** (bump from 3 to 1 or 2 based on event realism assessment), not rewrite to native Go.
- **D-03:** Tier bump includes **expected_events review** — ensure each technique's expected_events list accurately reflects which EIDs each command generates, for more precise SIEM validation.

### Persistence Techniques (TECH-02)
- **D-04:** TECH-02 is **already satisfied** by existing techniques: T1053.005 (Scheduled Task, Tier 1), T1547.001 (Registry Run Key, Tier 1), T1197 (BITS Jobs, Tier 1), T1543.003 (New Service, Tier 1). All have cleanup commands and use the RunWithCleanup defer pattern from Phase 14.

### Defense Evasion (TECH-03)
- **D-05:** TECH-03 is **already satisfied** by existing techniques: T1027 (Encoded Commands, Tier 1), T1036.005 (Masquerading, Tier 1), T1218.011 (Rundll32 LOLBin, Tier 1), T1574.002 (DLL Sideloading, Tier 2). Tier 2 for T1574.002 is appropriate — uses a benign DLL, not a real sideload.

### C2 & Exfiltration (TECH-04)
- **D-06:** TECH-04 is **already satisfied** by existing techniques: T1041 (HTTP Exfil, Tier 2), T1071.001 (HTTP C2, Tier 2), T1071.004 (DNS C2, Tier 2), T1560.001 (Archive/Encoding, Tier 2). Tier 2 is the correct classification — they use loopback/simulation patterns (real protocol, simulated target), which is the right safety tradeoff per Out of Scope constraints.

### Claude's Discretion
- Exact tier assignment for each of the 4 discovery techniques (1 vs 2) based on event realism assessment
- Expected_events enrichment details — which additional EIDs to add or descriptions to update
- Whether to update technique descriptions to better reflect their actual (non-stub) execution

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Techniques to Re-audit (TECH-01)
- `internal/playbooks/embedded/techniques/T1069_group_discovery.yaml` — Currently Tier 3, runs real net localgroup/Get-LocalGroup commands
- `internal/playbooks/embedded/techniques/T1082_system_info_discovery.yaml` — Currently Tier 3, runs real systeminfo/wmic/reg query commands
- `internal/playbooks/embedded/techniques/T1083_file_discovery.yaml` — Currently Tier 3, runs real dir/Get-ChildItem on user profile paths
- `internal/playbooks/embedded/techniques/T1135_network_share_discovery.yaml` — Currently Tier 3, runs real net view/share/Get-SmbShare commands

### Tier Classification Reference
- `docs/TECHNIQUE-CLASSIFICATION.md` — Tier definitions and per-technique rationales (Phase 14 output)
- `.planning/phases/14-safety-audit/14-CONTEXT.md` — Tier boundary definitions (D-02): Tier 1 = real events, Tier 2 = some real + shortcuts, Tier 3 = echo/stub

### Already-Satisfied Techniques (verify, don't modify)
- `internal/playbooks/embedded/techniques/T1053_005_scheduled_task.yaml` — TECH-02, Tier 1
- `internal/playbooks/embedded/techniques/T1547_001_registry_persistence.yaml` — TECH-02, Tier 1
- `internal/playbooks/embedded/techniques/T1197_bits_jobs.yaml` — TECH-02, Tier 1
- `internal/playbooks/embedded/techniques/T1543_003_new_service.yaml` — TECH-02, Tier 1
- `internal/playbooks/embedded/techniques/T1027_obfuscated_commands.yaml` — TECH-03, Tier 1
- `internal/playbooks/embedded/techniques/T1036_005_masquerading.yaml` — TECH-03, Tier 1
- `internal/playbooks/embedded/techniques/T1218_011_rundll32.yaml` — TECH-03, Tier 1
- `internal/playbooks/embedded/techniques/T1574_002_dll_sideloading.yaml` — TECH-03, Tier 2
- `internal/playbooks/embedded/techniques/T1041_exfiltration_http.yaml` — TECH-04, Tier 2
- `internal/playbooks/embedded/techniques/T1071_001_web_protocols.yaml` — TECH-04, Tier 2
- `internal/playbooks/embedded/techniques/T1071_004_dns.yaml` — TECH-04, Tier 2
- `internal/playbooks/embedded/techniques/T1560_001_archive_collected_data.yaml` — TECH-04, Tier 2

### Requirements
- `.planning/REQUIREMENTS.md` §Technique Realism Upgrades — TECH-01, TECH-02, TECH-03, TECH-04

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `tier:` field in all YAML technique files — established in Phase 14, just needs value update
- `expected_events` array in YAML with `event_id`, `channel`, `description` — structured format for enrichment
- `docs/TECHNIQUE-CLASSIFICATION.md` — must be updated when tiers change

### Established Patterns
- Tier boundaries from Phase 14 D-02: Tier 1 = generates real Windows events a SIEM would see in actual attack; Tier 2 = some real events but uses simulation shortcuts; Tier 3 = echo/stub
- YAML technique structure: id, name, description, tactic, technique_id, platform, phase, elevation_required, tier, expected_events, tags, executor, cleanup

### Integration Points
- 4 YAML files to update (tier + expected_events)
- `docs/TECHNIQUE-CLASSIFICATION.md` — update tier and rationale for 4 techniques
- No code changes needed — this is a YAML-only metadata update

</code_context>

<specifics>
## Specific Ideas

- The 4 discovery techniques are misclassified — they run real commands generating authentic events, making them Tier 1 or 2 by Phase 14's own tier definitions
- Expected_events review should verify Sysmon EID 1 (Process Create) coverage alongside Security EID 4688, since both detect the same tool invocations via different log sources

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 18-technique-realism-upgrades*
*Context gathered: 2026-04-10*
