package httptransport

import (
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"testing"
)

func TestPanelProjectionPreservesEmptyCollectionShapes(t *testing.T) {
	w := backoffice.Workspace{Categories: []catalog.Category{}, Products: []catalog.Product{}, CardPreviews: []catalog.CardPreview{}, Audit: []backoffice.Audit{}, Variants: []catalog.Variant{}, Attributes: []catalog.AttributeDefinition{}}
	before, _ := json.Marshal(w)
	after, _ := json.Marshal(panelWorkspace(w))
	if string(before) != string(after) {
		t.Fatal("empty arrays changed shape", string(after))
	}
	w.Categories = []catalog.Category{{ID: "category", Bindings: []catalog.AttributeBinding{}}}
	w.CardPreviews = []catalog.CardPreview{{ProductID: "product", Features: []catalog.CardFeature{}}}
	projected := panelWorkspace(w)
	if projected.Categories[0].Bindings == nil || projected.CardPreviews[0].Features == nil {
		t.Fatal("nested empty arrays became null")
	}
}
