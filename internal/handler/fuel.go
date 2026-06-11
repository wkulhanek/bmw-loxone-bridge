package handler

import "net/http"

func (h *Handler) Fuel(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, flatMap(filterByCategory(h.Store.GetAll(), "fuel")))
}
