package httptransport

import (
	"net/http"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func (h *CommerceHandler) OrderOperations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	surface := identity.SurfaceBackoffice
	if strings.HasPrefix(r.URL.Path, "/seller/") {
		surface = identity.SurfaceSeller
	}
	if strings.HasPrefix(r.URL.Path, "/storefront/") {
		surface = identity.SurfaceCustomer
	}
	a, ok := h.actor(w, r, surface)
	if !ok {
		return
	}
	id, parent := r.PathValue("id"), ""
	if surface == identity.SurfaceCustomer {
		id, parent = r.PathValue("order"), r.PathValue("id")
	}
	if r.Method == http.MethodGet {
		result, err := h.service.ReadOrderOperations(r.Context(), a, id, parent)
		commerceResult(w, r, result, err)
		return
	}
	var in backoffice.OrderOperationInput
	if !commerceBody(w, r, &in) {
		return
	}
	if in.OrderID != "" && in.OrderID != id {
		commerceResult(w, r, nil, backoffice.ErrInvalid)
		return
	}
	in.OrderID, in.PurchaseID = id, parent
	if surface == identity.SurfaceCustomer {
		if in.Action != "" && in.Action != "deliver" {
			commerceResult(w, r, nil, backoffice.ErrInvalid)
			return
		}
		in.Action = "deliver"
	}
	result, err := h.service.TransitionOrder(r.Context(), a, in, r.Header.Get("Idempotency-Key"))
	commerceResult(w, r, result, err)
}
