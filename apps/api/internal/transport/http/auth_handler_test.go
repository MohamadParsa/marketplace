package httptransport

import (
	"bytes"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRouter() http.Handler {
	store := authmemory.NewStore([]identity.User{{ID: "u1", Phone: "09120000001", Surface: identity.SurfaceBackoffice, Role: identity.RoleViewer, Active: true}})
	notifier := authmemory.NewNotifier()
	service := auth.NewService(store, store, store, notifier, authmemory.Clock{}, authmemory.Random{})
	return NewRouter(RouterConfig{Auth: NewAuthHandler(service, notifier, true)})
}

func TestOTPHTTPFlow(t *testing.T) {
	router := testRouter()
	requestBody, _ := json.Marshal(map[string]string{"phone": "09120000001"})
	request := httptest.NewRequest(http.MethodPost, "/backoffice/auth/request-otp", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("request status=%d body=%s", response.Code, response.Body.String())
	}
	var challenge map[string]any
	_ = json.Unmarshal(response.Body.Bytes(), &challenge)
	verifyBody, _ := json.Marshal(map[string]string{"request_id": challenge["request_id"].(string), "code": challenge["test_code"].(string)})
	verify := httptest.NewRequest(http.MethodPost, "/backoffice/auth/verify-otp", bytes.NewReader(verifyBody))
	verifyResponse := httptest.NewRecorder()
	router.ServeHTTP(verifyResponse, verify)
	if verifyResponse.Code != http.StatusOK {
		t.Fatalf("verify status=%d", verifyResponse.Code)
	}
}

func TestPasswordLoginHTTPFlow(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("test-password-123"), bcrypt.MinCost)
	store := authmemory.NewStore([]identity.User{{ID: "u1", Phone: "09120000001", Surface: identity.SurfaceBackoffice, Role: identity.RoleViewer, Active: true, PasswordHash: string(hash)}})
	notifier := authmemory.NewNotifier()
	service := auth.NewService(store, store, store, notifier, authmemory.Clock{}, authmemory.Random{})
	router := NewRouter(RouterConfig{Auth: NewAuthHandler(service, notifier, true)})
	body, _ := json.Marshal(map[string]string{"phone": "09120000001", "password": "test-password-123"})
	request := httptest.NewRequest(http.MethodPost, "/backoffice/auth/password-login", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), hash) {
		t.Fatal("password hash leaked in response")
	}
}
