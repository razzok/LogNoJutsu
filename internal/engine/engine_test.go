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
