package backoffice_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func pos(n int) *int { return &n }

func TestPanelResolveSamePricePrefersEarlierSupply(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	_ = repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Stores {
			if w.Stores[i].ID == "s1" {
				w.Stores[i].LeadDays = 5
			}
			if w.Stores[i].ID == "s2" {
				w.Stores[i].LeadDays = 1
			}
		}
		for i := range w.Offers {
			if w.Offers[i].ID == "mob-a-black" || w.Offers[i].ID == "mob-b-black" {
				w.Offers[i].Price = 100000000
			}
		}
		return nil
	})
	resolved, e := s.ResolveVariant(ctx, actor(identity.RoleOperations), "mobile-sample", selection("مشکی", "۱۸ ماهه"))
	if e != nil || len(resolved.Offers) < 2 || resolved.Offers[0].ID != "mob-b-black" {
		t.Fatal(resolved, e)
	}
}

func TestPanelRolePositionsAreIndependentAndLegacyCompatible(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	w, _ := repo.Snapshot(ctx)
	var category catalog.Category
	for _, c := range w.Categories {
		if c.ID == "mobile" {
			category = c
		}
	}
	for i := range category.Bindings {
		b := &category.Bindings[i]
		if b.AttributeID == "warranty" {
			b.AxisPosition = pos(0)
		}
		if b.AttributeID == "color" {
			b.AxisPosition = pos(1)
		}
		if b.AttributeID == "ram" {
			b.CardPosition = pos(0)
			b.FilterPosition = pos(1)
		}
		if b.AttributeID == "memory" {
			b.CardPosition = pos(1)
			b.FilterPosition = pos(0)
		}
	}
	if e := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "category.bindings.set", Target: category.ID, Version: category.Version, Bindings: category.Bindings, IdempotencyKey: "role-orders"}); e != nil {
		t.Fatal(e)
	}
	workspace, e := s.ReadProductWorkspace(ctx, actor(identity.RoleOperations), "mobile-sample")
	if e != nil || workspace.Axes[0].ID != "warranty" {
		t.Fatal(workspace.Axes, e)
	}
	w, _ = repo.Snapshot(ctx)
	for _, card := range backoffice.BuildCardPreviews(w) {
		if card.ProductID == "mobile-sample" && (len(card.Features) < 2 || card.Features[0].AttributeID != "ram" || card.Features[1].AttributeID != "memory") {
			t.Fatal(card.Features)
		}
	}
	for _, c := range w.Categories {
		if c.ID == "mobile" {
			for _, b := range c.Bindings {
				if b.AttributeID == "ram" && (b.PositionFor("card") != 0 || b.PositionFor("filter") != 1 || b.PositionFor("axis") != b.Position) {
					t.Fatal(b)
				}
			}
		}
	}
	category.Bindings[0].AxisPosition = pos(-1)
	if e := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "category.bindings.set", Target: category.ID, Version: category.Version + 1, Bindings: category.Bindings, IdempotencyKey: "invalid-order"}); !errors.Is(e, backoffice.ErrInvalid) {
		t.Fatal(e)
	}
}

func TestPanelCombinedOfferUpdateIsAtomicScopedAndIdempotent(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	ops := actor(identity.RoleOperations)
	w, _ := s.ReadProductWorkspace(ctx, ops, "mobile-sample")
	var original backoffice.OfferView
	for _, o := range w.Offers {
		if o.ID == "mob-a-black" {
			original = o
		}
	}
	price := original.Price + 100
	stock := original.Stock + 3
	command := backoffice.Command{Action: "offer.update", Target: original.ID, Version: original.Version, IdempotencyKey: "combined", Offer: &backoffice.OfferInput{Price: &price, Stock: &stock}}
	if e := s.Execute(ctx, ops, command); e != nil {
		t.Fatal(e)
	}
	if e := s.Execute(ctx, ops, command); e != nil {
		t.Fatal("retry", e)
	}
	after, _ := repo.Snapshot(ctx)
	for _, o := range after.Offers {
		if o.ID == original.ID && (o.Price != price || o.Stock != stock || o.Version != original.Version+1) {
			t.Fatal(o)
		}
	}
	command.IdempotencyKey = "stale-combined"
	if e := s.Execute(ctx, ops, command); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("stale", e)
	}
	_ = repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Offers {
			if w.Offers[i].ID == original.ID {
				w.Offers[i].Reserved = 2
			}
		}
		return nil
	})
	stock = 1
	command.Version++
	command.IdempotencyKey = "below-reserved"
	price++
	before, _ := repo.Snapshot(ctx)
	if e := s.Execute(ctx, ops, command); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal(e)
	}
	after, _ = repo.Snapshot(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("partial price change on stock failure")
	}
	_ = repo.Update(ctx, func(w *backoffice.Workspace) error {
		grants := w.Access.Grants[:0]
		for _, g := range w.Access.Grants {
			if g.UserID == sellerActor().ID && g.RoleID == "catalog-stock" {
				continue
			}
			grants = append(grants, g)
		}
		w.Access.Grants = grants
		return nil
	})
	command.IdempotencyKey = "price-only-role-cannot-write-stock"
	stock = original.Stock
	if e := s.Execute(ctx, sellerActor(), command); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal(e)
	}
	command.Offer.Stock = nil
	command.IdempotencyKey = "price-only-role"
	if e := s.Execute(ctx, sellerActor(), command); e != nil {
		t.Fatal(e)
	}
}

func TestPanelNumbersSurviveRemovalRestartAndAutomaticSKURetry(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	before, _ := repo.Snapshot(ctx)
	cmd := backoffice.Command{Action: "variant.create", Target: "mobile-sample", Version: 1, IdempotencyKey: "automatic-sku", Variant: &backoffice.VariantInput{ProductID: "mobile-sample", Values: selection("آبی", "۱۸ ماهه")}}
	if e := s.Execute(ctx, actor(identity.RoleOperations), cmd); e != nil {
		t.Fatal(e)
	}
	if e := s.Execute(ctx, actor(identity.RoleOperations), cmd); e != nil {
		t.Fatal(e)
	}
	after, _ := repo.Snapshot(ctx)
	if len(after.Variants) != len(before.Variants)+1 {
		t.Fatal("retry created duplicate")
	}
	created := after.Variants[len(after.Variants)-1]
	if !regexp.MustCompile(`^chr-[0-9]+$`).MatchString(created.SKU) {
		t.Fatal(created.SKU)
	}
	for i, v := range before.Variants {
		if after.Variants[i].SKU != v.SKU || after.Variants[i].ID != v.ID {
			t.Fatal("historical variant changed")
		}
	}
	raw, _ := json.Marshal(after)
	var reopened backoffice.Workspace
	if e := json.Unmarshal(raw, &reopened); e != nil {
		t.Fatal(e)
	}
	newRepo := backofficememory.New(reopened)
	restored, _ := newRepo.Snapshot(ctx)
	if !reflect.DeepEqual(after.RecordNumbers, restored.RecordNumbers) {
		t.Fatal("numbers changed on restart")
	}
	removed := after.Variants[0].ID
	number := after.RecordNumbers["variant|"+removed]
	_ = newRepo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Variants = w.Variants[1:]
		w.Variants = append(w.Variants, catalog.Variant{ID: "fresh-after-delete"})
		return nil
	})
	updated, _ := newRepo.Snapshot(ctx)
	if updated.RecordNumbers["variant|"+removed] != number || updated.RecordNumbers["variant|fresh-after-delete"] <= after.LastRecordNumber {
		t.Fatal("number reused")
	}
	visible, e := s.Read(ctx, sellerActor())
	if e != nil {
		t.Fatal(e)
	}
	if visible.RecordNumbers["user|u-admin"] != 0 || visible.RecordNumbers["offer|mob-b-black"] != 0 {
		t.Fatal("foreign identifiers leaked")
	}
}

func TestPanelCombinedOfferConcurrentLastVersion(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			price := int64(100000000 + n)
			stock := 10 + n
			results <- s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "offer.update", Target: "mob-a-black", Version: 1, IdempotencyKey: []string{"combined-a", "combined-b"}[n], Offer: &backoffice.OfferInput{Price: &price, Stock: &stock}})
		}(i)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for e := range results {
		if e == nil {
			success++
		} else if errors.Is(e, backoffice.ErrConflict) {
			conflict++
		} else {
			t.Fatal(e)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal(success, conflict)
	}
	w, _ := repo.Snapshot(ctx)
	for _, o := range w.Offers {
		if o.ID == "mob-a-black" && o.Stock != int(o.Price-100000000)+10 {
			t.Fatal("mixed command fields", o)
		}
	}
}

func TestPanelAxisConversionPersistsAllZoneSettingsTogether(t *testing.T) {
	h := axisHarness(t)
	ctx := context.Background()
	in := axisInput()
	in.Bindings = []catalog.AttributeBinding{{AttributeID: "old", Card: true, CardPosition: pos(1), Position: 0}, {AttributeID: "new", Axis: true, Card: true, Filter: true, AxisPosition: pos(0), CardPosition: pos(0), FilterPosition: pos(0), Position: 1}}
	preview, e := h.s.PreviewAxisConversion(ctx, h.admin, in)
	if e != nil || !preview.Ready {
		t.Fatal(preview, e)
	}
	plan, e := h.s.PrepareAxisConversion(ctx, h.admin, in, preview.Fingerprint, h.key())
	if e != nil {
		t.Fatal(e)
	}
	if _, e = h.s.ApplyAxisConversion(ctx, h.admin, plan.ID, plan.Fingerprint, true); e != nil {
		t.Fatal(e)
	}
	w, _ := h.repo.Snapshot(ctx)
	for _, c := range w.Categories {
		if c.ID == in.CategoryID && !reflect.DeepEqual(c.Bindings, in.Bindings) {
			t.Fatal(c.Bindings)
		}
	}
}
