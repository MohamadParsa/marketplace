package backoffice_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func sellerActor() identity.User {
	return identity.User{ID: "catalog-seller-a", Surface: identity.SurfaceSeller, Role: identity.RoleViewer, Active: true}
}
func pricePtr(n int64) *int64 { return &n }
func stockPtr(n int) *int     { return &n }
func selection(color, warranty string) []catalog.AttributeValue {
	return []catalog.AttributeValue{{AttributeID: "color", Value: color}, {AttributeID: "warranty", Value: warranty}}
}

func TestProductWorkspaceOfferUpdatesPreserveContentAndHistory(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	before, _ := repo.Snapshot(ctx)
	immutable, _ := json.Marshal([]any{before.Products, before.Variants, before.Orders, before.Stores})
	price := backoffice.Command{Action: "offer.price.update", Target: "mob-a-black", Version: 1, IdempotencyKey: "price-one", Offer: &backoffice.OfferInput{Price: pricePtr(121000000)}}
	if err := s.Execute(ctx, actor(identity.RoleOperations), price); err != nil {
		t.Fatal(err)
	}
	if err := s.Execute(ctx, actor(identity.RoleOperations), price); err != nil {
		t.Fatal("retry", err)
	}
	stock := backoffice.Command{Action: "offer.stock.update", Target: "mob-a-black", Version: 2, IdempotencyKey: "stock-one", Offer: &backoffice.OfferInput{Stock: stockPtr(8)}}
	if err := s.Execute(ctx, sellerActor(), stock); err != nil {
		t.Fatal(err)
	}
	after, _ := repo.Snapshot(ctx)
	afterImmutable, _ := json.Marshal([]any{after.Products, after.Variants, after.Orders, after.Stores})
	if string(immutable) != string(afterImmutable) {
		t.Fatal("offer edit changed content, variants, orders or seller lead time")
	}
	w, err := s.ReadProductWorkspace(ctx, actor(identity.RoleViewer), "mobile-sample")
	if err != nil {
		t.Fatal(err)
	}
	var changed, other backoffice.OfferView
	for _, o := range w.Offers {
		if o.ID == "mob-a-black" {
			changed = o
		}
		if o.ID == "mob-b-black" {
			other = o
		}
	}
	if changed.Price != 121000000 || changed.Stock != 8 || changed.Reserved != 2 || changed.Available != 6 || changed.Version != 3 || changed.LeadDays != 1 {
		t.Fatal(changed)
	}
	if other.Price != 118000000 || other.Version != 1 || len(after.Audit) != 2 {
		t.Fatal("other seller or audit changed")
	}
}

func TestProductWorkspaceStockValidationAndRollback(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	cases := []struct {
		name string
		c    backoffice.Command
		want error
	}{
		{"below-reserved", backoffice.Command{Action: "offer.stock.update", Target: "mob-a-black", Version: 1, Offer: &backoffice.OfferInput{Stock: stockPtr(1)}}, backoffice.ErrConflict},
		{"stale", backoffice.Command{Action: "offer.price.update", Target: "mob-a-black", Version: 99, Offer: &backoffice.OfferInput{Price: pricePtr(100)}}, backoffice.ErrConflict},
		{"negative-stock", backoffice.Command{Action: "offer.stock.update", Target: "mob-a-black", Version: 1, Offer: &backoffice.OfferInput{Stock: stockPtr(-1)}}, backoffice.ErrInvalid},
		{"nonpositive-price", backoffice.Command{Action: "offer.price.update", Target: "mob-a-black", Version: 1, Offer: &backoffice.OfferInput{Price: pricePtr(0)}}, backoffice.ErrInvalid},
		{"mixed-permission-payload", backoffice.Command{Action: "offer.price.update", Target: "mob-a-black", Version: 1, Offer: &backoffice.OfferInput{Price: pricePtr(100), Stock: stockPtr(100)}}, backoffice.ErrInvalid},
		{"reparent-offer", backoffice.Command{Action: "offer.price.update", Target: "mob-a-black", Version: 1, Offer: &backoffice.OfferInput{StoreID: "s2", Price: pricePtr(100)}}, backoffice.ErrInvalid},
		{"overflow", backoffice.Command{Action: "offer.price.update", Target: "mob-a-black", Version: 1, Offer: &backoffice.OfferInput{Price: pricePtr(backoffice.MaxOfferPrice + 1)}}, backoffice.ErrInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.c.IdempotencyKey = tc.name
			if err := s.Execute(ctx, actor(identity.RoleOperations), tc.c); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
	state, _ := repo.Snapshot(ctx)
	if len(state.Audit) != 0 {
		t.Fatal("failed writes left audit")
	}
}

func TestProductWorkspaceScopedPermissionsAndRevocation(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	seller := sellerActor()
	w, err := s.ReadProductWorkspace(ctx, seller, "mobile-sample")
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Offers) != 2 || len(w.Stores) != 1 || w.Permissions.CreateVariant {
		t.Fatal("seller data/permissions leaked", w)
	}
	for _, o := range w.Offers {
		if o.StoreID != "s1" {
			t.Fatal("other seller offer leaked")
		}
	}
	state, err := s.Read(ctx, seller)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Users)+len(state.Orders)+len(state.Cases)+len(state.Transfers)+len(state.Audit)+len(state.Memberships) > 0 {
		t.Fatal("private workspace leaked")
	}
	if _, err := s.ReadProductWorkspace(ctx, seller, "p2"); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("other independent product", err)
	}
	foreign := backoffice.Command{Action: "offer.price.update", Target: "mob-b-black", Version: 1, IdempotencyKey: "foreign", Offer: &backoffice.OfferInput{Price: pricePtr(100)}}
	if err := s.Execute(ctx, seller, foreign); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("IDOR", err)
	}
	foreign.Target = "mob-a-black"
	foreign.IdempotencyKey = "viewer"
	if err := s.Execute(ctx, actor(identity.RoleViewer), foreign); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("viewer write", err)
	}
	// Grant read of B only: union must not transfer A's write permission to B.
	_ = repo.Update(ctx, func(state *backoffice.Workspace) error {
		state.StoreMemberships = append(state.StoreMemberships, identity.StoreMembership{UserID: seller.ID, StoreID: "s2"})
		state.Access.Grants = append(state.Access.Grants, identity.RoleGrant{UserID: seller.ID, StoreID: "s2", RoleID: "store-view"})
		grants := state.Access.Grants[:0]
		for _, g := range state.Access.Grants {
			if g.UserID == seller.ID && g.RoleID == "catalog-stock" {
				continue
			}
			grants = append(grants, g)
		}
		state.Access.Grants = grants
		return nil
	})
	foreign.Target = "mob-b-black"
	foreign.IdempotencyKey = "foreign-read-grant"
	if err := s.Execute(ctx, seller, foreign); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("scoped write crossed memberships", err)
	}
	stock := backoffice.Command{Action: "offer.stock.update", Target: "mob-a-black", Version: 1, IdempotencyKey: "no-stock-permission", Offer: &backoffice.OfferInput{Stock: stockPtr(8)}}
	if err := s.Execute(ctx, seller, stock); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("price permission implied stock", err)
	}
	foreign.Target = "mob-a-black"
	foreign.IdempotencyKey = "own-price"
	if err := s.Execute(ctx, seller, foreign); err != nil {
		t.Fatal(err)
	}
	_ = repo.Update(ctx, func(state *backoffice.Workspace) error {
		state.StoreMemberships = nil
		return nil
	})
	if err := s.Execute(ctx, seller, foreign); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("revoked retry", err)
	}
}

func TestProductWorkspaceVariantValidationAndResolution(t *testing.T) {
	s, _ := testService()
	ctx := context.Background()
	ops := actor(identity.RoleOperations)
	resolved, err := s.ResolveVariant(ctx, ops, "mobile-sample", selection("مشکی", "۱۸ ماهه"))
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Variant.ID != "mobile-black-18" || len(resolved.Offers) != 2 || resolved.Offers[0].ID != "mob-b-black" {
		t.Fatal(resolved)
	}
	white, err := s.ResolveVariant(ctx, ops, "mobile-sample", selection("سفید", "۲۴ ماهه"))
	if err != nil {
		t.Fatal(err)
	}
	if len(white.Offers) != 1 || white.Offers[0].ID != "mob-b-white" {
		t.Fatal("fully reserved offer still purchasable", white)
	}
	sellerWhite, err := s.ResolveVariant(ctx, sellerActor(), "mobile-sample", selection("سفید", "۲۴ ماهه"))
	if err != nil || len(sellerWhite.Offers) != 0 {
		t.Fatal("resolution bypassed seller scope", err)
	}
	cmd := backoffice.Command{Action: "variant.create", Target: "mobile-sample", Version: 1, IdempotencyKey: "new-variant", Variant: &backoffice.VariantInput{ProductID: "mobile-sample", SKU: "TEST-NEW", Values: selection("آبی", "۱۸ ماهه")}}
	if err := s.Execute(ctx, sellerActor(), cmd); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal(err)
	}
	if err := s.Execute(ctx, ops, cmd); err != nil {
		t.Fatal(err)
	}
	if err := s.Execute(ctx, ops, cmd); err != nil {
		t.Fatal("retry", err)
	}
	cmd.Version = 2
	cmd.Variant.SKU = "DIFFERENT"
	cmd.IdempotencyKey = "duplicate-combination"
	if err := s.Execute(ctx, ops, cmd); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("duplicate allowed", err)
	}
	cmd.Variant.Values = selection("طلایی", "۱۸ ماهه")
	cmd.IdempotencyKey = "invalid-option"
	if err := s.Execute(ctx, ops, cmd); !errors.Is(err, backoffice.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.ResolveVariant(ctx, ops, "mobile-sample", selection("مشکی", "۲۴ ماهه")); !errors.Is(err, backoffice.ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := s.ResolveVariant(ctx, ops, "mobile-sample", selection("مشکی", "")); !errors.Is(err, backoffice.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestProductWorkspaceCreateOfferAndConcurrentWrites(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	ops := actor(identity.RoleOperations)
	create := backoffice.Command{Action: "offer.create", Target: "mobile-sample", Version: 1, IdempotencyKey: "offer-create", Offer: &backoffice.OfferInput{ProductID: "mobile-sample", VariantID: "mobile-white-24", StoreID: "s4", Price: pricePtr(124000000), Stock: stockPtr(10)}}
	if err := s.Execute(ctx, ops, create); err != nil {
		t.Fatal(err)
	}
	if err := s.Execute(ctx, ops, create); err != nil {
		t.Fatal(err)
	}
	create.IdempotencyKey = "duplicate-offer"
	if err := s.Execute(ctx, ops, create); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal(err)
	}
	create.Target = "p2"
	create.Offer.ProductID = "p2"
	create.IdempotencyKey = "mismatch-variant"
	create.Offer.StoreID = "s2"
	if err := s.Execute(ctx, ops, create); !errors.Is(err, backoffice.ErrInvalid) {
		t.Fatal("foreign variant", err)
	}
	var success atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := s.Execute(ctx, ops, backoffice.Command{Action: "offer.stock.update", Target: "mob-a-black", Version: 1, IdempotencyKey: fmt.Sprintf("stock-%d", i), Offer: &backoffice.OfferInput{Stock: stockPtr(10 + i)}})
			if err == nil {
				success.Add(1)
			} else if !errors.Is(err, backoffice.ErrConflict) {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if success.Load() != 1 {
		t.Fatal("lost update", success.Load())
	}
	state, _ := repo.Snapshot(ctx)
	if len(state.Audit) != 2 {
		t.Fatal("duplicate effects", len(state.Audit))
	}
}

func TestProductWorkspaceArbitraryAxisCountAndCurrentReservation(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	ops := actor(identity.RoleOperations)
	_ = repo.Update(ctx, func(state *backoffice.Workspace) error {
		state.Attributes = append(state.Attributes, catalog.AttributeDefinition{ID: "finish", Code: "finish", Name: "روکش آزمون", Kind: "enum", Options: []string{"مات", "براق"}, Version: 1})
		for i := range state.Categories {
			if state.Categories[i].ID == "mobile" {
				state.Categories[i].Bindings = append(state.Categories[i].Bindings, catalog.AttributeBinding{AttributeID: "finish", Axis: true})
			}
		}
		for i := range state.Variants {
			if state.Variants[i].ProductID == "mobile-sample" {
				state.Variants[i].Values = append(state.Variants[i].Values, catalog.AttributeValue{AttributeID: "finish", Value: "مات"})
			}
		}
		for i := range state.Offers {
			if state.Offers[i].ID == "mob-a-black" {
				state.Offers[i].Reserved = 4
			}
		}
		return nil
	})
	values := append(selection("مشکی", "۱۸ ماهه"), catalog.AttributeValue{AttributeID: "finish", Value: "مات"})
	resolved, err := s.ResolveVariant(ctx, ops, "mobile-sample", values)
	if err != nil || resolved.Variant.ID != "mobile-black-18" {
		t.Fatal("third axis was not supported", err)
	}
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "offer.stock.update", Target: "mob-a-black", Version: 1, IdempotencyKey: "changed-reservation", Offer: &backoffice.OfferInput{Stock: stockPtr(3)}}); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("stale displayed reservation used", err)
	}
}
