package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strconv"
	"strings"
)

func (h *CommerceHandler) Shipments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	surface := identity.SurfaceBackoffice
	if strings.HasPrefix(r.URL.Path, "/seller/") {
		surface = identity.SurfaceSeller
	}
	a, ok := h.actor(w, r, surface)
	if !ok {
		return
	}
	if id := r.PathValue("id"); id != "" {
		result, e := h.service.ReadShipment(r.Context(), a, id)
		commerceResult(w, r, result, e)
		return
	}
	q := r.URL.Query()
	page, size := 0, 20
	var e error
	if q.Has("page") {
		page, e = strconv.Atoi(q.Get("page"))
		if e != nil {
			commerceResult(w, r, nil, backoffice.ErrInvalid)
			return
		}
	}
	if q.Has("size") {
		size, e = strconv.Atoi(q.Get("size"))
		if e != nil {
			commerceResult(w, r, nil, backoffice.ErrInvalid)
			return
		}
	}
	result, err := h.service.QueryShipments(r.Context(), a, backoffice.ShipmentQuery{Search: q.Get("q"), Store: q.Get("store"), Status: q.Get("status"), Page: page, Size: size})
	commerceResult(w, r, result, err)
}
