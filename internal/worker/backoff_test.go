package worker

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	initial := 2 * time.Second
	factor := 2.0
	// No jitter so results are deterministic
	jitterPct := 0.0

	tests := []struct {
		retryNum int
		wantMin  time.Duration // allow small tolerance for float
		wantMax  time.Duration
	}{
		{0, 2*time.Second - time.Millisecond, 2*time.Second + time.Millisecond},
		{1, 4*time.Second - time.Millisecond, 4*time.Second + time.Millisecond},
		{2, 8*time.Second - time.Millisecond, 8*time.Second + time.Millisecond},
		{3, 16*time.Second - time.Millisecond, 16*time.Second + time.Millisecond},
	}
	for _, tt := range tests {
		got := Backoff(initial, factor, tt.retryNum, jitterPct)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("Backoff(2s, 2, %d, 0) = %v, want in [%v, %v]", tt.retryNum, got, tt.wantMin, tt.wantMax)
		}
	}
}

func TestBackoff_NegativeRetryNumber(t *testing.T) {
	got := Backoff(time.Second, 2.0, -1, 0)
	if got < time.Second-time.Millisecond || got > time.Second+time.Millisecond {
		t.Errorf("Backoff with retryNum -1 should treat as 0, got %v", got)
	}
}

func TestBackoff_ZeroInitialDelay(t *testing.T) {
	got := Backoff(0, 2.0, 3, 0)
	if got != 0 {
		t.Errorf("Backoff(0, ...) = %v, want 0", got)
	}
}
