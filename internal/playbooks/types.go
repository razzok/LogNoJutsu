package playbooks

// Technique represents a single MITRE ATT&CK technique definition.
type Technique struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Tactic      string            `yaml:"tactic"`
	TechniqueID string            `yaml:"technique_id"`
	Platform    string            `yaml:"platform"`
	Phase       string            `yaml:"phase"` // "discovery" or "attack"
	Executor    Executor          `yaml:"executor"`
	Cleanup     string            `yaml:"cleanup"`
	ExpectedEvents []string       `yaml:"expected_events"`
	ElevationRequired bool        `yaml:"elevation_required"`
	Tags        []string          `yaml:"tags"`
	InputArgs   map[string]string `yaml:"input_args"`
}

type Executor struct {
	Type    string `yaml:"type"` // "powershell" or "cmd"
	Command string `yaml:"command"`
}

// Campaign is an ordered list of techniques forming an attack scenario.
type Campaign struct {
	ID          string           `yaml:"id"`
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Industry    string           `yaml:"industry"`
	ThreatActor string           `yaml:"threat_actor"`
	Steps       []CampaignStep   `yaml:"steps"`
	Tags        []string         `yaml:"tags"`
}

type CampaignStep struct {
	TechniqueID string `yaml:"technique_id"` // references Technique.ID
	DelayAfter  int    `yaml:"delay_after"`  // seconds to wait after this step
	Optional    bool   `yaml:"optional"`
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
