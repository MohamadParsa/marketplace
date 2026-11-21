package bootstrap_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/application/customer"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

func requestAs(t *testing.T, handler http.Handler, token, method, path, body string, want int) []byte {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("%s %s: got %d, want %d (%s)", method, path, response.Code, want, response.Body)
	}
	return response.Body.Bytes()
}

func readJSON[T any](t *testing.T, data []byte) T {
	t.Helper()
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestUnifiedCustomerAccountAndPurchaseOwnership(t *testing.T) {
	handler := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	panel := loginAt(t, handler, "backoffice", "09000000002")
	seller := loginAt(t, handler, "seller", "09000000002")
	buyer := loginAt(t, handler, "storefront", "09000000002")
	otherBuyer := loginAt(t, handler, "storefront", "09000000005")
	buyerToken := buyer["token"].(string)
	userID := panel["user"].(map[string]any)["ID"]
	profile := readJSON[customer.Profile](t, requestAs(t, handler, buyerToken, "GET", "/storefront/account/profile", "", 200))
	if profile.UserID != userID || profile.UserID != seller["user"].(map[string]any)["ID"] || profile.UserID != buyer["user"].(map[string]any)["ID"] {
		t.Fatalf("expected one identity across surfaces: %v / %+v", userID, profile)
	}
	purchases := readJSON[[]customer.Purchase](t, requestAs(t, handler, buyerToken, "GET", "/storefront/account/purchases?user_id=viewer", "", 200))
	if len(purchases) != 1 || purchases[0].ID != "BUY-501" || len(purchases[0].Orders) != 2 {
		t.Fatalf("expected own purchase in two stores: %+v", purchases)
	}
	other := readJSON[[]customer.Purchase](t, requestAs(t, handler, otherBuyer["token"].(string), "GET", "/storefront/account/purchases", "", 200))
	if len(other) != 1 || other[0].ID != "BUY-502" {
		t.Fatalf("wrong customer's purchases: %+v", other)
	}
	requestAs(t, handler, buyerToken, "GET", "/backoffice/workspace", "", 401)
	requestAs(t, handler, seller["token"].(string), "GET", "/storefront/account/profile", "", 401)
}

func TestCustomRolesCombineWithoutGrantingAccessManagement(t *testing.T) {
	fixture := backofficememory.Fixture()
	for i := range fixture.Categories {
		if fixture.Categories[i].ID == "c2" {
			fixture.Categories[i].Published = true
		}
	}
	for i := range fixture.Products {
		if fixture.Products[i].ID == "p2" {
			fixture.Products[i].ImageID = "publish-ready-photo"
		}
	}
	fixture.IconMedia = append(fixture.IconMedia, catalog.IconMedia{ID: "publish-ready-photo", Data: iconPNG(t, 8)})
	handler := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, backofficememory.New(fixture))
	admin := login(t, handler, "09000000001")
	commands := "/backoffice/commands"
	requestAs(t, handler, admin, "POST", commands, `{"action":"role.create","name":"read","scope":"global","permissions":["workspace.read"],"idempotency_key":"read"}`, 204)
	requestAs(t, handler, admin, "POST", commands, `{"action":"role.create","name":"publish","scope":"global","permissions":["product.publish"],"idempotency_key":"publish"}`, 204)
	requestAs(t, handler, admin, "POST", commands, `{"action":"role.create","name":"escalation","scope":"global","permissions":["user.role"],"idempotency_key":"unknown"}`, 400)
	workspace := readJSON[backoffice.Workspace](t, requestAs(t, handler, admin, "GET", "/backoffice/workspace", "", 200))
	ids := make(map[string]string)
	for _, role := range workspace.Access.Roles {
		ids[role.Name] = role.ID
	}
	roleCommand := func(command backoffice.Command, status int) {
		t.Helper()
		body, err := json.Marshal(command)
		if err != nil {
			t.Fatal(err)
		}
		requestAs(t, handler, admin, "POST", commands, string(body), status)
	}
	setRole := func(action, user, label, key string) {
		t.Helper()
		roleCommand(backoffice.Command{Action: action, Target: user, Value: ids[label], IdempotencyKey: key}, 204)
	}
	setRole("user.role-grant", "seller-test", "read", "grant-read")
	setRole("user.role-grant", "viewer", "read", "grant-read-second")
	setRole("user.role-grant", "viewer", "publish", "grant-publish")
	edit := backoffice.Command{Action: "role.update", Target: ids["read"], Name: "read-updated", Permissions: []string{"workspace.read"}, Version: 1, IdempotencyKey: "edit"}
	roleCommand(edit, 204)
	edit.IdempotencyKey = "stale"
	roleCommand(edit, 409)
	removeRole := backoffice.Command{Action: "role.delete", Target: ids["publish"], Version: 1, IdempotencyKey: "remove-publish"}
	roleCommand(removeRole, 409)
	readOnly := login(t, handler, "09120000002")
	combined := login(t, handler, "09000000005")
	const publish = `{"action":"product.publish","target":"p2","version":1,"idempotency_key":"publish-p2"}`
	readState := readJSON[backoffice.Workspace](t, requestAs(t, handler, readOnly, "GET", "/backoffice/workspace", "", 200))
	if readState.Access != nil || len(readState.Users) > 0 || len(readState.Audit) > 0 {
		t.Fatal("read-only role received private access data")
	}
	requestAs(t, handler, readOnly, "POST", commands, publish, 403)
	requestAs(t, handler, combined, "POST", commands, publish, 204)
	requestAs(t, handler, combined, "POST", commands, `{"action":"role.create","name":"attack","scope":"global","permissions":["workspace.read"],"idempotency_key":"non-admin"}`, 403)
	setRole("user.role-revoke", "viewer", "publish", "revoke-publish")
	requestAs(t, handler, combined, "POST", commands, publish, 401)
	roleCommand(removeRole, 204)

	seller := loginAt(t, handler, "seller", "09000000002")["token"].(string)
	requestAs(t, handler, admin, "POST", commands, `{"action":"membership.remove","target":"ops","store_id":"s2","idempotency_key":"remove-s2"}`, 204)
	requestAs(t, handler, seller, "GET", "/seller/stores/s2/offers", "", 403)
	requestAs(t, handler, seller, "GET", "/seller/stores/s1/offers", "", 200)
}
