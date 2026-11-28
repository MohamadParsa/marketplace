package backoffice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"slices"
	"strings"
	"time"
)

type OrderRefundInput struct {
	ID                  string `json:"request_id"`
	Version             int    `json:"version"`
	Method              string `json:"method"`
	AcceptedAmountToman string `json:"accepted_amount_toman"`
	Instructions        string `json:"instructions"`
	Reason              string `json:"reason"`
}
type OrderRefundSummary struct {
	Method           string      `json:"method"`
	AmountToman      string      `json:"amount_toman"`
	ShippingToman    string      `json:"shipping_toman"`
	Status           string      `json:"status"`
	At               time.Time   `json:"at"`
	Withdrawal       *RefundView `json:"withdrawal,omitempty"`
	SettlementStatus string      `json:"settlement_status,omitempty"`
}

func findCaseRefund(c *commerce.State, id string) *commerce.OrderRefund {
	for i := range c.OrderRefunds {
		if c.OrderRefunds[i].RequestID == id {
			return &c.OrderRefunds[i]
		}
	}
	return nil
}
func caseRefundSummary(c *commerce.State, f commerce.OrderRefund, financial bool) OrderRefundSummary {
	v := OrderRefundSummary{Method: f.Method, AmountToman: commerce.Toman(f.Amount), ShippingToman: commerce.Toman(f.Shipping), Status: "wallet-returned", At: f.At}
	if financial {
		v.SettlementStatus = f.SettlementStatus
	}
	if f.WithdrawalID != "" {
		for _, r := range c.Refunds {
			if r.ID == f.WithdrawalID {
				withdrawal := refundView(r)
				v.Withdrawal = &withdrawal
				switch r.Status {
				case "pending":
					v.Status = "bank-pending"
				case "completed":
					v.Status = "bank-completed"
				default:
					v.Status = "wallet-returned"
				}
			}
		}
	}
	return v
}
func refundDiscount(q *commerce.Quote, key string) (total, platform, seller commerce.Amount) {
	if q.Discount != nil {
		for _, a := range q.Discount.Allocations {
			if a.Key == key {
				total += a.Amount
				platform += a.Platform
				seller += a.Seller
			}
		}
	}
	return
}

func caseRefundPlan(w *Workspace, p *commerce.Purchase, r *commerce.OrderRequest) (commerce.OrderRefund, error) {
	c := w.Commerce
	f := commerce.OrderRefund{PurchaseID: p.ID, RequestID: r.ID, UserID: p.UserID, Adjustments: []commerce.RefundAdjustment{}, SettlementStatus: "review-required"}
	if p.Status != "confirmed" || p.ConfirmedAt == nil || r.Status != "approved" || findCaseRefund(c, r.ID) != nil {
		return f, ErrConflict
	}
	q := findQuote(c, p.QuoteID)
	if q == nil {
		return f, ErrConflict
	}
	for _, e := range c.Entries {
		if e.Source == "order-cancel:"+p.ID {
			return f, ErrConflict
		}
	}
	for _, line := range r.Items {
		o, item, e := sourceOrderItem(w, line.OrderID, line.ItemID)
		if e != nil {
			return f, e
		}
		var snapshot *commerce.LineSnapshot
		for i := range q.Lines {
			l := &q.Lines[i]
			if l.OfferID == line.OfferID && l.StoreID == o.StoreID {
				snapshot = l
				break
			}
		}
		if snapshot == nil || item.OfferID != snapshot.OfferID || len(line.UnitIndices) != line.Quantity {
			return f, ErrConflict
		}
		adjustment := commerce.RefundAdjustment{OrderID: o.ID, ItemID: item.ID, StoreID: o.StoreID, UnitIndices: slices.Clone(line.UnitIndices)}
		seen := map[int]bool{}
		for _, unit := range line.UnitIndices {
			if unit < 0 || unit >= snapshot.Quantity || seen[unit] {
				return f, ErrConflict
			}
			seen[unit] = true
			discount, platform, seller := refundDiscount(q, fmt.Sprintf("item:%s:%03d", line.OfferID, unit))
			if discount < 0 || discount > snapshot.UnitPrice || platform+seller != discount {
				return f, ErrConflict
			}
			adjustment.Gross += snapshot.UnitPrice
			adjustment.DiscountPlatform += platform
			adjustment.DiscountSeller += seller
			adjustment.Net += snapshot.UnitPrice - discount
		}
		f.Amount += adjustment.Net
		f.Adjustments = append(f.Adjustments, adjustment)
	}
	allCancelled := len(p.OrderIDs) > 0
	for _, id := range p.OrderIDs {
		o, _, e := liveOrder(w, id)
		if e != nil || o.Status != "cancelled" || o.DispatchedAt != nil || o.DeliveredAt != nil {
			allCancelled = false
		}
	}
	shippingUsed := false
	for _, other := range c.OrderRefunds {
		if other.PurchaseID == p.ID && other.ShippingIncluded {
			shippingUsed = true
		}
	}
	if r.Kind == "cancel" && allCancelled && !shippingUsed {
		f.ShippingIncluded = true
		for _, ship := range q.Shipments {
			if ship.Method.CollectionMode == "collect" {
				continue
			}
			discount, platform, seller := refundDiscount(q, "shipping:"+ship.StoreID)
			net := ship.Method.Fee - discount
			if net < 0 {
				return f, ErrConflict
			}
			var order string
			for _, id := range p.OrderIDs {
				o, _, _ := liveOrder(w, id)
				if o != nil && o.StoreID == ship.StoreID {
					order = o.ID
				}
			}
			f.Shipping += net
			f.Adjustments = append(f.Adjustments, commerce.RefundAdjustment{OrderID: order, StoreID: ship.StoreID, Gross: ship.Method.Fee, DiscountPlatform: platform, DiscountSeller: seller, Net: net})
		}
		f.Amount += f.Shipping
	}
	var returned commerce.Amount
	for _, other := range c.OrderRefunds {
		if other.PurchaseID == p.ID {
			returned += other.Amount
		}
	}
	if f.Amount < 0 || f.Amount > q.Total-returned {
		return f, ErrConflict
	}
	return f, nil
}

func refundOriginalSources(c *commerce.State, q *commerce.Quote, amount commerce.Amount, bank bool) ([]commerce.CreditAllocation, error) {
	if amount == 0 {
		return []commerce.CreditAllocation{}, nil
	}
	if q.Credit == nil || q.Credit.Status != "posted" || creditSum(combinedCredit(q)) != q.Total {
		return nil, ErrConflict
	}
	spent := map[string]commerce.Amount{}
	for _, f := range c.OrderRefunds {
		if f.PurchaseID == q.PurchaseID {
			for _, a := range f.OriginalSources {
				spent[a.LotID] += a.Amount
			}
		}
	}
	lots := map[string]commerce.CreditLot{}
	for _, l := range c.CreditLots {
		lots[l.ID] = l
	}
	result := []commerce.CreditAllocation{}
	for _, allocation := range combinedCredit(q) {
		lot, ok := lots[allocation.LotID]
		if !ok || lot.UserID != q.UserID {
			return nil, ErrConflict
		}
		if bank && !lot.CanWithdraw() {
			continue
		}
		free := allocation.Amount - spent[lot.ID]
		if free < 0 {
			return nil, ErrConflict
		}
		take := min(free, amount)
		if take > 0 {
			result = append(result, commerce.CreditAllocation{LotID: lot.ID, Amount: take})
			amount -= take
		}
		if amount == 0 {
			break
		}
	}
	if amount != 0 {
		if bank {
			return nil, ErrWithdrawalUnavailable
		}
		return nil, ErrConflict
	}
	return result, nil
}

func (s *Service) RefundOrderRequest(ctx context.Context, a identity.User, in OrderRefundInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "order-refund:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, CommerceRefundsResolve) }, func(w *Workspace, op string, now time.Time) (any, error) {
		a, e := currentActor(*w, a)
		if e != nil {
			return nil, e
		}
		c := commerceState(w)
		r := findOrderRequest(c, in.ID)
		if r == nil {
			return nil, ErrNotFound
		}
		p := findPurchase(c, r.PurchaseID)
		if p == nil || r.Version != in.Version {
			return nil, ErrConflict
		}
		if in.Method != "wallet" && in.Method != "bank" || strings.TrimSpace(in.Reason) == "" || !textBound(in.Reason, 1000) || !textBound(in.Instructions, 1000) {
			return nil, ErrInvalid
		}
		f, e := caseRefundPlan(w, p, r)
		if e != nil {
			return nil, e
		}
		amount, e := commerce.ParseToman(normalizeDigits(in.AcceptedAmountToman))
		if e != nil || amount != f.Amount {
			return nil, ErrConflict
		}
		if in.Method == "bank" && (amount == 0 || strings.TrimSpace(in.Instructions) == "") {
			return nil, ErrInvalid
		}
		q := findQuote(c, p.QuoteID)
		sources, e := refundOriginalSources(c, q, amount, in.Method == "bank")
		if e != nil {
			return nil, e
		}
		f.ID = op
		f.Method = in.Method
		f.At = now
		f.ActorID = a.ID
		f.Reason = in.Reason
		f.OriginalSources = sources
		returned := []commerce.CreditAllocation{}
		if amount > 0 {
			entry := op + ":credit"
			if e = appendWallet(c, entry, p.UserID, "order-refund:"+r.ID, "order-cancel-credit", amount, now); e != nil {
				return nil, e
			}
			f.CreditEntryID = entry
			// Replace the generic credit lot with lots retaining each original restriction and FIFO age.
			if len(c.CreditLots) == 0 || c.CreditLots[len(c.CreditLots)-1].ID != entry {
				return nil, ErrConflict
			}
			c.CreditLots = c.CreditLots[:len(c.CreditLots)-1]
			for i, source := range sources {
				var original commerce.CreditLot
				for _, lot := range c.CreditLots {
					if lot.ID == source.LotID {
						original = lot
						break
					}
				}
				withdrawable := original.CanWithdraw()
				priority := original.AvailableSince()
				id := fmt.Sprintf("%s:%d", entry, i)
				c.CreditLots = append(c.CreditLots, commerce.CreditLot{ID: id, EntryID: entry, UserID: p.UserID, Source: "order-refund:" + r.ID, OriginLotID: original.ID, Amount: source.Amount, Withdrawable: &withdrawable, PriorityAt: &priority, At: now, ActorID: a.ID, Title: "بازگشت اعتبار سفارش", Description: original.Title})
				returned = append(returned, commerce.CreditAllocation{LotID: id, Amount: source.Amount})
			}
		}
		if in.Method == "bank" {
			f.WithdrawalID = op + ":withdrawal"
			c.Refunds = append(c.Refunds, commerce.RefundRequest{ID: f.WithdrawalID, PurchaseID: p.ID, RequestID: r.ID, UserID: p.UserID, Amount: amount, Instructions: strings.TrimSpace(in.Instructions), Allocations: returned, Status: "pending", Version: 1, CreatedAt: now})
		}
		c.OrderRefunds = append(c.OrderRefunds, f)
		r.Version++
		p.Version++
		recordPurchaseEvent(p, a, "refund."+in.Method, now)
		event := &p.Events[len(p.Events)-1]
		event.ID = op
		event.RequestID = r.ID
		commerceAudit(w, op, a, "order-refund."+in.Method, r.ID, "", now)
		return caseRefundSummary(c, f, true), nil
	})
}
