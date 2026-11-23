package backoffice

import (
	"context"
	"sync"
)

type WarmupResult struct {
	Products, Listings, Failures int
	Skipped                      bool
}

// WarmPublicCache prioritizes first-screen data and never caches user/private data.
// Called on every storefront startup with a bounded context; API startup is independent.
func (s *Service) WarmPublicCache(ctx context.Context, limit int) WarmupResult {
	result := WarmupResult{}
	if s.publicCache == nil || limit < 1 {
		result.Skipped = true
		return result
	}
	if _, e := s.publicCache.store.Get(ctx, "cherriva:public:warmup-probe"); e != nil {
		result.Skipped = true
		result.Failures++
		return result
	}
	if e := s.ReleaseExpiredCommerce(ctx); e != nil {
		result.Failures++
		return result
	}
	home, e := s.PublicHome(ctx)
	if e != nil {
		result.Failures++
		return result
	}
	if _, e = s.PublicBanners(ctx); e != nil {
		result.Failures++
	}
	if _, e := s.PublicNavigation(ctx); e != nil {
		result.Failures++
	}
	ids := []string{}
	seen := map[string]bool{}
	add := func(id string) {
		if id != "" && !seen[id] && len(ids) < limit {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	for _, p := range home.SpecialOffers {
		add(p.ID)
	}
	for _, rail := range home.Rails {
		for _, p := range rail.Products {
			add(p.ID)
		}
		if rail.CategoryID == "" {
			continue
		}
		listing, e := s.PublicListing(ctx, ListingQuery{CategoryID: rail.CategoryID, Sort: "default", Page: 1, PageSize: 20})
		if e != nil {
			result.Failures++
		} else {
			result.Listings++
			for _, p := range listing.Products {
				add(p.ID)
			}
		}
	}
	jobs := make(chan string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				_, e := s.PublicProduct(ctx, id, "", "")
				mu.Lock()
				if e != nil {
					result.Failures++
				} else {
					result.Products++
				}
				mu.Unlock()
			}
		}()
	}
enqueue:
	for _, id := range ids {
		select {
		case jobs <- id:
		case <-ctx.Done():
			break enqueue
		}
	}
	close(jobs)
	wg.Wait()
	return result
}
