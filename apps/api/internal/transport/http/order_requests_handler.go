package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strings"
)

func (h *CommerceHandler) OrderRequests(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == http.MethodGet {
		result, err := h.service.ReadOrderRequests(r.Context(), a, r.PathValue("id"))
		commerceResult(w, r, result, err)
		return
	}
	var in backoffice.OrderRequestInput
	if !commerceBody(w, r, &in) {
		return
	}
	if in.PurchaseID != "" && in.PurchaseID != r.PathValue("id") {
		commerceResult(w, r, nil, backoffice.ErrInvalid)
		return
	}
	in.PurchaseID = r.PathValue("id")
	result, err := h.service.CreateOrderRequest(r.Context(), a, in, r.Header.Get("Idempotency-Key"))
	commerceResult(w, r, result, err)
}

func (h *CommerceHandler) OrderRequestReview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	var in backoffice.OrderRequestReview
	if !commerceBody(w, r, &in) {
		return
	}
	if in.ID != "" && in.ID != r.PathValue("request") {
		commerceResult(w, r, nil, backoffice.ErrInvalid)
		return
	}
	in.ID = r.PathValue("request")
	result, err := h.service.ReviewOrderRequest(r.Context(), a, in, r.Header.Get("Idempotency-Key"))
	commerceResult(w, r, result, err)
}
