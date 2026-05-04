package server

import (
	"net/http"
	"strings"
)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/properties" && r.Method == http.MethodPost:
		h.CreateProperty(w, r)
	case strings.HasPrefix(path, "/properties/") && r.Method == http.MethodPut:
		h.UpdateProperty(w, r)
	case strings.HasSuffix(path, "/offline") && r.Method == http.MethodPost:
		h.OfflineProperty(w, r)
	case strings.HasSuffix(path, "/sold") && r.Method == http.MethodPost:
		h.SellProperty(w, r)
	case path == "/properties" && r.Method == http.MethodGet:
		h.GetLandlordProperties(w, r)
	case strings.HasPrefix(path, "/properties/") && r.Method == http.MethodGet:
		h.GetProperty(w, r)
	case path == "/properties/filter" && r.Method == http.MethodPost:
		h.FilterProperties(w, r)
	case path == "/favorites" && r.Method == http.MethodPost:
		h.AddFavorite(w, r)
	case path == "/favorites" && r.Method == http.MethodDelete:
		h.RemoveFavorite(w, r)
	case path == "/favorites" && r.Method == http.MethodGet:
		h.GetFavorites(w, r)
	case path == "/favorites/check" && r.Method == http.MethodGet:
		h.CheckFavorite(w, r)
	default:
		http.NotFound(w, r)
	}
}
