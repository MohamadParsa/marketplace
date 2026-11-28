package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/locallog"
	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
)

func closeLogHandler(t *testing.T, h http.Handler) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.(interface{ Close(context.Context) error }).Close(ctx); err != nil {
		t.Fatal(err)
	}
}
func TestOperationHTTPContextReplayAndPrivateBodies(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "private-logs")
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebSurface: "storefront", OperationLogDir: dir}, backofficememory.New(commerceSeed()))
	var guest map[string]string
	json.Unmarshal(commerceHTTP(t, h, "", "", "POST", "/storefront/cart/guest", nil, "", 200), &guest)
	input := map[string]any{"action": "add", "offer_id": "offer", "quantity": 1, "version": 1}
	op := "0123456789abcdef0123456789abcdef"
	for i := 0; i < 2; i++ {
		b, _ := json.Marshal(input)
		r := httptest.NewRequest("POST", "/storefront/cart?token=PRIVATE-QUERY", bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Cart-Token", guest["cart_token"])
		r.Header.Set("Idempotency-Key", "cart-add")
		r.Header.Set("X-Operation-ID", op)
		r.Header.Set("Cookie", "PRIVATE-COOKIE")
		r.Header.Set("X-Request-ID", "PRIVATE-HEADER")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)
		if rr.Code != 200 || rr.Header().Get("X-Operation-ID") != op || len(rr.Header().Get("X-Request-ID")) != 32 {
			t.Fatal(rr.Code, rr.Body.String())
		}
	}
	commerceHTTP(t, h, "", guest["cart_token"], "POST", "/storefront/cart", map[string]any{"password": "PRIVATE-BODY"}, "bad-input", 400)
	commerceHTTP(t, h, "invalid-secret-token", "", "POST", "/storefront/checkout/purchases", map[string]any{}, "invalid-auth", 401)
	closeLogHandler(t, h)
	reader, err := locallog.OpenReader(locallog.DefaultOptions(dir))
	if err != nil {
		t.Fatal(err)
	}
	page, err := operationlog.Read(context.Background(), reader, operationlog.Query{Limit: 200})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(page)
	for _, secret := range []string{"PRIVATE-BODY", "PRIVATE-QUERY", "PRIVATE-COOKIE", "PRIVATE-HEADER", "invalid-secret-token", guest["cart_token"]} {
		if strings.Contains(string(encoded), secret) {
			t.Fatal("private data in events")
		}
	}
	finished, replayed := 0, 0
	for _, e := range page.Items {
		if e.OperationID == op && e.EventName == "commerce.cart" && e.Outcome == "succeeded" {
			finished++
		}
		if e.OperationID == op && e.EventName == "commerce.cart" && e.Outcome == "replayed" {
			replayed++
		}
	}
	if finished != 1 || replayed != 1 {
		t.Fatalf("missing outcome/replay: %d %d", finished, replayed)
	}
}
func TestUnavailableLocalLogPathDoesNotBlockCart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", OperationLogDir: filepath.Join(path, "logs")}, backofficememory.New(commerceSeed()))
	commerceHTTP(t, h, "", "", "POST", "/storefront/cart/guest", nil, "", 200)
	closeLogHandler(t, h)
}
func TestLogsCannotBeConfiguredInsidePublicFiles(t *testing.T) {
	root := t.TempDir()
	defer func() {
		if recover() == nil {
			t.Fatal("accepted public log directory")
		}
	}()
	bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebRoot: root, OperationLogDir: filepath.Join(root, "logs")}, backofficememory.New(commerceSeed()))
}
