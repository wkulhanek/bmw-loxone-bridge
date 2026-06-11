package handler

import "net/http"

func (h *Handler) Location(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, flatMap(filterByCategory(h.Store.GetAll(), "location")))
}
