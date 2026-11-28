package kavenegar

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotifierSendsLookupFormWithoutLeakingSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/test-key/verify/lookup.json" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("receptor") != "09120000001" || r.Form.Get("token") != "123456" || r.Form.Get("template") != "login" {
			t.Fatalf("form=%v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"تایید شد"}}`))
	}))
	defer server.Close()
	n := New("test-key", "login", server.URL, server.Client())
	if err := n.Send(context.Background(), identity.SurfaceCustomer, "+989120000001", "123456"); err != nil {
		t.Fatal(err)
	}
}

func TestNotifierRejectsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"return":{"status":424,"message":"template"}}`))
	}))
	defer server.Close()
	err := New("key", "missing", server.URL, server.Client()).Send(context.Background(), identity.SurfaceCustomer, "0912", "123456")
	if err == nil || !strings.Contains(err.Error(), "provider") {
		t.Fatalf("err=%v", err)
	}
}

func TestProviderCallsHaveCorrelatedSanitizedSemanticOutcome(t *testing.T) {
	var events []operationlog.Event
	recorder := operationlog.New(operationlog.SinkFunc(func(e operationlog.Event) error { events = append(events, e); return nil }), "api", "test", "storefront")
	ctx := operationlog.WithRecorder(operationlog.WithContext(context.Background(), operationlog.ContextData{OperationID: "operation-1", RequestID: "request-1"}), recorder)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"return":{"status":424,"message":"PRIVATE-RESPONSE"}}`))
	}))
	defer server.Close()
	if err := New("PRIVATE-KEY", "login", server.URL, server.Client()).Send(ctx, identity.SurfaceCustomer, "09120000001", "PRIVATE-OTP"); err == nil {
		t.Fatal("accepted provider rejection")
	}
	if len(events) != 2 || events[0].Outcome != "started" || events[1].Outcome != "failed" || events[1].HTTPStatus != 200 || events[1].ErrorCode != "provider.424" || events[1].OperationID != "operation-1" {
		t.Fatal(events)
	}
	b, _ := json.Marshal(events)
	for _, v := range []string{"PRIVATE-KEY", "PRIVATE-OTP", "PRIVATE-RESPONSE", "09120000001"} {
		if strings.Contains(string(b), v) {
			t.Fatal("private provider data logged")
		}
	}
}
