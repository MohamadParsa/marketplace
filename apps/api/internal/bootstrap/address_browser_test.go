package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/neshan"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type addressBrowserReverse struct{}

func (addressBrowserReverse) Lookup(ctx context.Context, lat, lng float64) (neshan.Result, error) {
	delay := 30 * time.Millisecond
	if lat < 35.71 {
		delay = 450 * time.Millisecond
	}
	select {
	case <-ctx.Done():
		return neshan.Result{}, ctx.Err()
	case <-time.After(delay):
	}
	city := "تهران"
	if lat > 36 {
		city = "شهر ناشناخته"
	}
	return neshan.Result{Status: "OK", Address: fmt.Sprintf("نشانی پیشنهادی ساختگی %.2f", lat), Province: "استان تهران", City: city}, nil
}

// Opt-in, memory-only, no real map/geocoder requests and no production data.
func TestAddressBrowserHarness(t *testing.T) {
	if os.Getenv("CHERRIVA_ADDRESS_BROWSER") != "1" {
		t.Skip("isolated address browser harness")
	}
	w := commerceSeed()
	w.IconMedia[0].Data = iconPNG(t, 12)
	w.Commerce.Addresses = []commerce.Address{{ID: "legacy-address", UserID: "u3", Version: 1, Recipient: "گیرنده قدیمی", Phone: "09000000111", Province: "تهران", City: "تهران", Street: "نشانی قدیمی ساختگی", PostalCode: "1234567890"}}
	for _, u := range w.Users {
		if u.Phone == "09000000003" {
			w.Commerce.Addresses[0].UserID = u.ID
		}
	}
	repo := backofficememory.New(w)
	// Token checks are delegated to the real app; reverse data is explicitly fake.
	base := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebSurface: "storefront", WebRoot: filepath.Join("..", "..", "..", "storefront-web", "build")}, repo)
	// Production authentication is exercised through base before fake lookup.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /storefront/account/address-map", func(w http.ResponseWriter, r *http.Request) {
		rr := httptest.NewRecorder()
		base.ServeHTTP(rr, r)
		if rr.Code != 200 {
			w.WriteHeader(rr.Code)
			w.Write(rr.Body.Bytes())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, no-store")
		w.Write([]byte(`{"enabled":true,"provider":"neshan","web_key":"fake-web-key"}`))
	})
	fakeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		check := r.Clone(r.Context())
		check.Method = "GET"
		check.URL.Path = "/storefront/account/address-map"
		rr := httptest.NewRecorder()
		base.ServeHTTP(rr, check)
		if rr.Code != 200 {
			w.WriteHeader(rr.Code)
			return
		} // fake endpoint validation is covered separately by handler tests
		var in struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			w.WriteHeader(400)
			return
		}
		out, e := (addressBrowserReverse{}).Lookup(r.Context(), in.Latitude, in.Longitude)
		if e != nil {
			w.WriteHeader(503)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, no-store")
		json.NewEncoder(w).Encode(map[string]any{"full_address": out.Address, "province": "تهران", "city": out.City, "geography_matched": out.City == "تهران"})
	})
	mux.Handle("POST /storefront/account/address-map/reverse", fakeHandler)
	mux.Handle("/", base)
	listener, e := net.Listen("tcp", "127.0.0.1:5198")
	if e != nil {
		t.Fatal(e)
	}
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	t.Cleanup(func() { server.Close() })
	go server.Serve(listener)
	t.Log("address browser ready on5198; memory/fake only")
	time.Sleep(15 * time.Minute)
}
