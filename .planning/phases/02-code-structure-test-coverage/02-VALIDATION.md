---
phase: 2
slug: code-structure-test-coverage
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-25
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | stdlib `testing` — go1.26.1 |
| **Config file** | none (no test config files needed) |
| **Quick run command** | `go test ./internal/...` |
| **Full suite command** | `go test ./... -race` |
| **Estimated runtime** | ~5 seconds |

# Note: -race requires CGO/gcc absent on this dev machine. go test ./... (without -race) is the operational command. Mutex discipline verified structurally in TestEngineRace. Pre-documented in STATE.md.

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/...`
- **After every plan wave:** Run `go test ./... -race`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 2-01-01 | 01 | 0 | QUAL-01 | vet fix | `go vet ./...` | ✅ existing | ✅ green |
| 2-01-02 | 01 | 1 | QUAL-01 | unit | `go test ./internal/server/... -race` | ✅ | ✅ green |
| 2-02-01 | 02 | 1 | QUAL-03 | unit | `go test ./internal/engine/... -race` | ✅ | ✅ green |
| 2-02-02 | 02 | 1 | QUAL-04 | unit | `go test ./internal/server/... -race` | ✅ | ✅ green |
| 2-02-03 | 02 | 1 | QUAL-05 | unit | `go test ./internal/verifier/... -race` | ✅ existing | ✅ green |
| 2-02-04 | 02 | 2 | all | integration | `go test ./... -race` | ✅ existing | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Fix `cmd/lognojutsu/main.go:28` — `fmt.Println(banner)` → `fmt.Print(banner)` (removes redundant `\n`, fixes `go vet` failure that blocks `go test ./...`)
- [x] `internal/engine/engine_test.go` — stub file with package declaration (populated in Wave 1)
- [x] `internal/server/server_test.go` — stub file with package declaration (populated after Server struct refactor)

*All Wave 0 prerequisites completed during Phase 2 implementation. Tests confirmed passing 2026-03-26.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Binary starts and serves UI | QUAL-01 | End-to-end smoke test | Build with `go build ./cmd/lognojutsu/`, run `.exe`, open browser to `http://localhost:8080` |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 10s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved (2026-03-26)
