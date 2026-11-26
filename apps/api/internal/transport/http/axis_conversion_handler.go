package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strings"
)

func (h *CommerceHandler) AxisConversion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	if r.Method == http.MethodGet {
		out, e := h.service.ReadAxisSource(r.Context(), a, r.PathValue("id"))
		if e != nil {
			commerceError(w, e)
			return
		}
		for i := range out.Attributes {
			out.Attributes[i].ID = backoffice.PublicReference("a", out.Attributes[i].ID)
			out.Attributes[i].Code = backoffice.PublicReference("a", out.Attributes[i].Code)
		}
		for i := range out.Category.Bindings {
			out.Category.Bindings[i].AttributeID = backoffice.PublicReference("a", out.Category.Bindings[i].AttributeID)
		}
		for i := range out.Variants {
			out.Variants[i].ID = backoffice.PublicReference("v", out.Variants[i].ID)
			out.Variants[i].Values = panelValues(out.Variants[i].Values)
		}
		writeJSON(w, 200, out)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/apply") {
		var in struct {
			PlanID      string `json:"plan_id"`
			Fingerprint string `json:"fingerprint"`
			Confirm     bool   `json:"confirm"`
		}
		if !commerceBody(w, r, &in) {
			return
		}
		out, e := h.service.ApplyAxisConversion(r.Context(), a, in.PlanID, in.Fingerprint, in.Confirm)
		commerceResult(w, r, out, e)
		return
	}
	var in struct {
		Input       backoffice.AxisConversionInput `json:"input"`
		Fingerprint string                         `json:"fingerprint"`
	}
	if !commerceBody(w, r, &in) {
		return
	}
	if strings.HasSuffix(r.URL.Path, "/prepare") {
		out, e := h.service.PrepareAxisConversion(r.Context(), a, in.Input, in.Fingerprint, r.Header.Get("Idempotency-Key"))
		commerceResult(w, r, out, e)
		return
	}
	out, e := h.service.PreviewAxisConversion(r.Context(), a, in.Input)
	for i := range out.Conflicts {
		c := &out.Conflicts[i]
		c.VariantID = backoffice.PublicReference("v", c.VariantID)
		c.OtherVariantID = backoffice.PublicReference("v", c.OtherVariantID)
		c.AttributeID = backoffice.PublicReference("a", c.AttributeID)
	}
	commerceResult(w, r, out, e)
}
