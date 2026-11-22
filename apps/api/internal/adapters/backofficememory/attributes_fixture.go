package backofficememory

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

func fixtureAttributes(state *backoffice.Workspace) {
	state.DisplaySettings = catalog.DisplaySettings{CardAttributeLimit: 3, Version: 1} // Synthetic test value, not a production decision.
	state.Attributes = []catalog.AttributeDefinition{
		{ID: "color", Code: "color", Name: "رنگ", Kind: "enum", Options: []string{"مشکی", "سفید", "آبی"}, Version: 1},
		{ID: "warranty", Code: "warranty", Name: "گارانتی", Kind: "enum", Options: []string{"۱۸ ماهه", "۲۴ ماهه"}, Version: 1},
		{ID: "memory", Code: "memory", Name: "حافظه", Kind: "number", Unit: "گیگابایت", Options: []string{}, Version: 1},
		{ID: "ram", Code: "ram", Name: "رم", Kind: "number", Unit: "گیگابایت", Options: []string{}, Version: 1},
		{ID: "camera", Code: "camera", Name: "دوربین", Kind: "number", Unit: "مگاپیکسل", Options: []string{}, Version: 1},
		{ID: "screen", Code: "screen", Name: "نمایشگر", Kind: "number", Unit: "اینچ", Options: []string{}, Version: 1},
	}
	bind := func(id string, axis, card, filter bool, pos int) catalog.AttributeBinding {
		return catalog.AttributeBinding{AttributeID: id, Axis: axis, Card: card, Filter: filter, Position: pos}
	}
	state.Categories[0].Bindings = []catalog.AttributeBinding{bind("color", true, false, true, 0)}
	state.Categories[1].Bindings = []catalog.AttributeBinding{}
	state.Categories[2].Bindings = []catalog.AttributeBinding{bind("memory", false, true, true, 0)}
	state.Categories = append(state.Categories,
		catalog.Category{ID: "mobile", Name: "موبایل آزمایشی", ParentID: "c3", Version: 1, Bindings: []catalog.AttributeBinding{
			bind("color", true, false, true, 0), bind("warranty", true, false, true, 1), bind("memory", false, true, true, 2), bind("ram", false, true, true, 3), bind("camera", false, true, false, 4), bind("screen", false, true, false, 5),
		}},
		catalog.Category{ID: "charger", Name: "شارژر آزمایشی", ParentID: "c3", Version: 1, Bindings: []catalog.AttributeBinding{bind("color", true, false, true, 0)}},
		catalog.Category{ID: "empty-category", Name: "دسته خالی آزمایشی", Version: 1, Bindings: []catalog.AttributeBinding{}},
	)
	state.Products = append(state.Products, catalog.Product{ID: "mobile-sample", Name: "گوشی نمونه برای بررسی کارت", CategoryID: "mobile", Model: "shared", Status: "draft", Version: 1, Attributes: []catalog.AttributeValue{
		{AttributeID: "memory", Value: "256"}, {AttributeID: "ram", Value: "8"}, {AttributeID: "camera", Value: "50"}, {AttributeID: "screen", Value: "6.5"},
	}})
}
