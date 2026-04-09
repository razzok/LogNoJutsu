//go:build windows

package native

import (
	"strings"
	"testing"
)

func TestT1057WMI(t *testing.T) {
	result, err := runT1057()
	if err != nil {
		t.Fatalf("runT1057() returned unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("runT1057() Success = false, ErrorOutput = %q", result.ErrorOutput)
	}
	if len(result.Output) == 0 {
		t.Error("runT1057() Output is empty — expected process list")
	}
	if !strings.Contains(result.Output, "processes found") {
		t.Errorf("runT1057() Output does not contain 'processes found': %q", result.Output)
	}
}

func TestT1057WMIHasResults(t *testing.T) {
	result, err := runT1057()
	if err != nil {
		t.Fatalf("runT1057() returned unexpected error: %v", err)
	}
	// Any Windows machine will have at least one process; verify PID line present
	if !strings.Contains(result.Output, "PID=") {
		t.Errorf("runT1057() Output does not contain any PID lines: %q", result.Output)
	}
}
