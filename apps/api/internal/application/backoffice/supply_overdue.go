package backoffice

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type SupplyOverdueRow struct {
	OrderID         string    `json:"order_id"`
	PurchaseNumber  string    `json:"purchase_number"`
	StoreID         string    `json:"store_id"`
	StoreName       string    `json:"store_name"`
	PaidAt          time.Time `json:"paid_at"`
	Deadline        time.Time `json:"deadline"`
	OverdueMinutes  int64     `json:"overdue_minutes"`
	LeadDays        int       `json:"lead_days"`
	DayBasis        string    `json:"day_basis"`
	CalendarVersion int       `json:"calendar_version"`
}
type SupplyOverdueView struct {
	CheckedAt            time.Time          `json:"checked_at"`
	Rows                 []SupplyOverdueRow `json:"rows"`
	MissingDeadlineCount int                `json:"missing_deadline_count"`
}

// A computed queue, never a state transition or a notification producer.
func (s *Service) ReadSupplyOverdue(ctx context.Context, a identity.User, storeID, search string) (SupplyOverdueView, error) {
	out := SupplyOverdueView{Rows: []SupplyOverdueRow{}, CheckedAt: s.clock.Now()}
	if a.Surface != identity.SurfaceBackoffice && a.Surface != identity.SurfaceSeller {
		return out, ErrForbidden
	}
	if len(storeID) > 128 || len([]rune(search)) > 160 {
		return out, ErrInvalid
	}
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return out, err
	}
	actor, err := catalogActor(w, a)
	if err != nil {
		return out, err
	}
	allowed := func(id string) bool {
		if actor.Surface == identity.SurfaceBackoffice {
			return allowsAction(actor, CommerceOrdersRead)
		}
		return w.Access != nil && identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, actor.ID, id, "orders.read")
	}
	if actor.Surface == identity.SurfaceBackoffice && !allowsAction(actor, CommerceOrdersRead) || storeID != "" && !allowed(storeID) {
		return out, ErrForbidden
	}
	search = strings.ToLower(normalizeDigits(strings.TrimSpace(search)))
	for _, order := range w.Orders {
		if order.Status != "awaiting-supply" || order.Payment != "paid" || !allowed(order.StoreID) || storeID != "" && order.StoreID != storeID {
			continue
		}
		name := order.StoreDisplayName
		if name == "" {
			for _, store := range w.Stores {
				if store.ID == order.StoreID {
					name = store.Name
					break
				}
			}
		}
		if search != "" && !strings.Contains(strings.ToLower(normalizeDigits(order.ID+" "+order.PurchaseNumber+" "+name)), search) {
			continue
		}
		snapshot := order.Preparation
		if snapshot == nil || snapshot.SourceAt.IsZero() || snapshot.ReadyAt.IsZero() || snapshot.ReadyAt.Before(snapshot.SourceAt) {
			out.MissingDeadlineCount++
			continue
		}
		if out.CheckedAt.Before(snapshot.ReadyAt) {
			continue
		}
		out.Rows = append(out.Rows, SupplyOverdueRow{OrderID: order.ID, PurchaseNumber: order.PurchaseNumber, StoreID: order.StoreID, StoreName: name, PaidAt: snapshot.SourceAt, Deadline: snapshot.ReadyAt, OverdueMinutes: int64(out.CheckedAt.Sub(snapshot.ReadyAt) / time.Minute), LeadDays: snapshot.LeadDays, DayBasis: string(snapshot.Policy.DayBasis), CalendarVersion: snapshot.Calendar.Version})
	}
	sort.Slice(out.Rows, func(i, j int) bool {
		if out.Rows[i].Deadline.Equal(out.Rows[j].Deadline) {
			return out.Rows[i].OrderID < out.Rows[j].OrderID
		}
		return out.Rows[i].Deadline.Before(out.Rows[j].Deadline)
	})
	return out, nil
}
