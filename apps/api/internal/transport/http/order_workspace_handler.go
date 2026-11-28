package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strconv"
)

func (h *CommerceHandler) OrderWorkspace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	if id := r.PathValue("id"); id != "" {
		result, e := h.service.ReadOrderWorkspace(r.Context(), a, id)
		commerceResult(w, r, result, e)
		return
	}
	q := r.URL.Query()
	page, size := 0, 20
	var e error
	if q.Get("page") != "" {
		page, e = strconv.Atoi(q.Get("page"))
		if e != nil {
			commerceResult(w, r, nil, backoffice.ErrInvalid)
			return
		}
	}
	if q.Get("size") != "" {
		size, e = strconv.Atoi(q.Get("size"))
		if e != nil {
			commerceResult(w, r, nil, backoffice.ErrInvalid)
			return
		}
	}
	result, err := h.service.QueryOrderWorkspace(r.Context(), a, backoffice.OrderWorkspaceQuery{Search: q.Get("q"), Mobile: q.Get("mobile"), Store: q.Get("store"), Order: q.Get("order"), Payment: q.Get("payment"), Fulfillment: q.Get("fulfillment"), Page: page, Size: size})
	commerceResult(w, r, result, err)
}
