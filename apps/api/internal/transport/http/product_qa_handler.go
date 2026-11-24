package httptransport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type ProductQAHandler struct {
	service *backoffice.Service
	auth    *auth.Service
}

func NewProductQAHandler(s *backoffice.Service, a *auth.Service) *ProductQAHandler {
	return &ProductQAHandler{service: s, auth: a}
}

func (h *ProductQAHandler) Public(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	after, ok := qaCursor(w, r)
	if !ok {
		return
	}
	page, err := h.service.PublicProductQA(r.Context(), r.PathValue("id"), after)
	if err != nil {
		qaError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *ProductQAHandler) Ask(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r, identity.SurfaceCustomer)
	if !ok {
		return
	}
	var input struct {
		Text string `json:"text"`
	}
	if !commerceBody(w, r, &input) {
		return
	}
	err := h.service.AskProductQuestion(r.Context(), actor, r.PathValue("id"), r.Header.Get("Idempotency-Key"), input.Text)
	qaMutationResult(w, err)
}

func (h *ProductQAHandler) Admin(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	if r.Method == http.MethodGet {
		after, ok := qaCursor(w, r)
		if !ok {
			return
		}
		page, err := h.service.AdminProductQA(r.Context(), actor, r.PathValue("id"), after)
		if err != nil {
			qaError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, page)
		return
	}
	var input backoffice.ProductQAMutation
	if !commerceBody(w, r, &input) {
		return
	}
	err := h.service.ManageProductQuestion(r.Context(), actor, r.PathValue("id"), r.PathValue("question"), r.Header.Get("Idempotency-Key"), input)
	qaMutationResult(w, err)
}

func (h *ProductQAHandler) actor(w http.ResponseWriter, r *http.Request, surface identity.Surface) (identity.User, bool) {
	w.Header().Set("Cache-Control", "private, no-store")
	session, err := h.auth.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), surface)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return identity.User{}, false
	}
	return session.User, true
}

func qaCursor(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := r.URL.Query().Get("after")
	if raw == "" {
		return 0, true
	}
	after, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || after < 1 {
		qaError(w, backoffice.ErrInvalid)
		return 0, false
	}
	return after, true
}

func qaError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, backoffice.ErrInvalid):
		writeError(w, http.StatusBadRequest, "invalid_request")
	case errors.Is(err, backoffice.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, backoffice.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
	case errors.Is(err, backoffice.ErrConflict):
		writeError(w, http.StatusConflict, "state_conflict")
	default:
		writeError(w, http.StatusServiceUnavailable, "qa_unavailable")
	}
}

func qaMutationResult(w http.ResponseWriter, err error) {
	if err != nil {
		qaError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
