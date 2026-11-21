package bootstrap_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func TestAdminGrantRequiresReauthenticationAndRevocationTakesEffect(t *testing.T) {
	handler := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	admin := login(t, handler, "09000000001")
	operations := login(t, handler, "09000000002")
	call := func(token, method, path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("%s %s: got %d, want %d (%s)", method, path, response.Code, want, response.Body)
		}
		return response
	}
	const commands = "/backoffice/commands"
	const challengePath = "/backoffice/auth/request-admin-otp"
	const roleGrant = `{"action":"user.role","target":"viewer","value":"admin","idempotency_key":"grant"}`
	const demoteAdmin = `{"action":"user.role","target":"admin","value":"viewer","idempotency_key":"demote"}`
	call(operations, "POST", commands, roleGrant, http.StatusForbidden)
	call(operations, "POST", challengePath, "", http.StatusForbidden)
	call(admin, "POST", commands, demoteAdmin, http.StatusConflict)
	call(admin, "POST", commands, `{"action":"user.create","name":"Admin test","phone":"09000000009","value":"super-admin","idempotency_key":"create"}`, http.StatusForbidden)
	missing := call(admin, "POST", commands, roleGrant, http.StatusForbidden)
	if !strings.Contains(missing.Body.String(), "reauthentication_required") {
		t.Fatal(missing.Body.String())
	}
	challenge := call(admin, "POST", challengePath, "", http.StatusAccepted)
	var otp struct {
		RequestID string `json:"request_id"`
		Code      string `json:"test_code"`
	}
	if err := json.Unmarshal(challenge.Body.Bytes(), &otp); err != nil || otp.RequestID == "" || otp.Code == "" {
		t.Fatal("missing synthetic admin challenge")
	}
	verification, err := json.Marshal(map[string]string{"request_id": otp.RequestID, "code": otp.Code})
	if err != nil {
		t.Fatal(err)
	}
	call("", "POST", "/backoffice/auth/verify-otp", string(verification), http.StatusUnauthorized)
	withOTP := func(command string) string {
		return strings.TrimSuffix(command, "}") + `,"reauthentication":` + string(verification) + "}"
	}
	call(admin, "POST", commands, withOTP(roleGrant), http.StatusNoContent)
	call(admin, "POST", commands, withOTP(roleGrant), http.StatusNoContent)
	call(admin, "POST", commands, withOTP(`{"action":"user.role","target":"ops","value":"admin","idempotency_key":"replay"}`), http.StatusForbidden)
	newAdmin := login(t, handler, "09000000005")
	otherChallenge := call(newAdmin, "POST", challengePath, "", http.StatusAccepted)
	if err := json.Unmarshal(otherChallenge.Body.Bytes(), &otp); err != nil {
		t.Fatal(err)
	}
	verification, err = json.Marshal(map[string]string{"request_id": otp.RequestID, "code": otp.Code})
	if err != nil {
		t.Fatal(err)
	}
	call(admin, "POST", commands, withOTP(`{"action":"user.role","target":"ops","value":"admin","idempotency_key":"wrong-session"}`), http.StatusForbidden)
	call(newAdmin, "POST", commands, demoteAdmin, http.StatusNoContent)
	call(admin, "POST", commands, roleGrant, http.StatusUnauthorized)

	response := call(newAdmin, "GET", "/backoffice/workspace", "", http.StatusOK)
	var state backoffice.Workspace
	if err := json.Unmarshal(response.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if len(state.Audit) != 2 {
		t.Fatalf("expected one grant and one revoke, got %d audit entries", len(state.Audit))
	}
	grant, revoke := state.Audit[0], state.Audit[1]
	if grant.Actor != "admin" || grant.Target != "viewer" || grant.PreviousRole != identity.RoleViewer || grant.Role != identity.RoleSuperAdmin ||
		revoke.Actor != "viewer" || revoke.Target != "admin" || revoke.PreviousRole != identity.RoleSuperAdmin || revoke.Role != identity.RoleViewer {
		t.Fatalf("unexpected role audit: %+v %+v", grant, revoke)
	}
	if strings.Contains(response.Body.String(), otp.RequestID) {
		t.Fatal("challenge identifier leaked in workspace")
	}
}
