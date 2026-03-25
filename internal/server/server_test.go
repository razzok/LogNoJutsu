package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"lognojutsu/internal/engine"
	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/userstore"
)

// testServer creates a minimal Server with an in-memory registry and empty user store.
// No global state, no running HTTP server — all handlers are called directly.
func testServer(t *testing.T, password string) *Server {
	t.Helper()
	reg := &playbooks.Registry{
		Techniques: map[string]*playbooks.Technique{
			"T0001": {ID: "T0001", Name: "Test Alpha", Tactic: "discovery", Phase: "discovery"},
			"T0002": {ID: "T0002", Name: "Test Beta", Tactic: "execution", Phase: "attack"},
		},
		Campaigns: map[string]*playbooks.Campaign{},
	}
	us, _ := userstore.Load()
	eng := engine.New(reg, us)
	return &Server{
		eng:      eng,
		registry: reg,
		users:    us,
		cfg:      Config{Password: password},
	}
}

// TestHandleStatus_idle verifies GET /api/status returns 200 with phase=idle when engine is idle.
func TestHandleStatus_idle(t *testing.T) {
	s := testServer(t, "")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	s.handleStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"phase":"idle"`) {
		t.Errorf("expected phase idle in body: %s", body)
	}
}

// TestHandleStatus_running verifies GET /api/status returns non-idle phase when engine is running.
func TestHandleStatus_running(t *testing.T) {
	s := testServer(t, "")
	// Inject a slow runner so the engine stays in a running phase
	s.eng.SetRunner(func(tech *playbooks.Technique, p *userstore.UserProfile, pw string) playbooks.ExecutionResult {
		time.Sleep(2 * time.Second)
		return playbooks.ExecutionResult{
			TechniqueID:   tech.ID,
			TechniqueName: tech.Name,
			TacticID:      tech.Tactic,
			StartTime:     time.Now().Format(time.RFC3339),
			EndTime:       time.Now().Format(time.RFC3339),
			Success:       true,
			Output:        "test",
		}
	})
	// Start engine — it will call our slow runner in a goroutine
	err := s.eng.Start(engine.Config{})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer s.eng.Stop()
	// Give goroutine time to enter running phase
	time.Sleep(50 * time.Millisecond)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	s.handleStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	// Phase should NOT be idle (it's running)
	if strings.Contains(body, `"phase":"idle"`) {
		t.Errorf("expected non-idle phase, got: %s", body)
	}
}

// TestHandleStart_validConfig verifies POST /api/start with valid JSON returns 200 + started status.
func TestHandleStart_validConfig(t *testing.T) {
	s := testServer(t, "")
	// Inject fast runner so simulation completes quickly
	s.eng.SetRunner(func(tech *playbooks.Technique, p *userstore.UserProfile, pw string) playbooks.ExecutionResult {
		return playbooks.ExecutionResult{
			TechniqueID:   tech.ID,
			TechniqueName: tech.Name,
			TacticID:      tech.Tactic,
			StartTime:     time.Now().Format(time.RFC3339),
			EndTime:       time.Now().Format(time.RFC3339),
			Success:       true,
			Output:        "test",
		}
	})
	body := strings.NewReader(`{"whatif":true}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/start", body)
	s.handleStart(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"started"`) {
		t.Errorf("expected started status in body: %s", rec.Body.String())
	}
}

// TestHandleStop verifies POST /api/stop returns 200 + stopped status.
func TestHandleStop(t *testing.T) {
	s := testServer(t, "")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/stop", nil)
	s.handleStop(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"stopped"`) {
		t.Errorf("expected stopped in body: %s", rec.Body.String())
	}
}

// TestHandleTechniques verifies GET /api/techniques returns JSON array with both test techniques.
func TestHandleTechniques(t *testing.T) {
	s := testServer(t, "")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/techniques", nil)
	s.handleTechniques(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "T0001") || !strings.Contains(body, "T0002") {
		t.Errorf("expected both techniques in response: %s", body)
	}
}

// TestAuthMiddleware_rejectsWrongPassword verifies auth middleware returns 401 for wrong password
// and 200 for correct password.
func TestAuthMiddleware_rejectsWrongPassword(t *testing.T) {
	s := testServer(t, "correct-password")

	// Wrap a dummy handler with authMiddleware
	handler := s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Request with wrong password
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	req.SetBasicAuth("user", "wrong-password")
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}

	// Request with correct password should succeed
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	req2.SetBasicAuth("user", "correct-password")
	handler(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200 with correct password, got %d", rec2.Code)
	}
}
