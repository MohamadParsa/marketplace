package httptransport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficeidentity"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/neshan"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	httptransport "github.com/cherriva/marketplace/apps/api/internal/transport/http"
	"net/http/httptest"
	"testing"
)

type fakeReverse struct {
	result neshan.Result
	err    error
	calls  int
}

func (f *fakeReverse) Lookup(context.Context, float64, float64) (neshan.Result, error) {
	f.calls++
	return f.result, f.err
}
func TestAddressMapAuthenticationGeographyFailureAndBudget(t *testing.T) {
	w := backofficememory.Fixture()
	w.Commerce = &commerce.State{Settings: commerce.DefaultSettings()}
	w.Commerce.Settings.Geography = []commerce.Province{{ID: "tehran", Name: "تهران", Active: true, Cities: []commerce.City{{ID: "tehran-city", Name: "تهران", Active: true}}}}
	repo := backofficememory.New(w)
	store := authmemory.NewStore(nil)
	notifier := authmemory.NewNotifier()
	a := auth.NewService(backofficeidentity.New(repo), store, store, notifier, authmemory.Clock{}, authmemory.Random{})
	ctx := context.Background()
	id, e := a.RequestOTP(ctx, identity.SurfaceCustomer, "09000000003")
	if e != nil {
		t.Fatal(e)
	}
	code, _ := notifier.Code(identity.SurfaceCustomer, "09000000003")
	session, e := a.VerifyOTP(ctx, identity.SurfaceCustomer, id, code)
	if e != nil {
		t.Fatal(e)
	}
	fake := &fakeReverse{result: neshan.Result{Address: "نشانی ساختگی", Province: "استان تهران", City: "تهران"}}
	h := httptransport.NewAddressMapHandler(backoffice.NewService(repo, authmemory.Clock{}, authmemory.Random{}), a, "fake-web-key", fake, 4)
	call := func(method, token, body string, status int) map[string]any {
		t.Helper()
		r := httptest.NewRequest(method, "/storefront/account/address-map/reverse", bytes.NewBufferString(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+token)
		rw := httptest.NewRecorder()
		h.Handle(rw, r)
		if rw.Code != status {
			t.Fatalf("%d expected%d: %s", rw.Code, status, rw.Body.String())
		}
		if rw.Header().Get("Cache-Control") != "private, no-store" {
			t.Fatal("cache")
		}
		var out map[string]any
		json.Unmarshal(rw.Body.Bytes(), &out)
		return out
	}
	call("POST", "", `{"latitude":35,"longitude":51}`, 401)
	config := call("GET", session.Token, "", 200)
	if config["web_key"] != "fake-web-key" {
		t.Fatal("public web key")
	}
	for _, body := range []string{`{"latitude":35}`, `{"latitude":100,"longitude":51}`, `{"latitude":35,"longitude":51,"url":"http://private"}`, `{"latitude":null,"longitude":51}`} {
		call("POST", session.Token, body, 400)
	}
	out := call("POST", session.Token, `{"latitude":35,"longitude":51}`, 200)
	if out["geography_matched"] != true || out["province"] != "تهران" {
		t.Fatal(out)
	}
	fake.result.City = "شهر ناشناخته"
	out = call("POST", session.Token, `{"latitude":35,"longitude":51}`, 200)
	if out["geography_matched"] != false || out["city"] != "" {
		t.Fatal("unknown city silently chosen")
	}
	fake.err = errors.New("private-key-or-remote-body")
	out = call("POST", session.Token, `{"latitude":35,"longitude":51}`, 503)
	if out["code"] != "map_provider_unavailable" {
		t.Fatal(out)
	}
	call("POST", session.Token, `{"latitude":35,"longitude":51}`, 503)
	call("POST", session.Token, `{"latitude":35,"longitude":51}`, 429)
	if fake.calls != 4 {
		t.Fatal("budget not enforced")
	}
	repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].ID == session.User.ID {
				w.Users[i].Active = false
			}
		}
		return nil
	})
	call("GET", session.Token, "", 401)
}
