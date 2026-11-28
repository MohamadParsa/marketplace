package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
)

func (h *CommerceHandler) GatewaySettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	if r.Method == http.MethodGet {
		result, err := h.service.ReadGatewaySettings(r.Context(), a)
		commerceResult(w, r, result, err)
		return
	}
	var in backoffice.GatewaySettingsInput
	if !commerceBody(w, r, &in) {
		return
	}
	result, err := h.service.SaveGatewaySettings(r.Context(), a, in, r.Header.Get("Idempotency-Key"))
	commerceResult(w, r, result, err)
}
