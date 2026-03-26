# Phase 9: UI Polish — Discussion Log

**Date:** 2026-03-26
**Workflow:** discuss-phase

---

## Area: Error Panel Design

**Q: Where should the error message appear when a prep step fails?**
Options: Inline below step row / Replace status text / Error section below all steps
**Selected:** Inline below the step row — styled div expands below failing step, stays visible until next run

**Q: Should ALL alert() dialogs be replaced, or only the preparation step?**
Options: All alerts (no alert() anywhere) / Prep only
**Selected:** All alerts — no alert() anywhere (7 calls exist across prep, user management, campaign launch)

---

## Area: German String Scope

**Q: Which categories of German text should be translated?**
Options: HTML labels & descriptions / JS-generated schedule strings
**Selected:** Both — user note: "UI should be completely in English"

**Q: For the PoC schedule time format, what should replace 'X Tage • tägl. H:00 Uhr'?**
Options: X days • daily H:00 / X days • H:00 daily / You decide
**Selected:** X days • daily H:00

---

## Area: Library Count Stat

**Q: Where should the library technique count stat box appear?**
Options: Before session stats / After all session stats / Separate row/section
**Selected:** Before the session stats — order: Available → Run → Succeeded → Failed

**Q: What label should the stat box use?**
Options: Techniques Available / Library Size / Total Techniques
**Selected:** Techniques Available

---

*Log generated: 2026-03-26*
