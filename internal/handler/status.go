package handler

import "net/http"

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, flatMap(filterByCategory(h.Store.GetAll(), "status")))
}
