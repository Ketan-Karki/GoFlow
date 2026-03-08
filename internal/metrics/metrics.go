package metrics

import (
	"sync"
	"time"
)

// Collector holds in-memory metrics for GET /metrics.
type Collector struct {
	mu sync.RWMutex

	TotalJobs      int64
	FailedJobs     int64
	CompletedJobs  int64
	PendingCount   int64
	ProcessingCount int64

	// Processing time in seconds (for avg)
	TotalProcessingTimeSec float64
	ProcessingCountForAvg int64
}

func New() *Collector {
	return &Collector{}
}

func (c *Collector) JobCreated() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TotalJobs++
}

func (c *Collector) JobCompleted(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CompletedJobs++
	c.TotalProcessingTimeSec += duration.Seconds()
	c.ProcessingCountForAvg++
}

func (c *Collector) JobFailed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FailedJobs++
}

func (c *Collector) SetPendingCount(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PendingCount = n
}

func (c *Collector) SetProcessingCount(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ProcessingCount = n
}

// Snapshot returns a copy of current metrics.
func (c *Collector) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	avgSec := 0.0
	if c.ProcessingCountForAvg > 0 {
		avgSec = c.TotalProcessingTimeSec / float64(c.ProcessingCountForAvg)
	}
	return Snapshot{
		TotalJobs:               c.TotalJobs,
		FailedJobs:              c.FailedJobs,
		CompletedJobs:           c.CompletedJobs,
		PendingCount:            c.PendingCount,
		ProcessingCount:         c.ProcessingCount,
		AvgProcessingTimeSec:    avgSec,
	}
}

type Snapshot struct {
	TotalJobs            int64   `json:"total_jobs"`
	FailedJobs           int64   `json:"failed_jobs"`
	CompletedJobs        int64   `json:"completed_jobs"`
	PendingCount         int64   `json:"pending_count"`
	ProcessingCount      int64   `json:"processing_count"`
	AvgProcessingTimeSec float64 `json:"avg_processing_time_seconds"`
}
