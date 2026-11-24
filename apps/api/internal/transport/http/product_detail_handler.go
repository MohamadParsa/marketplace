package httptransport

import (
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ProductDetailHandler struct {
	service *backoffice.Service
	root    string
	review  bool
}

func NewProductDetailHandler(s *backoffice.Service, root string, review bool) *ProductDetailHandler {
	return &ProductDetailHandler{s, root, review}
}
func productQuery(r *http.Request) (string, string, error) {
	if len(r.URL.RawQuery) > 1024 {
		return "", "", backoffice.ErrInvalid
	}
	q, e := url.ParseQuery(r.URL.RawQuery)
	if e != nil {
		return "", "", backoffice.ErrInvalid
	}
	for k, v := range q {
		if k == "group" && (len(v) != 1 || v[0] != "1") {
			return "", "", backoffice.ErrInvalid
		}
		if (k != "variant" && k != "offer" && k != "group") || len(v) != 1 || v[0] == "" || len(v[0]) > 256 {
			return "", "", backoffice.ErrInvalid
		}
	}
	v, err := backoffice.ResolvePublicReference("v", q.Get("variant"))
	if err != nil {
		return "", "", err
	}
	o, err := backoffice.ResolvePublicReference("o", q.Get("offer"))
	if err != nil {
		return "", "", err
	}
	return v, o, nil
}
func productError(w http.ResponseWriter, e error) {
	w.Header().Set("Cache-Control", "no-store")
	if errors.Is(e, backoffice.ErrInvalid) {
		writeError(w, 400, "invalid_product_selection")
	} else if errors.Is(e, backoffice.ErrNotFound) {
		writeError(w, 404, "product_not_found")
	} else if errors.Is(e, backoffice.ErrConflict) {
		writeError(w, 409, "selection_unavailable")
	} else {
		writeError(w, 503, "product_unavailable")
	}
}
func (h *ProductDetailHandler) read(r *http.Request) (backoffice.ProductDetail, error) {
	v, o, e := productQuery(r)
	if e != nil {
		return backoffice.ProductDetail{}, e
	}
	return h.service.PublicProduct(r.Context(), r.PathValue("id"), v, o, r.URL.Query().Get("group") == "1")
}
func (h *ProductDetailHandler) Read(w http.ResponseWriter, r *http.Request) {
	d, e := h.read(r)
	if e != nil {
		productError(w, e)
		return
	}
	writeCatalogJSON(w, r, struct {
		backoffice.ProductDetail
		Robots string `json:"robots"`
	}{d, h.robots(d)})
}
func (h *ProductDetailHandler) Selection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var in backoffice.ProductSelectionInput
	if dec.Decode(&in) != nil {
		productError(w, backoffice.ErrInvalid)
		return
	}
	if dec.Decode(new(any)) != io.EOF {
		productError(w, backoffice.ErrInvalid)
		return
	}
	var refErr error
	in.VariantID, refErr = backoffice.ResolvePublicReference("v", in.VariantID)
	if refErr != nil {
		productError(w, refErr)
		return
	}
	in.OfferID, refErr = backoffice.ResolvePublicReference("o", in.OfferID)
	if refErr != nil {
		productError(w, refErr)
		return
	}
	out, e := h.service.ValidateSelection(r.Context(), r.PathValue("id"), in)
	if e != nil {
		productError(w, e)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	b, err := publicCatalogJSON(out)
	if err != nil {
		productError(w, err)
		return
	}
	_, _ = w.Write(b)
}
func (h *ProductDetailHandler) robots(d backoffice.ProductDetail) string {
	index, follow := "index", "follow"
	if h.review || d.Content.NoIndex {
		index = "noindex"
	}
	if h.review || d.Content.NoFollow {
		follow = "nofollow"
	}
	return index + ", " + follow
}

var robotsMeta = regexp.MustCompile(`(?i)<meta\s+name=["']robots["'][^>]*>`)

func (h *ProductDetailHandler) Page(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	d, e := h.read(r)
	if e == nil {
		q := r.URL.Query()
		if q.Has("variant") {
			q.Set("variant", backoffice.PublicReference("v", d.SelectedVariantID))
		}
		if q.Has("offer") {
			q.Set("offer", backoffice.PublicReference("o", d.SelectedOfferID))
		}
		target := d.URL
		if len(q) > 0 {
			target += "?" + q.Encode()
		}
		current := r.URL.Path
		if r.URL.RawQuery != "" {
			current += "?" + r.URL.Query().Encode()
		}
		if current != target {
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}
	}
	status := 200
	if e != nil {
		status = 503
		if errors.Is(e, backoffice.ErrNotFound) {
			status = 404
		}
		if errors.Is(e, backoffice.ErrInvalid) {
			status = 400
		}
		d.Content.SEO.MetaTitle = "محصول در دسترس نیست | Cherriva"
		d.Content.NoIndex = true
		d.Content.NoFollow = true
	}
	b, err := os.ReadFile(filepath.Join(h.root, "index.html"))
	if err != nil {
		http.Error(w, "Storefront build unavailable", 503)
		return
	}
	body := robotsMeta.ReplaceAllString(string(b), "")
	a, z := "<!--cherriva-page-meta:start-->", "<!--cherriva-page-meta:end-->"
	start, end := strings.Index(body, a), strings.Index(body, z)
	if start < 0 || end < start {
		http.Error(w, "Storefront metadata template unavailable", 503)
		return
	}
	seo := d.Content.SEO
	meta := "<title>" + html.EscapeString(seo.MetaTitle) + "</title>\n<meta name=\"description\" content=\"" + html.EscapeString(seo.MetaDescription) + "\">\n<meta name=\"keywords\" content=\"" + html.EscapeString(seo.MetaKeywords) + "\">\n<meta name=\"robots\" content=\"" + h.robots(d) + "\">"
	if status == 200 {
		meta += "\n<link rel=\"canonical\" href=\"" + html.EscapeString(d.URL) + "\">"
	}
	w.WriteHeader(status)
	io.WriteString(w, body[:start]+meta+body[end+len(z):])
}
