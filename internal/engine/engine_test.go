package engine

import (
	"sync"
	"testing"
	"time"

	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/userstore"
)

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
