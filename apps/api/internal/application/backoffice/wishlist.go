package backoffice

import (
	"context"
	"sort"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/domain/customer"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type WishlistItem struct {
	ProductID string       `json:"product_id"`
	Product   *ListingCard `json:"product"`
}
type WishlistView struct {
	Items    []WishlistItem `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
type WishlistMembership struct {
	ProductID string `json:"product_id"`
	Saved     bool   `json:"saved"`
	Total     int    `json:"total"`
}

func wishlistMembership(w Workspace, user, product string) WishlistMembership {
	result := WishlistMembership{ProductID: product}
	for _, entry := range w.Wishlist {
		if entry.UserID == user {
			result.Total++
			if entry.ProductID == product {
				result.Saved = true
			}
		}
	}
	return result
}
func validWishlistID(id string) bool {
	return id != "" && len(id) <= 128 && strings.TrimSpace(id) == id && !strings.ContainsAny(id, "\x00/\\")
}

func (s *Service) ReadWishlist(ctx context.Context, actor identity.User, page, size int) (WishlistView, error) {
	result := WishlistView{Items: []WishlistItem{}, Page: page, PageSize: size}
	if page < 1 || page > 1000000 || size < 1 || size > 48 {
		return result, ErrInvalid
	}
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return result, err
	}
	if err = customerActor(w, actor); err != nil {
		return result, err
	}
	entries := []customer.WishlistEntry{}
	for _, entry := range w.Wishlist {
		if entry.UserID == actor.ID {
			entries = append(entries, entry)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].CreatedAt.Equal(entries[j].CreatedAt) {
			return entries[i].ProductID < entries[j].ProductID
		}
		return entries[i].CreatedAt.After(entries[j].CreatedAt)
	})
	result.Total = len(entries)
	start := (page - 1) * size
	if start >= len(entries) {
		return result, nil
	}
	end := min(start+size, len(entries))
	w = indexedCatalog(w)
	for _, entry := range entries[start:end] {
		item := WishlistItem{ProductID: entry.ProductID}
		if p, e := publicProduct(w, entry.ProductID); e == nil {
			if card, ok := listingCard(w, p, ListingQuery{}); ok {
				item.Product = &card
			}
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}
func (s *Service) ReadWishlistMembership(ctx context.Context, actor identity.User, product string) (WishlistMembership, error) {
	if !validWishlistID(product) {
		return WishlistMembership{}, ErrInvalid
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return WishlistMembership{}, e
	}
	if e = customerActor(w, actor); e != nil {
		return WishlistMembership{}, e
	}
	return wishlistMembership(w, actor.ID, product), nil
}

// PUT/DELETE express desired state: replay never duplicates a bookmark or deletes another user's row.
func (s *Service) SetWishlist(ctx context.Context, actor identity.User, product string, saved bool) (WishlistMembership, error) {
	var result WishlistMembership
	if !validWishlistID(product) {
		return result, ErrInvalid
	}
	e := s.repository.Update(ctx, func(w *Workspace) error {
		if e := customerActor(*w, actor); e != nil {
			return e
		}
		if saved {
			if _, e := publicProduct(*w, product); e != nil {
				return e
			}
		}
		found := false
		entries := make([]customer.WishlistEntry, 0, len(w.Wishlist)+1)
		for _, entry := range w.Wishlist {
			if entry.UserID == actor.ID && entry.ProductID == product {
				found = true
				if !saved {
					continue
				}
			}
			entries = append(entries, entry)
		}
		if saved && !found {
			entries = append(entries, customer.WishlistEntry{UserID: actor.ID, ProductID: product, CreatedAt: s.clock.Now()})
		}
		w.Wishlist = entries
		result = wishlistMembership(*w, actor.ID, product)
		return nil
	})
	return result, e
}
