package backoffice

import (
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"reflect"
	"testing"
)

func TestCatalogURLPlanReservesImportedNumbersAndAliases(t *testing.T) {
	w := Workspace{Products: []catalog.Product{{ID: "native"}, {ID: "gsp-product-18189"}, {ID: "gsp-product-39782"}}, Categories: []catalog.Category{{ID: "gsp-cat-1", EnglishName: "Mobile"}, {ID: "gsp-cat-755", EnglishName: "gadget"}, {ID: "gsp-cat-163", EnglishName: "PowerBank"}}}
	registry := CatalogURLRegistry{Products: map[string]int64{"deleted": 50000}, Categories: map[string]string{"old-phone": "gsp-cat-1"}, LastNumber: 50000}
	p, e := PlanCatalogURLs(w, registry)
	if e != nil {
		t.Fatal(e)
	}
	if p.Products[0].Code != "chr-18189" || p.Products[1].Code != "chr-39782" || p.Products[2].Code != "chr-50001" {
		t.Fatal(p)
	}
	for _, c := range p.Categories {
		if c.ID == "gsp-cat-755" && c.Slug != "smart-watch" || c.ID == "gsp-cat-163" && c.Slug != "power-bank" {
			t.Fatal(c)
		}
	}
	for _, x := range p.Products {
		registry.Products[x.ID] = x.Number
	}
	for _, x := range p.Categories {
		registry.Categories[x.Slug] = x.ID
	}
	registry.LastNumber = p.LastNumber
	again, e := PlanCatalogURLs(w, registry)
	if e != nil || !reflect.DeepEqual(p, again) {
		t.Fatal("non-idempotent", again, e)
	}
	w.Categories[1].Slug = "old-phone"
	if _, e = PlanCatalogURLs(w, registry); !errors.Is(e, ErrConflict) {
		t.Fatal("alias stolen", e)
	}
	w.Categories[1].Slug = "bad/slug"
	if _, e = PlanCatalogURLs(w, registry); !errors.Is(e, ErrInvalid) {
		t.Fatal(e)
	}
	w.Categories[1].Slug = "smart-watch"
	w.Products[0].PublicCode = "chr-1"
	if _, e = PlanCatalogURLs(w, registry); !errors.Is(e, ErrConflict) {
		t.Fatal("code changed", e)
	}
}
func TestCategoryURLNormalizationAndMissingEnglish(t *testing.T) {
	for input, want := range map[string]string{" Mobile ": "mobile", "PowerBank": "power-bank", "gold_coin": "gold-coin", "  smart -- watch ": "smart-watch"} {
		if got := DefaultCategorySlug(input); got != want {
			t.Fatal(input, got)
		}
	}
	for _, s := range []string{"", "-", "UPPER", "with space", "a/b", "a--b"} {
		if ValidCategorySlug(s) {
			t.Fatal(s)
		}
	}
	_, e := PlanCatalogURLs(Workspace{Categories: []catalog.Category{{ID: "gsp-cat-1"}}}, CatalogURLRegistry{})
	if !errors.Is(e, ErrInvalid) {
		t.Fatal("missing source English accepted", e)
	}
	_, ambiguous := PlanCatalogURLs(Workspace{Products: []catalog.Product{{ID: "gsp-product-18189"}}}, CatalogURLRegistry{Products: map[string]int64{"chr-18189": 777}})
	if !errors.Is(ambiguous, ErrConflict) {
		t.Fatal("deleted internal route shadowed", ambiguous)
	}
	_, e = PlanCatalogURLs(Workspace{Products: []catalog.Product{{ID: "gsp-product-1"}}}, CatalogURLRegistry{Products: map[string]int64{"other": 1}})
	if !errors.Is(e, ErrConflict) {
		t.Fatal("source collision", e)
	}
}

func TestPublicCodeBackfillDoesNotInvalidatePendingContent(t *testing.T) {
	p := catalog.Product{ID: "gsp-product-1", Name: "Before", Version: 1}
	before := approvedContentFingerprint(p)
	p.PublicCode = "chr-1"
	if approvedContentFingerprint(p) != before {
		t.Fatal("URL metadata invalidated pending content")
	}
	p.Name = "Changed"
	if approvedContentFingerprint(p) == before {
		t.Fatal("content change was ignored")
	}
}
