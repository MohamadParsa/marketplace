package backoffice_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http/httptest"
)

func seedFixture() (backoffice.CatalogSeed, []catalog.IconMedia) {
	c := backoffice.SeedCatalog{
		Categories:   []catalog.Category{{ID: "gsp-cat-1", Name: "موبایل", EnglishName: "Imported Mobile", Published: true, Version: 1, Bindings: []catalog.AttributeBinding{{AttributeID: "gsp-axis-color", Axis: true}}}},
		Brands:       []catalog.Brand{{ID: "gsp-brand-1", Name: "برند", Active: true, Version: 1}},
		Attributes:   []catalog.AttributeDefinition{{ID: "gsp-axis-color", Code: "gsp-color", Name: "رنگ", Kind: "enum", Options: []string{"مشکی", "سفید"}, Version: 1}},
		Products:     []catalog.Product{{ID: "gsp-product-1", Name: "محصول", CategoryID: "gsp-cat-1", BrandID: "gsp-brand-1", ImageID: "gsp-media-1", Model: "shared", Status: "published", Version: 1, VariantVersion: 1, Content: &catalog.ProductContent{FullDescription: "<p>شرح <strong>کامل</strong></p><script>alert(1)</script>"}}},
		Variants:     []catalog.Variant{{ID: "gsp-variant-1", ProductID: "gsp-product-1", SKU: "GSP-1", Version: 1, Values: []catalog.AttributeValue{{AttributeID: "gsp-axis-color", Value: "مشکی"}}}},
		Offers:       []catalog.Offer{{ID: "gsp-offer-1", ProductID: "gsp-product-1", VariantID: "gsp-variant-1", StoreID: "gsp-store-1", Price: 100000, RegularPrice: 110000, Stock: 2, Version: 1}},
		Stores:       []store.Store{{ID: "gsp-store-1", Name: "فروشنده", Model: "shared", Active: true, ContentPolicy: "review_required", ContentPolicyVersion: 1}},
		HomeSettings: backoffice.HomeSettings{Rails: []backoffice.HomeRail{{Key: "mobile", CategoryID: "gsp-cat-1", Limit: 10}, {Key: "earbuds", CategoryID: "gsp-cat-1", Limit: 10}, {Key: "power-bank", CategoryID: "gsp-cat-1", Limit: 10}, {Key: "smartwatch", CategoryID: "gsp-cat-1", Limit: 10}}},
	}
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	s := backoffice.CatalogSeed{SchemaVersion: 1, DatasetID: "test-catalog", Purpose: "test-catalog-snapshot", Currency: "IRT", Namespace: "gsp-", Catalog: c, Media: []backoffice.SeedMedia{{ID: "gsp-media-1", URL: "https://cdn.gooshishop.ir/images/thumbs/n/Entity/test.png"}}}
	return s, []catalog.IconMedia{{ID: "gsp-media-1", Data: b.Bytes()}}
}
func TestCatalogSeedRejectsInvalidRelationsAndSourceURLs(t *testing.T) {
	cases := map[string]func(*backoffice.CatalogSeed){
		"currency":  func(s *backoffice.CatalogSeed) { s.Currency = "IRR" },
		"namespace": func(s *backoffice.CatalogSeed) { s.Catalog.Products[0].ID = "p1" },
		"duplicate id": func(s *backoffice.CatalogSeed) {
			s.Catalog.Products = append(s.Catalog.Products, s.Catalog.Products[0])
		},
		"private image URL": func(s *backoffice.CatalogSeed) { s.Media[0].URL = "http://127.0.0.1/private" },
		"URL credentials": func(s *backoffice.CatalogSeed) {
			s.Media[0].URL = "https://user:pass@cdn.gooshishop.ir/images/thumbs/a.png"
		},
		"bad hash":      func(s *backoffice.CatalogSeed) { s.Media[0].SHA256 = "invalid" },
		"cycle":         func(s *backoffice.CatalogSeed) { s.Catalog.Categories[0].ParentID = "gsp-cat-1" },
		"unknown brand": func(s *backoffice.CatalogSeed) { s.Catalog.Products[0].BrandID = "gsp-brand-missing" },
		"hidden parent": func(s *backoffice.CatalogSeed) { s.Catalog.Categories[0].Published = false },
		"axis mismatch": func(s *backoffice.CatalogSeed) { s.Catalog.Variants[0].Values = nil },
		"duplicate variant selection": func(s *backoffice.CatalogSeed) {
			v := s.Catalog.Variants[0]
			v.ID = "gsp-variant-2"
			v.SKU = "GSP-2"
			s.Catalog.Variants = append(s.Catalog.Variants, v)
		},
		"cross product offer": func(s *backoffice.CatalogSeed) { s.Catalog.Offers[0].ProductID = "gsp-product-missing" },
		"negative money":      func(s *backoffice.CatalogSeed) { s.Catalog.Offers[0].Price = -1 },
		"fake discount":       func(s *backoffice.CatalogSeed) { s.Catalog.Offers[0].RegularPrice = 1 },
		"reserved stock":      func(s *backoffice.CatalogSeed) { s.Catalog.Offers[0].Reserved = 1 },
		"no offer":            func(s *backoffice.CatalogSeed) { s.Catalog.Offers = nil },
		"foreign image":       func(s *backoffice.CatalogSeed) { s.Catalog.Products[0].ImageID = "gsp-media-unknown" },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			s, _ := seedFixture()
			change(&s)
			if err := backoffice.ValidateCatalogSeed(&s); err == nil {
				t.Fatal("invalid snapshot accepted")
			}
		})
	}
}
func TestCatalogSeedConcurrentRetriesPreserveAccountsAndMoney(t *testing.T) {
	before := backofficememory.Fixture()
	repo := backofficememory.New(before)
	ctx := context.Background()
	before, _ = repo.Snapshot(ctx)
	var wg sync.WaitGroup
	var mu sync.Mutex
	created := 0
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, m := seedFixture()
			r, e := backoffice.ApplyCatalogSeed(ctx, repo, s, m, true, time.Now())
			if e != nil {
				t.Error(e)
				return
			}
			mu.Lock()
			if !r.AlreadyApplied {
				created++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	after, _ := repo.Snapshot(ctx)
	if created != 1 || len(after.Products) != len(before.Products)+1 || len(after.Audit) != len(before.Audit)+1 {
		t.Fatal("retry duplicated records")
	}
	if !reflect.DeepEqual(before.Users, after.Users) || !reflect.DeepEqual(before.Access, after.Access) || !reflect.DeepEqual(before.Commerce, after.Commerce) || !reflect.DeepEqual(before.Ledger, after.Ledger) {
		t.Fatal("non-catalog data changed")
	}
	// The panel uses Cherriva aliases; reading it must not alter imported IDs or retry evidence.
	seed, media := seedFixture()
	if e := backoffice.ValidateCatalogSeed(&seed); e != nil {
		t.Fatal(e)
	}
	seedBytes, _ := json.Marshal(seed)
	inputHash := sha256.Sum256(seedBytes)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	call := func(path string, data any, token string) []byte {
		b, _ := json.Marshal(data)
		method := "POST"
		if data == nil {
			method = "GET"
		}
		r := httptest.NewRequest(method, path, bytes.NewReader(b))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 && w.Code != 202 {
			t.Fatalf("panel: %d %s", w.Code, w.Body.String())
		}
		return w.Body.Bytes()
	}
	var challenge struct {
		RequestID string `json:"request_id"`
		TestCode  string `json:"test_code"`
	}
	json.Unmarshal(call("/backoffice/auth/request-otp", map[string]string{"phone": "09000000001"}, ""), &challenge)
	var session struct {
		Token string `json:"token"`
	}
	json.Unmarshal(call("/backoffice/auth/verify-otp", map[string]string{"request_id": challenge.RequestID, "code": challenge.TestCode}, ""), &session)
	var panel backoffice.Workspace
	json.Unmarshal(call("/backoffice/workspace", nil, session.Token), &panel)
	found := false
	for _, a := range panel.Attributes {
		if a.ID == backoffice.PublicReference("a", "gsp-axis-color") {
			found = a.Code == backoffice.PublicReference("a", "gsp-color")
		}
	}
	if !found {
		t.Fatal("imported attribute not projected")
	}
	retry, e := backoffice.ApplyCatalogSeed(ctx, repo, seed, media, true, time.Now())
	if e != nil || !retry.AlreadyApplied {
		t.Fatal("panel broke import retry", e)
	}
	private, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(after, private) {
		t.Fatal("panel/retry rewrote imported or protected data")
	}
	seedBytes, _ = json.Marshal(seed)
	if inputHash != sha256.Sum256(seedBytes) {
		t.Fatal("snapshot input mutated")
	}
	p := after.Products[len(after.Products)-1]
	if strings.Contains(p.Content.FullDescription, "script") || strings.Contains(p.Content.FullDescription, "alert") || !strings.Contains(p.Content.FullDescription, "<strong>") {
		t.Fatal("unsafe HTML or formatting lost")
	}
}
func TestCatalogSeedCollisionRollsBackEverything(t *testing.T) {
	w := backofficememory.Fixture()
	w.Brands = append(w.Brands, catalog.Brand{ID: "gsp-brand-1", Name: "ویرایش محلی"})
	repo := backofficememory.New(w)
	s, m := seedFixture()
	if _, e := backoffice.ApplyCatalogSeed(context.Background(), repo, s, m, true, time.Now()); e == nil {
		t.Fatal("collision overwrote local data")
	}
	after, _ := repo.Snapshot(context.Background())
	before, _ := json.Marshal(w)
	actual, _ := json.Marshal(after)
	if !bytes.Equal(before, actual) {
		t.Fatal("partial import survived collision")
	}
}
func TestCatalogSeedMissingMediaWritesNothing(t *testing.T) {
	w := backofficememory.Fixture()
	repo := backofficememory.New(w)
	s, _ := seedFixture()
	if _, e := backoffice.ApplyCatalogSeed(context.Background(), repo, s, nil, false, time.Now()); e == nil {
		t.Fatal("unhydrated image accepted")
	}
	a, _ := repo.Snapshot(context.Background())
	if len(a.Audit) != len(w.Audit) || len(a.Products) != len(w.Products) {
		t.Fatal("partial write")
	}
}
func TestCatalogSeedPostgresRestartAndRetry(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("set isolated PostgreSQL URL")
	}
	ctx := context.Background()
	control, e := pgx.Connect(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer control.Close(ctx)
	schema := "seed_test_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	safe := pgx.Identifier{schema}.Sanitize()
	if _, e = control.Exec(ctx, "CREATE SCHEMA "+safe); e != nil {
		t.Fatal(e)
	}
	defer func() { _, _ = control.Exec(ctx, "DROP SCHEMA "+safe+" CASCADE") }()
	opts, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		t.Fatal(e)
	}
	opts.ConnConfig.RuntimeParams["search_path"] = schema
	open := func() *pgxpool.Pool {
		p, e := pgxpool.NewWithConfig(ctx, opts.Copy())
		if e != nil {
			t.Fatal(e)
		}
		return p
	}
	db := open()
	defer db.Close()
	if e = postgresrepo.Migrate(ctx, db, os.DirFS(filepath.Join("..", "..", "..", "..", "..", "db", "migrations"))); e != nil {
		t.Fatal(e)
	}
	repo, e := postgresrepo.NewReview(ctx, db, backofficememory.Fixture())
	if e != nil {
		t.Fatal(e)
	}
	before, _ := repo.Snapshot(ctx)
	s, m := seedFixture()
	path := filepath.Join(t.TempDir(), "image.png")
	if e = os.WriteFile(path, m[0].Data, 0600); e != nil {
		t.Fatal(e)
	}
	sum := sha256.Sum256(m[0].Data)
	files := []backoffice.CatalogSeedFile{{ID: m[0].ID, Path: path, SHA256: hex.EncodeToString(sum[:]), Bytes: int64(len(m[0].Data))}}
	if _, e = backoffice.ApplyCatalogSeedFiles(ctx, repo, s, files, true, time.Now()); e != nil {
		t.Fatal(e)
	}
	db.Close()
	db = open()
	defer db.Close()
	repo, e = postgresrepo.NewReview(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	s, m = seedFixture()
	r, e := backoffice.ApplyCatalogSeed(ctx, repo, s, m, true, time.Now())
	if e != nil || !r.AlreadyApplied {
		t.Fatalf("restart retry: %v %+v", e, r)
	}
	after, _ := repo.Snapshot(ctx)
	for _, m := range after.IconMedia {
		if len(m.Data) != 0 || !m.Stored {
			t.Fatal("blob remained inside workspace JSON")
		}
	}
	data, e := repo.ReadCatalogMedia(ctx, "gsp-media-1")
	if e != nil || !bytes.Equal(data, m[0].Data) {
		t.Fatal("separate media missing after restart")
	}
	handler := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/catalog/icons/gsp-media-1", nil))
	if rr.Code != 200 || !bytes.Equal(rr.Body.Bytes(), m[0].Data) {
		t.Fatal("public image did not resolve stored blob")
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/catalog/icons/unattached", nil))
	if rr.Code != 404 {
		t.Fatal("unattached blob exposed")
	}
	// A non-catalog write preserves the separated bytes.
	if e = repo.Update(ctx, func(w *backoffice.Workspace) error { return nil }); e != nil {
		t.Fatal(e)
	}
	if _, e = repo.ReadCatalogMedia(ctx, "gsp-media-1"); e != nil {
		t.Fatal("unrelated write deleted media")
	}
	if len(after.Products) != len(before.Products)+1 || !reflect.DeepEqual(after.Users, before.Users) || !reflect.DeepEqual(after.Commerce, before.Commerce) {
		t.Fatal("restart lost import or changed identity/finance")
	}
}

func TestCatalogSeedDecodeConvertsTomanExactlyOnce(t *testing.T) {
	seed, _ := seedFixture()
	seed.Catalog.Offers[0].Price = 2995000
	seed.Catalog.Offers[0].RegularPrice = 3195000
	raw, _ := json.Marshal(seed)
	decoded, e := backoffice.DecodeCatalogSeed(raw)
	if e != nil {
		t.Fatal(e)
	}
	if decoded.Catalog.Offers[0].Price != 29950000 || decoded.Catalog.Offers[0].RegularPrice != 31950000 {
		t.Fatal("toman was not converted to internal rial")
	}
	if e = backoffice.ValidateCatalogSeed(&decoded); e != nil {
		t.Fatal(e)
	}
	if decoded.Catalog.Offers[0].Price != 29950000 {
		t.Fatal("validation converted amount twice")
	}
	seed.Catalog.Offers[0].Price = backoffice.MaxOfferPrice/10 + 1
	raw, _ = json.Marshal(seed)
	if _, e = backoffice.DecodeCatalogSeed(raw); e == nil {
		t.Fatal("overflowing conversion accepted")
	}
}
