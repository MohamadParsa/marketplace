package httptransport

import (
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCatalogReferenceProjectionPreservesMoneyAndText(t *testing.T) {
	input := json.RawMessage(`{"public_code":"chr-12034","id":"gsp-product-12034","name":"gsp-product-12034 is descriptive text","gallery":["/catalog/icons/gsp-media-7"],"selected_variant_id":"gsp-variant-7","variants":[{"id":"gsp-variant-7","values":[{"attribute_id":"gsp-axis-color","value":"gold"}]}],"price":9223372036854775807,"wallet_toman":"1000","user_id":"native-user"}`)
	b, err := publicCatalogJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	var output map[string]json.RawMessage
	json.Unmarshal(b, &output)
	if string(output["price"]) != "9223372036854775807" || string(output["wallet_toman"]) != `"1000"` || string(output["name"]) != `"gsp-product-12034 is descriptive text"` || string(output["user_id"]) != `"native-user"` {
		t.Fatal(string(b))
	}
	if !strings.Contains(string(output["gallery"]), "/catalog/icons/chr-m-") || strings.Contains(string(output["variants"]), "gsp-") || string(output["selected_variant_id"]) != `"`+backoffice.PublicReference("v", "gsp-variant-7")+`"` {
		t.Fatal(string(b))
	}
}
func TestCatalogInputKindsDuplicatesAndShippingNormalization(t *testing.T) {
	b, err := publicCatalogJSON(map[string]any{"store_id": "s1", "store_ids": []string{"s1"}})
	if err != nil {
		t.Fatal(err)
	}
	var projected struct {
		StoreID  string   `json:"store_id"`
		StoreIDs []string `json:"store_ids"`
	}
	json.Unmarshal(b, &projected)
	if len(projected.StoreIDs) != 1 || projected.StoreID != projected.StoreIDs[0] {
		t.Fatal("shipping restrictions mismatch", string(b))
	}

	alias := backoffice.PublicReference("s", "s1")
	v, err := normalizeReferenceValue(map[string]any{"shipping": map[string]any{alias: "post"}, "offer_id": backoffice.PublicReference("o", "offer"), "wallet_toman": "2500"}, "")
	if err != nil {
		t.Fatal(err)
	}
	m := v.(map[string]any)
	if m["offer_id"] != "offer" || m["shipping"].(map[string]any)["s1"] != "post" || m["wallet_toman"] != "2500" {
		t.Fatal(v)
	}
	if _, err = normalizeReferenceValue(map[string]any{"offer_id": backoffice.PublicReference("v", "offer")}, ""); err == nil {
		t.Fatal("wrong type accepted")
	}
	if _, err = normalizeReferenceValue(map[string]any{"shipping": map[string]any{alias: "post", "s1": "local"}}, ""); err == nil {
		t.Fatal("aliased duplicate accepted")
	}
}
func TestProductQueryAliasesAndMalformedInputs(t *testing.T) {
	v := backoffice.PublicReference("v", strings.Repeat("v", 128))
	o := backoffice.PublicReference("o", "offer")
	r := httptest.NewRequest("GET", "/storefront/products/chr-1?variant="+v+"&offer="+o, nil)
	variant, offer, err := productQuery(r)
	if err != nil || len(variant) != 128 || offer != "offer" {
		t.Fatal(variant, offer, err)
	}
	for _, q := range []string{"variant=chr-v-!", "variant=" + o, "variant=" + v + "&variant=" + v} {
		if _, _, err := productQuery(httptest.NewRequest("GET", "/x?"+q, nil)); err == nil {
			t.Fatal(q)
		}
	}
}
