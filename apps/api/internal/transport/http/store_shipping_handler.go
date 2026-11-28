package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strings"
)

func (h *CommerceHandler) StoreShipping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	surface := identity.SurfaceBackoffice
	if strings.HasPrefix(r.URL.Path, "/seller/") {
		surface = identity.SurfaceSeller
	}
	actor, ok := h.actor(w, r, surface)
	if !ok {
		return
	}
	if r.Method == http.MethodGet {
		out, err := h.service.ReadStoreShipping(r.Context(), actor)
		commerceResult(w, r, out, err)
		return
	}
	var in struct {
		backoffice.StoreShippingCommand
		BaseVersion *int `json:"base_version"`
		Stores      []struct {
			ID      string `json:"id"`
			Version *int   `json:"version"`
		} `json:"stores"`
	}
	if !commerceBody(w, r, &in) {
		return
	}
	if in.BaseVersion == nil {
		commerceError(w, backoffice.ErrInvalid)
		return
	}
	in.StoreShippingCommand.BaseVersion = *in.BaseVersion
	for _, st := range in.Stores {
		if st.Version == nil {
			commerceError(w, backoffice.ErrInvalid)
			return
		}
		in.StoreShippingCommand.Stores = append(in.StoreShippingCommand.Stores, backoffice.StoreShippingTarget{ID: st.ID, Version: *st.Version})
	}
	out, err := h.service.SaveStoreShipping(r.Context(), actor, in.StoreShippingCommand, r.Header.Get("Idempotency-Key"))
	commerceResult(w, r, out, err)
}
