package bootstrap_test

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

func panelAttributeFixture(t *testing.T) backoffice.Workspace {
	w := listingFixture(t)
	for i := range w.Attributes {
		if w.Attributes[i].ID == "color" {
			w.Attributes[i].ID = "gsp-axis-color"
			w.Attributes[i].Code = "gsp-axis-color"
		}
	}
	for i := range w.Categories {
		for j := range w.Categories[i].Bindings {
			if w.Categories[i].Bindings[j].AttributeID == "color" {
				w.Categories[i].Bindings[j].AttributeID = "gsp-axis-color"
			}
		}
	}
	for i := range w.Variants {
		for j := range w.Variants[i].Values {
			if w.Variants[i].Values[j].AttributeID == "color" {
				w.Variants[i].Values[j].AttributeID = "gsp-axis-color"
			}
		}
	}
	w.Attributes = append(w.Attributes, catalog.AttributeDefinition{ID: "gsp-axis-warranty", Code: "gsp-warranty", Name: "گارانتی", Kind: "enum", Options: []string{"شرکتی الف", "شرکتی ب"}, Groups: []catalog.AttributeValueGroup{{ID: "company", Name: "گارانتی شرکتی", Values: []string{"شرکتی الف", "شرکتی ب"}, Active: true}}, Version: 1})
	return w
}
func TestPanelAttributeReferencesRoundTripAndAuthorization(t *testing.T) {
	ctx := context.Background()
	repo := backofficememory.New(panelAttributeFixture(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	before, _ := repo.Snapshot(ctx)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	var shell backoffice.Workspace
	json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", "/backoffice/workspace", nil, "", 200), &shell)
	alias := backoffice.PublicReference("a", "gsp-axis-color")
	var attr catalog.AttributeDefinition
	for _, a := range shell.Attributes {
		if a.ID == alias {
			attr = a
		}
	}
	if attr.ID != "" && attr.Code != alias {
		t.Fatal("canonical code mismatch")
	}
	if attr.ID == "" {
		t.Fatal("definition not projected")
	}
	for _, c := range shell.Categories {
		for _, b := range c.Bindings {
			if b.AttributeID == "gsp-axis-color" {
				t.Fatal("binding not projected")
			}
		}
	}
	for _, v := range shell.Variants {
		for _, a := range v.Values {
			if a.AttributeID == "gsp-axis-color" {
				t.Fatal("value not projected")
			}
		}
	}
	var product backoffice.ProductWorkspace
	json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", "/backoffice/products/mobile/workspace", nil, "", 200), &product)
	if product.Axes[0].ID != alias || product.Axes[0].Code != alias {
		t.Fatal("workspace axes not projected", product.Axes)
	}
	for _, key := range []string{alias, "gsp-axis-color"} {
		commerceHTTP(t, h, admin, "", "GET", "/backoffice/products/mobile/resolve?attr."+url.QueryEscape(key)+"="+url.QueryEscape("آبی"), nil, "", 200)
	}
	for _, key := range []string{"chr-o-Z3NwLWF4aXMtY29sb3I", "chr-a-!"} {
		commerceHTTP(t, h, admin, "", "GET", "/backoffice/products/mobile/resolve?attr."+key+"=x", nil, "", 400)
	}
	// No mutation from reads, including groups and protected financial objects.
	after, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("read projection mutated repository")
	}
	attr.Name = "رنگ بدنه"
	command := backoffice.Command{Action: "attribute.update", Target: alias, Attribute: &attr, Version: attr.Version, IdempotencyKey: "panel-attribute-update"}
	commerceHTTP(t, h, viewer, "", "POST", "/backoffice/commands", command, "", 403)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", command, "", 204)
	command.Target = "gsp-axis-color"
	command.Attribute.ID = "gsp-axis-color"
	command.Attribute.Code = "gsp-axis-color"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", command, "", 204)
	command.IdempotencyKey = "stale-update"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", command, "", 409)
	command.Target = backoffice.PublicReference("o", "gsp-axis-color")
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", command, "", 400)
	after, _ = repo.Snapshot(ctx)
	if len(after.Attributes) != len(before.Attributes) || after.Attributes[3].ID != "gsp-axis-color" || after.Attributes[3].Code != "gsp-axis-color" || after.Attributes[3].Version != 2 || after.Attributes[3].Name != "رنگ بدنه" {
		t.Fatal("private identity or retry changed")
	}
	if !reflect.DeepEqual(before.Products, after.Products) || !reflect.DeepEqual(before.Variants, after.Variants) || !reflect.DeepEqual(before.Offers, after.Offers) || !reflect.DeepEqual(before.Categories, after.Categories) || !reflect.DeepEqual(before.Attributes[4], after.Attributes[4]) || !reflect.DeepEqual(before.Commerce, after.Commerce) || !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Users, after.Users) || !reflect.DeepEqual(before.Ledger, after.Ledger) {
		t.Fatal("unrelated state rewritten")
	}
	// Aliased and legacy binding inputs share one fingerprint and remain one definition.
	c := shell.Categories[0]
	bind := backoffice.Command{Action: "category.bindings.set", Target: c.ID, Version: c.Version, Bindings: c.Bindings, IdempotencyKey: "alias-bindings"}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", bind, "", 204)
	for i := range bind.Bindings {
		bind.Bindings[i].AttributeID, _ = backoffice.ResolvePublicReference("a", bind.Bindings[i].AttributeID)
	}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", bind, "", 204)
	bad := bind
	bad.IdempotencyKey = "duplicate-binding"
	bad.Version++
	bad.Bindings = append(append([]catalog.AttributeBinding{}, bind.Bindings...), catalog.AttributeBinding{AttributeID: alias, Axis: true})
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", bad, "", 400)
}
func TestPanelAttributeReferencesPostgresRestart(t *testing.T) {
	repo, db := cachePostgres(t)
	ctx := context.Background()
	fixture := panelAttributeFixture(t)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = fixture; return nil }); err != nil {
		t.Fatal(err)
	}
	before, _ := repo.Snapshot(ctx)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	first := commerceHTTP(t, h, admin, "", "GET", "/backoffice/workspace", nil, "", 200)
	reopened, err := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if err != nil {
		t.Fatal(err)
	}
	next := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	second := commerceHTTP(t, next, login(t, next, "09000000001"), "", "GET", "/backoffice/workspace", nil, "", 200)
	if string(first) != string(second) {
		t.Fatal("aliases changed on restart")
	}
	after, _ := reopened.Snapshot(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("private data changed after projection/restart")
	}
}
