package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/content"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
	"time"
)

// CatalogStamp is durable and changes in the same transaction as its inputs.
type CatalogStamp struct {
	Namespace    string
	Generation   int64
	Dependencies int64
	Product      int64
	NextExpiry   *time.Time
}
type PublicCatalogRepository interface {
	CatalogStamp(context.Context, string) (CatalogStamp, error)
	PublicCatalog(context.Context, string) (Workspace, error)
}

// CatalogReadWorkspace contains catalog inputs only, never accounts or financial state.
// It is an internal PostgreSQL projection; only white-listed response DTOs enter Redis.
func CatalogReadWorkspace(w Workspace) Workspace {
	r := Workspace{Categories: w.Categories, Brands: w.Brands, Attributes: w.Attributes, DisplaySettings: w.DisplaySettings, HomeSettings: w.HomeSettings}
	for _, st := range w.Stores {
		r.Stores = append(r.Stores, store.Store{ID: st.ID, Name: st.Name, Active: st.Active, LeadDays: st.LeadDays})
	}
	for _, p := range w.Products {
		p.ContentRevision = nil
		r.Products = append(r.Products, p)
	}
	for _, v := range w.Variants {
		v.SKU = ""
		r.Variants = append(r.Variants, v)
	}
	r.Offers = w.Offers
	for _, m := range w.IconMedia {
		r.IconMedia = append(r.IconMedia, catalog.IconMedia{ID: m.ID, Stored: m.Stored || len(m.Data) > 0})
	}
	for _, b := range w.Banners {
		if b.Published != nil {
			r.Banners = append(r.Banners, content.Banner{ID: b.ID, Name: b.Name, Published: b.Published})
		}
	}
	return r
}
func (s *Service) publicSnapshot(ctx context.Context, id string) (Workspace, error) {
	if r, ok := s.repository.(PublicCatalogRepository); ok {
		return r.PublicCatalog(ctx, id)
	}
	return s.repository.Snapshot(ctx)
}

type publicCatalogIndex struct {
	categories  map[string]catalog.Category
	visible     map[string]bool
	brands      map[string]bool
	media       map[string]bool
	stores      map[string]store.Store
	variants    map[string][]catalog.Variant
	offers      map[string][]catalog.Offer
	definitions map[string]catalog.AttributeDefinition
}

func indexedCatalog(w Workspace) Workspace {
	if w.publicIndex != nil {
		return w
	}
	idx := &publicCatalogIndex{categories: map[string]catalog.Category{}, visible: map[string]bool{}, brands: map[string]bool{}, media: map[string]bool{}, stores: map[string]store.Store{}, variants: map[string][]catalog.Variant{}, offers: map[string][]catalog.Offer{}, definitions: map[string]catalog.AttributeDefinition{}}
	for _, c := range w.Categories {
		idx.categories[c.ID] = c
	}
	// Compute using the existing cycle-safe visibility rule, before attaching the index.
	for _, c := range w.Categories {
		idx.visible[c.ID] = publicCategory(w, c.ID) != nil
	}
	for _, b := range w.Brands {
		idx.brands[b.ID] = b.Active
	}
	for _, m := range w.IconMedia {
		idx.media[m.ID] = m.Stored || len(m.Data) > 0
	}
	for _, st := range w.Stores {
		idx.stores[st.ID] = st
	}
	for _, v := range w.Variants {
		idx.variants[v.ProductID] = append(idx.variants[v.ProductID], v)
	}
	for _, o := range w.Offers {
		idx.offers[o.ProductID] = append(idx.offers[o.ProductID], o)
	}
	for _, d := range w.Attributes {
		idx.definitions[d.ID] = d
	}
	w.publicIndex = idx
	return w
}

// PublicCatalogMediaIDs preserves the existing public-image authorization rules.
func PublicCatalogMediaIDs(w Workspace) []string {
	w = indexedCatalog(w)
	allowed := map[string]bool{}
	for _, c := range w.Categories {
		if !c.Archived && imageAttached(w, c.IconID) {
			allowed[c.IconID] = true
		}
	}
	for _, b := range w.Brands {
		if b.Active && imageAttached(w, b.IconID) {
			allowed[b.IconID] = true
		}
	}
	for _, p := range w.Products {
		if p.Status != "published" || publicCategory(w, p.CategoryID) == nil || !validBrand(w, p.BrandID) {
			continue
		}
		if imageAttached(w, p.ImageID) {
			allowed[p.ImageID] = true
		}
		for _, id := range p.GalleryIDs {
			if imageAttached(w, id) {
				allowed[id] = true
			}
		}
		for _, v := range w.publicIndex.variants[p.ID] {
			if imageAttached(w, v.ImageID) {
				allowed[v.ImageID] = true
			}
		}
	}
	ids := make([]string, 0, len(allowed))
	for id := range allowed {
		ids = append(ids, id)
	}
	return ids
}
