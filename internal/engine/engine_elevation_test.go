package engine

import (
	"testing"

	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/userstore"
)

// TestElevationSkip verifies that techniques requiring elevation are skipped
// (with VerifElevationRequired status) when isAdmin=false.
func TestElevationSkip(t *testing.T) {
	reg := testRegistry(testTechnique("T-ELEV-01", "attack", "persistence"))
	e := New(reg, nil)

	// Track whether runner was called
	runnerCalled := false
	e.SetRunner(func(tech *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
		runnerCalled = true
		return playbooks.ExecutionResult{
			TechniqueID: tech.ID,
			Success:     true,
		}
	})

	e.SetAdmin(false)

	tech := &playbooks.Technique{
		ID:                "T-ELEV-01",
		Name:              "Elevation Test",
		Tactic:            "attack",
		Phase:             "attack",
		ElevationRequired: true,
		Executor: playbooks.Executor{
			Type:    "powershell",
			Command: "Write-Output test",
		},
	}
	e.runTechnique(tech)

	if runnerCalled {
		t.Error("runner should NOT be called when elevation is required and isAdmin=false")
	}

	e.mu.RLock()
	results := e.status.Results
	e.mu.RUnlock()

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.VerificationStatus != playbooks.VerifElevationRequired {
		t.Errorf("VerificationStatus = %q, want %q", r.VerificationStatus, playbooks.VerifElevationRequired)
	}
	if r.Success {
		t.Error("Success should be false for elevation-skipped technique")
	}
}

// TestElevationRun verifies that techniques requiring elevation execute normally
// when isAdmin=true.
func TestElevationRun(t *testing.T) {
	reg := testRegistry(testTechnique("T-ELEV-02", "attack", "persistence"))
	e := New(reg, nil)

	runnerCalled := false
	e.SetRunner(func(tech *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
		runnerCalled = true
		return playbooks.ExecutionResult{
			TechniqueID: tech.ID,
			Success:     true,
		}
	})

	e.SetAdmin(true)

	tech := &playbooks.Technique{
		ID:                "T-ELEV-02",
		Name:              "Elevation Test",
		Tactic:            "attack",
		Phase:             "attack",
		ElevationRequired: true,
		Executor: playbooks.Executor{
			Type:    "powershell",
			Command: "Write-Output test",
		},
	}
	e.runTechnique(tech)

	if !runnerCalled {
		t.Error("runner SHOULD be called when elevation is required and isAdmin=true")
	}
}

// TestElevationNotRequired verifies that techniques without elevation_required
// run regardless of isAdmin value.
func TestElevationNotRequired(t *testing.T) {
	for _, admin := range []bool{true, false} {
		reg := testRegistry(testTechnique("T-ELEV-03", "attack", "persistence"))
		e := New(reg, nil)

		runnerCalled := false
		e.SetRunner(func(tech *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
			runnerCalled = true
			return playbooks.ExecutionResult{
				TechniqueID: tech.ID,
				Success:     true,
			}
		})

		e.SetAdmin(admin)

		tech := &playbooks.Technique{
			ID:                "T-ELEV-03",
			Name:              "No Elevation Test",
			Tactic:            "attack",
			Phase:             "attack",
			ElevationRequired: false,
			Executor: playbooks.Executor{
				Type:    "powershell",
				Command: "Write-Output test",
			},
		}
		e.runTechnique(tech)

		if !runnerCalled {
			t.Errorf("runner SHOULD be called for non-elevation technique (isAdmin=%v)", admin)
		}
	}
}
