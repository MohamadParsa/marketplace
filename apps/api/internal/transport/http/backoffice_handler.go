package httptransport

import (
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"io"
	"net/http"
	"strings"
)

type BackofficeHandler struct {
	service *backoffice.Service
	auth    *auth.Service
}

func NewBackofficeHandler(service *backoffice.Service, authService *auth.Service) *BackofficeHandler {
	return &BackofficeHandler{service: service, auth: authService}
}
func (h *BackofficeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	surface := identity.SurfaceBackoffice
	if strings.HasPrefix(r.URL.Path, "/seller/") {
		surface = identity.SurfaceSeller
	}
	session, err := h.auth.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), surface)
	if err != nil {
		writeError(w, 401, "unauthorized")
		return
	}
	if r.Method == http.MethodGet {
		state, err := h.service.Read(r.Context(), session.User)
		if err != nil {
			iconError(w, err)
			return
		}
		writeJSON(w, 200, state)
		return
	}
	var command backoffice.Command
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&command); err != nil {
		writeError(w, 400, "invalid_request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, 400, "invalid_request")
		return
	}
	err = h.service.Execute(r.Context(), session.User, command)
	if err != nil {
		status := 500
		code := "internal_error"
		switch {
		case errors.Is(err, backoffice.ErrForbidden):
			status = 403
		case errors.Is(err, backoffice.ErrInvalid):
			status = 400
		case errors.Is(err, backoffice.ErrConflict):
			status = 409
		case errors.Is(err, backoffice.ErrNotFound):
			status = 404
		}
		if status != 500 {
			code = strings.TrimSpace(err.Error())
		}
		writeError(w, status, code)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
