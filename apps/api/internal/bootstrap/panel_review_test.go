package bootstrap_test

import (
	"context"
	"encoding/json"
	"reflect"
	"regexp"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
)

func TestPanelReviewPostgresNumberingSKUAndCombinedOfferRestart(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = backofficememory.Fixture(); return nil }); e != nil {
		t.Fatal(e)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	before, _ := repo.Snapshot(ctx)
	command := map[string]any{"action": "variant.create", "target": "mobile-sample", "version": 1, "idempotency_key": "pg-auto-sku", "variant": map[string]any{"product_id": "mobile-sample", "values": []map[string]string{{"attribute_id": "color", "value": "آبی"}, {"attribute_id": "warranty", "value": "۱۸ ماهه"}}}}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", command, "", 204)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", command, "", 204)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", map[string]any{"action": "offer.update", "target": "mob-a-black", "version": 1, "idempotency_key": "pg-combined", "offer": map[string]any{"price": 119000000, "stock": 8}}, "", 204)
	state, _ := repo.Snapshot(ctx)
	if len(state.Variants) != len(before.Variants)+1 {
		t.Fatal("duplicate variant")
	}
	last := state.Variants[len(state.Variants)-1]
	if !regexp.MustCompile(`^chr-[0-9]+$`).MatchString(last.SKU) {
		t.Fatal(last)
	}
	for _, o := range state.Offers {
		if o.ID == "mob-a-black" && (o.Price != 119000000 || o.Stock != 8 || o.Reserved != 2 || o.Version != 2) {
			t.Fatal(o)
		}
	}
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	restored, _ := reopened.Snapshot(ctx)
	if !reflect.DeepEqual(state.RecordNumbers, restored.RecordNumbers) || !reflect.DeepEqual(state.Variants, restored.Variants) || !reflect.DeepEqual(state.Offers, restored.Offers) {
		t.Fatal("restart changed identifiers or offers")
	}
	public, e := reopened.PublicCatalog(ctx, "")
	if e != nil {
		t.Fatal(e)
	}
	if len(public.RecordNumbers) != 0 || public.LastRecordNumber != 0 {
		t.Fatal("private number registry leaked to public projection")
	}
	next := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	var workspace backoffice.Workspace
	json.Unmarshal(commerceHTTP(t, next, login(t, next, "09000000001"), "", "GET", "/backoffice/workspace", nil, "", 200), &workspace)
	if workspace.RecordNumbers["variant|"+last.ID] != restored.RecordNumbers["variant|"+last.ID] {
		t.Fatal("display code changed")
	}
}
