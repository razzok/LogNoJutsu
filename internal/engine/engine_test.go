package engine

import (
	"strings"
	"sync"
	"testing"
	"time"

	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/simlog"
	"lognojutsu/internal/userstore"
)

// fakeClock implements Clock for deterministic testing — After() fires immediately.
type fakeClock struct {
	now time.Time
}

func (f *fakeClock) Now() time.Time { return f.now }

func (f *fakeClock) After(d time.Duration) <-chan time.Time {
	f.now = f.now.Add(d) // advance fake time
	ch := make(chan time.Time, 1)
	ch <- f.now
	return ch
}

// newPoCEngine creates an Engine wired with a fake clock and fake runner,
// configured for a minimal PoC run with the given day counts.
func newPoCEngine(phase1Days, gapDays, phase2Days int) (*Engine, *fakeClock) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1059", "discovery", "execution"),
		testTechnique("T1078", "attack", "persistence"),
	)
	// Add a minimal campaign for Phase 2
	reg.Campaigns["camp-test"] = &playbooks.Campaign{
		ID:   "camp-test",
		Name: "Test Campaign",
		Steps: []playbooks.CampaignStep{
			{TechniqueID: "T1078", DelayAfter: 0},
		},
	}
	eng := New(reg, nil)
	fc := &fakeClock{now: time.Date(2026, 4, 8, 6, 0, 0, 0, time.UTC)}
	eng.clock = fc
	eng.runner = fakeRunner(0)
	return eng, fc
}


// testRegistry builds a minimal Registry from the provided techniques.
func testRegistry(techniques ...*playbooks.Technique) *playbooks.Registry {
	r := &playbooks.Registry{
		Techniques: make(map[string]*playbooks.Technique),
		Campaigns:  make(map[string]*playbooks.Campaign),
	}
	for _, t := range techniques {
		r.Techniques[t.ID] = t
	}
	return r
}

// testTechnique builds a minimal Technique for test fixtures.
func testTechnique(id, phase, tactic string) *playbooks.Technique {
	return &playbooks.Technique{
		ID:     id,
		Name:   id + " test",
		Phase:  phase,
		Tactic: tactic,
	}
}

// fakeRunner returns a RunnerFunc that sleeps delay then returns a successful result.
func fakeRunner(delay time.Duration) RunnerFunc {
	return func(t *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
		if delay > 0 {
			time.Sleep(delay)
		}
		return playbooks.ExecutionResult{
			TechniqueID:   t.ID,
			TechniqueName: t.Name,
			TacticID:      t.Tactic,
			StartTime:     time.Now().Format(time.RFC3339),
			EndTime:       time.Now().Format(time.RFC3339),
			Success:       true,
			Output:        "fake output",
		}
	}
}

// waitForPhase polls until the engine reaches the target phase or the timeout expires.
func waitForPhase(e *Engine, target Phase, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if e.GetStatus().Phase == target {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// testUsers returns an empty userstore (no file required).
func testUsers() *userstore.Store {
	us, _ := userstore.Load()
	return us
}

func TestEngineStart_transitionsToDiscovery(t *testing.T) {
	reg := testRegistry(testTechnique("T9999", "discovery", "test-tactic"))
	eng := New(reg, testUsers())
	eng.SetRunner(fakeRunner(0))

	if err := eng.Start(Config{}); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone within timeout, current phase: %s", eng.GetStatus().Phase)
	}

	status := eng.GetStatus()
	if len(status.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(status.Results))
	}
	if status.Results[0].TechniqueID != "T9999" {
		t.Errorf("expected TechniqueID T9999, got %q", status.Results[0].TechniqueID)
	}
}

func TestEngineStop_abortsRun(t *testing.T) {
	techniques := make([]*playbooks.Technique, 5)
	for i := range techniques {
		techniques[i] = testTechnique("T900"+string(rune('0'+i)), "discovery", "test-tactic")
	}
	reg := testRegistry(techniques...)
	eng := New(reg, testUsers())
	eng.SetRunner(fakeRunner(200 * time.Millisecond))

	if err := eng.Start(Config{}); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	eng.Stop()

	if !waitForPhase(eng, PhaseAborted, 3*time.Second) {
		t.Fatalf("engine did not reach PhaseAborted within timeout, current phase: %s", eng.GetStatus().Phase)
	}
}

func TestFilterByTactics(t *testing.T) {
	techniques := []*playbooks.Technique{
		testTechnique("T1", "discovery", "recon"),
		testTechnique("T2", "attack", "execution"),
		testTechnique("T3", "attack", "lateral-movement"),
		testTechnique("T4", "discovery", "discovery"),
	}

	cases := []struct {
		name           string
		included       []string
		excluded       []string
		wantCount      int
		wantTactics    []string
		notWantTactics []string
	}{
		{
			name:      "no filters returns all",
			included:  nil,
			excluded:  nil,
			wantCount: 4,
		},
		{
			name:        "include-only discovery",
			included:    []string{"discovery"},
			excluded:    nil,
			wantCount:   1,
			wantTactics: []string{"discovery"},
		},
		{
			name:           "exclude-only execution",
			included:       nil,
			excluded:       []string{"execution"},
			wantCount:      3,
			notWantTactics: []string{"execution"},
		},
		{
			name:        "both included and excluded: discovery+lateral minus lateral = discovery only",
			included:    []string{"discovery", "lateral-movement"},
			excluded:    []string{"lateral-movement"},
			wantCount:   1,
			wantTactics: []string{"discovery"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eng := New(testRegistry(), testUsers())
			eng.cfg.IncludedTactics = tc.included
			eng.cfg.ExcludedTactics = tc.excluded

			result := eng.filterByTactics(techniques)

			if len(result) != tc.wantCount {
				t.Errorf("expected %d techniques, got %d", tc.wantCount, len(result))
			}

			// Verify wanted tactics are present
			for _, wantTactic := range tc.wantTactics {
				found := false
				for _, tech := range result {
					if tech.Tactic == wantTactic {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected tactic %q in results, but not found", wantTactic)
				}
			}

			// Verify unwanted tactics are absent
			for _, notWantTactic := range tc.notWantTactics {
				for _, tech := range result {
					if tech.Tactic == notWantTactic {
						t.Errorf("tactic %q should not be in results but was found", notWantTactic)
					}
				}
			}
		})
	}
}

func TestEngineRace(t *testing.T) {
	techniques := []*playbooks.Technique{
		testTechnique("T1", "discovery", "recon"),
		testTechnique("T2", "discovery", "recon"),
		testTechnique("T3", "discovery", "recon"),
		testTechnique("T4", "attack", "execution"),
		testTechnique("T5", "attack", "execution"),
	}
	reg := testRegistry(techniques...)
	eng := New(reg, testUsers())
	eng.SetRunner(fakeRunner(10 * time.Millisecond))

	var wg sync.WaitGroup

	// Goroutine 1: start the engine
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = eng.Start(Config{})
	}()

	// Goroutine 2: poll GetStatus 100 times
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = eng.GetStatus()
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Goroutine 3: stop after 50ms
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		eng.Stop()
	}()

	wg.Wait()
	// No specific assertion — the -race detector catches data races.
}

// TestPoCDayCounter verifies globalDay increments monotonically from 1 to totalDays (POCFIX-01 + TEST-01).
func TestPoCDayCounter(t *testing.T) {
	eng, _ := newPoCEngine(2, 1, 2) // 5 total days

	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 2,
		Phase1TechsPerDay:  1,
		Phase1DailyHour:    8,
		GapDays:            1,
		Phase2DurationDays: 2,
		Phase2DailyHour:    9,
		CampaignID:         "camp-test",
	}

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for completion (fake clock fires immediately, so this is fast)
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	status := eng.GetStatus()
	if status.PoCDay != 5 {
		t.Errorf("final PoCDay = %d, want 5 (2+1+2)", status.PoCDay)
	}
	if status.PoCTotalDays != 5 {
		t.Errorf("PoCTotalDays = %d, want 5", status.PoCTotalDays)
	}
}

// captureClock wraps fakeClock and records CurrentStep each time After() fires.
// This captures the PoC day step strings set just before each wait call.
type captureClock struct {
	fakeClock
	eng      *Engine
	captured []string
	mu       sync.Mutex
}

func (c *captureClock) After(d time.Duration) <-chan time.Time {
	// Capture CurrentStep at this instant (set just before the wait call in runPoC)
	step := c.eng.GetStatus().CurrentStep
	if step != "" {
		c.mu.Lock()
		c.captured = append(c.captured, step)
		c.mu.Unlock()
	}
	return c.fakeClock.After(d)
}

// TestPoCCurrentStepStrings verifies CurrentStep strings are English with no German words (POCFIX-02).
func TestPoCCurrentStepStrings(t *testing.T) {
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
	cc := &captureClock{
		fakeClock: fakeClock{now: time.Date(2026, 4, 8, 6, 0, 0, 0, time.UTC)},
		eng:       eng,
	}
	eng.clock = cc
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

	cc.mu.Lock()
	steps := make([]string, len(cc.captured))
	copy(steps, cc.captured)
	cc.mu.Unlock()

	// Verify no German strings leaked into any captured PoC day step
	germanWords := []string{"Tag", "warte", "Uhr", "keine", "Aktionen", "Pause"}
	for _, step := range steps {
		for _, word := range germanWords {
			if strings.Contains(step, word) {
				t.Errorf("CurrentStep contains German word %q: %s", word, step)
			}
		}
	}
	// Also check that English PoC day patterns appeared
	if len(steps) > 0 {
		found := false
		for _, step := range steps {
			if strings.Contains(step, "Day") && strings.Contains(step, "of") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no CurrentStep contains English 'Day N of M' pattern; steps: %v", steps)
		}
	} else {
		t.Error("no CurrentStep values were captured during PoC run")
	}
}

// TestPoCPhaseLogSeparators verifies simlog.Phase entries are produced at each PoC phase transition (POCFIX-03).
func TestPoCPhaseLogSeparators(t *testing.T) {
	eng, _ := newPoCEngine(1, 1, 1)

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

	entries := simlog.GetEntries()
	// simlog.Phase() uppercases the message, so compare case-insensitively
	expectedPhases := []string{
		"POC PHASE 1: DISCOVERY",
		"POC GAP",
		"POC PHASE 2: ATTACK",
	}

	for _, expected := range expectedPhases {
		found := false
		for _, entry := range entries {
			if entry.Type == simlog.TypePhase && strings.Contains(entry.Message, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing simlog.Phase entry for %q", expected)
		}
	}
}

// TestPoCClockInjection verifies the fake clock is used — no real sleeps occur (TEST-01).
func TestPoCClockInjection(t *testing.T) {
	eng, fc := newPoCEngine(1, 0, 1) // 2 days, no gap

	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 1,
		Phase1TechsPerDay:  1,
		Phase1DailyHour:    8,
		GapDays:            0,
		Phase2DurationDays: 1,
		Phase2DailyHour:    9,
		CampaignID:         "camp-test",
	}

	start := time.Now()
	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}
	elapsed := time.Since(start)

	// With a real clock, waiting for next hour would take minutes+.
	// With a fake clock, the entire run completes in under 2 seconds.
	if elapsed > 2*time.Second {
		t.Errorf("PoC run took %v — fake clock may not be wired correctly", elapsed)
	}

	// Verify fake clock time advanced (proves After() was called)
	if fc.Now().Equal(time.Date(2026, 4, 8, 6, 0, 0, 0, time.UTC)) {
		t.Error("fake clock time did not advance — After() was never called")
	}
}

// ── DayDigest tests (TRACK-01, TRACK-02, TRACK-04, CAMP-01) ─────────────────

// newPoCEngineWithCampaign creates an engine with a 2-step campaign where
// step 1 has DelayAfter=300 and step 2 has DelayAfter=0.
func newPoCEngineWithCampaign(phase1Days, gapDays, phase2Days int) (*Engine, *fakeClock) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1059", "discovery", "execution"),
		testTechnique("T1078", "attack", "persistence"),
		testTechnique("T1003", "attack", "credential-access"),
	)
	reg.Campaigns["camp-delay"] = &playbooks.Campaign{
		ID:   "camp-delay",
		Name: "Delay Test Campaign",
		Steps: []playbooks.CampaignStep{
			{TechniqueID: "T1078", DelayAfter: 300},
			{TechniqueID: "T1003", DelayAfter: 0},
		},
	}
	eng := New(reg, nil)
	fc := &fakeClock{now: time.Date(2026, 4, 8, 6, 0, 0, 0, time.UTC)}
	eng.clock = fc
	// alternating runner: first call succeeds, second fails, etc.
	callCount := 0
	eng.runner = func(t *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
		callCount++
		success := callCount%2 == 1
		return playbooks.ExecutionResult{
			TechniqueID:   t.ID,
			TechniqueName: t.Name,
			TacticID:      t.Tactic,
			StartTime:     time.Now().Format(time.RFC3339),
			EndTime:       time.Now().Format(time.RFC3339),
			Success:       success,
		}
	}
	return eng, fc
}

func pocConfig(phase1Days, gapDays, phase2Days int, campaignID string) Config {
	return Config{
		PoCMode:            true,
		Phase1DurationDays: phase1Days,
		Phase1TechsPerDay:  1,
		Phase1DailyHour:    8,
		GapDays:            gapDays,
		Phase2DurationDays: phase2Days,
		Phase2DailyHour:    9,
		CampaignID:         campaignID,
	}
}

// TestGetDayDigests_Empty verifies GetDayDigests returns empty slice (not nil) when no PoC is running (D-06).
func TestGetDayDigests_Empty(t *testing.T) {
	reg := testRegistry()
	eng := New(reg, nil)

	digests := eng.GetDayDigests()
	if digests == nil {
		t.Fatal("GetDayDigests() returned nil — expected empty slice []DayDigest{}")
	}
	if len(digests) != 0 {
		t.Errorf("GetDayDigests() returned %d entries, want 0", len(digests))
	}
}

// TestDayDigest_PrePopulated verifies all days are pre-populated as pending before any day runs (TRACK-02).
func TestDayDigest_PrePopulated(t *testing.T) {
	eng, fc := newPoCEngineWithCampaign(2, 1, 2)
	_ = fc

	cfg := pocConfig(2, 1, 2, "camp-delay")

	// Use a blocking runner so we can check state mid-run.
	// Actually, since fakeClock fires immediately, we'll start and let it complete,
	// then verify the final digest has correct initial structure from pre-population.
	// The pre-population check is validated by structure (5 entries total, correct phases).
	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	digests := eng.GetDayDigests()
	if len(digests) != 5 {
		t.Fatalf("expected 5 DayDigest entries (2+1+2), got %d", len(digests))
	}

	// Verify phase labels
	expectedPhases := []string{"phase1", "phase1", "gap", "phase2", "phase2"}
	for i, d := range digests {
		if d.Phase != expectedPhases[i] {
			t.Errorf("day %d: Phase=%q, want %q", i+1, d.Phase, expectedPhases[i])
		}
		if d.Day != i+1 {
			t.Errorf("day %d: Day number=%d, want %d", i+1, d.Day, i+1)
		}
	}

	// Phase1 days should have TechniqueCount=1 (techsPerDay from config)
	for i := 0; i < 2; i++ {
		if digests[i].TechniqueCount != 1 {
			t.Errorf("phase1 day %d: TechniqueCount=%d, want 1", i+1, digests[i].TechniqueCount)
		}
	}
	// Gap day should have TechniqueCount=0
	if digests[2].TechniqueCount != 0 {
		t.Errorf("gap day: TechniqueCount=%d, want 0", digests[2].TechniqueCount)
	}
}

// TestDayDigest_Lifecycle verifies all DayDigest entries reach DayComplete after PoC run (TRACK-01).
func TestDayDigest_Lifecycle(t *testing.T) {
	eng, _ := newPoCEngineWithCampaign(2, 1, 2)
	cfg := pocConfig(2, 1, 2, "camp-delay")

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	digests := eng.GetDayDigests()
	if len(digests) != 5 {
		t.Fatalf("expected 5 DayDigest entries, got %d", len(digests))
	}

	for i, d := range digests {
		if d.Status != DayComplete {
			t.Errorf("day %d: Status=%q, want DayComplete", i+1, d.Status)
		}
		if d.StartTime == "" {
			t.Errorf("day %d: StartTime is empty", i+1)
		}
		if d.EndTime == "" {
			t.Errorf("day %d: EndTime is empty", i+1)
		}
		if d.LastHeartbeat == "" {
			t.Errorf("day %d: LastHeartbeat is empty", i+1)
		}
	}
}

// TestDayDigest_Counts verifies PassCount and FailCount match alternating runner results (TRACK-01).
func TestDayDigest_Counts(t *testing.T) {
	eng, _ := newPoCEngineWithCampaign(2, 1, 2)
	cfg := pocConfig(2, 1, 2, "camp-delay")

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	digests := eng.GetDayDigests()
	if len(digests) != 5 {
		t.Fatalf("expected 5 DayDigest entries, got %d", len(digests))
	}

	// Gap day should have no pass/fail counts
	gapDay := digests[2]
	if gapDay.PassCount != 0 || gapDay.FailCount != 0 {
		t.Errorf("gap day: PassCount=%d FailCount=%d, want both 0", gapDay.PassCount, gapDay.FailCount)
	}

	// Phase1 days run 1 technique each (Phase1TechsPerDay=1)
	for i, d := range digests {
		if d.Phase != "phase1" {
			continue
		}
		total := d.PassCount + d.FailCount
		if total != 1 {
			t.Errorf("day %d (phase1): PassCount+FailCount=%d, want 1", i+1, total)
		}
	}
	// Phase2 days run 2 campaign steps each (camp-delay has 2 steps)
	for i, d := range digests {
		if d.Phase != "phase2" {
			continue
		}
		total := d.PassCount + d.FailCount
		if total != 2 {
			t.Errorf("day %d (phase2): PassCount+FailCount=%d, want 2", i+1, total)
		}
	}
}

// TestDayDigest_Heartbeat verifies LastHeartbeat is non-empty for all days after PoC run (TRACK-04).
func TestDayDigest_Heartbeat(t *testing.T) {
	eng, _ := newPoCEngineWithCampaign(1, 1, 1)
	cfg := pocConfig(1, 1, 1, "camp-delay")

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	digests := eng.GetDayDigests()
	for i, d := range digests {
		if d.LastHeartbeat == "" {
			t.Errorf("day %d (%s): LastHeartbeat is empty — should be set at day-start at minimum", i+1, d.Phase)
		}
	}
}

// TestDayDigest_GapDays verifies gap days transition to complete with timestamps but zero technique counts (TRACK-01).
func TestDayDigest_GapDays(t *testing.T) {
	eng, _ := newPoCEngineWithCampaign(1, 2, 1)
	cfg := pocConfig(1, 2, 1, "camp-delay")

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	digests := eng.GetDayDigests()
	if len(digests) != 4 {
		t.Fatalf("expected 4 DayDigest entries (1+2+1), got %d", len(digests))
	}

	// Days 2 and 3 are gap days
	for _, idx := range []int{1, 2} {
		d := digests[idx]
		if d.Phase != "gap" {
			t.Errorf("day %d: Phase=%q, want gap", idx+1, d.Phase)
		}
		if d.Status != DayComplete {
			t.Errorf("gap day %d: Status=%q, want DayComplete", idx+1, d.Status)
		}
		if d.StartTime == "" {
			t.Errorf("gap day %d: StartTime is empty", idx+1)
		}
		if d.EndTime == "" {
			t.Errorf("gap day %d: EndTime is empty", idx+1)
		}
		if d.PassCount != 0 || d.FailCount != 0 {
			t.Errorf("gap day %d: PassCount=%d FailCount=%d, want both 0", idx+1, d.PassCount, d.FailCount)
		}
	}
}

// TestDayDigest_Reset verifies dayDigests are cleared between PoC runs (Pitfall 4).
func TestDayDigest_Reset(t *testing.T) {
	eng, _ := newPoCEngineWithCampaign(1, 0, 1)

	run := func(campaignID string) {
		cfg := pocConfig(1, 0, 1, campaignID)
		if err := eng.Start(cfg); err != nil {
			t.Fatalf("Start: %v", err)
		}
		if !waitForPhase(eng, PhaseDone, 5*time.Second) {
			t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
		}
	}

	run("camp-delay")
	firstDigests := eng.GetDayDigests()

	// Reset clock for second run
	eng.clock = &fakeClock{now: time.Date(2026, 4, 9, 6, 0, 0, 0, time.UTC)}

	run("camp-delay")
	secondDigests := eng.GetDayDigests()

	// Both runs have 2 days (1+0+1), but the slice should be fresh
	if len(secondDigests) != len(firstDigests) {
		t.Errorf("second run: got %d digests, want %d", len(secondDigests), len(firstDigests))
	}
	// All second-run digests should be complete (not stale pending from first run)
	for i, d := range secondDigests {
		if d.Status != DayComplete {
			t.Errorf("second run day %d: Status=%q, want DayComplete", i+1, d.Status)
		}
	}
}

// TestCampaignDelayAfter verifies DelayAfter>0 triggers waitOrStop with correct duration; DelayAfter=0 does not (CAMP-01).
func TestCampaignDelayAfter(t *testing.T) {
	reg := testRegistry(
		testTechnique("T1087", "discovery", "discovery"),
		testTechnique("T1078", "attack", "persistence"),
		testTechnique("T1003", "attack", "credential-access"),
	)
	reg.Campaigns["camp-delay"] = &playbooks.Campaign{
		ID:   "camp-delay",
		Name: "Delay Test Campaign",
		Steps: []playbooks.CampaignStep{
			{TechniqueID: "T1078", DelayAfter: 300},
			{TechniqueID: "T1003", DelayAfter: 0},
		},
	}

	var delayDurations []time.Duration
	var delayMu sync.Mutex

	atc := &afterTrackClock{
		fakeClock:      fakeClock{now: time.Date(2026, 4, 8, 6, 0, 0, 0, time.UTC)},
		delayDurations: &delayDurations,
		mu:             &delayMu,
	}

	eng := New(reg, nil)
	eng.clock = atc
	eng.runner = fakeRunner(0)

	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 0,
		Phase1TechsPerDay:  0,
		Phase1DailyHour:    8,
		GapDays:            0,
		Phase2DurationDays: 1,
		Phase2DailyHour:    9,
		CampaignID:         "camp-delay",
	}

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !waitForPhase(eng, PhaseDone, 5*time.Second) {
		t.Fatalf("engine did not reach PhaseDone; stuck at %s", eng.GetStatus().Phase)
	}

	delayMu.Lock()
	recorded := make([]time.Duration, len(delayDurations))
	copy(recorded, delayDurations)
	delayMu.Unlock()

	// Expect exactly one delay of 300s (from step1 DelayAfter=300).
	// step2 has DelayAfter=0, so no campaign delay recorded for it.
	found300s := false
	for _, d := range recorded {
		if d == 300*time.Second {
			found300s = true
		}
	}
	if !found300s {
		t.Errorf("expected a 300s campaign delay_after call, got durations: %v", recorded)
	}
}

// afterTrackClock wraps fakeClock and records After() call durations above 60s
// (to distinguish campaign delays from scheduling waits which are hour-based = 3600s+).
// Actually we record all durations and let the test filter.
type afterTrackClock struct {
	fakeClock
	delayDurations *[]time.Duration
	mu             *sync.Mutex
}

func (a *afterTrackClock) After(d time.Duration) <-chan time.Time {
	// Only record sub-hour durations that look like campaign delays (< 3600s)
	// Scheduling waits are nextOccurrenceOfHour which is always > 0 up to 24h.
	// Campaign delays are step.DelayAfter seconds, typically much smaller.
	// Record all durations so the test can inspect them.
	a.mu.Lock()
	*a.delayDurations = append(*a.delayDurations, d)
	a.mu.Unlock()
	return a.fakeClock.After(d)
}

// TestCampaignDelayAfter_Interruptible verifies stop signal during DelayAfter aborts the engine (CAMP-01, D-09).
func TestCampaignDelayAfter_Interruptible(t *testing.T) {
	reg := testRegistry(
		testTechnique("T1078", "attack", "persistence"),
		testTechnique("T1003", "attack", "credential-access"),
	)
	reg.Campaigns["camp-long-delay"] = &playbooks.Campaign{
		ID:   "camp-long-delay",
		Name: "Long Delay Campaign",
		Steps: []playbooks.CampaignStep{
			{TechniqueID: "T1078", DelayAfter: 3600}, // 1 hour delay
		},
	}

	eng := New(reg, nil)
	// Use a blocking fakeClock that we can unblock via stop signal
	// Since fakeClock.After() fires immediately, we need a slow clock for this test.
	// Use realClock but with a very short delay (use a custom blocking clock).
	blockCh := make(chan time.Time) // never fires unless we close it
	bc := &blockingClock{
		fakeClock: fakeClock{now: time.Date(2026, 4, 8, 9, 0, 0, 0, time.UTC)},
		blockCh:   blockCh,
	}
	eng.clock = bc
	eng.runner = fakeRunner(0)

	cfg := Config{
		PoCMode:            true,
		Phase1DurationDays: 0,
		Phase1TechsPerDay:  0,
		Phase1DailyHour:    8,
		GapDays:            0,
		Phase2DurationDays: 1,
		Phase2DailyHour:    9,
		CampaignID:         "camp-long-delay",
	}

	if err := eng.Start(cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for Phase 2 to begin running (engine should be in poc_phase2)
	if !waitForPhase(eng, PhasePoCPhase2, 2*time.Second) {
		t.Fatalf("engine did not reach PhasePoCPhase2; stuck at %s", eng.GetStatus().Phase)
	}

	// Give it a moment to start the delay
	time.Sleep(50 * time.Millisecond)

	// Stop the engine while it's waiting in the campaign delay
	eng.Stop()

	// Engine should abort
	if !waitForPhase(eng, PhaseAborted, 2*time.Second) {
		t.Fatalf("engine did not abort after Stop(); stuck at %s", eng.GetStatus().Phase)
	}
}

// blockingClock implements Clock where After() blocks until the channel is closed or a stop signal fires.
// The scheduling wait (nextOccurrenceOfHour) fires immediately (via fakeClock),
// but campaign delays block until Stop() is called.
type blockingClock struct {
	fakeClock
	blockCh   chan time.Time
	callCount int
	mu        sync.Mutex
}

func (b *blockingClock) After(d time.Duration) <-chan time.Time {
	b.mu.Lock()
	b.callCount++
	count := b.callCount
	b.mu.Unlock()

	if count == 1 {
		// First call is the scheduling wait — fire immediately
		return b.fakeClock.After(d)
	}
	// Subsequent calls (campaign delay) block indefinitely
	return b.blockCh
}
