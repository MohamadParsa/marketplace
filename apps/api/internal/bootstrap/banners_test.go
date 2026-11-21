package bootstrap_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/content"
)

func bannerUploadHTTP(t *testing.T, h http.Handler, token, id, slot string, version int, key string, data []byte, want int) backoffice.BannerMediaResult {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	_ = form.WriteField("version", strconv.Itoa(version))
	_ = form.WriteField("idempotency_key", key)
	file, _ := form.CreateFormFile("file", "banner.png")
	_, _ = file.Write(data)
	_ = form.Close()
	r := httptest.NewRequest("POST", "/backoffice/banners/"+id+"/media/"+slot, &body)
	r.Header.Set("Content-Type", form.FormDataContentType())
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != want {
		t.Fatalf("upload got %d want %d: %s", w.Code, want, w.Body.String())
	}
	var result backoffice.BannerMediaResult
	if want == 200 {
		_ = json.Unmarshal(w.Body.Bytes(), &result)
	}
	return result
}

func TestBannerHTTPPublicationResponsiveHTMLAndPrivateAssets(t *testing.T) {
	h := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	seller := sellerLogin(t, h, "09120000002")
	productCall(t, h, "GET", "/backoffice/banners", "", nil, 401)
	productCall(t, h, "GET", "/backoffice/banners", seller, nil, 401)
	attributeCommand(t, h, viewer, backoffice.Command{Action: "banner.create", Name: "Denied", IdempotencyKey: "deny-banner"}, 403)
	attributeCommand(t, h, admin, backoffice.Command{Action: "banner.create", Name: "Responsive test", IdempotencyKey: "create-banner"}, 204)
	var list backoffice.BannerList
	r := productCall(t, h, "GET", "/backoffice/banners", admin, nil, 200)
	_ = json.Unmarshal(r.Body.Bytes(), &list)
	if len(list.Banners) != 1 {
		t.Fatal(list)
	}
	b := list.Banners[0]
	if !strings.Contains(r.Header().Get("X-Robots-Tag"), "noindex") {
		t.Fatal("private preview indexable")
	}
	bannerUploadHTTP(t, h, "", b.ID, "desktop", 1, "no-auth", iconPNG(t, 8), 401)
	bannerUploadHTTP(t, h, viewer, b.ID, "desktop", 1, "viewer", iconPNG(t, 8), 403)
	desktop := bannerUploadHTTP(t, h, admin, b.ID, "desktop", 1, "desktop", iconPNG(t, 16), 200)
	mobile := bannerUploadHTTP(t, h, admin, b.ID, "mobile", 2, "mobile", iconPNG(t, 8), 200)
	productCall(t, h, "GET", "/content/banner-media/"+desktop.AssetID, "", nil, 404)
	productCall(t, h, "GET", "/backoffice/banner-media/"+desktop.AssetID, "", nil, 401)
	productCall(t, h, "GET", "/backoffice/banner-media/"+desktop.AssetID, viewer, nil, 200)
	cfg := content.BannerConfig{DesktopID: desktop.AssetID, MobileID: mobile.AssetID, Alt: "بنر آزمون", Link: "/products/mobile-sample", Enabled: true}
	attributeCommand(t, h, admin, backoffice.Command{Action: "banner.save", Target: b.ID, Version: 3, Banner: &cfg, IdempotencyKey: "save"}, 204)
	before := productCall(t, h, "GET", "/content/banners", "", nil, 200)
	if strings.Contains(before.Body.String(), b.ID) {
		t.Fatal("draft list leak")
	}
	attributeCommand(t, h, admin, backoffice.Command{Action: "banner.publish", Target: b.ID, Version: 4, IdempotencyKey: "publish"}, 204)
	view := productCall(t, h, "GET", "/content/banners/"+b.ID+"/view", "", nil, 200)
	if !strings.Contains(view.Body.String(), `<picture>`) || !strings.Contains(view.Body.String(), "max-width: 767px") || !strings.Contains(view.Body.String(), mobile.AssetID) || !strings.Contains(view.Body.String(), desktop.AssetID) {
		t.Fatal(view.Body.String())
	}
	asset := productCall(t, h, "GET", "/content/banner-media/"+desktop.AssetID, "", nil, 200)
	if asset.Header().Get("Content-Type") != "image/png" || asset.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("unsafe media headers")
	}
	bannerUploadHTTP(t, h, admin, b.ID, "desktop", 5, "invalid", []byte("<svg/>"), 400)
	attributeCommand(t, h, admin, backoffice.Command{Action: "banner.media.remove", Target: b.ID, Version: 5, Value: "mobile", IdempotencyKey: "remove"}, 204)
	attributeCommand(t, h, admin, backoffice.Command{Action: "banner.publish", Target: b.ID, Version: 6, IdempotencyKey: "cannot-publish"}, 400)
	still := productCall(t, h, "GET", "/content/banners/"+b.ID+"/view", "", nil, 200)
	if still.Body.String() != view.Body.String() {
		t.Fatal("failed publication changed public view")
	}
	attributeCommand(t, h, admin, backoffice.Command{Action: "banner.save", Target: b.ID, Version: 6, Banner: &cfg, IdempotencyKey: "inject-reference"}, 400)
}
