package handler

import "net/http"

func (h *Handler) Vehicle(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, flatMap(h.Store.GetAll()))
}
