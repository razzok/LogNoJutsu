//go:build windows

package native

import (
	"errors"
	"testing"
)

func TestT1482NoDC(t *testing.T) {
	t.Setenv("LOGONSERVER", "")
	t.Setenv("USERDNSDOMAIN", "")

	_, err := discoverDC()
	if !errors.Is(err, ErrNoDCReachable) {
		t.Errorf("discoverDC() with no env vars: got err=%v, want ErrNoDCReachable", err)
	}
}

func TestT1482DiscoverDCFromLogonserver(t *testing.T) {
	t.Setenv("LOGONSERVER", `\\DC01`)
	t.Setenv("USERDNSDOMAIN", "")

	dc, err := discoverDC()
	if err != nil {
		t.Fatalf("discoverDC() with LOGONSERVER set: unexpected error: %v", err)
	}
	if dc != "DC01:389" {
		t.Errorf("discoverDC() = %q, want %q", dc, "DC01:389")
	}
}

func TestT1482GracefulFallback(t *testing.T) {
	t.Setenv("LOGONSERVER", "")
	t.Setenv("USERDNSDOMAIN", "")

	result, err := runT1482()
	if result.Success {
		t.Error("runT1482() with no DC: expected Success=false, got true")
	}
	if !errors.Is(err, ErrNoDCReachable) {
		t.Errorf("runT1482() with no DC: expected ErrNoDCReachable, got %v", err)
	}
	const want = "no domain controller reachable"
	if result.ErrorOutput == "" {
		t.Error("runT1482() with no DC: ErrorOutput should not be empty")
	}
	found := false
	for i := 0; i < len(result.ErrorOutput)-len(want)+1; i++ {
		if result.ErrorOutput[i:i+len(want)] == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("runT1482() ErrorOutput = %q, want it to contain %q", result.ErrorOutput, want)
	}
}
