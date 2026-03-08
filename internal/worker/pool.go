package worker

import (
	"context"
	"log"
	"time"

	"goflow/internal/metrics"
	"goflow/internal/model"
	"goflow/internal/repository"
)

type Config struct {
	WorkerCount  int
	InitialDelay time.Duration
	Factor       float64
	JitterPct    float64
	PollInterval time.Duration
}

var DefaultConfig = Config{
	WorkerCount:  3,
	InitialDelay: 2 * time.Second,
	Factor:       2.0,
	JitterPct:    0.1,
	PollInterval: 500 * time.Millisecond,
}

type Pool struct {
	repo     *repository.Repository
	metrics  *metrics.Collector
	procs    map[model.JobType]Processor
	config   Config
}

func NewPool(repo *repository.Repository, metrics *metrics.Collector, config Config) *Pool {
	if config.WorkerCount <= 0 {
		config.WorkerCount = DefaultConfig.WorkerCount
	}
	if config.PollInterval <= 0 {
		config.PollInterval = DefaultConfig.PollInterval
	}
	return &Pool{
		repo:    repo,
		metrics: metrics,
		procs:   Processors(),
		config:  config,
	}
}

func (p *Pool) Run(ctx context.Context) {
	for i := 0; i < p.config.WorkerCount; i++ {
		workerID := i
		go p.runWorker(ctx, workerID)
	}
}

func (p *Pool) runWorker(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		job, err := p.repo.ClaimNext(ctx)
		if err != nil {
			log.Printf("[worker %d] claim error: %v", workerID, err)
			time.Sleep(p.config.PollInterval)
			continue
		}
		if job == nil {
			time.Sleep(p.config.PollInterval)
			continue
		}
		p.processJob(ctx, job)
	}
}

func (p *Pool) processJob(ctx context.Context, job *model.Job) {
	start := time.Now()
	proc, ok := p.procs[job.Type]
	if !ok {
		_ = p.repo.MarkFailed(ctx, job.ID, "unknown job type: "+string(job.Type))
		p.metrics.JobFailed()
		return
	}
	result, err := proc(ctx, job.Payload)
	if err != nil {
		retryNum := job.Retries - 1
		if retryNum < 0 {
			retryNum = 0
		}
		if job.Retries < job.MaxRetries {
			delay := Backoff(p.config.InitialDelay, p.config.Factor, retryNum, p.config.JitterPct)
			if scheduleErr := p.repo.ScheduleRetry(ctx, job.ID, delay, err.Error()); scheduleErr != nil {
				log.Printf("schedule retry failed for job %s: %v", job.ID, scheduleErr)
				_ = p.repo.MarkFailed(ctx, job.ID, err.Error())
				p.metrics.JobFailed()
			}
		} else {
			_ = p.repo.MarkFailed(ctx, job.ID, err.Error())
			p.metrics.JobFailed()
		}
		return
	}
	if err := p.repo.MarkCompleted(ctx, job.ID, result); err != nil {
		log.Printf("mark completed failed for job %s: %v", job.ID, err)
		return
	}
	p.metrics.JobCompleted(time.Since(start))
}
