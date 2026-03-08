package handler

import "net/http"

// HealthHandler responds to GET /health. If Check is set and returns an error, responds 503; otherwise 200.
type HealthHandler struct {
	Check func() error
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Check != nil {
		if err := h.Check(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
