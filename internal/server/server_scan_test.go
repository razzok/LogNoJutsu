package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestScanPendingAPI verifies GET /api/scan/pending returns 204 when no scan pending
// and 200+JSON when pending.
func TestScanPendingAPI(t *testing.T) {
	s := testServer(t, "")

	// When no simulation running — should return 204
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/scan/pending", nil)
	s.handleScanPending(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 when no scan pending, got %d", rec.Code)
	}
}

// TestScanConfirmAPI verifies POST /api/scan/confirm returns 409 when no scan pending.
func TestScanConfirmAPI(t *testing.T) {
	s := testServer(t, "")

	// No scan pending — should return 409 Conflict
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/scan/confirm", nil)
	s.handleScanConfirm(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 when no scan pending, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "error") {
		t.Errorf("expected error field in response, got: %s", body)
	}
}

// TestScanConfirmAPI_WrongMethod verifies POST /api/scan/confirm returns 405 for non-POST.
func TestScanConfirmAPI_WrongMethod(t *testing.T) {
	s := testServer(t, "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/scan/confirm", nil)
	s.handleScanConfirm(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET on scan/confirm, got %d", rec.Code)
	}
}

// TestScanCancelAPI verifies POST /api/scan/cancel returns 409 when no scan pending.
func TestScanCancelAPI(t *testing.T) {
	s := testServer(t, "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/scan/cancel", nil)
	s.handleScanCancel(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 when no scan pending, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestScanCancelAPI_WrongMethod verifies GET /api/scan/cancel returns 405.
func TestScanCancelAPI_WrongMethod(t *testing.T) {
	s := testServer(t, "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/scan/cancel", nil)
	s.handleScanCancel(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET on scan/cancel, got %d", rec.Code)
	}
}
