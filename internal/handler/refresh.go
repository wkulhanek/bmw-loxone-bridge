package handler

import "net/http"

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if h.RefreshFunc == nil {
		http.Error(w, "refresh not available", http.StatusServiceUnavailable)
		return
	}
	h.RefreshFunc()
	writeJSON(w, map[string]string{"status": "ok", "message": "REST API refresh triggered"})
}
