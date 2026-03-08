package worker

import (
	"math"
	"math/rand"
	"time"
)

// Backoff computes delay = initialDelay * (factor ^ retryNumber) with optional jitter.
// retryNumber is 0-based (first retry = 0).
func Backoff(initialDelay time.Duration, factor float64, retryNumber int, jitterPct float64) time.Duration {
	if retryNumber < 0 {
		retryNumber = 0
	}
	delay := float64(initialDelay) * math.Pow(factor, float64(retryNumber))
	if jitterPct > 0 {
		jitter := delay * jitterPct * (rand.Float64()*2 - 1)
		delay += jitter
	}
	if delay < 0 {
		delay = 0
	}
	return time.Duration(delay)
}
