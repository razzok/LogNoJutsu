package reporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lognojutsu/internal/playbooks"
)

// makeResult builds a minimal ExecutionResult for testing.
func makeResult(status playbooks.VerificationStatus, events []playbooks.VerifiedEvent) playbooks.ExecutionResult {
	return playbooks.ExecutionResult{
		TechniqueID:        "T1059",
		TechniqueName:      "Command Scripting",
		TacticID:           "execution",
		StartTime:          time.Now().Format(time.RFC3339),
		EndTime:            time.Now().Format(time.RFC3339),
		Success:            true,
		RunAsUser:          "testuser",
		VerificationStatus: status,
		VerifiedEvents:     events,
	}
}

// saveHTMLToDir calls SaveResults with a temp directory as the working dir
// and returns the HTML content of the generated file.
func saveHTMLToDir(t *testing.T, results []playbooks.ExecutionResult) string {
	t.Helper()
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to temp: %v", err)
	}
	defer func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("chdir back: %v", err)
		}
	}()

	htmlPath := SaveResults(results, "test.log", false)
	if htmlPath == "" {
		t.Fatal("SaveResults returned empty path — HTML not generated")
	}

	data, err := os.ReadFile(filepath.Join(dir, htmlPath))
	if err != nil {
		t.Fatalf("read html file: %v", err)
	}
	return string(data)
}

// TestHTMLVerificationColumn checks that a pass result produces verif-pass CSS class and "Pass" text.
func TestHTMLVerificationColumn(t *testing.T) {
	results := []playbooks.ExecutionResult{
		makeResult(playbooks.VerifPass, []playbooks.VerifiedEvent{
			{EventID: 4688, Channel: "Security", Description: "Process create", Found: true, Count: 1},
		}),
	}
	html := saveHTMLToDir(t, results)

	checks := []string{
		"verif-pass",
		"Pass",
		"Verifikation",
		"EID 4688",
		"Security",
	}
	for _, want := range checks {
		if !strings.Contains(html, want) {
			t.Errorf("expected HTML to contain %q", want)
		}
	}
}

// TestHTMLVerificationFail checks that a fail result produces verif-fail CSS class and "Fail" text.
func TestHTMLVerificationFail(t *testing.T) {
	results := []playbooks.ExecutionResult{
		makeResult(playbooks.VerifFail, []playbooks.VerifiedEvent{
			{EventID: 4688, Channel: "Security", Description: "Process create", Found: false, Count: 0},
		}),
	}
	html := saveHTMLToDir(t, results)

	if !strings.Contains(html, "verif-fail") {
		t.Error("expected HTML to contain 'verif-fail'")
	}
	if !strings.Contains(html, "Fail") {
		t.Error("expected HTML to contain 'Fail'")
	}
}

// TestHTMLVerificationNotExecuted checks that not_executed status shows "Nicht ausgeführt".
func TestHTMLVerificationNotExecuted(t *testing.T) {
	results := []playbooks.ExecutionResult{
		makeResult(playbooks.VerifNotExecuted, nil),
	}
	html := saveHTMLToDir(t, results)

	want := "Nicht ausgeführt"
	if !strings.Contains(html, want) {
		t.Errorf("expected HTML to contain %q", want)
	}
	if !strings.Contains(html, "verif-skip") {
		t.Error("expected HTML to contain 'verif-skip'")
	}
}

// TestHTMLCrowdStrikeColumn checks that a CrowdStrike coverage column renders
// conditionally based on whether any result has siem_coverage.crowdstrike populated.
func TestHTMLCrowdStrikeColumn(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		result := makeResult(playbooks.VerifPass, nil)
		result.SIEMCoverage = map[string][]string{
			"crowdstrike": {"Credential Dumping", "Suspicious LSASS Access"},
		}
		html := saveHTMLToDir(t, []playbooks.ExecutionResult{result})

		checks := []string{
			"CrowdStrike",           // column header
			"cs-badge",              // badge CSS class
			"CS",                    // badge text
			"Credential Dumping",    // detection rule name
			"Suspicious LSASS Access",
		}
		for _, want := range checks {
			if !strings.Contains(html, want) {
				t.Errorf("expected HTML to contain %q", want)
			}
		}
	})

	t.Run("absent", func(t *testing.T) {
		result := makeResult(playbooks.VerifPass, nil)
		// No SIEMCoverage set
		html := saveHTMLToDir(t, []playbooks.ExecutionResult{result})

		unwanted := []string{"CrowdStrike", "cs-badge"}
		for _, bad := range unwanted {
			if strings.Contains(html, bad) {
				t.Errorf("HTML should NOT contain %q when no CrowdStrike coverage", bad)
			}
		}
	})

	t.Run("na_cell", func(t *testing.T) {
		withCS := makeResult(playbooks.VerifPass, nil)
		withCS.TechniqueID = "FALCON_lsass"
		withCS.SIEMCoverage = map[string][]string{
			"crowdstrike": {"Credential Dumping"},
		}
		withoutCS := makeResult(playbooks.VerifPass, nil)
		withoutCS.TechniqueID = "T1016"
		// No SIEMCoverage

		html := saveHTMLToDir(t, []playbooks.ExecutionResult{withCS, withoutCS})

		if !strings.Contains(html, "cs-badge") {
			t.Error("expected cs-badge for technique with CrowdStrike mapping")
		}
		if !strings.Contains(html, "cs-na") {
			t.Error("expected cs-na class for technique without CrowdStrike mapping")
		}
		if !strings.Contains(html, "N/A") {
			t.Error("expected N/A text for technique without CrowdStrike mapping")
		}
	})
}

// TestHTMLVerificationEventList checks per-event checkmark/X rendering.
func TestHTMLVerificationEventList(t *testing.T) {
	results := []playbooks.ExecutionResult{
		makeResult(playbooks.VerifFail, []playbooks.VerifiedEvent{
			{EventID: 4688, Channel: "Security", Found: true, Count: 2},
			{EventID: 1, Channel: "Sysmon", Found: false, Count: 0},
		}),
	}
	html := saveHTMLToDir(t, results)

	// Both EIDs should appear
	if !strings.Contains(html, "EID 4688") {
		t.Error("expected HTML to contain 'EID 4688'")
	}
	if !strings.Contains(html, "EID 1") {
		t.Error("expected HTML to contain 'EID 1'")
	}
	// Checkmark (&#10003;) for found, X (&#10007;) for not found
	if !strings.Contains(html, "✓") && !strings.Contains(html, "&#10003;") {
		t.Error("expected HTML to contain checkmark for found event")
	}
	if !strings.Contains(html, "✗") && !strings.Contains(html, "&#10007;") {
		t.Error("expected HTML to contain X for missing event")
	}
	if !strings.Contains(html, "verif-list") {
		t.Error("expected HTML to contain 'verif-list' CSS class")
	}
}
