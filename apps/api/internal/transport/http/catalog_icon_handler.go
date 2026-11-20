package httptransport

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type CatalogIconHandler struct {
	service *backoffice.Service
	auth    *auth.Service
}

func NewCatalogIconHandler(service *backoffice.Service, authService *auth.Service) *CatalogIconHandler {
	return &CatalogIconHandler{service, authService}
}

func (h *CatalogIconHandler) Read(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.ReadCatalogIcon(r.Context(), r.PathValue("id"))
	if err != nil {
		iconError(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *CatalogIconHandler) Upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	session, err := h.auth.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), identity.SurfaceBackoffice)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if session.User.Role != identity.RoleSuperAdmin && session.User.Role != identity.RoleOperations {
		iconError(w, backoffice.ErrForbidden)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, backoffice.MaxIconBytes+16*1024)
	if err := r.ParseMultipartForm(backoffice.MaxIconBytes + 16*1024); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			iconError(w, backoffice.ErrIconTooLarge)
		} else {
			iconError(w, backoffice.ErrInvalid)
		}
		return
	}
	defer r.MultipartForm.RemoveAll()
	for key, fields := range r.MultipartForm.Value {
		if (key != "version" && key != "idempotency_key") || len(fields) != 1 {
			iconError(w, backoffice.ErrInvalid)
			return
		}
	}
	if len(r.MultipartForm.File) != 1 || len(r.MultipartForm.File["file"]) != 1 {
		iconError(w, backoffice.ErrInvalid)
		return
	}
	version, err := strconv.Atoi(r.FormValue("version"))
	if err != nil {
		iconError(w, backoffice.ErrInvalid)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		iconError(w, backoffice.ErrInvalid)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, backoffice.MaxIconBytes+1))
	if err != nil {
		iconError(w, backoffice.ErrInvalid)
		return
	}
	err = h.service.SetCatalogIcon(r.Context(), session.User, r.PathValue("kind"), r.PathValue("target"), version, r.FormValue("idempotency_key"), data)
	if err != nil {
		iconError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func iconError(w http.ResponseWriter, err error) {
	status, code := http.StatusInternalServerError, "internal_error"
	switch {
	case errors.Is(err, backoffice.ErrForbidden):
		status, code = http.StatusForbidden, "forbidden"
	case errors.Is(err, backoffice.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, backoffice.ErrConflict):
		status, code = http.StatusConflict, "version_conflict"
	case errors.Is(err, backoffice.ErrIconTooLarge):
		status, code = http.StatusRequestEntityTooLarge, "icon_too_large"
	case errors.Is(err, backoffice.ErrInvalidIcon):
		status, code = http.StatusBadRequest, "invalid_icon"
	case errors.Is(err, backoffice.ErrInvalid):
		status, code = http.StatusBadRequest, "invalid_request"
	}
	writeError(w, status, code)
}
