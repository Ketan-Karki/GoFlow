package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"goflow/internal/metrics"
)

func TestMetricsHandler(t *testing.T) {
	met := metrics.New()
	met.JobCreated()
	met.JobCreated()
	met.JobCompleted(100 * time.Millisecond)

	h := &MetricsHandler{Metric: met}
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"total_jobs":2`) {
		t.Errorf("response should contain total_jobs:2, got %s", body)
	}
	if !strings.Contains(body, `"completed_jobs":1`) {
		t.Errorf("response should contain completed_jobs:1, got %s", body)
	}
	if !strings.Contains(body, `"avg_processing_time_seconds"`) {
		t.Errorf("response should contain avg_processing_time_seconds, got %s", body)
	}
}
