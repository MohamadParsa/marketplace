package backoffice_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

func sampleGeographySeed() backoffice.GeographySeed {
	return backoffice.GeographySeed{
		DatasetID: "iran-geography-1404-v3.1.2",
		Source:    backoffice.GeographySeedSource{Repository: "https://github.com/sajaddp/list-of-cities-in-Iran", Commit: strings.Repeat("a", 40), Version: "3.1.2", OfficialYear: 1404, CityDataset: "cities-filtered", ProvinceCount: 2, CityCount: 3},
		Provinces: []commerce.Province{
			{ID: "ir-p-1", Name: "استان یک", Active: true, Cities: []commerce.City{{ID: "ir-c-1", Name: "علی آباد", Active: true}, {ID: "ir-c-2", Name: "شهر دوم", Active: true}}},
			{ID: "ir-p-2", Name: "استان دوم", Active: true, Cities: []commerce.City{{ID: "ir-c-3", Name: "شهر سوم", Active: true}}},
		},
	}
}

func bundledGeography(t *testing.T) backoffice.GeographySeed {
	t.Helper()
	raw, err := os.ReadFile("../../../../../data/iran-geography/iran-provinces-cities.json")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := backoffice.DecodeGeographySeed(raw)
	if err != nil {
		t.Fatal(err)
	}
	return seed
}

func TestBundledIranGeographyCoverageAndDisambiguation(t *testing.T) {
	seed := bundledGeography(t)
	if seed.Source.ProvinceCount != 31 || seed.Source.CityCount != 1481 {
		t.Fatal("unexpected pinned directory coverage")
	}
	names := map[string]bool{}
	for _, p := range seed.Provinces {
		for _, c := range p.Cities {
			names[p.Name+":"+c.Name] = true
		}
	}
	for _, pair := range []string{"تهران:تهران", "تهران:ری", "قم:قم", "هرمزگان:کیش", "اصفهان:اصفهان", "البرز:کرج", "مازندران:گلوگاه (بابل)", "مازندران:گلوگاه (گلوگاه)", "کرمان:کشکوییه (بافت)", "کرمان:کشکوییه (رفسنجان)", "کرمان:علی آباد (جیرفت)", "کرمان:علی آباد (ارزوییه)"} {
		if !names[pair] {
			t.Errorf("missing city %s", pair)
		}
	}
	encoded, _ := json.Marshal(seed.Provinces)
	if len(encoded) > 110*1024 {
		t.Fatal("directory would crowd the existing commerce mutation response limit")
	}
}

func TestGeographyMergePreservesLegacyIDsInactiveFlagsAndInput(t *testing.T) {
	previous := []commerce.Province{
		{ID: "legacy-p", Name: "استان يك", Active: false, Cities: []commerce.City{{ID: "legacy-c", Name: "علي‌آباد", Active: false}}},
		{ID: "custom-p", Name: "محلی", Active: true, Cities: []commerce.City{{ID: "custom-c", Name: "شهر محلی", Active: true}}},
	}
	before, _ := json.Marshal(previous)
	merged, result, err := backoffice.MergeGeographySeed(previous, sampleGeographySeed())
	if err != nil {
		t.Fatal(err)
	}
	if result.ProvincesAdded != 1 || result.CitiesAdded != 2 || result.CitiesMatched != 1 || merged[0].ID != "legacy-p" || merged[0].Active || merged[0].Cities[0] != previous[0].Cities[0] || !reflect.DeepEqual(merged[1], previous[1]) {
		t.Fatalf("legacy destinations changed: %+v", result)
	}
	after, _ := json.Marshal(previous)
	if string(before) != string(after) {
		t.Fatal("merge mutated caller's slices")
	}
	again, result, err := backoffice.MergeGeographySeed(merged, sampleGeographySeed())
	if err != nil || !result.AlreadyApplied || !reflect.DeepEqual(again, merged) {
		t.Fatal("repeated import is not a no-op", err)
	}
}

func TestGeographySeedRejectsAmbiguityCountsAndIDCollisions(t *testing.T) {
	for name, change := range map[string]func(*backoffice.GeographySeed){
		"count":             func(s *backoffice.GeographySeed) { s.Source.CityCount++ },
		"ambiguous spacing": func(s *backoffice.GeographySeed) { s.Provinces[0].Cities[1].Name = "علی‌آباد" },
		"city identity":     func(s *backoffice.GeographySeed) { s.Provinces[0].Cities[1].ID = "ir-c-1" },
		"wrong projection":  func(s *backoffice.GeographySeed) { s.Source.CityDataset = "counties" },
		"inactive source":   func(s *backoffice.GeographySeed) { s.Provinces[0].Active = false },
	} {
		t.Run(name, func(t *testing.T) {
			seed := sampleGeographySeed()
			change(&seed)
			if err := backoffice.ValidateGeographySeed(seed); err == nil {
				t.Fatal("invalid seed accepted")
			}
		})
	}
	seed := sampleGeographySeed()
	previous := []commerce.Province{{ID: "legacy-p", Name: "استان دوم", Active: true, Cities: []commerce.City{{ID: "ir-c-1", Name: "علی آباد", Active: true}}}}
	if _, _, err := backoffice.MergeGeographySeed(previous, seed); err == nil {
		t.Fatal("existing city was silently moved or duplicated")
	}
	raw, _ := json.Marshal(seed)
	for _, data := range [][]byte{append(append([]byte{}, raw...), []byte(` {}`)...), []byte(`{"unexpected":true}`), make([]byte, backoffice.MaxGeographySeedBytes+1)} {
		if _, err := backoffice.DecodeGeographySeed(data); err == nil {
			t.Fatal("malformed seed accepted")
		}
	}
}

func TestGeographyImportIsAtomicAuditedAndDoesNotRewriteBusinessState(t *testing.T) {
	ctx := context.Background()
	w := backofficememory.Fixture()
	w.Commerce = &commerce.State{Settings: commerce.DefaultSettings(), Addresses: []commerce.Address{{ID: "history", Province: "استان یک", City: "علی آباد"}}}
	repo := backofficememory.New(w)
	before, _ := repo.Snapshot(ctx)
	result, err := backoffice.ApplyGeographySeed(ctx, repo, sampleGeographySeed(), before.Commerce.Settings.Version, time.Unix(100, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	after, _ := repo.Snapshot(ctx)
	if result.AlreadyApplied || len(after.Audit) != len(before.Audit)+1 || after.Audit[len(after.Audit)-1].Actor != "geography-import-cli" {
		t.Fatal("missing maintenance audit")
	}
	again, err := backoffice.ApplyGeographySeed(ctx, repo, sampleGeographySeed(), result.SettingsVersion, time.Unix(200, 0).UTC())
	if err != nil || !again.AlreadyApplied {
		t.Fatal("replay", err)
	}
	final, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(final, after) {
		t.Fatal("replay changed state")
	}
	after.Commerce.Settings.Geography = before.Commerce.Settings.Geography
	after.Commerce.Settings.Version = before.Commerce.Settings.Version
	after.Audit = before.Audit
	if !reflect.DeepEqual(before, after) {
		t.Fatal("non-geography business data changed")
	}
	if _, err := backoffice.ApplyGeographySeed(ctx, repo, sampleGeographySeed(), before.Commerce.Settings.Version, time.Now()); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("stale settings version accepted", err)
	}
}

func TestConcurrentGeographyImportsPreserveSettingsVersion(t *testing.T) {
	w := backoffice.EmptyWorkspace()
	repo := backofficememory.New(w)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := backoffice.ApplyGeographySeed(context.Background(), repo, sampleGeographySeed(), w.Commerce.Settings.Version, time.Now())
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, backoffice.ErrConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal("concurrent imports did not serialize", success, conflict)
	}
}
