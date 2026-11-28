package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strings"
)

func (h *CommerceHandler) SupplyOverdue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	surface := identity.SurfaceBackoffice
	if strings.HasPrefix(r.URL.Path, "/seller/") {
		surface = identity.SurfaceSeller
	}
	actor, ok := h.actor(w, r, surface)
	if !ok {
		return
	}
	out, err := h.service.ReadSupplyOverdue(r.Context(), actor, r.URL.Query().Get("store_id"), r.URL.Query().Get("search"))
	commerceResult(w, r, out, err)
}
