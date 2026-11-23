package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"html"
	"net/url"
	"sort"
	"strings"
	"unicode/utf8"
)

type ProductBrand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type PublicProductVariant struct {
	ID        string                   `json:"id"`
	Values    []catalog.AttributeValue `json:"values"`
	Available bool                     `json:"available"`
	ImageURL  string                   `json:"image_url,omitempty"`
}
type PublicProductOffer struct {
	ID           string `json:"id"`
	VariantID    string `json:"variant_id"`
	StoreID      string `json:"store_id"`
	StoreName    string `json:"store_name"`
	LeadDays     int    `json:"lead_days"`
	Available    bool   `json:"available"`
	Price        *int64 `json:"price"`
	RegularPrice int64  `json:"regular_price,omitempty"`
	Version      int    `json:"version"`
}
type ProductDetail struct {
	ID                string                        `json:"id"`
	Name              string                        `json:"name"`
	EnglishName       string                        `json:"english_name"`
	URL               string                        `json:"url"`
	Version           int                           `json:"version"`
	Category          CatalogNav                    `json:"category"`
	Brand             *ProductBrand                 `json:"brand"`
	Breadcrumbs       []CatalogNav                  `json:"breadcrumbs"`
	Navigation        []CatalogNav                  `json:"navigation"`
	Content           catalog.ProductContent        `json:"content"`
	Gallery           []string                      `json:"gallery"`
	Axes              []catalog.AttributeDefinition `json:"axes"`
	Variants          []PublicProductVariant        `json:"variants"`
	Offers            []PublicProductOffer          `json:"offers"`
	SelectedVariantID string                        `json:"selected_variant_id"`
	SelectedOfferID   string                        `json:"selected_offer_id"`
	Features          []catalog.CardFeature         `json:"features"`
	Specifications    []catalog.CardFeature         `json:"specifications"`
	Interactions      struct {
		Reviews bool `json:"reviews"`
		QA      bool `json:"qa"`
		Notify  bool `json:"notify"`
	} `json:"interactions"`
	CheckoutEnabled bool `json:"checkout_enabled"`
}

func publicProduct(w Workspace, id string) (catalog.Product, error) {
	for _, p := range w.Products {
		if p.ID == id && p.Status == "published" && publicCategory(w, p.CategoryID) != nil && validBrand(w, p.BrandID) {
			return p, nil
		}
	}
	return catalog.Product{}, ErrNotFound
}
func productContent(p catalog.Product) catalog.ProductContent {
	c := catalog.ProductContent{Advantages: []string{}, Disadvantages: []string{}}
	if p.Content != nil {
		c = *p.Content
	}
	if c.Advantages == nil {
		c.Advantages = []string{}
	}
	if c.Disadvantages == nil {
		c.Disadvantages = []string{}
	}
	if c.FullDescription == "" {
		c.FullDescription = "<p>" + strings.ReplaceAll(html.EscapeString(p.Description), "\n", "<br>") + "</p>"
	}
	if strings.TrimSpace(c.SEO.MetaTitle) == "" {
		c.SEO.MetaTitle = p.Name + " | Cherriva"
	}
	if strings.TrimSpace(c.SEO.MetaDescription) == "" {
		text := c.ShortDescription
		if strings.TrimSpace(text) == "" {
			text = p.Description
		}
		text = strings.Join(strings.Fields(text), " ")
		r := []rune(text)
		if len(r) > 400 {
			r = r[:400]
		}
		c.SEO.MetaDescription = string(r)
	}
	c.FullDescription = sanitizeProductHTML(c.FullDescription)
	return c
}
func productPhotoURL(w Workspace, id string) string {
	if imageAttached(w, id) {
		return "/catalog/icons/" + url.PathEscape(id)
	}
	return ""
}

func BuildProductDetail(w Workspace, id, variantID, offerID string) (ProductDetail, error) {
	result := ProductDetail{CheckoutEnabled: true, Gallery: []string{}, Features: []catalog.CardFeature{}, Variants: []PublicProductVariant{}, Offers: []PublicProductOffer{}, Breadcrumbs: []CatalogNav{}, Specifications: []catalog.CardFeature{}, Navigation: publicNavigation(w)}
	p, err := publicProduct(w, id)
	if err != nil {
		return result, err
	}
	result.ID, result.Name, result.EnglishName, result.URL, result.Version = p.ID, p.Name, p.EnglishName, "/product/"+url.PathEscape(p.ID), p.Version
	result.Content = productContent(p)
	// Full interaction services remain deferred (CV-WEB-15/16); toggles alone do not enable them.
	policy := catalog.ProductInteractions(p.Content, true, catalog.InteractionCapabilities{})
	result.Interactions.Reviews = policy.Reviews.NewSubmissions
	result.Interactions.QA = policy.QA.NewSubmissions
	category := publicCategory(w, p.CategoryID)
	result.Category = catalogNav(w, *category)
	chain := []CatalogNav{}
	for c := category; c != nil; {
		chain = append(chain, catalogNav(w, *c))
		if c.ParentID == "" {
			break
		}
		c = publicCategory(w, c.ParentID)
	}
	for i := len(chain) - 1; i >= 0; i-- {
		result.Breadcrumbs = append(result.Breadcrumbs, chain[i])
	}
	for _, b := range w.Brands {
		if b.ID == p.BrandID {
			result.Brand = &ProductBrand{b.ID, b.Name}
		}
	}
	result.Axes = productAxes(w, p)
	vs := productVariants(w, p)
	// Shared definitions can contain hundreds of values. Public choices belong
	// to this product's variants, including declared but unavailable variants.
	for i := range result.Axes {
		a := &result.Axes[i]
		used := map[string]bool{}
		for _, v := range vs {
			for _, value := range v.Values {
				if value.AttributeID == a.ID {
					used[value.Value] = true
				}
			}
		}
		options := make([]string, 0, len(a.Options))
		for _, value := range a.Options {
			if used[value] {
				options = append(options, value)
			}
		}
		a.Options = options
	}
	if variantID == "" {
		offers := publicOffers(w, p)
		if len(offers) > 0 {
			variantID = offers[0].VariantID
		} else if len(vs) > 0 {
			variantID = vs[0].ID
		}
	}
	var selected *catalog.Variant
	availableOffers := publicOffers(w, p)
	for i, v := range vs {
		available := false
		for _, o := range availableOffers {
			available = available || o.VariantID == v.ID
		}
		result.Variants = append(result.Variants, PublicProductVariant{v.ID, v.Values, available, productPhotoURL(w, v.ImageID)})
		if v.ID == variantID {
			selected = &vs[i]
		}
	}
	if selected == nil && variantID != "" {
		return result, ErrInvalid
	}
	result.SelectedVariantID = variantID
	for _, o := range w.Offers {
		if o.ProductID != p.ID || o.VariantID != variantID || o.Price <= 0 {
			continue
		}
		if p.Model == "independent" && p.OwnerStoreID != o.StoreID {
			continue
		}
		for _, st := range w.Stores {
			if st.ID == o.StoreID && st.Active {
				a := o.Stock-o.Reserved > 0
				view := PublicProductOffer{ID: o.ID, VariantID: o.VariantID, StoreID: st.ID, StoreName: st.Name, LeadDays: st.LeadDays, Available: a, Version: o.Version}
				if a {
					price := o.Price
					view.Price = &price
					if o.RegularPrice > o.Price {
						view.RegularPrice = o.RegularPrice
					}
				}
				result.Offers = append(result.Offers, view)
				break
			}
		}
	}
	sort.Slice(result.Offers, func(i, j int) bool {
		a, b := result.Offers[i], result.Offers[j]
		if a.Available != b.Available {
			return a.Available
		}
		if a.Available && *a.Price != *b.Price {
			return *a.Price < *b.Price
		}
		if a.LeadDays != b.LeadDays {
			return a.LeadDays < b.LeadDays
		}
		return a.ID < b.ID
	})
	if offerID != "" {
		found := false
		for _, o := range result.Offers {
			if o.ID == offerID {
				found = true
			}
		}
		if !found {
			return result, ErrInvalid
		}
		result.SelectedOfferID = offerID
	} else if len(result.Offers) > 0 {
		result.SelectedOfferID = result.Offers[0].ID
	}
	images := []string{}
	if selected != nil {
		images = append(images, selected.ImageID)
	}
	images = append(images, p.ImageID)
	images = append(images, p.GalleryIDs...)
	seen := map[string]bool{}
	for _, image := range images {
		u := productPhotoURL(w, image)
		if u != "" && !seen[u] {
			result.Gallery = append(result.Gallery, u)
			seen[u] = true
		}
	}
	if selected != nil {
		result.Features = listingFeatures(w, p, *selected)
		values := variantValues(w, p, *selected)
		defs := definitions(w)
		for _, b := range categoryBindings(w, p.CategoryID) {
			if a, ok := defs[b.AttributeID]; ok && strings.TrimSpace(values[b.AttributeID]) != "" {
				result.Specifications = append(result.Specifications, catalog.CardFeature{AttributeID: a.ID, Name: a.Name, Value: values[a.ID], Unit: a.Unit})
			}
		}
	}
	return result, nil
}
func (s *Service) PublicProduct(ctx context.Context, id, variant, offer string) (ProductDetail, error) {
	if id == "" || len(id) > 128 || len(variant) > 128 || len(offer) > 128 {
		return ProductDetail{}, ErrInvalid
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return ProductDetail{}, e
	}
	return BuildProductDetail(w, id, variant, offer)
}

type ProductSelectionInput struct {
	VariantID string `json:"variant_id"`
	OfferID   string `json:"offer_id"`
	Quantity  int    `json:"quantity"`
}
type ProductSelection struct {
	ProductID    string `json:"product_id"`
	VariantID    string `json:"variant_id"`
	OfferID      string `json:"offer_id"`
	Quantity     int    `json:"quantity"`
	UnitPrice    int64  `json:"unit_price"`
	OfferVersion int    `json:"offer_version"`
}

// ValidateSelection creates no cart, order, reservation or financial quote.
func (s *Service) ValidateSelection(ctx context.Context, id string, in ProductSelectionInput) (ProductSelection, error) {
	var result ProductSelection
	if in.Quantity < 1 || in.Quantity > 99 || in.VariantID == "" || in.OfferID == "" || utf8.RuneCountInString(in.VariantID) > 128 || utf8.RuneCountInString(in.OfferID) > 128 {
		return result, ErrInvalid
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return result, e
	}
	p, e := publicProduct(w, id)
	if e != nil {
		return result, e
	}
	d, e := BuildProductDetail(w, id, in.VariantID, in.OfferID)
	if e != nil {
		return result, e
	}
	for _, v := range d.Offers {
		if v.ID == in.OfferID && v.Available {
			for _, o := range w.Offers {
				if o.ID == v.ID && o.Stock-o.Reserved >= in.Quantity {
					return ProductSelection{p.ID, in.VariantID, o.ID, in.Quantity, o.Price, o.Version}, nil
				}
			}
		}
	}
	return result, ErrConflict
}
