package playbooks

// Technique represents a single MITRE ATT&CK technique definition.
type Technique struct {
	ID                string            `yaml:"id"                 json:"id"`
	Name              string            `yaml:"name"               json:"name"`
	Description       string            `yaml:"description"        json:"description"`
	Tactic            string            `yaml:"tactic"             json:"tactic"`
	TechniqueID       string            `yaml:"technique_id"       json:"technique_id"`
	Platform          string            `yaml:"platform"           json:"platform"`
	Phase             string            `yaml:"phase"              json:"phase"`
	Executor          Executor          `yaml:"executor"           json:"executor"`
	Cleanup           string            `yaml:"cleanup"            json:"cleanup"`
	ExpectedEvents    []string          `yaml:"expected_events"    json:"expected_events"`
	ElevationRequired bool              `yaml:"elevation_required" json:"elevation_required"`
	Tags              []string          `yaml:"tags"               json:"tags"`
	InputArgs         map[string]string `yaml:"input_args"         json:"input_args,omitempty"`
	NistControls      []string          `yaml:"nist_controls"      json:"nist_controls,omitempty"`
}

type Executor struct {
	Type    string `yaml:"type"    json:"type"`
	Command string `yaml:"command" json:"command"`
}

// Campaign is an ordered list of techniques forming an attack scenario.
type Campaign struct {
	ID          string         `yaml:"id"           json:"id"`
	Name        string         `yaml:"name"         json:"name"`
	Description string         `yaml:"description"  json:"description"`
	Industry    string         `yaml:"industry"     json:"industry"`
	ThreatActor string         `yaml:"threat_actor" json:"threat_actor"`
	Steps       []CampaignStep `yaml:"steps"        json:"steps"`
	Tags        []string       `yaml:"tags"         json:"tags"`
}

type CampaignStep struct {
	TechniqueID string `yaml:"technique_id" json:"technique_id"`
	DelayAfter  int    `yaml:"delay_after"  json:"delay_after"`
	Optional    bool   `yaml:"optional"     json:"optional"`
}

// ExecutionResult records the output of a single technique run.
type ExecutionResult struct {
	TechniqueID   string `json:"technique_id"`
	TechniqueName string `json:"technique_name"`
	TacticID      string `json:"tactic_id"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Success       bool   `json:"success"`
	Output        string `json:"output"`
	ErrorOutput   string `json:"error_output"`
	CleanupRun    bool   `json:"cleanup_run"`
	RunAsUser     string `json:"run_as_user"` // "current user" or "DOMAIN\username"
}
