package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/finance"
	"github.com/cherriva/marketplace/apps/api/internal/domain/support"
)

func applySupport(state *Workspace, c Command, id string) error {
	if c.Reason == "" {
		return ErrInvalid
	}
	if c.Action == "order.deliver" {
		for i := range state.Orders {
			o := &state.Orders[i]
			if o.ID != c.Target {
				continue
			}
			if o.CheckoutID != "" {
				return ErrForbidden
			}
			if o.Version != c.Version || o.Status != "shipped" {
				return ErrConflict
			}
			o.Status = "delivered"
			o.Version++
			return nil
		}
		return ErrNotFound
	}
	for i := range state.Cases {
		ticket := &state.Cases[i]
		if ticket.ID != c.Target {
			continue
		}
		if ticket.Version != c.Version || ticket.Status == "closed" {
			return ErrConflict
		}
		switch c.Action {
		case "case.start":
			if ticket.Status != "open" {
				return ErrConflict
			}
			ticket.Status = "in-progress"
		case "case.close":
			ticket.Status = "closed"
		case "case.assign":
			active := false
			for _, seller := range state.Stores {
				if seller.ID == c.Value && seller.Active {
					active = true
				}
			}
			if !active {
				return ErrInvalid
			}
			state.Assignments = append(state.Assignments, support.Assignment{ID: id, CaseID: ticket.ID, StoreID: c.Value, ItemID: ticket.ItemID, Status: "awaiting-confirmation"})
			ticket.Status = "assigned"
		case "case.refund":
			var amount int64
			var seller string
			for j := range state.Orders {
				order := &state.Orders[j]
				if order.ID != ticket.OrderID {
					continue
				}
				if order.CheckoutID != "" {
					return ErrForbidden
				}
				seller = order.StoreID
				for k := range order.Items {
					item := &order.Items[k]
					if item.ID == ticket.ItemID {
						if item.Status == "refunded" || item.Status == "refund-pending" {
							return ErrConflict
						}
						amount = item.UnitPrice * int64(item.Quantity)
						item.Status = "refund-pending"
					}
				}
			}
			for _, transfer := range state.Transfers {
				if transfer.Kind == "refund" && transfer.ItemID == ticket.ItemID && transfer.Status != "failure" {
					return ErrConflict
				}
			}
			if amount <= 0 {
				return ErrInvalid
			}
			state.Transfers = append(state.Transfers, finance.Transfer{ID: id, Kind: "refund", OrderID: ticket.OrderID, ItemID: ticket.ItemID, StoreID: seller, Amount: amount, Status: "pending", Reserved: true, Source: "TEST-ORDER-" + ticket.OrderID, Version: 1})
			ticket.Status = "refund-pending"
		default:
			return ErrInvalid
		}
		ticket.Reason = c.Reason
		ticket.Version++
		return nil
	}
	return ErrNotFound
}
