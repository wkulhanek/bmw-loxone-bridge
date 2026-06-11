package handler

import (
	"net/http"
	"strings"

	"github.com/wkulhane/bmw-loxone-bridge/internal/names"
	"github.com/wkulhane/bmw-loxone-bridge/internal/store"
)

func (h *Handler) Battery(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, flatMap(filterByCategory(h.Store.GetAll(), "battery")))
}

func filterByCategory(all map[string]store.DataPoint, category string) map[string]store.DataPoint {
	prefixes := names.CategoryPrefixes(category)
	if len(prefixes) == 0 {
		return nil
	}
	return filterByPrefix(all, prefixes...)
}

func filterByPrefix(all map[string]store.DataPoint, prefixes ...string) map[string]store.DataPoint {
	result := make(map[string]store.DataPoint)
	for key, dp := range all {
		for _, prefix := range prefixes {
			if strings.HasPrefix(key, prefix) {
				result[key] = dp
				break
			}
		}
	}
	return result
}
