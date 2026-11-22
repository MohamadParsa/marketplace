package httptransport

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type ProductWorkspaceHandler struct {
	service *backoffice.Service
	auth    *auth.Service
}

func NewProductWorkspaceHandler(service *backoffice.Service, authService *auth.Service) *ProductWorkspaceHandler {
	return &ProductWorkspaceHandler{service, authService}
}
func (h *ProductWorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if strings.HasSuffix(r.URL.Path, "/resolve") {
		if len(r.URL.RawQuery) > 4096 {
			iconError(w, backoffice.ErrInvalid)
			return
		}
		query, e := url.ParseQuery(r.URL.RawQuery)
		if e != nil {
			iconError(w, backoffice.ErrInvalid)
			return
		}
		values := []catalog.AttributeValue{}
		for key, vs := range query {
			if !strings.HasPrefix(key, "attr.") || len(vs) != 1 {
				e = backoffice.ErrInvalid
				break
			}
			values = append(values, catalog.AttributeValue{AttributeID: strings.TrimPrefix(key, "attr."), Value: vs[0]})
		}
		if e != nil {
			iconError(w, e)
			return
		}
		result, err := h.service.ResolveVariant(r.Context(), session.User, r.PathValue("id"), values)
		if err != nil {
			iconError(w, err)
			return
		}
		writeJSON(w, 200, result)
		return
	}
	result, err := h.service.ReadProductWorkspace(r.Context(), session.User, r.PathValue("id"))
	if err != nil {
		iconError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
