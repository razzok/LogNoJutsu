---
phase: 03-additional-techniques
plan: "03"
subsystem: documentation
tags: [readme, german, mitre-attack, ueba, collection, command-and-control, exabeam]

# Dependency graph
requires:
  - phase: 03-additional-techniques/03-01
    provides: 5 ATT&CK technique YAML files (T1005, T1560.001, T1119, T1071.001, T1071.004)
  - phase: 03-additional-techniques/03-02
    provides: 4 UEBA scenario YAML files (UEBA-DATA-STAGING, UEBA-ACCOUNT-TAKEOVER, UEBA-PRIV-ESC, UEBA-LATERAL-NEW-ASSET)
provides:
  - README documentation for all 9 new techniques/scenarios in German
  - Per-technique property tables (Eigenschaft/Wert) matching existing format
  - Code blocks showing key commands extracted from YAML executors
  - UEBA-Erkennungslogik paragraphs explaining Exabeam detection logic
affects: [TECH-01, TECH-02, TECH-03 documentation completeness]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "README technique section format: H4 heading with em-dash, property table, Was wird ausgefuehrt code block, Erwartete SIEM-Events bullets"
    - "UEBA README section adds UEBA-Erkennungslogik paragraph in addition to standard fields"
    - "German content with umlauts handled via ASCII approximation in some sentences (ue/oe/ae)"

key-files:
  created: []
  modified:
    - README.md

key-decisions:
  - "Inserted ATT&CK sections at end of Phase 2: Attack block (after T1574.002) before UEBA-Szenarien heading — preserves phase/section organization"
  - "Inserted UEBA sections at end of UEBA-Szenarien block (after UEBA-LATERAL-CHAIN) before Kampagnen section — maintains section order"
  - "UEBA-Erkennungslogik paragraphs written with umlauts spelled out (ue/oe/ae) for cross-platform compatibility, matching plan specification"

# Metrics
duration: 3min
completed: 2026-03-25
---

# Phase 3 Plan 03: README Documentation Update Summary

**README updated with 9 German-language technique sections (5 ATT&CK + 4 UEBA) matching existing per-technique format with property tables, command blocks, and SIEM event lists**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-03-25T14:24:37Z
- **Completed:** 2026-03-25T14:27:23Z
- **Tasks:** 2
- **Files modified:** 1 (README.md, +332 lines)

## Accomplishments

- Added 5 ATT&CK technique sections to Phase 2: Attack block in README: T1005 (Data from Local System), T1560.001 (Archive Collected Data), T1119 (Automated Collection), T1071.001 (Web Protocols C2), T1071.004 (DNS C2)
- Added 4 UEBA scenario sections to UEBA-Szenarien block: UEBA-DATA-STAGING, UEBA-ACCOUNT-TAKEOVER, UEBA-PRIV-ESC, UEBA-LATERAL-NEW-ASSET
- Each section follows the existing `| Eigenschaft | Wert |` table format with MITRE ATT&CK link, Taktik, Exabeam-Regeln count, Admin erforderlich, Cleanup
- UEBA sections include `UEBA-Erkennungslogik` paragraph explaining Exabeam behavioral detection logic
- Commands in code blocks extracted from actual YAML executor files (Plans 03-01/03-02 output)
- All content in German, consistent with existing README language
- `go test ./... -count=1` passes green (all 5 packages with tests pass)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add 5 ATT&CK technique sections to README Phase 2: Attack** — `bd763c0` (feat)
2. **Task 2: Add 4 UEBA scenario sections to README UEBA-Szenarien** — `f848d63` (feat)

## Files Created/Modified

- `README.md` — 332 lines added: 5 ATT&CK sections (lines ~1903–2082) + 4 UEBA sections (lines ~2181–2332)

## Decisions Made

- **Insertion location for ATT&CK sections:** Placed at the end of Phase 2: Attack, immediately before `### UEBA-Szenarien` heading. This follows the natural phase structure and avoids disrupting the existing 43-technique layout.
- **Insertion location for UEBA sections:** Placed at the end of `### UEBA-Szenarien`, immediately before `## Kampagnen / Playbooks`. Total UEBA count is now 7.
- **Format matching:** UEBA sections use an extra `UEBA Use Case` property row (replacing `MITRE ATT&CK` + `Taktik`) plus `UEBA-Erkennungslogik` paragraph, exactly matching existing UEBA entries (UEBA-SPRAY-CHAIN, UEBA-OFFHOURS, UEBA-LATERAL-CHAIN).

## Deviations from Plan

None — plan executed exactly as written. All 9 sections added in German, matching existing format, with commands extracted from YAML files.

## Known Stubs

None — all README sections reference real YAML technique files created in Plans 03-01 and 03-02. No placeholder content.

## Self-Check: PASSED

- FOUND: `#### T1005` heading in README.md (line 1903)
- FOUND: `#### T1560.001` heading in README.md (line 1938)
- FOUND: `#### T1119` heading in README.md (line 1973)
- FOUND: `#### T1071.001` heading in README.md (line 2009)
- FOUND: `#### T1071.004` heading in README.md (line 2051)
- FOUND: `#### UEBA-DATA-STAGING` heading in README.md (line 2181)
- FOUND: `#### UEBA-ACCOUNT-TAKEOVER` heading in README.md (line 2220)
- FOUND: `#### UEBA-PRIV-ESC` heading in README.md (line 2255)
- FOUND: `#### UEBA-LATERAL-NEW-ASSET` heading in README.md (line 2292)
- FOUND: commit bd763c0 (Task 1: 5 ATT&CK sections)
- FOUND: commit f848d63 (Task 2: 4 UEBA sections)
- `go test ./... -count=1`: all packages pass (engine, playbooks, reporter, server, verifier)

---
*Phase: 03-additional-techniques*
*Completed: 2026-03-25*
