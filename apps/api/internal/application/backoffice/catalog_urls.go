package backoffice

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

var publicSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var camelBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)
var slugSeparators = regexp.MustCompile(`[^a-z0-9]+`)

func ValidCategorySlug(s string) bool {
	return len(s) > 0 && len(s) <= 100 && publicSlugPattern.MatchString(s)
}
func DefaultCategorySlug(name string) string {
	return strings.Trim(slugSeparators.ReplaceAllString(strings.ToLower(camelBoundary.ReplaceAllString(strings.TrimSpace(name), "$1-$2")), "-"), "-")
}
func ProductURL(p catalog.Product) string {
	token := p.PublicCode
	if token == "" {
		token = p.ID
	}
	return "/product/" + url.PathEscape(token)
}
func CategoryURL(c catalog.Category) string {
	token := c.Slug
	if token == "" {
		token = c.ID
	}
	return "/category/" + url.PathEscape(token)
}

// Registry retains deleted codes and old category aliases; neither can be reassigned.
type CatalogURLRegistry struct {
	Products   map[string]int64  `json:"products"`
	Categories map[string]string `json:"categories"` // slug/alias -> internal category ID
	LastNumber int64             `json:"last_number"`
}
type ProductURLMapping struct {
	ID     string `json:"id"`
	Code   string `json:"public_code"`
	Number int64  `json:"number"`
}
type CategoryURLMapping struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	EnglishName string `json:"english_name"`
}
type CatalogURLPlan struct {
	Products   []ProductURLMapping  `json:"products"`
	Categories []CategoryURLMapping `json:"categories"`
	LastNumber int64                `json:"last_number"`
}

// PlanCatalogURLs is read-only. Conflicts stop the entire backfill, never choose a silent alternative.
func PlanCatalogURLs(w Workspace, registry CatalogURLRegistry) (CatalogURLPlan, error) {
	plan := CatalogURLPlan{Products: []ProductURLMapping{}, Categories: []CategoryURLMapping{}}
	numbers := map[int64]string{}
	codes := map[string]int64{}
	aliases := map[string]string{}
	max := registry.LastNumber
	if max < 999 {
		max = 999
	}
	for id, n := range registry.Products {
		if n <= 0 || numbers[n] != "" && numbers[n] != id {
			return plan, fmt.Errorf("%w: product code collision", ErrConflict)
		}
		codes[id] = n
		numbers[n] = id
		if n > max {
			max = n
		}
	}
	productIDs := map[string]bool{}
	for id := range registry.Products {
		productIDs[id] = true
	}
	for _, p := range w.Products {
		productIDs[p.ID] = true
	}
	products := append([]catalog.Product{}, w.Products...)
	sort.Slice(products, func(i, j int) bool { return products[i].ID < products[j].ID })
	// Reserve all imported numbers before allocating any native product number.
	for _, p := range products {
		if !strings.HasPrefix(p.ID, "gsp-product-") {
			continue
		}
		s := strings.TrimPrefix(p.ID, "gsp-product-")
		n, e := strconv.ParseInt(s, 10, 64)
		if e != nil || n <= 0 || strconv.FormatInt(n, 10) != s {
			return plan, fmt.Errorf("%w: invalid source product number %s", ErrInvalid, p.ID)
		}
		if codes[p.ID] != 0 && codes[p.ID] != n || numbers[n] != "" && numbers[n] != p.ID {
			return plan, fmt.Errorf("%w: product code %d", ErrConflict, n)
		}
		codes[p.ID] = n
		numbers[n] = p.ID
		if n > max {
			max = n
		}
	}
	for _, p := range products {
		n := codes[p.ID]
		if n == 0 {
			if max == 9223372036854775807 {
				return plan, ErrConflict
			}
			max++
			n = max
			codes[p.ID] = n
		}
		code := "chr-" + strconv.FormatInt(n, 10)
		if productIDs[code] && p.ID != code {
			return plan, fmt.Errorf("%w: code collides with internal product route %s", ErrConflict, code)
		}
		if p.PublicCode != "" && p.PublicCode != code {
			return plan, fmt.Errorf("%w: immutable code %s", ErrConflict, p.ID)
		}
		plan.Products = append(plan.Products, ProductURLMapping{p.ID, code, n})
	}
	for slug, id := range registry.Categories {
		aliases[slug] = id
	}
	// Reserve old ID routes too, so another category cannot steal them.
	for _, c := range w.Categories {
		if owner := aliases[c.ID]; owner != "" && owner != c.ID {
			return plan, fmt.Errorf("%w: category ID alias %s", ErrConflict, c.ID)
		}
		aliases[c.ID] = c.ID
	}
	categories := append([]catalog.Category{}, w.Categories...)
	sort.Slice(categories, func(i, j int) bool { return categories[i].ID < categories[j].ID })
	for _, c := range categories {
		slug := c.Slug
		if slug == "" {
			slug = DefaultCategorySlug(c.EnglishName)
			if c.ID == "gsp-cat-755" {
				slug = "smart-watch"
			}
			if slug == "" && !strings.HasPrefix(c.ID, "gsp-cat-") {
				slug = c.ID
			}
		}
		if !ValidCategorySlug(slug) {
			return plan, fmt.Errorf("%w: missing/invalid English slug for %s", ErrInvalid, c.ID)
		}
		if owner := aliases[slug]; owner != "" && owner != c.ID {
			return plan, fmt.Errorf("%w: category slug %s", ErrConflict, slug)
		}
		aliases[slug] = c.ID
		plan.Categories = append(plan.Categories, CategoryURLMapping{c.ID, slug, c.EnglishName})
	}
	plan.LastNumber = max
	return plan, nil
}

type CatalogURLResolver interface {
	ResolveCatalogToken(context.Context, string, string) (string, error)
}

func (s *Service) resolveCatalogToken(ctx context.Context, kind, token string) (string, error) {
	if token == "" || len(token) > 128 {
		return "", ErrNotFound
	}
	if r, ok := s.repository.(CatalogURLResolver); ok {
		return r.ResolveCatalogToken(ctx, kind, token)
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return "", e
	}
	if kind == "product" {
		for _, p := range w.Products {
			if p.ID == token || p.PublicCode != "" && p.PublicCode == token {
				return p.ID, nil
			}
		}
	} else {
		for _, c := range w.Categories {
			if c.ID == token || c.Slug != "" && c.Slug == token {
				return c.ID, nil
			}
		}
	}
	return "", ErrNotFound
}
