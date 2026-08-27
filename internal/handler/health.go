package handler

import (
	"net/http"
	"time"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	latest := h.Store.LatestTimestamp()
	lastMsg := ""
	if !latest.IsZero() {
		lastMsg = latest.Format(time.RFC3339)
	}

	writeJSON(w, map[string]any{
		"status":         "ok",
		"data_points":    h.Store.Count(),
		"last_message":   lastMsg,
		"uptime_seconds": int(time.Since(h.StartTime).Seconds()),
	})
}
