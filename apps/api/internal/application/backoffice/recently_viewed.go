package backoffice

import (
	"context"
	"sort"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/customer"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const RecentHistoryLimit = 50
const RecentHistoryRetention = 30 * 24 * time.Hour

type RecentlyViewedItem struct {
	ProductID string       `json:"product_id"`
	ViewedAt  time.Time    `json:"viewed_at"`
	Product   *ListingCard `json:"product"`
}
type RecentlyViewedView struct {
	Items         []RecentlyViewedItem `json:"items"`
	Total         int                  `json:"total"`
	Page          int                  `json:"page"`
	PageSize      int                  `json:"page_size"`
	MaxItems      int                  `json:"max_items"`
	RetentionDays int                  `json:"retention_days"`
}

// Sorting a copy prevents read-only compaction from changing the repository's slices.
func compactRecentHistory(entries []customer.RecentlyViewedEntry, now time.Time, owner string) []customer.RecentlyViewedEntry {
	sorted := append([]customer.RecentlyViewedEntry{}, entries...)
	// Persisted array order breaks equal clock ticks by the latest committed view.
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].ViewedAt.After(sorted[j].ViewedAt) })
	result := make([]customer.RecentlyViewedEntry, 0, len(sorted))
	counts := map[string]int{}
	seen := map[[2]string]bool{}
	cutoff := now.Add(-RecentHistoryRetention)
	for _, entry := range sorted {
		if owner == "" || entry.UserID == owner {
			key := [2]string{entry.UserID, entry.ProductID}
			if !entry.ViewedAt.After(cutoff) || counts[entry.UserID] >= RecentHistoryLimit || seen[key] {
				continue
			}
			counts[entry.UserID]++
			seen[key] = true
		}
		result = append(result, entry)
	}
	return result
}
func (s *Service) ReadRecentlyViewed(ctx context.Context, a identity.User, page, size int) (RecentlyViewedView, error) {
	result := RecentlyViewedView{Items: []RecentlyViewedItem{}, Page: page, PageSize: size, MaxItems: RecentHistoryLimit, RetentionDays: 30}
	if page < 1 || page > 1000000 || size < 1 || size > 48 {
		return result, ErrInvalid
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return result, e
	}
	if e = customerActor(w, a); e != nil {
		return result, e
	}
	entries := []customer.RecentlyViewedEntry{}
	for _, entry := range compactRecentHistory(w.RecentlyViewed, s.clock.Now(), a.ID) {
		if entry.UserID == a.ID {
			entries = append(entries, entry)
		}
	}
	result.Total = len(entries)
	start := (page - 1) * size
	if start >= len(entries) {
		return result, nil
	}
	w = indexedCatalog(w)
	for _, entry := range entries[start:min(start+size, len(entries))] {
		item := RecentlyViewedItem{ProductID: entry.ProductID, ViewedAt: entry.ViewedAt}
		if p, e := publicProduct(w, entry.ProductID); e == nil {
			if card, ok := listingCard(w, p, ListingQuery{}); ok {
				item.Product = &card
			}
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}
func (s *Service) RecordRecentlyViewed(ctx context.Context, a identity.User, product string) error {
	if !validWishlistID(product) {
		return ErrInvalid
	}
	return s.repository.Update(ctx, func(w *Workspace) error {
		if e := customerActor(*w, a); e != nil {
			return e
		}
		if _, e := publicProduct(*w, product); e != nil {
			return e
		}
		now := s.clock.Now()
		entries := make([]customer.RecentlyViewedEntry, 0, len(w.RecentlyViewed)+1)
		for _, entry := range w.RecentlyViewed {
			if entry.UserID != a.ID || entry.ProductID != product {
				entries = append(entries, entry)
			}
		}
		entries = append([]customer.RecentlyViewedEntry{{UserID: a.ID, ProductID: product, ViewedAt: now}}, entries...)
		w.RecentlyViewed = compactRecentHistory(entries, now, a.ID)
		return nil
	})
}
func (s *Service) RemoveRecentlyViewed(ctx context.Context, a identity.User, product string) (int, error) {
	if product != "" && !validWishlistID(product) {
		return 0, ErrInvalid
	}
	removed := 0
	e := s.repository.Update(ctx, func(w *Workspace) error {
		if e := customerActor(*w, a); e != nil {
			return e
		}
		kept := make([]customer.RecentlyViewedEntry, 0, len(w.RecentlyViewed))
		for _, entry := range w.RecentlyViewed {
			if entry.UserID == a.ID && (product == "" || product == entry.ProductID) {
				removed++
				continue
			}
			kept = append(kept, entry)
		}
		w.RecentlyViewed = kept
		return nil
	})
	return removed, e
}

// Purge removes expired metadata physically, including accounts that never return.
func (s *Service) PurgeRecentlyViewed(ctx context.Context) (int, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return 0, e
	}
	if len(compactRecentHistory(w.RecentlyViewed, s.clock.Now(), "")) == len(w.RecentlyViewed) {
		return 0, nil
	}
	removed := 0
	e = s.repository.Update(ctx, func(w *Workspace) error {
		entries := compactRecentHistory(w.RecentlyViewed, s.clock.Now(), "")
		removed = len(w.RecentlyViewed) - len(entries)
		w.RecentlyViewed = entries
		return nil
	})
	return removed, e
}
