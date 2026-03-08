package metrics

import (
	"testing"
	"time"
)

func TestCollector_JobCreated(t *testing.T) {
	c := New()
	c.JobCreated()
	c.JobCreated()
	snap := c.Snapshot()
	if snap.TotalJobs != 2 {
		t.Errorf("TotalJobs = %d, want 2", snap.TotalJobs)
	}
}

func TestCollector_JobCompleted(t *testing.T) {
	c := New()
	c.JobCompleted(100 * time.Millisecond)
	c.JobCompleted(200 * time.Millisecond)
	snap := c.Snapshot()
	if snap.CompletedJobs != 2 {
		t.Errorf("CompletedJobs = %d, want 2", snap.CompletedJobs)
	}
	// 0.1 + 0.2 = 0.3 sec total, avg 0.15
	avg := snap.AvgProcessingTimeSec
	if avg < 0.14 || avg > 0.16 {
		t.Errorf("AvgProcessingTimeSec = %v, want ~0.15", avg)
	}
}

func TestCollector_JobFailed(t *testing.T) {
	c := New()
	c.JobFailed()
	c.JobFailed()
	snap := c.Snapshot()
	if snap.FailedJobs != 2 {
		t.Errorf("FailedJobs = %d, want 2", snap.FailedJobs)
	}
}

func TestCollector_Snapshot_NoDivisionByZero(t *testing.T) {
	c := New()
	snap := c.Snapshot()
	if snap.AvgProcessingTimeSec != 0 {
		t.Errorf("AvgProcessingTimeSec with no jobs = %v, want 0", snap.AvgProcessingTimeSec)
	}
}

func TestCollector_SetPendingAndProcessing(t *testing.T) {
	c := New()
	c.SetPendingCount(5)
	c.SetProcessingCount(2)
	snap := c.Snapshot()
	if snap.PendingCount != 5 || snap.ProcessingCount != 2 {
		t.Errorf("PendingCount=%d ProcessingCount=%d, want 5 and 2", snap.PendingCount, snap.ProcessingCount)
	}
}
