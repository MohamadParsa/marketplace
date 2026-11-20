package httptransport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/application/seller"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type SellerOffersHandler struct {
	offers *seller.Offers
	auth   *auth.Service
}

func NewSellerOffersHandler(offers *seller.Offers, authService *auth.Service) *SellerOffersHandler {
	return &SellerOffersHandler{offers: offers, auth: authService}
}

func (h *SellerOffersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := h.auth.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), identity.SurfaceSeller)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	storeID := r.PathValue("storeID")
	if r.Method == http.MethodGet {
		offers, err := h.offers.List(r.Context(), session.User.ID, storeID)
		if err != nil {
			writeSellerOffersError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, offers)
		return
	}
	var input struct {
		Stock   *int `json:"stock"`
		Version *int `json:"version"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || input.Stock == nil || input.Version == nil || decoder.Decode(&struct{}{}) != io.EOF {
		writeSellerOffersError(w, backoffice.ErrInvalid)
		return
	}
	err = h.offers.SetStock(r.Context(), session.User.ID, storeID, r.PathValue("offerID"), *input.Stock, *input.Version)
	if err != nil {
		writeSellerOffersError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeSellerOffersError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, backoffice.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, backoffice.ErrInvalid):
		writeError(w, http.StatusBadRequest, "invalid_request")
	case errors.Is(err, backoffice.ErrConflict):
		writeError(w, http.StatusConflict, "version_conflict")
	case errors.Is(err, backoffice.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error")
	}
}
