package playbooks

import (
	"strings"
	"testing"
)

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
	if len(reg.Techniques) < 54 {
		t.Errorf("expected at least 54 techniques, got %d", len(reg.Techniques))
	}
	required := []string{"T1005", "T1560.001", "T1119", "T1071.001", "T1071.004", "FALCON_process_injection", "FALCON_lsass_access", "FALCON_lateral_movement_psexec", "AZURE_kerberoasting", "AZURE_ldap_recon", "AZURE_dcsync"}
	for _, id := range required {
		if _, ok := reg.Techniques[id]; !ok {
			t.Errorf("missing required new ATT&CK technique: %s", id)
		}
	}
}

func TestFalconTechniques(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() failed: %v", err)
	}

	falconIDs := []string{"FALCON_process_injection", "FALCON_lsass_access", "FALCON_lateral_movement_psexec"}
	invalidTactics := map[string]bool{"crowdstrike-falcon": true, "falcon": true, "crowdstrike": true}

	for _, id := range falconIDs {
		tech, ok := reg.Techniques[id]
		if !ok {
			t.Errorf("missing FALCON technique: %s", id)
			continue
		}
		if len(tech.ExpectedEvents) == 0 {
			t.Errorf("%s has no expected_events", id)
		}
		cs := tech.SIEMCoverage["crowdstrike"]
		if len(cs) == 0 {
			t.Errorf("%s has no siem_coverage.crowdstrike entries", id)
		}
		if invalidTactics[tech.Tactic] {
			t.Errorf("%s uses non-MITRE tactic %q — must use standard MITRE tactic name", id, tech.Tactic)
		}
		if tech.Phase != "attack" {
			t.Errorf("%s phase should be 'attack', got %q", id, tech.Phase)
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

func TestSentinelCoverage(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() failed: %v", err)
	}

	// HIGH confidence Sentinel rule names (verified from Azure/Azure-Sentinel GitHub)
	highConfidence := map[string]string{
		"T1558.003": "Potential Kerberoasting",
		"T1003.001": "Dumping LSASS Process Into a File",
		"T1003.006": "Non Domain Controller Active Directory Replication",
	}
	for techID, ruleName := range highConfidence {
		tech, ok := reg.Techniques[techID]
		if !ok {
			t.Errorf("technique %s not found", techID)
			continue
		}
		sentinel := tech.SIEMCoverage["sentinel"]
		if len(sentinel) == 0 {
			t.Errorf("%s should have non-empty siem_coverage.sentinel", techID)
			continue
		}
		found := false
		for _, name := range sentinel {
			if name == ruleName {
				found = true
			}
		}
		if !found {
			t.Errorf("%s siem_coverage.sentinel should contain %q, got %v", techID, ruleName, sentinel)
		}
	}

	// MEDIUM confidence — just check non-empty
	mediumConfidence := []string{"T1059.001", "T1136.001"}
	for _, techID := range mediumConfidence {
		tech, ok := reg.Techniques[techID]
		if !ok {
			t.Errorf("technique %s not found", techID)
			continue
		}
		if len(tech.SIEMCoverage["sentinel"]) == 0 {
			t.Errorf("%s should have non-empty siem_coverage.sentinel", techID)
		}
	}

	// Discovery technique should NOT have sentinel coverage
	disc, ok := reg.Techniques["T1016"]
	if !ok {
		t.Fatal("T1016 not found")
	}
	if len(disc.SIEMCoverage["sentinel"]) != 0 {
		t.Errorf("T1016 (discovery) should have no siem_coverage.sentinel, got %v", disc.SIEMCoverage["sentinel"])
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

func TestAzureTechniques(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() failed: %v", err)
	}

	azureIDs := []string{"AZURE_kerberoasting", "AZURE_ldap_recon", "AZURE_dcsync"}
	invalidTactics := map[string]bool{"sentinel": true, "azure": true, "microsoft": true, "azure-ad": true}

	for _, id := range azureIDs {
		tech, ok := reg.Techniques[id]
		if !ok {
			t.Errorf("missing AZURE technique: %s", id)
			continue
		}
		if len(tech.ExpectedEvents) == 0 {
			t.Errorf("%s has no expected_events", id)
		}
		sentinel := tech.SIEMCoverage["sentinel"]
		if len(sentinel) == 0 {
			t.Errorf("%s has no siem_coverage.sentinel entries", id)
		}
		if invalidTactics[tech.Tactic] {
			t.Errorf("%s uses non-MITRE tactic %q — must use standard MITRE tactic name", id, tech.Tactic)
		}
		if tech.Phase != "attack" {
			t.Errorf("%s phase should be 'attack', got %q", id, tech.Phase)
		}
	}

	// Verify specific HIGH confidence rule names
	if tech, ok := reg.Techniques["AZURE_kerberoasting"]; ok {
		found := false
		for _, name := range tech.SIEMCoverage["sentinel"] {
			if name == "Potential Kerberoasting" {
				found = true
			}
		}
		if !found {
			t.Error("AZURE_kerberoasting sentinel should contain 'Potential Kerberoasting'")
		}
	}
	if tech, ok := reg.Techniques["AZURE_dcsync"]; ok {
		found := false
		for _, name := range tech.SIEMCoverage["sentinel"] {
			if name == "Non Domain Controller Active Directory Replication" {
				found = true
			}
		}
		if !found {
			t.Error("AZURE_dcsync sentinel should contain 'Non Domain Controller Active Directory Replication'")
		}
	}
}

func TestTierClassified(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() failed: %v", err)
	}
	for id, tech := range reg.Techniques {
		if tech.Tier < 1 || tech.Tier > 3 {
			t.Errorf("technique %q (%s) has invalid tier %d — must be 1, 2, or 3", id, tech.Name, tech.Tier)
		}
	}
}

func TestWriteArtifactsHaveCleanup(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() failed: %v", err)
	}
	// Techniques that write persistent artifacts (disk, registry, scheduled tasks, services)
	// and therefore require a cleanup command.
	// Excluded from this list:
	//   T1059.001 — no persistent artifacts written; only PowerShell invocation patterns
	//   T1550.002 — inline self-cleanup (cmdkey /delete, net use /delete) within the command body
	//   T1070.001 — cleanup added in Plan 02 (custom log channel rewrite)
	writeArtifacts := map[string]bool{
		"T1053.005": true, "T1543.003": true, "T1547.001": true,
		"T1546.003": true, "T1562.002": true, "T1548.002": true,
		"T1036.005": true, "T1134.001": true, "T1574.002": true,
		"T1490": true, "T1027": true,
		"T1005": true, "T1560.001": true, "T1119": true,
		"T1041": true, "T1047": true, "T1021.001": true,
		"T1021.002": true, "T1136.001": true,
		"T1486": true,
		"UEBA-DATA-STAGING": true, "UEBA-LATERAL-NEW-ASSET": true,
		"FALCON_process_injection": true, "FALCON_lsass_access": true,
		"FALCON_lateral_movement_psexec": true,
	}
	for id, tech := range reg.Techniques {
		if writeArtifacts[id] && strings.TrimSpace(tech.Cleanup) == "" {
			t.Errorf("technique %q (%s) writes artifacts but has empty cleanup", id, tech.Name)
		}
	}
}
