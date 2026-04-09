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
		cfg:      Config{Password: password, Version: "test-v0.0.0"},
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

// TestHandleInfo_returnsVersion verifies GET /api/info returns 200 with JSON containing version.
func TestHandleInfo_returnsVersion(t *testing.T) {
	s := testServer(t, "")
	s.cfg.Version = "test-v1.2.3"
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	s.handleInfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"version":"test-v1.2.3"`) {
		t.Errorf("expected version in body: %s", body)
	}
}

// TestHandleInfo_noAuthRequired verifies GET /api/info succeeds without credentials.
func TestHandleInfo_noAuthRequired(t *testing.T) {
	s := testServer(t, "secret-password")
	s.cfg.Version = "test-v2.0.0"

	// Call handleInfo directly (not through authMiddleware)
	// This proves the route handler works without credentials
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	// Do NOT set BasicAuth — simulates unauthenticated request
	s.handleInfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 without auth, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"version":"test-v2.0.0"`) {
		t.Errorf("expected version in body: %s", body)
	}
}

// TestRegisterRoutes_infoNoAuth verifies /api/info is registered without authMiddleware in the mux.
func TestRegisterRoutes_infoNoAuth(t *testing.T) {
	s := testServer(t, "secret-password")
	s.cfg.Version = "route-test"
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	// Hit /api/info through the mux WITHOUT credentials
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 via mux without auth, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"version":"route-test"`) {
		t.Errorf("expected version in mux response: %s", rec.Body.String())
	}
}

// TestHandlePoCDays_idle verifies GET /api/poc/days returns 200 with empty JSON array when engine is idle.
func TestHandlePoCDays_idle(t *testing.T) {
	s := testServer(t, "")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/poc/days", nil)
	s.handlePoCDays(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := strings.TrimSpace(rec.Body.String())
	if body != "[]" {
		t.Errorf("expected empty JSON array [], got: %s", body)
	}
}

// TestHandlePoCDays_auth verifies /api/poc/days requires authentication when password is set.
func TestHandlePoCDays_auth(t *testing.T) {
	s := testServer(t, "secret123")
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	// Without auth — should get 401
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/poc/days", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", rec.Code)
	}

	// With auth — should get 200
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/poc/days", nil)
	req2.SetBasicAuth("", "secret123")
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 with auth, got %d", rec2.Code)
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
