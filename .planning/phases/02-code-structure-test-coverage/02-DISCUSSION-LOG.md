# Phase 2: Discussion Log

**Session:** 2026-03-24
**Phase:** 02-code-structure-test-coverage

---

## Gray Areas Selected

User selected all four areas for discussion:
- Server struct pattern
- Executor abstraction
- Test breadth & depth
- Package rename: playbooks → techniques

---

## Area 1: Server struct pattern

**Q: How should server.go globals be restructured to enable testable handlers?**
Options presented:
- `Server` struct + method receivers (Recommended)
- `AppState` struct + closure injection

**Selected:** Server struct + method receivers

---

**Q: Should Start() become a method on Server, or stay a package-level function?**
Options presented:
- `server.Start(cfg)` stays — Server internal (Recommended)
- `New(cfg)` + `s.Start()` — exported constructor

**Selected:** server.Start(cfg) stays — Server internal (main.go unchanged)

---

## Area 2: Executor abstraction

**Q: To unit-test engine.Engine without shelling out to PowerShell, the executor calls need to be mockable. What approach?**
Options presented:
- RunnerFunc type injection (Recommended) — consistent with Phase 1 QueryFn pattern
- Executor interface
- Skip executor mocking — test around it

**Selected:** RunnerFunc type injection

---

## Area 3: Test breadth & depth

**Q: How much test coverage should Phase 2 aim for?**
Options presented:
- Critical paths + race detector (Recommended)
- Full QUAL coverage — all requirements

**Selected:** Critical paths + race detector (~13 specific tests)

---

**Q: stdlib testing package only, or add testify for cleaner assertions?**
Options presented:
- stdlib only (Recommended)
- Add testify/assert

**Selected:** stdlib only

---

## Area 4: Package rename

**Q: The ROADMAP targets a 'techniques' package but the existing package is 'playbooks'. Keep or rename?**
Options presented:
- Keep 'playbooks' (Recommended)
- Rename to 'techniques'

**Selected:** Keep 'playbooks' — no churn, document discrepancy

---

*Log generated: 2026-03-24*
