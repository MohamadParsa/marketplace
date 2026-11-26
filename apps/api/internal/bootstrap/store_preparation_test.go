package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
	"reflect"
	"testing"
)

func preparationSeed(t *testing.T) backoffice.Workspace {
	w := storeEligibilitySeed(t)
	w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "prep", Name: "تقویم فروشگاه آزمون", Version: 1, Scope: identity.ScopeStore, Permissions: []string{backoffice.StoresPreparationRead, backoffice.StoresPreparationWrite}})
	w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: "seller-test", StoreID: "s1", RoleID: "prep"})
	return w
}
func httpPreparation() backoffice.StorePreparationInput {
	return backoffice.StorePreparationInput{LeadDays: 2, DayBasis: shipping.CalendarDays, CalendarMode: shipping.Override, WorkingWeekdays: []int{6, 0, 1, 2, 3}}
}

func TestStorePreparationHTTPRequiredFieldsScopesRetryAndPreview(t *testing.T) {
	repo := backofficememory.New(preparationSeed(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := sellerLogin(t, h, "09120000002")
	in := httpPreparation()
	commerceHTTP(t, h, "", "", "GET", "/backoffice/stores/preparation", nil, "", 401)
	commerceHTTP(t, h, viewer, "", "GET", "/backoffice/stores/preparation", nil, "", 403)
	commerceHTTP(t, h, seller, "", "PUT", "/backoffice/stores/s1/preparation", in, "wrong-surface", 401)
	commerceHTTP(t, h, seller, "", "PUT", "/seller/stores/s2/preparation", in, "cross-store", 403)
	for _, bad := range []any{
		map[string]any{"lead_days": 2, "day_basis": "calendar", "calendar_mode": "inherit"},
		map[string]any{"version": 0, "day_basis": "calendar", "calendar_mode": "inherit"},
		map[string]any{"version": nil, "lead_days": 2, "day_basis": "calendar", "calendar_mode": "inherit"},
		map[string]any{"version": 0, "lead_days": 2, "day_basis": "calendar", "calendar_mode": "inherit", "owner": "injected"},
	} {
		commerceHTTP(t, h, admin, "", "PUT", "/backoffice/stores/s1/preparation", bad, "invalid", 400)
	}
	var rows []backoffice.StorePreparationView
	json.Unmarshal(commerceHTTP(t, h, seller, "", "GET", "/seller/stores/preparation", nil, "", 200), &rows)
	if len(rows) != 1 || rows[0].ID != "s1" || !rows[0].CanWrite {
		t.Fatal("seller projection", rows)
	}
	before, _ := repo.Snapshot(context.Background())
	commerceHTTP(t, h, seller, "", "POST", "/seller/stores/s1/preparation/preview", in, "", 200)
	after, _ := repo.Snapshot(context.Background())
	if !reflect.DeepEqual(before, after) {
		t.Fatal("preview mutated state")
	}
	first := commerceHTTP(t, h, seller, "", "PUT", "/seller/stores/s1/preparation", in, "save-prep", 200)
	if string(first) != string(commerceHTTP(t, h, seller, "", "PUT", "/seller/stores/s1/preparation", in, "save-prep", 200)) {
		t.Fatal("retry")
	}
	commerceHTTP(t, h, seller, "", "PUT", "/seller/stores/s1/preparation", in, "stale-prep", 409)
	var row backoffice.StorePreparationView
	json.Unmarshal(first, &row)
	if row.Version != 1 || row.Policy.Version != 1 || row.LeadDays != 2 || row.Policy.DayBasis != shipping.CalendarDays || row.Policy.WorkingWeekdays[0] != 0 {
		t.Fatal(row)
	}
}

func TestStorePreparationPostgresRestartPreservesCalendarAndHistory(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := preparationSeed(t)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	before, _ := repo.Snapshot(ctx)
	commerceHTTP(t, h, admin, "", "PUT", "/backoffice/stores/s1/preparation", httpPreparation(), "persist-prep", 200)
	restarted, err := reviewpostgres.New(ctx, db, backofficememory.Fixture())
	if err != nil {
		t.Fatal(err)
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, restarted)
	admin2 := login(t, h2, "09000000001")
	var rows []backoffice.StorePreparationView
	json.Unmarshal(commerceHTTP(t, h2, admin2, "", "GET", "/backoffice/stores/preparation", nil, "", 200), &rows)
	var found bool
	for _, row := range rows {
		if row.ID == "s1" {
			found = true
			if row.Policy.DayBasis != shipping.CalendarDays || row.Policy.CalendarMode != shipping.Override || row.Version != 1 {
				t.Fatal("restart", row)
			}
		}
	}
	if !found {
		t.Fatal("lost store")
	}
	after, _ := restarted.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Ledger, after.Ledger) || !reflect.DeepEqual(before.Commerce, after.Commerce) {
		t.Fatal("protected data changed")
	}
	public, err := restarted.PublicCatalog(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range public.Stores {
		if st.ID == "s1" && (st.Preparation == nil || st.Preparation.DayBasis != shipping.CalendarDays) {
			t.Fatal("public cache lost day basis")
		}
	}
}
