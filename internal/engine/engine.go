package engine

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"lognojutsu/internal/executor"
	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/reporter"
	"lognojutsu/internal/simlog"
	"lognojutsu/internal/userstore"
)

type Phase string

const (
	PhaseIdle      Phase = "idle"
	PhaseDiscovery Phase = "discovery"
	PhaseAttack    Phase = "attack"
	PhaseDone      Phase = "done"
	PhaseAborted   Phase = "aborted"
)

// UserRotation controls how user profiles are assigned to techniques.
type UserRotation string

const (
	RotationNone       UserRotation = "none"       // all techniques as current user
	RotationSequential UserRotation = "sequential" // cycle through profiles in order
	RotationRandom     UserRotation = "random"     // pick a random profile each time
)

// Config holds the engine scheduling and user configuration.
type Config struct {
	DelayBeforeDiscovery int          `json:"delay_before_discovery"`
	DelayBeforeAttack    int          `json:"delay_before_attack"`
	RunCleanup           bool         `json:"run_cleanup"`
	CampaignID           string       `json:"campaign_id"`
	SelectedTechniques   []string     `json:"selected_techniques"`
	UserProfileIDs       []string     `json:"user_profile_ids"` // empty = current user only
	UserRotation         UserRotation `json:"user_rotation"`
}

// Status is the current engine state.
type Status struct {
	Phase           Phase                       `json:"phase"`
	StartTime       string                      `json:"start_time"`
	PhaseStartTimes map[Phase]string            `json:"phase_start_times"`
	CurrentStep     string                      `json:"current_step"`
	CurrentUser     string                      `json:"current_user"`
	Results         []playbooks.ExecutionResult `json:"results"`
	Errors          []string                    `json:"errors"`
	LogFile         string                      `json:"log_file"`
	CleanupDone     bool                        `json:"cleanup_done"`
}

// Engine manages the simulation lifecycle.
type Engine struct {
	mu                 sync.RWMutex
	status             Status
	registry           *playbooks.Registry
	users              *userstore.Store
	cfg                Config
	stopCh             chan struct{}
	executedTechniques []*playbooks.Technique
	// resolved user profiles with decrypted passwords for this run
	resolvedProfiles   []resolvedProfile
	rotationIndex      int
}

type resolvedProfile struct {
	profile  *userstore.UserProfile
	password string
}

func New(registry *playbooks.Registry, users *userstore.Store) *Engine {
	return &Engine{
		registry: registry,
		users:    users,
		status: Status{
			Phase:           PhaseIdle,
			PhaseStartTimes: make(map[Phase]string),
			Results:         []playbooks.ExecutionResult{},
			Errors:          []string{},
		},
		stopCh: make(chan struct{}, 1),
	}
}

func (e *Engine) Start(cfg Config) error {
	e.mu.Lock()
	if e.status.Phase != PhaseIdle && e.status.Phase != PhaseDone && e.status.Phase != PhaseAborted {
		e.mu.Unlock()
		return fmt.Errorf("simulation already running (phase: %s)", e.status.Phase)
	}
	e.cfg = cfg
	e.stopCh = make(chan struct{}, 1)
	e.executedTechniques = nil
	e.rotationIndex = 0
	e.status = Status{
		Phase:           PhaseDiscovery,
		StartTime:       time.Now().Format(time.RFC3339),
		PhaseStartTimes: make(map[Phase]string),
		Results:         []playbooks.ExecutionResult{},
		Errors:          []string{},
	}
	e.mu.Unlock()

	// Resolve user profiles and decrypt passwords once at start
	if err := e.resolveProfiles(); err != nil {
		return fmt.Errorf("resolving user profiles: %w", err)
	}

	simlog.Start(cfg.CampaignID)
	simlog.Info(fmt.Sprintf("Configuration: delay_discovery=%ds delay_attack=%ds cleanup=%v campaign=%q rotation=%s profiles=%d",
		cfg.DelayBeforeDiscovery, cfg.DelayBeforeAttack, cfg.RunCleanup,
		cfg.CampaignID, cfg.UserRotation, len(e.resolvedProfiles)))

	if len(e.resolvedProfiles) > 0 {
		names := make([]string, len(e.resolvedProfiles))
		for i, rp := range e.resolvedProfiles {
			names[i] = rp.profile.QualifiedName()
		}
		simlog.Info(fmt.Sprintf("User profiles for simulation: %v", names))
	} else {
		simlog.Info("Running all techniques as current user (no profiles configured)")
	}

	e.mu.Lock()
	e.status.LogFile = simlog.GetFilePath()
	e.mu.Unlock()

	go e.run()
	return nil
}

func (e *Engine) Stop() {
	select {
	case e.stopCh <- struct{}{}:
	default:
	}
}

func (e *Engine) GetStatus() Status {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// resolveProfiles looks up each configured profile ID and decrypts its password.
func (e *Engine) resolveProfiles() error {
	e.resolvedProfiles = nil
	if len(e.cfg.UserProfileIDs) == 0 || e.cfg.UserRotation == RotationNone {
		return nil
	}
	for _, id := range e.cfg.UserProfileIDs {
		profile, ok := e.users.Get(id)
		if !ok {
			return fmt.Errorf("user profile %q not found", id)
		}
		if !profile.Enabled {
			continue
		}
		password, err := e.users.DecryptPassword(id)
		if err != nil {
			return fmt.Errorf("decrypting password for %s: %w", profile.QualifiedName(), err)
		}
		e.resolvedProfiles = append(e.resolvedProfiles, resolvedProfile{
			profile:  profile,
			password: password,
		})
	}
	return nil
}

// pickUser returns the next user for a technique based on the rotation strategy.
// Returns nil profile (= current user) if no profiles are configured.
func (e *Engine) pickUser() (*userstore.UserProfile, string) {
	if len(e.resolvedProfiles) == 0 {
		return nil, ""
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	var rp resolvedProfile
	switch e.cfg.UserRotation {
	case RotationRandom:
		rp = e.resolvedProfiles[rand.Intn(len(e.resolvedProfiles))]
	default: // RotationSequential
		rp = e.resolvedProfiles[e.rotationIndex%len(e.resolvedProfiles)]
		e.rotationIndex++
	}
	return rp.profile, rp.password
}

func (e *Engine) run() {
	simlog.Info(fmt.Sprintf("Starting simulation. Discovery delay: %ds, Attack delay: %ds",
		e.cfg.DelayBeforeDiscovery, e.cfg.DelayBeforeAttack))

	if e.cfg.DelayBeforeDiscovery > 0 {
		simlog.Info(fmt.Sprintf("Waiting %ds before Discovery phase...", e.cfg.DelayBeforeDiscovery))
		if !e.waitOrStop(time.Duration(e.cfg.DelayBeforeDiscovery) * time.Second) {
			e.abort()
			return
		}
	}

	// Phase 1: Discovery
	e.setPhase(PhaseDiscovery)
	simlog.Phase("discovery")
	discoveryTechniques := e.registry.GetTechniquesByPhase("discovery")
	simlog.Info(fmt.Sprintf("Discovery phase: %d techniques to run", len(discoveryTechniques)))
	for _, t := range discoveryTechniques {
		if e.isStopped() {
			e.abort()
			return
		}
		e.runTechnique(t)
	}

	if e.cfg.DelayBeforeAttack > 0 {
		simlog.Info(fmt.Sprintf("Discovery complete. Waiting %ds before Attack phase...", e.cfg.DelayBeforeAttack))
		if !e.waitOrStop(time.Duration(e.cfg.DelayBeforeAttack) * time.Second) {
			e.abort()
			return
		}
	}

	// Phase 2: Attack
	e.setPhase(PhaseAttack)
	simlog.Phase("attack")
	attackTechniques := e.getTechniquesForPhase()
	simlog.Info(fmt.Sprintf("Attack phase: %d techniques to run", len(attackTechniques)))
	for _, t := range attackTechniques {
		if e.isStopped() {
			e.abort()
			return
		}
		e.runTechnique(t)
	}

	e.finish()
}

func (e *Engine) runTechnique(t *playbooks.Technique) {
	profile, password := e.pickUser()

	userLabel := "current user"
	if profile != nil {
		userLabel = profile.QualifiedName()
	}

	e.mu.Lock()
	e.status.CurrentStep = fmt.Sprintf("%s — %s", t.ID, t.Name)
	e.status.CurrentUser = userLabel
	e.mu.Unlock()

	var result playbooks.ExecutionResult
	if e.cfg.RunCleanup {
		result = executor.RunWithCleanup(t, profile, password)
	} else {
		result = executor.RunAs(t, profile, password)
		if t.Cleanup != "" {
			e.mu.Lock()
			e.executedTechniques = append(e.executedTechniques, t)
			e.mu.Unlock()
		}
	}

	e.mu.Lock()
	e.status.Results = append(e.status.Results, result)
	if !result.Success {
		e.status.Errors = append(e.status.Errors,
			fmt.Sprintf("%s (as %s): %s", t.ID, userLabel, result.ErrorOutput))
	}
	e.mu.Unlock()
}

func (e *Engine) abort() {
	simlog.Phase("ABORTED — running cleanup")
	e.runPendingCleanups()
	count := len(e.GetStatus().Results)
	simlog.Stop(fmt.Sprintf("Simulation ABORTED. %d techniques had run before abort.", count))
	reporter.SaveResults(e.GetStatus().Results, e.GetStatus().LogFile)
	e.setPhase(PhaseAborted)
	e.mu.Lock()
	e.status.CurrentStep = ""
	e.status.CurrentUser = ""
	e.status.CleanupDone = true
	e.mu.Unlock()
}

func (e *Engine) finish() {
	if !e.cfg.RunCleanup {
		simlog.Phase("POST-SIMULATION CLEANUP")
		e.runPendingCleanups()
	}
	results := e.GetStatus().Results
	succeeded := 0
	for _, r := range results {
		if r.Success {
			succeeded++
		}
	}
	summary := fmt.Sprintf("Simulation COMPLETE. %d/%d techniques successful. Log: %s",
		succeeded, len(results), e.GetStatus().LogFile)
	simlog.Stop(summary)
	reporter.SaveResults(results, e.GetStatus().LogFile)
	e.setPhase(PhaseDone)
	e.mu.Lock()
	e.status.CurrentStep = ""
	e.status.CurrentUser = ""
	e.status.CleanupDone = true
	e.mu.Unlock()
}

func (e *Engine) runPendingCleanups() {
	e.mu.RLock()
	pending := make([]*playbooks.Technique, len(e.executedTechniques))
	copy(pending, e.executedTechniques)
	e.mu.RUnlock()
	if len(pending) == 0 {
		simlog.Info("No pending cleanups.")
		return
	}
	simlog.Info(fmt.Sprintf("Running cleanup for %d techniques...", len(pending)))
	for _, t := range pending {
		executor.RunCleanupOnly(t)
	}
	e.mu.Lock()
	e.executedTechniques = nil
	e.mu.Unlock()
}

func (e *Engine) getTechniquesForPhase() []*playbooks.Technique {
	if e.cfg.CampaignID != "" {
		return e.getTechniquesForCampaign(e.cfg.CampaignID)
	}
	return e.registry.GetTechniquesByPhase("attack")
}

func (e *Engine) getTechniquesForCampaign(campaignID string) []*playbooks.Technique {
	campaign, ok := e.registry.Campaigns[campaignID]
	if !ok {
		simlog.Info(fmt.Sprintf("WARNING: Campaign not found: %s", campaignID))
		return nil
	}
	simlog.Info(fmt.Sprintf("Running campaign: %s (%s) — %d steps", campaign.Name, campaign.Industry, len(campaign.Steps)))
	var techniques []*playbooks.Technique
	for _, step := range campaign.Steps {
		t, exists := e.registry.Techniques[step.TechniqueID]
		if !exists {
			simlog.Info(fmt.Sprintf("WARNING: Technique %s not found (campaign: %s)", step.TechniqueID, campaignID))
			continue
		}
		techniques = append(techniques, t)
	}
	return techniques
}

func (e *Engine) setPhase(p Phase) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.status.Phase = p
	e.status.PhaseStartTimes[p] = time.Now().Format(time.RFC3339)
}

func (e *Engine) waitOrStop(d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-e.stopCh:
		return false
	}
}

func (e *Engine) isStopped() bool {
	select {
	case <-e.stopCh:
		return true
	default:
		return false
	}
}
