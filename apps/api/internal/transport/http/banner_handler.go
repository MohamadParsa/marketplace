package httptransport

import (
	"errors"
	"html/template"
	"io"
	"net/http"
	"slices"
	"strconv"

	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type BannerHandler struct {
	service *backoffice.Service
	auth    *auth.Service
}

func NewBannerHandler(service *backoffice.Service, authentication *auth.Service) *BannerHandler {
	return &BannerHandler{service, authentication}
}
func (h *BannerHandler) actor(w http.ResponseWriter, r *http.Request) (identity.User, bool) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	session, err := h.auth.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), identity.SurfaceBackoffice)
	if err != nil {
		writeError(w, 401, "unauthorized")
		return identity.User{}, false
	}
	return session.User, true
}
func (h *BannerHandler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	result, err := h.service.ReadBanners(r.Context(), actor)
	if err != nil {
		iconError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
func (h *BannerHandler) Media(w http.ResponseWriter, r *http.Request) {
	var actor *identity.User
	if r.PathValue("access") == "private" {
		current, ok := h.actor(w, r)
		if !ok {
			return
		}
		actor = &current
	}
	asset, err := h.service.ReadBannerMedia(r.Context(), actor, r.PathValue("id"))
	if err != nil {
		iconError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Length", strconv.Itoa(len(asset.Data)))
	_, _ = w.Write(asset.Data)
}
func (h *BannerHandler) Upload(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}
	if actor.Role != identity.RoleSuperAdmin && actor.Role != identity.RoleOperations && !slices.Contains(actor.Permissions, "banner.media.set") {
		iconError(w, backoffice.ErrForbidden)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, backoffice.MaxBannerBytes+16384)
	if err := r.ParseMultipartForm(backoffice.MaxBannerBytes + 16384); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			iconError(w, backoffice.ErrIconTooLarge)
		} else {
			iconError(w, backoffice.ErrInvalid)
		}
		return
	}
	defer r.MultipartForm.RemoveAll()
	for key, values := range r.MultipartForm.Value {
		if (key != "version" && key != "idempotency_key") || len(values) != 1 {
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
	data, err := io.ReadAll(io.LimitReader(file, backoffice.MaxBannerBytes+1))
	if err != nil {
		iconError(w, backoffice.ErrInvalid)
		return
	}
	result, err := h.service.SetBannerMedia(r.Context(), actor, r.PathValue("id"), r.PathValue("slot"), version, r.FormValue("idempotency_key"), data)
	if err != nil {
		iconError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
func (h *BannerHandler) PublicList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	rows, err := h.service.PublicBanners(r.Context())
	if err != nil {
		iconError(w, err)
		return
	}
	writeJSON(w, 200, rows)
}

var bannerTemplate = template.Must(template.New("banner").Parse(`<!doctype html><html lang="fa" dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="robots" content="noindex,nofollow"><title>{{.Name}}</title><style>body{margin:0;background:#f5f7f8}main{max-width:1440px;margin:auto}img{display:block;width:100%;height:auto}a{display:block}</style></head><body><main data-revision="{{.Revision}}">{{if .Config.Link}}<a href="{{.Config.Link}}">{{end}}<picture><source media="(max-width: 767px)" srcset="/content/banner-media/{{.Config.MobileID}}"><img src="/content/banner-media/{{.Config.DesktopID}}" alt="{{.Config.Alt}}"></picture>{{if .Config.Link}}</a>{{end}}</main></body></html>`))

func (h *BannerHandler) View(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	rows, err := h.service.PublicBanners(r.Context())
	if err != nil {
		iconError(w, err)
		return
	}
	for _, row := range rows {
		if row.ID == r.PathValue("id") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'")
			if row.Config.CatalogCategoryID != "" {
				home, e := h.service.PublicHome(r.Context())
				if e != nil {
					iconError(w, e)
					return
				}
				for _, rail := range home.Rails {
					if rail.CategoryID == row.Config.CatalogCategoryID && len(rail.Products) > 0 {
						_ = catalogBannerTemplate.Execute(w, struct{ Title, Link, Image string }{row.Config.Alt, rail.URL, rail.Products[0].ImageURL})
						return
					}
				}
				iconError(w, backoffice.ErrNotFound)
				return
			}
			_ = bannerTemplate.Execute(w, row)
			return
		}
	}
	iconError(w, backoffice.ErrNotFound)
}

var catalogBannerTemplate = template.Must(template.New("catalog-promo").Parse(`<!doctype html><html lang="fa" dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="robots" content="noindex,nofollow"><title>{{.Title}}</title><style>body{margin:0;background:#f7eaf0;font-family:sans-serif;color:#402934}a{display:flex;align-items:center;gap:24px;max-width:900px;margin:auto;padding:32px;color:inherit;text-decoration:none}img{width:40%;max-height:260px;object-fit:contain;mix-blend-mode:multiply}h1{line-height:1.8}span{color:#b51e4f}</style></head><body><a href="{{.Link}}"><div><h1>{{.Title}}</h1><span>مشاهده مدل‌های موجود</span></div><img src="{{.Image}}" alt="{{.Title}}"></a></body></html>`))
