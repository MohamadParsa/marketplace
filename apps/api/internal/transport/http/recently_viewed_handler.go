package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (h *CommerceHandler) RecentlyViewed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Add("Vary", "Authorization")
	actor, ok := h.actor(w, r, identity.SurfaceCustomer)
	if !ok {
		return
	}
	id := r.PathValue("id")
	for key, values := range r.URL.Query() {
		if r.Method != "GET" || (key != "page" && key != "page_size") || len(values) != 1 {
			writeError(w, 400, "invalid_request")
			return
		}
	}
	if r.Method == "GET" {
		page, size := 1, 24
		for key, ptr := range map[string]*int{"page": &page, "page_size": &size} {
			if r.URL.Query().Has(key) {
				v, e := strconv.Atoi(r.URL.Query().Get(key))
				if e != nil {
					writeError(w, 400, "invalid_request")
					return
				}
				*ptr = v
			}
		}
		result, e := h.service.ReadRecentlyViewed(r.Context(), actor, page, size)
		commerceResult(w, result, e)
		return
	}
	body, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
	if e != nil || strings.TrimSpace(string(body)) != "" {
		writeError(w, 400, "invalid_request")
		return
	}
	if r.Method == "POST" {
		e = h.service.RecordRecentlyViewed(r.Context(), actor, id)
		commerceResult(w, map[string]bool{"recorded": e == nil}, e)
		return
	}
	removed, e := h.service.RemoveRecentlyViewed(r.Context(), actor, id)
	commerceResult(w, map[string]int{"removed": removed}, e)
}
