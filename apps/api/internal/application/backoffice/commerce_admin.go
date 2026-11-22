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

type DepositReviewInput struct {
	ID                  string `json:"id"`
	Version             int    `json:"version"`
	Approve             bool   `json:"approve"`
	Reason              string `json:"reason"`
	VerifiedReference   string `json:"verified_reference"`
	VerifiedDate        string `json:"verified_date"`
	VerifiedAmountToman string `json:"verified_amount_toman"`
}
type PurchaseAdminInput struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Action  string `json:"action"`
	Reason  string `json:"reason"`
}
type RefundResolveInput struct {
	ID        string `json:"id"`
	Version   int    `json:"version"`
	Complete  bool   `json:"complete"`
	Reference string `json:"reference"`
	Date      string `json:"date"`
	Reason    string `json:"reason"`
}
type AdminDepositView struct {
	DepositView
	UserID        string `json:"user_id"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
}
type AdminPurchaseView struct {
	PurchaseView
	UserID        string `json:"user_id"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
}
type AdminRefundView struct {
	RefundView
	UserID        string `json:"user_id"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
}
type CommerceAdminView struct {
	Purchases   []AdminPurchaseView `json:"purchases"`
	Deposits    []AdminDepositView  `json:"deposits"`
	Refunds     []AdminRefundView   `json:"refunds"`
	Audit       []Audit             `json:"audit"`
	Permissions map[string]bool     `json:"permissions"`
	Currency    string              `json:"currency"`
}

func (s *Service) ReadCommerceAdmin(ctx context.Context, a identity.User) (CommerceAdminView, error) {
	if a.Surface != identity.SurfaceBackoffice {
		return CommerceAdminView{}, ErrForbidden
	}
	if e := s.ReleaseExpiredCommerce(ctx); e != nil {
		return CommerceAdminView{}, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return CommerceAdminView{}, e
	}
	a, e = currentActor(w, a)
	if e != nil {
		return CommerceAdminView{}, e
	}
	v := CommerceAdminView{Purchases: []AdminPurchaseView{}, Deposits: []AdminDepositView{}, Refunds: []AdminRefundView{}, Audit: []Audit{}, Permissions: map[string]bool{}, Currency: "toman"}
	for _, p := range []string{CommerceSettingsRead, CommercePaymentWrite, CommerceShippingWrite, CommerceGeographyWrite, CommerceBusinessWrite, CommerceOrdersRead, CommerceOrdersConfirm, CommerceDepositsRead, CommerceDepositsReview, CommerceRefundsResolve} {
		v.Permissions[p] = allowsAction(a, p)
	}
	if !v.Permissions[CommerceOrdersRead] && !v.Permissions[CommerceDepositsRead] && !v.Permissions[CommerceSettingsRead] && !v.Permissions[CommerceRefundsResolve] {
		return v, ErrForbidden
	}
	c := commerceState(&w)
	name := func(id string) (string, string) {
		for _, u := range w.Users {
			if u.ID == id {
				return u.Name, u.Phone
			}
		}
		return "", ""
	}
	if v.Permissions[CommerceOrdersRead] {
		for _, p := range c.Purchases {
			n, phone := name(p.UserID)
			view := purchaseView(c, p)
			if !v.Permissions[CommerceDepositsRead] {
				view.Deposits = []DepositView{}
			}
			if !v.Permissions[CommerceOrdersConfirm] {
				view.WalletBalanceToman = ""
				view.WalletAvailableToman = ""
				view.ShortfallToman = ""
			}
			v.Purchases = append(v.Purchases, AdminPurchaseView{view, p.UserID, n, phone})
		}
	}
	if v.Permissions[CommerceDepositsRead] {
		for _, d := range c.Deposits {
			n, phone := name(d.UserID)
			v.Deposits = append(v.Deposits, AdminDepositView{depositView(d), d.UserID, n, phone})
		}
	}
	if v.Permissions[CommerceRefundsResolve] {
		for _, r := range c.Refunds {
			n, phone := name(r.UserID)
			v.Refunds = append(v.Refunds, AdminRefundView{refundView(r), r.UserID, n, phone})
		}
	}
	if v.Permissions[CommerceDepositsRead] || v.Permissions[CommerceOrdersConfirm] || v.Permissions[CommerceRefundsResolve] {
		for _, a := range w.Audit {
			if strings.HasPrefix(a.Action, "commerce.") {
				v.Audit = append(v.Audit, a)
			}
		}
	}
	return v, nil
}
func (s *Service) ReviewDeposit(ctx context.Context, a identity.User, in DepositReviewInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "deposit-review:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, CommerceDepositsReview) }, func(w *Workspace, op string, now time.Time) (any, error) {
		c := commerceState(w)
		for i := range c.Deposits {
			d := &c.Deposits[i]
			if d.ID != in.ID {
				continue
			}
			if d.Status == "approved" && in.Approve {
				ref, e := normalizeReference(in.VerifiedReference)
				if e == nil && ref == d.VerifiedReference && in.VerifiedDate == d.VerifiedDate && in.VerifiedAmountToman == commerce.Toman(d.Amount) {
					return depositView(*d), nil
				}
				return nil, ErrConflict
			}
			if d.Version != in.Version || d.Status != "pending" {
				return nil, ErrConflict
			}
			if !textBound(in.Reason, 1000) {
				return nil, ErrInvalid
			}
			if in.Approve {
				ref, e := normalizeReference(in.VerifiedReference)
				amount, ae := commerce.ParseToman(normalizeDigits(in.VerifiedAmountToman))
				if e != nil || ae != nil || amount != d.Amount || !validDepositDate(in.VerifiedDate, now) {
					return nil, ErrInvalid
				}
				verifiedKey := referenceKey(d.Method, d.Destination, in.VerifiedDate, ref)
				for _, other := range c.Deposits {
					if other.ID != d.ID && other.VerifiedKey == verifiedKey && other.Status == "approved" {
						return nil, ErrConflict
					}
				}
				if e = appendWallet(c, op+":credit", d.UserID, "deposit:"+d.ID, "deposit-credit", d.Amount, now); e != nil {
					return nil, e
				}
				d.Status = "approved"
				d.VerifiedReference = ref
				d.VerifiedDate = in.VerifiedDate
				d.VerifiedKey = verifiedKey
			} else {
				if strings.TrimSpace(in.Reason) == "" {
					return nil, ErrInvalid
				}
				d.Status = "rejected"
			}
			d.Reason = in.Reason
			d.ReviewedAt = &now
			d.ReviewedBy = a.ID
			d.Version++
			if p := findPurchase(c, d.PurchaseID); p != nil {
				recordPurchaseEvent(p, a, "deposit."+d.Status, now)
			}
			commerceAudit(w, op, a, "deposit.review", d.ID, "", now)
			return depositView(*d), nil
		}
		return nil, ErrNotFound
	})
}
func (s *Service) ConfirmPurchase(ctx context.Context, a identity.User, in PurchaseAdminInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "purchase-confirm:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, CommerceOrdersConfirm) }, func(w *Workspace, op string, now time.Time) (any, error) {
		c := commerceState(w)
		p := findPurchase(c, in.ID)
		if p == nil {
			return nil, ErrNotFound
		}
		if p.Status == "confirmed" {
			return purchaseView(c, *p), nil
		}
		if p.Version != in.Version {
			return nil, ErrConflict
		}
		if p.Status == "cancelled" {
			return nil, ErrConflict
		}
		q := findQuote(c, p.QuoteID)
		if q == nil {
			return nil, ErrConflict
		}
		if q.Status != "reserved" || !now.Before(q.ReservedUntil) {
			p.Status = "requires-reconfirmation"
			p.Version++
			recordPurchaseEvent(p, a, "order.reconfirmation-required", now)
			return purchaseView(c, *p), nil
		}
		eligible := false
		for _, m := range c.Settings.Methods {
			if m.ID == q.PaymentMethod.ID && methodEligible(*w, m, q.Lines, q.Total) {
				eligible = true
			}
		}
		if !eligible {
			releaseQuote(w, q)
			q.Status = "invalidated"
			p.Status = "requires-reconfirmation"
			p.Reason = "روش پرداخت یا محدودیت آن تغییر کرده است."
			p.Version++
			recordPurchaseEvent(p, a, "order.reconfirmation-required", now)
			return purchaseView(c, *p), nil
		}
		for _, l := range q.Lines {
			o, e := findOffer(*w, l.OfferID)
			copyWorkspace := *w
			copyWorkspace.Offers = append([]catalog.Offer{}, w.Offers...)
			for i := range copyWorkspace.Offers {
				if copyWorkspace.Offers[i].ID == l.OfferID {
					copyWorkspace.Offers[i].Reserved -= l.Quantity
				}
			}
			item, itemError := cartItem(copyWorkspace, commerce.CartLine{OfferID: l.OfferID, Quantity: l.Quantity})
			if e != nil || itemError != nil || !item.Available || o.ProductID != l.ProductID || o.VariantID != l.VariantID || o.Stock < l.Quantity || o.Reserved < l.Quantity {
				releaseQuote(w, q)
				q.Status = "invalidated"
				p.Status = "requires-reconfirmation"
				p.Reason = "موجودی نیاز به بررسی دوباره دارد."
				p.Version++
				recordPurchaseEvent(p, a, "order.reconfirmation-required", now)
				return purchaseView(c, *p), nil
			}
		}
		if e := appendWallet(c, op+":debit", p.UserID, "order:"+p.ID, "order-payment", -q.Total, now); e != nil {
			return nil, e
		}
		for _, l := range q.Lines {
			for i := range w.Offers {
				if w.Offers[i].ID == l.OfferID {
					w.Offers[i].Stock -= l.Quantity
					w.Offers[i].Reserved -= l.Quantity
					w.Offers[i].Version++
				}
			}
		}
		q.Status = "consumed"
		p.Status = "confirmed"
		p.ConfirmedAt = &now
		p.Version++
		for i := range w.Orders {
			if w.Orders[i].CheckoutID == p.ID {
				w.Orders[i].Payment = "paid"
				w.Orders[i].Status = "awaiting-supply"
				w.Orders[i].Version++
				for j := range w.Orders[i].Items {
					w.Orders[i].Items[j].Status = "awaiting-supply"
				}
			}
		}
		recordPurchaseEvent(p, a, "order.paid", now)
		commerceAudit(w, op, a, "purchase.confirm", p.ID, "", now)
		return purchaseView(c, *p), nil
	})
}
func (s *Service) AdminCancelPurchase(ctx context.Context, a identity.User, in PurchaseAdminInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "purchase-admin-cancel:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, CommerceOrdersConfirm) }, func(w *Workspace, op string, now time.Time) (any, error) {
		c := commerceState(w)
		p := findPurchase(c, in.ID)
		if p == nil {
			return nil, ErrNotFound
		}
		if p.Status == "cancelled" {
			return purchaseView(c, *p), nil
		}
		if p.Version != in.Version || strings.TrimSpace(in.Reason) == "" || !textBound(in.Reason, 1000) {
			return nil, ErrConflict
		}
		q := findQuote(c, p.QuoteID)
		if q == nil {
			return nil, ErrConflict
		}
		if p.Status == "confirmed" {
			for _, order := range w.Orders {
				if order.CheckoutID == p.ID && order.Status != "awaiting-supply" {
					return nil, ErrConflict
				}
			}
			if e := appendWallet(c, op+":credit", p.UserID, "order-cancel:"+p.ID, "order-cancel-credit", q.Total, now); e != nil {
				return nil, e
			}
			for _, l := range q.Lines {
				for i := range w.Offers {
					if w.Offers[i].ID == l.OfferID {
						if w.Offers[i].Stock > MaxOfferStock-l.Quantity {
							return nil, ErrInvalid
						}
						w.Offers[i].Stock += l.Quantity
						w.Offers[i].Version++
					}
				}
			}
		} else {
			releaseQuote(w, q)
			q.Status = "cancelled"
		}
		p.Status = "cancelled"
		p.CancelledAt = &now
		p.Reason = in.Reason
		p.Version++
		for i := range w.Orders {
			if w.Orders[i].CheckoutID == p.ID {
				w.Orders[i].Status = "cancelled"
				w.Orders[i].Version++
				if p.ConfirmedAt != nil {
					w.Orders[i].Payment = "returned-to-wallet"
				}
			}
		}
		recordPurchaseEvent(p, a, "order.cancelled", now)
		commerceAudit(w, op, a, "purchase.admin-cancel", p.ID, "", now)
		return purchaseView(c, *p), nil
	})
}
func (s *Service) ResolveRefund(ctx context.Context, a identity.User, in RefundResolveInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "refund-resolve:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, CommerceRefundsResolve) }, func(w *Workspace, op string, now time.Time) (any, error) {
		c := commerceState(w)
		for i := range c.Refunds {
			r := &c.Refunds[i]
			if r.ID != in.ID {
				continue
			}
			if r.Status == "completed" && in.Complete {
				ref, e := normalizeReference(in.Reference)
				if e == nil && r.Reference == ref && r.Date == in.Date {
					return refundView(*r), nil
				}
				return nil, ErrConflict
			}
			if r.Version != in.Version || r.Status != "pending" {
				return nil, ErrConflict
			}
			if !textBound(in.Reason, 1000) {
				return nil, ErrInvalid
			}
			if in.Complete {
				ref, e := normalizeReference(in.Reference)
				if e != nil || !validDepositDate(in.Date, now) || strings.TrimSpace(in.Reason) == "" {
					return nil, ErrInvalid
				}
				for _, other := range c.Refunds {
					if other.ID != r.ID && other.Status == "completed" && other.Reference == ref && other.Date == in.Date {
						return nil, ErrConflict
					}
				}
				r.Status = "completed"
				if e = appendWallet(c, op+":debit", r.UserID, "refund:"+r.ID, "manual-refund", -r.Amount, now); e != nil {
					return nil, e
				}
				r.Reference = ref
				r.Date = in.Date
			} else {
				if strings.TrimSpace(in.Reason) == "" {
					return nil, ErrInvalid
				}
				r.Status = "rejected"
			}
			r.Reason = in.Reason
			r.ResolvedAt = &now
			r.ResolvedBy = a.ID
			r.Version++
			commerceAudit(w, op, a, "refund.resolve", r.ID, "", now)
			return refundView(*r), nil
		}
		return nil, ErrNotFound
	})
}
