---
phase: 12
slug: daily-digest-timeline-calendar-ui
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-09
---

# Phase 12 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package (stdlib) |
| **Config file** | none — `go test ./...` convention |
| **Quick run command** | `go test ./internal/server/... -v -run TestHandlePoCDays` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/server/... -v -run TestHandlePoCDays`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 12-01-01 | 01 | 1 | DIGEST-01 | manual | visual inspection | N/A | ⬜ pending |
| 12-01-02 | 01 | 1 | DIGEST-02 | manual | visual inspection | N/A | ⬜ pending |
| 12-01-03 | 01 | 1 | DIGEST-03 | manual | visual inspection | N/A | ⬜ pending |
| 12-01-04 | 01 | 1 | CAL-01 | manual | visual inspection | N/A | ⬜ pending |
| 12-01-05 | 01 | 1 | CAL-02 | manual | visual inspection | N/A | ⬜ pending |
| 12-01-06 | 01 | 1 | CAL-03 | manual | visual inspection | N/A | ⬜ pending |
| 12-01-07 | 01 | 1 | CAL-04 | manual | visual inspection | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. No new test files needed for Phase 12 — all requirements are UI-only and verified manually.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Digest panel renders per-day summaries | DIGEST-01 | No DOM testing infrastructure (no jsdom/Playwright) | Start binary with short PoC schedule, open Dashboard, verify digest panel appears with day rows |
| Active day auto-expands | DIGEST-02 | Requires visual DOM state inspection | During active PoC run, verify current day row is expanded; completed days are collapsed |
| Day entry shows counts + time window | DIGEST-03 | Visual content check | Expand a day row, verify technique count, pass/fail counts, and start→end timestamps are shown |
| Horizontal day grid visible | CAL-01 | Visual layout check | Verify horizontal strip of day cells appears above digest panel |
| Days color-coded by status | CAL-02 | Visual color check | Verify green (complete), accent (active), gray (pending), muted (gap) colors on day cells |
| Technique count in cell/tooltip | CAL-03 | Visual content check | Hover over day cells, verify tooltip shows technique count and pass/fail |
| Phase labels above groups | CAL-04 | Visual layout check | Verify "Phase 1", "Gap", "Phase 2" labels appear above corresponding day cell groups |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
