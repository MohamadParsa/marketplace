package backoffice

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

var attributeCode = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

// Test contract safety bounds; these are not a selected production card limit.
const MaxCardAttributeLimit = 20

func validateAttribute(a catalog.AttributeDefinition) bool {
	if !attributeCode.MatchString(a.Code) || strings.TrimSpace(a.Name) == "" || a.Name != strings.TrimSpace(a.Name) || utf8.RuneCountInString(a.Name) > 80 || utf8.RuneCountInString(a.Unit) > 32 || len(a.Options) > 100 {
		return false
	}
	if a.Kind != "enum" && a.Kind != "text" && a.Kind != "number" {
		return false
	}
	if (a.Kind == "enum" && len(a.Options) == 0) || (a.Kind != "enum" && len(a.Options) != 0) {
		return false
	}
	seen := map[string]bool{}
	for _, option := range a.Options {
		key := strings.ToLower(strings.TrimSpace(option))
		if key == "" || option != strings.TrimSpace(option) || utf8.RuneCountInString(option) > 80 || seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}

func attributeUsed(state *Workspace, id string) bool {
	for _, c := range state.Categories {
		for _, b := range c.Bindings {
			if b.AttributeID == id {
				return true
			}
		}
	}
	for _, p := range state.Products {
		for _, v := range p.Attributes {
			if v.AttributeID == id {
				return true
			}
		}
	}
	return false
}

func applyAttributes(state *Workspace, c Command, id string) error {
	switch c.Action {
	case "settings.card-limit":
		if c.Target != "display" || c.Amount < 1 || c.Amount > MaxCardAttributeLimit {
			return ErrInvalid
		}
		if state.DisplaySettings.Version != c.Version {
			return ErrConflict
		}
		state.DisplaySettings.CardAttributeLimit = int(c.Amount)
		state.DisplaySettings.Version++
		return nil
	case "attribute.create", "attribute.update":
		if c.Attribute == nil || !validateAttribute(*c.Attribute) {
			return ErrInvalid
		}
		a := *c.Attribute
		for _, existing := range state.Attributes {
			if (c.Action == "attribute.create" || existing.ID != c.Target) && (existing.Code == a.Code || strings.EqualFold(existing.Name, a.Name)) {
				return ErrConflict
			}
		}
		if c.Action == "attribute.create" {
			a.ID, a.Version = id, 1
			state.Attributes = append(state.Attributes, a)
			return nil
		}
		for i, existing := range state.Attributes {
			if existing.ID != c.Target {
				continue
			}
			if existing.Version != c.Version {
				return ErrConflict
			}
			// Definition identity is stable; schema changes on used definitions need a later migration policy.
			if existing.Code != a.Code {
				return ErrConflict
			}
			if attributeUsed(state, existing.ID) && (existing.Kind != a.Kind || existing.Unit != a.Unit || !sameOptions(existing.Options, a.Options)) {
				return ErrConflict
			}
			a.ID, a.Version = existing.ID, existing.Version+1
			state.Attributes[i] = a
			return nil
		}
	case "attribute.delete":
		for i, a := range state.Attributes {
			if a.ID != c.Target {
				continue
			}
			if a.Version != c.Version || attributeUsed(state, a.ID) {
				return ErrConflict
			}
			state.Attributes = append(state.Attributes[:i], state.Attributes[i+1:]...)
			return nil
		}
	case "category.bindings.set":
		for i := range state.Categories {
			category := &state.Categories[i]
			if category.ID != c.Target {
				continue
			}
			if category.Archived || category.Version != c.Version {
				return ErrConflict
			}
			if len(c.Bindings) > 100 {
				return ErrInvalid
			}
			seen := map[string]bool{}
			for _, b := range c.Bindings {
				if seen[b.AttributeID] || b.Position < 0 || b.Position > 10000 {
					return ErrInvalid
				}
				seen[b.AttributeID] = true
				var def *catalog.AttributeDefinition
				for j := range state.Attributes {
					if state.Attributes[j].ID == b.AttributeID {
						def = &state.Attributes[j]
						break
					}
				}
				if def == nil || (b.Axis && def.Kind != "enum") {
					return ErrInvalid
				}
			}
			// Until the product/variant migration policy is selected, freeze the axis set on any used category.
			for _, product := range state.Products {
				if product.CategoryID != category.ID {
					continue
				}
				if !sameAxes(category.Bindings, c.Bindings) {
					return ErrConflict
				}
				for _, value := range product.Attributes {
					if !seen[value.AttributeID] {
						return ErrConflict
					}
				}
			}
			category.Bindings = append([]catalog.AttributeBinding{}, c.Bindings...)
			category.Version++
			return nil
		}
	}
	return ErrNotFound
}

func sameOptions(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameAxes(a, b []catalog.AttributeBinding) bool {
	axis := func(bindings []catalog.AttributeBinding) []string {
		ids := []string{}
		for _, x := range bindings {
			if x.Axis {
				ids = append(ids, x.AttributeID)
			}
		}
		sort.Strings(ids)
		return ids
	}
	return sameOptions(axis(a), axis(b))
}

// BuildCardPreviews is a read projection. It never mutates products, offers or historical orders.
func BuildCardPreviews(state Workspace) []catalog.CardPreview {
	result := []catalog.CardPreview{}
	for _, product := range state.Products {
		preview := catalog.CardPreview{ProductID: product.ID, Features: []catalog.CardFeature{}}
		for _, category := range state.Categories {
			if category.ID != product.CategoryID || category.Archived {
				continue
			}
			bindings := append([]catalog.AttributeBinding{}, category.Bindings...)
			sort.Slice(bindings, func(i, j int) bool {
				if bindings[i].Position == bindings[j].Position {
					return bindings[i].AttributeID < bindings[j].AttributeID
				}
				return bindings[i].Position < bindings[j].Position
			})
			for _, b := range bindings {
				if !b.Card || len(preview.Features) >= state.DisplaySettings.CardAttributeLimit {
					continue
				}
				for _, value := range product.Attributes {
					if value.AttributeID != b.AttributeID || strings.TrimSpace(value.Value) == "" {
						continue
					}
					for _, a := range state.Attributes {
						if a.ID == b.AttributeID {
							preview.Features = append(preview.Features, catalog.CardFeature{AttributeID: a.ID, Name: a.Name, Value: value.Value, Unit: a.Unit})
							break
						}
					}
					break
				}
			}
			break
		}
		result = append(result, preview)
	}
	return result
}
