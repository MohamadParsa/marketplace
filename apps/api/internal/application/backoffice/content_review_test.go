package backoffice_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
	"testing"
)

func setupContentReview(t *testing.T) (*commerceHarness, identity.User) {
	h := newCommerceHarness(t)
	var seller identity.User
	err := h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		w.Products[0].Model = "independent"
		w.Products[0].OwnerStoreID = "store1"
		w.Products[0].Name = "نسخه تأییدشده"
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "content-editor", Name: "ویرایش محتوا", Scope: identity.ScopeStore, Permissions: []string{"products.content.write"}, Version: 1})
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: "seller-test", StoreID: "store1"})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: "seller-test", StoreID: "store1", RoleID: "content-editor"})
		for _, u := range w.Users {
			if u.ID == "seller-test" {
				seller = identity.User{ID: u.ID, Active: u.Active, AccessVersion: u.AccessVersion, Surface: identity.SurfaceSeller}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return h, seller
}
func contentInput(name string) *backoffice.CatalogInput {
	return &backoffice.CatalogInput{Name: name, CategoryID: "regular", Content: &catalog.ProductContent{ShortDescription: "متن پیشنهادی", FullDescription: "<p>متن جدید</p><script>alert(1)</script>"}}
}
func currentProduct(t *testing.T, h *commerceHarness) catalog.Product {
	t.Helper()
	w, e := h.repo.Snapshot(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	return w.Products[0]
}
func TestSellerContentReviewPreservesPublishedAndCurrentOffers(t *testing.T) {
	h, seller := setupContentReview(t)
	ctx := context.Background()
	p := currentProduct(t, h)
	command := backoffice.Command{Action: "product.content.submit", Target: p.ID, Version: p.Version, Catalog: contentInput("ویرایش فروشنده"), IdempotencyKey: h.key()}
	if e := h.s.Execute(ctx, seller, command); e != nil {
		t.Fatal(e)
	}
	if e := h.s.Execute(ctx, seller, command); e != nil {
		t.Fatal("retry", e)
	}
	p = currentProduct(t, h)
	if p.Name != "نسخه تأییدشده" || p.Status != "published" || p.ContentRevision == nil || p.ContentRevision.State != "pending" {
		t.Fatal(p)
	}
	w, _ := h.repo.Snapshot(ctx)
	public, e := backoffice.BuildProductDetail(w, p.ID, "", "")
	if e != nil || public.Name != p.Name {
		t.Fatal("pending content leaked", e)
	}
	data, _ := json.Marshal(public)
	if containsString(string(data), "ویرایش فروشنده") {
		t.Fatal("revision public")
	}
	if e = h.s.Execute(ctx, h.admin, backoffice.Command{Action: "product.reject", Target: p.ID, Version: p.Version, Reason: "نباید نسخه قبلی را پنهان کند", IdempotencyKey: h.key()}); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("legacy rejection bypass", e)
	}
	if e = h.s.Execute(ctx, h.admin, backoffice.Command{Action: "product.content.reject", Target: p.ID, Version: p.Version, Reason: "اصلاح لازم است", IdempotencyKey: h.key()}); e != nil {
		t.Fatal(e)
	}
	p = currentProduct(t, h)
	if p.Status != "published" || p.Name != "نسخه تأییدشده" || p.ContentRevision.State != "rejected" {
		t.Fatal("rejection changed public")
	}
	command.Version = p.Version
	command.IdempotencyKey = h.key()
	if e = h.s.Execute(ctx, seller, command); e != nil {
		t.Fatal(e)
	}
	p = currentProduct(t, h)
	// Live price/stock changes after submission must survive content approval.
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Offers[0].Price = 7770000
		w.Offers[0].Reserved = 1
		w.Offers[0].Stock = 7
		return nil
	})
	if e = h.s.Execute(ctx, h.admin, backoffice.Command{Action: "product.content.approve", Target: p.ID, Version: p.Version, IdempotencyKey: h.key()}); e != nil {
		t.Fatal(e)
	}
	p = currentProduct(t, h)
	w, _ = h.repo.Snapshot(ctx)
	if p.Name != "ویرایش فروشنده" || p.Status != "published" || p.ContentRevision.State != "approved" || containsString(p.Content.FullDescription, "script") || w.Offers[0].Price != 7770000 || w.Offers[0].Stock != 7 || w.Offers[0].Reserved != 1 {
		t.Fatal("approval or live offers", p)
	}

}
func containsString(s, part string) bool { return strings.Contains(s, part) }
func TestContentSubmissionOwnershipAndRevocation(t *testing.T) {
	h, seller := setupContentReview(t)
	ctx := context.Background()
	p := currentProduct(t, h)
	command := backoffice.Command{Action: "product.content.submit", Target: p.ID, Version: p.Version, Catalog: contentInput("ویرایش"), IdempotencyKey: h.key()}
	for _, mode := range []string{"shared", "foreign", "no-membership"} {
		t.Run(mode, func(t *testing.T) {
			h, seller := setupContentReview(t)
			h.repo.Update(ctx, func(w *backoffice.Workspace) error {
				switch mode {
				case "shared":
					w.Products[0].Model = "shared"
				case "foreign":
					w.Products[0].OwnerStoreID = "store2"
				case "no-membership":
					w.StoreMemberships = nil
				}
				return nil
			})
			if e := h.s.Execute(ctx, seller, command); !errors.Is(e, backoffice.ErrForbidden) {
				t.Fatal(e)
			}
		})
	}
	if e := h.s.Execute(ctx, seller, command); e != nil {
		t.Fatal(e)
	}
	p = currentProduct(t, h)
	if e := h.s.Execute(ctx, seller, backoffice.Command{Action: "product.content.approve", Target: p.ID, Version: p.Version, IdempotencyKey: h.key()}); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal("self approval", e)
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Access.Grants = nil; return nil })
	if e := h.s.Execute(ctx, seller, command); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal("revoked retry", e)
	}
}
func TestContentRevisionRejectsInvalidStaleAndConcurrentAdminEdit(t *testing.T) {
	h, seller := setupContentReview(t)
	ctx := context.Background()
	p := currentProduct(t, h)
	bad := contentInput("bad")
	bad.Content.Shipping.WeightUnit = "lb"
	if e := h.s.Execute(ctx, seller, backoffice.Command{Action: "product.content.submit", Target: p.ID, Version: p.Version, Catalog: bad, IdempotencyKey: h.key()}); !errors.Is(e, backoffice.ErrInvalid) {
		t.Fatal("invalid accepted", e)
	}
	if currentProduct(t, h).ContentRevision != nil {
		t.Fatal("bad draft saved")
	}
	command := backoffice.Command{Action: "product.content.submit", Target: p.ID, Version: p.Version, Catalog: contentInput("new"), IdempotencyKey: h.key()}
	if e := h.s.Execute(ctx, seller, command); e != nil {
		t.Fatal(e)
	}
	command.IdempotencyKey = h.key()
	if e := h.s.Execute(ctx, seller, command); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("stale submit", e)
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Products[0].Description = "ویرایش تازه ادمین"
		w.Products[0].Version++
		return nil
	})
	p = currentProduct(t, h)
	if e := h.s.Execute(ctx, h.admin, backoffice.Command{Action: "product.content.approve", Target: p.ID, Version: p.Version, IdempotencyKey: h.key()}); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("overwrote newer approved content", e)
	}
}
func TestDirectContentPolicyIsScopedAndDoesNotApplyOldPending(t *testing.T) {
	h, seller := setupContentReview(t)
	ctx := context.Background()
	p := currentProduct(t, h)
	if e := h.s.Execute(ctx, seller, backoffice.Command{Action: "store.content-policy.set", Target: "store1", Value: "direct", Version: 0, IdempotencyKey: h.key()}); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal("seller policy", e)
	}
	if e := h.s.Execute(ctx, seller, backoffice.Command{Action: "product.content.submit", Target: p.ID, Version: p.Version, Catalog: contentInput("pending"), IdempotencyKey: h.key()}); e != nil {
		t.Fatal(e)
	}
	if e := h.s.Execute(ctx, h.admin, backoffice.Command{Action: "store.content-policy.set", Target: "store1", Value: "direct", Version: 0, IdempotencyKey: h.key()}); e != nil {
		t.Fatal(e)
	}
	p = currentProduct(t, h)
	if p.Name != "نسخه تأییدشده" {
		t.Fatal("policy published old pending")
	}
	if e := h.s.Execute(ctx, seller, backoffice.Command{Action: "product.content.submit", Target: p.ID, Version: p.Version, Catalog: contentInput("direct"), IdempotencyKey: h.key()}); e != nil {
		t.Fatal(e)
	}
	p = currentProduct(t, h)
	if p.Name != "direct" || p.ContentRevision.State != "applied-direct" {
		t.Fatal(p)
	}
}
