package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strconv"
)

func publicOrderEvent(p commerce.Purchase, e commerce.OrderEvent, index int) OrderEventView {
	id := e.ID
	if id == "" {
		id = "legacy:" + hashToken(p.ID+":"+strconv.Itoa(index))
	}
	description := e.CustomerDescription
	if description == "" {
		description = orderCustomerDescription(e.Action)
	}
	return OrderEventView{CustomerDescription: description, Version: e.Version, ID: id, OrderID: e.OrderID, RequestID: e.RequestID, From: e.FromStatus, To: e.ToStatus, Action: e.Action, FlowStatus: e.FlowStatus, OrderStatus: e.OrderStatus, PaymentStatus: e.PaymentStatus, ActorKind: e.ActorKind, At: e.At}
}

type OrderHistoryEvent struct {
	OrderEventView
	Sequence      int    `json:"sequence"`
	StoreName     string `json:"store_name,omitempty"`
	RequestNumber string `json:"request_number,omitempty"`
	ActorName     string `json:"actor_name,omitempty"`
	Reason        string `json:"reason,omitempty"`
}
type OrderHistoryPage struct {
	Events     []OrderHistoryEvent `json:"events"`
	NextBefore int                 `json:"next_before,omitempty"`
}

func (s *Service) ReadOrderHistory(ctx context.Context, a identity.User, id string, before, limit int) (OrderHistoryPage, error) {
	result := OrderHistoryPage{Events: []OrderHistoryEvent{}}
	if before < 0 || limit < 1 || limit > 50 {
		return result, ErrInvalid
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return result, e
	}
	p, e := requestPurchase(&w, id)
	if e != nil {
		return result, e
	}
	a, e = requestActor(w, a, *p)
	if e != nil {
		return result, e
	}
	private := a.Surface == identity.SurfaceBackoffice && (allowsAction(a, OrdersCorrect) || allowsAction(a, OrdersRequestsReview) || allowsAction(a, CommerceRefundsResolve))
	for i := len(p.Events) - 1; i >= 0; i-- {
		if before != 0 && i+1 >= before {
			continue
		}
		event := p.Events[i]
		entry := OrderHistoryEvent{OrderEventView: publicOrderEvent(*p, event, i), Sequence: i + 1}
		if event.OrderID != "" {
			o, _, err := liveOrder(&w, event.OrderID)
			if err != nil {
				continue
			}
			if _, err = orderAccess(w, a, *o, *p); err != nil {
				continue
			}
			entry.StoreName = o.StoreDisplayName
		} else if a.Surface == identity.SurfaceSeller {
			continue
		}
		if len(result.Events) == limit {
			result.NextBefore = result.Events[len(result.Events)-1].Sequence
			break
		}
		if r := findOrderRequest(w.Commerce, event.RequestID); r != nil {
			entry.RequestNumber = strconv.FormatInt(r.Number, 10)
		}
		if private {
			entry.Reason = event.Reason
			for _, u := range w.Users {
				if u.ID == event.ActorID {
					entry.ActorName = u.Name
					break
				}
			}
		}
		result.Events = append(result.Events, entry)
	}
	return result, nil
}
