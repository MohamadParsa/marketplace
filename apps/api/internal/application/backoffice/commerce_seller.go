package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
)

type SellerCheckoutLine struct {
	ProductID      string `json:"product_id"`
	VariantID      string `json:"variant_id"`
	OfferID        string `json:"offer_id"`
	Name           string `json:"name"`
	Quantity       int    `json:"quantity"`
	UnitPriceToman string `json:"unit_price_toman"`
}
type SellerCheckoutOrder struct {
	OrderStatus       string                 `json:"order_status"`
	Delivery          *shipping.DeliveryPlan `json:"delivery,omitempty"`
	CollectionMode    string                 `json:"collection_mode,omitempty"`
	ShippingFeeToman  string                 `json:"shipping_fee_toman,omitempty"`
	TransportDayBasis shipping.DayBasis      `json:"transport_day_basis,omitempty"`
	MethodVersion     int                    `json:"method_version,omitempty"`
	StoreRevision     int                    `json:"store_revision,omitempty"`
	PurchaseNumber    string                 `json:"purchase_number"`
	ID                string                 `json:"id"`
	PurchaseID        string                 `json:"purchase_id"`
	StoreID           string                 `json:"store_id"`
	City              string                 `json:"city"`
	Status            string                 `json:"status"`
	PaymentStatus     string                 `json:"payment_status"`
	Items             []SellerCheckoutLine   `json:"items"`
	ShippingMethod    string                 `json:"shipping_method"`
	DeliveryFrom      string                 `json:"delivery_from"`
	DeliveryTo        string                 `json:"delivery_to"`
}

func (s *Service) ReadSellerCheckoutOrders(ctx context.Context, a identity.User, id string) ([]SellerCheckoutOrder, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return nil, e
	}
	if a.Surface != identity.SurfaceSeller {
		return nil, ErrForbidden
	}
	a, e = catalogActor(w, a)
	if e != nil {
		return nil, e
	}
	if w.Commerce != nil {
		commerceState(&w)
	}
	out := []SellerCheckoutOrder{}
	for _, o := range w.Orders {
		if o.CheckoutID == "" || o.Status == "superseded" || id != "" && o.ID != id {
			continue
		}
		if w.Access == nil || !identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, a.ID, o.StoreID, "orders.read") {
			continue
		}
		v := SellerCheckoutOrder{PurchaseNumber: o.PurchaseNumber, ID: o.ID, PurchaseID: o.PurchaseID, StoreID: o.StoreID, City: o.City, Status: o.Status, PaymentStatus: o.Payment, Items: []SellerCheckoutLine{}, ShippingMethod: o.ShippingMethod, DeliveryFrom: o.DeliveryFrom, DeliveryTo: o.DeliveryTo}
		v.Delivery = o.Delivery
		v.OrderStatus = o.OrderStatus
		if o.ShippingTerms != nil {
			terms := shippingMethodInput(*o.ShippingTerms)
			v.CollectionMode, v.ShippingFeeToman, v.TransportDayBasis = terms.CollectionMode, terms.FeeToman, terms.DayBasis
			v.MethodVersion, v.StoreRevision = o.ShippingTerms.Version, o.ShippingTerms.StoreRevision
		}
		for _, l := range o.Items {
			v.Items = append(v.Items, SellerCheckoutLine{ProductID: l.ProductID, VariantID: l.VariantID, OfferID: l.OfferID, Name: l.Name, Quantity: l.Quantity, UnitPriceToman: commerce.Toman(commerce.Amount(l.UnitPrice))})
		}
		out = append(out, v)
	}
	if id != "" && len(out) == 0 {
		return nil, ErrNotFound
	}
	return out, nil
}
