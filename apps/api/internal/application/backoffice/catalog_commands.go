package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
)

func applyCatalog(state *Workspace, c Command, id string) error {
	switch c.Action {
	case "store.create":
		if c.Name == "" || c.Value == "" {
			return ErrInvalid
		}
		state.Stores = append(state.Stores, store.Store{ID: id, Name: c.Name, Owner: "مالک آزمایشی", City: c.Value, Active: true, Model: "shared"})
		return nil
	case "store.toggle":
		for i := range state.Stores {
			if state.Stores[i].ID == c.Target {
				state.Stores[i].Active = !state.Stores[i].Active
				return nil
			}
		}
	case "category.create":
		if c.Name == "" {
			return ErrInvalid
		}
		state.Categories = append(state.Categories, catalog.Category{ID: id, Name: c.Name, Bindings: []catalog.AttributeBinding{}, Version: 1})
		return nil
	case "category.archive":
		for i := range state.Categories {
			if state.Categories[i].ID == c.Target {
				state.Categories[i].Archived = true
				state.Categories[i].Version++
				return nil
			}
		}
	case "product.publish", "product.reject":
		for i := range state.Products {
			p := &state.Products[i]
			if p.ID != c.Target {
				continue
			}
			if p.Version != c.Version {
				return ErrConflict
			}
			if c.Action == "product.publish" {
				if !canPublishProduct(*state, *p) {
					return ErrInvalid
				}
				p.Status = "published"
			} else {
				if c.Reason == "" {
					return ErrInvalid
				}
				p.Status = "rejected"
			}
			p.Version++
			return nil
		}
	case "rule.toggle":
		for i := range state.Rules {
			if state.Rules[i].ID == c.Target {
				state.Rules[i].Active = !state.Rules[i].Active
				return nil
			}
		}
	case "shipping.update":
		for i := range state.Shipping {
			if state.Shipping[i].ID == c.Target {
				if c.Amount < 0 {
					return ErrInvalid
				}
				state.Shipping[i].Tariff = c.Amount
				return nil
			}
		}
	}
	return ErrNotFound
}
