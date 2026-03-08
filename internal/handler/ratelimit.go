package handler

import (
	"net/http"
	"sync"
	"time"
)

// RateLimitMiddleware returns a middleware that limits requests per second per client (by RemoteAddr).
func RateLimitMiddleware(maxPerSecond int) func(http.Handler) http.Handler {
	type bucket struct {
		count int
		start time.Time
	}
	var mu sync.Mutex
	buckets := make(map[string]*bucket)
	tick := time.NewTicker(time.Second)
	go func() {
		for range tick.C {
			mu.Lock()
			for k, b := range buckets {
				b.count = 0
				b.start = time.Now()
				_ = b
				_ = k
			}
			mu.Unlock()
		}
	}()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			if key == "" {
				key = "unknown"
			}
			mu.Lock()
			b, ok := buckets[key]
			if !ok {
				b = &bucket{start: time.Now()}
				buckets[key] = b
			}
			if time.Since(b.start) > time.Second {
				b.count = 0
				b.start = time.Now()
			}
			b.count++
			allowed := b.count <= maxPerSecond
			mu.Unlock()
			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
