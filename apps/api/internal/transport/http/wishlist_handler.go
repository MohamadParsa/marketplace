package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (h *CommerceHandler) Wishlist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Add("Vary", "Authorization")
	actor, ok := h.actor(w, r, identity.SurfaceCustomer)
	if !ok {
		return
	}
	id := r.PathValue("id")
	for k, values := range r.URL.Query() {
		if id != "" || (k != "page" && k != "page_size") || len(values) != 1 {
			writeError(w, 400, "invalid_request")
			return
		}
	}
	if id == "" {
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
		result, e := h.service.ReadWishlist(r.Context(), actor, page, size)
		commerceResult(w, r, result, e)
		return
	}
	if r.Method == "GET" {
		result, e := h.service.ReadWishlistMembership(r.Context(), actor, id)
		commerceResult(w, r, result, e)
		return
	}
	body, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
	if e != nil || strings.TrimSpace(string(body)) != "" {
		writeError(w, 400, "invalid_request")
		return
	}
	result, e := h.service.SetWishlist(r.Context(), actor, id, r.Method == "PUT")
	commerceResult(w, r, result, e)
}
