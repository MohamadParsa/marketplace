package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strings"
)

func (h *CommerceHandler) ShippingChoices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	actor, ok := h.actor(w, r, identity.SurfaceCustomer)
	if !ok {
		return
	}
	out, err := h.service.ReadShippingChoices(r.Context(), actor, r.URL.Query().Get("address_id"), r.URL.Query().Get("purchase_id"))
	commerceResult(w, r, out, err)
}

func (h *CommerceHandler) ShippingCapacity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	surface := identity.SurfaceBackoffice
	if strings.HasPrefix(r.URL.Path, "/seller/") {
		surface = identity.SurfaceSeller
	}
	actor, ok := h.actor(w, r, surface)
	if !ok {
		return
	}
	q := r.URL.Query()
	out, err := h.service.ReadShippingCapacity(r.Context(), actor, q.Get("store_id"), q.Get("method_id"), q.Get("date"))
	commerceResult(w, r, out, err)
}
