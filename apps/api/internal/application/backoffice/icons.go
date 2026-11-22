package backoffice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg"
	"image/png"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const MaxIconBytes = 1 << 20
const MaxIconDimension = 1024
const MaxProductImageBytes = 5 << 20

var ErrIconTooLarge = errors.New("icon_too_large")
var ErrInvalidIcon = errors.New("invalid_icon")

type IconPolicy struct {
	MaxBytes     int      `json:"max_bytes"`
	MaxDimension int      `json:"max_dimension"`
	ContentTypes []string `json:"content_types"`
}

func CatalogIconPolicy() *IconPolicy {
	return &IconPolicy{MaxIconBytes, MaxIconDimension, []string{"image/png", "image/jpeg", "image/webp"}}
}

// Icons are decoded and re-encoded, stripping filenames, metadata and trailing content.
func normalizeIcon(input []byte) ([]byte, error) {
	return normalizeRaster(input, MaxIconBytes, MaxIconDimension, MaxIconDimension*MaxIconDimension)
}

func normalizeRaster(input []byte, maxBytes, maxDimension, maxPixels int) ([]byte, error) {
	if len(input) > maxBytes {
		return nil, ErrIconTooLarge
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(input))
	if err != nil || (format != "png" && format != "jpeg" && format != "webp") || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > maxDimension || cfg.Height > maxDimension || cfg.Width > maxPixels/cfg.Height {
		return nil, ErrInvalidIcon
	}
	decoded, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, ErrInvalidIcon
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, decoded); err != nil {
		return nil, ErrInvalidIcon
	}
	return encoded.Bytes(), nil
}

func iconReference(state *Workspace, kind, target string) (*string, *int, error) {
	switch kind {
	case "product-gallery":
		p, err := findProduct(state, target)
		if err != nil {
			return nil, nil, err
		}
		if len(p.GalleryIDs) >= MaxGalleryImages {
			return nil, nil, ErrInvalid
		}
		p.GalleryIDs = append(p.GalleryIDs, "")
		return &p.GalleryIDs[len(p.GalleryIDs)-1], &p.Version, nil
	case "product":
		for i := range state.Products {
			p := &state.Products[i]
			if p.ID == target {
				return &p.ImageID, &p.Version, nil
			}
		}
		return nil, nil, ErrNotFound
	case "category":
		for i := range state.Categories {
			x := &state.Categories[i]
			if x.ID == target {
				if x.Archived {
					return nil, nil, ErrConflict
				}
				return &x.IconID, &x.Version, nil
			}
		}
	case "brand":
		for i := range state.Brands {
			x := &state.Brands[i]
			if x.ID == target {
				if !x.Active {
					return nil, nil, ErrConflict
				}
				return &x.IconID, &x.Version, nil
			}
		}
	default:
		return nil, nil, ErrInvalid
	}
	return nil, nil, ErrNotFound
}

func collectUnusedIcons(state *Workspace) {
	used := map[string]bool{}
	for _, v := range state.Variants {
		used[v.ImageID] = true
	}
	for _, p := range state.Products {
		for _, id := range p.GalleryIDs {
			used[id] = true
		}
		if p.ImageID != "" {
			used[p.ImageID] = true
		}
	}
	for _, x := range state.Categories {
		used[x.IconID] = true
	}
	for _, x := range state.Brands {
		used[x.IconID] = true
	}
	kept := make([]catalog.IconMedia, 0, len(state.IconMedia))
	for _, x := range state.IconMedia {
		if used[x.ID] {
			kept = append(kept, x)
		}
	}
	state.IconMedia = kept
}

func removeIcon(state *Workspace, c Command) error {
	kind := strings.TrimSuffix(c.Action, ".icon.remove")
	ref, version, err := iconReference(state, kind, c.Target)
	if err != nil {
		return err
	}
	if *version != c.Version {
		return ErrConflict
	}
	*ref = ""
	*version = *version + 1
	collectUnusedIcons(state)
	return nil
}

// SetCatalogIcon attaches the validated icon and audit in the same repository transaction.
func (s *Service) SetCatalogIcon(ctx context.Context, actor identity.User, kind, target string, version int, key string, input []byte) error {
	action := kind + ".icon.set"
	if kind == "product-gallery" {
		action = "product.icon.set"
	}
	if !allowsAction(actor, action) {
		return ErrForbidden
	}
	if (kind != "category" && kind != "brand" && kind != "product" && kind != "product-gallery") || target == "" || version < 0 || strings.TrimSpace(key) == "" || len(key) > 128 {
		return ErrInvalid
	}
	var data []byte
	var err error
	if kind == "product" || kind == "product-gallery" {
		data, err = normalizeProductPhoto(input)
	} else {
		data, err = normalizeIcon(input)
	}
	if err != nil {
		return err
	}
	id, err := s.identifiers.Token()
	if err != nil {
		return err
	}
	digest := sha256.Sum256(input)
	encoded, _ := json.Marshal(struct {
		Kind, Target, Digest string
		Version              int
	}{kind, target, hex.EncodeToString(digest[:]), version})
	sum := sha256.Sum256(encoded)
	fingerprint := hex.EncodeToString(sum[:])
	return s.repository.Update(ctx, func(state *Workspace) error {
		current, err := currentActor(*state, actor)
		if err != nil {
			return err
		}
		if !allowsAction(current, action) {
			return ErrForbidden
		}
		for _, event := range state.Audit {
			if event.ID == key {
				if event.Actor == actor.ID && event.Action == action && event.Target == target && event.Fingerprint == fingerprint {
					return nil
				}
				return ErrConflict
			}
		}
		ref, currentVersion, err := iconReference(state, kind, target)
		if err != nil {
			return err
		}
		if *currentVersion != version {
			return ErrConflict
		}
		*ref = id
		*currentVersion = *currentVersion + 1
		state.IconMedia = append(state.IconMedia, catalog.IconMedia{ID: id, Data: data})
		collectUnusedIcons(state)
		state.Audit = append(state.Audit, Audit{ID: key, Actor: actor.ID, Action: action, Target: target, Fingerprint: fingerprint, At: s.clock.Now()})
		return nil
	})
}

// ReadCatalogIcon exposes only images attached to an active public catalog record.
func (s *Service) ReadCatalogIcon(ctx context.Context, id string) ([]byte, error) {
	state, err := s.repository.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	attached := false
	for _, p := range state.Products {
		if p.Status == "published" && productHasImage(state, p, id) && publicCategory(state, p.CategoryID) != nil && validBrand(state, p.BrandID) {
			attached = true
		}
	}
	for _, x := range state.Categories {
		attached = attached || (!x.Archived && x.IconID == id)
	}
	for _, x := range state.Brands {
		attached = attached || (x.Active && x.IconID == id)
	}
	if id == "" || !attached {
		return nil, ErrNotFound
	}
	for _, x := range state.IconMedia {
		if x.ID == id {
			return x.Data, nil
		}
	}
	return nil, ErrNotFound
}

func (s *Service) ReadManagedIcon(ctx context.Context, actor identity.User, id string) ([]byte, error) {
	state, err := s.repository.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	current, err := currentActor(state, actor)
	if err != nil || !allowsAction(current, "product.icon.set") {
		return nil, ErrForbidden
	}
	for _, m := range state.IconMedia {
		if m.ID == id {
			return m.Data, nil
		}
	}
	return nil, ErrNotFound
}

func normalizeProductPhoto(input []byte) ([]byte, error) {
	if len(input) > MaxProductImageBytes {
		return nil, ErrIconTooLarge
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(input))
	if err != nil || (format != "png" && format != "jpeg" && format != "webp") || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 4096 || cfg.Height > 4096 || cfg.Width > (16*1024*1024)/cfg.Height {
		return nil, ErrInvalidIcon
	}
	source, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, ErrInvalidIcon
	}
	var normalized image.Image = source
	if cfg.Width > 1024 || cfg.Height > 1024 {
		width, height := cfg.Width, cfg.Height
		if width >= height {
			height = height * 1024 / width
			width = 1024
		} else {
			width = width * 1024 / height
			height = 1024
		}
		if width < 1 {
			width = 1
		}
		if height < 1 {
			height = 1
		}
		target := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.ApproxBiLinear.Scale(target, target.Bounds(), source, source.Bounds(), draw.Src, nil)
		normalized = target
	}
	var out bytes.Buffer
	if err := png.Encode(&out, normalized); err != nil {
		return nil, ErrInvalidIcon
	}
	return out.Bytes(), nil
}

func productHasImage(w Workspace, p catalog.Product, id string) bool {
	if p.ImageID == id {
		return true
	}
	for _, x := range p.GalleryIDs {
		if x == id {
			return true
		}
	}
	for _, v := range w.Variants {
		if v.ProductID == p.ID && v.ImageID == id {
			return true
		}
	}
	return false
}
