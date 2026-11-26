package backoffice

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
)

type VariantInput struct {
	ProductID string                   `json:"product_id"`
	SKU       string                   `json:"sku"`
	Values    []catalog.AttributeValue `json:"values"`
}
type OfferInput struct {
	RegularPrice *int64 `json:"regular_price,omitempty"`
	ProductID    string `json:"product_id"`
	VariantID    string `json:"variant_id"`
	StoreID      string `json:"store_id"`
	Price        *int64 `json:"price,omitempty"`
	Stock        *int   `json:"stock,omitempty"`
}
type ProductPermissions struct {
	EditContent       bool     `json:"edit_content"`
	ReviewContent     bool     `json:"review_content"`
	CreateVariant     bool     `json:"create_variant"`
	CreateOfferStores []string `json:"create_offer_stores"`
	PriceStores       []string `json:"price_stores"`
	StockStores       []string `json:"stock_stores"`
}
type OfferView struct {
	SaleRestricted bool `json:"sale_restricted,omitempty"`
	catalog.Offer
	Available    int    `json:"available"`
	LeadDayBasis string `json:"lead_day_basis"`
	LeadDays     int    `json:"lead_days"`
	StoreActive  bool   `json:"store_active"`
}
type ProductWorkspace struct {
	Product     catalog.Product               `json:"product"`
	Axes        []catalog.AttributeDefinition `json:"axes"`
	Variants    []catalog.Variant             `json:"variants"`
	Offers      []OfferView                   `json:"offers"`
	Stores      []store.Store                 `json:"stores"`
	Permissions ProductPermissions            `json:"permissions"`
}
type VariantResolution struct {
	Variant catalog.Variant `json:"variant"`
	Offers  []OfferView     `json:"offers"`
}

func isProductAction(action string) bool {
	return action == "variant.create" || action == "offer.create" || action == "offer.price.update" || action == "offer.stock.update"
}

func catalogActor(state Workspace, actor identity.User) (identity.User, error) {
	if actor.Surface != identity.SurfaceSeller {
		return currentActor(state, actor)
	}
	for _, u := range state.Users {
		if u.ID == actor.ID && u.Active && u.AccessVersion == actor.AccessVersion && len(identity.MemberStoreIDs(state.StoreMemberships, u.ID)) > 0 {
			actor.Role = u.Role
			return actor, nil
		}
	}
	return identity.User{}, ErrForbidden
}
func canStore(state Workspace, actor identity.User, storeID, permission string) bool {
	if actor.Surface != identity.SurfaceSeller {
		action := map[string]string{"read": WorkspaceRead, "create": "offer.create", "price": "offer.price.update", "stock": "offer.stock.update"}[permission]
		return allowsAction(actor, action) || actor.Role == identity.RoleSuperAdmin || actor.Role == identity.RoleOperations
	}
	if state.Access == nil {
		return false
	}
	perm := map[string]identity.StorePermission{"read": identity.StoreOffersRead, "create": "offers.create", "price": "offers.price.write", "stock": identity.StoreStockWrite}[permission]
	return identity.AllowsStore(state.StoreMemberships, state.Access.Roles, state.Access.Grants, actor.ID, storeID, perm)
}

func visibleProduct(state Workspace, actor identity.User, p catalog.Product) bool {
	return actor.Surface != identity.SurfaceSeller || p.Model == "shared" || (p.OwnerStoreID != "" && (canStore(state, actor, p.OwnerStoreID, "read") || sellerCanEditContent(state, actor, p)))
}

func findProduct(state *Workspace, id string) (*catalog.Product, error) {
	for i := range state.Products {
		if state.Products[i].ID == id {
			return &state.Products[i], nil
		}
	}
	return nil, ErrNotFound
}
func productAxes(state Workspace, p catalog.Product) []catalog.AttributeDefinition {
	result := []catalog.AttributeDefinition{}
	for _, c := range state.Categories {
		if c.ID == p.CategoryID {
			bindings := append([]catalog.AttributeBinding{}, c.Bindings...)
			sort.Slice(bindings, func(i, j int) bool {
				if bindings[i].Position == bindings[j].Position {
					return bindings[i].AttributeID < bindings[j].AttributeID
				}
				return bindings[i].Position < bindings[j].Position
			})
			for _, b := range bindings {
				if b.Axis {
					for _, a := range state.Attributes {
						if a.ID == b.AttributeID {
							result = append(result, a)
							break
						}
					}
				}
			}
			break
		}
	}
	return result
}

func (s *Service) ReadProductWorkspace(ctx context.Context, actor identity.User, id string) (ProductWorkspace, error) {
	state, err := s.repository.Snapshot(ctx)
	if err != nil {
		return ProductWorkspace{}, err
	}
	actor, err = catalogActor(state, actor)
	if err != nil {
		return ProductWorkspace{}, err
	}
	p, err := findProduct(&state, id)
	if err != nil {
		return ProductWorkspace{}, err
	}
	if !visibleProduct(state, actor, *p) {
		return ProductWorkspace{}, ErrForbidden
	}
	result := ProductWorkspace{Product: *p, Axes: productAxes(state, *p), Variants: []catalog.Variant{}, Offers: []OfferView{}, Stores: []store.Store{}, Permissions: ProductPermissions{EditContent: actor.Surface == identity.SurfaceSeller && sellerCanEditContent(state, actor, *p), ReviewContent: actor.Surface != identity.SurfaceSeller && allowsAction(actor, "product.content.review"), CreateVariant: actor.Surface != identity.SurfaceSeller && allowsAction(actor, "variant.create"), CreateOfferStores: []string{}, PriceStores: []string{}, StockStores: []string{}}}
	for _, v := range state.Variants {
		if v.ProductID == id {
			result.Variants = append(result.Variants, v)
		}
	}
	for _, st := range state.Stores {
		if actor.Surface == identity.SurfaceSeller && !canStore(state, actor, st.ID, "read") && !hasContentStoreAccess(state, actor, st.ID) {
			continue
		}
		st.Profile = nil
		result.Stores = append(result.Stores, st)
		if !storeHasSale(st, state, p.CategoryID) || (p.Model == "independent" && p.OwnerStoreID != st.ID) {
			continue
		}
		if canStore(state, actor, st.ID, "create") {
			result.Permissions.CreateOfferStores = append(result.Permissions.CreateOfferStores, st.ID)
		}
		if canStore(state, actor, st.ID, "price") {
			result.Permissions.PriceStores = append(result.Permissions.PriceStores, st.ID)
		}
		if canStore(state, actor, st.ID, "stock") {
			result.Permissions.StockStores = append(result.Permissions.StockStores, st.ID)
		}
	}
	for _, o := range state.Offers {
		if o.ProductID != id {
			continue
		}
		if actor.Surface == identity.SurfaceSeller && !canStore(state, actor, o.StoreID, "read") {
			continue
		}
		for _, st := range result.Stores {
			if st.ID == o.StoreID {
				result.Offers = append(result.Offers, OfferView{SaleRestricted: !StoreAllowsCategory(state, st, p.CategoryID) || !StoreAllowsMethod(st, "cash"), Offer: o, Available: o.Stock - o.Reserved, LeadDayBasis: string(preparationPolicy(st).DayBasis), LeadDays: st.LeadDays, StoreActive: st.Active})
				break
			}
		}
	}
	return result, nil
}

func canonicalSelection(axes []catalog.AttributeDefinition, values []catalog.AttributeValue) (string, error) {
	if len(values) != len(axes) {
		return "", ErrInvalid
	}
	byID := map[string]string{}
	for _, v := range values {
		if _, ok := byID[v.AttributeID]; ok {
			return "", ErrInvalid
		}
		byID[v.AttributeID] = v.Value
	}
	parts := []catalog.AttributeValue{}
	for _, a := range axes {
		value, ok := byID[a.ID]
		if !ok || a.Kind != "enum" {
			return "", ErrInvalid
		}
		valid := false
		for _, option := range a.Options {
			if option == value {
				valid = true
				break
			}
		}
		if !valid {
			return "", ErrInvalid
		}
		parts = append(parts, catalog.AttributeValue{AttributeID: a.ID, Value: value})
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].AttributeID < parts[j].AttributeID })
	encoded, _ := json.Marshal(parts)
	return string(encoded), nil
}

func sellerWorkspace(state Workspace, actor identity.User) Workspace {
	result := Workspace{DisplaySettings: state.DisplaySettings, Categories: state.Categories, Attributes: state.Attributes, Brands: state.Brands, Products: []catalog.Product{}, Stores: []store.Store{}, Offers: []catalog.Offer{}, Variants: []catalog.Variant{}}
	visible := map[string]bool{}
	for _, p := range state.Products {
		if visibleProduct(state, actor, p) {
			result.Products = append(result.Products, p)
			visible[p.ID] = true
		}
	}
	for _, st := range state.Stores {
		if canStore(state, actor, st.ID, "read") || hasContentStoreAccess(state, actor, st.ID) {
			st.Profile = nil
			result.Stores = append(result.Stores, st)
		}
	}
	for _, o := range state.Offers {
		if visible[o.ProductID] && canStore(state, actor, o.StoreID, "read") {
			result.Offers = append(result.Offers, o)
		}
	}
	for _, v := range state.Variants {
		if visible[v.ProductID] {
			result.Variants = append(result.Variants, v)
		}
	}
	result.CardPreviews = BuildCardPreviews(result)
	return result
}

func (s *Service) ResolveVariant(ctx context.Context, actor identity.User, productID string, values []catalog.AttributeValue) (VariantResolution, error) {
	w, err := s.ReadProductWorkspace(ctx, actor, productID)
	if err != nil {
		return VariantResolution{}, err
	}
	values, err = normalizeAttributeValues(values)
	if err != nil {
		return VariantResolution{}, err
	}
	key, err := canonicalSelection(w.Axes, values)
	if err != nil {
		return VariantResolution{}, err
	}
	for _, v := range w.Variants {
		candidate, e := canonicalSelection(w.Axes, v.Values)
		if e != nil || candidate != key {
			continue
		}
		result := VariantResolution{Variant: v, Offers: []OfferView{}}
		for _, o := range w.Offers {
			if o.VariantID == v.ID && o.StoreActive && o.Available > 0 && o.Price > 0 {
				result.Offers = append(result.Offers, o)
			}
		}
		sort.Slice(result.Offers, func(i, j int) bool {
			if result.Offers[i].Price == result.Offers[j].Price {
				return result.Offers[i].ID < result.Offers[j].ID
			}
			return result.Offers[i].Price < result.Offers[j].Price
		})
		return result, nil
	}
	return VariantResolution{}, ErrNotFound
}

func authorizeProductCommand(state Workspace, actor identity.User, c Command) error {
	current, err := catalogActor(state, actor)
	if err != nil {
		return err
	}
	if c.Action == "variant.create" {
		if current.Surface == identity.SurfaceSeller || (current.Role != identity.RoleSuperAdmin && current.Role != identity.RoleOperations) {
			return ErrForbidden
		}
		return nil
	}
	storeID := ""
	if c.Action == "offer.create" {
		if c.Offer == nil {
			return ErrInvalid
		}
		storeID = c.Offer.StoreID
	} else {
		for _, o := range state.Offers {
			if o.ID == c.Target {
				storeID = o.StoreID
				break
			}
		}
		if storeID == "" {
			if current.Surface == identity.SurfaceSeller {
				return ErrForbidden
			}
			return ErrNotFound
		}
	}
	permission := "create"
	if c.Action == "offer.price.update" {
		permission = "price"
	}
	if c.Action == "offer.stock.update" {
		permission = "stock"
	}
	if !canStore(state, current, storeID, permission) {
		return ErrForbidden
	}
	return nil
}

var skuPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

const MaxOfferPrice int64 = 9000000000000000
const MaxOfferStock = 1000000000

func applyProduct(state *Workspace, c Command, id string) error {
	if c.Action == "variant.create" {
		if c.Variant == nil || c.Variant.ProductID != c.Target || !skuPattern.MatchString(c.Variant.SKU) {
			return ErrInvalid
		}
		p, err := findProduct(state, c.Target)
		if err != nil {
			return err
		}
		if p.VariantVersion != c.Version {
			return ErrConflict
		}
		for _, category := range state.Categories {
			if category.ID == p.CategoryID && category.Archived {
				return ErrConflict
			}
		}
		axes := productAxes(*state, *p)
		key, err := canonicalSelection(axes, c.Variant.Values)
		if err != nil {
			return err
		}
		for _, v := range state.Variants {
			if strings.EqualFold(v.SKU, c.Variant.SKU) {
				return ErrConflict
			}
			if v.ProductID == p.ID {
				existing, e := canonicalSelection(axes, v.Values)
				if e == nil && existing == key {
					return ErrConflict
				}
			}
		}
		state.Variants = append(state.Variants, catalog.Variant{ID: id, ProductID: p.ID, SKU: c.Variant.SKU, Values: append([]catalog.AttributeValue{}, c.Variant.Values...), Version: 1})
		p.VariantVersion++
		return nil
	}
	if c.Offer == nil {
		return ErrInvalid
	}
	input := c.Offer
	if c.Action == "offer.create" {
		if c.Target != input.ProductID || input.Price == nil || input.Stock == nil || *input.Price <= 0 || *input.Price > MaxOfferPrice || *input.Stock < 0 || *input.Stock > MaxOfferStock {
			return ErrInvalid
		}
		p, err := findProduct(state, input.ProductID)
		if err != nil {
			return err
		}
		if p.VariantVersion != c.Version {
			return ErrConflict
		}
		validStore := false
		for _, st := range state.Stores {
			if st.ID == input.StoreID && st.Active {
				validStore = storeHasSale(st, *state, p.CategoryID)
			}
		}
		if !validStore || (p.Model == "independent" && p.OwnerStoreID != input.StoreID) {
			return ErrConflict
		}
		validVariant := false
		for _, v := range state.Variants {
			if v.ID == input.VariantID && v.ProductID == p.ID {
				validVariant = true
			}
		}
		if !validVariant {
			return ErrInvalid
		}
		for _, o := range state.Offers {
			if o.StoreID == input.StoreID && o.VariantID == input.VariantID {
				return ErrConflict
			}
		}
		regular := int64(0)
		if input.RegularPrice != nil {
			regular = *input.RegularPrice
		}
		if regular < 0 || regular > MaxOfferPrice || (regular > 0 && regular <= *input.Price) {
			return ErrInvalid
		}
		state.Offers = append(state.Offers, catalog.Offer{RegularPrice: regular, ID: id, ProductID: p.ID, VariantID: input.VariantID, StoreID: input.StoreID, Price: *input.Price, Stock: *input.Stock, Version: 1})
		return nil
	}
	for i := range state.Offers {
		o := &state.Offers[i]
		if o.ID != c.Target {
			continue
		}
		if o.Version != c.Version {
			return ErrConflict
		}
		if input.ProductID != "" || input.VariantID != "" || input.StoreID != "" {
			return ErrInvalid
		}
		active := false
		for _, st := range state.Stores {
			if st.ID == o.StoreID && st.Active {
				active = true
			}
		}
		product, e := findProduct(state, o.ProductID)
		if e != nil {
			return e
		}
		for _, st := range state.Stores {
			if st.ID == o.StoreID && !storeHasSale(st, *state, product.CategoryID) {
				return ErrForbidden
			}
		}
		if !active {
			return ErrConflict
		}
		if c.Action == "offer.price.update" {
			if input.Price == nil || input.Stock != nil || *input.Price <= 0 || *input.Price > MaxOfferPrice {
				return ErrInvalid
			}
			regular := o.RegularPrice
			if input.RegularPrice != nil {
				regular = *input.RegularPrice
			}
			if regular < 0 || regular > MaxOfferPrice || (regular > 0 && regular <= *input.Price) {
				return ErrInvalid
			}
			o.RegularPrice = regular
			o.Price = *input.Price
		} else if c.Action == "offer.stock.update" {
			if input.Stock == nil || input.Price != nil || input.RegularPrice != nil || *input.Stock < 0 || *input.Stock > MaxOfferStock {
				return ErrInvalid
			}
			if *input.Stock < o.Reserved {
				return ErrConflict
			}
			o.Stock = *input.Stock
		} else {
			return ErrInvalid
		}
		o.Version++
		return nil
	}
	return ErrNotFound
}
