package backoffice_test

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"testing"
)

func TestGatewayCallbackUsesEnvironmentOriginAndCodeOwnedPath(t *testing.T) {
	for _, tc := range []struct {
		name, base string
		local      bool
	}{
		{"local", "http://localhost:5197", true},
		{"test", "http://127.0.0.1:18105", true},
		{"staging", "https://stage.example.test", false},
		{"production", "https://shop.example.test", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newCommerceHarness(t)
			if err := h.s.ConfigureGatewayCallbackBase(tc.base+"/", tc.local); err != nil {
				t.Fatal(err)
			}
			got, err := h.s.GatewayCallbackURL("zarinpal")
			if err != nil || got != tc.base+backoffice.ZarinpalCallbackPath {
				t.Fatal(got, err)
			}
		})
	}
	h := newCommerceHarness(t)
	if _, err := h.s.GatewayCallbackURL("zarinpal"); err != backoffice.ErrGatewayCallbackUnavailable {
		t.Fatal(err)
	}
	for _, bad := range []string{"http://example.test", "https://user:pass@example.test", "https://example.test?callback=elsewhere", "https://example.test/path", "//example.test", "javascript:alert(1)"} {
		if err := h.s.ConfigureGatewayCallbackBase(bad, false); err != backoffice.ErrGatewayCallbackUnavailable {
			t.Fatal("unsafe origin", bad, err)
		}
	}
	if err := h.s.ConfigureGatewayCallbackBase("http://localhost:5197", false); err != backoffice.ErrGatewayCallbackUnavailable {
		t.Fatal("production HTTP allowed", err)
	}
	for _, provider := range backoffice.GatewayProviders() {
		for _, field := range provider.Fields {
			if field.Key == "callback_url" {
				t.Fatal("callback exposed as a gateway setting")
			}
		}
	}
}
