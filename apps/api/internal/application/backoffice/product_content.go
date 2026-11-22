package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/microcosm-cc/bluemonday"
	"math"
	"strings"
	"unicode/utf8"
)

const MaxGalleryImages = 8

func validProductContent(c catalog.ProductContent) bool {
	validText := func(s string, max int) bool {
		return utf8.RuneCountInString(s) <= max && !strings.ContainsRune(s, '\x00')
	}
	if !validText(c.ShortDescription, 2000) || !validText(c.FullDescription, 20000) || !validCategorySEO(c.SEO) || len(c.Advantages) > 20 || len(c.Disadvantages) > 20 {
		return false
	}
	for _, items := range [][]string{c.Advantages, c.Disadvantages} {
		for _, s := range items {
			if strings.TrimSpace(s) == "" || !validText(s, 300) {
				return false
			}
		}
	}
	d := c.Shipping
	if d.WeightUnit != "" && d.WeightUnit != "g" {
		return false
	}
	if d.DimensionUnit != "" && d.DimensionUnit != "mm" {
		return false
	}
	for i, v := range []*float64{d.Weight, d.Length, d.Width, d.Height} {
		if v != nil {
			if math.IsNaN(*v) || math.IsInf(*v, 0) || *v <= 0 || *v > 1e9 || math.Abs(*v*1000-math.Round(*v*1000)) > 1e-5 {
				return false
			}
			if i == 0 && d.WeightUnit == "" || i > 0 && d.DimensionUnit == "" {
				return false
			}
		}
	}
	return true
}

type ProductMediaInput struct {
	GalleryIDs    []string          `json:"gallery_ids"`
	VariantImages map[string]string `json:"variant_images"`
}

// Media updates can only reuse images already attached to this product.
func applyProductMedia(w *Workspace, c Command) error {
	if c.ProductMedia == nil || len(c.ProductMedia.GalleryIDs) > MaxGalleryImages {
		return ErrInvalid
	}
	p, err := findProduct(w, c.Target)
	if err != nil {
		return err
	}
	if p.Version != c.Version {
		return ErrConflict
	}
	allowed := map[string]bool{p.ImageID: true}
	for _, id := range p.GalleryIDs {
		allowed[id] = true
	}
	for _, v := range w.Variants {
		if v.ProductID == p.ID {
			allowed[v.ImageID] = true
		}
	}
	seen := map[string]bool{}
	for _, id := range c.ProductMedia.GalleryIDs {
		if id == "" || seen[id] || !allowed[id] || !imageAttached(*w, id) {
			return ErrInvalid
		}
		seen[id] = true
	}
	for vid, id := range c.ProductMedia.VariantImages {
		found := false
		for _, v := range w.Variants {
			if v.ID == vid && v.ProductID == p.ID {
				found = true
			}
		}
		if !found || id != "" && (!allowed[id] || !imageAttached(*w, id)) {
			return ErrInvalid
		}
	}
	p.GalleryIDs = append([]string{}, c.ProductMedia.GalleryIDs...)
	p.Version++
	for i := range w.Variants {
		v := &w.Variants[i]
		if v.ProductID == p.ID {
			if id, ok := c.ProductMedia.VariantImages[v.ID]; ok {
				v.ImageID = id
				v.Version++
			}
		}
	}
	collectUnusedIcons(w)
	return nil
}

// A narrow policy is shared by writes and public reads, including legacy records.
var productHTMLPolicy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "h2", "h3", "strong", "b", "em", "i", "u", "s", "ul", "ol", "li", "blockquote", "a", "hr", "code", "pre")
	p.AllowAttrs("href").OnElements("a")
	p.AllowURLSchemes("http", "https")
	p.AllowRelativeURLs(true)
	p.RequireParseableURLs(true)
	p.RequireNoFollowOnLinks(true)
	p.RequireNoReferrerOnLinks(true)
	return p
}()

func sanitizeProductHTML(s string) string { return productHTMLPolicy.Sanitize(s) }
