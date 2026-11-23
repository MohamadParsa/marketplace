package backoffice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/content"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const MaxBannerBytes = 5 << 20
const MaxBannerDimension = 4096
const MaxBannerPixels = 8 * 1024 * 1024

type BannerPolicy struct {
	MaxBytes     int      `json:"max_bytes"`
	MaxDimension int      `json:"max_dimension"`
	MaxPixels    int      `json:"max_pixels"`
	ContentTypes []string `json:"content_types"`
}
type BannerList struct {
	Banners []content.Banner `json:"banners"`
	CanEdit bool             `json:"can_edit"`
	Policy  BannerPolicy     `json:"policy"`
}
type BannerMediaResult struct {
	AssetID string `json:"asset_id"`
	Version int    `json:"version"`
}

func bannerActor(state Workspace, actor identity.User, write bool) error {
	current, err := catalogActor(state, actor)
	if err != nil || current.Surface == identity.SurfaceSeller {
		return ErrForbidden
	}
	if !write && current.Role == "" && !allowsAction(current, "banner.read") {
		return ErrForbidden
	}
	if write && !allowsAction(current, "banner.save") {
		return ErrForbidden
	}
	return nil
}

func (s *Service) ReadBanners(ctx context.Context, actor identity.User) (BannerList, error) {
	state, err := s.repository.Snapshot(ctx)
	if err != nil {
		return BannerList{}, err
	}
	if err := bannerActor(state, actor, false); err != nil {
		return BannerList{}, err
	}
	current, _ := catalogActor(state, actor)
	rows := state.Banners
	if rows == nil {
		rows = []content.Banner{}
	}
	return BannerList{Banners: rows, CanEdit: allowsAction(current, "banner.save"), Policy: BannerPolicy{MaxBannerBytes, MaxBannerDimension, MaxBannerPixels, []string{"image/png", "image/jpeg", "image/webp"}}}, nil
}

func findBanner(state *Workspace, id string) (*content.Banner, error) {
	for i := range state.Banners {
		if state.Banners[i].ID == id {
			return &state.Banners[i], nil
		}
	}
	return nil, ErrNotFound
}

// Only an internal path is allowed in this test slice; never accept executable or protocol-relative URLs.
func validBannerLink(link string) bool {
	if link == "" {
		return true
	}
	if len(link) > 2048 || !strings.HasPrefix(link, "/") {
		return false
	}
	decoded, err := url.PathUnescape(link)
	if err != nil {
		return false
	}
	if strings.HasPrefix(decoded, "//") || strings.Contains(decoded, "\\") || strings.IndexFunc(decoded, func(r rune) bool { return unicode.IsControl(r) || unicode.IsSpace(r) }) >= 0 {
		return false
	}
	u, err := url.ParseRequestURI(link)
	return err == nil && u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//")
}

func validBannerConfig(config content.BannerConfig) bool {
	if config.Placement != "" && config.Placement != "hero" && config.Placement != "promo" {
		return false
	}
	return len(config.CatalogCategoryID) <= 128 && utf8.RuneCountInString(config.Alt) <= 200 && !strings.ContainsAny(config.Alt, "\x00\r\n") && validBannerLink(config.Link)
}

func bannerOwnsAsset(b content.Banner, id string) bool {
	return id != "" && (b.Draft.DesktopID == id || b.Draft.MobileID == id || (b.Published != nil && (b.Published.Config.DesktopID == id || b.Published.Config.MobileID == id)))
}

func collectBannerAssets(state *Workspace) {
	kept := make([]content.BannerAsset, 0, len(state.BannerAssets))
	for _, asset := range state.BannerAssets {
		for _, b := range state.Banners {
			if bannerOwnsAsset(b, asset.ID) {
				kept = append(kept, asset)
				break
			}
		}
	}
	state.BannerAssets = kept
}

func applyBanner(state *Workspace, c Command, id string) error {
	if c.Action == "banner.create" {
		name := strings.TrimSpace(c.Name)
		if name == "" || utf8.RuneCountInString(name) > 80 {
			return ErrInvalid
		}
		state.Banners = append(state.Banners, content.Banner{ID: id, Name: name, Version: 1, Draft: content.BannerConfig{Enabled: true}})
		return nil
	}
	b, err := findBanner(state, c.Target)
	if err != nil {
		return err
	}
	if b.Version != c.Version {
		return ErrConflict
	}
	switch c.Action {
	case "banner.save":
		if c.Banner == nil || !validBannerConfig(*c.Banner) {
			return ErrInvalid
		}
		if !validCatalogBanner(*state, *c.Banner) {
			return ErrInvalid
		}
		// Image references are changed exclusively by the upload/remove operations.
		if c.Banner.DesktopID != b.Draft.DesktopID || c.Banner.MobileID != b.Draft.MobileID {
			return ErrInvalid
		}
		b.Draft = *c.Banner
	case "banner.media.remove":
		if c.Value == "desktop" {
			b.Draft.DesktopID = ""
		} else if c.Value == "mobile" {
			b.Draft.MobileID = ""
		} else {
			return ErrInvalid
		}
	case "banner.publish":
		if !validBannerConfig(b.Draft) || !validCatalogBanner(*state, b.Draft) || strings.TrimSpace(b.Draft.Alt) == "" || (b.Draft.CatalogCategoryID == "" && (b.Draft.DesktopID == "" || b.Draft.MobileID == "")) {
			return ErrInvalid
		}
		refs := []string{}
		if b.Draft.CatalogCategoryID == "" {
			refs = []string{b.Draft.DesktopID, b.Draft.MobileID}
		}
		for _, ref := range refs {
			found := false
			for _, asset := range state.BannerAssets {
				if asset.ID == ref && validStoredBannerAsset(asset) {
					found = true
					break
				}
			}
			if !found {
				return ErrInvalid
			}
		}
		revision := 1
		if b.Published != nil {
			revision = b.Published.Revision + 1
		}
		b.Published = &content.PublishedBanner{Config: b.Draft, Revision: revision}
	default:
		return ErrInvalid
	}
	b.Version++
	collectBannerAssets(state)
	return nil
}

func (s *Service) SetBannerMedia(ctx context.Context, actor identity.User, target, slot string, version int, key string, input []byte) (BannerMediaResult, error) {
	if actor.Surface == identity.SurfaceSeller || !allowsAction(actor, "banner.media.set") {
		return BannerMediaResult{}, ErrForbidden
	}
	if (slot != "desktop" && slot != "mobile") || strings.TrimSpace(key) == "" || len(key) > 128 || version < 1 {
		return BannerMediaResult{}, ErrInvalid
	}
	data, err := normalizeRaster(input, MaxBannerBytes, MaxBannerDimension, MaxBannerPixels)
	if err != nil {
		return BannerMediaResult{}, err
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return BannerMediaResult{}, ErrInvalid
	}
	id, err := s.identifiers.Token()
	if err != nil {
		return BannerMediaResult{}, err
	}
	digest := sha256.Sum256(input)
	encoded, _ := json.Marshal(struct {
		Target, Slot, Digest string
		Version              int
	}{target, slot, hex.EncodeToString(digest[:]), version})
	sum := sha256.Sum256(encoded)
	fingerprint := hex.EncodeToString(sum[:])
	var result BannerMediaResult
	err = s.repository.Update(ctx, func(state *Workspace) error {
		current, err := catalogActor(*state, actor)
		if err != nil || current.Surface == identity.SurfaceSeller || !allowsAction(current, "banner.media.set") {
			return ErrForbidden
		}
		if err := bannerActor(*state, current, false); err != nil {
			return err
		}
		b, err := findBanner(state, target)
		if err != nil {
			return err
		}
		for _, event := range state.Audit {
			if event.ID == key {
				if event.Actor == actor.ID && event.Action == "banner.media.set" && event.Target == target && event.Fingerprint == fingerprint {
					result = BannerMediaResult{AssetID: event.ResultID, Version: version + 1}
					return nil
				}
				return ErrConflict
			}
		}
		if b.Version != version {
			return ErrConflict
		}
		if slot == "desktop" {
			b.Draft.DesktopID = id
		} else {
			b.Draft.MobileID = id
		}
		b.Version++
		state.BannerAssets = append(state.BannerAssets, content.BannerAsset{ID: id, Data: data, Width: cfg.Width, Height: cfg.Height})
		collectBannerAssets(state)
		state.Audit = append(state.Audit, Audit{ID: key, Actor: actor.ID, Action: "banner.media.set", Target: target, Fingerprint: fingerprint, At: s.clock.Now(), ResultID: id})
		result = BannerMediaResult{AssetID: id, Version: b.Version}
		return nil
	})
	return result, err
}

func validStoredBannerAsset(asset content.BannerAsset) bool {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(asset.Data))
	if err != nil || format != "png" || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > MaxBannerDimension || cfg.Height > MaxBannerDimension || cfg.Width > MaxBannerPixels/cfg.Height {
		return false
	}
	_, _, err = image.Decode(bytes.NewReader(asset.Data))
	return err == nil
}

func (s *Service) PublicBanners(ctx context.Context) ([]content.PublicBanner, error) {
	return cachedPublic(ctx, s, "banners", "", nil, 5*time.Minute, func(ctx context.Context) ([]content.PublicBanner, error) {
		state, err := s.publicSnapshot(ctx, "__shared__")
		if err != nil {
			return nil, err
		}
		result := []content.PublicBanner{}
		for _, b := range state.Banners {
			if b.Published != nil && b.Published.Config.Enabled {
				result = append(result, content.PublicBanner{ID: b.ID, Name: b.Name, Revision: b.Published.Revision, Config: b.Published.Config})
			}
		}
		return result, nil
	})
}

func (s *Service) ReadBannerMedia(ctx context.Context, actor *identity.User, id string) (content.BannerAsset, error) {
	if actor == nil {
		if r, ok := s.repository.(interface {
			ReadPublicBannerMedia(context.Context, string) (content.BannerAsset, error)
		}); ok {
			return r.ReadPublicBannerMedia(ctx, id)
		}
	}
	state, err := s.repository.Snapshot(ctx)
	if err != nil {
		return content.BannerAsset{}, err
	}
	if actor != nil {
		if err := bannerActor(state, *actor, false); err != nil {
			return content.BannerAsset{}, err
		}
	}
	allowed := false
	for _, b := range state.Banners {
		if actor != nil {
			allowed = allowed || bannerOwnsAsset(b, id)
		} else if b.Published != nil && b.Published.Config.Enabled {
			allowed = allowed || (b.Published.Config.DesktopID == id || b.Published.Config.MobileID == id)
		}
	}
	if id == "" || !allowed {
		return content.BannerAsset{}, ErrNotFound
	}
	for _, asset := range state.BannerAssets {
		if asset.ID == id {
			return asset, nil
		}
	}
	return content.BannerAsset{}, ErrNotFound
}

// Catalog promos deliberately use one of the four configured rails in this MVP.
func validCatalogBanner(w Workspace, c content.BannerConfig) bool {
	if c.CatalogCategoryID == "" {
		return true
	}
	if c.Placement != "promo" || publicCategory(w, c.CatalogCategoryID) == nil {
		return false
	}
	for _, r := range w.HomeSettings.Rails {
		if r.CategoryID == c.CatalogCategoryID {
			return true
		}
	}
	return false
}
