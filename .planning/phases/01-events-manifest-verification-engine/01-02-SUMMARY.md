---
phase: 01-events-manifest-verification-engine
plan: 02
subsystem: playbooks/techniques
tags: [yaml-migration, event-specs, structured-data]
dependency_graph:
  requires: [01-01]
  provides: [structured-expected-events]
  affects: [internal/verifier]
tech_stack:
  added: []
  patterns: [EventSpec-yaml-format]
key_files:
  created: []
  modified:
    - internal/playbooks/embedded/techniques/T1003_001_lsass.yaml
    - internal/playbooks/embedded/techniques/T1003_006_dcsync.yaml
    - internal/playbooks/embedded/techniques/T1016_network_config.yaml
    - internal/playbooks/embedded/techniques/T1021_001_rdp.yaml
    - internal/playbooks/embedded/techniques/T1021_002_smb_shares.yaml
    - internal/playbooks/embedded/techniques/T1027_obfuscated_commands.yaml
    - internal/playbooks/embedded/techniques/T1036_005_masquerading.yaml
    - internal/playbooks/embedded/techniques/T1041_exfiltration_http.yaml
    - internal/playbooks/embedded/techniques/T1046_network_scan.yaml
    - internal/playbooks/embedded/techniques/T1047_wmi_execution.yaml
    - internal/playbooks/embedded/techniques/T1049_network_connections.yaml
    - internal/playbooks/embedded/techniques/T1053_005_scheduled_task.yaml
    - internal/playbooks/embedded/techniques/T1057_process_discovery.yaml
    - internal/playbooks/embedded/techniques/T1059_001_powershell.yaml
    - internal/playbooks/embedded/techniques/T1059_003_cmd_shell.yaml
    - internal/playbooks/embedded/techniques/T1069_group_discovery.yaml
    - internal/playbooks/embedded/techniques/T1070_001_clear_logs.yaml
    - internal/playbooks/embedded/techniques/T1082_system_info_discovery.yaml
    - internal/playbooks/embedded/techniques/T1083_file_discovery.yaml
    - internal/playbooks/embedded/techniques/T1087_account_discovery.yaml
    - internal/playbooks/embedded/techniques/T1098_account_manipulation.yaml
    - internal/playbooks/embedded/techniques/T1110_001_bruteforce.yaml
    - internal/playbooks/embedded/techniques/T1110_003_password_spraying.yaml
    - internal/playbooks/embedded/techniques/T1134_001_token_impersonation.yaml
    - internal/playbooks/embedded/techniques/T1135_network_share_discovery.yaml
    - internal/playbooks/embedded/techniques/T1136_001_create_local_account.yaml
    - internal/playbooks/embedded/techniques/T1197_bits_jobs.yaml
    - internal/playbooks/embedded/techniques/T1218_011_rundll32.yaml
    - internal/playbooks/embedded/techniques/T1482_domain_trust_discovery.yaml
    - internal/playbooks/embedded/techniques/T1486_data_encrypted_for_impact.yaml
    - internal/playbooks/embedded/techniques/T1490_inhibit_recovery.yaml
    - internal/playbooks/embedded/techniques/T1543_003_new_service.yaml
    - internal/playbooks/embedded/techniques/T1546_003_wmi_event_subscription.yaml
    - internal/playbooks/embedded/techniques/T1547_001_registry_persistence.yaml
    - internal/playbooks/embedded/techniques/T1548_002_uac_bypass.yaml
    - internal/playbooks/embedded/techniques/T1550_002_pass_the_hash.yaml
    - internal/playbooks/embedded/techniques/T1552_001_credentials_in_files.yaml
    - internal/playbooks/embedded/techniques/T1558_003_kerberoasting.yaml
    - internal/playbooks/embedded/techniques/T1562_002_disable_logging.yaml
    - internal/playbooks/embedded/techniques/T1574_002_dll_sideloading.yaml
    - internal/playbooks/embedded/techniques/UEBA_credential_spray_chain.yaml
    - internal/playbooks/embedded/techniques/UEBA_lateral_discovery_chain.yaml
    - internal/playbooks/embedded/techniques/UEBA_offhours_activity.yaml
decisions:
  - "Removed non-queryable entries (proxy/firewall logs, Exabeam behavioral labels) — EventSpec requires a real Windows channel"
  - "T1490 retains contains field on bcdedit entry per RESEARCH.md Pitfall 4 guidance"
  - "T1057 keeps two 4688 entries with different descriptions for wmic vs tasklist distinction"
metrics:
  duration: 20min
  completed: 2026-03-24
---

# Phase 1 Plan 02: YAML EventSpec Migration Summary

**One-liner:** Migrated all 43 technique YAML files from free-text strings to structured `{event_id, channel, description}` EventSpec format queryable by the verifier.

## What Was Done

Converted every `expected_events` entry in all 43 YAML files from plain strings like:
```yaml
- "Sysmon 10 (ProcessAccess - TargetImage: lsass.exe, GrantedAccess: 0x1010)"
```
to structured EventSpec maps:
```yaml
- event_id: 10
  channel: "Microsoft-Windows-Sysmon/Operational"
  description: "ProcessAccess - TargetImage: lsass.exe, GrantedAccess: 0x1010 or 0x1410"
```

Channel values were assigned using the authoritative mapping from RESEARCH.md:
- Sysmon EIDs (1,3,7,10,11,12,13,19,20,21,22,23) → `Microsoft-Windows-Sysmon/Operational`
- Security EIDs (4xxx, 5xxx range) → `Security`
- System EIDs (104, 1102, 7045) → `System`
- PowerShell EIDs (4103, 4104) → `Microsoft-Windows-PowerShell/Operational`
- WMI EIDs (5857-5861) → `Microsoft-Windows-WMI-Activity/Operational`
- BITS EIDs (59, 60) → `Microsoft-Windows-Bits-Client/Operational`
- Defender EIDs (1116, 1117) → `Microsoft-Windows-Windows Defender/Operational`

## Verification Results

- `grep -rl 'event_id:' techniques/*.yaml | wc -l` = 43 (all files converted)
- `grep -rL 'event_id:' techniques/*.yaml` = empty (no unconverted files)
- `grep -c 'channel:' T1003_001_lsass.yaml` = 5
- `grep -c 'channel:' T1059_001_powershell.yaml` = 4
- `go build ./...` exits 0

## Deviations from Plan

### Auto-fixed Issues

None — plan executed as written with the following intentional scope decisions:

**1. Removed non-queryable entries**
- Found during: All files
- Issue: Several original strings referenced non-Windows-Event-Log sources: "Proxy/firewall log", "Exabeam: Brute Force use case trigger", "Exabeam: Abnormal enumeration activity". These cannot be queried via `Get-WinEvent`.
- Fix: Removed these entries. Only real Windows Event Log channels with valid EIDs are included.
- Rationale: EventSpec requires a `channel` that `Get-WinEvent` can query. Non-queryable entries would cause verifier errors.

**2. T1490 duplicate 4688 entries retained**
- Per RESEARCH.md Pitfall 4: T1490 has two distinct 4688 entries (vssadmin vs bcdedit). Both retained with `contains: "bcdedit"` on the second per the research recommendation.

## Known Stubs

None. All EventSpec entries reference real Windows Event Log channels and event IDs. The verifier (Plan 03) will query these directly.

## Self-Check: PASSED

- All 43 YAML files exist and contain `event_id:` entries
- Commit 744712b verified in git log
- `go build ./...` exits 0
