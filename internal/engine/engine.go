package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"lognojutsu/internal/executor"
	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/reporter"
	"lognojutsu/internal/simlog"
	"lognojutsu/internal/userstore"
	"lognojutsu/internal/verifier"
)

type Phase string

const (
	PhaseIdle      Phase = "idle"
	PhaseDiscovery Phase = "discovery"
	PhaseAttack    Phase = "attack"
	PhaseDone      Phase = "done"
	PhaseAborted   Phase = "aborted"
	PhasePoCPhase1 Phase = "poc_phase1"
	PhasePoCGap    Phase = "poc_gap"
	PhasePoCPhase2 Phase = "poc_phase2"
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

	// Execution options
	WhatIf                bool     `json:"whatif"`                  // preview without executing
	DelayBetweenTechniques int     `json:"delay_between_techniques"` // seconds between each technique
	ExcludedTactics       []string `json:"excluded_tactics"`         // skip techniques with these tactics
	IncludedTactics       []string `json:"included_tactics"`         // only run these tactics (empty = all)

	VerificationWaitSecs int `json:"verification_wait_secs"` // 0 = default 3s

	// PoC multi-day scheduling mode
	PoCMode            bool `json:"poc_mode"`
	Phase1DurationDays int  `json:"phase1_duration_days"` // days to run discovery phase
	Phase1TechsPerDay  int  `json:"phase1_techs_per_day"` // discovery techniques per day
	Phase1DailyHour    int  `json:"phase1_daily_hour"`    // 0-23: hour at which to run each day
	GapDays            int  `json:"gap_days"`             // silent days between phases
	Phase2DurationDays int  `json:"phase2_duration_days"` // days to run attack phase
	Phase2DailyHour    int  `json:"phase2_daily_hour"`    // 0-23: hour at which to run each day
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

	// WhatIf mode indicator
	WhatIf     bool   `json:"whatif"`
	ReportFile string `json:"report_file,omitempty"` // path to generated HTML report

	// PoC mode fields (populated only when PoCMode is active)
	PoCDay           int    `json:"poc_day,omitempty"`
	PoCTotalDays     int    `json:"poc_total_days,omitempty"`
	PoCPhase         string `json:"poc_phase,omitempty"`          // "phase1", "gap", "phase2"
	NextScheduledRun string `json:"next_scheduled_run,omitempty"` // RFC3339
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
	runner             RunnerFunc // nil = real executor
	clock              Clock     // injectable; defaults to realClock{}
}

type resolvedProfile struct {
	profile  *userstore.UserProfile
	password string
}

// RunnerFunc abstracts technique execution for testability.
// Mirrors the QueryFn pattern in the verifier package (D-06).
// nil means use the real executor (production path unchanged).
type RunnerFunc func(t *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult

// Clock abstracts time operations for deterministic testing.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

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
		clock:  realClock{},
	}
}

// SetRunner injects a custom execution function for testing.
// Pass nil to restore the default (real executor).
func (e *Engine) SetRunner(fn RunnerFunc) {
	e.runner = fn
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
	initialPhase := PhaseDiscovery
	if cfg.PoCMode {
		initialPhase = PhasePoCPhase1
	}
	e.status = Status{
		Phase:           initialPhase,
		StartTime:       time.Now().Format(time.RFC3339),
		PhaseStartTimes: make(map[Phase]string),
		Results:         []playbooks.ExecutionResult{},
		Errors:          []string{},
		WhatIf:          cfg.WhatIf,
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
	if e.cfg.PoCMode {
		e.runPoC()
		return
	}
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
	discoveryTechniques := e.filterByTactics(e.registry.GetTechniquesByPhase("discovery"))
	simlog.Info(fmt.Sprintf("Discovery phase: %d techniques to run%s", len(discoveryTechniques), e.whatIfLabel()))
	for _, t := range discoveryTechniques {
		if e.isStopped() {
			e.abort()
			return
		}
		e.runTechnique(t)
		e.delayBetween()
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
	simlog.Info(fmt.Sprintf("Attack phase: %d techniques to run%s", len(attackTechniques), e.whatIfLabel()))
	for _, t := range attackTechniques {
		if e.isStopped() {
			e.abort()
			return
		}
		e.runTechnique(t)
		e.delayBetween()
	}

	e.finish()
}

// nextOccurrenceOfHour returns the duration until the next occurrence of hour:00:00
// on the local clock. If that time has already passed today, returns duration to tomorrow.
func nextOccurrenceOfHour(hour int, now time.Time) time.Duration {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}

func (e *Engine) runPoC() {
	cfg := e.cfg
	totalDays := cfg.Phase1DurationDays + cfg.GapDays + cfg.Phase2DurationDays

	simlog.Info(fmt.Sprintf("[PoC] Multi-day simulation: Phase1=%dd (%d techs/day @ %02d:00)  Gap=%dd  Phase2=%dd (campaign=%q @ %02d:00)  Total=%dd",
		cfg.Phase1DurationDays, cfg.Phase1TechsPerDay, cfg.Phase1DailyHour,
		cfg.GapDays,
		cfg.Phase2DurationDays, cfg.CampaignID, cfg.Phase2DailyHour,
		totalDays))

	e.mu.Lock()
	e.status.PoCTotalDays = totalDays
	e.mu.Unlock()

	// Shuffle discovery techniques once — cycle through them across days
	discoveryTechs := e.filterByTactics(e.registry.GetTechniquesByPhase("discovery"))
	rand.Shuffle(len(discoveryTechs), func(i, j int) {
		discoveryTechs[i], discoveryTechs[j] = discoveryTechs[j], discoveryTechs[i]
	})
	techsPerDay := cfg.Phase1TechsPerDay
	if techsPerDay < 1 {
		techsPerDay = 2
	}

	// ── Phase 1: Discovery ──────────────────────────────────────────
	e.setPhase(PhasePoCPhase1)
	for day := 1; day <= cfg.Phase1DurationDays; day++ {
		if e.isStopped() {
			e.abort()
			return
		}
		d := nextOccurrenceOfHour(cfg.Phase1DailyHour, e.clock.Now())
		nextRun := e.clock.Now().Add(d)
		simlog.Info(fmt.Sprintf("[PoC Phase1] Day %d/%d — waiting until %s", day, cfg.Phase1DurationDays, nextRun.Format("2006-01-02 15:04")))
		e.mu.Lock()
		e.status.PoCDay = day
		e.status.PoCPhase = "phase1"
		e.status.NextScheduledRun = nextRun.Format(time.RFC3339)
		e.status.CurrentStep = fmt.Sprintf("PoC Phase 1 — Tag %d/%d — warte bis %02d:00 Uhr", day, cfg.Phase1DurationDays, cfg.Phase1DailyHour)
		e.mu.Unlock()

		if !e.waitOrStop(d) {
			e.abort()
			return
		}

		if len(discoveryTechs) > 0 {
			start := ((day - 1) * techsPerDay) % len(discoveryTechs)
			for i := 0; i < techsPerDay; i++ {
				if e.isStopped() {
					e.abort()
					return
				}
				t := discoveryTechs[(start+i)%len(discoveryTechs)]
				simlog.Info(fmt.Sprintf("[PoC Phase1] Day %d — technique %d/%d: %s%s", day, i+1, techsPerDay, t.ID, e.whatIfLabel()))
				e.runTechnique(t)
				e.delayBetween()
			}
		}
		simlog.Info(fmt.Sprintf("[PoC Phase1] Day %d complete", day))
	}

	// ── Gap ─────────────────────────────────────────────────────────
	if cfg.GapDays > 0 {
		e.setPhase(PhasePoCGap)
		for day := 1; day <= cfg.GapDays; day++ {
			if e.isStopped() {
				e.abort()
				return
			}
			d := nextOccurrenceOfHour(cfg.Phase2DailyHour, e.clock.Now())
			nextRun := e.clock.Now().Add(d)
			simlog.Info(fmt.Sprintf("[PoC Gap] Day %d/%d — silent day, waiting until %s", day, cfg.GapDays, nextRun.Format("2006-01-02 15:04")))
			e.mu.Lock()
			e.status.PoCPhase = "gap"
			e.status.NextScheduledRun = nextRun.Format(time.RFC3339)
			e.status.CurrentStep = fmt.Sprintf("PoC Pause — Tag %d/%d (keine Aktionen)", day, cfg.GapDays)
			e.mu.Unlock()
			if !e.waitOrStop(d) {
				e.abort()
				return
			}
		}
	}

	// ── Phase 2: Attack ──────────────────────────────────────────────
	e.setPhase(PhasePoCPhase2)
	for day := 1; day <= cfg.Phase2DurationDays; day++ {
		if e.isStopped() {
			e.abort()
			return
		}
		d := nextOccurrenceOfHour(cfg.Phase2DailyHour, e.clock.Now())
		nextRun := e.clock.Now().Add(d)
		simlog.Info(fmt.Sprintf("[PoC Phase2] Day %d/%d — waiting until %s", day, cfg.Phase2DurationDays, nextRun.Format("2006-01-02 15:04")))
		e.mu.Lock()
		e.status.PoCPhase = "phase2"
		e.status.NextScheduledRun = nextRun.Format(time.RFC3339)
		e.status.CurrentStep = fmt.Sprintf("PoC Phase 2 — Tag %d/%d — warte bis %02d:00 Uhr", day, cfg.Phase2DurationDays, cfg.Phase2DailyHour)
		e.mu.Unlock()

		if !e.waitOrStop(d) {
			e.abort()
			return
		}

		attackTechs := e.getTechniquesForPhase()
		simlog.Info(fmt.Sprintf("[PoC Phase2] Day %d — running %d attack techniques%s", day, len(attackTechs), e.whatIfLabel()))
		for _, t := range attackTechs {
			if e.isStopped() {
				e.abort()
				return
			}
			e.runTechnique(t)
			e.delayBetween()
		}
		simlog.Info(fmt.Sprintf("[PoC Phase2] Day %d complete", day))
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
	if e.cfg.WhatIf {
		// WhatIf: record what would run without executing anything
		now := time.Now().Format(time.RFC3339)
		result = playbooks.ExecutionResult{
			TechniqueID:        t.ID,
			TechniqueName:      t.Name,
			TacticID:           t.Tactic,
			StartTime:          now,
			EndTime:            now,
			Success:            true,
			Output:             "[WhatIf] Nicht ausgeführt — Vorschau-Modus aktiv",
			RunAsUser:          userLabel,
			VerificationStatus: playbooks.VerifNotRun,
		}
		simlog.Info(fmt.Sprintf("[WhatIf] Would run: %s — %s (as %s)", t.ID, t.Name, userLabel))
	} else if e.runner != nil {
		// Injected runner for testing (D-05)
		result = e.runner(t, profile, password)
		if t.Cleanup != "" && !e.cfg.RunCleanup {
			e.mu.Lock()
			e.executedTechniques = append(e.executedTechniques, t)
			e.mu.Unlock()
		}
	} else if e.cfg.RunCleanup {
		result = executor.RunWithCleanup(t, profile, password)
	} else {
		result = executor.RunAs(t, profile, password)
		if t.Cleanup != "" {
			e.mu.Lock()
			e.executedTechniques = append(e.executedTechniques, t)
			e.mu.Unlock()
		}
	}
	result.SIEMCoverage = t.SIEMCoverage

	// ── Post-execution verification ──────────────────────────────────────
	if !e.cfg.WhatIf && len(t.ExpectedEvents) > 0 {
		waitSecs := e.cfg.VerificationWaitSecs
		if waitSecs <= 0 {
			waitSecs = 3
		}
		simlog.Info(fmt.Sprintf("[Verify] %s — waiting %ds for event log writes...", t.ID, waitSecs))
		time.Sleep(time.Duration(waitSecs) * time.Second)

		startTime, _ := time.Parse(time.RFC3339, result.StartTime)
		status, verified := verifier.Verify(t.ExpectedEvents, startTime, result.Success, verifier.DefaultQueryFn)
		result.VerificationStatus = status
		result.VerifiedEvents = verified
		result.VerifyTime = time.Now().Format(time.RFC3339)

		foundCount := 0
		for _, v := range verified {
			if v.Found {
				foundCount++
			}
		}
		simlog.Info(fmt.Sprintf("[Verify] %s — %s (%d/%d events found)", t.ID, status, foundCount, len(verified)))
	} else if !e.cfg.WhatIf {
		result.VerificationStatus = playbooks.VerifNotRun
	}

	e.mu.Lock()
	e.status.Results = append(e.status.Results, result)
	if !result.Success {
		e.status.Errors = append(e.status.Errors,
			fmt.Sprintf("%s (as %s): %s", t.ID, userLabel, result.ErrorOutput))
	}
	e.mu.Unlock()
}

// filterByTactics filters a technique list by the IncludedTactics / ExcludedTactics config.
func (e *Engine) filterByTactics(techniques []*playbooks.Technique) []*playbooks.Technique {
	if len(e.cfg.IncludedTactics) == 0 && len(e.cfg.ExcludedTactics) == 0 {
		return techniques
	}
	included := make(map[string]bool, len(e.cfg.IncludedTactics))
	for _, t := range e.cfg.IncludedTactics {
		included[strings.ToLower(t)] = true
	}
	excluded := make(map[string]bool, len(e.cfg.ExcludedTactics))
	for _, t := range e.cfg.ExcludedTactics {
		excluded[strings.ToLower(t)] = true
	}
	var result []*playbooks.Technique
	for _, t := range techniques {
		tactic := strings.ToLower(t.Tactic)
		if excluded[tactic] {
			continue
		}
		if len(included) > 0 && !included[tactic] {
			continue
		}
		result = append(result, t)
	}
	return result
}

// delayBetween waits DelayBetweenTechniques seconds if configured and not stopped.
func (e *Engine) delayBetween() {
	if e.cfg.DelayBetweenTechniques > 0 && !e.isStopped() {
		e.waitOrStop(time.Duration(e.cfg.DelayBetweenTechniques) * time.Second)
	}
}

func (e *Engine) whatIfLabel() string {
	if e.cfg.WhatIf {
		return " [WhatIf — Vorschau-Modus, keine echte Ausführung]"
	}
	return ""
}

func (e *Engine) abort() {
	simlog.Phase("ABORTED — running cleanup")
	e.runPendingCleanups()
	count := len(e.GetStatus().Results)
	simlog.Stop(fmt.Sprintf("Simulation ABORTED. %d techniques had run before abort.", count))
	reporter.SaveResults(e.GetStatus().Results, e.GetStatus().LogFile, e.cfg.WhatIf)
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
	htmlReport := reporter.SaveResults(results, e.GetStatus().LogFile, e.cfg.WhatIf)
	if htmlReport != "" {
		e.mu.Lock()
		e.status.ReportFile = htmlReport
		e.mu.Unlock()
	}
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
	var techs []*playbooks.Technique
	if e.cfg.CampaignID != "" {
		techs = e.getTechniquesForCampaign(e.cfg.CampaignID)
	} else {
		techs = e.registry.GetTechniquesByPhase("attack")
	}
	return e.filterByTactics(techs)
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
	e.status.PhaseStartTimes[p] = e.clock.Now().Format(time.RFC3339)
}

func (e *Engine) waitOrStop(d time.Duration) bool {
	select {
	case <-e.clock.After(d):
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
