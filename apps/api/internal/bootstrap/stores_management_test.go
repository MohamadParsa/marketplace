package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
)

func storeUpload(t *testing.T, h http.Handler, token, id string, version int, key string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	form := multipart.NewWriter(&b)
	form.WriteField("version", strconv.Itoa(version))
	f, e := form.CreateFormFile("file", "private.png")
	if e != nil {
		t.Fatal(e)
	}
	f.Write(data)
	form.Close()
	r := httptest.NewRequest("POST", "/backoffice/stores/"+id+"/image", &b)
	r.Header.Set("Content-Type", form.FormDataContentType())
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Idempotency-Key", key)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func sampleStoreInput() backoffice.StoreInput {
	return backoffice.StoreInput{Name: "فروشگاه ساختگی", Active: true, City: "تهران", Profile: backoffice.StoreProfileInput{Description: "توضیح ساختگی", Email: "Store@Example.test", Province: "تهران", Address: "نشانی ساختگی", PostalCode: "۱۲۳۴۵۶۷۸۹۰", Phone: "۰۲۱۱۲۳۴۵۶۷۸", Mobile: "۰۹۱۲۰۰۰۰۹۸۱"}}
}
func TestStoreManagementRoundtripPrivacyVersionAndValidation(t *testing.T) {
	ctx := context.Background()
	repo := backofficememory.New(commerceSeed())
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := loginAt(t, h, "seller", "09120000002")["token"].(string)
	in := sampleStoreInput()
	var row backoffice.AdminStoreRow
	raw := commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/save", in, "new-store", 200)
	json.Unmarshal(raw, &row)
	if row.ID == "" || row.Version != 1 || row.Profile.Email != "store@example.test" || row.Profile.Mobile != "09120000981" || row.Profile.PostalCode != "1234567890" {
		t.Fatal("normalization", row)
	}
	if string(raw) != string(commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/save", in, "new-store", 200)) {
		t.Fatal("duplicate create")
	}
	in.Name = "نام دیگر"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/save", in, "new-store", 409)
	for _, tok := range []string{viewer, seller} {
		status := 403
		if tok == seller {
			status = 401
		}
		commerceHTTP(t, h, tok, "", "POST", "/backoffice/stores/query", map[string]any{}, "", status)
		commerceHTTP(t, h, tok, "", "POST", "/backoffice/stores/save", in, "forbidden", status)
	}
	for _, payload := range []map[string]any{{"name": "x", "version": 0}, {"name": "x", "active": true}, {"name": "x", "active": true, "version": 0, "owner": "injected"}, {"name": "x", "active": true, "version": 0, "profile": map[string]any{"image_id": "injected"}}} {
		commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/save", payload, "invalid-fields", 400)
	}
	bad := sampleStoreInput()
	bad.ID = row.ID
	bad.Version = row.Version
	for n, change := range []func(*backoffice.StoreInput){func(v *backoffice.StoreInput) { v.Profile.Email = "invalid" }, func(v *backoffice.StoreInput) { v.Profile.Phone = "abc" }, func(v *backoffice.StoreInput) { v.Profile.Mobile = "1234" }, func(v *backoffice.StoreInput) { v.Profile.PostalCode = "123" }, func(v *backoffice.StoreInput) { v.City = "قم" }, func(v *backoffice.StoreInput) { v.Profile.Province = "" }, func(v *backoffice.StoreInput) { v.Name = "" }} {
		v := bad
		change(&v)
		commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/save", v, fmt.Sprint("bad-", n), 400)
	}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/save", backoffice.StoreInput{ID: row.ID, Version: 0, Name: row.Name, Active: true}, "stale", 409)
	if got := storeUpload(t, h, admin, row.ID, 1, "oversize-store-image", make([]byte, 2<<20)); got.Code != 413 {
		t.Fatal("oversized request", got.Code)
	}
	image := storeUpload(t, h, admin, row.ID, 1, "store-image", append(iconPNG(t, 12), []byte("private-trailing-data")...))
	if image.Code != 200 {
		t.Fatal(image.Code, image.Body.String())
	}
	json.Unmarshal(image.Body.Bytes(), &row)
	if got := storeUpload(t, h, admin, row.ID, 1, "store-image", append(iconPNG(t, 12), []byte("private-trailing-data")...)); got.Code != 200 {
		t.Fatal("image retry", got.Code)
	}
	if got := storeUpload(t, h, admin, row.ID, 2, "invalid-image", []byte("not an image")); got.Code != 400 {
		t.Fatal("invalid image", got.Code)
	}
	data := commerceHTTP(t, h, admin, "", "GET", "/backoffice/stores/"+row.ID+"/image", nil, "", 200)
	if bytes.Contains(data, []byte("private-trailing-data")) {
		t.Fatal("image metadata retained")
	}
	commerceHTTP(t, h, viewer, "", "GET", "/backoffice/stores/"+row.ID+"/image", nil, "", 403)
	commerceHTTP(t, h, "", "", "GET", "/catalog/icons/"+row.Profile.ImageID, nil, "", 404)
	privateState, _ := repo.Snapshot(ctx)
	for _, actorToken := range []string{viewer, seller} {
		path := "/backoffice/workspace"
		if actorToken == seller {
			path = "/seller/workspace"
		}
		body := commerceHTTP(t, h, actorToken, "", "GET", path, nil, "", 200)
		if strings.Contains(string(body), "store@example.test") || strings.Contains(string(body), "store_mutations") {
			t.Fatal("private profile in workspace")
		}
	}
	// A catalog icon update must not garbage-collect unrelated store images.
	if got := iconUpload(t, h, admin, "category", "regular", 0, "keep-store-media", iconPNG(t, 8)); got.Code != 204 {
		t.Fatal(got.Code)
	}
	commerceHTTP(t, h, admin, "", "GET", "/backoffice/stores/"+row.ID+"/image", nil, "", 200)
	commerceHTTP(t, h, admin, "", "DELETE", "/backoffice/stores/"+row.ID+"/image", map[string]int{"version": 2}, "remove-store-image", 200)
	commerceHTTP(t, h, admin, "", "DELETE", "/backoffice/stores/"+row.ID+"/image", map[string]int{"version": 2}, "remove-store-image", 200)
	commerceHTTP(t, h, admin, "", "GET", "/backoffice/stores/"+row.ID+"/image", nil, "", 404)
	after, _ := repo.Snapshot(ctx)
	if len(after.StoreMutations) != len(privateState.StoreMutations) {
		t.Fatal("image changed store retries")
	}
}
func TestStoreDirectoryFiltersMembersAndDelegatedPermissions(t *testing.T) {
	seed := commerceSeed()
	for i := 0; i < 125; i++ {
		seed.Stores = append(seed.Stores, store.Store{ID: fmt.Sprintf("directory-%03d", i), Name: fmt.Sprintf("نمونه %03d", i), City: "تهران", Active: true, Profile: &store.Profile{Province: "تهران"}})
	}
	seed.Access.Roles = append(seed.Access.Roles, identity.RoleDefinition{ID: "store-reader", Name: "مشاهده فروشگاه", Scope: identity.ScopeGlobal, Version: 1, Permissions: []string{backoffice.StoresRead}})
	seed.Access.Grants = append(seed.Access.Grants, identity.RoleGrant{UserID: "viewer", RoleID: "store-reader"})
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	reader := login(t, h, "09000000005")
	var out backoffice.StoreDirectory
	query := func(token string, f backoffice.StoreFilter) {
		t.Helper()
		json.Unmarshal(commerceHTTP(t, h, token, "", "POST", "/backoffice/stores/query", map[string]any{"filter": f}, "", 200), &out)
	}
	query(reader, backoffice.StoreFilter{Name: "نمونه", Province: "تهران", City: "تهران", Active: "true", Page: 2, Size: 20})
	if out.Total != 125 || len(out.Rows) != 20 || out.Page != 2 || out.CanWrite || out.CanMembers {
		t.Fatal("paging/rbac")
	}
	query(reader, backoffice.StoreFilter{ID: "s1"})
	if len(out.Rows) != 1 || out.Rows[0].Members != nil {
		t.Fatal("members leaked")
	}
	commerceHTTP(t, h, reader, "", "POST", "/backoffice/stores/query", map[string]any{"filter": backoffice.StoreFilter{Member: "seller-test"}}, "", 403)
	query(admin, backoffice.StoreFilter{Member: "seller-test"})
	if out.Total == 0 {
		t.Fatal("fixture membership")
	}
	for _, r := range out.Rows {
		found := false
		for _, m := range r.Members {
			if m.UserID == "seller-test" {
				found = true
			}
		}
		if !found {
			t.Fatal("wrong member filter")
		}
	}
	commerceHTTP(t, h, reader, "", "POST", "/backoffice/stores/export", map[string]any{}, "", 403)
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/export", map[string]any{"filter": backoffice.StoreFilter{Name: "نمونه"}}, "", 200), &out)
	if len(out.Rows) != 125 {
		t.Fatal("partial export")
	}
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/export", map[string]any{"filter": backoffice.StoreFilter{Name: "نمونه"}, "ids": []string{"directory-000", "directory-100", "s1"}}, "", 200), &out)
	if len(out.Rows) != 2 {
		t.Fatal("export filter intersection")
	}
}
func TestStorePostgresRestartKeepsContactsMediaAndExistingHistory(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce = commerceSeed().Commerce
		w.Offers[0].StoreID = w.Stores[0].ID
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	before, _ := repo.Snapshot(ctx)
	var phone string
	for _, u := range before.Users {
		if u.Role == identity.RoleSuperAdmin {
			phone = u.Phone
		}
	}
	cache := backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}})
	h := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	path := "/storefront/products/" + before.Products[0].ID
	initial := cacheRequest(h, path, "")
	if initial.Code != 200 {
		t.Fatal("public product", initial.Code)
	}
	admin := login(t, h, phone)
	in := sampleStoreInput()
	old := before.Stores[0]
	in.ID = old.ID
	in.Version = old.Version
	var row backoffice.AdminStoreRow
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/stores/save", in, "pg-store-profile", 200), &row)
	if row.Owner != old.Owner || row.Model != old.Model || row.LeadDays != old.LeadDays || row.ContentPolicy != old.ContentPolicy || row.ContentPolicyVersion != old.ContentPolicyVersion {
		t.Fatal("legacy store rewritten")
	}
	fresh := cacheRequest(h, path, initial.Header().Get("ETag"))
	if fresh.Code != 200 || !strings.Contains(fresh.Body.String(), in.Name) || strings.Contains(fresh.Body.String(), "store@example.test") {
		t.Fatal("public store name/cache privacy", fresh.Code)
	}
	image := storeUpload(t, h, admin, row.ID, row.Version, "pg-store-image", iconPNG(t, 16))
	if image.Code != 200 {
		t.Fatal(image.Code, image.Body.String())
	}
	json.Unmarshal(image.Body.Bytes(), &row)
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, _ := reopened.Snapshot(ctx)
	for name, pair := range map[string][2]any{"users": {before.Users, after.Users}, "roles": {before.Access, after.Access}, "members": {before.StoreMemberships, after.StoreMemberships}, "commerce": {before.Commerce, after.Commerce}, "orders": {before.Orders, after.Orders}, "ledger": {before.Ledger, after.Ledger}, "transfers": {before.Transfers, after.Transfers}, "products": {before.Products, after.Products}, "variants": {before.Variants, after.Variants}, "offers": {before.Offers, after.Offers}} {
		if !reflect.DeepEqual(pair[0], pair[1]) {
			t.Fatal("changed", name)
		}
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	admin2 := login(t, h2, phone)
	var listing backoffice.StoreDirectory
	json.Unmarshal(commerceHTTP(t, h2, admin2, "", "POST", "/backoffice/stores/query", map[string]any{"filter": backoffice.StoreFilter{ID: row.ID}}, "", 200), &listing)
	if listing.Total != 1 || listing.Rows[0].Profile.Email != "store@example.test" || listing.Rows[0].Version != old.Version+2 {
		t.Fatal("store profile not persisted")
	}
	commerceHTTP(t, h2, admin2, "", "GET", "/backoffice/stores/"+row.ID+"/image", nil, "", 200)
	commerceHTTP(t, h2, admin2, "", "POST", "/backoffice/stores/save", in, "pg-store-profile", 200)
	public, e := reopened.PublicCatalog(ctx, "")
	if e != nil {
		t.Fatal(e)
	}
	for _, s := range public.Stores {
		if s.Profile != nil {
			t.Fatal("private store profile in public projection")
		}
	}
}
