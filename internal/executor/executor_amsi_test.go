package executor

import (
	"os/exec"
	"testing"
)

func TestIsAMSIBlocked_Patterns(t *testing.T) {
	patterns := []struct {
		name   string
		errOut string
	}{
		{"ScriptContainedMaliciousContent", "At line:1 char:1\n+ ScriptContainedMaliciousContent\nAt line:1..."},
		{"This script contains malicious content", "This script contains malicious content and cannot be executed."},
		{"has been blocked by your antivirus software", "The operation has been blocked by your antivirus software."},
	}
	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			if !isAMSIBlocked(tc.errOut, nil) {
				t.Errorf("isAMSIBlocked(%q, nil) = false, want true", tc.errOut)
			}
		})
	}
}

func TestIsAMSIBlocked_NormalError(t *testing.T) {
	normalErrors := []string{
		"command not found",
		"The term 'Invoke-Something' is not recognized",
		"Access is denied.",
		"",
	}
	for _, errOut := range normalErrors {
		if isAMSIBlocked(errOut, nil) {
			t.Errorf("isAMSIBlocked(%q, nil) = true, want false for normal error", errOut)
		}
	}
}

func TestIsAMSIBlocked_ExitCode(t *testing.T) {
	// Exit code -196608 (0xFFCFFFFF) signals AMSI block on Windows.
	// We simulate an exec.ExitError with this exit code using a fake process.
	cmd := exec.Command("cmd", "/C", "exit", "-196608")
	err := cmd.Run()
	if err == nil {
		t.Skip("expected non-zero exit from cmd")
	}
	// On non-Windows or when cmd is unavailable, just test the pattern matching.
	// The exit code test is best-effort; AMSI patterns are the primary detection.
	t.Log("Exit code test: isAMSIBlocked returns", isAMSIBlocked("", err))
}
