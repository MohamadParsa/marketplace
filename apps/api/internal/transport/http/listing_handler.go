package httptransport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
)

type ListingHandler struct {
	service *backoffice.Service
	root    string
}

func NewListingHandler(s *backoffice.Service, root string) *ListingHandler {
	return &ListingHandler{s, root}
}
func listingQuery(r *http.Request) (backoffice.ListingQuery, error) {
	if len(r.URL.RawQuery) > 4096 {
		return backoffice.ListingQuery{}, backoffice.ErrListingQuery
	}
	v, e := url.ParseQuery(r.URL.RawQuery)
	if e != nil {
		return backoffice.ListingQuery{}, backoffice.ErrListingQuery
	}
	if id := r.PathValue("id"); id != "" {
		if len(v["category"]) > 1 || (len(v["category"]) > 0 && v.Get("category") != id) {
			return backoffice.ListingQuery{}, backoffice.ErrListingQuery
		}
		v.Set("category", id)
	}
	if r.URL.Path == "/offers" {
		v.Set("special", "1")
	}
	return backoffice.ParseListingQuery(v)
}
func listingError(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, backoffice.ErrListingQuery):
		writeError(w, 400, "invalid_listing_query")
	case errors.Is(e, backoffice.ErrNotFound):
		writeError(w, 404, "category_not_found")
	default:
		writeError(w, 503, "catalog_unavailable")
	}
}
func writeCatalogJSON(w http.ResponseWriter, r *http.Request, data any) {
	b, e := json.Marshal(data)
	if e != nil {
		writeError(w, 500, "internal_error")
		return
	}
	sum := sha256.Sum256(b)
	tag := "\"" + hex.EncodeToString(sum[:]) + "\""
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, no-cache")
	w.Header().Set("ETag", tag)
	if r.Header.Get("If-None-Match") == tag {
		w.WriteHeader(304)
		return
	}
	w.WriteHeader(200)
	_, _ = w.Write(b)
}
func (h *ListingHandler) Read(w http.ResponseWriter, r *http.Request) {
	q, e := listingQuery(r)
	if e != nil {
		listingError(w, e)
		return
	}
	data, e := h.service.PublicListing(r.Context(), q)
	if e != nil {
		listingError(w, e)
		return
	}
	writeCatalogJSON(w, r, data)
}
func (h *ListingHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.RawQuery) > 1024 {
		listingError(w, backoffice.ErrListingQuery)
		return
	}
	v, e := url.ParseQuery(r.URL.RawQuery)
	if e != nil || len(v) != 1 || len(v["q"]) != 1 {
		listingError(w, backoffice.ErrListingQuery)
		return
	}
	data, e := h.service.Suggestions(r.Context(), v.Get("q"))
	if e != nil {
		listingError(w, e)
		return
	}
	writeCatalogJSON(w, r, data)
}

// Page serves the same React build with metadata in the initial HTML response.
// It is enabled only on an explicitly configured storefront web surface.
func (h *ListingHandler) Page(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	q, e := listingQuery(r)
	status := 200
	var data backoffice.ListingResult
	if e == nil {
		data, e = h.service.PublicListing(r.Context(), q)
	}
	if e != nil {
		status = 503
		if errors.Is(e, backoffice.ErrListingQuery) {
			status = 400
		}
		if errors.Is(e, backoffice.ErrNotFound) {
			status = 404
		}
		data.SEO.MetaTitle = "صفحه در دسترس نیست | Cherriva"
	}
	b, readErr := os.ReadFile(filepath.Join(h.root, "index.html"))
	if readErr != nil {
		http.Error(w, "Storefront build unavailable", 503)
		return
	}
	marker := "<!--cherriva-page-meta:start-->"
	endMarker := "<!--cherriva-page-meta:end-->"
	if !strings.Contains(string(b), marker) || !strings.Contains(string(b), endMarker) {
		http.Error(w, "Storefront metadata template unavailable", 503)
		return
	}
	meta := "<title>" + html.EscapeString(data.SEO.MetaTitle) + "</title>\n<meta name=\"description\" content=\"" + html.EscapeString(data.SEO.MetaDescription) + "\">\n<meta name=\"keywords\" content=\"" + html.EscapeString(data.SEO.MetaKeywords) + "\">"
	start := strings.Index(string(b), marker)
	end := strings.Index(string(b), endMarker)
	if end < start {
		http.Error(w, "Invalid storefront template", 503)
		return
	}
	body := string(b[:start]) + meta + string(b[end+len(endMarker):])
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}
