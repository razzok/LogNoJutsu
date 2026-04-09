package engine

import (
	"sync"
	"testing"
	"time"

	"lognojutsu/internal/playbooks"
)

// ── Task 1: Day counter monotonicity + DayDigest lifecycle (TEST-02, TEST-04) ──

// dayCaptureClock wraps fakeClock and captures PoCDay + PoCPhase on each After() call.
// Used to verify monotonic day progression across Phase1->Gap->Phase2.
type dayCaptureClock struct {
	fakeClock
	eng    *Engine
	days   []int
	phases []string
	mu     sync.Mutex
}

func (c *dayCaptureClock) After(d time.Duration) <-chan time.Time {
	s := c.eng.GetStatus()
	if s.PoCDay > 0 {
		c.mu.Lock()
		c.days = append(c.days, s.PoCDay)
		c.phases = append(c.phases, s.PoCPhase)
		c.mu.Unlock()
	}
	return c.fakeClock.After(d)
}

// TestPoCDayCounter_Monotonic verifies globalDay increments monotonically 1..N
// across Phase1, Gap, and Phase2 without gaps or resets (TEST-02).
func TestPoCDayCounter_Monotonic(t *testing.T) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1059", "discovery", "execution"),
		testTechnique("T1078", "attack", "persistence"),
	)
	reg.Campaigns["camp-test"] = &playbooks.Campaign{
		ID:   "camp-test",
		Name: "Test Campaign",
		Steps: []playbooks.CampaignStep{
			{TechniqueID: "T1078", DelayAfter: 0},
		},
	}

	eng := New(reg, nil)
	dc := &dayCaptureClock{
		fakeClock: fakeClock{now: time.Date(2026, 4, 8, 6, 0, 0, 0, time.UTC)},
		eng:       eng,
	}
	eng.clock = dc
	eng.runner = fakeRunner(0)

	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 3,
		Phase1TechsPerDay:  1,
		Phase1DailyHour:    8,
		GapDays:            2,
		Phase2DurationDays: 2,
		Phase2DailyHour:    9,
		CampaignID:         "camp-test",
	}

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	dc.mu.Lock()
	days := make([]int, len(dc.days))
	phases := make([]string, len(dc.phases))
	copy(days, dc.days)
	copy(phases, dc.phases)
	dc.mu.Unlock()

	if len(days) == 0 {
		t.Fatal("no PoCDay values captured — After() may not be wired")
	}

	// Assert monotonically increasing (each value >= previous)
	for i := 1; i < len(days); i++ {
		if days[i] < days[i-1] {
			t.Errorf("day sequence not monotonic at index %d: days[%d]=%d < days[%d]=%d",
				i, i, days[i], i-1, days[i-1])
		}
	}

	// Assert span: first captured >= 1, last captured == 7
	if days[0] < 1 {
		t.Errorf("first captured PoCDay=%d, want >= 1", days[0])
	}
	if days[len(days)-1] != 7 {
		t.Errorf("last captured PoCDay=%d, want 7 (3+2+2)", days[len(days)-1])
	}

	// Assert phase transition ordering: no "phase2" before "gap", no "gap" before "phase1"
	// Find indices of first occurrence of each phase
	firstGap := -1
	firstPhase2 := -1
	for i, p := range phases {
		if p == "gap" && firstGap == -1 {
			firstGap = i
		}
		if p == "phase2" && firstPhase2 == -1 {
			firstPhase2 = i
		}
	}

	// If we have gap days, gap must appear after phase1 entries
	if firstGap != -1 && firstGap == 0 {
		t.Error("gap appeared before any phase1 entries")
	}
	// If we have phase2, it must appear after gap (or after phase1 if no gap)
	if firstPhase2 != -1 && firstGap != -1 && firstPhase2 < firstGap {
		t.Errorf("phase2 appeared (index %d) before gap (index %d)", firstPhase2, firstGap)
	}

	// Final status check
	finalStatus := eng.GetStatus()
	if finalStatus.PoCDay != 7 {
		t.Errorf("final PoCDay=%d, want 7", finalStatus.PoCDay)
	}
}

// digestCaptureClock wraps fakeClock and snapshots GetDayDigests() on each After() call.
// Used to observe DayDigest state transitions mid-run.
type digestCaptureClock struct {
	fakeClock
	eng       *Engine
	snapshots [][]DayDigest
	mu        sync.Mutex
}

func (c *digestCaptureClock) After(d time.Duration) <-chan time.Time {
	digests := c.eng.GetDayDigests()
	if len(digests) > 0 {
		snap := make([]DayDigest, len(digests))
		copy(snap, digests)
		c.mu.Lock()
		c.snapshots = append(c.snapshots, snap)
		c.mu.Unlock()
	}
	return c.fakeClock.After(d)
}

// TestDayDigest_PendingActiveComplete verifies DayDigest status transitions
// pending->active->complete are observed during PoC execution (TEST-04).
func TestDayDigest_PendingActiveComplete(t *testing.T) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1059", "discovery", "execution"),
		testTechnique("T1078", "attack", "persistence"),
	)
	reg.Campaigns["camp-test"] = &playbooks.Campaign{
		ID:   "camp-test",
		Name: "Test Campaign",
		Steps: []playbooks.CampaignStep{
			{TechniqueID: "T1078", DelayAfter: 0},
		},
	}

	eng := New(reg, nil)
	dc := &digestCaptureClock{
		fakeClock: fakeClock{now: time.Date(2026, 4, 8, 6, 0, 0, 0, time.UTC)},
		eng:       eng,
	}
	eng.clock = dc
	eng.runner = fakeRunner(0)

	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 1,
		Phase1TechsPerDay:  1,
		Phase1DailyHour:    8,
		GapDays:            1,
		Phase2DurationDays: 1,
		Phase2DailyHour:    9,
		CampaignID:         "camp-test",
	}

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	dc.mu.Lock()
	snapshots := make([][]DayDigest, len(dc.snapshots))
	copy(snapshots, dc.snapshots)
	dc.mu.Unlock()

	if len(snapshots) == 0 {
		t.Fatal("no digest snapshots captured — After() may not be wired")
	}

	// Assert: at least one snapshot shows a day with DayActive status
	foundActive := false
	for _, snap := range snapshots {
		for _, d := range snap {
			if d.Status == DayActive {
				foundActive = true
				break
			}
		}
		if foundActive {
			break
		}
	}
	if !foundActive {
		t.Error("no snapshot showed DayActive status — active state was never observed mid-run")
	}

	// Assert: for each day index, status transitions are valid
	// Valid transitions: pending->active, active->complete, pending->complete (no reversals)
	numDays := 3
	for dayIdx := 0; dayIdx < numDays; dayIdx++ {
		var observed []DayStatus
		for _, snap := range snapshots {
			if dayIdx < len(snap) {
				st := snap[dayIdx].Status
				// Only append if status changed
				if len(observed) == 0 || observed[len(observed)-1] != st {
					observed = append(observed, st)
				}
			}
		}
		// Check no reverse transitions
		for i := 1; i < len(observed); i++ {
			prev := observed[i-1]
			curr := observed[i]
			// complete -> active or complete -> pending are invalid
			if prev == DayComplete && curr != DayComplete {
				t.Errorf("day %d: invalid transition %s -> %s", dayIdx+1, prev, curr)
			}
			// active -> pending is invalid
			if prev == DayActive && curr == DayPending {
				t.Errorf("day %d: invalid transition %s -> %s", dayIdx+1, prev, curr)
			}
		}
	}

	// Assert final state: all 3 digests are DayComplete with StartTime and EndTime set
	finalDigests := eng.GetDayDigests()
	if len(finalDigests) != 3 {
		t.Fatalf("expected 3 final DayDigest entries, got %d", len(finalDigests))
	}
	for i, d := range finalDigests {
		if d.Status != DayComplete {
			t.Errorf("day %d final status=%q, want DayComplete", i+1, d.Status)
		}
		if d.StartTime == "" {
			t.Errorf("day %d: StartTime is empty", i+1)
		}
		if d.EndTime == "" {
			t.Errorf("day %d: EndTime is empty", i+1)
		}
	}
}

// ── Task 2: Stop-signal handling tests (TEST-03) ─────────────────────────────

// stopOnNthClock fires immediately for the first N-1 After() calls and blocks on the Nth call.
// This allows precise control of when the engine gets stuck in waitOrStop().
type stopOnNthClock struct {
	fakeClock
	blockAt   int
	blockCh   chan time.Time
	callCount int
	mu        sync.Mutex
}

func (c *stopOnNthClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	c.callCount++
	n := c.callCount
	c.mu.Unlock()
	if n >= c.blockAt {
		return c.blockCh
	}
	return c.fakeClock.After(d)
}

// newStopOnNthEngine creates an engine with stopOnNthClock injected for stop-signal tests.
func newStopOnNthEngine(phase1Days, gapDays, phase2Days, blockAt int) (*Engine, *stopOnNthClock) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1059", "discovery", "execution"),
		testTechnique("T1078", "attack", "persistence"),
	)
	reg.Campaigns["camp-test"] = &playbooks.Campaign{
		ID:   "camp-test",
		Name: "Test Campaign",
		Steps: []playbooks.CampaignStep{
			{TechniqueID: "T1078", DelayAfter: 0},
		},
	}
	eng := New(reg, nil)
	sc := &stopOnNthClock{
		fakeClock: fakeClock{now: time.Date(2026, 4, 8, 6, 0, 0, 0, time.UTC)},
		blockAt:   blockAt,
		blockCh:   make(chan time.Time), // never-firing channel
	}
	eng.clock = sc
	eng.runner = fakeRunner(0)
	return eng, sc
}

// TestPoCStop_DuringDayWait verifies stop signal during scheduling wait aborts the engine (D-03).
func TestPoCStop_DuringDayWait(t *testing.T) {
	// 2+1+1 = 4 days; blockAt=2 means day 1 wait fires, day 2 scheduling wait blocks
	eng, _ := newStopOnNthEngine(2, 1, 1, 2)

	cfg := pocConfig(2, 1, 1, "camp-test")

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for Phase1 to begin (day 1 fires immediately, day 2 will block)
	if !waitForPhase(eng, PhasePoCPhase1, 2*time.Second) {
		t.Fatalf("engine did not reach PhasePoCPhase1; stuck at %s", eng.GetStatus().Phase)
	}

	// Give engine time to enter the blocking second After() call
	time.Sleep(50 * time.Millisecond)

	eng.Stop()

	if !waitForPhase(eng, PhaseAborted, 2*time.Second) {
		t.Fatalf("engine did not reach PhaseAborted after Stop(); stuck at %s", eng.GetStatus().Phase)
	}

	// Verify engine did NOT reach PhaseDone
	if eng.GetStatus().Phase == PhaseDone {
		t.Error("engine reached PhaseDone — should have been aborted")
	}
}

// TestPoCStop_BetweenPhaseTransitions verifies stop at Phase1->Gap boundary aborts engine (D-04).
func TestPoCStop_BetweenPhaseTransitions(t *testing.T) {
	// 1+1+1 = 3 days; blockAt=2: Phase1 day 1 fires (call 1), gap day 1 blocks (call 2)
	eng, _ := newStopOnNthEngine(1, 1, 1, 2)

	cfg := pocConfig(1, 1, 1, "camp-test")

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for gap phase (Phase1 day 1 completes, engine enters gap)
	if !waitForPhase(eng, PhasePoCGap, 2*time.Second) {
		t.Fatalf("engine did not reach PhasePoCGap; stuck at %s", eng.GetStatus().Phase)
	}

	// Give engine time to enter the blocking After() call in gap
	time.Sleep(50 * time.Millisecond)

	eng.Stop()

	if !waitForPhase(eng, PhaseAborted, 2*time.Second) {
		t.Fatalf("engine did not reach PhaseAborted after Stop(); stuck at %s", eng.GetStatus().Phase)
	}

	finalStatus := eng.GetStatus()
	if finalStatus.Phase != PhaseAborted {
		t.Errorf("final phase=%q, want PhaseAborted", finalStatus.Phase)
	}
}

// TestPoCStop_DuringGapDays verifies stop during gap day wait aborts engine,
// and gap day 1 is DayActive (not complete) while phase1 day is DayComplete (D-05).
func TestPoCStop_DuringGapDays(t *testing.T) {
	// 1+2+1 = 4 days; blockAt=2: Phase1 fires (call 1), gap day 1 blocks (call 2)
	eng, _ := newStopOnNthEngine(1, 2, 1, 2)

	cfg := pocConfig(1, 2, 1, "camp-test")

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if !waitForPhase(eng, PhasePoCGap, 2*time.Second) {
		t.Fatalf("engine did not reach PhasePoCGap; stuck at %s", eng.GetStatus().Phase)
	}

	time.Sleep(50 * time.Millisecond)

	eng.Stop()

	if !waitForPhase(eng, PhaseAborted, 2*time.Second) {
		t.Fatalf("engine did not reach PhaseAborted after Stop(); stuck at %s", eng.GetStatus().Phase)
	}

	digests := eng.GetDayDigests()
	if len(digests) < 2 {
		t.Fatalf("expected at least 2 DayDigest entries, got %d", len(digests))
	}

	// Day 1 (phase1) should be DayComplete
	if digests[0].Status != DayComplete {
		t.Errorf("day 1 (phase1): Status=%q, want DayComplete", digests[0].Status)
	}

	// Day 2 (gap day 1) should be DayActive (started but not completed due to stop)
	if digests[1].Status != DayActive {
		t.Errorf("day 2 (gap day 1): Status=%q, want DayActive", digests[1].Status)
	}
}

// TestPoCStop_ImmediateAfterStart verifies stop before first day executes aborts engine (D-06).
// blockAt=1 means the very first After() call blocks.
func TestPoCStop_ImmediateAfterStart(t *testing.T) {
	// 2+1+1 = 4 days; blockAt=1: first scheduling wait blocks immediately
	eng, _ := newStopOnNthEngine(2, 1, 1, 1)

	cfg := pocConfig(2, 1, 1, "camp-test")

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Phase is set to PhasePoCPhase1 before the first After() call
	if !waitForPhase(eng, PhasePoCPhase1, 2*time.Second) {
		t.Fatalf("engine did not reach PhasePoCPhase1; stuck at %s", eng.GetStatus().Phase)
	}

	time.Sleep(50 * time.Millisecond)

	eng.Stop()

	if !waitForPhase(eng, PhaseAborted, 2*time.Second) {
		t.Fatalf("engine did not reach PhaseAborted after Stop(); stuck at %s", eng.GetStatus().Phase)
	}

	status := eng.GetStatus()
	// Day 1 was set but never executed (blocked in scheduling wait)
	if status.PoCDay != 1 {
		t.Errorf("PoCDay=%d, want 1 (stopped during day 1 scheduling wait)", status.PoCDay)
	}

	// All DayDigests should be DayPending or DayActive — none DayComplete
	digests := eng.GetDayDigests()
	for i, d := range digests {
		if d.Status == DayComplete {
			t.Errorf("day %d: Status=DayComplete — should not have completed (engine stopped immediately)", i+1)
		}
	}
}
