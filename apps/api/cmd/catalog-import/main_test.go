package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
)

func TestMediaCacheEnforcesChecksumsAndFullDecode(t *testing.T) {
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 3, 3)))
	raw := b.Bytes()
	m := backoffice.SeedMedia{ID: "gsp-media-1", URL: "https://cdn.gooshishop.ir/images/thumbs/n/Entity/test.png"}
	hash := sha256.Sum256([]byte(m.URL))
	cache := t.TempDir()
	file := filepath.Join(cache, hex.EncodeToString(hash[:])+".source")
	if e := os.WriteFile(file, raw, 0600); e != nil {
		t.Fatal(e)
	}
	sum := sha256.Sum256(raw)
	m.SHA256 = hex.EncodeToString(sum[:])
	if _, e := loadMedia(context.Background(), &http.Client{}, m, cache); e != nil {
		t.Fatal(e)
	}
	m.SHA256 = hex.EncodeToString(make([]byte, 32))
	if _, e := loadMedia(context.Background(), &http.Client{}, m, cache); e == nil {
		t.Fatal("changed source image accepted")
	}
	m.SHA256 = ""
	_ = os.WriteFile(file, raw[:len(raw)/2], 0600)
	if _, e := loadMedia(context.Background(), &http.Client{}, m, cache); e == nil {
		t.Fatal("truncated image accepted")
	}
	_ = os.WriteFile(file, []byte("<svg><script>alert(1)</script></svg>"), 0600)
	if _, e := loadMedia(context.Background(), &http.Client{}, m, cache); e == nil {
		t.Fatal("SVG accepted")
	}
}
