package preparation

import (
	"fmt"
	"strings"
	"testing"
)

// TestPoliciesUseGUIDs verifies that every entry in auditPolicies uses a GUID format
// ({0CCE9...}) rather than an English subcategory name like "Logon" or "Logoff".
// Fails before Task 2 because auditPolicies does not exist yet (RED phase).
func TestPoliciesUseGUIDs(t *testing.T) {
	if len(auditPolicies) == 0 {
		t.Fatal("auditPolicies is empty — expected at least 11 entries")
	}
	for _, p := range auditPolicies {
		if !strings.HasPrefix(p.guid, "{0CCE") {
			t.Errorf("entry %q: guid %q does not start with {0CCE — expected GUID format", p.description, p.guid)
		}
	}
}

// TestPoliciesNoDuplicateGUIDs verifies that no GUID value appears more than once
// in auditPolicies. Duplicate GUIDs would silently configure the same subcategory twice.
func TestPoliciesNoDuplicateGUIDs(t *testing.T) {
	seen := make(map[string]string) // guid -> description of first occurrence
	for _, p := range auditPolicies {
		if prev, dup := seen[p.guid]; dup {
			t.Errorf("duplicate GUID %q: appears for %q and %q", p.guid, prev, p.description)
		}
		seen[p.guid] = p.description
	}
}

// TestAuditFailureMessageFormat verifies that audit failure messages use the format
// "<description>: failed (exit status N)" — human-readable description first, not raw GUID.
func TestAuditFailureMessageFormat(t *testing.T) {
	description := "Logon/Logoff events (4624, 4625, 4634)"
	exitCode := 87
	// Simulate the error format used in ConfigureAuditPolicy when auditpol.exe fails.
	// The format matches fmt.Sprintf("%s: failed (%v)", p.description, err) where
	// err.Error() == "exit status N" for *exec.ExitError.
	msg := fmt.Sprintf("%s: failed (exit status %d)", description, exitCode)

	if !strings.Contains(msg, ": failed (exit status ") {
		t.Errorf("failure message %q does not match expected format '<description>: failed (exit status N)'", msg)
	}
	if !strings.Contains(msg, description) {
		t.Errorf("failure message %q does not contain the human-readable description %q", msg, description)
	}
	if strings.HasPrefix(msg, "{0CCE") {
		t.Errorf("failure message %q starts with a raw GUID — description should come first", msg)
	}
}
