package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"sort"
	"strings"
	"unicode/utf8"
)

type HomeSettings struct {
	Version int        `json:"version"`
	Rails   []HomeRail `json:"rails"`
}
type HomeRail struct {
	Key        string `json:"key"`
	CategoryID string `json:"category_id"`
	Limit      int    `json:"limit"`
}
type CatalogInput struct {
	Name        string `json:"name"`
	EnglishName string `json:"english_name"`
	Description string `json:"description"`
	CategoryID  string `json:"category_id"`
	BrandID     string `json:"brand_id"`
	ParentID    string `json:"parent_id"`
	Published   bool   `json:"published"`
}

var homeKeys = []string{"mobile", "earbuds", "power-bank", "smartwatch"}
var homeTitles = []string{"موبایل", "هندزفری", "پاوربانک", "ساعت هوشمند"}

func DefaultHomeSettings() HomeSettings {
	h := HomeSettings{Rails: []HomeRail{}}
	for _, key := range homeKeys {
		h.Rails = append(h.Rails, HomeRail{Key: key, Limit: 10})
	}
	return h
}
func IsHomeCatalogAction(a string) bool {
	return a == "homepage.save" || strings.HasPrefix(a, "catalog.")
}
func validCatalogText(c CatalogInput) bool {
	return strings.TrimSpace(c.Name) != "" && utf8.RuneCountInString(c.Name) <= 160 && utf8.RuneCountInString(c.EnglishName) <= 160 && utf8.RuneCountInString(c.Description) <= 4000
}
func publicCategory(w Workspace, id string) *catalog.Category {
	for _, c := range w.Categories {
		if c.ID == id && c.Published && !c.Archived {
			return &c
		}
	}
	return nil
}
func validBrand(w Workspace, id string) bool {
	if id == "" {
		return true
	}
	for _, b := range w.Brands {
		if b.ID == id && b.Active {
			return true
		}
	}
	return false
}
func imageAttached(w Workspace, id string) bool {
	if id == "" {
		return false
	}
	for _, m := range w.IconMedia {
		if m.ID == id && len(m.Data) > 0 {
			return true
		}
	}
	return false
}
func canPublishProduct(w Workspace, p catalog.Product) bool {
	return strings.TrimSpace(p.Name) != "" && publicCategory(w, p.CategoryID) != nil && validBrand(w, p.BrandID) && imageAttached(w, p.ImageID) && len(publicOffers(w, p)) > 0
}
func applyHomeCatalog(w *Workspace, c Command, id string) error {
	if c.Action == "homepage.save" {
		if c.Home == nil || c.Version != w.HomeSettings.Version {
			return ErrConflict
		}
		if len(c.Home.Rails) != 4 {
			return ErrInvalid
		}
		next := HomeSettings{Version: w.HomeSettings.Version + 1, Rails: []HomeRail{}}
		used := map[string]bool{}
		for i, key := range homeKeys {
			r := c.Home.Rails[i]
			if r.Key != key || r.Limit < 1 || r.Limit > 10 {
				return ErrInvalid
			}
			if r.CategoryID != "" {
				if publicCategory(*w, r.CategoryID) == nil || used[r.CategoryID] {
					return ErrInvalid
				}
				used[r.CategoryID] = true
			}
			next.Rails = append(next.Rails, r)
		}
		w.HomeSettings = next
		return nil
	}
	if c.Catalog == nil || !validCatalogText(*c.Catalog) {
		return ErrInvalid
	}
	in := *c.Catalog
	in.Name = strings.TrimSpace(in.Name)
	in.EnglishName = strings.TrimSpace(in.EnglishName)
	switch c.Action {
	case "catalog.category.create", "catalog.category.update":
		if in.ParentID != "" {
			found := false
			for _, x := range w.Categories {
				if x.ID == in.ParentID && !x.Archived {
					found = true
				}
			}
			if !found {
				return ErrInvalid
			}
		}
		if c.Action == "catalog.category.create" {
			w.Categories = append(w.Categories, catalog.Category{ID: id, Name: in.Name, EnglishName: in.EnglishName, Description: in.Description, Published: in.Published, ParentID: in.ParentID, Bindings: []catalog.AttributeBinding{}, Version: 1})
			return nil
		}
		for i := range w.Categories {
			x := &w.Categories[i]
			if x.ID != c.Target {
				continue
			}
			if x.Version != c.Version {
				return ErrConflict
			}
			if x.Archived {
				return ErrConflict
			}
			if in.ParentID != x.ParentID {
				return ErrConflict
			}
			x.Name = in.Name
			x.EnglishName = in.EnglishName
			x.Description = in.Description
			x.Published = in.Published
			x.Version++
			return nil
		}
	case "catalog.brand.create", "catalog.brand.update":
		if c.Action == "catalog.brand.create" {
			w.Brands = append(w.Brands, catalog.Brand{ID: id, Name: in.Name, Active: in.Published, Version: 1})
			return nil
		}
		for i := range w.Brands {
			x := &w.Brands[i]
			if x.ID == c.Target {
				if x.Version != c.Version {
					return ErrConflict
				}
				x.Name = in.Name
				x.Active = in.Published
				x.Version++
				return nil
			}
		}
	case "catalog.product.create", "catalog.product.update":
		found := false
		for _, x := range w.Categories {
			if x.ID == in.CategoryID && !x.Archived {
				found = true
			}
		}
		if !found || !validBrand(*w, in.BrandID) {
			return ErrInvalid
		}
		if c.Action == "catalog.product.create" {
			if in.Published {
				return ErrInvalid
			}
			p := catalog.Product{ID: id, Name: in.Name, EnglishName: in.EnglishName, Description: in.Description, CategoryID: in.CategoryID, BrandID: in.BrandID, Model: "shared", Status: "draft", Version: 1, VariantVersion: 1}
			w.Products = append(w.Products, p)
			if len(productAxes(*w, p)) == 0 {
				w.Variants = append(w.Variants, catalog.Variant{ID: "base-" + id, ProductID: id, SKU: "BASE-" + id, Values: []catalog.AttributeValue{}, Version: 1})
			}
			return nil
		}
		for i := range w.Products {
			p := &w.Products[i]
			if p.ID != c.Target {
				continue
			}
			if p.Version != c.Version {
				return ErrConflict
			}
			if p.CategoryID != in.CategoryID {
				return ErrConflict
			}
			candidate := *p
			candidate.Name, candidate.BrandID = in.Name, in.BrandID
			if in.Published && !canPublishProduct(*w, candidate) {
				return ErrInvalid
			}
			p.Name = in.Name
			p.EnglishName = in.EnglishName
			p.Description = in.Description
			p.BrandID = in.BrandID
			if in.Published {
				p.Status = "published"
			} else {
				p.Status = "draft"
			}
			p.Version++
			return nil
		}
	default:
		return ErrInvalid
	}
	return ErrNotFound
}

type PublicOffer struct {
	RegularPrice int64  `json:"regular_price,omitempty"`
	ID           string `json:"id"`
	VariantID    string `json:"variant_id"`
	Price        int64  `json:"price"`
	LeadDays     int    `json:"lead_days"`
}

func publicOffers(w Workspace, p catalog.Product) []PublicOffer {
	result := []PublicOffer{}
	for _, o := range w.Offers {
		if o.ProductID != p.ID || o.Price <= 0 || o.Stock-o.Reserved <= 0 {
			continue
		}
		valid := false
		for _, v := range w.Variants {
			if v.ID == o.VariantID && v.ProductID == p.ID {
				valid = true
			}
		}
		if !valid {
			continue
		}
		for _, s := range w.Stores {
			if s.ID == o.StoreID && s.Active {
				result = append(result, PublicOffer{RegularPrice: o.RegularPrice, ID: o.ID, VariantID: o.VariantID, Price: o.Price, LeadDays: s.LeadDays})
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.Price != b.Price {
			return a.Price < b.Price
		}
		if a.LeadDays != b.LeadDays {
			return a.LeadDays < b.LeadDays
		}
		return a.ID < b.ID
	})
	return result
}

type HomeCard struct {
	RegularPrice    int64  `json:"regular_price,omitempty"`
	DiscountPercent int    `json:"discount_percent,omitempty"`
	ID              string `json:"id"`
	Name            string `json:"name"`
	CategoryName    string `json:"category_name"`
	URL             string `json:"url"`
	ImageURL        string `json:"image_url"`
	Price           int64  `json:"price"`
	OfferID         string `json:"offer_id"`
	VariantID       string `json:"variant_id"`
}
type HomeSection struct {
	Key        string     `json:"key"`
	Title      string     `json:"title"`
	CategoryID string     `json:"category_id"`
	URL        string     `json:"url"`
	Products   []HomeCard `json:"products"`
}
type HomeCategory struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	ImageURL string `json:"image_url"`
	Count    int    `json:"count"`
}
type HomePage struct {
	Categories       []HomeCategory `json:"categories"`
	SpecialOffers    []HomeCard     `json:"special_offers"`
	SpecialOffersURL string         `json:"special_offers_url"`
	Rails            []HomeSection  `json:"rails"`
}

func BuildHome(w Workspace) HomePage {
	result := HomePage{Rails: []HomeSection{}, Categories: []HomeCategory{}, SpecialOffers: []HomeCard{}, SpecialOffersURL: "/offers"}
	h := w.HomeSettings
	if len(h.Rails) != 4 {
		h = DefaultHomeSettings()
	}
	products := append([]catalog.Product{}, w.Products...)
	sort.Slice(products, func(i, j int) bool { return products[i].ID < products[j].ID })
	counts := map[string]int{}
	images := map[string]string{}
	for _, p := range products {
		cat := publicCategory(w, p.CategoryID)
		if cat == nil || p.Status != "published" || !validBrand(w, p.BrandID) || !imageAttached(w, p.ImageID) {
			continue
		}
		offers := publicOffers(w, p)
		if len(offers) == 0 {
			continue
		}
		counts[cat.ID]++
		images[cat.ID] = "/catalog/icons/" + p.ImageID
		card := makeHomeCard(p, *cat, offers[0])
		if card.RegularPrice > card.Price && len(result.SpecialOffers) < 10 {
			result.SpecialOffers = append(result.SpecialOffers, card)
		}
	}
	for _, cat := range w.Categories {
		if counts[cat.ID] == 0 {
			continue
		}
		img := images[cat.ID]
		if imageAttached(w, cat.IconID) {
			img = "/catalog/icons/" + cat.IconID
		}
		result.Categories = append(result.Categories, HomeCategory{cat.ID, cat.Name, "/category/" + cat.ID, img, counts[cat.ID]})
	}
	sort.Slice(result.Categories, func(i, j int) bool { return result.Categories[i].Name < result.Categories[j].Name })
	for i, key := range homeKeys {
		r := h.Rails[i]
		section := HomeSection{Key: key, Title: homeTitles[i], CategoryID: r.CategoryID, Products: []HomeCard{}}
		cat := publicCategory(w, r.CategoryID)
		if cat != nil {
			section.Title = cat.Name
			section.URL = "/category/" + cat.ID
			for _, p := range products {
				if len(section.Products) >= r.Limit {
					break
				}
				if p.Status != "published" || p.CategoryID != cat.ID || !validBrand(w, p.BrandID) || !imageAttached(w, p.ImageID) {
					continue
				}
				offers := publicOffers(w, p)
				if len(offers) == 0 {
					continue
				}
				best := offers[0]
				section.Products = append(section.Products, makeHomeCard(p, *cat, best))
			}
		}
		result.Rails = append(result.Rails, section)
	}
	return result
}
func (s *Service) PublicHome(ctx context.Context) (HomePage, error) {
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return HomePage{}, err
	}
	return BuildHome(w), nil
}

func makeHomeCard(p catalog.Product, c catalog.Category, o PublicOffer) HomeCard {
	card := HomeCard{ID: p.ID, Name: p.Name, CategoryName: c.Name, URL: "/product/" + p.ID, ImageURL: "/catalog/icons/" + p.ImageID, Price: o.Price, OfferID: o.ID, VariantID: o.VariantID}
	if o.RegularPrice > o.Price {
		card.RegularPrice = o.RegularPrice
		card.DiscountPercent = int((o.RegularPrice - o.Price) * 100 / o.RegularPrice)
	}
	return card
}
