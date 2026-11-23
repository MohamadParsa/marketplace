package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
	"testing"
)

func TestSellerContentHTTPPublicProjectionAndReview(t *testing.T) {
	w := commerceSeed()
	w.Products[0].Model = "independent"
	w.Products[0].OwnerStoreID = "s1"
	w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "content-writer", Name: "محتوا", Scope: identity.ScopeStore, Permissions: []string{"products.content.write"}, Version: 1})
	w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: "seller-test", StoreID: "s1", RoleID: "content-writer"})
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	seller := loginAt(t, h, "seller", "09120000002")["token"].(string)
	command := backoffice.Command{Action: "product.content.submit", Target: "product", Version: 1, IdempotencyKey: "seller-content-submit", Catalog: &backoffice.CatalogInput{Name: "نام پیشنهادی فروشنده", CategoryID: "regular", Content: &catalog.ProductContent{FullDescription: "<p>" + strings.Repeat("متن طولانی فارسی ", 900) + "</p>", ShortDescription: "نسخه پیشنهادی"}}}
	payload, _ := json.Marshal(command)
	if len(payload) < 16384 {
		t.Fatal("test should exceed old HTTP limit")
	}
	requestAs(t, h, seller, "POST", "/seller/commands", string(payload), 204)
	p, _ := pdpRequest(t, h, "/storefront/products/product", 200)
	if p.Name != "محصول آزمون" || strings.Contains(p.Content.ShortDescription, "پیشنهادی") {
		t.Fatal("pending leaked")
	}
	state, _ := repo.Snapshot(context.Background())
	revision := state.Products[0].ContentRevision
	if revision == nil {
		t.Fatal("missing pending")
	}
	raw, _ := json.Marshal(backoffice.Command{Action: "product.content.approve", Target: "product", Version: 2, IdempotencyKey: "seller-self-approve"})
	requestAs(t, h, seller, "POST", "/seller/commands", string(raw), 403)
	raw, _ = json.Marshal(backoffice.Command{Action: "product.content.approve", Target: "product", Version: 2, IdempotencyKey: "admin-approve"})
	requestAs(t, h, admin, "POST", "/backoffice/commands", string(raw), 204)
	p, _ = pdpRequest(t, h, "/storefront/products/product", 200)
	if p.Name != "نام پیشنهادی فروشنده" {
		t.Fatal("approval not public")
	}
}

func TestProductInteractionSettingsDefaultOffAndRemainIndependent(t *testing.T) {
	repo := backofficememory.New(commerceSeed())
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	p, _ := pdpRequest(t, h, "/storefront/products/product", 200)
	if p.Content.ReviewsEnabled || p.Content.QAEnabled {
		t.Fatal("new product defaults")
	}
	for _, flags := range []struct{ reviews, qa bool }{{true, false}, {true, true}, {false, true}, {false, false}} {
		state, _ := repo.Snapshot(context.Background())
		product := state.Products[0]
		c := catalog.ProductContent{ReviewsEnabled: flags.reviews, QAEnabled: flags.qa}
		command := backoffice.Command{Action: "catalog.product.update", Target: product.ID, Version: product.Version, IdempotencyKey: fmt.Sprintf("interaction-flags-%t-%t", flags.reviews, flags.qa), Catalog: &backoffice.CatalogInput{Name: product.Name, CategoryID: product.CategoryID, Published: true, Content: &c}}
		raw, _ := json.Marshal(command)
		requestAs(t, h, admin, "POST", "/backoffice/commands", string(raw), 204)
		view, _ := pdpRequest(t, h, "/storefront/products/product", 200)
		if view.Content.ReviewsEnabled != flags.reviews || view.Content.QAEnabled != flags.qa {
			t.Fatal("settings coupled or not persisted")
		}
		if view.Interactions.Reviews || view.Interactions.QA != flags.qa {
			t.Fatal("Q&A availability must follow only its own toggle and reviews remain deferred")
		}
	}
}
