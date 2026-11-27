package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"slices"
	"strings"
)

// Only live catalog fields and attribute audit target presentation are projected.
// Financial/order snapshots, source references and repository state stay intact.
func panelDefinitions(defs []catalog.AttributeDefinition) []catalog.AttributeDefinition {
	if defs == nil {
		return nil
	}
	out := append([]catalog.AttributeDefinition{}, defs...)
	for i := range out {
		out[i].ID = backoffice.PublicReference("a", out[i].ID)
		out[i].Code = backoffice.PublicReference("a", out[i].Code)
	}
	return out
}
func panelValues(values []catalog.AttributeValue) []catalog.AttributeValue {
	if values == nil {
		return nil
	}
	out := append([]catalog.AttributeValue{}, values...)
	for i := range out {
		out[i].AttributeID = backoffice.PublicReference("a", out[i].AttributeID)
	}
	return out
}
func panelProduct(p catalog.Product) catalog.Product {
	p.Attributes = panelValues(p.Attributes)
	if p.ContentRevision != nil {
		r := *p.ContentRevision
		r.Attributes = panelValues(r.Attributes)
		p.ContentRevision = &r
	}
	return p
}
func panelVariants(values []catalog.Variant) []catalog.Variant {
	if values == nil {
		return nil
	}
	out := append([]catalog.Variant{}, values...)
	for i := range out {
		out[i].Values = panelValues(out[i].Values)
	}
	return out
}
func panelWorkspace(state backoffice.Workspace) backoffice.Workspace {
	if state.RecordNumbers != nil {
		numbers := make(map[string]int64, len(state.RecordNumbers))
		for key, number := range state.RecordNumbers {
			if strings.HasPrefix(key, "attribute|") {
				key = "attribute|" + backoffice.PublicReference("a", strings.TrimPrefix(key, "attribute|"))
			}
			numbers[key] = number
		}
		state.RecordNumbers = numbers
	}
	state.Attributes = panelDefinitions(state.Attributes)
	state.Categories = slices.Clone(state.Categories)
	for i := range state.Categories {
		c := &state.Categories[i]
		c.Bindings = slices.Clone(c.Bindings)
		for j := range c.Bindings {
			c.Bindings[j].AttributeID = backoffice.PublicReference("a", c.Bindings[j].AttributeID)
		}
	}
	state.Products = slices.Clone(state.Products)
	for i := range state.Products {
		state.Products[i] = panelProduct(state.Products[i])
	}
	state.Variants = panelVariants(state.Variants)
	state.CardPreviews = slices.Clone(state.CardPreviews)
	for i := range state.CardPreviews {
		c := &state.CardPreviews[i]
		c.Features = slices.Clone(c.Features)
		for j := range c.Features {
			c.Features[j].AttributeID = backoffice.PublicReference("a", c.Features[j].AttributeID)
		}
	}
	state.Audit = slices.Clone(state.Audit)
	for i := range state.Audit {
		if strings.HasPrefix(state.Audit[i].Action, "attribute.") {
			state.Audit[i].Target = backoffice.PublicReference("a", state.Audit[i].Target)
		}
	}
	return state
}
func panelProductWorkspace(state backoffice.ProductWorkspace) backoffice.ProductWorkspace {
	state.Product = panelProduct(state.Product)
	state.Axes = panelDefinitions(state.Axes)
	state.Variants = panelVariants(state.Variants)
	return state
}
