-- GoFlow jobs table: PostgreSQL as queue with FOR UPDATE SKIP LOCKED

CREATE TABLE IF NOT EXISTS jobs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status       VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    type         VARCHAR(50) NOT NULL,
    payload      JSONB,
    result       JSONB,
    retries      INT NOT NULL DEFAULT 0,
    max_retries  INT NOT NULL DEFAULT 3,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at   TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error   TEXT,
    priority     INT NOT NULL DEFAULT 0
);

-- Worker polling: next eligible job
CREATE INDEX idx_jobs_poll ON jobs (status, scheduled_at) WHERE status = 'pending';

-- Priority queue: higher priority first, then scheduled_at
CREATE INDEX idx_jobs_priority_poll ON jobs (status, priority DESC, scheduled_at) WHERE status = 'pending';

-- GET /jobs/:id
-- (id is PK, already indexed)

COMMENT ON TABLE jobs IS 'GoFlow job queue; workers claim via SELECT FOR UPDATE SKIP LOCKED';
