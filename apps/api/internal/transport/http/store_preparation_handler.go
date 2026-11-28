package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strings"
)

func (h *CommerceHandler) StorePreparation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	surface := identity.SurfaceBackoffice
	if strings.HasPrefix(r.URL.Path, "/seller/") {
		surface = identity.SurfaceSeller
	}
	actor, ok := h.actor(w, r, surface)
	if !ok {
		return
	}
	if strings.HasSuffix(r.URL.Path, "/preparation/batch") {
		var in struct {
			backoffice.StorePreparationBatchInput
			Stores []struct {
				ID      string `json:"id"`
				Version *int   `json:"version"`
			} `json:"stores"`
			LeadDays *int `json:"lead_days"`
		}
		if !commerceBody(w, r, &in) {
			return
		}
		if in.LeadDays == nil {
			commerceError(w, backoffice.ErrInvalid)
			return
		}
		in.StorePreparationBatchInput.LeadDays = *in.LeadDays
		for _, item := range in.Stores {
			if item.Version == nil {
				commerceError(w, backoffice.ErrInvalid)
				return
			}
			in.StorePreparationBatchInput.Stores = append(in.StorePreparationBatchInput.Stores, backoffice.StorePreparationVersion{ID: item.ID, Version: *item.Version})
		}
		out, err := h.service.SaveStorePreparationBatch(r.Context(), actor, in.StorePreparationBatchInput, r.Header.Get("Idempotency-Key"))
		commerceResult(w, r, out, err)
		return
	}
	if r.Method == http.MethodGet {
		out, err := h.service.ReadStorePreparations(r.Context(), actor)
		commerceResult(w, r, out, err)
		return
	}
	var in struct {
		backoffice.StorePreparationInput
		Version  *int `json:"version"`
		LeadDays *int `json:"lead_days"`
	}
	if !commerceBody(w, r, &in) {
		return
	}
	if in.Version == nil || in.LeadDays == nil {
		commerceError(w, backoffice.ErrInvalid)
		return
	}
	in.StorePreparationInput.Version = *in.Version
	in.StorePreparationInput.LeadDays = *in.LeadDays
	if r.Method == http.MethodPost {
		out, err := h.service.PreviewStorePreparation(r.Context(), actor, r.PathValue("id"), in.StorePreparationInput)
		commerceResult(w, r, out, err)
		return
	}
	out, err := h.service.SaveStorePreparation(r.Context(), actor, r.PathValue("id"), in.StorePreparationInput, r.Header.Get("Idempotency-Key"))
	commerceResult(w, r, out, err)
}
