package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"lognojutsu/internal/engine"
	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/preparation"
	"lognojutsu/internal/simlog"
	"lognojutsu/internal/userstore"
)

//go:embed static
var staticFiles embed.FS

// Config holds server configuration.
type Config struct {
	Host     string
	Port     int
	Password string
	Version  string
}

// Server holds all dependencies for the HTTP layer.
type Server struct {
	eng      *engine.Engine
	registry *playbooks.Registry
	users    *userstore.Store
	cfg      Config
}

// Start initializes the server and begins listening.
func Start(c Config) error {
	reg, err := playbooks.LoadEmbedded()
	if err != nil {
		return fmt.Errorf("loading playbooks: %w", err)
	}
	log.Printf("Loaded %d techniques, %d campaigns", len(reg.Techniques), len(reg.Campaigns))

	us, err := userstore.Load()
	if err != nil {
		log.Printf("WARNING: Could not load user store: %v (starting with empty store)", err)
		us, _ = userstore.Load()
	}
	log.Printf("Loaded %d user profiles", len(us.List()))

	s := &Server{
		eng:      engine.New(reg, us),
		registry: reg,
		users:    us,
		cfg:      c,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	log.Printf("LogNoJutsu UI: http://%s", addr)
	if c.Host == "0.0.0.0" {
		log.Printf("WARNING: UI accessible from network — ensure firewall rules are in place")
	}
	return http.ListenAndServe(addr, mux)
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Static UI
	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// Version info — public, no auth required (per D-10)
	mux.HandleFunc("/api/info", s.handleInfo)

	// Simulation API
	mux.HandleFunc("/api/status", s.authMiddleware(s.handleStatus))
	mux.HandleFunc("/api/techniques", s.authMiddleware(s.handleTechniques))
	mux.HandleFunc("/api/campaigns", s.authMiddleware(s.handleCampaigns))
	mux.HandleFunc("/api/tactics", s.authMiddleware(s.handleTactics))
	mux.HandleFunc("/api/start", s.authMiddleware(s.handleStart))
	mux.HandleFunc("/api/stop", s.authMiddleware(s.handleStop))
	mux.HandleFunc("/api/logs", s.authMiddleware(s.handleLogs))
	mux.HandleFunc("/api/report", s.authMiddleware(s.handleReport))

	// Preparation API
	mux.HandleFunc("/api/prepare", s.authMiddleware(s.handlePrepare))
	mux.HandleFunc("/api/prepare/step", s.authMiddleware(s.handlePrepareStep))

	// User management API
	mux.HandleFunc("/api/users", s.authMiddleware(s.handleUsers))
	mux.HandleFunc("/api/users/discover", s.authMiddleware(s.handleUsersDiscover))
	mux.HandleFunc("/api/users/test", s.authMiddleware(s.handleUsersTest))
	mux.HandleFunc("/api/users/delete", s.authMiddleware(s.handleUsersDelete))
}

// authMiddleware optionally enforces basic password protection.
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Password != "" {
			_, pass, ok := r.BasicAuth()
			if !ok || pass != s.cfg.Password {
				w.Header().Set("WWW-Authenticate", `Basic realm="LogNoJutsu"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	writeJSON(w, map[string]string{"error": msg})
}

// ── Simulation ────────────────────────────────────────────────────────────────

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.eng.GetStatus())
}

func (s *Server) handleTechniques(w http.ResponseWriter, r *http.Request) {
	var list []*playbooks.Technique
	for _, t := range s.registry.Techniques {
		list = append(list, t)
	}
	writeJSON(w, list)
}

func (s *Server) handleCampaigns(w http.ResponseWriter, r *http.Request) {
	var list []*playbooks.Campaign
	for _, c := range s.registry.Campaigns {
		list = append(list, c)
	}
	writeJSON(w, list)
}

func (s *Server) handleTactics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.registry.GetAllTactics())
}

// handleReport serves the latest HTML report file if one exists.
func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	reportFile := s.eng.GetStatus().ReportFile
	if reportFile == "" {
		writeError(w, "No report available yet — run a simulation first", http.StatusNotFound)
		return
	}
	data, err := os.ReadFile(reportFile)
	if err != nil {
		writeError(w, "Could not read report file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", `inline; filename="lognojutsu_report.html"`)
	w.Write(data)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var cfg engine.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.eng.Start(cfg); err != nil {
		writeError(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, map[string]string{"status": "started"})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	s.eng.Stop()
	writeJSON(w, map[string]string{"status": "stopped"})
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	entries := simlog.GetEntries()
	if entries == nil {
		entries = []simlog.Entry{}
	}
	writeJSON(w, entries)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]string{"version": s.cfg.Version})
}

// ── Preparation ───────────────────────────────────────────────────────────────

func (s *Server) handlePrepare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	results := preparation.RunAll()
	writeJSON(w, results)
}

func (s *Server) handlePrepareStep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Step string `json:"step"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid body", http.StatusBadRequest)
		return
	}
	var result preparation.Result
	switch req.Step {
	case "powershell":
		result = preparation.EnablePowerShellLogging()
	case "auditpol":
		result = preparation.ConfigureAuditPolicy()
	case "sysmon":
		result = preparation.InstallSysmon()
	default:
		writeError(w, "Unknown step: "+req.Step, http.StatusBadRequest)
		return
	}
	writeJSON(w, result)
}

// ── User Management ───────────────────────────────────────────────────────────

// GET  /api/users          → list all profiles (no passwords)
// POST /api/users          → add/update a profile
func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, s.users.List())

	case http.MethodPost:
		var req struct {
			Username    string            `json:"username"`
			Domain      string            `json:"domain"`
			Password    string            `json:"password"`
			UserType    userstore.UserType `json:"user_type"`
			DisplayName string            `json:"display_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Username) == "" {
			writeError(w, "username is required", http.StatusBadRequest)
			return
		}
		if req.UserType == "" {
			req.UserType = userstore.UserTypeLocal
		}
		profile, err := s.users.Add(req.Username, req.Domain, req.Password, req.UserType, req.DisplayName)
		if err != nil {
			writeError(w, "Failed to save profile: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, profile)

	default:
		writeError(w, "GET or POST required", http.StatusMethodNotAllowed)
	}
}

// POST /api/users/delete  → delete a profile by ID
func (s *Server) handleUsersDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid body", http.StatusBadRequest)
		return
	}
	if err := s.users.Delete(req.ID); err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]string{"status": "deleted"})
}

// POST /api/users/discover → enumerate local and recently-logged-on domain users
func (s *Server) handleUsersDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	type discoveryResponse struct {
		LocalUsers  []userstore.DiscoveredUser `json:"local_users"`
		DomainUsers []userstore.DiscoveredUser `json:"domain_users"`
		Error       string                     `json:"error,omitempty"`
	}

	resp := discoveryResponse{}

	localUsers, err := userstore.DiscoverLocalUsers()
	if err != nil {
		resp.Error = "Local user discovery error: " + err.Error()
	} else {
		resp.LocalUsers = localUsers
	}

	domainUsers, err := userstore.DiscoverRecentDomainUsers()
	if err != nil {
		if resp.Error != "" {
			resp.Error += "; "
		}
		resp.Error += "Domain user discovery error: " + err.Error()
	} else {
		resp.DomainUsers = domainUsers
	}

	if resp.LocalUsers == nil {
		resp.LocalUsers = []userstore.DiscoveredUser{}
	}
	if resp.DomainUsers == nil {
		resp.DomainUsers = []userstore.DiscoveredUser{}
	}

	writeJSON(w, resp)
}

// POST /api/users/test → validate credentials for a stored profile
func (s *Server) handleUsersTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid body", http.StatusBadRequest)
		return
	}
	profile, ok := s.users.Get(req.ID)
	if !ok {
		writeError(w, "Profile not found", http.StatusNotFound)
		return
	}
	password, err := s.users.DecryptPassword(req.ID)
	if err != nil {
		writeError(w, "Could not decrypt password: "+err.Error(), http.StatusInternalServerError)
		return
	}
	ok, msg := userstore.TestCredentials(profile, password)
	s.users.SetTestResult(req.ID, ok, msg)
	writeJSON(w, map[string]any{
		"id":      req.ID,
		"success": ok,
		"message": msg,
	})
}
