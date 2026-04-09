package engine

import (
	"sync"
	"testing"
	"time"

	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/userstore"
)

// scanTechnique builds a minimal Technique with RequiresConfirmation set.
func scanTechnique(id string, requiresConfirmation bool) *playbooks.Technique {
	return &playbooks.Technique{
		ID:                   id,
		Name:                 id + " scan test",
		Phase:                "discovery",
		Tactic:               "discovery",
		RequiresConfirmation: requiresConfirmation,
		Executor: playbooks.Executor{
			Type:    "powershell",
			Command: "Write-Output scan",
		},
	}
}

// TestScanConfirmBlocks verifies that the engine does NOT proceed until ConfirmScan() is called.
func TestScanConfirmBlocks(t *testing.T) {
	reg := testRegistry(scanTechnique("T-SCAN-01", true))
	eng := New(reg, nil)

	runnerCalled := false
	var mu sync.Mutex
	eng.SetRunner(func(tech *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
		mu.Lock()
		runnerCalled = true
		mu.Unlock()
		return playbooks.ExecutionResult{TechniqueID: tech.ID, Success: true}
	})
	eng.SetAdmin(true)

	err := eng.Start(Config{})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer eng.Stop()

	// Give the engine goroutine time to reach the scan confirmation gate
	time.Sleep(100 * time.Millisecond)

	// Engine should be blocked — runner should NOT have been called yet
	mu.Lock()
	called := runnerCalled
	mu.Unlock()
	if called {
		t.Error("runner should NOT be called before scan is confirmed")
	}

	// Verify pending info is available
	info := eng.GetScanPending()
	if info == nil {
		t.Fatal("GetScanPending should return ScanInfo when confirmation is pending")
	}
	if len(info.Techniques) == 0 {
		t.Error("ScanInfo.Techniques should list scan techniques")
	}
	if info.TargetSubnet == "" {
		t.Error("ScanInfo.TargetSubnet should not be empty")
	}
	if info.RateLimitNote == "" {
		t.Error("ScanInfo.RateLimitNote should not be empty")
	}
	if info.IDSWarning == "" {
		t.Error("ScanInfo.IDSWarning should not be empty")
	}

	// Confirm the scan — engine should proceed
	if err := eng.ConfirmScan(); err != nil {
		t.Fatalf("ConfirmScan failed: %v", err)
	}

	// Give the engine time to run the technique
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	called = runnerCalled
	mu.Unlock()
	if !called {
		t.Error("runner SHOULD be called after scan confirmation")
	}

	// After confirming, pending info should be cleared
	if eng.GetScanPending() != nil {
		t.Error("GetScanPending should return nil after confirmation")
	}
}

// TestScanConfirmCancel verifies that CancelScan() aborts the simulation.
func TestScanConfirmCancel(t *testing.T) {
	reg := testRegistry(scanTechnique("T-SCAN-02", true))
	eng := New(reg, nil)

	runnerCalled := false
	var mu sync.Mutex
	eng.SetRunner(func(tech *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
		mu.Lock()
		runnerCalled = true
		mu.Unlock()
		return playbooks.ExecutionResult{TechniqueID: tech.ID, Success: true}
	})
	eng.SetAdmin(true)

	err := eng.Start(Config{})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give the engine goroutine time to reach the scan confirmation gate
	time.Sleep(100 * time.Millisecond)

	// Cancel the scan
	if err := eng.CancelScan(); err != nil {
		t.Fatalf("CancelScan failed: %v", err)
	}

	// Give the engine time to abort
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	called := runnerCalled
	mu.Unlock()
	if called {
		t.Error("runner should NOT be called after scan cancel")
	}

	// Engine should be in aborted phase
	status := eng.GetStatus()
	if status.Phase != PhaseAborted {
		t.Errorf("phase should be aborted after cancel, got %q", status.Phase)
	}
}

// TestScanConfirmNoBlockWithoutFlag verifies that techniques without RequiresConfirmation don't block.
func TestScanConfirmNoBlockWithoutFlag(t *testing.T) {
	reg := testRegistry(scanTechnique("T-SCAN-03", false))
	eng := New(reg, nil)

	runnerCalled := make(chan struct{})
	eng.SetRunner(func(tech *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
		close(runnerCalled)
		return playbooks.ExecutionResult{TechniqueID: tech.ID, Success: true}
	})
	eng.SetAdmin(true)

	err := eng.Start(Config{})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer eng.Stop()

	// Runner should be called without needing any confirmation
	select {
	case <-runnerCalled:
		// Good — no blocking
	case <-time.After(2 * time.Second):
		t.Error("runner was NOT called — engine should not block when no RequiresConfirmation techniques")
	}

	// No pending scan info
	if eng.GetScanPending() != nil {
		t.Error("GetScanPending should be nil when no confirmation techniques")
	}
}

// TestScanConfirmWhatIfSkips verifies that WhatIf mode skips the confirmation gate entirely.
func TestScanConfirmWhatIfSkips(t *testing.T) {
	reg := testRegistry(scanTechnique("T-SCAN-04", true))
	eng := New(reg, nil)

	runnerCalled := make(chan struct{}, 1)
	eng.SetRunner(func(tech *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
		select {
		case runnerCalled <- struct{}{}:
		default:
		}
		return playbooks.ExecutionResult{TechniqueID: tech.ID, Success: true}
	})
	eng.SetAdmin(true)

	err := eng.Start(Config{WhatIf: true})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer eng.Stop()

	// WhatIf mode skips both the runner AND the gate — just check no blocking
	// WhatIf bypasses runner entirely; just ensure engine reaches done/finishes
	time.Sleep(200 * time.Millisecond)

	// No pending scan info in WhatIf mode
	if eng.GetScanPending() != nil {
		t.Error("GetScanPending should be nil in WhatIf mode")
	}
}

// TestGetScanPending verifies GetScanPending returns nil when no scan is pending.
func TestGetScanPending(t *testing.T) {
	reg := testRegistry()
	eng := New(reg, nil)

	// No simulation started — should return nil
	if eng.GetScanPending() != nil {
		t.Error("GetScanPending should return nil when idle")
	}
}

// TestConfirmScanNoPending verifies ConfirmScan returns an error when no scan is pending.
func TestConfirmScanNoPending(t *testing.T) {
	reg := testRegistry()
	eng := New(reg, nil)

	err := eng.ConfirmScan()
	if err == nil {
		t.Error("ConfirmScan should return error when no scan pending")
	}

	err = eng.CancelScan()
	if err == nil {
		t.Error("CancelScan should return error when no scan pending")
	}
}
