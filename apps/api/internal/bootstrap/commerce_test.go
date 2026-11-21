package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

func commerceSeed() backoffice.Workspace {
	w := backofficememory.Fixture()
	w.Orders = nil
	w.Purchases = nil
	w.Cases = nil
	w.Ledger = nil
	w.Transfers = nil
	w.Categories = []catalog.Category{{ID: "regular", Name: "دیجیتال", Published: true}, {ID: "gold", Name: "طلا", Published: true}}
	w.Products = []catalog.Product{{ID: "product", Name: "محصول آزمون", CategoryID: "regular", Status: "published", ImageID: "photo", Version: 1}}
	w.IconMedia = []catalog.IconMedia{{ID: "photo", Data: []byte{1}}}
	w.Variants = []catalog.Variant{{ID: "variant", ProductID: "product", Version: 1}}
	w.Offers = []catalog.Offer{{ID: "offer", ProductID: "product", VariantID: "variant", StoreID: "s1", Price: 1000000, Stock: 10, Version: 1}}
	settings := commerce.DefaultSettings()
	settings.Destination = commerce.Destination{ID: "test-card", CardNumber: "4242424242424242", Holder: "آزمون", Bank: "آزمایشی"}
	for i := range settings.Methods {
		settings.Methods[i].Active = true
		if settings.Methods[i].Kind == "local" {
			settings.Methods[i].Instructions = "حضوری آزمون"
			settings.Methods[i].ExcludedCategories = []string{"gold"}
		}
	}
	settings.Shipping = []commerce.ShippingMethod{{ID: "post", Name: "پست", Active: true, AllCities: true, Fee: 100000, MinDays: 1, MaxDays: 3}}
	w.Commerce = &commerce.State{Settings: settings}
	return w
}
func commerceHTTP(t *testing.T, h http.Handler, token, guest, method, path string, input any, key string, status int) []byte {
	t.Helper()
	var body []byte
	if input != nil {
		body, _ = json.Marshal(input)
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if guest != "" {
		r.Header.Set("X-Cart-Token", guest)
	}
	if key != "" {
		r.Header.Set("Idempotency-Key", key)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	if rr.Code != status {
		t.Fatalf("%s %s status%d expected%d: %s", method, path, rr.Code, status, rr.Body.String())
	}
	if strings.Contains(path, "/wallet") || strings.Contains(path, "/cart") || strings.Contains(path, "/account/orders") {
		if rr.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("private response cache")
		}
	}
	return rr.Body.Bytes()
}
func TestCommerceHTTPRegistrationCreditAndPayment(t *testing.T) {
	repo := backofficememory.New(commerceSeed())
	commerceRegistration(t, repo)
}
func commerceRegistration(t *testing.T, repo backoffice.Repository) {
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	customer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	admin := login(t, h, "09000000001")
	other := loginAt(t, h, "storefront", "09000000004")["token"].(string)
	get := func(token, method, path string, in any, key string, status int) []byte {
		return commerceHTTP(t, h, token, "", method, path, in, key, status)
	}
	var guest map[string]string
	json.Unmarshal(commerceHTTP(t, h, "", "", "POST", "/storefront/cart/guest", map[string]any{}, "", 200), &guest)
	commerceHTTP(t, h, "", guest["cart_token"], "POST", "/storefront/cart", map[string]any{"action": "add", "offer_id": "offer", "quantity": 1, "version": 1}, "add", 200)
	get(customer, "POST", "/storefront/cart/merge", map[string]any{"guest_token": guest["cart_token"]}, "merge", 200)
	var cart backoffice.CartView
	json.Unmarshal(get(customer, "GET", "/storefront/cart", nil, "", 200), &cart)
	var address commerce.Address
	json.Unmarshal(get(customer, "POST", "/storefront/account/addresses", backoffice.AddressInput{Recipient: "گیرنده ساختگی", Phone: "09000000003", Province: "تهران", City: "تهران", Street: "نشانی آزمون", PostalCode: "1234567890"}, "address", 200), &address)
	get(other, "POST", "/storefront/account/addresses", backoffice.AddressInput{ID: address.ID, Version: 1, Delete: true}, "foreign", 404)
	in := backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"s1": "post"}, PaymentMethodID: "card-transfer"}
	get(customer, "POST", "/storefront/checkout/preview", in, "", 200)
	var quote backoffice.QuoteView
	json.Unmarshal(get(customer, "POST", "/storefront/checkout/quotes", in, "quote", 200), &quote)
	get(customer, "POST", "/storefront/checkout/purchases", backoffice.AcceptQuoteInput{QuoteID: quote.ID, AcceptedTotalToman: "1"}, "tamper", 409)
	var p backoffice.PurchaseView
	json.Unmarshal(get(customer, "POST", "/storefront/checkout/purchases", backoffice.AcceptQuoteInput{QuoteID: quote.ID, AcceptedTotalToman: quote.TotalToman}, "accept", 200), &p)
	if p.PaymentStatus != "unpaid" || p.OrderStatus != "awaiting-payment" {
		t.Fatal(p)
	}
	get(other, "GET", "/storefront/account/orders/"+p.ID, nil, "", 404)
	var d backoffice.DepositView
	json.Unmarshal(get(customer, "POST", "/storefront/wallet/deposits", backoffice.DepositInput{MethodID: "card-transfer", PurchaseID: p.ID, DestinationID: "test-card", AmountToman: quote.TotalToman, Date: time.Now().UTC().Format("2006-01-02"), Reference: "HTTP-001"}, "deposit", 200), &d)
	var before backoffice.WalletView
	json.Unmarshal(get(customer, "GET", "/storefront/wallet", nil, "", 200), &before)
	if before.BalanceToman != "0" {
		t.Fatal("declaration credited")
	}
	review := backoffice.DepositReviewInput{ID: d.ID, Version: d.Version, Approve: true, VerifiedReference: "ACTUAL-001", VerifiedDate: d.Date, VerifiedAmountToman: d.AmountToman}
	get(customer, "POST", "/backoffice/commerce/deposits/review", review, "bad", 401)
	get(admin, "POST", "/backoffice/commerce/deposits/review", review, "review", 200)
	get(admin, "POST", "/backoffice/commerce/deposits/review", review, "review", 200)
	json.Unmarshal(get(customer, "GET", "/storefront/account/orders/"+p.ID, nil, "", 200), &p)
	if p.PaymentStatus != "unpaid" || p.WalletBalanceToman != quote.TotalToman {
		t.Fatal("credit conflated payment")
	}
	get(admin, "POST", "/backoffice/commerce/orders/confirm", backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, "confirm", 200)
	json.Unmarshal(get(customer, "GET", "/storefront/account/orders/"+p.ID, nil, "", 200), &p)
	if p.PaymentStatus != "paid" || p.OrderStatus != "awaiting-supply" || p.WalletBalanceToman != "0" || len(p.Events) < 3 {
		t.Fatal(p)
	}
	get(admin, "POST", "/backoffice/commerce/orders/confirm", backoffice.PurchaseAdminInput{ID: p.ID, Version: 1}, "confirm-retry", 200)
	state, _ := repo.Snapshot(context.Background())
	if len(state.Commerce.Entries) != 2 || state.Offers[0].Stock != 9 {
		t.Fatal("double effect")
	}
	get(customer, "POST", "/storefront/cart", map[string]any{"action": "add", "offer_id": "offer", "quantity": 1, "version": 3, "price": 1}, "extra", 400)
}
func TestCommerceStartupRejectsFixtureInPublicMode(t *testing.T) {
	for _, cfg := range []config.Config{{Environment: "production"}, {Environment: "production", ReviewDatabaseURL: "postgres://test"}, {Environment: "test", DatabaseURL: "postgres://test"}, {Environment: "production", DatabaseURL: "postgres://test"}} {
		if _, e := bootstrap.NewHTTPServer(cfg, nil); e == nil {
			t.Fatal("unsafe config accepted")
		}
	}
}

func TestCommercePostgresLedgerReceiptAndRestart(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("PostgreSQL integration requires test DSN")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	control, e := pgx.Connect(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer control.Close(context.Background())
	schema := "commerce_test_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	safe := pgx.Identifier{schema}.Sanitize()
	if _, e = control.Exec(ctx, "CREATE SCHEMA "+safe); e != nil {
		t.Fatal(e)
	}
	defer control.Exec(context.Background(), "DROP SCHEMA "+safe+" CASCADE")
	options, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		t.Fatal(e)
	}
	options.ConnConfig.RuntimeParams["search_path"] = schema
	pool, e := pgxpool.NewWithConfig(ctx, options)
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	if e = postgresrepo.Migrate(ctx, pool, os.DirFS(filepath.Join("..", "..", "..", "..", "db", "migrations"))); e != nil {
		t.Fatal(e)
	}
	repo, e := postgresrepo.NewReview(ctx, pool, commerceSeed())
	if e != nil {
		t.Fatal(e)
	}
	commerceRegistration(t, repo)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	other := loginAt(t, h, "storefront", "09000000004")["token"].(string)
	admin := login(t, h, "09000000001")
	var d backoffice.DepositView
	json.Unmarshal(commerceHTTP(t, h, buyer, "", "POST", "/storefront/wallet/deposits", backoffice.DepositInput{MethodID: "card-transfer", DestinationID: "test-card", AmountToman: "200000", Date: time.Now().Format("2006-01-02"), Reference: "RECEIPT-001"}, "receipt-deposit", 200), &d)
	var binary bytes.Buffer
	png.Encode(&binary, image.NewRGBA(image.Rect(0, 0, 8, 8)))
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("version", fmt.Sprint(d.Version))
	part, _ := writer.CreateFormFile("file", "receipt.png")
	part.Write(binary.Bytes())
	writer.Close()
	r := httptest.NewRequest("POST", "/storefront/wallet/deposits/"+d.ID+"/receipt", &body)
	r.Header.Set("Authorization", "Bearer "+buyer)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("Idempotency-Key", "receipt-upload")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	if rr.Code != 200 {
		t.Fatal(rr.Code, rr.Body.String())
	}
	json.Unmarshal(rr.Body.Bytes(), &d)
	if d.ReceiptID == "" {
		t.Fatal("no receipt")
	}
	state, e := repo.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	for _, r := range state.Commerce.Receipts {
		if len(r.Data) > 0 {
			t.Fatal("receipt bytes leaked into JSON workspace")
		}
	}
	walletBefore := commerceHTTP(t, h, buyer, "", "GET", "/storefront/wallet", nil, "", 200)
	pool.Close()
	pool, e = pgxpool.NewWithConfig(ctx, options)
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	repo, e = postgresrepo.NewReview(ctx, pool, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = postgresrepo.NewLive(ctx, pool); e == nil {
		t.Fatal("review DB reused as live")
	}
	h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	buyer = loginAt(t, h, "storefront", "09000000003")["token"].(string)
	other = loginAt(t, h, "storefront", "09000000004")["token"].(string)
	admin = login(t, h, "09000000001")
	walletAfter := commerceHTTP(t, h, buyer, "", "GET", "/storefront/wallet", nil, "", 200)
	if !bytes.Equal(walletBefore, walletAfter) {
		t.Fatal("wallet restart mismatch")
	}
	commerceHTTP(t, h, other, "", "GET", "/storefront/wallet/receipts/"+d.ReceiptID, nil, "", 404)
	imageData := commerceHTTP(t, h, buyer, "", "GET", "/storefront/wallet/receipts/"+d.ReceiptID, nil, "", 200)
	if !bytes.Equal(binary.Bytes(), imageData) {
		t.Fatal("receipt restart mismatch")
	}
	commerceHTTP(t, h, admin, "", "GET", "/backoffice/commerce/receipts/"+d.ReceiptID, nil, "", 200)
	before, _ := repo.Snapshot(ctx)
	err := repo.Update(ctx, func(w *backoffice.Workspace) error { w.Commerce.Entries[0].Amount++; return nil })
	if err == nil {
		t.Fatal("immutable ledger edit accepted")
	}
	after, _ := repo.Snapshot(ctx)
	if after.Commerce.Entries[0].Amount != before.Commerce.Entries[0].Amount {
		t.Fatal("failed edit not rolled back")
	}
	if _, e = pool.Exec(ctx, `INSERT INTO commerce_wallet_entries(id,user_id,source,amount,kind,created_at,fingerprint) SELECT 'duplicate',user_id,source,amount,kind,created_at,fingerprint FROM commerce_wallet_entries LIMIT 1`); e == nil {
		t.Fatal("duplicate source accepted")
	}
	// A separate clean schema has no sample users, money or catalog in live mode.
	liveSchema := schema + "_live"
	liveSafe := pgx.Identifier{liveSchema}.Sanitize()
	control.Exec(ctx, "CREATE SCHEMA "+liveSafe)
	defer control.Exec(context.Background(), "DROP SCHEMA "+liveSafe+" CASCADE")
	liveOptions := options.Copy()
	liveOptions.ConnConfig.RuntimeParams["search_path"] = liveSchema
	livePool, e := pgxpool.NewWithConfig(ctx, liveOptions)
	if e != nil {
		t.Fatal(e)
	}
	defer livePool.Close()
	if e = postgresrepo.Migrate(ctx, livePool, os.DirFS(filepath.Join("..", "..", "..", "..", "db", "migrations"))); e != nil {
		t.Fatal(e)
	}
	live, e := postgresrepo.NewLive(ctx, livePool)
	if e != nil {
		t.Fatal(e)
	}
	empty, e := live.Snapshot(ctx)
	if e != nil || len(empty.Users) != 0 || len(empty.Products) != 0 || len(empty.Commerce.Entries) != 0 {
		t.Fatal("live fixture data")
	}
	if _, e = postgresrepo.NewReview(ctx, livePool, commerceSeed()); e == nil {
		t.Fatal("live DB opened as review")
	}
}

func TestCommerceVerifiedMobileCreatesOneSharedCustomerOnly(t *testing.T) {
	repo := backofficememory.New(commerceSeed())
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	request := func(t *testing.T, h http.Handler, method, path, body string, status int) []byte {
		var input any
		json.Unmarshal([]byte(body), &input)
		return commerceHTTP(t, h, "", "", method, path, input, "", status)
	}
	phone := "09000000777"
	before, _ := repo.Snapshot(context.Background())
	initialRequest := request(t, h, "POST", "/storefront/auth/request-otp", `{"phone":"`+phone+`"}`, 202)
	var challenge map[string]any
	json.Unmarshal(initialRequest, &challenge)
	after, _ := repo.Snapshot(context.Background())
	if len(after.Users) != len(before.Users) {
		t.Fatal("account created before OTP proof")
	}
	request(t, h, "POST", "/storefront/auth/verify-otp", `{"request_id":"`+challenge["request_id"].(string)+`","code":"bad"}`, 401)
	good := request(t, h, "POST", "/storefront/auth/verify-otp", `{"request_id":"`+challenge["request_id"].(string)+`","code":"`+challenge["test_code"].(string)+`"}`, 200)
	var session map[string]any
	json.Unmarshal(good, &session)
	after, _ = repo.Snapshot(context.Background())
	if len(after.Users) != len(before.Users)+1 {
		t.Fatal("customer not created exactly once")
	}
	for _, u := range after.Users {
		if u.Phone == phone && (u.Role != "" || u.AccessVersion != 1) {
			t.Fatal("signup privilege")
		}
	}
	token := session["token"].(string)
	commerceHTTP(t, h, token, "", "GET", "/storefront/cart", nil, "", 200)
	commerceHTTP(t, h, token, "", "GET", "/backoffice/commerce", nil, "", 401)
	request(t, h, "POST", "/storefront/auth/verify-otp", `{"request_id":"`+challenge["request_id"].(string)+`","code":"`+challenge["test_code"].(string)+`"}`, 401)
	request(t, h, "POST", "/backoffice/auth/request-otp", `{"phone":"`+phone+`"}`, 401)
}
