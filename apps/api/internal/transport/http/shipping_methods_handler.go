package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strconv"
)

func (h *CommerceHandler) ShippingMethods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	actor, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	if r.Method == http.MethodGet {
		q := r.URL.Query()
		page, size := 0, 20
		var err error
		if q.Has("page") {
			page, err = strconv.Atoi(q.Get("page"))
			if err != nil {
				commerceError(w, backoffice.ErrInvalid)
				return
			}
		}
		if q.Has("size") {
			size, err = strconv.Atoi(q.Get("size"))
			if err != nil {
				commerceError(w, backoffice.ErrInvalid)
				return
			}
		}
		out, err := h.service.ReadShippingMethods(r.Context(), actor, backoffice.ShippingMethodFilter{Query: q.Get("query"), Active: q.Get("active"), CollectionMode: q.Get("collection_mode"), DeliveryMode: q.Get("delivery_mode"), CategoryID: q.Get("category_id"), CityID: q.Get("city_id"), Page: page, Size: size})
		commerceResult(w, r, out, err)
		return
	}
	var in struct {
		backoffice.ShippingMethodInput
		Version *int `json:"version"`
	}
	if !commerceBody(w, r, &in) {
		return
	}
	if in.Version == nil {
		commerceError(w, backoffice.ErrInvalid)
		return
	}
	in.ShippingMethodInput.Version = *in.Version
	if r.Method == http.MethodPut {
		if in.ID != "" && in.ID != r.PathValue("id") {
			commerceError(w, backoffice.ErrInvalid)
			return
		}
		in.ID = r.PathValue("id")
	}
	out, err := h.service.SaveShippingMethod(r.Context(), actor, in.ShippingMethodInput, r.Method == http.MethodPost, r.Header.Get("Idempotency-Key"))
	commerceResult(w, r, out, err)
}
