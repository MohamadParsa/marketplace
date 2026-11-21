package kavenegar

import (
	"context"
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
