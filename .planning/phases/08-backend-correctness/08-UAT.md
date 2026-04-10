---
status: complete
phase: 08-backend-correctness
source: [08-01-SUMMARY.md, 08-02-SUMMARY.md]
started: 2026-03-26T00:00:00Z
updated: 2026-04-10T19:35:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Default build banner shows "dev"
expected: Build without ldflags, run binary. Banner ends with "dev" (e.g. "SIEM Validation & ATT&CK Simulation Tool  dev"). The old hardcoded "v0.1.0" should be gone.
result: pass

### 2. ldflags build injects version into banner
expected: Build with ldflags: `go build -ldflags "-X main.version=v1.1.0" -o lognojutsu.exe ./cmd/lognojutsu/` then run it. Banner should end with "v1.1.0" instead of "dev".
result: pass

### 3. GET /api/info returns version without auth
expected: Start the server (`./lognojutsu.exe`), then in another terminal: `curl http://localhost:8080/api/info` (no credentials). Should return `{"version":"dev"}` with HTTP 200. No 401 Unauthorized.
result: pass

### 4. Audit policy GUID format (code review)
expected: Open `internal/preparation/preparation.go`. The `auditPolicies` var should contain entries with GUIDs like `{0CCE9215-69AE-11D9-BED3-505054503030}` — not English names like "Logon". Error messages in the loop use `p.description` (readable description), not `p.guid`. Two entries have `// VERIFY` comments.
result: pass

## Summary

total: 4
passed: 4
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps
