package engine

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"lognojutsu/internal/playbooks"
)

// TestRandomSlotsInWindow verifies that randomSlotsInWindow generates slot
// durations that all fall within the configured window boundaries (POC-03).
func TestRandomSlotsInWindow(t *testing.T) {
	// Use a fixed seed for determinism
	src := rand.NewSource(42)
	n := 5
	windowStart := 8
	windowEnd := 17
	// Anchor: midnight today so window is clearly in the future
	dayStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	durations := randomSlotsInWindow(n, windowStart, windowEnd, dayStart, src)

	if len(durations) != n {
		t.Fatalf("expected %d durations, got %d", n, len(durations))
	}

	// Reconstruct absolute slot times and verify they fall within window
	cursor := dayStart
	for i, d := range durations {
		if d < 0 {
			t.Errorf("slot %d has negative duration: %v", i, d)
		}
		cursor = cursor.Add(d)
		hour := cursor.Hour()
		if hour < windowStart || hour >= windowEnd {
			t.Errorf("slot %d at %v (hour %d) outside window [%d, %d)", i, cursor, hour, windowStart, windowEnd)
		}
	}

	// Guard: windowEnd <= windowStart should not panic
	durations2 := randomSlotsInWindow(1, 17, 8, dayStart, rand.NewSource(99))
	if len(durations2) != 1 {
		t.Fatalf("guard case: expected 1 duration, got %d", len(durations2))
	}
}

// afterCountClock counts how many times After() is called, for distributed-slot assertions.
type afterCountClock struct {
	mu    sync.Mutex
	inner *fakeClock
	count int
}

func (c *afterCountClock) Now() time.Time { return c.inner.Now() }
func (c *afterCountClock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
	return c.inner.After(d)
}
func (c *afterCountClock) getCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// TestPoCPhase1_DistributedSlots verifies that Phase 1 executes one technique
// per randomly-timed slot instead of all at once at a fixed hour (POC-01).
func TestPoCPhase1_DistributedSlots(t *testing.T) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1059", "discovery", "execution"),
	)

	fc := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	cc := &afterCountClock{inner: fc}

	eng := New(reg, nil)
	eng.clock = cc
	eng.runner = fakeRunner(0)

	// Phase 1: 1 day with 2 techniques. Distributed scheduling means 2 After() calls
	// (one per slot), not 1 (the old single-wait-then-burst pattern).
	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 1,
		Phase1TechsPerDay:  2,
		Phase1WindowStart:  0,
		Phase1WindowEnd:    23,
		GapDays:            0,
		Phase2DurationDays: 0,
	}
	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	// Distributed: Phase 1 day with 2 techniques = at least 2 After() calls (one per slot).
	got := cc.getCount()
	if got < 2 {
		t.Errorf("expected at least 2 After() calls for 2 distributed Phase 1 slots, got %d", got)
	}
}

// TestDayDigest_DistributedCounts verifies TechniqueCount and PassCount+FailCount
// are accurate when Phase 1 runs 3 techniques per day across distributed slots (D-03).
// This closes the coverage gap: existing TestDayDigest_Counts uses Phase1TechsPerDay=1
// and therefore does not exercise multi-technique distributed behavior.
func TestDayDigest_DistributedCounts(t *testing.T) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1059", "discovery", "execution"),
		testTechnique("T1078", "attack", "persistence"),
	)

	fc := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	cc := &afterCountClock{inner: fc}

	eng := New(reg, nil)
	eng.clock = cc
	eng.runner = fakeRunner(0) // all pass — simplifies PassCount assertion

	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 2,
		Phase1TechsPerDay:  3, // 3 distributed slots/day
		Phase1WindowStart:  0,
		Phase1WindowEnd:    23,
		GapDays:            0,
		Phase2DurationDays: 0,
	}
	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	digests := eng.GetDayDigests()
	if len(digests) != 2 {
		t.Fatalf("expected 2 DayDigest entries, got %d", len(digests))
	}
	for i, d := range digests {
		if d.TechniqueCount != 3 {
			t.Errorf("day %d: TechniqueCount=%d, want 3", i+1, d.TechniqueCount)
		}
		// fakeRunner(0) always succeeds, so PassCount should equal TechniqueCount
		if d.PassCount+d.FailCount != 3 {
			t.Errorf("day %d: PassCount+FailCount=%d, want 3", i+1, d.PassCount+d.FailCount)
		}
		if d.PassCount != 3 {
			t.Errorf("day %d: PassCount=%d, want 3 (fakeRunner always succeeds)", i+1, d.PassCount)
		}
		if d.Status != DayComplete {
			t.Errorf("day %d: Status=%q, want DayComplete", i+1, d.Status)
		}
	}

	// Verify distributed: 3 techs/day x 2 days = at least 6 After() calls (one per slot)
	got := cc.getCount()
	if got < 6 {
		t.Errorf("expected >= 6 After() calls for 3 distributed slots x 2 days, got %d", got)
	}
}

// TestDayDigest_Phase2StepCount verifies Phase 2 DayDigest TechniqueCount reflects
// the total number of campaign steps executed, not the number of batches (D-03).
func TestDayDigest_Phase2StepCount(t *testing.T) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1078", "attack", "persistence"),
		testTechnique("T1059", "attack", "execution"),
		testTechnique("T1003", "attack", "credential-access"),
		testTechnique("T1082", "attack", "discovery"),
	)
	reg.Campaigns["camp-5step"] = &playbooks.Campaign{
		ID:   "camp-5step",
		Name: "Five Step Campaign",
		Steps: []playbooks.CampaignStep{
			{TechniqueID: "T1078"},
			{TechniqueID: "T1059"},
			{TechniqueID: "T1003"},
			{TechniqueID: "T1082"},
			{TechniqueID: "T1087"},
		},
	}

	fc := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	eng := New(reg, nil)
	eng.clock = fc
	eng.runner = fakeRunner(0)

	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 0,
		GapDays:            0,
		Phase2DurationDays: 1,
		Phase2WindowStart:  0,
		Phase2WindowEnd:    23,
		CampaignID:         "camp-5step",
	}
	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	digests := eng.GetDayDigests()
	if len(digests) != 1 {
		t.Fatalf("expected 1 DayDigest entry, got %d", len(digests))
	}
	d := digests[0]
	// TechniqueCount should be 5 (total steps), not 2-3 (batch count)
	if d.TechniqueCount != 5 {
		t.Errorf("TechniqueCount=%d, want 5 (total campaign steps, not batch count)", d.TechniqueCount)
	}
	if d.PassCount+d.FailCount != 5 {
		t.Errorf("PassCount+FailCount=%d, want 5", d.PassCount+d.FailCount)
	}
	if d.Status != DayComplete {
		t.Errorf("Status=%q, want DayComplete", d.Status)
	}
}

// TestPoCPhase2_BatchedSlots verifies that Phase 2 executes techniques in
// batches of 2-3 at randomly-timed slots instead of all at once (POC-02).
func TestPoCPhase2_BatchedSlots(t *testing.T) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1078", "attack", "persistence"),
		testTechnique("T1059", "attack", "execution"),
		testTechnique("T1003", "attack", "credential-access"),
		testTechnique("T1082", "attack", "discovery"),
	)
	// Campaign with 5 steps forces at least 2 batch slots (ceil(5/3) = 2)
	reg.Campaigns["camp-multi"] = &playbooks.Campaign{
		ID:   "camp-multi",
		Name: "Multi-step Campaign",
		Steps: []playbooks.CampaignStep{
			{TechniqueID: "T1078"},
			{TechniqueID: "T1059"},
			{TechniqueID: "T1003"},
			{TechniqueID: "T1082"},
			{TechniqueID: "T1087"},
		},
	}

	fc := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	cc := &afterCountClock{inner: fc}

	eng := New(reg, nil)
	eng.clock = cc
	eng.runner = fakeRunner(0)

	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 0,
		GapDays:            0,
		Phase2DurationDays: 1,
		Phase2WindowStart:  0,
		Phase2WindowEnd:    23,
		CampaignID:         "camp-multi",
	}
	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	// Distributed: 5 steps in batches of 2-3 = at least 2 After() calls for batch slots.
	got := cc.getCount()
	if got < 2 {
		t.Errorf("expected at least 2 After() calls for batched Phase 2 slots, got %d", got)
	}
}
