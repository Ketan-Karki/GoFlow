package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goflow/internal/metrics"
	"goflow/internal/model"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// mockJobRepo is a in-memory JobRepo for tests.
type mockJobRepo struct {
	job *model.Job
	err error
}

func (m *mockJobRepo) Create(ctx context.Context, jobType model.JobType, payload []byte, priority int) (*model.Job, error) {
	if m.err != nil {
		return nil, m.err
	}
	id := uuid.Must(uuid.NewRandom())
	job := &model.Job{ID: id, Status: model.StatusPending, Type: jobType, Payload: payload, Priority: priority}
	m.job = job
	return job, nil
}

func (m *mockJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.job != nil && m.job.ID == id {
		return m.job, nil
	}
	return nil, nil
}

func TestJobsHandler_Create(t *testing.T) {
	met := metrics.New()
	repo := &mockJobRepo{}
	h := &JobsHandler{Repo: repo, Metric: met}

	r := chi.NewRouter()
	r.Post("/jobs", h.Create)

	body := `{"type": "report", "payload": {"id": "r1"}}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Create status = %d, want 201", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", rec.Header().Get("Content-Type"))
	}
	// Response should contain "id" and "pending"
	resp := rec.Body.String()
	if !strings.Contains(resp, `"id"`) || !strings.Contains(resp, `"pending"`) {
		t.Errorf("response missing id or status: %s", resp)
	}
	snap := met.Snapshot()
	if snap.TotalJobs != 1 {
		t.Errorf("TotalJobs = %d, want 1", snap.TotalJobs)
	}
}

func TestJobsHandler_Create_BadJSON(t *testing.T) {
	met := metrics.New()
	repo := &mockJobRepo{}
	h := &JobsHandler{Repo: repo, Metric: met}
	r := chi.NewRouter()
	r.Post("/jobs", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Create bad JSON status = %d, want 400", rec.Code)
	}
}

func TestJobsHandler_Create_InvalidType(t *testing.T) {
	met := metrics.New()
	repo := &mockJobRepo{}
	h := &JobsHandler{Repo: repo, Metric: met}
	r := chi.NewRouter()
	r.Post("/jobs", h.Create)

	body := `{"type": "invalid_type"}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Create invalid type status = %d, want 400", rec.Code)
	}
}

func TestJobsHandler_GetByID(t *testing.T) {
	met := metrics.New()
	repo := &mockJobRepo{}
	h := &JobsHandler{Repo: repo, Metric: met}
	// Create a job first so GetByID has something to return
	_, _ = repo.Create(context.Background(), model.TypeReport, []byte("{}"), 0)
	id := repo.job.ID

	r := chi.NewRouter()
	r.Get("/jobs/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/jobs/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GetByID status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"report"`) {
		t.Errorf("response should contain job type: %s", rec.Body.String())
	}
}

func TestJobsHandler_GetByID_NotFound(t *testing.T) {
	met := metrics.New()
	repo := &mockJobRepo{} // no job created
	h := &JobsHandler{Repo: repo, Metric: met}
	r := chi.NewRouter()
	r.Get("/jobs/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/jobs/"+uuid.Must(uuid.NewRandom()).String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GetByID not found status = %d, want 404", rec.Code)
	}
}

func TestJobsHandler_GetByID_InvalidUUID(t *testing.T) {
	met := metrics.New()
	repo := &mockJobRepo{}
	h := &JobsHandler{Repo: repo, Metric: met}
	r := chi.NewRouter()
	r.Get("/jobs/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/jobs/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("GetByID bad uuid status = %d, want 400", rec.Code)
	}
}
