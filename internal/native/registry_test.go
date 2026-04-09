package native

import (
	"errors"
	"testing"
)

func TestRegisterLookup(t *testing.T) {
	// Register a test function
	fn := func() (NativeResult, error) {
		return NativeResult{Output: "test output", Success: true}, nil
	}
	Register("TEST-001", fn, nil)

	// Lookup registered ID
	got := Lookup("TEST-001")
	if got == nil {
		t.Fatal("Lookup(TEST-001) returned nil, want non-nil")
	}

	// Invoke to verify it's the right function
	result, err := got()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output != "test output" {
		t.Errorf("Output = %q, want %q", result.Output, "test output")
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}

	// Lookup missing ID
	missing := Lookup("MISSING")
	if missing != nil {
		t.Error("Lookup(MISSING) returned non-nil, want nil")
	}
}

func TestRegisterWithCleanup(t *testing.T) {
	cleanupCalled := false
	fn := func() (NativeResult, error) {
		return NativeResult{Success: true}, nil
	}
	cleanup := func() error {
		cleanupCalled = true
		return nil
	}

	Register("TEST-CLEANUP-001", fn, cleanup)

	// LookupCleanup should return non-nil
	got := LookupCleanup("TEST-CLEANUP-001")
	if got == nil {
		t.Fatal("LookupCleanup(TEST-CLEANUP-001) returned nil, want non-nil")
	}

	// Invoke cleanup to verify it works
	if err := got(); err != nil {
		t.Fatalf("cleanup returned error: %v", err)
	}
	if !cleanupCalled {
		t.Error("cleanup was not called")
	}

	// Register same ID with nil cleanup — LookupCleanup should return nil
	Register("TEST-CLEANUP-001", fn, nil)
	nilCleanup := LookupCleanup("TEST-CLEANUP-001")
	if nilCleanup != nil {
		t.Error("LookupCleanup after re-register with nil cleanup returned non-nil")
	}
}

func TestRegisterOverwrite(t *testing.T) {
	firstCalled := false
	secondCalled := false

	first := func() (NativeResult, error) {
		firstCalled = true
		return NativeResult{Output: "first", Success: true}, nil
	}
	second := func() (NativeResult, error) {
		secondCalled = true
		return NativeResult{Output: "second", Success: true}, nil
	}

	Register("TEST-OVERWRITE", first, nil)
	Register("TEST-OVERWRITE", second, nil)

	got := Lookup("TEST-OVERWRITE")
	if got == nil {
		t.Fatal("Lookup(TEST-OVERWRITE) returned nil")
	}

	result, err := got()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !secondCalled {
		t.Error("second function was not called")
	}
	if firstCalled {
		t.Error("first function was called instead of second")
	}
	if result.Output != "second" {
		t.Errorf("Output = %q, want %q", result.Output, "second")
	}
}

func TestLookupCleanupMissing(t *testing.T) {
	// Should not panic, should return nil
	got := LookupCleanup("MISSING-CLEANUP")
	if got != nil {
		t.Error("LookupCleanup(MISSING-CLEANUP) returned non-nil, want nil")
	}
}

// Ensure NativeResult fields are correct types
func TestNativeResultFields(t *testing.T) {
	r := NativeResult{
		Output:      "stdout",
		ErrorOutput: "stderr",
		Success:     true,
	}
	if r.Output != "stdout" {
		t.Errorf("Output = %q", r.Output)
	}
	if r.ErrorOutput != "stderr" {
		t.Errorf("ErrorOutput = %q", r.ErrorOutput)
	}
	if !r.Success {
		t.Error("Success = false")
	}
}

// Ensure CleanupFunc signature is correct (returns error)
func TestCleanupFuncSignature(t *testing.T) {
	var fn CleanupFunc = func() error {
		return errors.New("test error")
	}
	err := fn()
	if err == nil || err.Error() != "test error" {
		t.Errorf("CleanupFunc error = %v", err)
	}
}
