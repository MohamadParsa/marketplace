package httptransport

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"net/http"
)

type HomeHandler struct{ service *backoffice.Service }

func NewHomeHandler(s *backoffice.Service) *HomeHandler { return &HomeHandler{s} }
func (h *HomeHandler) Read(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.PublicHome(r.Context())
	if err != nil {
		writeError(w, 503, "catalog_unavailable")
		return
	}
	encoded, err := publicCatalogJSON(data)
	if err != nil {
		writeError(w, 500, "internal_error")
		return
	}
	sum := sha256.Sum256(encoded)
	tag := "\"" + hex.EncodeToString(sum[:]) + "\""
	w.Header().Set("ETag", tag)
	w.Header().Set("Cache-Control", "public, no-cache")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Header.Get("If-None-Match") == tag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.WriteHeader(200)
	_, _ = w.Write(encoded)
}
