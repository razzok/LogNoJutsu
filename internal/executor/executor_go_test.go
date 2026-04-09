package executor

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"lognojutsu/internal/native"
	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/userstore"
)

// makeTechnique builds a minimal Technique with type:go for testing.
func makeTechnique(id string) *playbooks.Technique {
	return &playbooks.Technique{
		ID:   id,
		Name: "Test Go Technique " + id,
		Executor: playbooks.Executor{
			Type:    "go",
			Command: "",
		},
	}
}

func TestGoDispatch(t *testing.T) {
	id := "TEST-GO-DISPATCH-001"
	native.Register(id, func() (native.NativeResult, error) {
		return native.NativeResult{Output: "hello", Success: true}, nil
	}, nil)

	result := Run(makeTechnique(id))

	if !result.Success {
		t.Errorf("Success = false, want true; ErrorOutput = %q", result.ErrorOutput)
	}
	if result.Output != "hello" {
		t.Errorf("Output = %q, want %q", result.Output, "hello")
	}
	if result.ErrorOutput != "" {
		t.Errorf("ErrorOutput = %q, want empty", result.ErrorOutput)
	}
}

func TestGoDispatchUnregistered(t *testing.T) {
	id := "MISSING-GO-999"
	// Ensure not registered
	result := Run(makeTechnique(id))

	if result.Success {
		t.Error("Success = true, want false for unregistered technique")
	}
	if !strings.Contains(result.ErrorOutput, "no native") {
		t.Errorf("ErrorOutput = %q, want to contain %q", result.ErrorOutput, "no native")
	}
}

func TestGoDispatchWithError(t *testing.T) {
	id := "TEST-GO-ERROR-001"
	native.Register(id, func() (native.NativeResult, error) {
		return native.NativeResult{Output: "partial"}, fmt.Errorf("infra failure")
	}, nil)

	result := Run(makeTechnique(id))

	if result.Success {
		t.Error("Success = true, want false when func returns error")
	}
	if !strings.Contains(result.ErrorOutput, "infra failure") {
		t.Errorf("ErrorOutput = %q, want to contain %q", result.ErrorOutput, "infra failure")
	}
}

func TestGoCleanupInRunWithCleanup(t *testing.T) {
	id := "TEST-GO-CLEANUP-001"
	var cleanupCalled atomic.Bool

	native.Register(id, func() (native.NativeResult, error) {
		return native.NativeResult{Output: "ran", Success: true}, nil
	}, func() error {
		cleanupCalled.Store(true)
		return nil
	})

	result := RunWithCleanup(makeTechnique(id), nil, "")

	if !result.Success {
		t.Errorf("Success = false; ErrorOutput = %q", result.ErrorOutput)
	}
	if !result.CleanupRun {
		t.Error("CleanupRun = false, want true")
	}
	if !cleanupCalled.Load() {
		t.Error("cleanup function was not called")
	}
}

func TestGoNoCleanupWhenNil(t *testing.T) {
	id := "TEST-GO-NOCLEANUP-001"
	native.Register(id, func() (native.NativeResult, error) {
		return native.NativeResult{Output: "ran", Success: true}, nil
	}, nil)

	result := RunWithCleanup(makeTechnique(id), nil, "")

	if !result.Success {
		t.Errorf("Success = false; ErrorOutput = %q", result.ErrorOutput)
	}
	if result.CleanupRun {
		t.Error("CleanupRun = true, want false when no cleanup registered")
	}
}

func TestGoRunAsLogNote(t *testing.T) {
	id := "TEST-GO-RUNAS-001"
	native.Register(id, func() (native.NativeResult, error) {
		return native.NativeResult{Success: true}, nil
	}, nil)

	// Create a non-current user profile
	profile := &userstore.UserProfile{
		UserType: "domain",
		Username: "testuser",
		Domain:   "TESTDOMAIN",
	}

	// Should not panic; RunAs info note is logged but execution proceeds
	result := RunAs(makeTechnique(id), profile, "password")

	if !result.Success {
		t.Errorf("Success = false; ErrorOutput = %q", result.ErrorOutput)
	}
}
