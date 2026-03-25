package playbooks

import "testing"

func TestExpectedEvents(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() failed: %v", err)
	}
	for id, tech := range reg.Techniques {
		if len(tech.ExpectedEvents) == 0 {
			t.Errorf("technique %q (%s) has no expected_events", id, tech.Name)
		}
	}
}

func TestNewTechniqueCount(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() failed: %v", err)
	}
	if len(reg.Techniques) < 48 {
		t.Errorf("expected at least 48 techniques, got %d", len(reg.Techniques))
	}
	required := []string{"T1005", "T1560.001", "T1119", "T1071.001", "T1071.004"}
	for _, id := range required {
		if _, ok := reg.Techniques[id]; !ok {
			t.Errorf("missing required new ATT&CK technique: %s", id)
		}
	}
}

func TestSIEMCoverage(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() failed: %v", err)
	}
	// T1059.001 should have CrowdStrike mappings after YAML population
	ps, ok := reg.Techniques["T1059.001"]
	if !ok {
		t.Fatal("T1059.001 not found in registry")
	}
	cs := ps.SIEMCoverage["crowdstrike"]
	if len(cs) == 0 {
		t.Error("T1059.001 should have non-empty siem_coverage.crowdstrike")
	}
	found := false
	for _, name := range cs {
		if name == "Suspicious Scripts and Commands" {
			found = true
		}
	}
	if !found {
		t.Error("T1059.001 siem_coverage.crowdstrike should contain 'Suspicious Scripts and Commands'")
	}

	// A technique with no siem_coverage should have nil map
	disc, ok := reg.Techniques["T1016"]
	if !ok {
		t.Fatal("T1016 not found in registry")
	}
	if len(disc.SIEMCoverage) != 0 {
		t.Errorf("T1016 (discovery) should have no siem_coverage, got %v", disc.SIEMCoverage)
	}
}

func TestNewUEBACount(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() failed: %v", err)
	}
	uebaCount := 0
	for _, tech := range reg.Techniques {
		if tech.Tactic == "ueba-scenario" {
			uebaCount++
		}
	}
	if uebaCount < 7 {
		t.Errorf("expected at least 7 UEBA scenarios, got %d", uebaCount)
	}
	required := []string{"UEBA-DATA-STAGING", "UEBA-ACCOUNT-TAKEOVER", "UEBA-PRIV-ESC", "UEBA-LATERAL-NEW-ASSET"}
	for _, id := range required {
		if _, ok := reg.Techniques[id]; !ok {
			t.Errorf("missing required new UEBA scenario: %s", id)
		}
	}
}
