package playbooks

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEventSpecParsing(t *testing.T) {
	yamlData := `
id: T1234
name: Test Technique
description: A test
tactic: attack
technique_id: T1234
platform: windows
phase: attack
executor:
  type: powershell
  command: echo test
expected_events:
  - event_id: 10
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "ProcessAccess"
  - event_id: 4688
    channel: "Security"
    description: "Process Creation"
    contains: "mimikatz"
`
	var tech Technique
	if err := yaml.Unmarshal([]byte(yamlData), &tech); err != nil {
		t.Fatalf("failed to unmarshal technique YAML: %v", err)
	}

	if len(tech.ExpectedEvents) != 2 {
		t.Fatalf("expected 2 EventSpec, got %d", len(tech.ExpectedEvents))
	}

	ev := tech.ExpectedEvents[0]
	if ev.EventID != 10 {
		t.Errorf("expected EventID=10, got %d", ev.EventID)
	}
	if ev.Channel != "Microsoft-Windows-Sysmon/Operational" {
		t.Errorf("unexpected Channel: %q", ev.Channel)
	}
	if ev.Description != "ProcessAccess" {
		t.Errorf("unexpected Description: %q", ev.Description)
	}

	ev2 := tech.ExpectedEvents[1]
	if ev2.EventID != 4688 {
		t.Errorf("expected EventID=4688, got %d", ev2.EventID)
	}
	if ev2.Contains != "mimikatz" {
		t.Errorf("expected Contains=mimikatz, got %q", ev2.Contains)
	}
}

func TestEventSpecEmptyList(t *testing.T) {
	yamlData := `
id: T5678
name: No Events Technique
description: Technique with no expected events
tactic: discovery
technique_id: T5678
platform: windows
phase: discovery
executor:
  type: powershell
  command: echo hello
`
	var tech Technique
	if err := yaml.Unmarshal([]byte(yamlData), &tech); err != nil {
		t.Fatalf("failed to unmarshal technique YAML: %v", err)
	}

	if len(tech.ExpectedEvents) != 0 {
		t.Errorf("expected 0 EventSpec, got %d", len(tech.ExpectedEvents))
	}
}

func TestVerificationStatusConstants(t *testing.T) {
	if VerifNotRun != "not_run" {
		t.Errorf("VerifNotRun = %q, want %q", VerifNotRun, "not_run")
	}
	if VerifPass != "pass" {
		t.Errorf("VerifPass = %q, want %q", VerifPass, "pass")
	}
	if VerifFail != "fail" {
		t.Errorf("VerifFail = %q, want %q", VerifFail, "fail")
	}
	if VerifNotExecuted != "not_executed" {
		t.Errorf("VerifNotExecuted = %q, want %q", VerifNotExecuted, "not_executed")
	}
}
