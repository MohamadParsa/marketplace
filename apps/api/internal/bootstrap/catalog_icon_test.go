package bootstrap_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
)

func iconPNG(t *testing.T, size int) []byte {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, size, size))
	im.SetNRGBA(0, 0, color.NRGBA{R: 170, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func iconUpload(t *testing.T, h http.Handler, token, kind, id string, version int, key string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	form := multipart.NewWriter(&b)
	_ = form.WriteField("version", strconv.Itoa(version))
	_ = form.WriteField("idempotency_key", key)
	file, err := form.CreateFormFile("file", "untrusted-name.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.Write(data)
	_ = form.Close()
	r := httptest.NewRequest("POST", "/backoffice/catalog/"+kind+"/"+id+"/icon", &b)
	r.Header.Set("Content-Type", form.FormDataContentType())
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func iconWorkspace(t *testing.T, h http.Handler, token string) backoffice.Workspace {
	t.Helper()
	r := httptest.NewRequest("GET", "/backoffice/workspace", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("workspace %d", w.Code)
	}
	if strings.Contains(w.Body.String(), `"icon_media"`) {
		t.Fatal("icon binary storage leaked")
	}
	var state backoffice.Workspace
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func TestCatalogIconHTTPRoundTripAndRemoval(t *testing.T) {
	for _, kind := range []string{"category", "brand"} {
		t.Run(kind, func(t *testing.T) {
			h := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
			token := login(t, h, "09000000002")
			adminToken := login(t, h, "09000000001")
			state := iconWorkspace(t, h, token)
			id, version := state.Categories[0].ID, state.Categories[0].Version
			if kind == "brand" {
				id, version = state.Brands[0].ID, state.Brands[0].Version
			}
			data := append(iconPNG(t, 8), []byte("<script>must-not-be-served</script>")...)
			for range 2 {
				if got := iconUpload(t, h, token, kind, id, version, "upload-once", data); got.Code != 204 {
					t.Fatalf("upload %d: %s", got.Code, got.Body.String())
				}
			}
			state = iconWorkspace(t, h, token)
			icon, next := state.Categories[0].IconID, state.Categories[0].Version
			if kind == "brand" {
				icon, next = state.Brands[0].IconID, state.Brands[0].Version
			}
			if len(state.Audit) != 0 {
				t.Fatal("operator received private audit")
			}
			auditState := iconWorkspace(t, h, adminToken)
			if icon == "" || next != version+1 || len(auditState.Audit) != 1 {
				t.Fatalf("attachment/version/retry failed: %s/%d/%d", icon, next, len(state.Audit))
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", "/catalog/icons/"+icon, nil))
			if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" || bytes.Contains(w.Body.Bytes(), []byte("must-not-be-served")) {
				t.Fatal("unsafe or missing public image")
			}
			if _, err := png.Decode(bytes.NewReader(w.Body.Bytes())); err != nil {
				t.Fatal(err)
			}
			if bad := iconUpload(t, h, token, kind, id, version, "stale", iconPNG(t, 8)); bad.Code != 409 {
				t.Fatalf("stale: %d", bad.Code)
			}
			if bad := iconUpload(t, h, token, kind, id, next, "invalid", []byte("<svg onload='bad'/>")); bad.Code != 400 {
				t.Fatalf("invalid: %d", bad.Code)
			}
			state = iconWorkspace(t, h, token)
			preserved := state.Categories[0].IconID
			if kind == "brand" {
				preserved = state.Brands[0].IconID
			}
			if preserved != icon {
				t.Fatal("failed write lost previous icon")
			}
			body, _ := json.Marshal(backoffice.Command{Action: kind + ".icon.remove", Target: id, Version: next, IdempotencyKey: "remove-once"})
			for range 2 {
				r := httptest.NewRequest("POST", "/backoffice/commands", bytes.NewReader(body))
				r.Header.Set("Authorization", "Bearer "+token)
				w = httptest.NewRecorder()
				h.ServeHTTP(w, r)
				if w.Code != 204 {
					t.Fatalf("remove %d", w.Code)
				}
			}
			state = iconWorkspace(t, h, token)
			remaining := state.Categories[0].IconID
			if kind == "brand" {
				remaining = state.Brands[0].IconID
			}
			auditState = iconWorkspace(t, h, adminToken)
			if remaining != "" || len(auditState.Audit) != 2 {
				t.Fatal("remove/retry failed")
			}
			w = httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", "/catalog/icons/"+icon, nil))
			if w.Code != 404 {
				t.Fatal("removed media still exposed")
			}
		})
	}
}

func TestCatalogIconPermissionsAndValidation(t *testing.T) {
	h := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	data := iconPNG(t, 8)
	for _, tc := range []struct {
		name, token string
		data        []byte
		status      int
	}{
		{"anonymous", "", data, 401}, {"viewer", viewer, data, 403},
		{"fake-png", admin, []byte("not a PNG despite filename"), 400},
		{"truncated", admin, data[:len(data)/2], 400},
		{"oversize", admin, make([]byte, backoffice.MaxIconBytes+1), 413},
		{"dimensions", admin, iconPNG(t, backoffice.MaxIconDimension+1), 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := iconUpload(t, h, tc.token, "category", "c1", 0, tc.name, tc.data)
			if w.Code != tc.status {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
		})
	}
	state := iconWorkspace(t, h, admin)
	if len(state.Audit) != 0 || state.Categories[0].IconID != "" {
		t.Fatal("rejected upload changed state")
	}
	body := bytes.NewBufferString(`{"action":"brand.icon.remove","target":"b1","version":1,"idempotency_key":"viewer-delete"}`)
	r := httptest.NewRequest("POST", "/backoffice/commands", body)
	r.Header.Set("Authorization", "Bearer "+viewer)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatalf("viewer remove %d", w.Code)
	}
}

func TestCatalogIconAcceptsRasterFormats(t *testing.T) {
	h := bootstrap.NewHTTPHandler(config.Config{Environment: "test"})
	token := login(t, h, "09000000001")
	var jpegData bytes.Buffer
	if err := jpeg.Encode(&jpegData, image.NewRGBA(image.Rect(0, 0, 8, 8)), nil); err != nil {
		t.Fatal(err)
	}
	webpData, err := base64.StdEncoding.DecodeString("UklGRrICAABXRUJQVlA4WAoAAAAgAAAAHwAAHwAASUNDUMgBAAAAAAHIAAAAAAQwAABtbnRyUkdCIFhZWiAH4AABAAEAAAAAAABhY3NwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAA9tYAAQAAAADTLQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAlkZXNjAAAA8AAAACRyWFlaAAABFAAAABRnWFlaAAABKAAAABRiWFlaAAABPAAAABR3dHB0AAABUAAAABRyVFJDAAABZAAAAChnVFJDAAABZAAAAChiVFJDAAABZAAAAChjcHJ0AAABjAAAADxtbHVjAAAAAAAAAAEAAAAMZW5VUwAAAAgAAAAcAHMAUgBHAEJYWVogAAAAAAAAb6IAADj1AAADkFhZWiAAAAAAAABimQAAt4UAABjaWFlaIAAAAAAAACSgAAAPhAAAts9YWVogAAAAAAAA9tYAAQAAAADTLXBhcmEAAAAAAAQAAAACZmYAAPKnAAANWQAAE9AAAApbAAAAAAAAAABtbHVjAAAAAAAAAAEAAAAMZW5VUwAAACAAAAAcAEcAbwBvAGcAbABlACAASQBuAGMALgAgADIAMAAxADZWUDggxAAAABAHAJ0BKiAAIAAuTTabTaEkJCQEAExLUAO4L4B+AHIGaAhADeAPsALYEyx+rW6AHG10ggYlNskvSZ3joWIWnacAAP73ccv4v/Tvc+d581kG3p09CQETU8Sk4e77y///aWdaHrLOpXPybIgS9zDMdxmmfn+uJqMtX4xCSYVfP3sVdqtW8yMwe3uT9j/cWIfChVbjIkY1Lp7RI6SpLFregNMZwh5iq+Epy+YzS+eGt5SnmbGyOnjp2c+2rXz3xGYF3bOIAAA=")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		data []byte
	}{{"png", iconPNG(t, 8)}, {"jpeg", jpegData.Bytes()}, {"webp", webpData}} {
		t.Run(tc.name, func(t *testing.T) {
			if w := iconUpload(t, h, token, "brand", "b1", iconWorkspace(t, h, token).Brands[0].Version, "format-"+tc.name, tc.data); w.Code != 204 {
				t.Fatalf("format %s: %d %s", tc.name, w.Code, w.Body.String())
			}
			state := iconWorkspace(t, h, token)
			body, _ := json.Marshal(backoffice.Command{Action: "brand.icon.remove", Target: "b1", Version: state.Brands[0].Version, IdempotencyKey: "remove-" + tc.name})
			r := httptest.NewRequest("POST", "/backoffice/commands", bytes.NewReader(body))
			r.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 204 {
				t.Fatal(w.Code)
			}
		})
	}
}
