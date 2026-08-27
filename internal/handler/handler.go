package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/wkulhane/bmw-loxone-bridge/internal/store"
)

type Handler struct {
	Store       *store.Store
	StartTime   time.Time
	RefreshFunc func()
}

func New(s *store.Store) *Handler {
	return &Handler{
		Store:     s,
		StartTime: time.Now(),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/vehicle", h.Vehicle)
	mux.HandleFunc("GET /api/battery", h.Battery)
	mux.HandleFunc("GET /api/location", h.Location)
	mux.HandleFunc("GET /api/status", h.Status)
	mux.HandleFunc("GET /api/fuel", h.Fuel)
	mux.HandleFunc("GET /api/raw/{name...}", h.Raw)
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("POST /api/refresh", h.Refresh)
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func flatMap(points map[string]store.DataPoint) map[string]any {
	result := make(map[string]any, len(points)+1)
	var latest time.Time
	for key, dp := range points {
		if dp.Numeric != nil {
			result[key] = *dp.Numeric
		} else {
			result[key] = dp.Value
		}
		if dp.Timestamp.After(latest) {
			latest = dp.Timestamp
		}
	}
	if !latest.IsZero() {
		result["last_updated"] = latest.Format(time.RFC3339)
	}
	return result
}
