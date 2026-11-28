package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
)

func (h *CommerceHandler) OperationReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	if r.Method == "GET" {
		out, err := h.service.ReadOperationLogStats(r.Context(), a)
		commerceResult(w, r, out, err)
		return
	}
	var in struct {
		operationlog.Query
		StoreID string `json:"store_id"`
	}
	if !commerceBody(w, r, &in) {
		return
	}
	if r.URL.Path == "/backoffice/observability/funnel/query" {
		out, err := h.service.ReadActivityFunnel(r.Context(), a, in.Query)
		commerceResult(w, r, out, err)
		return
	}
	if r.URL.Path == "/backoffice/observability/audit/query" {
		out, err := h.service.ReadAuditReport(r.Context(), a, in.Query, in.StoreID)
		commerceResult(w, r, out, err)
		return
	}
	out, err := h.service.ReadOperationLogs(r.Context(), a, in.Query)
	commerceResult(w, r, out, err)
}
