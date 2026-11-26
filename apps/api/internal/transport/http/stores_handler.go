package httptransport

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func (h *CommerceHandler) Stores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	actor, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	if strings.HasSuffix(r.URL.Path, "/image") {
		id := r.PathValue("id")
		if r.Method == http.MethodGet {
			data, e := h.service.ReadStoreImage(r.Context(), actor, id)
			if e != nil {
				commerceError(w, e)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Write(data)
			return
		}
		if r.Method == http.MethodDelete {
			var in struct {
				Version *int `json:"version"`
			}
			if !commerceBody(w, r, &in) {
				return
			}
			if in.Version == nil {
				commerceError(w, backoffice.ErrInvalid)
				return
			}
			out, e := h.service.ChangeStoreImage(r.Context(), actor, id, *in.Version, r.Header.Get("Idempotency-Key"), nil, true)
			commerceResult(w, r, out, e)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, backoffice.MaxIconBytes+16384)
		if e := r.ParseMultipartForm(backoffice.MaxIconBytes + 16384); e != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(e, &tooLarge) {
				iconError(w, backoffice.ErrIconTooLarge)
				return
			}
			iconError(w, backoffice.ErrInvalidIcon)
			return
		}
		defer r.MultipartForm.RemoveAll()
		if len(r.MultipartForm.File) != 1 || len(r.MultipartForm.File["file"]) != 1 || len(r.MultipartForm.Value) != 1 || len(r.MultipartForm.Value["version"]) != 1 {
			iconError(w, backoffice.ErrInvalid)
			return
		}
		version, e := strconv.Atoi(r.FormValue("version"))
		if e != nil {
			iconError(w, backoffice.ErrInvalid)
			return
		}
		f, e := r.MultipartForm.File["file"][0].Open()
		if e != nil {
			iconError(w, backoffice.ErrInvalidIcon)
			return
		}
		defer f.Close()
		data, e := io.ReadAll(io.LimitReader(f, backoffice.MaxIconBytes+1))
		if e != nil {
			iconError(w, backoffice.ErrInvalidIcon)
			return
		}
		out, e := h.service.ChangeStoreImage(r.Context(), actor, id, version, r.Header.Get("Idempotency-Key"), data, false)
		if e != nil {
			iconError(w, e)
			return
		}
		writeJSON(w, 200, out)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/save") {
		var in backoffice.StoreInput
		if !commerceBody(w, r, &in) {
			return
		}
		out, e := h.service.SaveStore(r.Context(), actor, in, r.Header.Get("Idempotency-Key"))
		commerceResult(w, r, out, e)
		return
	}
	var in struct {
		Filter backoffice.StoreFilter `json:"filter"`
		IDs    []string               `json:"ids"`
	}
	if !commerceBody(w, r, &in) {
		return
	}
	out, e := h.service.ListStores(r.Context(), actor, in.Filter, strings.HasSuffix(r.URL.Path, "/export"), in.IDs)
	commerceResult(w, r, out, e)
}
