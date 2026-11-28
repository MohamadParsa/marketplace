package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"reflect"
	"testing"
)

func TestAttributeCommandAliasesNormalizeWithoutChangingCaller(t *testing.T) {
	id := "gsp-axis-color"
	alias := PublicReference("a", id)
	values := []catalog.AttributeValue{{AttributeID: alias, Value: "آبی"}}
	c := Command{Action: "attribute.update", Target: alias, Attribute: &catalog.AttributeDefinition{ID: alias, Code: alias}, Bindings: []catalog.AttributeBinding{{AttributeID: alias}}, Variant: &VariantInput{Values: values}, Catalog: &CatalogInput{Attributes: &values}}
	state := Workspace{Attributes: []catalog.AttributeDefinition{{ID: id, Code: id}}}
	got, err := normalizeAttributeCommand(state, c)
	if err != nil {
		t.Fatal(err)
	}
	if got.Target != id || got.Attribute.ID != id || got.Attribute.Code != id || got.Bindings[0].AttributeID != id || got.Variant.Values[0].AttributeID != id || (*got.Catalog.Attributes)[0].AttributeID != id {
		t.Fatal("normalization incomplete", got)
	}
	if c.Target != alias || c.Attribute.Code != alias || c.Bindings[0].AttributeID != alias || c.Variant.Values[0].AttributeID != alias || values[0].AttributeID != alias {
		t.Fatal("caller input mutated")
	}
	again, err := normalizeAttributeCommand(state, got)
	if err != nil || !reflect.DeepEqual(got, again) {
		t.Fatal("normalization not idempotent")
	}
}
func TestAttributeCommandRejectsWrongKindsAndReservedCreateCode(t *testing.T) {
	bad := PublicReference("o", "color")
	for _, c := range []Command{{Action: "attribute.delete", Target: bad}, {Bindings: []catalog.AttributeBinding{{AttributeID: bad}}}, {Variant: &VariantInput{Values: []catalog.AttributeValue{{AttributeID: "chr-a-!"}}}}, {Action: "attribute.create", Attribute: &catalog.AttributeDefinition{Code: PublicReference("a", "color")}}} {
		if _, err := normalizeAttributeCommand(Workspace{}, c); err == nil {
			t.Fatal("invalid alias accepted", c)
		}
	}
}
