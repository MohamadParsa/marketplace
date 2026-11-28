package neshan_test

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/neshan"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestReverseProtocolBoundariesAndPrivateError(t *testing.T) {
	var valid bool
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		valid = r.Header.Get("Api-Key") == "fake-server-key" && r.URL.Path == "/v5/reverse" && r.URL.Query().Get("lat") == "35.700000" && r.URL.Query().Get("lng") == "51.400000"
		w.Write([]byte(`{"status":"OK","formatted_address":"نشانی ساختگی","state":"استان تهران","city":"تهران"}`))
	}))
	defer s.Close()
	out, e := neshan.NewWithEndpoint("fake-server-key", s.URL+"/v5/reverse", nil).Lookup(context.Background(), 35.7, 51.4)
	if e != nil || !valid || out.City != "تهران" {
		t.Fatal("protocol", e)
	}
	for _, body := range []string{`{"status":"ERROR","message":"fake-server-key"}`, `not json`, strings.Repeat("x", 32769)} {
		t.Run(body[:5], func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
			defer ts.Close()
			_, e := neshan.NewWithEndpoint("fake-server-key", ts.URL, nil).Lookup(context.Background(), 0, 0)
			if e != neshan.ErrUnavailable || strings.Contains(e.Error(), "fake-server-key") {
				t.Fatal(e)
			}
		})
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(30 * time.Millisecond); w.Write([]byte(`{}`)) }))
	defer ts.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, e = neshan.NewWithEndpoint("fake", ts.URL, nil).Lookup(ctx, 0, 0); e != neshan.ErrUnavailable {
		t.Fatal("timeout", e)
	}
}
