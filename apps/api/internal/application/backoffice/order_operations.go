package backoffice

import (
	"context"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
)

const OrdersOperate = "orders.operate"
const OrdersReceive = "orders.receive"
const OrdersCorrect = "orders.correct"

type OrderOperationInput struct {
	OrderID            string   `json:"order_id,omitempty"`
	PurchaseID         string   `json:"-"`
	Version            int      `json:"version"`
	Action             string   `json:"action"`
	Reason             string   `json:"reason"`
	DispatchIdentifier string   `json:"dispatch_identifier,omitempty"`
	Identifiers        []string `json:"identifiers,omitempty"`
	AllItems           bool     `json:"all_items"`
}

type OrderOperationEvent struct {
	CustomerDescription string    `json:"customer_description,omitempty"`
	ID                  string    `json:"id,omitempty"`
	Action              string    `json:"action"`
	From                string    `json:"from"`
	To                  string    `json:"to"`
	ActorKind           string    `json:"actor_kind"`
	At                  time.Time `json:"at"`
	Reason              string    `json:"reason,omitempty"`
}

type OrderOperationsView struct {
	StoreID            string                `json:"store_id"`
	ID                 string                `json:"id"`
	PurchaseID         string                `json:"purchase_id"`
	PurchaseNumber     string                `json:"purchase_number"`
	StoreName          string                `json:"store_name"`
	Version            int                   `json:"version"`
	OrderStatus        string                `json:"order_status"`
	Status             string                `json:"status"`
	PaymentStatus      string                `json:"payment_status"`
	ShippingName       string                `json:"shipping_name"`
	DispatchIdentifier string                `json:"dispatch_identifier"`
	Identifiers        []string              `json:"identifiers"`
	Items              []orders.Item         `json:"items"`
	Address            *commerce.Address     `json:"address,omitempty"`
	AllowedActions     []string              `json:"allowed_actions"`
	Events             []OrderOperationEvent `json:"events"`
}

func liveOrder(w *Workspace, id string) (*orders.Order, *commerce.Purchase, error) {
	c := commerceState(w)
	for i := range w.Orders {
		o := &w.Orders[i]
		if o.ID != id || o.CheckoutID == "" {
			continue
		}
		p := findPurchase(c, o.CheckoutID)
		if p != nil && slices.Contains(p.OrderIDs, o.ID) {
			return o, p, nil
		}
	}
	return nil, nil, ErrNotFound
}

func orderAccess(w Workspace, a identity.User, o orders.Order, p commerce.Purchase) (identity.User, error) {
	switch a.Surface {
	case identity.SurfaceCustomer:
		if err := customerActor(w, a); err != nil {
			return a, err
		}
		if p.UserID != a.ID {
			return a, ErrNotFound
		}
		return a, nil
	case identity.SurfaceSeller:
		current, err := catalogActor(w, a)
		if err != nil {
			return a, err
		}
		if !slices.Contains(identity.MemberStoreIDs(w.StoreMemberships, a.ID), o.StoreID) {
			return a, ErrNotFound
		}
		if w.Access == nil || !identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, a.ID, o.StoreID, "orders.read") {
			return a, ErrForbidden
		}
		return current, nil
	case identity.SurfaceBackoffice:
		current, err := currentActor(w, a)
		if err != nil {
			return a, err
		}
		if !allowsAction(current, CommerceOrdersRead) {
			return a, ErrForbidden
		}
		return current, nil
	}
	return a, ErrForbidden
}

func canOrderOperate(w Workspace, a identity.User, store string) bool {
	if a.Surface == identity.SurfaceSeller {
		return w.Access != nil && identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, a.ID, store, OrdersOperate)
	}
	return a.Surface == identity.SurfaceBackoffice && allowsAction(a, OrdersOperate)
}

func requireOrderAction(w Workspace, a identity.User, o orders.Order, action string) error {
	switch action {
	case "start", "supply", "pack", "ship":
		if canOrderOperate(w, a, o.StoreID) {
			return nil
		}
	case "deliver":
		if a.Surface == identity.SurfaceCustomer || a.Surface == identity.SurfaceBackoffice && allowsAction(a, OrdersReceive) {
			return nil
		}
	case "configure-dispatch", "correct":
		if a.Surface == identity.SurfaceBackoffice && allowsAction(a, OrdersCorrect) {
			return nil
		}
	default:
		return ErrInvalid
	}
	return ErrForbidden
}

func orderSupplyComplete(o orders.Order) bool {
	any := false
	for _, item := range o.Items {
		if item.Quantity <= item.CancelledQuantity {
			continue
		}
		any = true
		if !slices.Contains([]string{"supplied", "packing", "shipped", "delivered"}, item.Status) {
			return false
		}
	}
	return any
}

func orderDispatchPolicy(o orders.Order) string {
	if o.DispatchIdentifier != "" {
		return o.DispatchIdentifier
	}
	if o.ShippingTerms != nil {
		return o.ShippingTerms.DispatchIdentifier
	}
	return ""
}

func orderHasSupplyHold(w Workspace, o orders.Order) bool {
	if o.Status == "on-hold" {
		return true
	}
	for _, item := range o.Items {
		if item.Quantity > item.CancelledQuantity && slices.Contains([]string{"supply-exception", "refund-pending", "on-hold"}, item.Status) {
			return true
		}
	}
	for _, c := range w.Cases {
		if c.OrderID == o.ID && c.Status != "closed" {
			return true
		}
	}
	return false
}

func operationActions(w Workspace, a identity.User, o orders.Order) []string {
	result := []string{}
	if o.Payment != "paid" || o.Status == "cancelled" {
		return result
	}
	if canOrderOperate(w, a, o.StoreID) && !orderHasSupplyHold(w, o) {
		switch o.Status {
		case "awaiting-supply":
			if o.OrderStatus == "processing" {
				result = append(result, "supply")
			} else {
				result = append(result, "start")
			}
		case "supplied":
			result = append(result, "pack")
		case "packing":
			if orderDispatchPolicy(o) != "" {
				result = append(result, "ship")
			}
		}
	}
	if o.Status == "shipped" && (a.Surface == identity.SurfaceCustomer || a.Surface == identity.SurfaceBackoffice && allowsAction(a, OrdersReceive)) {
		result = append(result, "deliver")
	}
	if a.Surface == identity.SurfaceBackoffice && allowsAction(a, OrdersCorrect) {
		if beforeShipping(o.Status) && orderDispatchPolicy(o) == "" {
			result = append(result, "configure-dispatch")
		}
		if o.Status == "shipped" || o.Status == "delivered" {
			result = append(result, "correct")
		}
	}
	return result
}

func orderOperationsView(w Workspace, a identity.User, o orders.Order, p commerce.Purchase) OrderOperationsView {
	v := OrderOperationsView{StoreID: o.StoreID, ID: o.ID, PurchaseID: p.ID, PurchaseNumber: o.PurchaseNumber, StoreName: o.StoreDisplayName, Version: o.Version, OrderStatus: o.OrderStatus, Status: o.Status, PaymentStatus: o.Payment, ShippingName: o.ShippingMethod, DispatchIdentifier: orderDispatchPolicy(o), Identifiers: slices.Clone(o.DispatchIdentifiers), Items: slices.Clone(o.Items), AllowedActions: operationActions(w, a, o), Events: []OrderOperationEvent{}}
	for _, s := range w.Stores {
		if s.ID == o.StoreID && v.StoreName == "" {
			v.StoreName = s.Name
		}
	}
	if p.ConfirmedAt == nil {
		v.AllowedActions = []string{}
	}
	if a.Surface != identity.SurfaceSeller || slices.Contains([]string{"supplied", "packing", "shipped", "delivered"}, o.Status) && orderSupplyComplete(o) {
		if q := findQuote(w.Commerce, p.QuoteID); q != nil {
			address := q.Address
			if a.Surface == identity.SurfaceSeller {
				address.ID = ""
				address.UserID = ""
				address.RecipientNationalID = ""
			}
			v.Address = &address
		}
	}
	for _, e := range p.Events {
		if e.OrderID == o.ID {
			reason := e.Reason
			if a.Surface == identity.SurfaceCustomer || a.Surface == identity.SurfaceSeller && e.RequestID != "" {
				reason = ""
			}
			description := e.CustomerDescription
			if description == "" {
				description = orderCustomerDescription(e.Action)
			}
			v.Events = append(v.Events, OrderOperationEvent{CustomerDescription: description, ID: e.ID, Action: e.Action, From: e.FromStatus, To: e.ToStatus, ActorKind: e.ActorKind, At: e.At, Reason: reason})
		}
	}
	return v
}

func (s *Service) ReadOrderOperations(ctx context.Context, a identity.User, id, purchase string) (OrderOperationsView, error) {
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return OrderOperationsView{}, err
	}
	o, p, err := liveOrder(&w, id)
	if err != nil {
		return OrderOperationsView{}, err
	}
	if purchase != "" && p.ID != purchase && strconv.FormatInt(p.Number, 10) != purchase {
		return OrderOperationsView{}, ErrNotFound
	}
	a, err = orderAccess(w, a, *o, *p)
	if err != nil {
		return OrderOperationsView{}, err
	}
	return orderOperationsView(w, a, *o, *p), nil
}

func normalizeDispatchIdentifiers(values []string, policy string) ([]string, error) {
	if !slices.Contains([]string{"none", "tracking", "delivery"}, policy) || len(values) > 20 {
		return nil, ErrInvalid
	}
	ids := []string{}
	for _, value := range values {
		v := normalizeDigits(strings.TrimSpace(value))
		if v == "" || !textBound(v, 160) || strings.ContainsFunc(v, unicode.IsControl) || slices.Contains(ids, v) {
			return nil, ErrInvalid
		}
		ids = append(ids, v)
	}
	if policy == "none" && len(ids) > 0 || policy != "none" && len(ids) == 0 {
		return nil, ErrInvalid
	}
	return ids, nil
}

func recordOrderOperation(p *commerce.Purchase, o orders.Order, a identity.User, id, action, from, reason string, now time.Time) {
	recordPurchaseEvent(p, a, "fulfillment."+action, now)
	e := &p.Events[len(p.Events)-1]
	e.ID = id
	e.OrderID = o.ID
	e.FromStatus = from
	e.ToStatus = o.Status
	e.Reason = reason
}

func (s *Service) TransitionOrder(ctx context.Context, a identity.User, in OrderOperationInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "order-operation:"+a.ID, key, in, func(w Workspace) error {
		o, p, e := liveOrder(&w, in.OrderID)
		if e != nil {
			return e
		}
		if in.PurchaseID != "" && p.ID != in.PurchaseID && strconv.FormatInt(p.Number, 10) != in.PurchaseID {
			return ErrNotFound
		}
		current, e := orderAccess(w, a, *o, *p)
		if e != nil {
			return e
		}
		return requireOrderAction(w, current, *o, in.Action)
	}, func(w *Workspace, op string, now time.Time) (any, error) {
		o, p, e := liveOrder(w, in.OrderID)
		if e != nil {
			return nil, e
		}
		a, e = orderAccess(*w, a, *o, *p)
		if e != nil {
			return nil, e
		}
		if o.Version != in.Version {
			return nil, ErrConflict
		}
		if !textBound(in.Reason, 1000) {
			return nil, ErrInvalid
		}
		if p.ConfirmedAt == nil || p.Status == "cancelled" || !slices.Contains(operationActions(*w, a, *o), in.Action) {
			return nil, ErrConflict
		}
		from := o.Status
		previousIdentifiers := slices.Clone(o.DispatchIdentifiers)
		switch in.Action {
		case "start":
			o.OrderStatus = "processing"
			o.ProcessingStartedAt = &now
		case "supply", "pack":
			status := "supplied"
			if in.Action == "pack" {
				status = "packing"
			}
			o.Status = status
			for i := range o.Items {
				if o.Items[i].Quantity > o.Items[i].CancelledQuantity {
					o.Items[i].Status = status
				}
			}
		case "ship":
			if !in.AllItems {
				return nil, ErrInvalid
			}
			ids, e := normalizeDispatchIdentifiers(in.Identifiers, orderDispatchPolicy(*o))
			if e != nil {
				return nil, e
			}
			o.DispatchIdentifiers = ids
			o.Packages = slices.Clone(ids)
			o.Status = "shipped"
			o.DispatchedAt = &now
			for i := range o.Items {
				if o.Items[i].Quantity > o.Items[i].CancelledQuantity {
					o.Items[i].Status = "shipped"
				}
			}
		case "deliver":
			if !in.AllItems || a.Surface == identity.SurfaceBackoffice && strings.TrimSpace(in.Reason) == "" {
				return nil, ErrInvalid
			}
			o.Status = "delivered"
			o.DeliveredAt = &now
			for i := range o.Items {
				if o.Items[i].Quantity > o.Items[i].CancelledQuantity {
					o.Items[i].Status = "delivered"
				}
			}
		case "configure-dispatch":
			if strings.TrimSpace(in.Reason) == "" || !slices.Contains([]string{"none", "tracking", "delivery"}, in.DispatchIdentifier) {
				return nil, ErrInvalid
			}
			o.DispatchIdentifier = in.DispatchIdentifier
		case "correct":
			if strings.TrimSpace(in.Reason) == "" {
				return nil, ErrInvalid
			}
			ids, e := normalizeDispatchIdentifiers(in.Identifiers, orderDispatchPolicy(*o))
			if e != nil {
				return nil, e
			}
			o.DispatchIdentifiers = ids
			o.Packages = slices.Clone(ids)
		default:
			return nil, ErrInvalid
		}
		o.Version++
		p.Version++
		refreshPurchaseOrderState(w, p)
		recordOrderOperation(p, *o, a, op, in.Action, from, in.Reason, now)
		commerceAudit(w, op, a, "order."+in.Action, o.ID, in.Reason, now)
		p.Events[len(p.Events)-1].PreviousIdentifiers = previousIdentifiers
		p.Events[len(p.Events)-1].Identifiers = slices.Clone(o.DispatchIdentifiers)
		return orderOperationsView(*w, a, *o, *p), nil
	})
}
