package httptransport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
)

// Only catalog reference fields and catalog URLs are projected. Amounts, user
// fields, identifiers in the ledger and backoffice payloads are never rewritten.
var catalogFieldKinds = map[string]string{
	"desktop_id": "m", "mobile_id": "m",
	"store_ids": "s", "allowed_categories": "c", "excluded_categories": "c", "category_ids": "c",
	"variant_id": "v", "selected_variant_id": "v", "offer_id": "o", "new_offer_id": "o", "selected_offer_id": "o",
	"product_id": "p", "category_id": "c", "parent_id": "c", "catalog_category_id": "c", "brand_id": "b", "store_id": "s", "attribute_id": "a",
}
var catalogCollectionKinds = map[string]string{
	"variants": "v", "offers": "o", "axes": "a", "attributes": "a", "filters": "a", "products": "p", "product": "p",
	"categories": "c", "category": "c", "navigation": "c", "breadcrumbs": "c", "children": "c", "brand": "b",
}

func importedReferenceKind(s string) string {
	for prefix, kind := range map[string]string{"gsp-variant-": "v", "gsp-offer-": "o", "gsp-media-": "m", "gsp-product-": "p", "gsp-cat-": "c", "gsp-brand-": "b", "gsp-store-": "s", "gsp-axis-": "a", "gsp-attr-": "a"} {
		if strings.HasPrefix(s, prefix) {
			return kind
		}
	}
	return ""
}
func projectCatalogURL(s string) string {
	// Scope to local catalog URLs; never alter user text, external hosts or receipts.
	if !strings.HasPrefix(s, "/") || strings.HasPrefix(s, "//") {
		return s
	}
	u, err := url.Parse(s)
	if err != nil {
		return s
	}
	segments := strings.Split(u.Path, "/")
	if len(segments) == 4 && ((segments[1] == "catalog" && segments[2] == "icons") || (segments[1] == "content" && segments[2] == "banner-media")) {
		segments[3] = backoffice.PublicReference("m", segments[3])
		u.Path = strings.Join(segments, "/")
	} else if len(segments) == 3 && (segments[1] == "product" || segments[1] == "category") && strings.HasPrefix(segments[2], "gsp-") {
		kind := "p"
		if segments[1] == "category" {
			kind = "c"
		}
		segments[2] = backoffice.PublicReference(kind, segments[2])
		u.Path = strings.Join(segments, "/")
	}
	q := u.Query()
	for key, values := range q {
		newKey := key
		if strings.HasPrefix(key, "attr.") {
			newKey = "attr." + backoffice.PublicReference("a", strings.TrimPrefix(key, "attr."))
			q.Del(key)
		}
		for i, value := range values {
			if kind := importedReferenceKind(value); kind != "" {
				values[i] = backoffice.PublicReference(kind, value)
			}
		}
		q[newKey] = values
	}
	u.RawQuery = q.Encode()
	return u.String()
}
func projectCatalogValue(v any, field, container string) any {
	switch value := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		// Root PDP and root navigation arrays have no collection envelope.
		inferred := field
		if _, ok := value["public_code"]; ok {
			inferred = "product"
		}
		if raw, ok := value["url"].(string); ok && strings.HasPrefix(raw, "/category/") {
			inferred = "category"
		}
		for k, child := range value {
			key := k
			if kind := importedReferenceKind(k); kind != "" {
				key = backoffice.PublicReference(kind, k)
			}
			out[key] = projectCatalogValue(child, k, inferred)
		}
		return out
	case []any:
		for i, child := range value {
			value[i] = projectCatalogValue(child, field, container)
		}
		return value
	case string:
		if value == "" {
			return value
		}
		if field == "code" && strings.HasPrefix(value, "gsp-") {
			return backoffice.PublicReference("a", value)
		}
		if kind := catalogFieldKinds[field]; kind != "" {
			return backoffice.PublicReference(kind, value)
		}
		if field == "id" {
			if kind := catalogCollectionKinds[container]; kind != "" {
				return backoffice.PublicReference(kind, value)
			}
			if kind := importedReferenceKind(value); kind != "" {
				return backoffice.PublicReference(kind, value)
			}
		}
		if field == "url" || field == "image_url" || field == "icon_url" || field == "link" || field == "gallery" || strings.HasSuffix(field, "_url") {
			return projectCatalogURL(value)
		}
	}
	return v
}
func publicCatalogJSON(data any) ([]byte, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var value any
	if err = dec.Decode(&value); err != nil {
		return nil, err
	}
	return json.Marshal(projectCatalogValue(value, "", ""))
}
func normalizeReferenceValue(value any, field string) (any, error) {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, child := range v {
			k := key
			// Shipping selections are keyed by store ID. Other arbitrary keys are preserved.
			if field == "shipping" || field == "shipments" || field == "shipping_selections" {
				var err error
				k, err = backoffice.ResolvePublicReference("s", key)
				if err != nil {
					return nil, err
				}
			}
			result, err := normalizeReferenceValue(child, key)
			if err != nil {
				return nil, err
			}
			if _, exists := out[k]; exists {
				return nil, backoffice.ErrInvalid
			}
			out[k] = result
		}
		return out, nil
	case []any:
		for i, child := range v {
			result, err := normalizeReferenceValue(child, field)
			if err != nil {
				return nil, err
			}
			v[i] = result
		}
		return v, nil
	case string:
		if kind := catalogFieldKinds[field]; kind != "" {
			return backoffice.ResolvePublicReference(kind, v)
		}
	}
	return value, nil
}
func publicCatalogHandler(kind string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if id := r.PathValue("id"); id != "" {
			resolved, err := backoffice.ResolvePublicReference(kind, id)
			if err != nil {
				w.Header().Set("Cache-Control", "no-store")
				writeError(w, 400, "invalid_catalog_reference")
				return
			}
			r.SetPathValue("id", resolved)
		}
		next(w, r)
	}
}
