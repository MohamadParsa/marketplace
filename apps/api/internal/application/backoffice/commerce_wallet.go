package backoffice

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
	"time"
)

type DepositInput struct {
	MethodID      string `json:"method_id"`
	PurchaseID    string `json:"purchase_id,omitempty"`
	DestinationID string `json:"destination_id,omitempty"`
	AmountToman   string `json:"amount_toman"`
	Date          string `json:"date"`
	Reference     string `json:"reference"`
}
type DepositView struct {
	ID           string               `json:"id"`
	PurchaseID   string               `json:"purchase_id,omitempty"`
	MethodID     string               `json:"method_id"`
	MethodName   string               `json:"method_name"`
	AmountToman  string               `json:"amount_toman"`
	Date         string               `json:"date"`
	Reference    string               `json:"reference"`
	Status       string               `json:"status"`
	Version      int                  `json:"version"`
	ReceiptID    string               `json:"receipt_id,omitempty"`
	Destination  commerce.Destination `json:"destination"`
	Instructions string               `json:"instructions"`
	Reason       string               `json:"reason,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
}
type WalletEntryView struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Kind        string    `json:"kind"`
	AmountToman string    `json:"amount_toman"`
	At          time.Time `json:"at"`
}
type RefundView struct {
	PurchaseID   string    `json:"purchase_id,omitempty"`
	RequestID    string    `json:"request_id,omitempty"`
	ID           string    `json:"id"`
	AmountToman  string    `json:"amount_toman"`
	Instructions string    `json:"instructions"`
	Status       string    `json:"status"`
	Version      int       `json:"version"`
	Reference    string    `json:"reference,omitempty"`
	Date         string    `json:"date,omitempty"`
	Reason       string    `json:"reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
type WalletView struct {
	WithdrawableToman string               `json:"withdrawable_toman"`
	PurchaseOnlyToman string               `json:"purchase_only_toman"`
	Credits           []CreditLotView      `json:"credits"`
	BalanceToman      string               `json:"balance_toman"`
	HeldToman         string               `json:"held_toman"`
	AvailableToman    string               `json:"available_toman"`
	Entries           []WalletEntryView    `json:"entries"`
	Deposits          []DepositView        `json:"deposits"`
	Refunds           []RefundView         `json:"refunds"`
	Methods           []PaymentMethodInput `json:"funding_methods"`
	Destination       commerce.Destination `json:"destination"`
	Currency          string               `json:"currency"`
}
type RefundInput struct {
	AmountToman  string `json:"amount_toman"`
	Instructions string `json:"instructions"`
}

func depositView(d commerce.Deposit) DepositView {
	return DepositView{ID: d.ID, PurchaseID: d.PurchaseID, MethodID: d.Method.ID, MethodName: d.Method.Name, AmountToman: commerce.Toman(d.Amount), Date: d.Date, Reference: d.Reference, Status: d.Status, Version: d.Version, ReceiptID: d.ReceiptID, Destination: d.Destination, Instructions: d.Method.Instructions, Reason: d.Reason, CreatedAt: d.CreatedAt}
}
func refundView(r commerce.RefundRequest) RefundView {
	return RefundView{PurchaseID: r.PurchaseID, RequestID: r.RequestID, ID: r.ID, AmountToman: commerce.Toman(r.Amount), Instructions: r.Instructions, Status: r.Status, Version: r.Version, Reference: r.Reference, Date: r.Date, Reason: r.Reason, CreatedAt: r.CreatedAt}
}
func walletView(c *commerce.State, user string) WalletView {
	balance, held := walletAmounts(c, user)
	v := WalletView{BalanceToman: commerce.Toman(balance), HeldToman: commerce.Toman(held), AvailableToman: commerce.Toman(balance - held), Entries: []WalletEntryView{}, Deposits: []DepositView{}, Refunds: []RefundView{}, Methods: []PaymentMethodInput{}, Destination: c.Settings.Destination, Currency: "toman"}
	for _, e := range c.Entries {
		if e.UserID == user {
			v.Entries = append(v.Entries, WalletEntryView{e.ID, e.Source, e.Kind, commerce.Toman(e.Amount), e.At})
		}
	}
	for _, d := range c.Deposits {
		if d.UserID == user {
			v.Deposits = append(v.Deposits, depositView(d))
		}
	}
	for _, r := range c.Refunds {
		if r.UserID == user {
			v.Refunds = append(v.Refunds, refundView(r))
		}
	}
	for _, m := range c.Settings.Methods {
		if m.Active && m.Kind != "wallet" {
			v.Methods = append(v.Methods, methodInput(m))
		}
	}
	v.WithdrawableToman = commerce.Toman(withdrawableCredit(c, user))
	v.PurchaseOnlyToman = commerce.Toman(max(0, balance-held-withdrawableCredit(c, user)))
	v.Credits = creditLotViews(c, user)
	return v
}
func (s *Service) ReadWallet(ctx context.Context, a identity.User) (WalletView, error) {
	if e := s.ReleaseExpiredCommerce(ctx); e != nil {
		return WalletView{}, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return WalletView{}, e
	}
	if e = customerActor(w, a); e != nil {
		return WalletView{}, e
	}
	c := commerceState(&w)
	if e = ensureCreditState(c, a.ID, s.clock.Now()); e != nil {
		return WalletView{}, e
	}
	return walletView(c, a.ID), nil
}
func normalizeReference(s string) (string, error) {
	s = strings.ToUpper(normalizeDigits(s))
	if len(s) < 3 || len(s) > 80 {
		return "", ErrInvalid
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r == '-') {
			return "", ErrInvalid
		}
	}
	return s, nil
}
func validDepositDate(date string, now time.Time) bool {
	d, e := time.Parse("2006-01-02", date)
	if e != nil || d.Format("2006-01-02") != date {
		return false
	}
	loc, e := time.LoadLocation("Asia/Tehran")
	if e != nil {
		loc = time.FixedZone("Tehran", 12600)
	}
	return date <= now.In(loc).Format("2006-01-02")
}
func referenceKey(m commerce.PaymentMethod, d commerce.Destination, date, reference string) string {
	account := m.Kind
	if m.Kind == "card-transfer" {
		account += "|" + d.CardNumber
	}
	return hashToken(account + "|" + date + "|" + reference)
}
func (s *Service) CreateDeposit(ctx context.Context, a identity.User, in DepositInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "deposit:"+a.ID, key, in, func(w Workspace) error { return customerActor(w, a) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		amount, e := commerce.ParseToman(normalizeDigits(in.AmountToman))
		if e != nil || amount <= 0 {
			return nil, ErrInvalid
		}
		ref, e := normalizeReference(in.Reference)
		if e != nil || !validDepositDate(in.Date, now) {
			return nil, ErrInvalid
		}
		var method commerce.PaymentMethod
		found := false
		for _, m := range c.Settings.Methods {
			if m.ID == in.MethodID && m.Kind != "wallet" {
				method = m
				found = true
			}
		}
		if !found {
			return nil, ErrInvalidMethod
		}
		destination := c.Settings.Destination
		if in.PurchaseID != "" {
			p := findPurchase(c, in.PurchaseID)
			if p == nil || p.UserID != a.ID {
				return nil, ErrNotFound
			}
			if p.Status == "confirmed" || p.Status == "cancelled" {
				return nil, ErrConflict
			}
			q := findQuote(c, p.QuoteID)
			if q == nil {
				return nil, ErrConflict
			}
			if q.PaymentMethod.Kind != "wallet" && q.PaymentMethod.ID != method.ID {
				return nil, ErrInvalidMethod
			}
			if q.PaymentMethod.ID == method.ID {
				method = q.PaymentMethod
				if method.Kind == "card-transfer" {
					destination = q.Destination
				}
			}
			if !methodEligible(*w, method, q.Lines, q.Total) {
				return nil, ErrInvalidMethod
			}
		} else if !method.Active {
			return nil, ErrInvalidMethod
		}
		if amount < method.Min || method.Max > 0 && amount > method.Max {
			return nil, ErrInvalidMethod
		}
		if method.Kind == "card-transfer" {
			if in.DestinationID != "" && in.DestinationID != destination.ID {
				found := false
				for _, d := range c.Settings.DestinationHistory {
					if d.ID == in.DestinationID {
						destination = d
						found = true
						break
					}
				}
				if !found {
					return nil, ErrInvalid
				}
			}
			if destination.ID == "" || !validCard(destination.CardNumber) {
				return nil, ErrInvalid
			}
		}
		rk := referenceKey(method, destination, in.Date, ref)
		for _, d := range c.Deposits {
			if d.ReferenceKey == rk {
				return nil, ErrConflict
			}
		}
		d := commerce.Deposit{ID: id, UserID: a.ID, PurchaseID: in.PurchaseID, Method: method, Destination: destination, Amount: amount, Date: in.Date, Reference: ref, ReferenceKey: rk, Status: "pending", Version: 1, CreatedAt: now}
		c.Deposits = append(c.Deposits, d)
		updateFundingAttempt(c, d, now)
		commerceAudit(w, id+":audit", a, "deposit.declare", id, "", now)
		return depositView(d), nil
	})
}
func (s *Service) AttachReceipt(ctx context.Context, a identity.User, depositID string, version int, key string, input []byte) (json.RawMessage, error) {
	if len(input) > 2<<20 {
		return nil, ErrIconTooLarge
	}
	data, e := normalizeRaster(input, 2<<20, 4096, 16*1024*1024)
	if e != nil {
		return nil, e
	}
	if len(data) > 4<<20 {
		return nil, ErrIconTooLarge
	}
	finger := struct {
		ID      string
		Version int
		Digest  string
	}{depositID, version, hashToken(string(input))}
	return s.mutateCommerce(ctx, "receipt:"+a.ID, key, finger, func(w Workspace) error { return customerActor(w, a) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		for i := range c.Deposits {
			d := &c.Deposits[i]
			if d.ID != depositID || d.UserID != a.ID {
				continue
			}
			if d.Version != version || d.Status != "pending" {
				return nil, ErrConflict
			}
			d.ReceiptID = id
			updateFundingAttempt(c, *d, now)
			d.Version++
			kept := []commerce.Receipt{}
			for _, r := range c.Receipts {
				if r.DepositID != d.ID {
					kept = append(kept, r)
				}
			}
			c.Receipts = append(kept, commerce.Receipt{ID: id, UserID: a.ID, DepositID: d.ID, Data: data})
			commerceAudit(w, id+":audit", a, "receipt.attach", d.ID, "", now)
			return depositView(*d), nil
		}
		return nil, ErrNotFound
	})
}
func (s *Service) ReadReceipt(ctx context.Context, a identity.User, id string, admin bool) ([]byte, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return nil, e
	}
	if admin {
		if e = requireCommerceAdmin(w, a, CommerceDepositsRead); e != nil {
			return nil, e
		}
	} else if e = customerActor(w, a); e != nil {
		return nil, e
	}
	if w.Commerce != nil {
		for _, r := range w.Commerce.Receipts {
			if r.ID == id && (admin || r.UserID == a.ID) {
				if len(r.Data) > 0 {
					return append([]byte{}, r.Data...), nil
				}
				if reader, ok := s.repository.(interface {
					ReadReceiptData(context.Context, string, string, string) ([]byte, error)
				}); ok {
					return reader.ReadReceiptData(ctx, id, r.UserID, r.DepositID)
				}
				return nil, ErrNotFound
			}
		}
	}
	return nil, ErrNotFound
}
func (s *Service) RequestRefund(ctx context.Context, a identity.User, in RefundInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "refund:"+a.ID, key, in, func(w Workspace) error { return customerActor(w, a) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		amount, e := commerce.ParseToman(normalizeDigits(in.AmountToman))
		if e != nil || amount <= 0 || strings.TrimSpace(in.Instructions) == "" || !textBound(in.Instructions, 1000) {
			return nil, ErrInvalid
		}
		if e = ensureCreditState(c, a.ID, now); e != nil {
			return nil, e
		}
		allocations, e := allocateCreditMode(c, a.ID, amount, "", true)
		if e != nil {
			return nil, e
		}
		r := commerce.RefundRequest{Allocations: allocations, ID: id, UserID: a.ID, Amount: amount, Instructions: strings.TrimSpace(in.Instructions), Status: "pending", Version: 1, CreatedAt: now}
		c.Refunds = append(c.Refunds, r)
		commerceAudit(w, id+":audit", a, "refund.request", id, "", now)
		return refundView(r), nil
	})
}
