package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"reflect"
	"strings"
	"testing"
)

func businessInput() commerce.Business {
	return commerce.Business{Name: "کسب‌وکار ساختگی", About: "معرفی تأییدشده آزمون", Contact: "متن خصوصی پیش‌نویس", PublishedPages: map[string]bool{"about": true, "contact": false}, SocialLinks: []commerce.SocialLink{{ID: "second", Title: "شبکه دوم", Kind: "instagram", URL: "https://example.test/second", Position: 20, Active: true}, {ID: "first", Title: "شبکه اول", Kind: "other", URL: "https://example.test/first", Position: 10, Active: true}, {ID: "draft", Title: "پیش‌نویس", Kind: "other", Position: 0, Active: false}}}
}
func TestBusinessContentPermissionsValidationAndPublicDraftRedaction(t *testing.T) {
	repo := backofficememory.New(commerceSeed())
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := sellerLogin(t, h, "09120000002")
	b := businessInput()
	in := backoffice.CommerceSettingsInput{Business: &b}
	for _, token := range []string{"", viewer, seller} {
		code := 403
		if token == "" || token == seller {
			code = 401
		}
		commerceHTTP(t, h, token, "", "PATCH", "/backoffice/commerce/settings", in, "denied-business", code)
	}
	current, _ := repo.Snapshot(context.Background())
	in.Version = current.Commerce.Settings.Version
	commerceHTTP(t, h, admin, "", "PATCH", "/backoffice/commerce/settings", in, "save-business", 200)
	commerceHTTP(t, h, admin, "", "PATCH", "/backoffice/commerce/settings", in, "save-business", 200)
	body := commerceHTTP(t, h, "", "", "GET", "/storefront/business", nil, "", 200)
	var public commerce.Business
	json.Unmarshal(body, &public)
	if public.About != b.About || public.Contact != "" || public.PublishedPages != nil || len(public.SocialLinks) != 2 || public.SocialLinks[0].ID != "first" || strings.Contains(string(body), "پیش‌نویس") {
		t.Fatal("draft or order", string(body))
	}
	cfg := commerceHTTP(t, h, "", "", "GET", "/storefront/commerce/config", nil, "", 200)
	if strings.Contains(string(cfg), b.Contact) {
		t.Fatal("draft leaked through config")
	}
	in.Version++
	for _, bad := range []string{"javascript:alert(1)", "data:text/html,test", "//example.test", "http://example.test", "https://user:pass@example.test/x", "https://example.test/space x"} {
		t.Run(bad, func(t *testing.T) {
			copy := businessInput()
			copy.SocialLinks[0].URL = bad
			commerceHTTP(t, h, admin, "", "PATCH", "/backoffice/commerce/settings", backoffice.CommerceSettingsInput{Version: in.Version, Business: &copy}, "bad-url-"+bad, 400)
		})
	}
	invalid := businessInput()
	invalid.PublishedPages["terms"] = true
	commerceHTTP(t, h, admin, "", "PATCH", "/backoffice/commerce/settings", backoffice.CommerceSettingsInput{Version: in.Version, Business: &invalid}, "empty-publication", 400)
	invalid = businessInput()
	invalid.SocialLinks[1].ID = invalid.SocialLinks[0].ID
	commerceHTTP(t, h, admin, "", "PATCH", "/backoffice/commerce/settings", backoffice.CommerceSettingsInput{Version: in.Version, Business: &invalid}, "duplicate-social", 400)
	in.Version--
	commerceHTTP(t, h, admin, "", "PATCH", "/backoffice/commerce/settings", in, "stale-business", 409)
}
func TestBusinessContentPostgresCacheRestartAndProtectedState(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := commerceSeed()
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); e != nil {
		t.Fatal(e)
	}
	before, _ := repo.Snapshot(ctx)
	h := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}}))
	admin := login(t, h, "09000000001")
	old := cacheRequest(h, "/storefront/business", "")
	home := cacheRequest(h, "/storefront/home", "")
	b := businessInput()
	in := backoffice.CommerceSettingsInput{Version: before.Commerce.Settings.Version, Business: &b}
	commerceHTTP(t, h, admin, "", "PATCH", "/backoffice/commerce/settings", in, "pg-business", 200)
	fresh := cacheRequest(h, "/storefront/business", old.Header().Get("ETag"))
	if fresh.Code != 200 || fresh.Header().Get("ETag") == old.Header().Get("ETag") || !strings.Contains(fresh.Body.String(), b.About) || strings.Contains(fresh.Body.String(), b.Contact) {
		t.Fatal("stale cache/draft")
	}
	changed := cacheRequest(h, "/storefront/home", home.Header().Get("ETag"))
	if changed.Code != 200 && changed.Code != 304 {
		t.Fatal("home unavailable")
	}
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, _ := reopened.Snapshot(ctx)
	if after.Commerce.Settings.Business.Contact != b.Contact || after.Commerce.Settings.Business.PublishedPages["contact"] {
		t.Fatal("private draft lost")
	}
	for name, pair := range map[string][2]any{"users": {before.Users, after.Users}, "roles": {before.Access, after.Access}, "orders": {before.Orders, after.Orders}, "ledger": {before.Ledger, after.Ledger}, "bank": {before.Commerce.Settings.Destination, after.Commerce.Settings.Destination}, "methods": {before.Commerce.Settings.Methods, after.Commerce.Settings.Methods}, "shipping": {before.Commerce.Settings.Shipping, after.Commerce.Settings.Shipping}, "entries": {before.Commerce.Entries, after.Commerce.Entries}} {
		if !reflect.DeepEqual(pair[0], pair[1]) {
			t.Fatal("changed " + name)
		}
	}
	privateReads := repo.privateReads.Load()
	commerceHTTP(t, h, "", "", "GET", "/storefront/business", nil, "", 200)
	if repo.privateReads.Load() != privateReads {
		t.Fatal("footer loaded private workspace")
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	a2 := login(t, h2, "09000000001")
	commerceHTTP(t, h2, a2, "", "PATCH", "/backoffice/commerce/settings", in, "pg-business", 200)
	b.SocialLinks[0].Active = false
	b.PublishedPages["about"] = false
	in.Version++
	commerceHTTP(t, h2, a2, "", "PATCH", "/backoffice/commerce/settings", in, "hide-business", 200)
	res := cacheRequest(h, "/storefront/business", fresh.Header().Get("ETag"))
	if res.Code != 200 || strings.Contains(res.Body.String(), b.About) || strings.Contains(res.Body.String(), "شبکه دوم") {
		t.Fatal("deactivation cache")
	}
}
