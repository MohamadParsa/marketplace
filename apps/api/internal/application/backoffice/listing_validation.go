package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Plain-text bounds for the initial category editor; rendering escapes all text.
func validCategorySEO(s catalog.SEO) bool {
	return utf8.RuneCountInString(s.MetaTitle) <= 160 && utf8.RuneCountInString(s.MetaKeywords) <= 500 && utf8.RuneCountInString(s.MetaDescription) <= 400 && !strings.ContainsRune(s.MetaTitle, '\x00') && !strings.ContainsRune(s.MetaKeywords, '\x00') && !strings.ContainsRune(s.MetaDescription, '\x00')
}
func validProductAttributes(w Workspace, category string, values []catalog.AttributeValue) bool {
	if len(values) > 100 {
		return false
	}
	seen := map[string]bool{}
	for _, v := range values {
		if seen[v.AttributeID] || strings.TrimSpace(v.Value) == "" || utf8.RuneCountInString(v.Value) > 160 {
			return false
		}
		seen[v.AttributeID] = true
		found := false
		for _, c := range w.Categories {
			if c.ID != category {
				continue
			}
			for _, b := range c.Bindings {
				if b.AttributeID != v.AttributeID {
					continue
				}
				for _, a := range w.Attributes {
					if a.ID != b.AttributeID {
						continue
					}
					found = true
					if a.Kind == "enum" {
						valid := false
						for _, o := range a.Options {
							valid = valid || o == v.Value
						}
						if !valid {
							return false
						}
					}
					if a.Kind == "number" {
						n, err := strconv.ParseFloat(v.Value, 64)
						if err != nil || n != n || n > 1e15 || n < -1e15 {
							return false
						}
					}
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func NormalizeSearch(s string) string {
	return strings.Join(strings.Fields(strings.Map(func(r rune) rune {
		switch r {
		case 'ي', 'ى':
			return 'ی'
		case 'ك':
			return 'ک'
		case '\u200c', '\u200d':
			return ' '
		case '\u0640':
			return -1
		}
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		if r >= '۰' && r <= '۹' {
			return '0' + r - '۰'
		}
		if r >= '٠' && r <= '٩' {
			return '0' + r - '٠'
		}
		return unicode.ToLower(r)
	}, s)), " ")
}

// Axis assignments belong to variants. Preserve unchanged legacy duplicates
// during content edits, but never introduce a new axis value on a product.
func validContentAttributes(w Workspace, category string, previous, incoming []catalog.AttributeValue) bool {
	if !validProductAttributes(w, category, incoming) {
		return false
	}
	axes := map[string]bool{}
	for _, c := range w.Categories {
		if c.ID == category {
			for _, b := range c.Bindings {
				if b.Axis {
					axes[b.AttributeID] = true
				}
			}
		}
	}
	for _, v := range incoming {
		if axes[v.AttributeID] {
			same := false
			for _, old := range previous {
				if old == v {
					same = true
					break
				}
			}
			if !same {
				return false
			}
		}
	}
	return true
}
