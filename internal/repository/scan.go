package repository

import (
	"time"

	"goflow/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func scanJob(row pgx.Row) (*model.Job, error) {
	var (
		id          uuid.UUID
		status      string
		jobType     string
		payload     []byte
		result      []byte
		retries     int
		maxRetries  int
		scheduledAt interface{}
		startedAt   interface{}
		completedAt interface{}
		createdAt   interface{}
		updatedAt   interface{}
		lastError   *string
		priority    int
	)
	err := row.Scan(
		&id, &status, &jobType, &payload, &result, &retries, &maxRetries,
		&scheduledAt, &startedAt, &completedAt, &createdAt, &updatedAt, &lastError, &priority,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	j := &model.Job{
		ID:          id,
		Status:      model.JobStatus(status),
		Type:        model.JobType(jobType),
		Payload:     payload,
		Result:      result,
		Retries:     retries,
		MaxRetries:  maxRetries,
		Priority:    priority,
		LastError:   lastError,
	}
	j.ScheduledAt = mustTime(scheduledAt)
	j.CreatedAt = mustTime(createdAt)
	j.UpdatedAt = mustTime(updatedAt)
	if startedAt != nil {
		t := mustTime(startedAt)
		j.StartedAt = &t
	}
	if completedAt != nil {
		t := mustTime(completedAt)
		j.CompletedAt = &t
	}
	return j, nil
}

func mustTime(v interface{}) time.Time {
	// pgx returns time.Time for timestamptz
	if t, ok := v.(time.Time); ok {
		return t
	}
	return time.Time{}
}
