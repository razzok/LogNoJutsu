---
phase: 9
slug: ui-polish
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-26
---

# Phase 9 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none (standard `go test ./...`) |
| **Quick run command** | `go test ./internal/reporter/... -run TestTacticColor -v` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go build ./... && go vet ./...`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green + manual browser check of all 5 requirements
- **Max feedback latency:** ~5 seconds (automated); manual browser checks per wave

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 9-W0-01 | Wave 0 | 0 | UI-04 | unit | `go test ./internal/reporter/... -run TestTacticColor -v` | ❌ W0 | ⬜ pending |
| 9-01-01 | 01 | 1 | UI-04 | unit | `go test ./internal/reporter/... -run TestTacticColor -v` | ✅ after W0 | ⬜ pending |
| 9-02-01 | 02 | 1 | VER-03, UI-03 | manual | Browser: check `.version-badge` text + `dashAvailable` count | — | ⬜ pending |
| 9-03-01 | 03 | 1 | UI-01, UI-02 | manual | Browser: grep no German strings + trigger prep step failure | — | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/reporter/reporter_test.go` — add `TestTacticColor` test covering `command-and-control` → `#f85149`, `ueba-scenario` → `#bc8cff`, and unknown tactic fallback → `#8b949e`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Version badge shows live version from `/api/info` | VER-03 | Browser DOM — no HTML test infrastructure | Load UI, check `.version-badge` text matches `go run . --version` output |
| No German strings in any visible UI text | UI-01 | Visual inspection of static HTML + JS-generated content | Load Scheduler tab, expand PoC mode, verify all labels/descriptions are English |
| Inline error panel on prep step failure; no `alert()` fires | UI-02 | Requires UI interaction + browser observation | Trigger prep step while WEF is offline; confirm panel appears below step row; no browser dialog |
| Dashboard "Techniques Available" count = 57 | UI-03 | Browser DOM verification | Load Dashboard tab; verify "Techniques Available" stat box shows correct count |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
