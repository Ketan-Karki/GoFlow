package repository

import (
	"context"
	"time"

	"goflow/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, jobType model.JobType, payload []byte, priority int) (*model.Job, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO jobs (status, type, payload, priority)
		VALUES ('pending', $1, $2, $3)
		RETURNING id, status, type, payload, result, retries, max_retries,
		          scheduled_at, started_at, completed_at, created_at, updated_at, last_error, priority
	`, jobType, payload, priority)
	return scanJob(row)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, status, type, payload, result, retries, max_retries,
		       scheduled_at, started_at, completed_at, created_at, updated_at, last_error, priority
		FROM jobs WHERE id = $1
	`, id)
	return scanJob(row)
}

// ClaimNext claims one pending job (FOR UPDATE SKIP LOCKED), sets status to processing, increments retries.
// Returns nil, nil if no job available.
func (r *Repository) ClaimNext(ctx context.Context) (*model.Job, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		SELECT id, status, type, payload, result, retries, max_retries,
		       scheduled_at, started_at, completed_at, created_at, updated_at, last_error, priority
		FROM jobs
		WHERE status = 'pending' AND scheduled_at <= now()
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`)
	job, err := scanJob(row)
	if err != nil || job == nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE jobs
		SET status = 'processing', started_at = now(), retries = retries + 1, updated_at = now()
		WHERE id = $1
	`, job.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	job.Status = model.StatusProcessing
	job.Retries++
	now := time.Now()
	job.StartedAt = &now
	return job, nil
}

func (r *Repository) MarkCompleted(ctx context.Context, id uuid.UUID, result []byte) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE jobs
		SET status = 'completed', result = $2, completed_at = now(), updated_at = now()
		WHERE id = $1
	`, id, result)
	return err
}

// ScheduleRetry sets status back to pending and scheduled_at = now() + backoffDelay.
func (r *Repository) ScheduleRetry(ctx context.Context, id uuid.UUID, backoffDelay time.Duration, lastError string) error {
	scheduledAt := time.Now().Add(backoffDelay)
	_, err := r.pool.Exec(ctx, `
		UPDATE jobs
		SET status = 'pending', last_error = $2, scheduled_at = $3, updated_at = now()
		WHERE id = $1
	`, id, lastError, scheduledAt)
	return err
}

func (r *Repository) MarkFailed(ctx context.Context, id uuid.UUID, lastError string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE jobs
		SET status = 'failed', last_error = $2, completed_at = now(), updated_at = now()
		WHERE id = $1
	`, id, lastError)
	return err
}
