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

func TestSharedAttributeBindingsAndNoInheritance(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	before, _ := repo.Snapshot(ctx)
	var mobile, charger catalog.Category
	for _, c := range before.Categories {
		if c.ID == "mobile" {
			mobile = c
		}
		if c.ID == "charger" {
			charger = c
		}
	}
	if mobile.Bindings[0].AttributeID != charger.Bindings[0].AttributeID {
		t.Fatal("fixture must reuse color identity")
	}
	changed := append([]catalog.AttributeBinding{}, mobile.Bindings...)
	changed[0].Filter = false
	command := backoffice.Command{Action: "category.bindings.set", Target: "mobile", Version: mobile.Version, Bindings: changed, IdempotencyKey: "mobile-flags"}
	if err := s.Execute(ctx, actor(identity.RoleOperations), command); err != nil {
		t.Fatal(err)
	}
	after, _ := repo.Snapshot(ctx)
	for _, c := range after.Categories {
		if c.ID == "charger" {
			if len(c.Bindings) != 1 || c.Bindings[0].AttributeID != "color" || !c.Bindings[0].Filter {
				t.Fatal("sibling flags changed or parent memory inherited")
			}
		}
	}
	if err := s.Execute(ctx, actor(identity.RoleOperations), command); err != nil {
		t.Fatal("same key retry", err)
	}
	command.IdempotencyKey = "stale-flags"
	if err := s.Execute(ctx, actor(identity.RoleOperations), command); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatalf("stale: %v", err)
	}
	state, _ := repo.Snapshot(ctx)
	if len(state.Audit) != 1 {
		t.Fatal("retry duplicated audit")
	}
	// No fallback to parent: charger has a value but no explicit memory binding.
	state.Products = append(state.Products, catalog.Product{ID: "charger-with-memory", CategoryID: "charger", Attributes: []catalog.AttributeValue{{AttributeID: "memory", Value: "128"}}})
	for _, p := range backoffice.BuildCardPreviews(state) {
		if p.ProductID == "charger-with-memory" && len(p.Features) != 0 {
			t.Fatal("parent attribute leaked into child preview")
		}
	}
}

func TestCardProjectionSettingsOrderingAndHistoricalIsolation(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	before, _ := repo.Snapshot(ctx)
	history, _ := json.Marshal([]any{before.Orders, before.Products, before.Offers})
	features := func(state backoffice.Workspace) []catalog.CardFeature {
		for _, p := range backoffice.BuildCardPreviews(state) {
			if p.ProductID == "mobile-sample" {
				return p.Features
			}
		}
		t.Fatal("no sample")
		return nil
	}
	initial := features(before)
	if len(initial) != 3 || initial[0].AttributeID != "memory" || initial[2].AttributeID != "camera" {
		t.Fatal(initial)
	}
	if err := s.Execute(ctx, actor(identity.RoleSuperAdmin), backoffice.Command{Action: "settings.card-limit", Target: "display", Amount: 4, Version: 1, IdempotencyKey: "limit-four"}); err != nil {
		t.Fatal(err)
	}
	state, _ := repo.Snapshot(ctx)
	if len(features(state)) != 4 {
		t.Fatal("new cap not applied")
	}
	for _, c := range state.Categories {
		if c.ID == "mobile" {
			for i := range c.Bindings {
				if c.Bindings[i].AttributeID == "screen" {
					c.Bindings[i].Position = 0
				}
			}
			if err := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "category.bindings.set", Target: c.ID, Version: c.Version, Bindings: c.Bindings, IdempotencyKey: "screen-first"}); err != nil {
				t.Fatal(err)
			}
		}
	}
	state, _ = repo.Snapshot(ctx)
	if features(state)[0].AttributeID != "screen" {
		t.Fatal("order not applied")
	}
	if err := s.Execute(ctx, actor(identity.RoleSuperAdmin), backoffice.Command{Action: "settings.card-limit", Target: "display", Amount: 2, Version: 2, IdempotencyKey: "limit-two"}); err != nil {
		t.Fatal(err)
	}
	state, _ = repo.Snapshot(ctx)
	if len(features(state)) != 2 || features(state)[0].AttributeID != "screen" {
		t.Fatal("updated projection not applied")
	}
	historyAfter, _ := json.Marshal([]any{state.Orders, state.Products, state.Offers})
	if string(history) != string(historyAfter) {
		t.Fatal("display changes rewrote product, offer or order data")
	}
}

func TestAttributeConflictsValidationAndAxisGuard(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	ops := actor(identity.RoleOperations)
	state, _ := repo.Snapshot(ctx)
	color := state.Attributes[0]
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "attribute.delete", Target: "color", Version: 1, IdempotencyKey: "delete-used"}); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal(err)
	}
	color.Options = append(color.Options, "new-option")
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "attribute.update", Target: "color", Version: 1, Attribute: &color, IdempotencyKey: "change-used"}); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal(err)
	}
	color = state.Attributes[0]
	color.Name = "رنگ مشترک"
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "attribute.update", Target: "color", Version: 1, Attribute: &color, IdempotencyKey: "rename"}); err != nil {
		t.Fatal(err)
	}
	for _, c := range state.Categories {
		if c.ID == "mobile" {
			c.Bindings[0].Axis = false
			if err := s.Execute(ctx, ops, backoffice.Command{Action: "category.bindings.set", Target: c.ID, Version: c.Version, Bindings: c.Bindings, IdempotencyKey: "remove-axis"}); !errors.Is(err, backoffice.ErrConflict) {
				t.Fatal(err)
			}
		}
	}
	invalid := []backoffice.Command{
		{Action: "attribute.create", Attribute: &catalog.AttributeDefinition{Code: "bad", Name: "Bad", Kind: "enum", Options: []string{"x", "x"}}},
		{Action: "attribute.create", Attribute: &catalog.AttributeDefinition{Code: "X", Name: "Bad", Kind: "text"}},
		{Action: "category.bindings.set", Target: "empty-category", Version: 1, Bindings: []catalog.AttributeBinding{{AttributeID: "missing"}}},
		{Action: "category.bindings.set", Target: "empty-category", Version: 1, Bindings: []catalog.AttributeBinding{{AttributeID: "memory", Axis: true}}},
		{Action: "category.bindings.set", Target: "empty-category", Version: 1, Bindings: []catalog.AttributeBinding{{AttributeID: "color"}, {AttributeID: "color"}}},
	}
	for i, c := range invalid {
		c.IdempotencyKey = fmt.Sprintf("invalid-%d", i)
		if err := s.Execute(ctx, ops, c); !errors.Is(err, backoffice.ErrInvalid) {
			t.Fatalf("%d: %v", i, err)
		}
	}
	for _, n := range []int64{0, -1, 21} {
		if err := s.Execute(ctx, actor(identity.RoleSuperAdmin), backoffice.Command{Action: "settings.card-limit", Target: "display", Amount: n, Version: 1, IdempotencyKey: fmt.Sprintf("cap-%d", n)}); !errors.Is(err, backoffice.ErrInvalid) {
			t.Fatal(err)
		}
	}
	after, _ := repo.Snapshot(ctx)
	if len(after.Audit) != 1 || after.DisplaySettings.CardAttributeLimit != 3 {
		t.Fatal("invalid mutations were committed")
	}
}

func TestConcurrentBindingsAndRevokedRole(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "category.bindings.set", Target: "empty-category", Version: 1, Bindings: []catalog.AttributeBinding{{AttributeID: "color", Axis: true}}, IdempotencyKey: fmt.Sprintf("concurrent-%d", i)})
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, backoffice.ErrConflict) {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatal(successes.Load())
	}
	_ = repo.Update(ctx, func(state *backoffice.Workspace) error {
		for i := range state.Users {
			if state.Users[i].ID == "ops" {
				state.Users[i].Role = identity.RoleViewer
			}
		}
		return nil
	})
	if err := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "attribute.delete", Target: "color", Version: 1, IdempotencyKey: "revoked"}); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal(err)
	}
}
