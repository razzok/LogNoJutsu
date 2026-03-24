# LogNoJutsu

## What This Is

LogNoJutsu is a SIEM validation tool that simulates MITRE ATT&CK techniques on Windows, generating the system artifacts (Windows Event Logs, Sysmon events, PowerShell logs) that SIEMs should detect. It is used by security consultants during SIEM onboarding and validation engagements with clients. When a technique runs and the SIEM does not alert, LogNoJutsu helps pinpoint whether the problem is technique execution, log forwarding, parsing, or rule configuration.

## Core Value

Automated pass/fail verification that SIEM detection rules fire when attack techniques execute — eliminating manual log correlation during client SIEM validation engagements.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

- ✓ Web UI with simulation control (start/stop/status) — v0.1
- ✓ Multi-phase simulation (Preparation → Discovery → Attack) — v0.1
- ✓ Multi-user simulation mode — v0.2
- ✓ WhatIf dry-run mode — v0.3
- ✓ HTML report generation — v0.3
- ✓ Tactic filter (run only specific phases) — v0.3
- ✓ Technique-level logging to .log file — v0.1

### Active

<!-- Current scope. Building toward these. -->

- [ ] Events manifest: each technique declares expected Windows Event IDs / log sources
- [ ] Verification engine: after execution, query local Event Log and report pass/fail per technique
- [ ] Enhanced HTML report showing verification results (expected vs. observed events)
- [ ] Test coverage: Go unit and integration tests for engine, HTTP handlers, verification logic
- [ ] Code structure: refactor package-level globals to struct, split into packages
- [ ] Additional MITRE ATT&CK techniques and Exabeam UEBA scenarios
- [ ] CrowdStrike SIEM coverage: detection mappings + Falcon-sensor-specific techniques
- [ ] Microsoft Sentinel coverage: detection mappings + Azure AD / Sentinel-specific techniques

### Out of Scope

- SIEM API queries at runtime — adds external dependency, breaks standalone deployment model
- Non-Windows platforms — techniques are fundamentally Windows Event Log / Sysmon based
- Real credential extraction or actual exploitation — tool simulates artifacts only, never real attacks

## Context

- Inspired by Exabeam's internal "Magneto" tool (PowerShell-based, web UI, configurable timing)
- Single Go binary (~10 MB), no runtime dependencies, no internet required at runtime
- Distributed to clients as a standalone .exe for use on Windows 10/11/Server 2016+
- Requires local admin for Attack phase techniques; normal user for Discovery
- README written in German (target user base)
- Current state: zero test files, package-level globals in server.go block clean testing
- Codebase map available at `.planning/codebase/`

## Constraints

- **Platform**: Windows-only — techniques use Windows Event Log, PowerShell, Sysmon
- **Distribution**: Single binary — no installer, no runtime dependencies
- **Security**: Tool must not perform real attacks — simulate artifacts only
- **Language**: Go — existing codebase, no runtime needed on target machines

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go over PowerShell | Single binary distribution, no PS execution policy issues | ✓ Good |
| Local event log verification (not SIEM API) | Keeps tool standalone; SIEM-agnostic approach | — Pending |
| Flat package structure | Fast to build initially | ⚠️ Revisit — globals blocking tests |

---
*Last updated: 2026-03-24 after initial project initialization*
