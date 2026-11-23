package backoffice_test

import (
	"context"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"reflect"
	"testing"
	"time"
)

func TestSeedFictionalStoresKeepSourceFingerprintAndIDs(t *testing.T) {
	seed, media := seedFixture()
	seed.AnonymousStoreNames = true
	rawName := seed.Catalog.Stores[0].Name
	repo := backofficememory.New(backofficememory.Fixture())
	ctx := context.Background()
	_, e := backoffice.ApplyCatalogSeed(ctx, repo, seed, media, false, time.Now())
	if e != nil {
		t.Fatal(e)
	}
	state, e := repo.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, st := range state.Stores {
		if st.ID == "gsp-store-1" {
			found = true
			if st.Name != backoffice.AnonymousCatalogStoreName(st.ID) {
				t.Fatal("real source name stored")
			}
		}
	}
	if !found {
		t.Fatal("missing store")
	}
	if seed.Catalog.Stores[0].Name != rawName {
		t.Fatal("source snapshot mutated")
	}
	seed.AnonymousStoreNames = false
	result, e := backoffice.ApplyCatalogSeed(ctx, repo, seed, media, false, time.Now())
	if e != nil || !result.AlreadyApplied {
		t.Fatal("source fingerprint changed", result, e)
	}
	after, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(after.Stores, state.Stores) {
		t.Fatal("replay restored real name")
	}
}
func TestStoreAnonymizationOnlyNamesAndIdempotence(t *testing.T) {
	seed, media := seedFixture()
	repo := backofficememory.New(backofficememory.Fixture())
	ctx := context.Background()
	_, e := backoffice.ApplyCatalogSeed(ctx, repo, seed, media, false, time.Now())
	if e != nil {
		t.Fatal(e)
	}
	before, _ := repo.Snapshot(ctx)
	n, e := backoffice.AnonymizeImportedStores(ctx, repo, []string{"gsp-store-1"}, time.Now())
	if e != nil || n != 1 {
		t.Fatal(n, e)
	}
	after, _ := repo.Snapshot(ctx)
	for i := range before.Stores {
		if before.Stores[i].ID == "gsp-store-1" {
			before.Stores[i].Name = backoffice.AnonymousCatalogStoreName("gsp-store-1")
		}
	}
	before.Audit = after.Audit
	if !reflect.DeepEqual(before, after) {
		t.Fatal("anonymization touched other state")
	}
	n, e = backoffice.AnonymizeImportedStores(ctx, repo, []string{"gsp-store-1"}, time.Now())
	if e != nil || n != 0 {
		t.Fatal("retry", n, e)
	}
	stable, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(stable, after) {
		t.Fatal("retry wrote state")
	}
	_, e = backoffice.AnonymizeImportedStores(ctx, repo, []string{"s1"}, time.Now())
	if !errors.Is(e, backoffice.ErrInvalid) {
		t.Fatal("non-imported store allowed", e)
	}
	_, e = backoffice.AnonymizeImportedStores(ctx, repo, []string{"gsp-store-1", "gsp-store-missing"}, time.Now())
	if !errors.Is(e, backoffice.ErrNotFound) {
		t.Fatal("unknown store", e)
	}
	rolled, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(rolled, stable) {
		t.Fatal("failed alias operation wrote state")
	}
}
