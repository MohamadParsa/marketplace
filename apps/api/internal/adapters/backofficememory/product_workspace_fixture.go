package backofficememory

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func fixtureProductWorkspace(state *backoffice.Workspace) {
	state.Users = append(state.Users, backoffice.User{ID: "catalog-seller-a", Name: "فروشنده آزمون الف", Phone: "09120000012", Role: identity.RoleViewer, Active: true, SellerOnly: true}, backoffice.User{ID: "catalog-seller-b", Name: "فروشنده آزمون ب", Phone: "09120000013", Role: identity.RoleViewer, Active: true, SellerOnly: true})
	state.Memberships = []backoffice.StoreMembership{
		{UserID: "catalog-seller-a", StoreID: "s1", Active: true, CreateOffers: true, EditPrice: true, EditStock: true},
		{UserID: "catalog-seller-b", StoreID: "s2", Active: true, CreateOffers: true, EditPrice: true, EditStock: true},
	}
	state.Access.Roles = append(state.Access.Roles, identity.RoleDefinition{ID: "catalog-price", Name: "عرضه و قیمت", Scope: identity.ScopeStore, Permissions: []string{"offers.read", "offers.create", "offers.price.write"}, Version: 1}, identity.RoleDefinition{ID: "catalog-stock", Name: "موجودی کاتالوگ", Scope: identity.ScopeStore, Permissions: []string{"stock.write"}, Version: 1})
	for _, m := range state.Memberships {
		state.StoreMemberships = append(state.StoreMemberships, identity.StoreMembership{UserID: m.UserID, StoreID: m.StoreID})
		for _, r := range []string{"catalog-price", "catalog-stock"} {
			state.Access.Grants = append(state.Access.Grants, identity.RoleGrant{UserID: m.UserID, StoreID: m.StoreID, RoleID: r})
		}
	}
	for i := range state.Stores {
		state.Stores[i].LeadDays = i + 1
	}
	owners := map[string]string{"p2": "s2", "p4": "s4", "p5": "s2"}
	for i := range state.Products {
		p := &state.Products[i]
		p.OwnerStoreID = owners[p.ID]
		p.VariantVersion = 1
		values := []catalog.AttributeValue{}
		if p.CategoryID == "c1" {
			values = append(values, catalog.AttributeValue{AttributeID: "color", Value: "مشکی"})
		}
		if p.ID != "mobile-sample" {
			state.Variants = append(state.Variants, catalog.Variant{ID: "v-" + p.ID, ProductID: p.ID, SKU: "TEST-" + p.ID, Values: values, Version: 1})
		}
	}
	state.Variants = append(state.Variants,
		catalog.Variant{ID: "mobile-black-18", ProductID: "mobile-sample", SKU: "TEST-MOB-B18", Version: 1, Values: []catalog.AttributeValue{{AttributeID: "color", Value: "مشکی"}, {AttributeID: "warranty", Value: "۱۸ ماهه"}}},
		catalog.Variant{ID: "mobile-white-24", ProductID: "mobile-sample", SKU: "TEST-MOB-W24", Version: 1, Values: []catalog.AttributeValue{{AttributeID: "color", Value: "سفید"}, {AttributeID: "warranty", Value: "۲۴ ماهه"}}},
	)
	for i := range state.Offers {
		state.Offers[i].VariantID = "v-" + state.Offers[i].ProductID
		state.Offers[i].Version = 1
	}
	state.Offers = append(state.Offers,
		catalog.Offer{ID: "mob-a-black", ProductID: "mobile-sample", VariantID: "mobile-black-18", StoreID: "s1", Price: 120000000, Stock: 5, Reserved: 2, Version: 1},
		catalog.Offer{ID: "mob-b-black", ProductID: "mobile-sample", VariantID: "mobile-black-18", StoreID: "s2", Price: 118000000, Stock: 3, Reserved: 1, Version: 1},
		catalog.Offer{ID: "mob-a-white", ProductID: "mobile-sample", VariantID: "mobile-white-24", StoreID: "s1", Price: 125000000, Stock: 2, Reserved: 2, Version: 1},
		catalog.Offer{ID: "mob-b-white", ProductID: "mobile-sample", VariantID: "mobile-white-24", StoreID: "s2", Price: 126000000, Stock: 4, Reserved: 0, Version: 1},
	)
}
