---
phase: 04-crowdstrike-siem-coverage
plan: 03
subsystem: reporter
tags: [html-report, crowdstrike, siem-coverage, go-templates, tdd]

# Dependency graph
requires:
  - phase: 04-01
    provides: SIEMCoverage field on ExecutionResult and populated crowdstrike keys in YAML techniques
provides:
  - Conditional CrowdStrike coverage column in HTML report (cs-badge + rule names or N/A)
  - HasCrowdStrike bool flag computed by scanning results
  - siemCoverage funcMap helper for safe nil-map access in templates
  - TestHTMLCrowdStrikeColumn unit test covering present/absent/na_cell scenarios
affects: [04-02, future-sentinel-coverage]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Conditional CSS rendering: cs-badge/cs-na/cs-list classes emitted inside {{if .HasCrowdStrike}} block so absent HTML contains no SIEM vendor-specific markup"
    - "$.HasCrowdStrike root data access inside {{range .Results}} to reach parent template context"
    - "siemCoverage funcMap helper for nil-safe map[string][]string access in Go templates"

key-files:
  created: []
  modified:
    - internal/reporter/reporter.go
    - internal/reporter/reporter_test.go

key-decisions:
  - "CSS classes for CrowdStrike (cs-badge, cs-na, cs-list) rendered inside {{if .HasCrowdStrike}} conditional — ensures absent HTML has zero vendor-specific markup, satisfying test assertion that cs-badge string not present"
  - "siemCoverage funcMap helper used instead of direct index template function — prevents nil map panic when SIEMCoverage is nil on ExecutionResult"

patterns-established:
  - "Conditional SIEM vendor column pattern: HasVendor bool in htmlData, computed by scanning results, CSS + header + cells all gated on the same flag"

requirements-completed: [CROW-03]

# Metrics
duration: 15min
completed: 2026-03-25
---

# Phase 4 Plan 03: CrowdStrike HTML Column Summary

**Conditional CrowdStrike coverage column added to HTML report: CS badge + detection rule names for mapped techniques, grey N/A for unmapped, column entirely absent when no results carry CrowdStrike mappings.**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-03-25T19:44:00Z
- **Completed:** 2026-03-25T20:00:00Z
- **Tasks:** 2 completed (1 TDD RED + 1 TDD GREEN)
- **Files modified:** 2

## Accomplishments

- Added `HasCrowdStrike bool` to `htmlData` struct, computed by scanning `ExecutionResult.SIEMCoverage["crowdstrike"]`
- Implemented conditional `<th>CrowdStrike</th>` column header and per-row cell with CS badge and rule name list
- Added `siemCoverage` funcMap helper for nil-safe map access in Go templates
- CSS classes (`cs-badge`, `cs-na`, `cs-list`) conditionally rendered so absent HTML has zero CrowdStrike markup
- `TestHTMLCrowdStrikeColumn` with three subtests (present/absent/na_cell) added via TDD — RED then GREEN
- All 5 reporter tests pass; full suite (`go test ./...`) green with no regressions

## Task Commits

Each task was committed atomically:

1. **Task 1: Add TestHTMLCrowdStrikeColumn test** - `f9974c8` (test) — RED state
2. **Task 2: Implement CrowdStrike column in reporter** - `1abfa0f` (feat) — GREEN state

**Plan metadata:** (docs commit below)

_TDD: test commit at RED, implementation commit at GREEN_

## Files Created/Modified

- `internal/reporter/reporter.go` - Added HasCrowdStrike field, hasCrowdStrike computation, siemCoverage funcMap helper, conditional CSS classes, conditional column header and cells in HTML template
- `internal/reporter/reporter_test.go` - Added TestHTMLCrowdStrikeColumn with present/absent/na_cell subtests

## Decisions Made

- CSS classes for CrowdStrike rendered inside `{{if .HasCrowdStrike}}` conditional rather than always-present in stylesheet. This ensures the "absent" test subtest can verify that `cs-badge` string does not appear in the HTML at all when no technique has CrowdStrike mappings.
- Used `siemCoverage` funcMap helper instead of the built-in `index` template function to handle nil `SIEMCoverage` maps without panic (nil map indexed with `index` returns zero value safely, but the helper makes intent explicit and matches the plan's recommendation).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] CSS rendered conditionally instead of always in stylesheet**
- **Found during:** Task 2 (first test run after implementation)
- **Issue:** Plan placed cs-badge/cs-na/cs-list CSS always in the `<style>` block. The "absent" subtest asserts that "cs-badge" string does NOT appear in HTML when no CrowdStrike coverage — but CSS class names in the stylesheet would fail this assertion.
- **Fix:** Wrapped the three CS CSS rules inside `{{if .HasCrowdStrike}}...{{end}}` in the template's style block so the class definitions are only emitted when the column is visible.
- **Files modified:** internal/reporter/reporter.go
- **Verification:** TestHTMLCrowdStrikeColumn/absent passes; present and na_cell also pass
- **Committed in:** 1abfa0f (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - bug in plan's CSS placement assumption)
**Impact on plan:** Strengthens test accuracy — absent scenario is fully clean with no vendor-specific markup.

## Self-Check: PASSED

- `internal/reporter/reporter.go` contains `HasCrowdStrike bool` — verified
- `internal/reporter/reporter.go` contains `hasCrowdStrike := false` — verified
- `internal/reporter/reporter.go` contains `HasCrowdStrike: hasCrowdStrike,` — verified
- `internal/reporter/reporter.go` contains `"siemCoverage"` in funcMap — verified
- `internal/reporter/reporter.go` htmlTemplate contains `{{if .HasCrowdStrike}}<th>CrowdStrike</th>{{end}}` — verified
- `internal/reporter/reporter.go` htmlTemplate contains `cs-badge` CSS definition — verified
- `internal/reporter/reporter.go` htmlTemplate contains `cs-na` CSS definition — verified
- `internal/reporter/reporter.go` htmlTemplate contains `$.HasCrowdStrike` — verified
- `internal/reporter/reporter_test.go` contains `func TestHTMLCrowdStrikeColumn` — verified
- `go test ./internal/reporter/... -run TestHTMLCrowdStrikeColumn` exits 0 — verified
- `go test ./...` exits 0 — verified
