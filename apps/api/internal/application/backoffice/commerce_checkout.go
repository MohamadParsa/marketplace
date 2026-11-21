package backoffice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"strings"
	"time"
)

type QuoteInput struct {
	CartVersion     int               `json:"cart_version"`
	AddressID       string            `json:"address_id"`
	Shipping        map[string]string `json:"shipping"`
	PaymentMethodID string            `json:"payment_method_id"`
	PurchaseID      string            `json:"purchase_id,omitempty"`
}
type AcceptQuoteInput struct {
	QuoteID            string `json:"quote_id"`
	AcceptedTotalToman string `json:"accepted_total_toman"`
}
type QuoteLineView struct {
	OfferID        string `json:"offer_id"`
	Quantity       int    `json:"quantity"`
	ProductID      string `json:"product_id"`
	VariantID      string `json:"variant_id"`
	StoreID        string `json:"store_id"`
	Name           string `json:"name"`
	VariantLabel   string `json:"variant_label"`
	ImageURL       string `json:"image_url"`
	UnitPriceToman string `json:"unit_price_toman"`
	TotalToman     string `json:"total_toman"`
}
type ShipmentView struct {
	StoreID      string `json:"store_id"`
	StoreName    string `json:"store_name"`
	MethodID     string `json:"method_id"`
	MethodName   string `json:"method_name"`
	FeeToman     string `json:"fee_toman"`
	LeadDays     int    `json:"lead_days"`
	MinDays      int    `json:"min_days"`
	MaxDays      int    `json:"max_days"`
	DeliveryFrom string `json:"delivery_from"`
	DeliveryTo   string `json:"delivery_to"`
}
type QuoteView struct {
	Destination   commerce.Destination `json:"destination"`
	ID            string               `json:"id"`
	Status        string               `json:"status"`
	Address       commerce.Address     `json:"address"`
	Lines         []QuoteLineView      `json:"lines"`
	Shipments     []ShipmentView       `json:"shipments"`
	PaymentMethod PaymentMethodInput   `json:"payment_method"`
	SubtotalToman string               `json:"subtotal_toman"`
	ShippingToman string               `json:"shipping_toman"`
	TotalToman    string               `json:"total_toman"`
	ExpiresAt     time.Time            `json:"expires_at"`
	ReservedUntil time.Time            `json:"reserved_until"`
	Currency      string               `json:"currency"`
}
type PurchaseView struct {
	OrderStatus           string           `json:"order_status"`
	PaymentStatus         string           `json:"payment_status"`
	Events                []OrderEventView `json:"events"`
	CancellationRequested bool             `json:"cancellation_requested"`
	CancellationReason    string           `json:"cancellation_reason,omitempty"`
	ID                    string           `json:"id"`
	Status                string           `json:"status"`
	Version               int              `json:"version"`
	Quote                 QuoteView        `json:"quote"`
	OrderIDs              []string         `json:"order_ids"`
	CreatedAt             time.Time        `json:"created_at"`
	ConfirmedAt           *time.Time       `json:"confirmed_at,omitempty"`
	WalletBalanceToman    string           `json:"wallet_balance_toman"`
	WalletAvailableToman  string           `json:"wallet_available_toman"`
	ShortfallToman        string           `json:"shortfall_toman"`
	Deposits              []DepositView    `json:"deposits"`
	Reason                string           `json:"reason,omitempty"`
}
type MethodEligibility struct {
	PaymentMethodInput
	Eligible bool   `json:"eligible"`
	Reason   string `json:"reason,omitempty"`
}
type PrepareView struct {
	Cart                 CartView              `json:"cart"`
	Addresses            []commerce.Address    `json:"addresses"`
	Shipping             []ShippingMethodInput `json:"shipping"`
	Methods              []PaymentMethodInput  `json:"methods"`
	WalletBalanceToman   string                `json:"wallet_balance_toman"`
	WalletAvailableToman string                `json:"wallet_available_toman"`
	Currency             string                `json:"currency"`
}

func quoteView(q commerce.Quote) QuoteView {
	v := QuoteView{Destination: q.Destination, ID: q.ID, Status: q.Status, Address: q.Address, Lines: []QuoteLineView{}, Shipments: []ShipmentView{}, PaymentMethod: methodInput(q.PaymentMethod), SubtotalToman: commerce.Toman(q.Subtotal), ShippingToman: commerce.Toman(q.ShippingTotal), TotalToman: commerce.Toman(q.Total), ExpiresAt: q.ExpiresAt, ReservedUntil: q.ReservedUntil, Currency: "toman"}
	for _, l := range q.Lines {
		total, _ := commerce.Multiply(l.UnitPrice, l.Quantity)
		v.Lines = append(v.Lines, QuoteLineView{OfferID: l.OfferID, Quantity: l.Quantity, ProductID: l.ProductID, VariantID: l.VariantID, StoreID: l.StoreID, Name: l.Name, VariantLabel: l.VariantLabel, ImageURL: l.ImageURL, UnitPriceToman: commerce.Toman(l.UnitPrice), TotalToman: commerce.Toman(total)})
	}
	for _, s := range q.Shipments {
		v.Shipments = append(v.Shipments, ShipmentView{StoreID: s.StoreID, StoreName: s.StoreName, MethodID: s.Method.ID, MethodName: s.Method.Name, FeeToman: commerce.Toman(s.Method.Fee), LeadDays: s.LeadDays, MinDays: s.Method.MinDays, MaxDays: s.Method.MaxDays, DeliveryFrom: s.DeliveryFrom, DeliveryTo: s.DeliveryTo})
	}
	return v
}
func purchaseView(c *commerce.State, p commerce.Purchase) PurchaseView {
	q := findQuote(c, p.QuoteID)
	orderStatus, paymentStatus := purchaseStatuses(p)
	v := PurchaseView{OrderStatus: orderStatus, PaymentStatus: paymentStatus, Events: []OrderEventView{}, CancellationRequested: p.CancellationRequested, CancellationReason: p.CancellationReason, ID: p.ID, Status: p.Status, Version: p.Version, OrderIDs: p.OrderIDs, CreatedAt: p.CreatedAt, ConfirmedAt: p.ConfirmedAt, Deposits: []DepositView{}, Reason: p.Reason}
	for _, e := range p.Events {
		v.Events = append(v.Events, OrderEventView{e.Action, e.FlowStatus, e.OrderStatus, e.PaymentStatus, e.ActorKind, e.At})
	}
	balance, held := walletAmounts(c, p.UserID)
	v.WalletBalanceToman = commerce.Toman(balance)
	v.WalletAvailableToman = commerce.Toman(balance - held)
	if q != nil {
		v.Quote = quoteView(*q)
		short := q.Total - (balance - held)
		if short < 0 {
			short = 0
		}
		v.ShortfallToman = commerce.Toman(short)
	}
	for _, d := range c.Deposits {
		if d.UserID == p.UserID && d.PurchaseID == p.ID {
			v.Deposits = append(v.Deposits, depositView(d))
		}
	}
	return v
}
func buildQuote(w Workspace, a identity.User, in QuoteInput, id string, now time.Time) (commerce.Quote, error) {
	c := commerceState(&w)
	q := commerce.Quote{ID: id, UserID: a.ID, Status: "preview", CreatedAt: now, ExpiresAt: now.Add(commerce.ReservationMinutes * time.Minute), Lines: []commerce.LineSnapshot{}, Shipments: []commerce.ShipmentSnapshot{}, ReplacesPurchaseID: in.PurchaseID}
	found := false
	for _, v := range c.Addresses {
		if v.ID == in.AddressID && v.UserID == a.ID {
			q.Address = v
			found = true
		}
	}
	if !found {
		return q, ErrNotFound
	}
	lines := []commerce.CartLine{}
	ownHeld := map[string]int{}
	if in.PurchaseID != "" {
		p := findPurchase(c, in.PurchaseID)
		if p == nil || p.UserID != a.ID {
			return q, ErrNotFound
		}
		if p.Status == "confirmed" || p.Status == "cancelled" {
			return q, ErrConflict
		}
		old := findQuote(c, p.QuoteID)
		if old == nil {
			return q, ErrConflict
		}
		for _, l := range old.Lines {
			lines = append(lines, commerce.CartLine{OfferID: l.OfferID, Quantity: l.Quantity, SeenPrice: l.UnitPrice})
			if old.Status == "reserved" && now.Before(old.ReservedUntil) {
				ownHeld[l.OfferID] += l.Quantity
			}
		}
	} else {
		i := cartIndex(c, CartOwner{Actor: a})
		if i < 0 {
			return q, ErrNotFound
		}
		cart := c.Carts[i]
		if cart.Version != in.CartVersion {
			return q, ErrCheckoutChanged
		}
		q.CartID = cart.ID
		q.CartVersion = cart.Version
		lines = cart.Lines
	}
	if len(lines) == 0 || len(lines) > 100 {
		return q, ErrInvalid
	}
	seen := map[string]bool{}
	for _, l := range lines {
		if seen[l.OfferID] || l.Quantity < 1 || l.Quantity > 99 {
			return q, ErrInvalid
		}
		seen[l.OfferID] = true
		o, e := findOffer(w, l.OfferID)
		if e != nil {
			return q, e
		}
		copy := w
		copy.Offers = append([]catalog.Offer{}, w.Offers...)
		for i := range copy.Offers {
			if copy.Offers[i].ID == l.OfferID {
				copy.Offers[i].Reserved -= ownHeld[l.OfferID]
			}
		}
		item, e := cartItem(copy, l)
		if e != nil {
			return q, ErrCheckoutChanged
		}
		if !item.Available {
			return q, ErrStock
		}
		price := commerce.Amount(o.Price)
		total, e := commerce.Multiply(price, l.Quantity)
		if e != nil {
			return q, ErrInvalid
		}
		q.Subtotal, e = commerce.Add(q.Subtotal, total)
		if e != nil {
			return q, ErrInvalid
		}
		q.Lines = append(q.Lines, commerce.LineSnapshot{OfferID: o.ID, ProductID: o.ProductID, VariantID: o.VariantID, StoreID: o.StoreID, Name: item.Name, VariantLabel: item.VariantLabel, ImageURL: item.ImageURL, UnitPrice: price, Quantity: l.Quantity, OfferVersion: o.Version})
	}
	stores := sortedStoreIDs(q.Lines)
	if len(in.Shipping) != len(stores) {
		return q, ErrInvalid
	}
	for _, storeID := range stores {
		found := false
		for _, st := range w.Stores {
			if st.ID != storeID || !st.Active {
				continue
			}
			for _, method := range c.Settings.Shipping {
				if method.ID == in.Shipping[storeID] && shippingEligible(method, q.Address.City, storeID) {
					q.Shipments = append(q.Shipments, commerce.ShipmentSnapshot{StoreID: st.ID, StoreName: st.Name, Method: method, LeadDays: st.LeadDays, DeliveryFrom: deliveryDate(now, st.LeadDays+method.MinDays, c.Settings), DeliveryTo: deliveryDate(now, st.LeadDays+method.MaxDays, c.Settings)})
					var e error
					q.ShippingTotal, e = commerce.Add(q.ShippingTotal, method.Fee)
					if e != nil {
						return q, ErrInvalid
					}
					found = true
					break
				}
			}
		}
		if !found {
			return q, ErrInvalid
		}
	}
	var e error
	q.Total, e = commerce.Add(q.Subtotal, q.ShippingTotal)
	if e != nil {
		return q, ErrInvalid
	}
	q.Destination = c.Settings.Destination
	if in.PaymentMethodID == "" {
		return q, nil
	}
	for _, m := range c.Settings.Methods {
		if m.ID == in.PaymentMethodID && methodEligible(w, m, q.Lines, q.Total) {
			q.PaymentMethod = m
			return q, nil
		}
	}
	return q, ErrInvalidMethod
}
func (s *Service) PrepareCheckout(ctx context.Context, a identity.User) (PrepareView, error) {
	cart, e := s.ReadCart(ctx, CartOwner{Actor: a})
	if e != nil {
		return PrepareView{}, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return PrepareView{}, e
	}
	if e = customerActor(w, a); e != nil {
		return PrepareView{}, e
	}
	c := commerceState(&w)
	settings := settingsView(c.Settings)
	addresses := []commerce.Address{}
	for _, v := range c.Addresses {
		if v.UserID == a.ID {
			addresses = append(addresses, v)
		}
	}
	balance, held := walletAmounts(c, a.ID)
	methods := []PaymentMethodInput{}
	for _, m := range settings.Methods {
		if m.Active {
			methods = append(methods, m)
		}
	}
	shipping := []ShippingMethodInput{}
	for _, m := range settings.Shipping {
		if m.Active {
			shipping = append(shipping, m)
		}
	}
	return PrepareView{Cart: cart, Addresses: addresses, Shipping: shipping, Methods: methods, WalletBalanceToman: commerce.Toman(balance), WalletAvailableToman: commerce.Toman(balance - held), Currency: "toman"}, nil
}
func (s *Service) CreateQuote(ctx context.Context, a identity.User, in QuoteInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "quote:"+a.ID, key, in, func(w Workspace) error { return customerActor(w, a) }, func(w *Workspace, id string, now time.Time) (any, error) {
		if in.PaymentMethodID == "" {
			return nil, ErrInvalidMethod
		}
		q, e := buildQuote(*w, a, in, id, now)
		if e != nil {
			return nil, e
		}
		c := commerceState(w)
		c.Quotes = append(c.Quotes, q)
		return quoteView(q), nil
	})
}
func qComparable(q commerce.Quote) string {
	q.ID = ""
	q.CreatedAt = time.Time{}
	q.ExpiresAt = time.Time{}
	q.ReservedUntil = time.Time{}
	q.AcceptedAt = nil
	q.Lines = append([]commerce.LineSnapshot{}, q.Lines...)
	for i := range q.Lines {
		q.Lines[i].OfferVersion = 0
	}
	b, _ := json.Marshal(q)
	return string(b)
}
func quoteAsInput(q commerce.Quote) QuoteInput {
	in := QuoteInput{CartVersion: q.CartVersion, AddressID: q.Address.ID, Shipping: map[string]string{}, PaymentMethodID: q.PaymentMethod.ID, PurchaseID: q.ReplacesPurchaseID}
	for _, v := range q.Shipments {
		in.Shipping[v.StoreID] = v.Method.ID
	}
	return in
}
func syncPurchaseOrders(w *Workspace, p *commerce.Purchase, q commerce.Quote, now time.Time) {
	p.OrderIDs = []string{}
	for _, ship := range q.Shipments {
		id := p.ID + ":" + ship.StoreID
		order := orders.Order{ID: id, CheckoutID: p.ID, PurchaseID: p.ID, StoreID: ship.StoreID, City: q.Address.City, Status: "awaiting-payment", Payment: "unpaid", CreatedAt: now.Format(time.RFC3339), Version: 1, Packages: []string{}, ShippingFee: int64(ship.Method.Fee), ShippingMethod: ship.Method.Name, DeliveryFrom: ship.DeliveryFrom, DeliveryTo: ship.DeliveryTo, Items: []orders.Item{}}
		for i, l := range q.Lines {
			if l.StoreID == ship.StoreID {
				order.Items = append(order.Items, orders.Item{ID: fmt.Sprintf("%s:%d", id, i), ProductID: l.ProductID, OfferID: l.OfferID, VariantID: l.VariantID, Name: l.Name, Quantity: l.Quantity, UnitPrice: int64(l.UnitPrice), Status: "awaiting-payment"})
			}
		}
		found := false
		for i := range w.Orders {
			if w.Orders[i].ID == id {
				order.Version = w.Orders[i].Version + 1
				order.CreatedAt = w.Orders[i].CreatedAt
				w.Orders[i] = order
				found = true
			}
		}
		if !found {
			w.Orders = append(w.Orders, order)
		}
		p.OrderIDs = append(p.OrderIDs, id)
	}
}
func (s *Service) AcceptQuote(ctx context.Context, a identity.User, in AcceptQuoteInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "purchase:"+a.ID, key, in, func(w Workspace) error { return customerActor(w, a) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		q := findQuote(c, in.QuoteID)
		if q == nil || q.UserID != a.ID {
			return nil, ErrNotFound
		}
		accepted, e := commerce.ParseToman(in.AcceptedTotalToman)
		if e != nil || accepted != q.Total {
			return nil, ErrCheckoutChanged
		}
		if q.PurchaseID != "" {
			p := findPurchase(c, q.PurchaseID)
			if p != nil && p.QuoteID == q.ID {
				return purchaseView(c, *p), nil
			}
			return nil, ErrConflict
		}
		if q.Status != "preview" || !now.Before(q.ExpiresAt) {
			return nil, ErrCheckoutChanged
		}
		current, e := buildQuote(*w, a, quoteAsInput(*q), q.ID, now)
		if e != nil {
			return nil, e
		}
		if qComparable(current) != qComparable(*q) {
			return nil, ErrCheckoutChanged
		}
		var p *commerce.Purchase
		if q.ReplacesPurchaseID != "" {
			p = findPurchase(c, q.ReplacesPurchaseID)
			if p == nil || p.UserID != a.ID || p.Status == "confirmed" || p.Status == "cancelled" {
				return nil, ErrConflict
			}
			old := findQuote(c, p.QuoteID)
			if old != nil {
				releaseQuote(w, old)
				old.Status = "replaced"
			}
			p.PreviousQuotes = append(p.PreviousQuotes, p.QuoteID)
			p.QuoteID = q.ID
			p.Status = "awaiting-confirmation"
			p.Version++
		} else {
			pending := 0
			for _, v := range c.Purchases {
				if v.UserID == a.ID && v.Status != "confirmed" && v.Status != "cancelled" {
					pending++
				}
			}
			if pending >= 10 {
				return nil, ErrConflict
			}
			c.Purchases = append(c.Purchases, commerce.Purchase{ID: id, UserID: a.ID, QuoteID: q.ID, Status: "awaiting-confirmation", Version: 1, CreatedAt: now, PreviousQuotes: []string{}})
			p = &c.Purchases[len(c.Purchases)-1]
			w.Purchases = append(w.Purchases, PurchaseOwner{PurchaseID: id, UserID: a.ID})
		}
		for _, l := range q.Lines {
			found := false
			for i := range w.Offers {
				o := &w.Offers[i]
				if o.ID == l.OfferID {
					if o.Stock-o.Reserved < l.Quantity {
						return nil, ErrStock
					}
					o.Reserved += l.Quantity
					o.Version++
					found = true
				}
			}
			if !found {
				return nil, ErrStock
			}
		}
		q.Status = "reserved"
		q.AcceptedAt = &now
		q.ReservedUntil = now.Add(commerce.ReservationMinutes * time.Minute)
		q.ExpiresAt = q.ReservedUntil
		q.PurchaseID = p.ID
		if q.CartID != "" {
			for i := range c.Carts {
				if c.Carts[i].ID == q.CartID && c.Carts[i].UserID == a.ID {
					c.Carts[i].Lines = []commerce.CartLine{}
					c.Carts[i].Version++
					c.Carts[i].UpdatedAt = now
				}
			}
		}
		syncPurchaseOrders(w, p, *q, now)
		recordPurchaseEvent(p, a, "order.accepted", now)
		commerceAudit(w, id+":audit", a, "purchase.accept", p.ID, "", now)
		return purchaseView(c, *p), nil
	})
}
func (s *Service) ReadPurchases(ctx context.Context, a identity.User, id string) ([]PurchaseView, error) {
	if e := s.ReleaseExpiredCommerce(ctx); e != nil {
		return nil, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return nil, e
	}
	if e = customerActor(w, a); e != nil {
		return nil, e
	}
	c := commerceState(&w)
	out := []PurchaseView{}
	for _, p := range c.Purchases {
		if p.UserID == a.ID && (id == "" || p.ID == id) {
			out = append(out, purchaseView(c, p))
		}
	}
	if id != "" && len(out) == 0 {
		return nil, ErrNotFound
	}
	return out, nil
}
func (s *Service) CancelPurchase(ctx context.Context, a identity.User, id, reason, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "purchase-cancel:"+a.ID, key, struct{ ID, Reason string }{id, reason}, func(w Workspace) error { return customerActor(w, a) }, func(w *Workspace, op string, now time.Time) (any, error) {
		if !textBound(reason, 1000) {
			return nil, ErrInvalid
		}
		c := commerceState(w)
		p := findPurchase(c, id)
		if p == nil || p.UserID != a.ID {
			return nil, ErrNotFound
		}
		if p.Status == "confirmed" {
			if strings.TrimSpace(reason) == "" || !textBound(reason, 1000) {
				return nil, ErrInvalid
			}
			if !p.CancellationRequested {
				p.CancellationRequested = true
				p.CancellationReason = reason
				recordPurchaseEvent(p, a, "order.cancellation-requested", now)
				p.Version++
				commerceAudit(w, op, a, "purchase.cancel.request", p.ID, "", now)
			}
			return purchaseView(c, *p), nil
		}
		if p.Status == "cancelled" {
			return purchaseView(c, *p), nil
		}
		if q := findQuote(c, p.QuoteID); q != nil {
			releaseQuote(w, q)
			q.Status = "cancelled"
		}
		p.Status = "cancelled"
		p.Reason = strings.TrimSpace(reason)
		p.CancelledAt = &now
		p.Version++
		for i := range w.Orders {
			if w.Orders[i].CheckoutID == p.ID {
				w.Orders[i].Status = "cancelled"
				w.Orders[i].Version++
			}
		}
		recordPurchaseEvent(p, a, "order.cancelled", now)
		commerceAudit(w, op, a, "purchase.cancel", p.ID, "", now)
		return purchaseView(c, *p), nil
	})
}

type CheckoutPreview struct {
	Quote   QuoteView           `json:"quote"`
	Methods []MethodEligibility `json:"methods"`
}

func (s *Service) PreviewCheckout(ctx context.Context, a identity.User, in QuoteInput) (CheckoutPreview, error) {
	if e := s.ReleaseExpiredCommerce(ctx); e != nil {
		return CheckoutPreview{}, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return CheckoutPreview{}, e
	}
	if e = customerActor(w, a); e != nil {
		return CheckoutPreview{}, e
	}
	in.PaymentMethodID = ""
	q, e := buildQuote(w, a, in, "", s.clock.Now())
	if e != nil {
		return CheckoutPreview{}, e
	}
	v := CheckoutPreview{Quote: quoteView(q), Methods: []MethodEligibility{}}
	for _, m := range commerceState(&w).Settings.Methods {
		if !m.Active {
			continue
		}
		eligible := methodEligible(w, m, q.Lines, q.Total)
		reason := ""
		if !eligible {
			reason = "این روش برای همه اقلام سفارش مجاز نیست."
			if q.Total < m.Min {
				reason = "مبلغ سفارش از حداقل این روش کمتر است."
			}
			if m.Max > 0 && q.Total > m.Max {
				reason = "مبلغ کل سفارش از سقف این روش بیشتر است."
			}
		}
		v.Methods = append(v.Methods, MethodEligibility{methodInput(m), eligible, reason})
	}
	return v, nil
}
func (s *Service) PublicCommerceConfig(ctx context.Context) (map[string]any, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return nil, e
	}
	settings := commerceState(&w).Settings
	methods := []PaymentMethodInput{}
	destination := commerce.Destination{}
	for _, m := range settings.Methods {
		if m.Active {
			methods = append(methods, methodInput(m))
			if m.Kind == "card-transfer" {
				destination = settings.Destination
			}
		}
	}
	loc, _ := time.LoadLocation("Asia/Tehran")
	if loc == nil {
		loc = time.FixedZone("Tehran", 12600)
	}
	return map[string]any{"methods": methods, "destination": destination, "business": settings.Business, "currency": "toman", "reservation_minutes": commerce.ReservationMinutes, "server_date": s.clock.Now().In(loc).Format("2006-01-02")}, nil
}
