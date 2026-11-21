package bootstrap_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
)

func productCall(t *testing.T, h http.Handler, method, path, token string, body any, want int) *httptest.ResponseRecorder {
	t.Helper()
	var data []byte
	if raw, ok := body.(string); ok {
		data = []byte(raw)
	} else if body != nil {
		data, _ = json.Marshal(body)
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(data))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != want {
		t.Fatalf("%s %s got %d want %d: %s", method, path, w.Code, want, w.Body.String())
	}
	return w
}

func sellerLogin(t *testing.T, h http.Handler, phone string) string {
	t.Helper()
	var challenge struct {
		RequestID string `json:"request_id"`
		TestCode  string `json:"test_code"`
	}
	w := productCall(t, h, "POST", "/seller/auth/request-otp", "", map[string]string{"phone": phone}, 202)
	if err := json.Unmarshal(w.Body.Bytes(), &challenge); err != nil {
		t.Fatal(err)
	}
	w = productCall(t, h, "POST", "/seller/auth/verify-otp", "", map[string]string{"request_id": challenge.RequestID, "code": challenge.TestCode}, 200)
	var session struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &session)
	return session.Token
}

func TestProductWorkspaceHTTPReadWriteResolveAndPrivacy(t *testing.T) {
	h := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := sellerLogin(t, h, "09120000012")
	sellerB := sellerLogin(t, h, "09120000013")
	path := "/backoffice/products/mobile-sample/workspace"
	productCall(t, h, "GET", path, "", nil, 401)
	productCall(t, h, "GET", path, seller, nil, 401)
	var w backoffice.ProductWorkspace
	r := productCall(t, h, "GET", path, viewer, nil, 200)
	_ = json.Unmarshal(r.Body.Bytes(), &w)
	if w.Permissions.CreateVariant || len(w.Permissions.PriceStores) > 0 || len(w.Variants) != 2 || len(w.Offers) != 4 {
		t.Fatal(w)
	}
	productCall(t, h, "GET", "/seller/products/mobile-sample/workspace", admin, nil, 401)
	r = productCall(t, h, "GET", "/seller/products/mobile-sample/workspace", seller, nil, 200)
	_ = json.Unmarshal(r.Body.Bytes(), &w)
	if strings.Contains(r.Body.String(), "mob-b-black") || len(w.Offers) != 2 || w.Permissions.CreateVariant {
		t.Fatal("seller leaked other offer")
	}
	workspace := productCall(t, h, "GET", "/seller/workspace", seller, nil, 200)
	for _, forbidden := range []string{`"users"`, `"orders"`, `"audit"`, `"transfers"`, `"memberships"`, `"mob-b-black"`, `"p2"`} {
		if strings.Contains(workspace.Body.String(), forbidden) {
			t.Fatal("private data leaked", forbidden)
		}
	}
	productCall(t, h, "GET", "/seller/products/p2/workspace", seller, nil, 403)
	productCall(t, h, "GET", "/seller/products/p2/workspace", sellerB, nil, 200)
	price := `{"action":"offer.price.update","target":"mob-a-black","version":1,"idempotency_key":"seller-price","offer":{"price":123450000}}`
	productCall(t, h, "POST", "/seller/commands", seller, price, 204)
	productCall(t, h, "POST", "/seller/commands", seller, price, 204)
	productCall(t, h, "POST", "/seller/commands", sellerB, price, 403)
	productCall(t, h, "POST", "/backoffice/commands", viewer, price, 403)
	productCall(t, h, "POST", "/seller/commands", seller, `{"action":"offer.stock.update","target":"mob-a-black","version":2,"idempotency_key":"low","offer":{"stock":1}}`, 409)
	productCall(t, h, "POST", "/seller/commands", seller, `{"action":"offer.stock.update","target":"mob-a-black","version":2,"idempotency_key":"reserved-injection","offer":{"stock":8,"reserved":0}}`, 400)
	productCall(t, h, "POST", "/seller/commands", seller, `{"action":"offer.stock.update","target":"mob-a-black","version":2,"idempotency_key":"lead-injection","offer":{"stock":8,"lead_days":0}}`, 400)
	productCall(t, h, "POST", "/seller/commands", seller, `{"action":"product.publish","target":"mobile-sample","version":1,"idempotency_key":"seller-content"}`, 403)
	productCall(t, h, "POST", "/seller/commands", seller, `{"action":"offer.stock.update","target":"mob-a-black","version":2,"idempotency_key":"stock-ok","offer":{"stock":9}}`, 204)
	resolve := "/backoffice/products/mobile-sample/resolve?attr.color=" + url.QueryEscape("مشکی") + "&attr.warranty=" + url.QueryEscape("۱۸ ماهه")
	resolved := productCall(t, h, "GET", resolve, admin, nil, 200)
	if !strings.Contains(resolved.Body.String(), "mobile-black-18") || !strings.Contains(resolved.Body.String(), "mob-a-black") || !strings.Contains(resolved.Body.String(), "mob-b-black") {
		t.Fatal(resolved.Body.String())
	}
	if resolved.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("private resolve cached")
	}
	productCall(t, h, "GET", resolve+"&attr.color=duplicate", admin, nil, 400)
	productCall(t, h, "GET", "/backoffice/products/mobile-sample/resolve?attr.color=x", admin, nil, 400)
	productCall(t, h, "GET", "/backoffice/products/missing/workspace", admin, nil, 404)
}
