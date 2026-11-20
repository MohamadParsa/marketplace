package bootstrap_test

import (
	"bytes"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func login(t *testing.T, handler http.Handler, phone string) string {
	t.Helper()
	return loginAt(t, handler, "backoffice", phone)["token"].(string)
}

func loginAt(t *testing.T, handler http.Handler, surface, phone string) map[string]any {
	t.Helper()
	call := func(path string, body any) map[string]any {
		encoded, _ := json.Marshal(body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("POST", path, bytes.NewReader(encoded)))
		if response.Code != 200 && response.Code != 202 {
			t.Fatalf("login %s: %d", path, response.Code)
		}
		var result map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	challenge := call("/"+surface+"/auth/request-otp", map[string]string{"phone": phone})
	return call("/"+surface+"/auth/verify-otp", map[string]any{"request_id": challenge["request_id"], "code": challenge["test_code"]})
}
func TestBackofficeHTTPPermissions(t *testing.T) {
	handler := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	viewer := login(t, handler, "09000000005")
	request := httptest.NewRequest("GET", "/backoffice/workspace", nil)
	request.Header.Set("Authorization", "Bearer "+viewer)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatal(response.Code)
	}
	var state map[string]any
	_ = json.Unmarshal(response.Body.Bytes(), &state)
	if _, ok := state["users"]; ok {
		t.Fatal("users leaked")
	}
	if _, ok := state["transfers"]; ok {
		t.Fatal("financial data leaked")
	}
	body := bytes.NewBufferString(`{"action":"store.toggle","target":"s1","idempotency_key":"forbidden"}`)
	request = httptest.NewRequest("POST", "/backoffice/commands", body)
	request.Header.Set("Authorization", "Bearer "+viewer)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 403 {
		t.Fatalf("viewer write = %d", response.Code)
	}
}
func TestDeactivationRevokesAuthorization(t *testing.T) {
	handler := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	admin := login(t, handler, "09000000001")
	viewer := login(t, handler, "09000000005")
	request := httptest.NewRequest("POST", "/backoffice/commands", bytes.NewBufferString(`{"action":"user.toggle","target":"viewer","reason":"test","idempotency_key":"deactivate"}`))
	request.Header.Set("Authorization", "Bearer "+admin)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 204 {
		t.Fatalf("deactivate: %d", response.Code)
	}
	request = httptest.NewRequest("GET", "/backoffice/workspace", nil)
	request.Header.Set("Authorization", "Bearer "+viewer)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 401 {
		t.Fatalf("deactivated session access: %d", response.Code)
	}
}
