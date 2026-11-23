package backoffice_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/content"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func bannerPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	im.SetNRGBA(0, 0, color.NRGBA{R: 180, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func newBanner(t *testing.T, s *backoffice.Service) content.Banner {
	t.Helper()
	ctx := context.Background()
	if err := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "banner.create", Name: "بنر آزمون", IdempotencyKey: "create-banner"}); err != nil {
		t.Fatal(err)
	}
	list, err := s.ReadBanners(ctx, actor(identity.RoleViewer))
	if err != nil || len(list.Banners) != 1 {
		t.Fatal(err, list)
	}
	return list.Banners[0]
}
func readyBanner(t *testing.T, s *backoffice.Service) content.Banner {
	t.Helper()
	ctx := context.Background()
	b := newBanner(t, s)
	for _, slot := range []string{"desktop", "mobile"} {
		result, err := s.SetBannerMedia(ctx, actor(identity.RoleOperations), b.ID, slot, b.Version, "upload-"+slot, bannerPNG(t, 32, 16))
		if err != nil {
			t.Fatal(err)
		}
		b.Version = result.Version
	}
	list, _ := s.ReadBanners(ctx, actor(identity.RoleViewer))
	b = list.Banners[0]
	config := b.Draft
	config.Alt = "تصویر آزمون"
	config.Link = "/products/mobile-sample"
	if err := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "banner.save", Target: b.ID, Version: b.Version, Banner: &config, IdempotencyKey: "save-config"}); err != nil {
		t.Fatal(err)
	}
	list, _ = s.ReadBanners(ctx, actor(identity.RoleViewer))
	return list.Banners[0]
}

func TestBannerDraftPublicationAndMediaIsolation(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	b := readyBanner(t, s)
	ops := actor(identity.RoleOperations)
	if rows, _ := s.PublicBanners(ctx); len(rows) != 0 {
		t.Fatal("draft leaked")
	}
	if _, err := s.ReadBannerMedia(ctx, nil, b.Draft.DesktopID); !errors.Is(err, backoffice.ErrNotFound) {
		t.Fatal("draft image public", err)
	}
	viewer := actor(identity.RoleViewer)
	if _, err := s.ReadBannerMedia(ctx, &viewer, b.Draft.DesktopID); err != nil {
		t.Fatal(err)
	}
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "banner.publish", Target: b.ID, Version: b.Version, IdempotencyKey: "publish-one"}); err != nil {
		t.Fatal(err)
	}
	b.Version++
	rows, _ := s.PublicBanners(ctx)
	if len(rows) != 1 || rows[0].Revision != 1 || rows[0].Config.Alt != "تصویر آزمون" {
		t.Fatal(rows)
	}
	oldID := rows[0].Config.DesktopID
	result, err := s.SetBannerMedia(ctx, ops, b.ID, "desktop", b.Version, "replace-desktop", bannerPNG(t, 64, 24))
	if err != nil {
		t.Fatal(err)
	}
	rows, _ = s.PublicBanners(ctx)
	if rows[0].Config.DesktopID != oldID {
		t.Fatal("draft replacement changed published")
	}
	if _, err := s.ReadBannerMedia(ctx, nil, oldID); err != nil {
		t.Fatal("published asset collected early", err)
	}
	if _, err := s.ReadBannerMedia(ctx, nil, result.AssetID); !errors.Is(err, backoffice.ErrNotFound) {
		t.Fatal("replacement leaked", err)
	}
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "banner.publish", Target: b.ID, Version: result.Version, IdempotencyKey: "publish-two"}); err != nil {
		t.Fatal(err)
	}
	rows, _ = s.PublicBanners(ctx)
	if rows[0].Revision != 2 || rows[0].Config.DesktopID != result.AssetID {
		t.Fatal(rows)
	}
	if _, err := s.ReadBannerMedia(ctx, nil, oldID); !errors.Is(err, backoffice.ErrNotFound) {
		t.Fatal("old orphan remains public", err)
	}
	state, _ := repo.Snapshot(ctx)
	if len(state.BannerAssets) != 2 {
		t.Fatal("orphan not collected", len(state.BannerAssets))
	}
	private, err := s.Read(ctx, viewer)
	if err != nil || len(private.BannerAssets)+len(private.Banners) != 0 {
		t.Fatal("binary/draft leaked in generic workspace", err)
	}
}

func TestBannerInvalidMediaLinksAndMissingAssetCannotPublish(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	b := readyBanner(t, s)
	ops := actor(identity.RoleOperations)
	for i, link := range []string{"javascript:alert(1)", "https://example.test/", "//example.test/", "/%2fexample.test", "/\\example.test", "/%5cexample.test", "/x%0ay", "/ bad"} {
		cfg := b.Draft
		cfg.Link = link
		err := s.Execute(ctx, ops, backoffice.Command{Action: "banner.save", Target: b.ID, Version: b.Version, Banner: &cfg, IdempotencyKey: fmt.Sprintf("bad-link-%d", i)})
		if !errors.Is(err, backoffice.ErrInvalid) {
			t.Fatalf("%q: %v", link, err)
		}
	}
	for i, data := range [][]byte{[]byte("<svg onload='alert(1)'/>"), []byte("not png"), bannerPNG(t, 4097, 1), bytes.Repeat([]byte{0}, backoffice.MaxBannerBytes+1)} {
		_, err := s.SetBannerMedia(ctx, ops, b.ID, "desktop", b.Version, fmt.Sprintf("bad-image-%d", i), data)
		if err == nil {
			t.Fatal("invalid image accepted", i)
		}
	}
	list, _ := s.ReadBanners(ctx, ops)
	if list.Banners[0].Version != b.Version || list.Banners[0].Draft != b.Draft {
		t.Fatal("invalid write changed valid draft")
	}
	backup, _ := repo.Snapshot(ctx)
	remove := backoffice.Command{Action: "banner.media.remove", Target: b.ID, Value: "mobile", Version: b.Version, IdempotencyKey: "remove-mobile"}
	if err := s.Execute(ctx, ops, remove); err != nil {
		t.Fatal(err)
	}
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "banner.publish", Target: b.ID, Version: b.Version + 1, IdempotencyKey: "missing-image"}); !errors.Is(err, backoffice.ErrInvalid) {
		t.Fatal("missing image publish", err)
	}
	_ = repo.Update(ctx, func(state *backoffice.Workspace) error {
		state.BannerAssets = backup.BannerAssets
		state.Banners[0].Draft.MobileID = b.Draft.MobileID
		for i := range state.BannerAssets {
			if state.BannerAssets[i].ID == b.Draft.DesktopID {
				state.BannerAssets[i].Data = []byte("corrupted stored png")
			}
		}
		return nil
	})
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "banner.publish", Target: b.ID, Version: b.Version + 1, IdempotencyKey: "corrupt-image"}); !errors.Is(err, backoffice.ErrInvalid) {
		t.Fatal("corrupt image publish", err)
	}
	if rows, _ := s.PublicBanners(ctx); len(rows) != 0 {
		t.Fatal("invalid publication occurred")
	}
}

func TestBannerDisabledDraftDoesNotHidePublicUntilPublished(t *testing.T) {
	s, _ := testService()
	ctx := context.Background()
	b := readyBanner(t, s)
	ops := actor(identity.RoleOperations)
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "banner.publish", Target: b.ID, Version: b.Version, IdempotencyKey: "enable-public"}); err != nil {
		t.Fatal(err)
	}
	b.Version++
	cfg := b.Draft
	cfg.Enabled = false
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "banner.save", Target: b.ID, Version: b.Version, Banner: &cfg, IdempotencyKey: "disable-draft"}); err != nil {
		t.Fatal(err)
	}
	if rows, _ := s.PublicBanners(ctx); len(rows) != 1 {
		t.Fatal("draft disabled public early")
	}
	if err := s.Execute(ctx, ops, backoffice.Command{Action: "banner.publish", Target: b.ID, Version: b.Version + 1, IdempotencyKey: "disable-public"}); err != nil {
		t.Fatal(err)
	}
	if rows, _ := s.PublicBanners(ctx); len(rows) != 0 {
		t.Fatal("disabled publication still visible")
	}
	if _, err := s.ReadBannerMedia(ctx, nil, b.Draft.DesktopID); !errors.Is(err, backoffice.ErrNotFound) {
		t.Fatal("disabled publication image still public", err)
	}
}

func TestBannerConcurrentPublishRetryAndCurrentPermissions(t *testing.T) {
	for _, sameKey := range []bool{false, true} {
		t.Run(fmt.Sprint(sameKey), func(t *testing.T) {
			s, repo := testService()
			ctx := context.Background()
			b := readyBanner(t, s)
			before, _ := repo.Snapshot(ctx)
			var successes atomic.Int32
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					key := "same-publish"
					if !sameKey {
						key = fmt.Sprint("publish-", i)
					}
					err := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "banner.publish", Target: b.ID, Version: b.Version, IdempotencyKey: key})
					if err == nil {
						successes.Add(1)
					} else if !errors.Is(err, backoffice.ErrConflict) {
						t.Error(err)
					}
				}(i)
			}
			wg.Wait()
			want := int32(1)
			if sameKey {
				want = 8
			}
			if successes.Load() != want {
				t.Fatal(successes.Load())
			}
			after, _ := repo.Snapshot(ctx)
			if len(after.Audit) != len(before.Audit)+1 || after.Banners[0].Published.Revision != 1 {
				t.Fatal("duplicate effect")
			}
		})
	}
	s, repo := testService()
	ctx := context.Background()
	b := newBanner(t, s)
	if _, err := s.SetBannerMedia(ctx, actor(identity.RoleViewer), b.ID, "desktop", b.Version, "viewer", bannerPNG(t, 8, 8)); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal(err)
	}
	if _, err := s.ReadBanners(ctx, sellerActor()); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("seller read", err)
	}
	_ = repo.Update(ctx, func(state *backoffice.Workspace) error {
		for i := range state.Users {
			if state.Users[i].ID == "ops" {
				state.Users[i].Role = identity.RoleViewer
			}
		}
		return nil
	})
	if _, err := s.SetBannerMedia(ctx, actor(identity.RoleOperations), b.ID, "desktop", b.Version, "revoked", bannerPNG(t, 8, 8)); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("stale actor", err)
	}
}

func TestBannerUploadRetryReturnsOriginalReceipt(t *testing.T) {
	s, _ := testService()
	ctx := context.Background()
	b := newBanner(t, s)
	data := bannerPNG(t, 8, 8)
	ops := actor(identity.RoleOperations)
	first, err := s.SetBannerMedia(ctx, ops, b.ID, "desktop", 1, "receipt", data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SetBannerMedia(ctx, ops, b.ID, "mobile", 2, "next-media", data)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := s.SetBannerMedia(ctx, ops, b.ID, "desktop", 1, "receipt", data)
	if err != nil || retry != first {
		t.Fatal("retry silently advanced expected version", err, retry, first)
	}
	if _, err := s.SetBannerMedia(ctx, ops, b.ID, "mobile", 1, "receipt", data); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("key payload conflict", err)
	}
}

func TestCatalogPromoUsesConfiguredPublishedRailWithoutStaticImages(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Categories = append(w.Categories, catalog.Category{ID: "promo-watch", Name: "ساعت", Published: true})
		w.HomeSettings = backoffice.DefaultHomeSettings()
		w.HomeSettings.Rails[3].CategoryID = "promo-watch"
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	b := newBanner(t, s)
	config := content.BannerConfig{Placement: "promo", CatalogCategoryID: "promo-watch", Alt: "ساعت‌های موجود", Link: "/category/promo-watch", Enabled: true}
	if e := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "banner.save", Target: b.ID, Version: b.Version, Banner: &config, IdempotencyKey: "catalog-promo-save"}); e != nil {
		t.Fatal(e)
	}
	if e := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "banner.publish", Target: b.ID, Version: b.Version + 1, IdempotencyKey: "catalog-promo-publish"}); e != nil {
		t.Fatal(e)
	}
	public, e := s.PublicBanners(ctx)
	if e != nil || len(public) != 1 || public[0].Config.CatalogCategoryID != "promo-watch" {
		t.Fatal("missing data-driven config", public, e)
	}
	config.Placement = "hero"
	if e := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "banner.save", Target: b.ID, Version: b.Version + 2, Banner: &config, IdempotencyKey: "catalog-promo-wrong-placement"}); !errors.Is(e, backoffice.ErrInvalid) {
		t.Fatal("catalog promo accepted as hero", e)
	}
	config.Placement = "promo"
	config.CatalogCategoryID = "missing"
	if e := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "banner.save", Target: b.ID, Version: b.Version + 2, Banner: &config, IdempotencyKey: "catalog-promo-missing-category"}); !errors.Is(e, backoffice.ErrInvalid) {
		t.Fatal("missing category accepted", e)
	}
}
