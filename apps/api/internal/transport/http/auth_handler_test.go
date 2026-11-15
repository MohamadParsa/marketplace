package httptransport

import (
	"bytes"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
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
	var challenge map[string]string
	_ = json.Unmarshal(response.Body.Bytes(), &challenge)
	verifyBody, _ := json.Marshal(map[string]string{"request_id": challenge["request_id"], "code": challenge["test_code"]})
	verify := httptest.NewRequest(http.MethodPost, "/backoffice/auth/verify-otp", bytes.NewReader(verifyBody))
	verifyResponse := httptest.NewRecorder()
	router.ServeHTTP(verifyResponse, verify)
	if verifyResponse.Code != http.StatusOK {
		t.Fatalf("verify status=%d", verifyResponse.Code)
	}
}
