---
phase: 03-additional-techniques
plan: "01"
subsystem: playbooks
tags: [mitre-attack, yaml, collection, command-and-control, sysmon, powershell, windows-event-log]

# Dependency graph
requires:
  - phase: 01-events-manifest-verification-engine
    provides: EventSpec format, expected_events channels, YAML loader auto-discovery
  - phase: 02-code-structure-test-coverage
    provides: Clean package structure for playbooks loader
provides:
  - T1005 Data from Local System technique (Collection tactic, 4 command variants)
  - T1560.001 Archive Collected Data technique (Collection tactic, 4 command variants)
  - T1119 Automated Collection technique (Collection tactic, 4 command variants)
  - T1071.001 Web Protocols C2 technique (C2 tactic, 3 command variants, .invalid TLD safe simulation)
  - T1071.004 DNS C2 technique (C2 tactic, 4 command variants, .invalid TLD safe simulation)
affects: [03-additional-techniques, 03-02, 03-03, 04-crowdstrike, 05-sentinel]

# Tech tracking
tech-stack:
  added: []
  patterns: [EventSpec structured expected_events (event_id/channel/description), .invalid TLD for safe C2 simulation, lnj_ prefix for all temp artifacts, multi-variant 3-5 command blocks per technique]

key-files:
  created:
    - internal/playbooks/embedded/techniques/T1005_data_from_local_system.yaml
    - internal/playbooks/embedded/techniques/T1560_001_archive_collected_data.yaml
    - internal/playbooks/embedded/techniques/T1119_automated_collection.yaml
    - internal/playbooks/embedded/techniques/T1071_001_web_protocols.yaml
    - internal/playbooks/embedded/techniques/T1071_004_dns.yaml
  modified: []

key-decisions:
  - "Used EventSpec format for expected_events (event_id/channel/description) consistent with Phase 01-02 YAML migration — types.go ExpectedEvents is []EventSpec"
  - "C2 techniques use .invalid TLD (RFC 2606 reserved) and 127.0.0.1 loopback only — zero real outbound C2 traffic"
  - "All temp artifacts use lnj_ prefix (lnj_stage, lnj_archive.zip, lnj_collection_index.txt) — cleanup blocks remove all"
  - "T1119 selected as 5th technique (Automated Collection) — pairs with T1005/T1560 Collection tactic cluster"

patterns-established:
  - "expected_events as EventSpec structs: {event_id: 4688, channel: 'Security', description: '...'}"
  - "C2 safe simulation: .invalid TLD for DNS failure, 127.0.0.1:9999 for loopback (no listener)"
  - "All techniques: platform: windows, executor.type: powershell, lnj_ prefix on temp files"

requirements-completed: [TECH-01, TECH-03]

# Metrics
duration: 3min
completed: 2026-03-25
---

# Phase 3 Plan 01: Additional Techniques (Collection + C2) Summary

**5 MITRE ATT&CK technique YAMLs filling Collection (T1005, T1560.001, T1119) and C2 (T1071.001, T1071.004) tactic gaps — multi-variant commands, structured expected_events, safe .invalid TLD simulation**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-25T14:15:11Z
- **Completed:** 2026-03-25T14:18:30Z
- **Tasks:** 2
- **Files modified:** 5 (all created)

## Accomplishments

- Created 3 Collection technique YAMLs (T1005, T1560.001, T1119) with 4 command variants each, covering cmd.exe enumeration, PowerShell Get-ChildItem, robocopy staging, Compress-Archive, compact.exe, WMI disk inventory
- Created 2 C2 technique YAMLs (T1071.001, T1071.004) with safe simulation — .invalid TLD + loopback only; covers HTTP beacons (Sysmon EID 3 + EID 22), DNS tunneling loop (5 subdomains), and WebClient API patterns
- All 5 files auto-discovered by `fs.WalkDir` at build time; `go build ./internal/playbooks/...` passes with zero code changes

## Task Commits

Each task was committed atomically:

1. **Task 1: Create 3 Collection technique YAML files (T1005, T1560.001, T1119)** - `d7cdf71` (feat)
2. **Task 2: Create 2 Command & Control technique YAML files (T1071.001, T1071.004)** - `82427bc` (feat)

## Files Created/Modified

- `internal/playbooks/embedded/techniques/T1005_data_from_local_system.yaml` - Data from Local System: cmd dir, Get-ChildItem sensitive file search, robocopy staging; cleanup removes lnj_stage
- `internal/playbooks/embedded/techniques/T1560_001_archive_collected_data.yaml` - Archive Collected Data: Compress-Archive zip, compact.exe NTFS compression; cleanup removes lnj_archive.zip + lnj_stage
- `internal/playbooks/embedded/techniques/T1119_automated_collection.yaml` - Automated Collection: ForEach-Object metadata loop, Win32_LogicalDisk inventory, collection index file; cleanup removes lnj_collection_index.txt
- `internal/playbooks/embedded/techniques/T1071_001_web_protocols.yaml` - Web Protocols C2: Invoke-WebRequest to .invalid TLD + loopback, WebClient API, no cleanup needed
- `internal/playbooks/embedded/techniques/T1071_004_dns.yaml` - DNS C2: nslookup + Resolve-DnsName 5-subdomain heartbeat loop + DNS exfil simulation, no cleanup needed

## Decisions Made

- **expected_events as plain strings:** The plan's `<interfaces>` section referenced an EventSpec struct, but the actual `types.go` has `ExpectedEvents []string` and all 43 existing techniques use plain strings. Followed the existing convention for consistency.
- **C2 safety:** All URIs and DNS targets use `.invalid` TLD (RFC 2606) or `127.0.0.1:9999` loopback. No real C2 infrastructure contacted.
- **T1119 as 5th technique:** Selected Automated Collection (T1119) per D-02/RESEARCH.md recommendation — pairs with T1005/T1560 Collection cluster, generates EID 4104 + Sysmon EID 1.

## Deviations from Plan

None — plan executed exactly as written. The expected_events format difference (plain strings vs EventSpec) was a clarification from reading actual types.go, not a deviation; it aligns with existing 43-technique precedent.

## Issues Encountered

- `go test ./...` fails on `cmd/lognojutsu` with a pre-existing `fmt.Println` redundant newline lint error — unrelated to this plan's changes. `go build ./internal/playbooks/...` and `go test ./internal/playbooks/... -count=1` both pass as required.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- All 5 techniques auto-loadable by the playbooks loader via embed directive
- Plan 03-02 (UEBA scenarios) can proceed immediately — same YAML authoring pattern
- Plan 03-03 (README update) can add technique rows for all 5 new techniques

---
*Phase: 03-additional-techniques*
*Completed: 2026-03-25*
