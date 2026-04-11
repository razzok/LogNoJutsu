# Phase 19: Distributed Technique Scheduling - Discussion Log

**Date:** 2026-04-10
**Participants:** User, Claude

## Areas Discussed

### 1. Time Window Configuration

**Q: How should the daily execution window be defined?**
Options: Start + end hour | Start hour + duration | You decide
**Selected:** Start + end hour
Notes: Replace Phase1DailyHour with Phase1WindowStart/Phase1WindowEnd pairs.

**Q: Should Phase 1 and Phase 2 have separate time windows, or share one?**
Options: Separate windows | Shared window | You decide
**Selected:** You decide
Notes: Claude picks based on existing config structure.

**Q: What default window values make sense for client engagements?**
Options: 8:00-17:00 (business hours) | 6:00-22:00 (extended) | 0:00-23:59 (full day)
**Selected:** 8:00-17:00 (business hours)
Notes: Blends with regular employee activity on the SIEM.

### 2. Jitter Algorithm

**Q: How should techniques be distributed within the time window?**
Options: Even spread + jitter | Fully random | You decide
**Selected:** Fully random
Notes: More unpredictable, realistic attacker behavior pattern.

**Q: Should NextScheduledRun show the next technique time or just the window?**
Options: Next technique time | Just the window | You decide
**Selected:** You decide
Notes: Claude picks what fits existing status polling pattern.

### 3. Phase 2 Batching

**Q: How should Phase 2 technique batches be sized?**
Options: Fixed 2-3 random | Configurable batch size | You decide
**Selected:** You decide
Notes: Claude picks what fits engagement workflow best.

**Q: Within a batch, should techniques run simultaneously or with short delays?**
Options: Short delays (existing DelayBetween) | Back to back (no delay) | You decide
**Selected:** Short delays (existing DelayBetween)
Notes: Preserves the burst-of-activity detection pattern.

**Q: How should campaign step.DelayAfter interact with jitter scheduling?**
Options: DelayAfter overrides jitter | Jitter replaces DelayAfter | You decide
**Selected:** You decide
Notes: Claude picks what preserves campaign semantics while enabling distribution.
