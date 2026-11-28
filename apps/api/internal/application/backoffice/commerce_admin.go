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
	VerifiedAmountToman string `json:"verified_amount_toman,omitempty"`
	ID                  string `json:"id"`
	Version             int    `json:"version"`
	Complete            bool   `json:"complete"`
	Reference           string `json:"reference"`
	Date                string `json:"date"`
	Reason              string `json:"reason"`
}
type AdminDepositView struct {
	DepositView
	UserID        string `json:"user_id"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
}
type DiscountFundingView struct {
	PlatformToman string `json:"platform_toman"`
	SellerToman   string `json:"seller_toman"`
	RuleVersion   int    `json:"rule_version"`
}
type CreditSourceView struct {
	LotID        string `json:"lot_id"`
	Title        string `json:"title,omitempty"`
	Withdrawable bool   `json:"withdrawable"`
	OriginLotID  string `json:"origin_lot_id,omitempty"`
	Source       string `json:"source"`
	AmountToman  string `json:"amount_toman"`
	Kind         string `json:"kind"`
	Status       string `json:"status"`
}
type AdminPurchaseView struct {
	DiscountFunding *DiscountFundingView `json:"discount_funding,omitempty"`
	CreditSources   []CreditSourceView   `json:"credit_sources,omitempty"`
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
				view.PaymentAttempts = nil
				view.Quote.Destination = commerce.Destination{}
				view.Quote.PaymentMethod.Instructions = ""
			}
			if !v.Permissions[CommerceOrdersConfirm] {
				view.WalletBalanceToman = ""
				view.WalletAvailableToman = ""
				view.ShortfallToman = ""
				view.CreditSources = nil
			}
			entry := AdminPurchaseView{PurchaseView: view, UserID: p.UserID, CustomerName: n, CustomerPhone: phone}
			if q := findQuote(c, p.QuoteID); q != nil {
				if q.Discount != nil && allowsAction(a, CouponsRead) {
					var platform, seller commerce.Amount
					for _, v := range q.Discount.Allocations {
						platform += v.Platform
						seller += v.Seller
					}
					entry.DiscountFunding = &DiscountFundingView{commerce.Toman(platform), commerce.Toman(seller), q.Discount.Coupon.Version}
				}
				if q.Credit != nil && v.Permissions[CommerceOrdersConfirm] {
					entry.CreditSources = creditSourceViews(c, q)
				}
			}

			v.Purchases = append(v.Purchases, entry)
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
				earmarkDeposit(c, *d, op+":credit")
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
			updateFundingAttempt(c, *d, now)
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
		return finalizePurchase(w, a, in, op, now)
	})
}

// Called inside the placement/confirmation transaction, after the caller's authority is checked.
func finalizePurchase(w *Workspace, a identity.User, in PurchaseAdminInput, op string, now time.Time) (any, error) {
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
	eligible := q.Total == 0 || q.PaymentMethod.ID == "" && q.ContractVersion == 2 && q.Credit != nil && q.Total == q.Credit.Requested
	for _, m := range c.Settings.Methods {
		if m.ID == q.PaymentMethod.ID && methodEligible(*w, m, q.Lines, q.Total) && salesPaymentStillAllowed(*w, *q) {
			eligible = true
		}
	}
	eligible = eligible && storePoliciesStillAllowed(*w, *q) && shippingStillAllowed(*w, *q)
	if q.ContractVersion == 2 && q.Credit != nil && q.Credit.Requested > 0 && !walletEligible(*w, q.Lines, q.Total) {
		eligible = false
	}
	if !eligible {
		releaseQuote(w, q, now)
		q.Status = "invalidated"
		p.Status = "requires-reconfirmation"
		p.Reason = "روش پرداخت، ارسال یا محدودیت آن تغییر کرده است."
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
			releaseQuote(w, q, now)
			q.Status = "invalidated"
			p.Status = "requires-reconfirmation"
			p.Reason = "موجودی نیاز به بررسی دوباره دارد."
			p.Version++
			recordPurchaseEvent(p, a, "order.reconfirmation-required", now)
			return purchaseView(c, *p), nil
		}
	}
	finalized := *q
	finalized.Shipments = append([]commerce.ShipmentSnapshot{}, q.Shipments...)
	shippingError := finalizeShipmentCalendars(&finalized, now)
	if shippingError == nil {
		shippingError = commitShipping(w, finalized, now)
	}
	if shippingError != nil {
		releaseQuote(w, q, now)
		q.Status = "invalidated"
		p.Status = "requires-reconfirmation"
		p.Reason = "فرصت آماده‌سازی یا رزرو ظرفیت تغییر کرده؛ روش و موعد تحویل باید دوباره انتخاب و تأیید شوند."
		p.Version++
		recordPurchaseEvent(p, a, "order.reconfirmation-required", now)
		return purchaseView(c, *p), nil
	}
	q.Shipments = finalized.Shipments
	if q.ContractVersion == 2 {
		if q.Credit == nil || q.Credit.Status != "held" {
			return nil, ErrConflict
		}
		allocations := combinedCredit(q)
		if creditSum(allocations) != q.Total {
			return nil, ErrInsufficientWallet
		}
		q.Credit.Status = "posted"
		if q.Total > 0 {
			if e := appendWallet(c, op+":debit", p.UserID, "order:"+p.ID, "order-payment", -q.Total, now, allocations); e != nil {
				return nil, e
			}
		}
	} else if q.Total > 0 {
		if err := ensureCreditState(c, p.UserID, now); err != nil {
			return nil, err
		}
		allocations, err := allocateCredit(c, p.UserID, q.Total, "")
		if err != nil {
			return nil, err
		}
		q.Credit = &commerce.CreditUsage{Requested: q.Total, Existing: allocations, Funding: []commerce.CreditAllocation{}, Status: "posted"}
		if err = appendWallet(c, op+":debit", p.UserID, "order:"+p.ID, "order-payment", -q.Total, now, allocations); err != nil {
			return nil, err
		}
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
	updateOrderPayment(c, q.ID, "paid", "", now)
	p.Version++
	for i := range w.Orders {
		if w.Orders[i].CheckoutID == p.ID {
			w.Orders[i].Payment = "paid"
			for _, ship := range q.Shipments {
				if ship.StoreID == w.Orders[i].StoreID {
					w.Orders[i].Preparation = ship.Preparation
					w.Orders[i].Delivery = ship.Delivery
					w.Orders[i].DeliveryFrom, w.Orders[i].DeliveryTo = ship.DeliveryFrom, ship.DeliveryTo
				}
			}
			w.Orders[i].Status = "awaiting-supply"
			w.Orders[i].Version++
			for j := range w.Orders[i].Items {
				w.Orders[i].Items[j].Status = "awaiting-supply"
			}
		}
	}
	refreshPurchaseOrderState(w, p)
	recordPurchaseEvent(p, a, "order.paid", now)
	commerceAudit(w, op, a, "purchase.confirm", p.ID, "", now)
	return purchaseView(c, *p), nil
}
func (s *Service) AdminCancelPurchase(ctx context.Context, a identity.User, in PurchaseAdminInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "purchase-admin-cancel:"+a.ID, key, in, func(w Workspace) error {
		if e := requireCommerceAdmin(w, a, CommerceOrdersConfirm); e != nil {
			return e
		}
		if p := findPurchase(commerceState(&w), in.ID); p != nil && p.ConfirmedAt != nil {
			return requireCommerceAdmin(w, a, CommerceRefundsResolve)
		}
		return nil
	}, func(w *Workspace, op string, now time.Time) (any, error) {
		c := commerceState(w)
		p := findPurchase(c, in.ID)
		if p == nil {
			return nil, ErrNotFound
		}
		if hasLiveOrderRequest(c, p.ID) {
			return nil, ErrConflict
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
				if order.CheckoutID == p.ID && !beforeShipping(order.Status) {
					return nil, ErrConflict
				}
			}
			releaseShipping(w, q.ID, true, now)
			if q.Total > 0 {
				if e := appendWallet(c, op+":credit", p.UserID, "order-cancel:"+p.ID, "order-cancel-credit", q.Total, now); e != nil {
					return nil, e
				}
				if err := returnPurchaseCredit(c, q, op+":credit", a, now); err != nil {
					return nil, err
				}
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
			releaseQuote(w, q, now)
			q.Status = "cancelled"
		}
		p.Status = "cancelled"
		if p.ConfirmedAt == nil {
			updateOrderPayment(c, p.QuoteID, "failed", "staff_cancelled", now)
		}
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
		refreshPurchaseOrderState(w, p)
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
			if err := ensureCreditState(c, r.UserID, now); err != nil {
				return nil, err
			}
			if r.RequestID != "" && in.Complete {
				amount, e := commerce.ParseToman(normalizeDigits(in.VerifiedAmountToman))
				if e != nil || amount != r.Amount {
					return nil, ErrInvalid
				}
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
				if e = appendWallet(c, op+":debit", r.UserID, "refund:"+r.ID, "manual-refund", -r.Amount, now, r.Allocations); e != nil {
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
			if r.PurchaseID != "" {
				if p := findPurchase(c, r.PurchaseID); p != nil {
					p.Version++
					recordPurchaseEvent(p, a, "refund.bank-"+r.Status, now)
					event := &p.Events[len(p.Events)-1]
					event.ID = op
					event.RequestID = r.RequestID
				}
			}
			commerceAudit(w, op, a, "refund.resolve", r.ID, "", now)
			return refundView(*r), nil
		}
		return nil, ErrNotFound
	})
}
