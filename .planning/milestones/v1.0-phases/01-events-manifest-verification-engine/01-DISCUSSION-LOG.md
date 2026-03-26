# Phase 1: Events Manifest & Verification Engine - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-24
**Phase:** 01-events-manifest-verification-engine
**Areas discussed:** Event ID data model, Report presentation, Verification timing, "Not executed" detection

---

## Event ID Data Model

| Option | Description | Selected |
|--------|-------------|----------|
| Rich structured type | New EventSpec struct: event_id (int), channel (string), description (string). Replaces free-text strings. | ✓ |
| Keep strings + add structured field | Keep existing strings as docs, add separate event_ids field for querying. | |
| Minimal — integers only | Just a list of Event ID integers. | |

**User's choice:** Rich structured type (EventSpec with event_id, channel, description)
**Notes:** Replaces the existing `ExpectedEvents []string` field in the Technique struct. All 39+ YAML technique files need updating.

---

## Optional keyword filter on EventSpec

| Option | Description | Selected |
|--------|-------------|----------|
| ID + channel + description only | Simple exact-match on event ID + channel | |
| Add optional keyword filter | EventSpec also has a 'contains' string for message filtering | |
| You decide | Left to Claude's discretion | ✓ |

**User's choice:** Claude's discretion
**Notes:** Balance precision vs complexity.

---

## Report Presentation

| Option | Description | Selected |
|--------|-------------|----------|
| Inline column in results table | Add verification column to existing per-technique table | ✓ |
| Separate verification section | New section below existing results table | |
| Summary only | Single-line badge in report header | |

**User's choice:** Inline column in existing results table

---

## Verification column detail level

| Option | Description | Selected |
|--------|-------------|----------|
| Pass/fail + event list | Badge + compact per-event-ID status list | ✓ |
| Pass/fail badge only | Just green/red badge | |
| Full event data | Timestamp, channel, message snippet per event | |

**User's choice:** Pass/fail badge + per-event-ID status (✓/✗ EID N channel)

---

## Verification Timing

| Option | Description | Selected |
|--------|-------------|----------|
| After each technique with short delay | Verify after each technique, configurable wait (default 3s) | ✓ |
| After full simulation completes | Single verification pass at end | |
| Manual trigger only | 'Verify' button in UI | |

**User's choice:** After each technique with configurable delay (default 3s)

---

## "Not Executed" Detection

| Option | Description | Selected |
|--------|-------------|----------|
| Use existing execution success flag | ExecutionResult.Success=false → "Not Executed" | ✓ |
| Check for specific output markers | Grep technique output for expected marker strings | |
| You decide | Left to Claude's discretion | |

**User's choice:** Use existing `ExecutionResult.Success` flag as gate

---

## Claude's Discretion

- Whether EventSpec includes optional `contains` keyword filter
- Query mechanism for Windows Event Log (PowerShell vs Win32 API)
- Time window for event search (how far back to look)
- New fields to add to ExecutionResult
- WhatIf mode behavior during verification

## Deferred Ideas

None.
