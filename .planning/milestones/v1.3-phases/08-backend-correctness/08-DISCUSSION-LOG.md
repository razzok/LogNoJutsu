# Phase 8: Backend Correctness - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-26
**Phase:** 08-backend-correctness
**Areas discussed:** Failure message content, /api/info auth, /api/info response shape

---

## Failure Message Content

| Option | Description | Selected |
|--------|-------------|----------|
| Description + error code | e.g. "Logon/Logoff events (4624, 4625, 4634): failed (exit status 87)" | ✓ |
| Description + remediation hint | e.g. "Logon/Logoff events: failed — run LogNoJutsu as Administrator" | |
| Description only | e.g. "Logon/Logoff events: failed" — clean, no technical noise | |

**User's choice:** Description + error code (Recommended)
**Notes:** Human-readable description first, raw exit code retained for debugging.

---

## /api/info Auth

| Option | Description | Selected |
|--------|-------------|----------|
| Bypass auth | Version is not sensitive — badge loads immediately, even before password entry | ✓ |
| Keep auth required | Consistent with all other API routes; badge only loads after login | |

**User's choice:** Yes, bypass auth (Recommended)
**Notes:** Registered without `authMiddleware` wrapper so the version badge works on the login screen.

---

## /api/info Response Shape

| Option | Description | Selected |
|--------|-------------|----------|
| Version only | `{"version":"v1.1.0"}` — simple, badge only needs version string | ✓ |
| Version + technique count | `{"version":"v1.1.0","technique_count":57}` | |
| Version + name + count | `{"version":"v1.1.0","name":"LogNoJutsu","technique_count":57}` | |

**User's choice:** Version only (Recommended)
**Notes:** Phase 9 uses `/api/techniques` for the Dashboard technique count — no need to duplicate it in `/api/info`.
