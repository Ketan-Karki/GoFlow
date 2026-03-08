package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"goflow/internal/metrics"
	"goflow/internal/model"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// JobRepo is the subset of job storage used by the HTTP handlers (for testing with mocks).
type JobRepo interface {
	Create(ctx context.Context, jobType model.JobType, payload []byte, priority int) (*model.Job, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Job, error)
}

type JobsHandler struct {
	Repo   JobRepo
	Metric *metrics.Collector
}

func (h *JobsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input model.CreateJobInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if input.Type == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type is required"})
		return
	}
	switch input.Type {
	case model.TypeReport, model.TypeImage, model.TypeEmail, model.TypeHeavyTask:
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid type: use report, image, email, heavy_task"})
		return
	}
	payload := input.Payload
	if payload == nil {
		payload = []byte("{}")
	}
	job, err := h.Repo.Create(r.Context(), input.Type, payload, input.Priority)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.Metric.JobCreated()
	writeJSON(w, http.StatusCreated, model.JobResponse{
		ID:        job.ID,
		Status:    job.Status,
		CreatedAt: job.CreatedAt,
	})
}

func (h *JobsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid job id"})
		return
	}
	job, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}
