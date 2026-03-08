package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

type JobType string

const (
	TypeReport    JobType = "report"
	TypeImage     JobType = "image"
	TypeEmail     JobType = "email"
	TypeHeavyTask JobType = "heavy_task"
)

type Job struct {
	ID          uuid.UUID       `json:"id"`
	Status      JobStatus       `json:"status"`
	Type        JobType         `json:"type"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	Retries     int             `json:"retries"`
	MaxRetries  int             `json:"max_retries"`
	ScheduledAt time.Time       `json:"scheduled_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	LastError   *string         `json:"last_error,omitempty"`
	Priority    int             `json:"priority"`
}

type CreateJobInput struct {
	Type     JobType         `json:"type"`
	Payload  json.RawMessage `json:"payload,omitempty"`
	Priority int             `json:"priority,omitempty"` // higher = process first
}

type JobResponse struct {
	ID        uuid.UUID       `json:"id"`
	Status    JobStatus       `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}
