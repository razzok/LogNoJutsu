package engine

import (
	"fmt"
	"math/rand"
	"net"
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

// DayStatus represents the execution state of a single PoC day.
type DayStatus string

const (
	DayPending  DayStatus = "pending"
	DayActive   DayStatus = "active"
	DayComplete DayStatus = "complete"
)

// DayDigest records per-day execution summary for a PoC run.
type DayDigest struct {
	Day            int       `json:"day"`
	Phase          string    `json:"phase"`           // "phase1", "gap", "phase2"
	Status         DayStatus `json:"status"`
	TechniqueCount int       `json:"technique_count"`
	PassCount      int       `json:"pass_count"`
	FailCount      int       `json:"fail_count"`
	StartTime      string    `json:"start_time,omitempty"`
	EndTime        string    `json:"end_time,omitempty"`
	LastHeartbeat  string    `json:"last_heartbeat,omitempty"`
}

// ScanInfo holds what the scan confirmation modal displays.
type ScanInfo struct {
	TargetSubnet  string   `json:"target_subnet"`
	RateLimitNote string   `json:"rate_limit_note"`
	IDSWarning    string   `json:"ids_warning"`
	Techniques    []string `json:"techniques"`
}

// Engine manages the simulation lifecycle.
type Engine struct {
	mu                 sync.RWMutex
	status             Status
	dayDigests         []DayDigest  // per-day tracking for PoC runs, guarded by mu
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
	isAdmin            bool      // set once at Start(); guards per-technique elevation skip

	// Scan confirmation state (D-07, D-08, D-09)
	scanConfirmCh   chan struct{} // closed when consultant confirms; nil = no scan pending
	scanCancelCh    chan struct{} // closed when consultant cancels scan
	scanConfirmMu   sync.Mutex   // guards scanConfirmCh and scanPendingInfo
	scanPendingInfo *ScanInfo    // non-nil when confirmation is pending
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

// SetAdmin overrides the admin check result for testing.
func (e *Engine) SetAdmin(admin bool) {
	e.isAdmin = admin
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
	e.dayDigests = nil
	e.rotationIndex = 0
	e.scanConfirmMu.Lock()
	e.scanConfirmCh = nil
	e.scanCancelCh = nil
	e.scanPendingInfo = nil
	e.scanConfirmMu.Unlock()
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

	e.isAdmin = checkIsElevated()
	simlog.Info(fmt.Sprintf("Elevation check: isAdmin=%v", e.isAdmin))

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

// ConfirmScan unblocks the engine after scan confirmation.
func (e *Engine) ConfirmScan() error {
	e.scanConfirmMu.Lock()
	defer e.scanConfirmMu.Unlock()
	if e.scanConfirmCh == nil {
		return fmt.Errorf("no scan confirmation pending")
	}
	close(e.scanConfirmCh)
	e.scanConfirmCh = nil
	return nil
}

// CancelScan aborts the simulation from the scan confirmation modal.
func (e *Engine) CancelScan() error {
	e.scanConfirmMu.Lock()
	defer e.scanConfirmMu.Unlock()
	if e.scanCancelCh == nil {
		return fmt.Errorf("no scan confirmation pending")
	}
	close(e.scanCancelCh)
	e.scanCancelCh = nil
	return nil
}

// GetScanPending returns the current scan info if confirmation is pending, nil otherwise.
func (e *Engine) GetScanPending() *ScanInfo {
	e.scanConfirmMu.Lock()
	defer e.scanConfirmMu.Unlock()
	return e.scanPendingInfo
}

// runScanConfirmation collects all requires_confirmation techniques and blocks until
// the consultant confirms or cancels via the API. Returns nil if confirmed or no scan
// techniques exist; returns an error (after calling abort()) if cancelled or stopped.
func (e *Engine) runScanConfirmation() error {
	allTechniques := append(
		e.filterByTactics(e.registry.GetTechniquesByPhase("discovery")),
		e.getTechniquesForPhase()...,
	)
	var confirmTechs []string
	for _, t := range allTechniques {
		if t.RequiresConfirmation {
			confirmTechs = append(confirmTechs, t.ID)
		}
	}
	if len(confirmTechs) == 0 {
		return nil
	}

	info := &ScanInfo{
		TargetSubnet:  detectLocalSubnet(),
		RateLimitNote: "Rate limit: 50 connections/second",
		IDSWarning:    "Active scanning may trigger IDS/IPS alerts on monitored networks.",
		Techniques:    confirmTechs,
	}
	e.scanConfirmMu.Lock()
	e.scanPendingInfo = info
	e.scanConfirmCh = make(chan struct{})
	e.scanCancelCh = make(chan struct{})
	confirmCh := e.scanConfirmCh
	cancelCh := e.scanCancelCh
	e.scanConfirmMu.Unlock()

	simlog.Info("[ScanConfirm] Waiting for consultant confirmation...")
	select {
	case <-confirmCh:
		simlog.Info("[ScanConfirm] Confirmed — proceeding with simulation")
		e.scanConfirmMu.Lock()
		e.scanPendingInfo = nil
		e.scanConfirmMu.Unlock()
		return nil
	case <-cancelCh:
		simlog.Info("[ScanConfirm] Cancelled by consultant")
		e.scanConfirmMu.Lock()
		e.scanPendingInfo = nil
		e.scanConfirmMu.Unlock()
		e.abort()
		return fmt.Errorf("scan cancelled")
	case <-e.stopCh:
		e.scanConfirmMu.Lock()
		e.scanPendingInfo = nil
		e.scanConfirmMu.Unlock()
		e.abort()
		return fmt.Errorf("stopped")
	}
}

// detectLocalSubnet returns the /24 subnet of the first non-loopback IPv4 interface.
func detectLocalSubnet() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipNet.IP.To4(); ip4 != nil {
					return fmt.Sprintf("%d.%d.%d.0/24", ip4[0], ip4[1], ip4[2])
				}
			}
		}
	}
	return "unknown"
}

func (e *Engine) GetStatus() Status {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// GetDayDigests returns a copy of the per-day digest slice.
// Returns an empty slice (never nil) when no PoC is running, so JSON encodes as [] not null.
func (e *Engine) GetDayDigests() []DayDigest {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if len(e.dayDigests) == 0 {
		return []DayDigest{}
	}
	result := make([]DayDigest, len(e.dayDigests))
	copy(result, e.dayDigests)
	return result
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

	// Scan confirmation pre-flight — fires once per simulation (per D-07)
	if !e.cfg.WhatIf {
		if err := e.runScanConfirmation(); err != nil {
			// Cancelled or stopped during confirmation
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
	globalDay := 0

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

	// Pre-populate all days as pending so /api/poc/days returns the full schedule from first poll (D-02).
	{
		digests := make([]DayDigest, totalDays)
		globalIdx := 0
		for i := 0; i < cfg.Phase1DurationDays; i++ {
			digests[globalIdx] = DayDigest{Day: globalIdx + 1, Phase: "phase1", Status: DayPending, TechniqueCount: techsPerDay}
			globalIdx++
		}
		for i := 0; i < cfg.GapDays; i++ {
			digests[globalIdx] = DayDigest{Day: globalIdx + 1, Phase: "gap", Status: DayPending}
			globalIdx++
		}
		// Phase 2 technique count: campaign step count if campaign, else attack registry length.
		phase2TechCount := 0
		if cfg.CampaignID != "" {
			if c, ok := e.registry.Campaigns[cfg.CampaignID]; ok {
				phase2TechCount = len(c.Steps)
			}
		} else {
			phase2TechCount = len(e.registry.GetTechniquesByPhase("attack"))
		}
		for i := 0; i < cfg.Phase2DurationDays; i++ {
			digests[globalIdx] = DayDigest{Day: globalIdx + 1, Phase: "phase2", Status: DayPending, TechniqueCount: phase2TechCount}
			globalIdx++
		}
		e.mu.Lock()
		e.dayDigests = digests
		e.mu.Unlock()
	}

	// Scan confirmation pre-flight — fires once per PoC simulation (per D-07)
	if !cfg.WhatIf {
		if err := e.runScanConfirmation(); err != nil {
			// Cancelled or stopped during confirmation
			return
		}
	}

	// ── Phase 1: Discovery ──────────────────────────────────────────
	e.setPhase(PhasePoCPhase1)
	simlog.Phase("PoC Phase 1: Discovery")
	for day := 1; day <= cfg.Phase1DurationDays; day++ {
		globalDay++
		if e.isStopped() {
			e.abort()
			return
		}
		d := nextOccurrenceOfHour(cfg.Phase1DailyHour, e.clock.Now())
		nextRun := e.clock.Now().Add(d)
		simlog.Info(fmt.Sprintf("[PoC Phase1] Day %d/%d — waiting until %s", day, cfg.Phase1DurationDays, nextRun.Format("2006-01-02 15:04")))
		e.mu.Lock()
		e.status.PoCDay = globalDay
		e.status.PoCPhase = "phase1"
		e.status.NextScheduledRun = nextRun.Format(time.RFC3339)
		e.status.CurrentStep = fmt.Sprintf("PoC Phase 1 — Day %d of %d — Waiting until %02d:00", globalDay, totalDays, cfg.Phase1DailyHour)
		e.dayDigests[globalDay-1].Status = DayActive
		e.dayDigests[globalDay-1].StartTime = e.clock.Now().Format(time.RFC3339)
		e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
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
				// Update pass/fail count and heartbeat (D-10)
				e.mu.RLock()
				lastResult := e.status.Results[len(e.status.Results)-1]
				e.mu.RUnlock()
				e.mu.Lock()
				if lastResult.Success {
					e.dayDigests[globalDay-1].PassCount++
				} else {
					e.dayDigests[globalDay-1].FailCount++
				}
				e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
				e.mu.Unlock()
				e.delayBetween()
			}
		}
		e.mu.Lock()
		e.dayDigests[globalDay-1].Status = DayComplete
		e.dayDigests[globalDay-1].EndTime = e.clock.Now().Format(time.RFC3339)
		e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
		e.mu.Unlock()
		simlog.Info(fmt.Sprintf("[PoC Phase1] Day %d complete", day))
	}

	// ── Gap ─────────────────────────────────────────────────────────
	if cfg.GapDays > 0 {
		e.setPhase(PhasePoCGap)
		simlog.Phase("PoC Gap")
		for day := 1; day <= cfg.GapDays; day++ {
			globalDay++
			if e.isStopped() {
				e.abort()
				return
			}
			d := nextOccurrenceOfHour(cfg.Phase2DailyHour, e.clock.Now())
			nextRun := e.clock.Now().Add(d)
			simlog.Info(fmt.Sprintf("[PoC Gap] Day %d/%d — silent day, waiting until %s", day, cfg.GapDays, nextRun.Format("2006-01-02 15:04")))
			e.mu.Lock()
			e.status.PoCDay = globalDay
			e.status.PoCPhase = "gap"
			e.status.NextScheduledRun = nextRun.Format(time.RFC3339)
			e.status.CurrentStep = fmt.Sprintf("PoC Gap — Day %d of %d (no actions)", globalDay, totalDays)
			e.dayDigests[globalDay-1].Status = DayActive
			e.dayDigests[globalDay-1].StartTime = e.clock.Now().Format(time.RFC3339)
			e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
			e.mu.Unlock()
			if !e.waitOrStop(d) {
				e.abort()
				return
			}
			// Gap days have no techniques; mark complete after the wait (Pitfall 5)
			e.mu.Lock()
			e.dayDigests[globalDay-1].Status = DayComplete
			e.dayDigests[globalDay-1].EndTime = e.clock.Now().Format(time.RFC3339)
			e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
			e.mu.Unlock()
		}
	}

	// ── Phase 2: Attack ──────────────────────────────────────────────
	e.setPhase(PhasePoCPhase2)
	simlog.Phase("PoC Phase 2: Attack")
	for day := 1; day <= cfg.Phase2DurationDays; day++ {
		globalDay++
		if e.isStopped() {
			e.abort()
			return
		}
		d := nextOccurrenceOfHour(cfg.Phase2DailyHour, e.clock.Now())
		nextRun := e.clock.Now().Add(d)
		simlog.Info(fmt.Sprintf("[PoC Phase2] Day %d/%d — waiting until %s", day, cfg.Phase2DurationDays, nextRun.Format("2006-01-02 15:04")))
		e.mu.Lock()
		e.status.PoCDay = globalDay
		e.status.PoCPhase = "phase2"
		e.status.NextScheduledRun = nextRun.Format(time.RFC3339)
		e.status.CurrentStep = fmt.Sprintf("PoC Phase 2 — Day %d of %d — Waiting until %02d:00", globalDay, totalDays, cfg.Phase2DailyHour)
		e.dayDigests[globalDay-1].Status = DayActive
		e.dayDigests[globalDay-1].StartTime = e.clock.Now().Format(time.RFC3339)
		e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
		e.mu.Unlock()

		if !e.waitOrStop(d) {
			e.abort()
			return
		}

		if cfg.CampaignID != "" {
			campaign, ok := e.registry.Campaigns[cfg.CampaignID]
			if ok {
				simlog.Info(fmt.Sprintf("[PoC Phase2] Day %d — running %d campaign steps%s", day, len(campaign.Steps), e.whatIfLabel()))
				for _, step := range campaign.Steps {
					if e.isStopped() {
						e.abort()
						return
					}
					t, exists := e.registry.Techniques[step.TechniqueID]
					if !exists {
						continue
					}
					e.runTechnique(t)
					// Update pass/fail count and heartbeat (D-10)
					e.mu.RLock()
					lastResult := e.status.Results[len(e.status.Results)-1]
					e.mu.RUnlock()
					e.mu.Lock()
					if lastResult.Success {
						e.dayDigests[globalDay-1].PassCount++
					} else {
						e.dayDigests[globalDay-1].FailCount++
					}
					e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
					e.mu.Unlock()
					// Apply campaign delay_after (D-07, D-08, D-09)
					if step.DelayAfter > 0 {
						if !e.waitOrStop(time.Duration(step.DelayAfter) * time.Second) {
							e.abort()
							return
						}
					}
					e.delayBetween()
				}
			}
		} else {
			attackTechs := e.getTechniquesForPhase()
			simlog.Info(fmt.Sprintf("[PoC Phase2] Day %d — running %d attack techniques%s", day, len(attackTechs), e.whatIfLabel()))
			for _, t := range attackTechs {
				if e.isStopped() {
					e.abort()
					return
				}
				e.runTechnique(t)
				// Update pass/fail count and heartbeat (D-10)
				e.mu.RLock()
				lastResult := e.status.Results[len(e.status.Results)-1]
				e.mu.RUnlock()
				e.mu.Lock()
				if lastResult.Success {
					e.dayDigests[globalDay-1].PassCount++
				} else {
					e.dayDigests[globalDay-1].FailCount++
				}
				e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
				e.mu.Unlock()
				e.delayBetween()
			}
		}
		e.mu.Lock()
		e.dayDigests[globalDay-1].Status = DayComplete
		e.dayDigests[globalDay-1].EndTime = e.clock.Now().Format(time.RFC3339)
		e.dayDigests[globalDay-1].LastHeartbeat = e.clock.Now().Format(time.RFC3339)
		e.mu.Unlock()
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

	// Elevation gating (per D-05) — before any execution attempt
	if t.ElevationRequired && !e.isAdmin {
		simlog.Info(fmt.Sprintf("[ElevationSkip] %s — requires elevation, skipping", t.ID))
		now := e.clock.Now().Format(time.RFC3339)
		result = playbooks.ExecutionResult{
			TechniqueID:        t.ID,
			TechniqueName:      t.Name,
			TacticID:           t.Tactic,
			StartTime:          now,
			EndTime:            now,
			Success:            false,
			Output:             "Elevation required — technique skipped (not running as Administrator)",
			RunAsUser:          userLabel,
			VerificationStatus: playbooks.VerifElevationRequired,
			SIEMCoverage:       t.SIEMCoverage,
		}
		e.mu.Lock()
		e.status.Results = append(e.status.Results, result)
		e.mu.Unlock()
		return
	}

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
