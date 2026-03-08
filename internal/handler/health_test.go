package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler_OK(t *testing.T) {
	h := &HealthHandler{Check: func() error { return nil }}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Errorf("response = %s, want status ok", rec.Body.String())
	}
}

func TestHealthHandler_Unhealthy(t *testing.T) {
	h := &HealthHandler{Check: func() error { return errFake }}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"unhealthy"`) {
		t.Errorf("response = %s, want status unhealthy", rec.Body.String())
	}
}

func TestHealthHandler_NoCheck(t *testing.T) {
	h := &HealthHandler{} // Check is nil
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 when Check is nil", rec.Code)
	}
}

type errType struct{ msg string }

func (e errType) Error() string { return e.msg }

var errFake = errType{msg: "fake error"}
