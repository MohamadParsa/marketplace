package backoffice

import (
	"context"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
)

const OrdersRequestsWrite = "orders.requests.write"
const OrdersRequestsReview = "orders.requests.review"

type RequestedItemInput struct {
	OrderID  string `json:"order_id"`
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}
type OrderRequestInput struct {
	PurchaseID string               `json:"purchase_id"`
	Version    int                  `json:"version"`
	Kind       string               `json:"kind"`
	Reason     string               `json:"reason"`
	Items      []RequestedItemInput `json:"items"`
}
type OrderRequestReview struct {
	ID      string `json:"request_id"`
	Version int    `json:"version"`
	Action  string `json:"action"`
	Reason  string `json:"reason"`
	Healthy *bool  `json:"healthy,omitempty"`
}
type RequestItemView struct {
	OrderID         string `json:"order_id"`
	ItemID          string `json:"item_id"`
	Name            string `json:"name"`
	StoreName       string `json:"store_name"`
	Quantity        int    `json:"quantity"`
	AvailableCancel int    `json:"available_cancel"`
	AvailableReturn int    `json:"available_return"`
}
type OrderRequestView struct {
	Refund          *OrderRefundSummary `json:"refund,omitempty"`
	RefundableToman string              `json:"refundable_toman,omitempty"`
	CanRefund       bool                `json:"can_refund"`
	CanBank         bool                `json:"can_bank"`
	ID              string              `json:"id"`
	Number          string              `json:"number"`
	Kind            string              `json:"kind"`
	Status          string              `json:"status"`
	Version         int                 `json:"version"`
	Reason          string              `json:"reason,omitempty"`
	ReviewReason    string              `json:"review_reason,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	ReviewedAt      *time.Time          `json:"reviewed_at,omitempty"`
	GoodsReceived   bool                `json:"goods_received"`
	Inspected       bool                `json:"inspected"`
	Healthy         bool                `json:"healthy"`
	StockApplied    bool                `json:"stock_applied"`
	Items           []RequestItemView   `json:"items"`
	AllowedActions  []string            `json:"allowed_actions"`
}
type OrderRequestDirectory struct {
	PurchaseID      string             `json:"purchase_id"`
	PurchaseNumber  string             `json:"purchase_number"`
	Version         int                `json:"version"`
	CanCreate       bool               `json:"can_create"`
	CanReview       bool               `json:"can_review"`
	LegacyRequested bool               `json:"legacy_requested"`
	LegacyReason    string             `json:"legacy_reason,omitempty"`
	Items           []RequestItemView  `json:"items"`
	Requests        []OrderRequestView `json:"requests"`
}

func requestActor(w Workspace, a identity.User, p commerce.Purchase) (identity.User, error) {
	if a.Surface == identity.SurfaceCustomer {
		if e := customerActor(w, a); e != nil {
			return a, e
		}
		if a.ID != p.UserID {
			return a, ErrNotFound
		}
		return a, nil
	}
	if a.Surface == identity.SurfaceBackoffice {
		current, e := currentActor(w, a)
		if e != nil {
			return a, e
		}
		if !allowsAction(current, CommerceOrdersRead) {
			return a, ErrForbidden
		}
		return current, nil
	}
	if a.Surface == identity.SurfaceSeller {
		current, e := catalogActor(w, a)
		if e != nil {
			return a, e
		}
		for _, o := range w.Orders {
			if slices.Contains(p.OrderIDs, o.ID) {
				if _, e = orderAccess(w, current, o, p); e == nil {
					return current, nil
				}
			}
		}
		return a, ErrNotFound
	}
	return a, ErrForbidden
}
func canCreateOrderRequest(a identity.User) bool {
	return a.Surface == identity.SurfaceCustomer || a.Surface == identity.SurfaceBackoffice && allowsAction(a, OrdersRequestsWrite)
}
func canReviewOrderRequest(a identity.User) bool {
	return a.Surface == identity.SurfaceBackoffice && allowsAction(a, OrdersRequestsReview)
}

func requestPurchase(w *Workspace, id string) (*commerce.Purchase, error) {
	c := commerceState(w)
	for i := range c.Purchases {
		p := &c.Purchases[i]
		if p.ID == id || strconv.FormatInt(p.Number, 10) == id {
			return p, nil
		}
	}
	return nil, ErrNotFound
}
func findOrderRequest(c *commerce.State, id string) *commerce.OrderRequest {
	for i := range c.OrderRequests {
		if c.OrderRequests[i].ID == id {
			return &c.OrderRequests[i]
		}
	}
	return nil
}
func hasLiveOrderRequest(c *commerce.State, purchase string) bool {
	for _, r := range c.OrderRequests {
		if r.PurchaseID == purchase && r.Status != "rejected" {
			return true
		}
	}
	return false
}

func requestUnits(c *commerce.State, order, item string) map[int]bool {
	used := map[int]bool{}
	for _, r := range c.OrderRequests {
		if r.Status == "rejected" {
			continue
		}
		for _, line := range r.Items {
			if line.OrderID == order && line.ItemID == item {
				for _, unit := range line.UnitIndices {
					used[unit] = true
				}
			}
		}
	}
	return used
}
func sourceOrderItem(w *Workspace, order, item string) (*orders.Order, *orders.Item, error) {
	o, _, e := liveOrder(w, order)
	if e != nil {
		return nil, nil, e
	}
	for i := range o.Items {
		if o.Items[i].ID == item {
			return o, &o.Items[i], nil
		}
	}
	return nil, nil, ErrNotFound
}
func requestItemView(w Workspace, o orders.Order, item orders.Item, quantity int) RequestItemView {
	name := item.Name
	if name == "" {
		name = "قلم سفارش"
	}
	store := o.StoreDisplayName
	for _, s := range w.Stores {
		if s.ID == o.StoreID && store == "" {
			store = s.Name
		}
	}
	return RequestItemView{OrderID: o.ID, ItemID: item.ID, Name: name, StoreName: store, Quantity: quantity}
}
func requestReviewActions(a identity.User, r commerce.OrderRequest) []string {
	result := []string{}
	if !canReviewOrderRequest(a) {
		return result
	}
	if r.Status == "requested" {
		result = append(result, "approve", "reject")
	}
	if r.Kind == "return" && r.Status != "rejected" && !r.StockApplied {
		if !r.GoodsReceived {
			result = append(result, "receive")
		} else {
			result = append(result, "inspect")
		}
	}
	return result
}
func orderRequestView(w *Workspace, a identity.User, p commerce.Purchase, r commerce.OrderRequest) OrderRequestView {
	v := OrderRequestView{ID: r.ID, Number: strconv.FormatInt(r.Number, 10), Kind: r.Kind, Status: r.Status, Version: r.Version, Reason: r.Reason, ReviewReason: r.ReviewReason, CreatedAt: r.CreatedAt, ReviewedAt: r.ReviewedAt, GoodsReceived: r.GoodsReceived, Inspected: r.Inspected, Healthy: r.Healthy, StockApplied: r.StockApplied, Items: []RequestItemView{}, AllowedActions: requestReviewActions(a, r)}
	for _, line := range r.Items {
		if o, item, e := sourceOrderItem(w, line.OrderID, line.ItemID); e == nil {
			if _, e = orderAccess(*w, a, *o, p); e == nil {
				v.Items = append(v.Items, requestItemView(*w, *o, *item, line.Quantity))
			}
		}
	}
	if a.Surface == identity.SurfaceSeller {
		v.Reason = ""
		v.ReviewReason = ""
	} else {
		financial := a.Surface == identity.SurfaceBackoffice && allowsAction(a, CommerceRefundsResolve)
		if f := findCaseRefund(w.Commerce, r.ID); f != nil {
			summary := caseRefundSummary(w.Commerce, *f, financial)
			v.Refund = &summary
		} else if financial {
			if plan, e := caseRefundPlan(w, &p, &r); e == nil {
				v.RefundableToman = commerce.Toman(plan.Amount)
				q := findQuote(w.Commerce, p.QuoteID)
				_, e = refundOriginalSources(w.Commerce, q, plan.Amount, false)
				v.CanRefund = e == nil
				_, e = refundOriginalSources(w.Commerce, q, plan.Amount, true)
				v.CanBank = e == nil && plan.Amount > 0
			}
		}
	}
	return v
}

func (s *Service) ReadOrderRequests(ctx context.Context, a identity.User, id string) (OrderRequestDirectory, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return OrderRequestDirectory{}, e
	}
	p, e := requestPurchase(&w, id)
	if e != nil {
		return OrderRequestDirectory{}, e
	}
	a, e = requestActor(w, a, *p)
	if e != nil {
		return OrderRequestDirectory{}, e
	}
	v := OrderRequestDirectory{PurchaseID: p.ID, PurchaseNumber: strconv.FormatInt(p.Number, 10), Version: p.Version, CanCreate: canCreateOrderRequest(a) && p.ConfirmedAt != nil && p.Status != "cancelled", CanReview: canReviewOrderRequest(a), Items: []RequestItemView{}, Requests: []OrderRequestView{}}
	for _, o := range w.Orders {
		if !slices.Contains(p.OrderIDs, o.ID) {
			continue
		}
		if _, e := orderAccess(w, a, o, *p); e != nil {
			continue
		}
		for _, item := range o.Items {
			line := requestItemView(w, o, item, item.Quantity)
			available := max(0, item.Quantity-len(requestUnits(w.Commerce, o.ID, item.ID)))
			if p.ConfirmedAt != nil && p.Status != "cancelled" {
				if beforeShipping(o.Status) {
					line.AvailableCancel = available
				}
				if o.Status == "shipped" || o.Status == "delivered" {
					line.AvailableReturn = available
				}
			}
			v.Items = append(v.Items, line)
		}
	}
	count := 0
	for _, r := range w.Commerce.OrderRequests {
		if r.PurchaseID != p.ID {
			continue
		}
		count++
		entry := orderRequestView(&w, a, *p, r)
		if len(entry.Items) > 0 {
			v.Requests = append(v.Requests, entry)
		}
	}
	if a.Surface != identity.SurfaceSeller {
		v.LegacyRequested = p.CancellationRequested && count == 0
		v.LegacyReason = p.CancellationReason
	} else {
		v.Version = 0
	}
	return v, nil
}

func appendOrderRequest(w *Workspace, a identity.User, p *commerce.Purchase, in OrderRequestInput, id, origin string, now time.Time) (*commerce.OrderRequest, error) {
	if p.ConfirmedAt == nil || p.Status == "cancelled" {
		return nil, ErrConflict
	}
	if in.Kind != "cancel" && in.Kind != "return" || len(in.Items) == 0 || len(in.Items) > 200 || strings.TrimSpace(in.Reason) == "" || !textBound(in.Reason, 1000) {
		return nil, ErrInvalid
	}
	r := commerce.OrderRequest{ID: id, PurchaseID: p.ID, UserID: p.UserID, Kind: in.Kind, Origin: origin, Status: "requested", Version: 1, Reason: strings.TrimSpace(in.Reason), CreatedAt: now, CreatedBy: a.ID, Items: []commerce.OrderRequestItem{}}
	seen := map[string]bool{}
	for _, input := range in.Items {
		key := input.OrderID + "\x00" + input.ItemID
		if seen[key] || input.Quantity < 1 {
			return nil, ErrInvalid
		}
		seen[key] = true
		o, item, e := sourceOrderItem(w, input.OrderID, input.ItemID)
		if e != nil {
			return nil, e
		}
		if !slices.Contains(p.OrderIDs, o.ID) {
			return nil, ErrNotFound
		}
		if in.Kind == "cancel" && !beforeShipping(o.Status) || in.Kind == "return" && o.Status != "shipped" && o.Status != "delivered" {
			return nil, ErrConflict
		}
		if item.OfferID == "" {
			return nil, ErrConflict
		}
		used := requestUnits(w.Commerce, o.ID, item.ID)
		indices := []int{}
		for i := 0; i < item.Quantity && len(indices) < input.Quantity; i++ {
			if !used[i] {
				indices = append(indices, i)
			}
		}
		if len(indices) != input.Quantity {
			return nil, ErrConflict
		}
		r.Items = append(r.Items, commerce.OrderRequestItem{OrderID: o.ID, ItemID: item.ID, OfferID: item.OfferID, Quantity: input.Quantity, UnitIndices: indices})
	}
	c := w.Commerce
	if c.NextRequestNumber < 1000 {
		c.NextRequestNumber = 1000
	}
	if c.NextRequestNumber == 9223372036854775807 {
		return nil, ErrInvalid
	}
	r.Number = c.NextRequestNumber
	c.NextRequestNumber++
	c.OrderRequests = append(c.OrderRequests, r)
	p.CancellationRequested = true
	p.CancellationReason = r.Reason
	p.Version++
	recordPurchaseEvent(p, a, "request.created", now)
	event := &p.Events[len(p.Events)-1]
	event.ID = hashToken(id + ":created")
	event.RequestID = id
	return &c.OrderRequests[len(c.OrderRequests)-1], nil
}

func (s *Service) CreateOrderRequest(ctx context.Context, a identity.User, in OrderRequestInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "order-request-create:"+a.ID, key, in, func(w Workspace) error {
		p, e := requestPurchase(&w, in.PurchaseID)
		if e != nil {
			return e
		}
		current, e := requestActor(w, a, *p)
		if e != nil {
			return e
		}
		if !canCreateOrderRequest(current) {
			return ErrForbidden
		}
		return nil
	}, func(w *Workspace, op string, now time.Time) (any, error) {
		p, e := requestPurchase(w, in.PurchaseID)
		if e != nil {
			return nil, e
		}
		a, e = requestActor(*w, a, *p)
		if e != nil {
			return nil, e
		}
		if p.Version != in.Version {
			return nil, ErrConflict
		}
		r, e := appendOrderRequest(w, a, p, in, op, "request", now)
		if e != nil {
			return nil, e
		}
		commerceAudit(w, op, a, "order.request.create", r.ID, "", now)
		return orderRequestView(w, a, *p, *r), nil
	})
}

func applyRequestStock(w *Workspace, p *commerce.Purchase, r *commerce.OrderRequest, a identity.User, op string, now time.Time) error {
	if r.StockApplied {
		return nil
	}
	if r.Status != "approved" || r.Kind == "return" && (!r.GoodsReceived || !r.Inspected || !r.Healthy) {
		return nil
	}
	changed := map[string]string{}
	for _, line := range r.Items {
		o, item, e := sourceOrderItem(w, line.OrderID, line.ItemID)
		if e != nil {
			return e
		}
		if r.Kind == "cancel" && !beforeShipping(o.Status) {
			return ErrConflict
		}
		if _, ok := changed[o.ID]; !ok {
			changed[o.ID] = o.Status
		}
		found := false
		for i := range w.Offers {
			offer := &w.Offers[i]
			if offer.ID == line.OfferID {
				if offer.Stock > MaxOfferStock-line.Quantity {
					return ErrInvalid
				}
				offer.Stock += line.Quantity
				offer.Version++
				found = true
				break
			}
		}
		if !found {
			return ErrConflict
		}
		if r.Kind == "cancel" {
			if item.CancelledQuantity > item.Quantity-line.Quantity {
				return ErrConflict
			}
			item.CancelledQuantity += line.Quantity
			if item.CancelledQuantity == item.Quantity {
				item.Status = "cancelled"
			}
		} else {
			if item.ReturnedQuantity > item.Quantity-line.Quantity {
				return ErrConflict
			}
			item.ReturnedQuantity += line.Quantity
		}
	}
	for id, from := range changed {
		o, _, e := liveOrder(w, id)
		if e != nil {
			return e
		}
		o.Version++
		if r.Kind == "cancel" {
			all := true
			for _, item := range o.Items {
				if item.CancelledQuantity < item.Quantity {
					all = false
				}
			}
			if all {
				o.Status = "cancelled"
				o.OrderStatus = "cancelled"
				releaseCancelledOrderCapacity(w, p.QuoteID, o.StoreID, now)
			}
		}
		refreshPurchaseOrderState(w, p)
		action := "items-cancelled"
		if r.Kind == "return" {
			action = "stock-returned"
		}
		recordOrderOperation(p, *o, a, hashToken(op+id), action, from, r.ReviewReason, now)
		p.Events[len(p.Events)-1].RequestID = r.ID
	}
	r.StockApplied = true
	return nil
}

func releaseCancelledOrderCapacity(w *Workspace, quote, store string, now time.Time) {
	for i := range w.Commerce.ShippingReservations {
		r := &w.Commerce.ShippingReservations[i]
		if r.QuoteID == quote && r.StoreID == store && r.Status == "committed" {
			r.Status = "released"
			r.ReleasedAt = &now
		}
	}
}

func (s *Service) ReviewOrderRequest(ctx context.Context, a identity.User, in OrderRequestReview, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "order-request-review:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, OrdersRequestsReview) }, func(w *Workspace, op string, now time.Time) (any, error) {
		c := commerceState(w)
		r := findOrderRequest(c, in.ID)
		if r == nil {
			return nil, ErrNotFound
		}
		p := findPurchase(c, r.PurchaseID)
		if p == nil || p.Status == "cancelled" {
			return nil, ErrConflict
		}
		a, e := currentActor(*w, a)
		if e != nil {
			return nil, e
		}
		if r.Version != in.Version {
			return nil, ErrConflict
		}
		if strings.TrimSpace(in.Reason) == "" || !textBound(in.Reason, 1000) {
			return nil, ErrInvalid
		}
		if !slices.Contains(requestReviewActions(a, *r), in.Action) {
			return nil, ErrConflict
		}
		switch in.Action {
		case "approve":
			if r.Kind == "cancel" {
				for _, line := range r.Items {
					o, _, e := sourceOrderItem(w, line.OrderID, line.ItemID)
					if e != nil {
						return nil, e
					}
					if !beforeShipping(o.Status) {
						return nil, ErrConflict
					}
				}
			}
			r.Status = "approved"
			r.ReviewedAt = &now
			r.ReviewedBy = a.ID
			r.ReviewReason = strings.TrimSpace(in.Reason)
		case "reject":
			r.Status = "rejected"
			r.ReviewedAt = &now
			r.ReviewedBy = a.ID
			r.ReviewReason = strings.TrimSpace(in.Reason)
		case "receive":
			r.GoodsReceived = true
			r.ReceivedAt = &now
			if in.Healthy != nil {
				r.Inspected = true
				r.Healthy = *in.Healthy
				r.InspectedAt = &now
			}
		case "inspect":
			if in.Healthy == nil {
				return nil, ErrInvalid
			}
			r.Inspected = true
			r.Healthy = *in.Healthy
			r.InspectedAt = &now
		default:
			return nil, ErrInvalid
		}
		if e := applyRequestStock(w, p, r, a, op, now); e != nil {
			return nil, e
		}
		r.Version++
		p.Version++
		p.CancellationRequested = false
		for _, other := range c.OrderRequests {
			if other.PurchaseID == p.ID && other.Status == "requested" {
				p.CancellationRequested = true
			}
		}
		refreshPurchaseOrderState(w, p)
		recordPurchaseEvent(p, a, "request."+in.Action, now)
		event := &p.Events[len(p.Events)-1]
		event.ID = op
		event.RequestID = r.ID
		event.Reason = in.Reason
		commerceAudit(w, op, a, "order.request."+in.Action, r.ID, in.Reason, now)
		return orderRequestView(w, a, *p, *r), nil
	})
}

func legacyOrderCancellationRequests(w *Workspace, a identity.User, p *commerce.Purchase, reason, op string, now time.Time) error {
	created := false
	for _, kind := range []string{"cancel", "return"} {
		in := OrderRequestInput{PurchaseID: p.ID, Kind: kind, Reason: reason}
		for _, o := range w.Orders {
			if !slices.Contains(p.OrderIDs, o.ID) {
				continue
			}
			matches := beforeShipping(o.Status)
			if kind == "return" {
				matches = o.Status == "shipped" || o.Status == "delivered"
			}
			if !matches {
				continue
			}
			for _, item := range o.Items {
				free := item.Quantity - len(requestUnits(w.Commerce, o.ID, item.ID))
				if free > 0 {
					in.Items = append(in.Items, RequestedItemInput{OrderID: o.ID, ItemID: item.ID, Quantity: free})
				}
			}
		}
		if len(in.Items) > 0 {
			if _, e := appendOrderRequest(w, a, p, in, hashToken(op+kind), "legacy-cancel", now); e != nil {
				return e
			}
			created = true
		}
	}
	if !created {
		return ErrConflict
	}
	return nil
}
