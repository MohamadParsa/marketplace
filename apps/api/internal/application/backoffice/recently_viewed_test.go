package backoffice_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"sync"
	"testing"
	"time"
)

func TestRecentHistoryRevisitCapAndIndependentRemoval(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	if e := h.s.RecordRecentlyViewed(ctx, h.buyer, "phone"); e != nil {
		t.Fatal(e)
	}
	h.clock.Advance(time.Second)
	h.s.RecordRecentlyViewed(ctx, h.buyer, "gold-product")
	h.clock.Advance(time.Second)
	h.s.RecordRecentlyViewed(ctx, h.buyer, "phone")
	v, e := h.s.ReadRecentlyViewed(ctx, h.buyer, 1, 24)
	if e != nil || v.Total != 2 || v.Items[0].ProductID != "phone" {
		t.Fatal("revisit", v, e)
	}
	if _, e = h.s.SetWishlist(ctx, h.buyer, "phone", true); e != nil {
		t.Fatal(e)
	}
	if e = h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := 0; i < 55; i++ {
			p := w.Products[0]
			p.ID = fmt.Sprintf("recent-%02d", i)
			w.Products = append(w.Products, p)
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 55; i++ {
		h.clock.Advance(time.Second)
		if e = h.s.RecordRecentlyViewed(ctx, h.buyer, fmt.Sprintf("recent-%02d", i)); e != nil {
			t.Fatal(e)
		}
	}
	v, e = h.s.ReadRecentlyViewed(ctx, h.buyer, 1, 24)
	if e != nil || v.Total != 50 || v.Items[0].ProductID != "recent-54" || v.MaxItems != 50 || v.RetentionDays != 30 {
		t.Fatal("cap/order", v, e)
	}
	last, _ := h.s.ReadRecentlyViewed(ctx, h.buyer, 3, 24)
	if len(last.Items) != 2 || last.Items[1].ProductID != "recent-05" {
		t.Fatal("last page/eviction", last)
	}
	h.s.RecordRecentlyViewed(ctx, h.other, "recent-54")
	if n, e := h.s.RemoveRecentlyViewed(ctx, h.buyer, "recent-54"); e != nil || n != 1 {
		t.Fatal(n, e)
	}
	if n, e := h.s.RemoveRecentlyViewed(ctx, h.buyer, "recent-54"); e != nil || n != 0 {
		t.Fatal("delete retry", n, e)
	}
	if n, e := h.s.RemoveRecentlyViewed(ctx, h.buyer, ""); e != nil || n != 49 {
		t.Fatal("clear", n, e)
	}
	other, _ := h.s.ReadRecentlyViewed(ctx, h.other, 1, 24)
	wish, _ := h.s.ReadWishlist(ctx, h.buyer, 1, 24)
	if other.Total != 1 || wish.Total != 1 {
		t.Fatal("clear affected another user/wishlist", other.Total, wish.Total)
	}
}
func TestRecentHistoryRetentionBoundaryAndPhysicalPurge(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.s.RecordRecentlyViewed(ctx, h.buyer, "phone")
	h.clock.Advance(time.Second)
	h.s.RecordRecentlyViewed(ctx, h.other, "gold-product")
	h.clock.Advance(backoffice.RecentHistoryRetention - time.Second)
	v, e := h.s.ReadRecentlyViewed(ctx, h.buyer, 1, 24)
	if e != nil || v.Total != 0 {
		t.Fatal("exact boundary is not expired", e, v)
	}
	state, _ := h.repo.Snapshot(ctx)
	if len(state.RecentlyViewed) != 2 {
		t.Fatal("read unexpectedly writes")
	}
	if n, e := h.s.PurgeRecentlyViewed(ctx); e != nil || n != 1 {
		t.Fatal("physical purge", n, e)
	}
	state, _ = h.repo.Snapshot(ctx)
	if len(state.RecentlyViewed) != 1 || state.RecentlyViewed[0].UserID != h.other.ID {
		t.Fatal("valid history lost")
	}
	h.clock.Advance(time.Second)
	if n, e := h.s.PurgeRecentlyViewed(ctx); e != nil || n != 1 {
		t.Fatal("inactive user purge", n, e)
	}
	if n, e := h.s.PurgeRecentlyViewed(ctx); e != nil || n != 0 {
		t.Fatal("purge retry", n, e)
	}
}
func TestRecentHistoryConcurrentUniqueAndActorRecheck(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if e := h.s.RecordRecentlyViewed(ctx, h.buyer, "phone"); e != nil {
				t.Error(e)
			}
		}()
	}
	group.Wait()
	v, _ := h.s.ReadRecentlyViewed(ctx, h.buyer, 1, 24)
	if v.Total != 1 {
		t.Fatal("duplicate rows", v.Total)
	}
	if e := h.s.RecordRecentlyViewed(ctx, h.admin, "phone"); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal("wrong surface", e)
	}
	if e := h.s.RecordRecentlyViewed(ctx, h.buyer, "missing"); !errors.Is(e, backoffice.ErrNotFound) {
		t.Fatal("missing product", e)
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].ID == h.buyer.ID {
				w.Users[i].AccessVersion++
			}
		}
		return nil
	})
	if _, e := h.s.RemoveRecentlyViewed(ctx, h.buyer, ""); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal("stale actor", e)
	}
	state, _ := h.repo.Snapshot(ctx)
	if len(state.RecentlyViewed) != 1 {
		t.Fatal("stale request cleared history")
	}
}

func TestRecentHistorySameClockRevisitFirst(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	for _, id := range []string{"gold-product", "phone", "gold-product"} {
		if e := h.s.RecordRecentlyViewed(ctx, h.buyer, id); e != nil {
			t.Fatal(e)
		}
	}
	v, e := h.s.ReadRecentlyViewed(ctx, h.buyer, 1, 24)
	if e != nil || v.Total != 2 || v.Items[0].ProductID != "gold-product" {
		t.Fatal("equal timestamps lost latest view", v, e)
	}
}
