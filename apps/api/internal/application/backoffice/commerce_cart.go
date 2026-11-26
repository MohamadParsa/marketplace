package backoffice

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
	"time"
)

type CartOwner struct {
	Actor      identity.User
	GuestToken string
}
type CartMutation struct {
	ReviewVariantVersion *int   `json:"review_variant_version,omitempty"`
	OfferID              string `json:"offer_id"`
	NewOfferID           string `json:"new_offer_id,omitempty"`
	Quantity             int    `json:"quantity"`
	Version              int    `json:"version"`
	Action               string `json:"action"`
}
type CartItemView struct {
	Restricted           bool   `json:"restricted,omitempty"`
	ReviewRequired       bool   `json:"review_required,omitempty"`
	PreviousVariantLabel string `json:"previous_variant_label,omitempty"`
	VariantVersion       int    `json:"variant_version,omitempty"`
	ProductURL           string `json:"product_url"`
	ProductID            string `json:"product_id"`
	VariantID            string `json:"variant_id"`
	OfferID              string `json:"offer_id"`
	StoreID              string `json:"store_id"`
	StoreName            string `json:"store_name"`
	Name                 string `json:"name"`
	VariantLabel         string `json:"variant_label"`
	ImageURL             string `json:"image_url"`
	Quantity             int    `json:"quantity"`
	UnitPriceToman       string `json:"unit_price_toman"`
	PreviousPriceToman   string `json:"previous_price_toman"`
	TotalToman           string `json:"total_toman"`
	Available            bool   `json:"available"`
	PriceChanged         bool   `json:"price_changed"`
	LeadDays             int    `json:"lead_days"`
}
type CartView struct {
	ID            string         `json:"id"`
	Version       int            `json:"version"`
	Items         []CartItemView `json:"items"`
	SubtotalToman string         `json:"subtotal_toman"`
	ItemCount     int            `json:"item_count"`
	CanCheckout   bool           `json:"can_checkout"`
	Currency      string         `json:"currency"`
}

func cartIndex(c *commerce.State, o CartOwner) int {
	for i, v := range c.Carts {
		if v.MergedInto != "" {
			continue
		}
		if o.Actor.ID != "" && v.UserID == o.Actor.ID {
			return i
		}
		if o.Actor.ID == "" && o.GuestToken != "" && v.GuestHash == hashToken(o.GuestToken) {
			return i
		}
	}
	return -1
}
func cartGuard(w Workspace, o CartOwner) error {
	if o.Actor.ID != "" {
		return customerActor(w, o.Actor)
	}
	if len(o.GuestToken) < 32 || len(o.GuestToken) > 128 || w.Commerce == nil || cartIndex(w.Commerce, o) < 0 {
		return ErrForbidden
	}
	return nil
}
func cartScope(o CartOwner) string {
	if o.Actor.ID != "" {
		return "cart:user:" + o.Actor.ID
	}
	return "cart:guest:" + hashToken(o.GuestToken)
}
func findOffer(w Workspace, id string) (catalog.Offer, error) {
	for _, o := range w.Offers {
		if o.ID == id {
			return o, nil
		}
	}
	return catalog.Offer{}, ErrNotFound
}
func cartItem(w Workspace, l commerce.CartLine) (CartItemView, error) {
	o, e := findOffer(w, l.OfferID)
	if e != nil {
		return CartItemView{OfferID: l.OfferID, Quantity: l.Quantity, Name: "کالای در دسترس نیست"}, e
	}
	p, e := publicProduct(w, o.ProductID)
	if e != nil {
		return CartItemView{ProductID: o.ProductID, VariantID: o.VariantID, OfferID: o.ID, StoreID: o.StoreID, Quantity: l.Quantity, Name: "کالای در دسترس نیست"}, e
	}
	d, e := BuildProductDetail(w, p.ID, o.VariantID, o.ID)
	if e != nil {
		return CartItemView{ProductID: p.ID, VariantID: o.VariantID, OfferID: o.ID, Quantity: l.Quantity, Name: p.Name}, e
	}
	var ov *PublicProductOffer
	for i := range d.Offers {
		if d.Offers[i].ID == o.ID {
			ov = &d.Offers[i]
		}
	}
	if ov == nil {
		return CartItemView{}, ErrNotFound
	}
	v := CartItemView{Restricted: ov.Restricted, ProductURL: ProductURL(p), ProductID: p.ID, VariantID: o.VariantID, OfferID: o.ID, StoreID: o.StoreID, StoreName: ov.StoreName, Name: p.Name, Quantity: l.Quantity, Available: o.Stock-o.Reserved >= l.Quantity && ov.Available, LeadDays: ov.LeadDays, PreviousPriceToman: commerce.Toman(l.SeenPrice), PriceChanged: l.SeenPrice != commerce.Amount(o.Price)}
	for _, variant := range d.Variants {
		if variant.ID == o.VariantID {
			values := []string{}
			for _, a := range variant.Values {
				values = append(values, a.Value)
			}
			v.VariantLabel = strings.Join(values, " / ")
			for _, raw := range w.Variants {
				if raw.ID == o.VariantID {
					v.VariantVersion = raw.Version
					break
				}
			}
			v.ReviewRequired = l.ReviewRequired
			v.PreviousVariantLabel = l.PreviousVariantLabel
		}
	}
	if len(d.Gallery) > 0 {
		v.ImageURL = d.Gallery[0]
	}
	v.UnitPriceToman = commerce.Toman(commerce.Amount(o.Price))
	total, e := commerce.Multiply(commerce.Amount(o.Price), l.Quantity)
	if e != nil {
		return v, ErrInvalid
	}
	v.TotalToman = commerce.Toman(total)
	return v, nil
}
func buildCart(w Workspace, c commerce.Cart) CartView {
	v := CartView{ID: c.ID, Version: c.Version, Items: []CartItemView{}, CanCheckout: len(c.Lines) > 0, Currency: "toman"}
	var subtotal commerce.Amount
	for _, l := range c.Lines {
		item, e := cartItem(w, l)
		if e != nil || !item.Available || item.ReviewRequired {
			v.CanCheckout = false
		}
		v.Items = append(v.Items, item)
		v.ItemCount += l.Quantity
		if item.TotalToman != "" {
			t, _ := commerce.ParseToman(item.TotalToman)
			next, e := commerce.Add(subtotal, t)
			if e != nil {
				v.CanCheckout = false
			} else {
				subtotal = next
			}
		}
	}
	v.SubtotalToman = commerce.Toman(subtotal)
	return v
}
func (s *Service) CreateGuestCart(ctx context.Context) (map[string]any, error) {
	token, e := s.identifiers.Token()
	if e != nil {
		return nil, e
	}
	id, e := s.identifiers.Token()
	if e != nil {
		return nil, e
	}
	e = s.repository.Update(ctx, func(w *Workspace) error {
		c := commerceState(w)
		if len(c.Carts) > 10000 {
			return ErrConflict
		}
		c.Carts = append(c.Carts, commerce.Cart{ID: id, GuestHash: hashToken(token), Version: 1, Lines: []commerce.CartLine{}, UpdatedAt: s.clock.Now()})
		return nil
	})
	return map[string]any{"cart_token": token}, e
}
func (s *Service) ReadCart(ctx context.Context, o CartOwner) (CartView, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return CartView{}, e
	}
	if o.Actor.ID != "" {
		if e = customerActor(w, o.Actor); e != nil {
			return CartView{}, e
		}
	} else if e = cartGuard(w, o); e != nil {
		return CartView{}, e
	}
	changed := expireCommerce(&w, s.clock.Now())
	c := commerceState(&w)
	if changed || o.Actor.ID != "" && cartIndex(c, o) < 0 {
		e = s.repository.Update(ctx, func(w *Workspace) error {
			if o.Actor.ID != "" {
				if e := customerActor(*w, o.Actor); e != nil {
					return e
				}
			} else if e := cartGuard(*w, o); e != nil {
				return e
			}
			expireCommerce(w, s.clock.Now())
			c := commerceState(w)
			if cartIndex(c, o) < 0 {
				if o.Actor.ID == "" {
					return ErrForbidden
				}
				id, e := s.identifiers.Token()
				if e != nil {
					return e
				}
				c.Carts = append(c.Carts, commerce.Cart{ID: id, UserID: o.Actor.ID, Version: 1, Lines: []commerce.CartLine{}, UpdatedAt: s.clock.Now()})
			}
			return nil
		})
		if e != nil {
			return CartView{}, e
		}
		w, e = s.repository.Snapshot(ctx)
		if e != nil {
			return CartView{}, e
		}
	}
	if e = cartGuard(w, o); e != nil {
		return CartView{}, e
	}
	i := cartIndex(w.Commerce, o)
	if i < 0 {
		return CartView{}, ErrNotFound
	}
	return buildCart(w, w.Commerce.Carts[i]), nil
}
func (s *Service) MutateCart(ctx context.Context, o CartOwner, in CartMutation, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, cartScope(o), key, in, func(w Workspace) error { return cartGuard(w, o) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		i := cartIndex(c, o)
		if i < 0 {
			return nil, ErrNotFound
		}
		cart := &c.Carts[i]
		if in.Version != cart.Version {
			return nil, ErrConflict
		}
		index := -1
		for i, l := range cart.Lines {
			if l.OfferID == in.OfferID {
				index = i
			}
		}
		if in.Action == "review" {
			if index < 0 {
				return nil, ErrNotFound
			}
			line := &cart.Lines[index]
			view, e := cartItem(*w, *line)
			if e != nil {
				return nil, e
			}
			if in.ReviewVariantVersion == nil || *in.ReviewVariantVersion != view.VariantVersion {
				return nil, ErrCheckoutChanged
			}
			if !view.Available {
				return nil, ErrStock
			}
			line.ReviewRequired = false
			line.PreviousVariantLabel = ""
		} else if in.Action == "remove" {
			if index < 0 {
				return nil, ErrNotFound
			}
			cart.Lines = append(cart.Lines[:index], cart.Lines[index+1:]...)
		} else {
			if in.Action != "add" && in.Action != "set" && in.Action != "replace" || in.Quantity < 1 || in.Quantity > 99 {
				return nil, ErrInvalid
			}
			target := in.OfferID
			if in.Action == "replace" {
				if index < 0 || in.NewOfferID == "" {
					return nil, ErrInvalid
				}
				before, e := findOffer(*w, in.OfferID)
				if e != nil {
					return nil, e
				}
				after, e := findOffer(*w, in.NewOfferID)
				if e != nil || after.ProductID != before.ProductID {
					return nil, ErrInvalid
				}
				target = in.NewOfferID
			}
			view, e := cartItem(*w, commerce.CartLine{OfferID: target, Quantity: in.Quantity})
			if e != nil {
				return nil, e
			}
			if !view.Available {
				return nil, ErrStock
			}
			price, _ := commerce.ParseToman(view.UnitPriceToman)
			if in.Action == "add" && index >= 0 {
				qty := cart.Lines[index].Quantity + in.Quantity
				if qty > 99 {
					return nil, ErrInvalid
				}
				test, e := cartItem(*w, commerce.CartLine{OfferID: target, Quantity: qty})
				if e != nil || !test.Available {
					return nil, ErrStock
				}
				cart.Lines[index].Quantity = qty
				cart.Lines[index].SeenPrice = price
			} else if in.Action == "set" {
				if index < 0 {
					return nil, ErrNotFound
				}
				cart.Lines[index].OfferID = target
				cart.Lines[index].Quantity = in.Quantity
				cart.Lines[index].SeenPrice = price
			} else if in.Action == "replace" {
				destination := -1
				for i, l := range cart.Lines {
					if l.OfferID == target && i != index {
						destination = i
					}
				}
				if destination >= 0 {
					qty := cart.Lines[destination].Quantity + in.Quantity
					if qty > 99 {
						return nil, ErrInvalid
					}
					test, e := cartItem(*w, commerce.CartLine{OfferID: target, Quantity: qty})
					if e != nil || !test.Available {
						return nil, ErrStock
					}
					cart.Lines[destination].Quantity = qty
					cart.Lines[destination].SeenPrice = price
					if cart.Lines[index].ReviewRequired {
						cart.Lines[destination].ReviewRequired = true
						cart.Lines[destination].PreviousVariantLabel = cart.Lines[index].PreviousVariantLabel
					}
					cart.Lines = append(cart.Lines[:index], cart.Lines[index+1:]...)
				} else {
					cart.Lines[index].OfferID = target
					cart.Lines[index].Quantity = in.Quantity
					cart.Lines[index].SeenPrice = price
				}
			} else {
				if len(cart.Lines) >= 100 {
					return nil, ErrInvalid
				}
				cart.Lines = append(cart.Lines, commerce.CartLine{OfferID: target, Quantity: in.Quantity, SeenPrice: price})
			}
		}
		cart.Version++
		cart.UpdatedAt = now
		return buildCart(*w, *cart), nil
	})
}
func (s *Service) MergeCart(ctx context.Context, a identity.User, guest, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "cart-merge:"+a.ID, key, struct{ Guest string }{hashToken(guest)}, func(w Workspace) error { return customerActor(w, a) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		guestIndex := -1
		for i, cart := range c.Carts {
			if cart.GuestHash == hashToken(guest) && cart.GuestHash != "" {
				guestIndex = i
			}
		}
		if guestIndex < 0 {
			return nil, ErrNotFound
		}
		if c.Carts[guestIndex].MergedInto != "" {
			if c.Carts[guestIndex].MergedInto != a.ID {
				return nil, ErrForbidden
			}
			i := cartIndex(c, CartOwner{Actor: a})
			if i < 0 {
				return nil, ErrNotFound
			}
			return buildCart(*w, c.Carts[i]), nil
		}
		i := cartIndex(c, CartOwner{Actor: a})
		if i < 0 {
			c.Carts = append(c.Carts, commerce.Cart{ID: id, UserID: a.ID, Version: 1, Lines: []commerce.CartLine{}})
			i = len(c.Carts) - 1
		}
		cart := &c.Carts[i]
		for _, line := range c.Carts[guestIndex].Lines {
			found := false
			for j := range cart.Lines {
				if cart.Lines[j].OfferID == line.OfferID {
					if cart.Lines[j].Quantity+line.Quantity > 99 {
						return nil, ErrConflict
					}
					cart.Lines[j].Quantity += line.Quantity
					if line.ReviewRequired {
						cart.Lines[j].ReviewRequired = true
						cart.Lines[j].PreviousVariantLabel = line.PreviousVariantLabel
					}
					found = true
				}
			}
			if !found {
				if len(cart.Lines) >= 100 {
					return nil, ErrConflict
				}
				cart.Lines = append(cart.Lines, line)
			}
		}
		cart.Version++
		cart.UpdatedAt = now
		c.Carts[guestIndex].MergedInto = a.ID
		return buildCart(*w, *cart), nil
	})
}
