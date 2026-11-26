package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"strings"
)

// Normalize copies at the command boundary, before validation and fingerprinting.
// Aliases never become persisted catalog IDs or a second definition.
func normalizeAttributeValues(values []catalog.AttributeValue) ([]catalog.AttributeValue, error) {
	if values == nil {
		return nil, nil
	}
	out := append([]catalog.AttributeValue{}, values...)
	for i := range out {
		id, err := ResolvePublicReference("a", out[i].AttributeID)
		if err != nil {
			return nil, err
		}
		out[i].AttributeID = id
	}
	return out, nil
}
func normalizeAttributeCommand(state Workspace, c Command) (Command, error) {
	var err error
	if strings.HasPrefix(c.Action, "attribute.") {
		c.Target, err = ResolvePublicReference("a", c.Target)
		if err != nil {
			return c, err
		}
	}
	if c.Attribute != nil {
		a := *c.Attribute
		if c.Action == "attribute.create" && strings.HasPrefix(a.Code, "chr-a-") {
			return c, ErrInvalid
		}
		a.ID, err = ResolvePublicReference("a", a.ID)
		if err != nil {
			return c, err
		}
		if c.Action != "attribute.create" {
			for _, original := range state.Attributes {
				if original.ID == c.Target && a.Code == PublicReference("a", original.Code) {
					a.Code = original.Code
					break
				}
			}
		}
		c.Attribute = &a
	}
	if c.Bindings != nil {
		c.Bindings = append([]catalog.AttributeBinding{}, c.Bindings...)
		for i := range c.Bindings {
			c.Bindings[i].AttributeID, err = ResolvePublicReference("a", c.Bindings[i].AttributeID)
			if err != nil {
				return c, err
			}
		}
	}
	if c.Variant != nil {
		v := *c.Variant
		v.Values, err = normalizeAttributeValues(v.Values)
		if err != nil {
			return c, err
		}
		c.Variant = &v
	}
	if c.Catalog != nil && c.Catalog.Attributes != nil {
		v := *c.Catalog
		values, e := normalizeAttributeValues(*v.Attributes)
		if e != nil {
			return c, e
		}
		v.Attributes = &values
		c.Catalog = &v
	}
	return c, nil
}
