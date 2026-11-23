package backoffice

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

var ErrListingQuery = errors.New("invalid_listing_query")

type ListingQuery struct {
	CategoryID string
	Q          string
	Sort       string
	InStock    bool
	Special    bool
	MinPrice   *int64
	MaxPrice   *int64
	Page       int
	PageSize   int
	Attributes map[string][]string
}
type CatalogNav struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	ParentID string `json:"parent_id,omitempty"`
	IconURL  string `json:"icon_url,omitempty"`
}
type CatalogCategory struct {
	CatalogNav
	Description string `json:"description"`
	EnglishName string `json:"english_name,omitempty"`
}
type ListingFilter struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
	Unit    string   `json:"unit,omitempty"`
	Options []string `json:"options"`
}
type ListingCard struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	CategoryName    string                `json:"category_name"`
	URL             string                `json:"url"`
	ImageURL        string                `json:"image_url,omitempty"`
	Available       bool                  `json:"available"`
	Price           *int64                `json:"price"`
	RegularPrice    int64                 `json:"regular_price,omitempty"`
	DiscountPercent int                   `json:"discount_percent,omitempty"`
	OfferID         string                `json:"offer_id,omitempty"`
	VariantID       string                `json:"variant_id,omitempty"`
	Features        []catalog.CardFeature `json:"features"`
}
type ListingPriceRange struct {
	Min *int64 `json:"min"`
	Max *int64 `json:"max"`
}
type ListingResult struct {
	PriceRange  ListingPriceRange `json:"price_range"`
	Category    *CatalogCategory  `json:"category"`
	Breadcrumbs []CatalogNav      `json:"breadcrumbs"`
	Children    []CatalogNav      `json:"children"`
	Navigation  []CatalogNav      `json:"navigation"`
	Products    []ListingCard     `json:"products"`
	Filters     []ListingFilter   `json:"filters"`
	SEO         catalog.SEO       `json:"seo"`
	Title       string            `json:"title"`
	Query       string            `json:"query"`
	Total       int               `json:"total"`
	Page        int               `json:"page"`
	PageSize    int               `json:"page_size"`
	TotalPages  int               `json:"total_pages"`
	Sort        string            `json:"sort"`
	SortOptions []string          `json:"sort_options"`
}

func ParseListingQuery(v url.Values) (ListingQuery, error) {
	q := ListingQuery{Sort: "default", Page: 1, PageSize: 20, Attributes: map[string][]string{}}
	for key, values := range v {
		if strings.HasPrefix(key, "attr.") {
			id := strings.TrimPrefix(key, "attr.")
			if id == "" || len(id) > 128 || len(values) > 20 {
				return q, ErrListingQuery
			}
			seen := map[string]bool{}
			for _, value := range values {
				if value == "" || utf8.RuneCountInString(value) > 160 || seen[value] {
					return q, ErrListingQuery
				}
				seen[value] = true
			}
			q.Attributes[id] = values
			continue
		}
		if len(values) != 1 {
			return q, ErrListingQuery
		}
		value := values[0]
		switch key {
		case "category":
			q.CategoryID = value
			if len(value) > 128 {
				return q, ErrListingQuery
			}
		case "q":
			q.Q = strings.TrimSpace(value)
			if utf8.RuneCountInString(q.Q) > 80 || strings.ContainsAny(q.Q, "\x00\r\n") {
				return q, ErrListingQuery
			}
		case "sort":
			if value != "default" && value != "newest" && value != "price-asc" && value != "price-desc" && value != "best-selling" {
				return q, ErrListingQuery
			}
			q.Sort = value
		case "in_stock":
			if value != "1" && value != "0" {
				return q, ErrListingQuery
			}
			q.InStock = value == "1"
		case "special":
			if value != "1" && value != "0" {
				return q, ErrListingQuery
			}
			q.Special = value == "1"
		case "min_price", "max_price":
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil || n < 0 || n > MaxOfferPrice {
				return q, ErrListingQuery
			}
			if key == "min_price" {
				q.MinPrice = &n
			} else {
				q.MaxPrice = &n
			}
		case "page", "page_size":
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return q, ErrListingQuery
			}
			if key == "page" {
				if n > 10000 {
					return q, ErrListingQuery
				}
				q.Page = n
			} else {
				if n > 40 {
					return q, ErrListingQuery
				}
				q.PageSize = n
			}
		default:
			return q, ErrListingQuery
		}
	}
	if len(q.Attributes) > 20 || (len(q.Attributes) > 0 && q.CategoryID == "") || (q.MinPrice != nil && q.MaxPrice != nil && *q.MinPrice > *q.MaxPrice) {
		return q, ErrListingQuery
	}
	return q, nil
}
func catalogNav(w Workspace, c catalog.Category) CatalogNav {
	n := CatalogNav{ID: c.ID, Name: c.Name, URL: CategoryURL(c), ParentID: c.ParentID}
	if imageAttached(w, c.IconID) {
		n.IconURL = "/catalog/icons/" + url.PathEscape(c.IconID)
	}
	return n
}
func publicNavigation(w Workspace) []CatalogNav {
	result := []CatalogNav{}
	for _, c := range w.Categories {
		if publicCategory(w, c.ID) != nil {
			result = append(result, catalogNav(w, c))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == result[j].Name {
			return result[i].ID < result[j].ID
		}
		return result[i].Name < result[j].Name
	})
	return result
}
func categoryBindings(w Workspace, id string) []catalog.AttributeBinding {
	for _, c := range w.Categories {
		if c.ID == id {
			bs := append([]catalog.AttributeBinding{}, c.Bindings...)
			sort.Slice(bs, func(i, j int) bool {
				if bs[i].Position == bs[j].Position {
					return bs[i].AttributeID < bs[j].AttributeID
				}
				return bs[i].Position < bs[j].Position
			})
			return bs
		}
	}
	return nil
}
func definitions(w Workspace) map[string]catalog.AttributeDefinition {
	if w.publicIndex != nil {
		return w.publicIndex.definitions
	}
	m := map[string]catalog.AttributeDefinition{}
	for _, a := range w.Attributes {
		m[a.ID] = a
	}
	return m
}
func variantValues(w Workspace, p catalog.Product, v catalog.Variant) map[string]string {
	allowed := map[string]bool{}
	for _, b := range categoryBindings(w, p.CategoryID) {
		allowed[b.AttributeID] = true
	}
	values := map[string]string{}
	for _, a := range p.Attributes {
		if allowed[a.AttributeID] {
			values[a.AttributeID] = a.Value
		}
	}
	for _, a := range v.Values {
		if allowed[a.AttributeID] {
			values[a.AttributeID] = a.Value
		}
	}
	return values
}
func matchesAttributes(values map[string]string, attrs map[string][]string) bool {
	for id, choices := range attrs {
		ok := false
		for _, v := range choices {
			ok = ok || NormalizeSearch(values[id]) == NormalizeSearch(v)
		}
		if !ok {
			return false
		}
	}
	return true
}
func productVariants(w Workspace, p catalog.Product) []catalog.Variant {
	vs := []catalog.Variant{}
	variants := w.Variants
	if w.publicIndex != nil {
		variants = w.publicIndex.variants[p.ID]
	}
	for _, v := range variants {
		if v.ProductID == p.ID {
			vs = append(vs, v)
		}
	}
	sort.Slice(vs, func(i, j int) bool { return vs[i].ID < vs[j].ID })
	return vs
}
func productCreatedAt(w Workspace, p catalog.Product) time.Time {
	if !p.CreatedAt.IsZero() {
		return p.CreatedAt
	}
	var created time.Time
	for _, a := range w.Audit {
		if a.Action == "catalog.product.create" && a.Target == p.ID && (created.IsZero() || a.At.Before(created)) {
			created = a.At
		}
	}
	return created
}
func listingFeatures(w Workspace, p catalog.Product, v catalog.Variant) []catalog.CardFeature {
	result := []catalog.CardFeature{}
	values := variantValues(w, p, v)
	defs := definitions(w)
	for _, b := range categoryBindings(w, p.CategoryID) {
		if len(result) >= w.DisplaySettings.CardAttributeLimit {
			break
		}
		a, ok := defs[b.AttributeID]
		if b.Card && ok && strings.TrimSpace(values[a.ID]) != "" {
			result = append(result, catalog.CardFeature{AttributeID: a.ID, Name: a.Name, Value: values[a.ID], Unit: a.Unit})
		}
	}
	return result
}
func listingCard(w Workspace, p catalog.Product, q ListingQuery) (ListingCard, bool) {
	c := publicCategory(w, p.CategoryID)
	if c == nil || p.Status != "published" || !validBrand(w, p.BrandID) {
		return ListingCard{}, false
	}
	card := ListingCard{ID: p.ID, Name: p.Name, CategoryName: c.Name, URL: ProductURL(p), Features: []catalog.CardFeature{}}
	if imageAttached(w, p.ImageID) {
		card.ImageURL = "/catalog/icons/" + url.PathEscape(p.ImageID)
	}
	variants := productVariants(w, p)
	matching := map[string]catalog.Variant{}
	var display catalog.Variant
	for _, v := range variants {
		if matchesAttributes(variantValues(w, p, v), q.Attributes) {
			matching[v.ID] = v
			if display.ID == "" {
				display = v
			}
		}
	}
	if len(q.Attributes) > 0 && len(matching) == 0 {
		return card, false
	}
	for _, o := range publicOffers(w, p) {
		v, ok := matching[o.VariantID]
		if !ok {
			continue
		}
		if q.MinPrice != nil && o.Price < *q.MinPrice {
			continue
		}
		if q.MaxPrice != nil && o.Price > *q.MaxPrice {
			continue
		}
		card.Available = true
		price := o.Price
		card.Price = &price
		card.OfferID = o.ID
		card.VariantID = o.VariantID
		display = v
		if o.RegularPrice > o.Price {
			card.RegularPrice = o.RegularPrice
			card.DiscountPercent = int((o.RegularPrice - o.Price) * 100 / o.RegularPrice)
		}
		break
	}
	// A price filter must not turn an otherwise available product into a fake unavailable one.
	if !card.Available && (q.InStock || q.Special || q.MinPrice != nil || q.MaxPrice != nil) {
		return card, false
	}
	if q.Special && card.RegularPrice == 0 {
		return card, false
	}
	card.Features = listingFeatures(w, p, display)
	return card, true
}
func BuildListing(w Workspace, q ListingQuery) (ListingResult, error) {
	return BuildListingWithSales(w, q, BuildCatalogSales(w).Counts(time.Now()))
}
func BuildListingWithSales(w Workspace, q ListingQuery, sales map[string]int64) (ListingResult, error) {
	w = indexedCatalog(w)
	if q.Page < 1 || q.Page > 10000 || q.PageSize < 1 || q.PageSize > 40 {
		return ListingResult{}, ErrListingQuery
	}
	result := ListingResult{Breadcrumbs: []CatalogNav{}, Children: []CatalogNav{}, Navigation: publicNavigation(w), Products: []ListingCard{}, Filters: []ListingFilter{}, Page: q.Page, PageSize: q.PageSize, Sort: q.Sort, SortOptions: []string{"default", "newest", "price-asc", "price-desc", "best-selling"}, Title: "جست‌وجوی کالا", Query: q.Q}
	var scope map[string]bool
	if q.CategoryID != "" {
		c := publicCategory(w, q.CategoryID)
		if c == nil {
			return result, ErrNotFound
		}
		scope = map[string]bool{c.ID: true}
		result.Category = &CatalogCategory{CatalogNav: catalogNav(w, *c), Description: c.Description, EnglishName: c.EnglishName}
		result.Title = c.Name
		result.SEO = c.SEO
		seen := map[string]bool{}
		current := c
		for current != nil && !seen[current.ID] {
			seen[current.ID] = true
			result.Breadcrumbs = append([]CatalogNav{catalogNav(w, *current)}, result.Breadcrumbs...)
			current = publicCategory(w, current.ParentID)
		}
		for _, n := range result.Navigation {
			if n.ParentID == c.ID {
				result.Children = append(result.Children, n)
			}
		}
	}
	defs := definitions(w)
	allowed := map[string]bool{}
	if result.Category != nil {
		for _, b := range categoryBindings(w, q.CategoryID) {
			if b.Filter {
				if a, ok := defs[b.AttributeID]; ok {
					allowed[a.ID] = true
					result.Filters = append(result.Filters, ListingFilter{ID: a.ID, Name: a.Name, Kind: a.Kind, Unit: a.Unit, Options: []string{}})
				}
			}
		}
	}
	for id, values := range q.Attributes {
		a, ok := defs[id]
		if !ok || !allowed[id] {
			return result, ErrListingQuery
		}
		for _, v := range values {
			if a.Kind == "enum" {
				valid := false
				for _, o := range a.Options {
					valid = valid || v == o
				}
				if !valid {
					return result, ErrListingQuery
				}
			}
			if a.Kind == "number" {
				n, e := strconv.ParseFloat(v, 64)
				if e != nil || n != n || n > 1e15 || n < -1e15 {
					return result, ErrListingQuery
				}
			}
		}
	}
	candidates := []catalog.Product{}
	tokens := strings.Fields(NormalizeSearch(q.Q))
	for _, p := range w.Products {
		if p.Status != "published" || publicCategory(w, p.CategoryID) == nil || !validBrand(w, p.BrandID) || (scope != nil && !scope[p.CategoryID]) {
			continue
		}
		text := NormalizeSearch(p.Name + " " + p.EnglishName)
		match := true
		for _, t := range tokens {
			match = match && strings.Contains(text, t)
		}
		if !match {
			continue
		}
		candidates = append(candidates, p)
	}
	for i := range result.Filters {
		f := &result.Filters[i]
		options := map[string]bool{}
		for _, p := range candidates {
			for _, v := range productVariants(w, p) {
				if value := variantValues(w, p, v)[f.ID]; value != "" {
					options[value] = true
				}
			}
		}
		for v := range options {
			f.Options = append(f.Options, v)
		}
		sort.Slice(f.Options, func(i, j int) bool {
			if f.Kind == "number" {
				a, e := strconv.ParseFloat(f.Options[i], 64)
				b, e2 := strconv.ParseFloat(f.Options[j], 64)
				if e == nil && e2 == nil && a != b {
					return a < b
				}
			}
			return f.Options[i] < f.Options[j]
		})
	}
	cards := []ListingCard{}
	dates := map[string]time.Time{}
	for _, p := range candidates {
		extendListingPriceRange(&result.PriceRange, w, p, q)
		if card, ok := listingCard(w, p, q); ok {
			cards = append(cards, card)
			dates[p.ID] = productCreatedAt(w, p)
		}
	}
	sort.Slice(cards, func(i, j int) bool {
		a, b := cards[i], cards[j]
		if a.Available != b.Available {
			return a.Available
		}
		if a.Available && b.Available && *a.Price != *b.Price {
			if q.Sort == "price-asc" {
				return *a.Price < *b.Price
			}
			if q.Sort == "price-desc" {
				return *a.Price > *b.Price
			}
		}
		if q.Sort == "best-selling" && sales[a.ID] != sales[b.ID] {
			return sales[a.ID] > sales[b.ID]
		}
		if !dates[a.ID].Equal(dates[b.ID]) {
			return dates[a.ID].After(dates[b.ID])
		}
		return a.ID < b.ID
	})
	if q.CategoryID == "" {
		if q.Special {
			result.Title = "تخفیف‌های ویژه"
		} else if q.Q != "" {
			result.Title = "نتایج جست‌وجوی «" + q.Q + "»"
		}
	}
	if strings.TrimSpace(result.SEO.MetaTitle) == "" {
		result.SEO.MetaTitle = result.Title + " | Cherriva"
	}
	if strings.TrimSpace(result.SEO.MetaDescription) == "" && result.Category != nil {
		text := []rune(strings.Join(strings.Fields(result.Category.Description), " "))
		if len(text) > 400 {
			text = text[:400]
		}
		result.SEO.MetaDescription = string(text)
	}
	result.Total = len(cards)
	result.TotalPages = (len(cards) + q.PageSize - 1) / q.PageSize
	start := (q.Page - 1) * q.PageSize
	if start < len(cards) {
		end := start + q.PageSize
		if end > len(cards) {
			end = len(cards)
		}
		result.Products = append(result.Products, cards[start:end]...)
	}
	return result, nil
}
func (s *Service) PublicListing(ctx context.Context, q ListingQuery) (ListingResult, error) {
	if q.CategoryID != "" {
		id, e := s.resolveCatalogToken(ctx, "category", q.CategoryID)
		if e != nil {
			return ListingResult{}, e
		}
		q.CategoryID = id
	}

	attrs := map[string][]string{}
	for k, v := range q.Attributes {
		attrs[k] = append([]string{}, v...)
		sort.Strings(attrs[k])
	}
	q.Attributes = attrs
	if q.Sort == "" {
		q.Sort = "default"
	}
	var salesRevision int64
	now := s.clock.Now()
	if q.Sort == "best-selling" {
		if r, ok := s.repository.(CatalogSalesRepository); ok {
			var e error
			salesRevision, e = r.CatalogSalesRevision(ctx)
			if e != nil {
				return ListingResult{}, e
			}
		}
	}
	var cacheArgs any = q
	if q.Sort == "best-selling" {
		cacheArgs = struct {
			Query         ListingQuery
			SalesRevision int64
			Minute        int64
		}{q, salesRevision, now.Unix() / 60}
	}
	return cachedPublic(ctx, s, "listing", "", cacheArgs, time.Minute, func(ctx context.Context) (ListingResult, error) {
		w, e := s.publicSnapshot(ctx, "")
		if e != nil {
			return ListingResult{}, e
		}
		if q.Sort == "best-selling" {
			var projection CatalogSales
			if r, ok := s.repository.(CatalogSalesRepository); ok {
				projection, e = r.CatalogSales(ctx)
			} else {
				var private Workspace
				private, e = s.repository.Snapshot(ctx)
				if e == nil {
					projection = BuildCatalogSales(private)
				}
			}
			if e != nil {
				return ListingResult{}, e
			}
			return BuildListingWithSales(w, q, projection.Counts(now))
		}
		return BuildListingWithSales(w, q, nil)
	})
}

type Suggestion struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Kind string `json:"kind"`
}

func (s *Service) Suggestions(ctx context.Context, query string) ([]Suggestion, error) {
	if utf8.RuneCountInString(query) > 80 {
		return nil, ErrListingQuery
	}
	query = NormalizeSearch(query)
	return cachedPublic(ctx, s, "suggestions", "", query, 30*time.Second, func(ctx context.Context) ([]Suggestion, error) { return s.buildSuggestions(ctx, query) })
}
func (s *Service) buildSuggestions(ctx context.Context, query string) ([]Suggestion, error) {
	result := []Suggestion{}
	if strings.TrimSpace(query) == "" {
		return result, nil
	}
	if utf8.RuneCountInString(query) > 80 {
		return nil, ErrListingQuery
	}
	w, e := s.publicSnapshot(ctx, "")
	if e != nil {
		return nil, e
	}
	q, e := ParseListingQuery(url.Values{"q": {query}, "page_size": {"6"}})
	if e != nil {
		return nil, e
	}
	data, e := BuildListing(w, q)
	if e != nil {
		return nil, e
	}
	for _, p := range data.Products {
		result = append(result, Suggestion{p.ID, p.Name, p.URL, "product"})
	}
	tokens := strings.Fields(NormalizeSearch(query))
	n := 0
	for _, c := range data.Navigation {
		match := true
		for _, t := range tokens {
			match = match && strings.Contains(NormalizeSearch(c.Name), t)
		}
		if match && n < 4 {
			result = append(result, Suggestion{c.ID, c.Name, c.URL, "category"})
			n++
		}
	}
	return result, nil
}

func (s *Service) PublicNavigation(ctx context.Context) ([]CatalogNav, error) {
	return cachedPublic(ctx, s, "navigation", "__shared__", nil, 5*time.Minute, func(ctx context.Context) ([]CatalogNav, error) {
		w, e := s.publicSnapshot(ctx, "__shared__")
		if e != nil {
			return nil, e
		}
		return publicNavigation(indexedCatalog(w)), nil
	})
}

// Price domain is computed from all buyable matching offers, before the price filter and pagination.
// For special offers, the deterministic representative of each price must itself be discounted,
// matching listingCard's behavior for a range narrowed to that exact price.
func extendListingPriceRange(r *ListingPriceRange, w Workspace, p catalog.Product, q ListingQuery) {
	matching := map[string]bool{}
	for _, v := range productVariants(w, p) {
		if matchesAttributes(variantValues(w, p, v), q.Attributes) {
			matching[v.ID] = true
		}
	}
	seenPrice := map[int64]bool{}
	for _, o := range publicOffers(w, p) {
		if !matching[o.VariantID] || seenPrice[o.Price] {
			continue
		}
		seenPrice[o.Price] = true
		if q.Special && o.RegularPrice <= o.Price {
			continue
		}
		value := o.Price
		if r.Min == nil || value < *r.Min {
			r.Min = &value
		}
		if r.Max == nil || value > *r.Max {
			r.Max = &value
		}
	}
}
