package bootstrap_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

func attributeCommand(t *testing.T, h http.Handler, token string, c backoffice.Command, want int) {
	t.Helper()
	data, _ := json.Marshal(c)
	r := httptest.NewRequest("POST", "/backoffice/commands", bytes.NewReader(data))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != want {
		t.Fatalf("%s got %d want %d: %s", c.Action, w.Code, want, w.Body.String())
	}
}

func TestAttributesHTTPCRUDAndPermissions(t *testing.T) {
	h := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	admin := login(t, h, "09000000001")
	ops := login(t, h, "09000000002")
	viewer := login(t, h, "09000000005")
	cmd := backoffice.Command{Action: "attribute.create", Attribute: &catalog.AttributeDefinition{Code: "material", Name: "جنس آزمون", Kind: "enum", Options: []string{"فلز", "پلاستیک"}}, IdempotencyKey: "create-material"}
	attributeCommand(t, h, "", cmd, 401)
	attributeCommand(t, h, viewer, cmd, 403)
	attributeCommand(t, h, ops, cmd, 204)
	attributeCommand(t, h, ops, cmd, 204)
	state := iconWorkspace(t, h, ops)
	var a catalog.AttributeDefinition
	for _, x := range state.Attributes {
		if x.Code == backoffice.PublicReference("a", "material") {
			a = x
		}
	}
	if a.ID == "" || a.Version != 1 {
		t.Fatal("missing shared definition")
	}
	duplicate := cmd
	duplicate.Target = a.ID
	duplicate.IdempotencyKey = "duplicate"
	attributeCommand(t, h, ops, duplicate, 409)
	a.Name = "جنس و مواد آزمون"
	update := backoffice.Command{Action: "attribute.update", Target: a.ID, Version: 1, Attribute: &a, IdempotencyKey: "update"}
	attributeCommand(t, h, ops, update, 204)
	update.IdempotencyKey = "stale"
	attributeCommand(t, h, ops, update, 409)
	remove := backoffice.Command{Action: "attribute.delete", Target: a.ID, Version: 2, IdempotencyKey: "delete"}
	attributeCommand(t, h, ops, remove, 204)
	attributeCommand(t, h, ops, remove, 204)
	for _, x := range iconWorkspace(t, h, ops).Attributes {
		if x.ID == a.ID {
			t.Fatal("delete failed")
		}
	}
	settings := backoffice.Command{Action: "settings.card-limit", Target: "display", Amount: 4, Version: 1, IdempotencyKey: "settings"}
	attributeCommand(t, h, ops, settings, 403)
	attributeCommand(t, h, viewer, settings, 403)
	attributeCommand(t, h, admin, settings, 204)
	settings.IdempotencyKey = "settings-stale"
	attributeCommand(t, h, admin, settings, 409)
	result := iconWorkspace(t, h, viewer)
	if result.DisplaySettings.CardAttributeLimit != 4 {
		t.Fatal("viewer did not receive public display value")
	}
	found := false
	for _, p := range result.CardPreviews {
		if p.ProductID == "mobile-sample" {
			found = true
			if len(p.Features) != 4 {
				t.Fatal("cap not consumed")
			}
		}
	}
	if !found {
		t.Fatal("missing preview")
	}
}
