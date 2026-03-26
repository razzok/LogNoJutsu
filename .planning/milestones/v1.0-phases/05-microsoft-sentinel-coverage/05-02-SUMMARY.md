---
phase: 05-microsoft-sentinel-coverage
plan: 02
subsystem: playbooks
tags: [sentinel, azure, kerberoasting, ldap-recon, dcsync, tdd, yaml-techniques]
dependency_graph:
  requires: []
  provides: [AZURE_kerberoasting, AZURE_ldap_recon, AZURE_dcsync, TestAzureTechniques]
  affects: [internal/playbooks/embedded/techniques, internal/playbooks/loader_test.go]
tech_stack:
  added: []
  patterns: [SIEM-targeted YAML technique format (AZURE_ prefix), siem_coverage.sentinel map field, TDD RED/GREEN]
key_files:
  created:
    - internal/playbooks/embedded/techniques/AZURE_kerberoasting.yaml
    - internal/playbooks/embedded/techniques/AZURE_ldap_recon.yaml
    - internal/playbooks/embedded/techniques/AZURE_dcsync.yaml
  modified:
    - internal/playbooks/loader_test.go
decisions:
  - "AZURE_ldap_recon uses 'Anomalous LDAP Activity' with MEDIUM confidence comment — no dedicated built-in Sentinel rule for EID 1644 exists; community hunting query only"
  - "AZURE_kerberoasting adds machine SPNs as fallback to ensure volume threshold is reachable in environments with few service accounts"
  - "AZURE_dcsync differs from T1003.006 by using LDAP ACL access to DS-Replication GUIDs rather than repadmin /syncall"
  - "TestNewTechniqueCount threshold updated from 51 to 54 to include 3 new AZURE_ techniques"
requirements-completed: [SENT-02]
metrics:
  duration: 4min
  completed: "2026-03-25T20:26:45Z"
  tasks: 1
  files: 4
---

# Phase 05 Plan 02: AZURE_ Sentinel Technique YAML Files Summary

**One-liner:** 3 Sentinel-targeted AZURE_ technique YAMLs (Kerberoasting/LDAP recon/DCSync) with siem_coverage.sentinel rule mappings, validated by TDD TestAzureTechniques.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 (RED) | Add TestAzureTechniques test | c2ee639 | internal/playbooks/loader_test.go |
| 1 (GREEN) | Create 3 AZURE_ technique YAML files | f871d92 | AZURE_kerberoasting.yaml, AZURE_ldap_recon.yaml, AZURE_dcsync.yaml |

## What Was Built

Three new AZURE_-prefixed technique YAML files for Microsoft Sentinel coverage, following the same pattern as FALCON_ techniques (Phase 4):

**AZURE_kerberoasting** — Sentinel-optimized Kerberoasting variant targeting the "Potential Kerberoasting" analytic rule. Requests TGS tickets for ALL enumerated SPNs plus machine SPNs to exceed the 15+ SPN threshold that triggers the Sentinel KQL rule. Generates EID 4769 (EncryptionType=0x17) on the DC. Tactic: credential-access / T1558.003.

**AZURE_ldap_recon** — LDAP directory reconnaissance using DirectorySearcher, ldifde, and dsquery. Primary detection via EID 4688 (process creation) and EID 4104 (PowerShell ScriptBlock). Mapped to "Anomalous LDAP Activity" community hunting query (MEDIUM confidence — not a built-in Sentinel analytic rule). Tactic: discovery / T1087.002.

**AZURE_dcsync** — DCSync simulation via LDAP ACL access to DS-Replication extended rights (GUIDs 1131f6aa/1131f6ad). Triggers EID 4662 that the Sentinel "Non Domain Controller Active Directory Replication" rule queries. Differs from T1003.006 (repadmin /syncall) by using direct LDAP ACL enumeration. Tactic: credential-access / T1003.006.

All 3 techniques use standard MITRE tactic names (not "sentinel"/"azure"), phase=attack, have non-empty expected_events and siem_coverage.sentinel.

## Test Coverage

- `TestAzureTechniques` — validates all 3 AZURE_ IDs load, have ExpectedEvents, SIEMCoverage["sentinel"] non-empty, standard MITRE tactics, phase="attack", and specific HIGH-confidence rule names
- `TestNewTechniqueCount` — updated to check >= 54 techniques (was 51) and includes AZURE_ IDs in required list
- Full suite: `go test ./...` green (9 packages)

## Deviations from Plan

None — plan executed exactly as written. The worktree required a `git rebase master` first to pull in Phase 3/4 work (FALCON files) that the branch was missing — this was a setup step, not a code deviation.

## Known Stubs

None. All 3 YAML files have fully populated:
- `expected_events` (2-3 events each with real Windows Event IDs)
- `siem_coverage.sentinel` (named Sentinel analytic rules)
- `executor.command` (multi-method PowerShell simulation commands)

## Self-Check: PASSED

- AZURE_kerberoasting.yaml: FOUND
- AZURE_ldap_recon.yaml: FOUND
- AZURE_dcsync.yaml: FOUND
- loader_test.go contains TestAzureTechniques: FOUND
- Commit c2ee639 (RED): FOUND
- Commit f871d92 (GREEN): FOUND
- go test ./internal/playbooks/... -run TestAzureTechniques: PASSED
- go test ./...: PASSED (all 9 packages)
