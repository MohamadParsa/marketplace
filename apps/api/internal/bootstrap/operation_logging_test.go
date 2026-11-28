package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/locallog"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
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

func TestPublicLogPathCannotHideBehindSymlinkAncestor(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "public")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(root, alias); err != nil {
		if runtime.GOOS != "windows" {
			t.Skip("symlink privilege unavailable")
		}
		// Directory junctions need no symlink privilege on Windows. Both paths
		// are owned temporary directories; mklink only creates the alias.
		if out, err := exec.Command("cmd.exe", "/c", "mklink", "/J", alias, root).CombinedOutput(); err != nil {
			t.Fatalf("cannot create test junction: %v: %s", err, out)
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("accepted a public directory through an alias")
		}
	}()
	bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebRoot: root, OperationLogDir: filepath.Join(alias, "new", "logs")}, backofficememory.New(commerceSeed()))
}

func TestRejectedCheckoutCanBeFollowedFromHTTPToDomain(t *testing.T) {
	repo := backofficememory.New(commerceSeed())
	dir := filepath.Join(t.TempDir(), "logs")
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebSurface: "storefront", OperationLogDir: dir}, repo)
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	before, _ := repo.Snapshot(context.Background())
	body, _ := json.Marshal(backoffice.QuoteInput{AddressID: "missing", CouponCode: "PRIVATE-COUPON-09120000001", WalletToman: "1"})
	r := httptest.NewRequest("POST", "/storefront/checkout/preview?private=PRIVATE-QUERY", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+buyer)
	op := "0123456789abcdef0123456789abcdef"
	r.Header.Set("X-Operation-ID", op)
	out := httptest.NewRecorder()
	h.ServeHTTP(out, r)
	if out.Code != 404 {
		t.Fatal(out.Code, out.Body.String())
	}
	closeLogHandler(t, h)
	reader, err := locallog.OpenReader(locallog.DefaultOptions(dir))
	if err != nil {
		t.Fatal(err)
	}
	page, err := operationlog.Read(context.Background(), reader, operationlog.Query{OperationID: op, Limit: 200})
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]operationlog.Event{}
	for _, e := range page.Items {
		found[e.EventName] = e
		if e.RequestID != out.Header().Get("X-Request-ID") {
			t.Fatal("request correlation broken")
		}
	}
	if len(page.Items) != 3 || found["http.request-started"].Outcome != "started" || found["checkout.preview"].Outcome != "rejected" || found["checkout.preview"].ErrorCode != "not_found" || found["http.request-finished"].Outcome != "rejected" || found["checkout.preview"].ActorID == "" {
		t.Fatal("missing rejection chain", found)
	}
	encoded, _ := json.Marshal(page)
	if strings.Contains(string(encoded), "PRIVATE") || strings.Contains(string(encoded), buyer) {
		t.Fatal("private checkout data in logs")
	}
	after, _ := repo.Snapshot(context.Background())
	if len(after.Commerce.Purchases) != len(before.Commerce.Purchases) || len(after.Commerce.Entries) != len(before.Commerce.Entries) || after.Offers[0].Stock != before.Offers[0].Stock {
		t.Fatal("rejected checkout changed business state")
	}
}
