package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ShipmentQuery struct {
	Search, Store, Status string
	Page, Size            int
}
type ShipmentRow struct {
	ID             string     `json:"id"`
	Number         string     `json:"number"`
	PurchaseID     string     `json:"purchase_id"`
	PurchaseNumber string     `json:"purchase_number"`
	StoreID        string     `json:"store_id"`
	StoreName      string     `json:"store_name"`
	Status         string     `json:"status"`
	OrderStatus    string     `json:"order_status"`
	PaymentStatus  string     `json:"payment_status"`
	ShippingName   string     `json:"shipping_name"`
	Quantity       int        `json:"quantity"`
	StartedAt      *time.Time `json:"started_at"`
}
type ShipmentPage struct {
	Items  []ShipmentRow      `json:"items"`
	Stores []OrderStoreOption `json:"stores"`
	Total  int                `json:"total"`
	Page   int                `json:"page"`
	Size   int                `json:"size"`
}
type ShipmentDetail struct {
	Number     string              `json:"number"`
	Operations OrderOperationsView `json:"operations"`
}

func shipmentActor(w Workspace, a identity.User) (identity.User, error) {
	switch a.Surface {
	case identity.SurfaceBackoffice:
		return orderWorkspaceActor(w, a)
	case identity.SurfaceSeller:
		current, e := catalogActor(w, a)
		if e != nil {
			return a, e
		}
		for _, store := range identity.MemberStoreIDs(w.StoreMemberships, a.ID) {
			if w.Access != nil && identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, a.ID, store, "orders.read") {
				return current, nil
			}
		}
	}
	return a, ErrForbidden
}
func shipmentEligible(o orders.Order, confirmed bool) bool {
	return confirmed && o.ProcessingStartedAt != nil && (o.Payment == "paid" || o.Payment == "returned-to-wallet" && (o.Status == "cancelled" || o.OrderStatus == "cancelled" || o.Status == "delivered"))
}

func (s *Service) QueryShipments(ctx context.Context, a identity.User, q ShipmentQuery) (ShipmentPage, error) {
	result := ShipmentPage{Items: []ShipmentRow{}, Stores: []OrderStoreOption{}, Page: q.Page, Size: q.Size}
	if q.Page < 0 || q.Page > 1000000 || q.Size < 1 || q.Size > 100 || !textBound(q.Search, 150) || !textBound(q.Store, 128) || !slices.Contains([]string{"", "all", "awaiting-supply", "supplied", "packing", "shipped", "delivered", "cancelled", "on-hold", "supply-exception", "refund-pending"}, q.Status) {
		return result, ErrInvalid
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return result, e
	}
	a, e = shipmentActor(w, a)
	if e != nil {
		return result, e
	}
	for _, store := range w.Stores {
		if a.Surface == identity.SurfaceSeller && (w.Access == nil || !identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, a.ID, store.ID, "orders.read")) {
			continue
		}
		result.Stores = append(result.Stores, OrderStoreOption{store.ID, store.Name})
	}
	needle := strings.ToLower(normalizeDigits(q.Search))
	for _, o := range w.Orders {
		live, p, e := liveOrder(&w, o.ID)
		if e != nil || !shipmentEligible(*live, p.ConfirmedAt != nil) {
			continue
		}
		if _, e = orderAccess(w, a, o, *p); e != nil {
			continue
		}
		if q.Store != "" && q.Store != o.StoreID {
			continue
		}
		terminal := o.Status == "delivered" || o.Status == "cancelled" || o.OrderStatus == "cancelled"
		if q.Status == "" && terminal || q.Status != "" && q.Status != "all" && q.Status != o.Status {
			continue
		}
		row := ShipmentRow{ID: o.ID, Number: strconv.FormatInt(recordNumber(&w, "order", o.ID), 10), PurchaseID: p.ID, PurchaseNumber: strconv.FormatInt(p.Number, 10), StoreID: o.StoreID, StoreName: o.StoreDisplayName, Status: o.Status, OrderStatus: o.OrderStatus, PaymentStatus: o.Payment, ShippingName: o.ShippingMethod, StartedAt: o.ProcessingStartedAt}
		for _, store := range result.Stores {
			if store.ID == o.StoreID && row.StoreName == "" {
				row.StoreName = store.Name
			}
		}
		for _, item := range o.Items {
			row.Quantity += max(0, item.Quantity-item.CancelledQuantity)
		}
		if needle != "" && !strings.Contains(strings.ToLower(row.Number+" "+row.PurchaseNumber+" "+row.StoreName), needle) {
			continue
		}
		result.Items = append(result.Items, row)
	}
	sort.SliceStable(result.Items, func(i, j int) bool {
		a, b := result.Items[i], result.Items[j]
		if a.StartedAt.Equal(*b.StartedAt) {
			return a.ID < b.ID
		}
		return a.StartedAt.After(*b.StartedAt)
	})
	result.Total = len(result.Items)
	start := min(q.Page*q.Size, result.Total)
	end := min(start+q.Size, result.Total)
	result.Items = result.Items[start:end]
	return result, nil
}
func (s *Service) ReadShipment(ctx context.Context, a identity.User, id string) (ShipmentDetail, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return ShipmentDetail{}, e
	}
	a, e = shipmentActor(w, a)
	if e != nil {
		return ShipmentDetail{}, e
	}
	o, p, e := liveOrder(&w, id)
	if e != nil {
		return ShipmentDetail{}, e
	}
	a, e = orderAccess(w, a, *o, *p)
	if e != nil {
		return ShipmentDetail{}, e
	}
	if !shipmentEligible(*o, p.ConfirmedAt != nil) {
		return ShipmentDetail{}, ErrNotFound
	}
	return ShipmentDetail{Number: strconv.FormatInt(recordNumber(&w, "order", o.ID), 10), Operations: orderOperationsView(w, a, *o, *p)}, nil
}
