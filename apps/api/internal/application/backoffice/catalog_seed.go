package backoffice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
)

// CatalogSeed intentionally has no users, permissions, commerce or bank-account fields.
// SourceRecords and other extraction metadata are not applied to the target workspace.
type CatalogSeed struct {
	SchemaVersion int         `json:"schema_version"`
	DatasetID     string      `json:"dataset_id"`
	Currency      string      `json:"currency"`
	Namespace     string      `json:"namespace"`
	Purpose       string      `json:"purpose"`
	Catalog       SeedCatalog `json:"catalog"`
	Media         []SeedMedia `json:"media"`
}
type SeedCatalog struct {
	Categories   []catalog.Category            `json:"categories"`
	Brands       []catalog.Brand               `json:"brands"`
	Attributes   []catalog.AttributeDefinition `json:"attributes"`
	Products     []catalog.Product             `json:"products"`
	Variants     []catalog.Variant             `json:"variants"`
	Offers       []catalog.Offer               `json:"offers"`
	Stores       []store.Store                 `json:"stores"`
	HomeSettings HomeSettings                  `json:"home_settings"`
}
type SeedMedia struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}
type SeedImportResult struct {
	DatasetID      string `json:"dataset_id"`
	Products       int    `json:"products"`
	Variants       int    `json:"variants"`
	Offers         int    `json:"offers"`
	Media          int    `json:"media"`
	AlreadyApplied bool   `json:"already_applied"`
}

var seedID = regexp.MustCompile(`^gsp-[a-zA-Z0-9-]{1,120}$`)

func DecodeCatalogSeed(data []byte) (CatalogSeed, error) {
	var s CatalogSeed
	if len(data) > 128<<20 {
		return s, fmt.Errorf("seed exceeds 128 MiB")
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("invalid catalog seed JSON")
	}
	// The portable file declares IRT (toman); catalog.Price uses rial internally.
	// Convert only at the decode boundary, never on retries or workspace writes.
	if s.Currency == "IRT" {
		for i := range s.Catalog.Offers {
			o := &s.Catalog.Offers[i]
			if o.Price < 0 || o.RegularPrice < 0 || o.Price > MaxOfferPrice/10 || o.RegularPrice > MaxOfferPrice/10 {
				return s, fmt.Errorf("seed toman amount exceeds internal price bound")
			}
			o.Price *= 10
			o.RegularPrice *= 10
		}
	}
	return s, nil
}
func seedURL(raw string) bool {
	u, e := url.Parse(raw)
	return e == nil && u.Scheme == "https" && u.Host == "cdn.gooshishop.ir" && u.User == nil && u.RawQuery == "" && u.Fragment == "" && strings.HasPrefix(u.Path, "/images/thumbs/")
}

// ValidateCatalogSeed applies the current editor's text, content, attribute and price bounds.
// It rejects broken relations and duplicate purchasable selections before any database write.
func ValidateCatalogSeed(s *CatalogSeed) error {
	bad := func(where string) error { return fmt.Errorf("invalid seed: %s", where) }
	if s.SchemaVersion != 1 || s.Currency != "IRT" || s.Namespace != "gsp-" || s.Purpose != "test-catalog-snapshot" || len(s.DatasetID) < 1 || len(s.DatasetID) > 100 || len(s.Catalog.Products) == 0 || len(s.Catalog.Products) > 10000 || len(s.Media) > 50000 {
		return bad("header or bounds")
	}
	w := s.Catalog.workspace()
	ids := map[string]bool{}
	add := func(id string) bool {
		if !seedID.MatchString(id) || ids[id] {
			return false
		}
		ids[id] = true
		return true
	}
	media := map[string]bool{}
	for _, m := range s.Media {
		if !add(m.ID) || !seedURL(m.URL) {
			return bad("media " + m.ID)
		}
		if m.SHA256 != "" {
			b, e := hex.DecodeString(m.SHA256)
			if e != nil || len(b) != 32 {
				return bad("media hash " + m.ID)
			}
		}
		media[m.ID] = true
	}
	imageOK := func(id string) bool { return id == "" || media[id] }
	cats := map[string]catalog.Category{}
	for _, c := range w.Categories {
		if !add(c.ID) || !validCatalogText(CatalogInput{Name: c.Name, EnglishName: c.EnglishName, Description: c.Description}) || !validCategorySEO(c.SEO) || c.Version != 1 || !imageOK(c.IconID) || len(c.Bindings) > 100 {
			return bad("category " + c.ID)
		}
		cats[c.ID] = c
	}
	for _, c := range w.Categories {
		seen := map[string]bool{}
		current := c
		for current.ParentID != "" {
			if seen[current.ID] {
				return bad("category cycle")
			}
			seen[current.ID] = true
			var ok bool
			current, ok = cats[current.ParentID]
			if !ok {
				return bad("category parent")
			}
		}
	}
	brands := map[string]bool{}
	for _, b := range w.Brands {
		if !add(b.ID) || !validCatalogText(CatalogInput{Name: b.Name}) || !imageOK(b.IconID) || b.Version != 1 {
			return bad("brand " + b.ID)
		}
		brands[b.ID] = b.Active
	}
	defs := map[string]catalog.AttributeDefinition{}
	codes := map[string]bool{}
	names := map[string]bool{}
	for _, a := range w.Attributes {
		if !add(a.ID) || !validateAttribute(a) || codes[a.Code] || names[strings.ToLower(a.Name)] || a.Version != 1 {
			return bad("attribute " + a.ID)
		}
		codes[a.Code] = true
		names[strings.ToLower(a.Name)] = true
		defs[a.ID] = a
	}
	for _, c := range w.Categories {
		seen := map[string]bool{}
		axes := 0
		for _, b := range c.Bindings {
			a, ok := defs[b.AttributeID]
			if !ok || seen[b.AttributeID] || b.Position < 0 || b.Position > 10000 || b.Axis && a.Kind != "enum" {
				return bad("category binding " + c.ID)
			}
			seen[b.AttributeID] = true
			if b.Axis {
				axes++
			}
		}
		if axes > 2 {
			return bad("more than two axes " + c.ID)
		}
	}
	stores := map[string]bool{}
	for _, st := range w.Stores {
		if !add(st.ID) || strings.TrimSpace(st.Name) == "" || !st.Active || st.LeadDays < 0 || st.LeadDays > 365 || st.Owner != "" || st.Model != "shared" || st.ContentPolicy != "review_required" || st.ContentPolicyVersion != 1 {
			return bad("store " + st.ID)
		}
		stores[st.ID] = true
	}
	products := map[string]catalog.Product{}
	for i := range s.Catalog.Products {
		p := &s.Catalog.Products[i]
		if !add(p.ID) || p.Status != "published" || p.Model != "shared" || p.OwnerStoreID != "" || p.ContentRevision != nil || p.Version != 1 || p.VariantVersion != 1 || p.ImageID == "" || !media[p.ImageID] || len(p.GalleryIDs) > MaxGalleryImages || p.Content == nil || !validProductContent(*p.Content) || !validCatalogText(CatalogInput{Name: p.Name, EnglishName: p.EnglishName, Description: p.Description}) || !brands[p.BrandID] || publicCategory(w, p.CategoryID) == nil || !validProductAttributes(w, p.CategoryID, p.Attributes) {
			return bad("product " + p.ID)
		}
		seen := map[string]bool{p.ImageID: true}
		for _, id := range p.GalleryIDs {
			if id == "" || !media[id] || seen[id] {
				return bad("gallery " + p.ID)
			}
			seen[id] = true
		}
		p.Content.FullDescription = sanitizeProductHTML(p.Content.FullDescription)
		products[p.ID] = *p
	}
	variants := map[string]catalog.Variant{}
	skus := map[string]bool{}
	selections := map[string]bool{}
	hasVariant := map[string]bool{}
	for _, v := range w.Variants {
		p, ok := products[v.ProductID]
		if !add(v.ID) || !ok || !skuPattern.MatchString(v.SKU) || skus[strings.ToLower(v.SKU)] || v.Version != 1 || !imageOK(v.ImageID) {
			return bad("variant " + v.ID)
		}
		key, e := canonicalSelection(productAxes(w, p), v.Values)
		key = p.ID + ":" + key
		if e != nil || selections[key] {
			return bad("variant selection " + v.ID)
		}
		selections[key] = true
		skus[strings.ToLower(v.SKU)] = true
		variants[v.ID] = v
		hasVariant[p.ID] = true
	}
	hasOffer := map[string]bool{}
	pairs := map[string]bool{}
	for _, o := range w.Offers {
		v, ok := variants[o.VariantID]
		key := o.VariantID + ":" + o.StoreID
		if !add(o.ID) || !ok || v.ProductID != o.ProductID || !stores[o.StoreID] || pairs[key] || o.Stock <= 0 || o.Stock > MaxOfferStock || o.Reserved != 0 || o.Price <= 0 || o.Price > MaxOfferPrice || o.RegularPrice < 0 || o.RegularPrice > MaxOfferPrice || o.RegularPrice > 0 && o.RegularPrice <= o.Price || o.Version != 1 {
			return bad("offer " + o.ID)
		}
		pairs[key] = true
		hasOffer[o.ProductID] = true
	}
	for id := range products {
		if !hasVariant[id] || !hasOffer[id] {
			return bad("product without purchasable offer " + id)
		}
	}
	h := w.HomeSettings
	if len(h.Rails) != 4 {
		return bad("home rail count")
	}
	for i, r := range h.Rails {
		if r.Key != homeKeys[i] || r.Limit < 1 || r.Limit > 10 {
			return bad("home rail")
		}
		if _, ok := cats[r.CategoryID]; !ok {
			return bad("home category")
		}
	}
	return nil
}
func (c SeedCatalog) workspace() Workspace {
	return Workspace{Categories: c.Categories, Brands: c.Brands, Attributes: c.Attributes, Products: c.Products, Variants: c.Variants, Offers: c.Offers, Stores: c.Stores, HomeSettings: c.HomeSettings}
}

// LimitCatalogSeed returns a dependency-complete sample. Zero imports the entire snapshot.
func LimitCatalogSeed(s CatalogSeed, n int) CatalogSeed {
	if n <= 0 || n >= len(s.Catalog.Products) {
		return s
	}
	s.Catalog.Products = append([]catalog.Product{}, s.Catalog.Products[:n]...)
	pids, cids, bids, sids, aids, mids := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, p := range s.Catalog.Products {
		pids[p.ID] = true
		cids[p.CategoryID] = true
		bids[p.BrandID] = true
		mids[p.ImageID] = true
		for _, id := range p.GalleryIDs {
			mids[id] = true
		}
	}
	// Preserve the four homepage category paths even when the sample has no products in one rail.
	for _, r := range s.Catalog.HomeSettings.Rails {
		cids[r.CategoryID] = true
	}
	for changed := true; changed; {
		changed = false
		for _, c := range s.Catalog.Categories {
			if cids[c.ID] && c.ParentID != "" && !cids[c.ParentID] {
				cids[c.ParentID] = true
				changed = true
			}
		}
	}
	s.Catalog.Variants = filterSeed(s.Catalog.Variants, func(v catalog.Variant) bool {
		if pids[v.ProductID] {
			mids[v.ImageID] = true
			return true
		}
		return false
	})
	s.Catalog.Offers = filterSeed(s.Catalog.Offers, func(o catalog.Offer) bool {
		if pids[o.ProductID] {
			sids[o.StoreID] = true
			return true
		}
		return false
	})
	s.Catalog.Stores = filterSeed(s.Catalog.Stores, func(st store.Store) bool { return sids[st.ID] })
	s.Catalog.Categories = filterSeed(s.Catalog.Categories, func(c catalog.Category) bool {
		if cids[c.ID] {
			mids[c.IconID] = true
			for _, b := range c.Bindings {
				aids[b.AttributeID] = true
			}
			return true
		}
		return false
	})
	s.Catalog.Brands = filterSeed(s.Catalog.Brands, func(b catalog.Brand) bool {
		if bids[b.ID] {
			mids[b.IconID] = true
			return true
		}
		return false
	})
	s.Catalog.Attributes = filterSeed(s.Catalog.Attributes, func(a catalog.AttributeDefinition) bool { return aids[a.ID] })
	s.Media = filterSeed(s.Media, func(m SeedMedia) bool { return mids[m.ID] })
	return s
}
func filterSeed[T any](in []T, keep func(T) bool) []T {
	out := []T{}
	for _, v := range in {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// ApplyCatalogSeed never overwrites any existing entity. The repository transaction makes
// collision failure and concurrent retries atomic; financial/customer/access data is untouched.
func ApplyCatalogSeed(ctx context.Context, repo Repository, s CatalogSeed, media []catalog.IconMedia, configureHome bool, now time.Time) (SeedImportResult, error) {
	result := seedResult(s, len(media))
	if err := ValidateCatalogSeed(&s); err != nil {
		return result, err
	}
	digests := []SeedMediaDigest{}
	for _, m := range media {
		if err := validateSeedPNG(m.Data); err != nil {
			return result, err
		}
		sum := sha256.Sum256(m.Data)
		digests = append(digests, SeedMediaDigest{m.ID, hex.EncodeToString(sum[:])})
	}
	return applySeed(ctx, repo.Update, s, media, digests, configureHome, now)
}

// File-backed import bounds memory to one image per worker/transaction write.
type CatalogSeedFile struct {
	ID     string
	Path   string
	SHA256 string
	Bytes  int64
}
type SeedMediaDigest struct {
	ID     string
	SHA256 string
}

var errSeedAlreadyApplied = errors.New("catalog_seed_already_applied")

func seedResult(s CatalogSeed, n int) SeedImportResult {
	return SeedImportResult{DatasetID: s.DatasetID, Products: len(s.Catalog.Products), Variants: len(s.Catalog.Variants), Offers: len(s.Catalog.Offers), Media: n}
}
func validateSeedPNG(data []byte) error {
	if len(data) > MaxProductImageBytes {
		return fmt.Errorf("hydrated media too large")
	}
	cfg, e := png.DecodeConfig(bytes.NewReader(data))
	if e != nil || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 1024 || cfg.Height > 1024 {
		return fmt.Errorf("invalid hydrated media")
	}
	if _, e = png.Decode(bytes.NewReader(data)); e != nil {
		return fmt.Errorf("invalid hydrated media")
	}
	return nil
}
func ApplyCatalogSeedFiles(ctx context.Context, repo interface {
	UpdateCatalogWithFiles(context.Context, []CatalogSeedFile, func(*Workspace) error) error
}, s CatalogSeed, files []CatalogSeedFile, configureHome bool, now time.Time) (SeedImportResult, error) {
	result := seedResult(s, len(files))
	if e := ValidateCatalogSeed(&s); e != nil {
		return result, e
	}
	media := make([]catalog.IconMedia, 0, len(files))
	digests := make([]SeedMediaDigest, 0, len(files))
	for _, f := range files {
		info, e := os.Stat(f.Path)
		if e != nil || info.Size() != f.Bytes || info.Size() > MaxProductImageBytes {
			return result, fmt.Errorf("invalid normalized media file %s", f.ID)
		}
		data, e := os.ReadFile(f.Path)
		if e != nil {
			return result, fmt.Errorf("cannot read normalized media %s", f.ID)
		}
		if e = validateSeedPNG(data); e != nil {
			return result, e
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != f.SHA256 {
			return result, fmt.Errorf("normalized media checksum changed %s", f.ID)
		}
		media = append(media, catalog.IconMedia{ID: f.ID, Stored: true})
		digests = append(digests, SeedMediaDigest{f.ID, f.SHA256})
	}
	update := func(ctx context.Context, apply func(*Workspace) error) error {
		return repo.UpdateCatalogWithFiles(ctx, files, apply)
	}
	return applySeed(ctx, update, s, media, digests, configureHome, now)
}
func applySeed(ctx context.Context, update func(context.Context, func(*Workspace) error) error, s CatalogSeed, media []catalog.IconMedia, digests []SeedMediaDigest, configureHome bool, now time.Time) (SeedImportResult, error) {
	result := seedResult(s, len(media))
	attached := map[string]bool{}
	for _, m := range media {
		if !seedID.MatchString(m.ID) || attached[m.ID] {
			return result, fmt.Errorf("invalid hydrated media ID")
		}
		attached[m.ID] = true
	}
	if len(attached) != len(s.Media) {
		return result, fmt.Errorf("hydrated media count mismatch")
	}
	for _, m := range s.Media {
		if !attached[m.ID] {
			return result, fmt.Errorf("unhydrated media %s", m.ID)
		}
	}
	// Prefer source checksums so a Go PNG encoder upgrade does not change retry identity.
	sourceHashes := map[string]string{}
	for _, m := range s.Media {
		if m.SHA256 != "" {
			sourceHashes[m.ID] = m.SHA256
		}
	}
	for i := range digests {
		if h := sourceHashes[digests[i].ID]; h != "" {
			digests[i].SHA256 = h
		}
	}
	sort.Slice(digests, func(i, j int) bool { return digests[i].ID < digests[j].ID })
	fingerprintData, _ := json.Marshal(struct {
		Catalog SeedCatalog
		Media   []SeedMediaDigest
		Home    bool
	}{s.Catalog, digests, configureHome})
	hash := sha256.Sum256(fingerprintData)
	fingerprint := hex.EncodeToString(hash[:])
	key := "catalog-seed-" + fingerprint
	err := update(ctx, func(w *Workspace) error {
		for _, a := range w.Audit {
			if a.ID == key && a.Action == "catalog.seed.import" {
				if a.Fingerprint != fingerprint {
					return ErrConflict
				}
				result.AlreadyApplied = true
				return errSeedAlreadyApplied
			}
		}
		existing := seedCatalogIDs(SeedCatalog{Categories: w.Categories, Brands: w.Brands, Attributes: w.Attributes, Products: w.Products, Variants: w.Variants, Offers: w.Offers, Stores: w.Stores})
		for _, m := range w.IconMedia {
			existing[m.ID] = true
		}
		incoming := seedCatalogIDs(s.Catalog)
		for _, m := range media {
			incoming[m.ID] = true
		}
		for id := range incoming {
			if existing[id] {
				return fmt.Errorf("catalog ID already exists: %s; no data changed", id)
			}
		}
		w.Categories = append(w.Categories, s.Catalog.Categories...)
		w.Brands = append(w.Brands, s.Catalog.Brands...)
		w.Attributes = append(w.Attributes, s.Catalog.Attributes...)
		w.Products = append(w.Products, s.Catalog.Products...)
		w.Variants = append(w.Variants, s.Catalog.Variants...)
		w.Offers = append(w.Offers, s.Catalog.Offers...)
		w.Stores = append(w.Stores, s.Catalog.Stores...)
		w.IconMedia = append(w.IconMedia, media...)
		if configureHome {
			h := s.Catalog.HomeSettings
			h.Version = w.HomeSettings.Version + 1
			w.HomeSettings = h
		}
		w.Audit = append(w.Audit, Audit{ID: key, Actor: "catalog-seed-cli", Action: "catalog.seed.import", Target: s.DatasetID, Reason: "test catalog snapshot; no identity or financial data imported", Fingerprint: fingerprint, At: now})
		return nil
	})
	if errors.Is(err, errSeedAlreadyApplied) {
		err = nil
	}
	return result, err
}

// Stable hydrated ordering is needed for retry fingerprints independent of network completion order.
func SortSeedMedia(media []catalog.IconMedia) {
	sort.Slice(media, func(i, j int) bool { return media[i].ID < media[j].ID })
}
func NormalizedSeedImage(raw []byte) ([]byte, error) { return normalizeProductPhoto(raw) }
func seedCatalogIDs(c SeedCatalog) map[string]bool {
	out := map[string]bool{}
	for _, v := range c.Categories {
		out[v.ID] = true
	}
	for _, v := range c.Brands {
		out[v.ID] = true
	}
	for _, v := range c.Attributes {
		out[v.ID] = true
	}
	for _, v := range c.Products {
		out[v.ID] = true
	}
	for _, v := range c.Variants {
		out[v.ID] = true
	}
	for _, v := range c.Offers {
		out[v.ID] = true
	}
	for _, v := range c.Stores {
		out[v.ID] = true
	}
	return out
}
