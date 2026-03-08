package handler

import (
	"net/http"

	"goflow/internal/metrics"
)

type MetricsHandler struct {
	Metric *metrics.Collector
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snap := h.Metric.Snapshot()
	writeJSON(w, http.StatusOK, snap)
}
