package verifier

import (
	"errors"
	"testing"
	"time"

	"lognojutsu/internal/playbooks"
)

// mockQueryFn returns a QueryFn that maps (channel, eventID) to a predetermined count.
func mockQueryFn(counts map[int]int) QueryFn {
	return func(channel string, eventID int, since time.Time) (int, error) {
		if c, ok := counts[eventID]; ok {
			return c, nil
		}
		return 0, nil
	}
}

func specs(ids ...int) []playbooks.EventSpec {
	s := make([]playbooks.EventSpec, len(ids))
	for i, id := range ids {
		s[i] = playbooks.EventSpec{
			EventID:     id,
			Channel:     "TestChannel",
			Description: "Test event",
		}
	}
	return s
}

func TestDetermineStatus(t *testing.T) {
	since := time.Now().Add(-1 * time.Minute)

	// All found -> VerifPass
	status, verified := Verify(specs(4688, 4624), since, true, mockQueryFn(map[int]int{4688: 1, 4624: 3}))
	if status != playbooks.VerifPass {
		t.Errorf("all found: expected VerifPass, got %q", status)
	}
	if len(verified) != 2 {
		t.Fatalf("expected 2 verified events, got %d", len(verified))
	}
	for _, v := range verified {
		if !v.Found {
			t.Errorf("event %d should be Found=true", v.EventID)
		}
	}

	// One missing -> VerifFail
	status, _ = Verify(specs(4688, 9999), since, true, mockQueryFn(map[int]int{4688: 1}))
	if status != playbooks.VerifFail {
		t.Errorf("one missing: expected VerifFail, got %q", status)
	}

	// Empty specs -> VerifNotRun
	status, verified = Verify(nil, since, true, mockQueryFn(map[int]int{}))
	if status != playbooks.VerifNotRun {
		t.Errorf("empty specs: expected VerifNotRun, got %q", status)
	}
	if verified != nil {
		t.Errorf("empty specs: expected nil verified, got %v", verified)
	}
}

func TestNotExecutedVsEventsMissing(t *testing.T) {
	since := time.Now().Add(-1 * time.Minute)

	// executionSuccess=false -> VerifNotExecuted
	status, verified := Verify(specs(4688), since, false, mockQueryFn(map[int]int{4688: 5}))
	if status != playbooks.VerifNotExecuted {
		t.Errorf("not executed: expected VerifNotExecuted, got %q", status)
	}
	if verified != nil {
		t.Errorf("not executed: expected nil verified, got %v", verified)
	}

	// executionSuccess=true, all counts=0 -> VerifFail
	status, _ = Verify(specs(4688), since, true, mockQueryFn(map[int]int{}))
	if status != playbooks.VerifFail {
		t.Errorf("all missing: expected VerifFail, got %q", status)
	}
}

func TestVerifyAllFound(t *testing.T) {
	since := time.Now().Add(-1 * time.Minute)
	counts := map[int]int{10: 2, 4688: 5, 4624: 1}

	status, verified := Verify(specs(10, 4688, 4624), since, true, mockQueryFn(counts))
	if status != playbooks.VerifPass {
		t.Errorf("expected VerifPass, got %q", status)
	}
	if len(verified) != 3 {
		t.Fatalf("expected 3 verified events, got %d", len(verified))
	}
	expectedCounts := map[int]int{10: 2, 4688: 5, 4624: 1}
	for _, v := range verified {
		if !v.Found {
			t.Errorf("event %d should be Found=true", v.EventID)
		}
		if v.Count != expectedCounts[v.EventID] {
			t.Errorf("event %d: expected Count=%d, got %d", v.EventID, expectedCounts[v.EventID], v.Count)
		}
	}
}

func TestQueryCountMock(t *testing.T) {
	since := time.Now().Add(-5 * time.Minute)

	var calledChannel string
	var calledEventID int
	var calledSince time.Time

	trackingFn := func(channel string, eventID int, s time.Time) (int, error) {
		calledChannel = channel
		calledEventID = eventID
		calledSince = s
		return 1, nil
	}

	sp := []playbooks.EventSpec{{
		EventID:     4688,
		Channel:     "Security",
		Description: "Process Creation",
	}}

	Verify(sp, since, true, trackingFn)

	if calledChannel != "Security" {
		t.Errorf("expected channel Security, got %q", calledChannel)
	}
	if calledEventID != 4688 {
		t.Errorf("expected eventID 4688, got %d", calledEventID)
	}
	if !calledSince.Equal(since) {
		t.Errorf("since time mismatch: got %v, want %v", calledSince, since)
	}
}

// Ensure QueryFn type is usable (compile-time check).
var _ QueryFn = func(channel string, eventID int, since time.Time) (int, error) {
	return 0, errors.New("unused")
}
