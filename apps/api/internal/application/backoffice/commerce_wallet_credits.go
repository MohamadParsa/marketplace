package backoffice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const CommerceWalletCredit = "commerce.wallet.credit"

type CreditLotView struct {
	ID           string    `json:"id"`
	Source       string    `json:"source"`
	OriginLotID  string    `json:"origin_lot_id,omitempty"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	AmountToman  string    `json:"amount_toman"`
	SpentToman   string    `json:"spent_toman"`
	HeldToman    string    `json:"held_toman"`
	FreeToman    string    `json:"free_toman"`
	Withdrawable bool      `json:"withdrawable"`
	At           time.Time `json:"at"`
}

func creditLotViews(c *commerce.State, user string) []CreditLotView {
	result := []CreditLotView{}
	held := creditHeld(c, user, "")
	for _, l := range c.CreditLots {
		if l.UserID != user {
			continue
		}
		title := l.Title
		if title == "" {
			title = "اعتبار نقدی"
			if l.Source == "legacy-unattributed" {
				title = "مانده قدیمی کیف پول"
			}
		}
		result = append(result, CreditLotView{ID: l.ID, Source: l.Source, OriginLotID: l.OriginLotID, Title: title, Description: l.Description, AmountToman: commerce.Toman(l.Amount), SpentToman: commerce.Toman(l.Spent), HeldToman: commerce.Toman(held[l.ID]), FreeToman: commerce.Toman(max(0, l.Amount-l.Spent-held[l.ID])), Withdrawable: l.CanWithdraw(), At: l.At})
	}
	return result
}
func creditSourceViews(c *commerce.State, q *commerce.Quote) []CreditSourceView {
	result := []CreditSourceView{}
	if q == nil || q.Credit == nil {
		return result
	}
	for _, part := range []struct {
		kind        string
		allocations []commerce.CreditAllocation
	}{{"selected-balance", q.Credit.Existing}, {"checkout-funding", q.Credit.Funding}} {
		for _, a := range part.allocations {
			for _, lot := range c.CreditLots {
				if lot.ID == a.LotID && lot.UserID == q.UserID {
					result = append(result, CreditSourceView{LotID: lot.ID, Source: lot.Source, AmountToman: commerce.Toman(a.Amount), Kind: part.kind, Status: q.Credit.Status, Withdrawable: lot.CanWithdraw(), Title: lot.Title, OriginLotID: lot.OriginLotID})
					break
				}
			}
		}
	}
	return result
}

type AdminCreditInput struct {
	UserID       string `json:"user_id"`
	AmountToman  string `json:"amount_toman"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Withdrawable bool   `json:"withdrawable"`
}
type AdminCreditResult struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	CustomerName      string    `json:"customer_name"`
	AmountToman       string    `json:"amount_toman"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Withdrawable      bool      `json:"withdrawable"`
	BalanceToman      string    `json:"balance_toman"`
	AvailableToman    string    `json:"available_toman"`
	WithdrawableToman string    `json:"withdrawable_toman"`
	At                time.Time `json:"at"`
}
type AdminCreditWallet struct {
	UserID            string             `json:"user_id"`
	CustomerName      string             `json:"customer_name"`
	BalanceToman      string             `json:"balance_toman"`
	AvailableToman    string             `json:"available_toman"`
	WithdrawableToman string             `json:"withdrawable_toman"`
	Credits           []CreditLotView    `json:"credits"`
	PreviousOperation *AdminCreditResult `json:"previous_operation,omitempty"`
}

func (s *Service) ReadCustomerCreditWallet(ctx context.Context, a identity.User, user, key string) (AdminCreditWallet, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return AdminCreditWallet{}, e
	}
	if e = requireCommerceAdmin(w, a, CommerceWalletCredit); e != nil {
		return AdminCreditWallet{}, e
	}
	target := findUser(&w, user)
	if target < 0 {
		return AdminCreditWallet{}, ErrNotFound
	}
	c := commerceState(&w)
	if e = ensureCreditState(c, user, s.clock.Now()); e != nil {
		return AdminCreditWallet{}, e
	}
	balance, held := walletAmounts(c, user)
	result := AdminCreditWallet{UserID: user, CustomerName: w.Users[target].Name, BalanceToman: commerce.Toman(balance), AvailableToman: commerce.Toman(balance - held), WithdrawableToman: commerce.Toman(withdrawableCredit(c, user)), Credits: creditLotViews(c, user)}
	if len(key) > 128 {
		return result, ErrInvalid
	}
	if key != "" {
		for _, op := range c.Operations {
			if op.Scope == "wallet-admin-credit:"+a.ID && op.Key == key {
				var v AdminCreditResult
				if json.Unmarshal(op.Result, &v) != nil || v.UserID != user {
					return AdminCreditWallet{}, ErrConflict
				}
				result.PreviousOperation = &v
				break
			}
		}
	}
	return result, nil
}
func (s *Service) CreditCustomerWallet(ctx context.Context, a identity.User, in AdminCreditInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "wallet-admin-credit:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, CommerceWalletCredit) }, func(w *Workspace, id string, now time.Time) (any, error) {
		index := findUser(w, in.UserID)
		if index < 0 || !w.Users[index].Active {
			return nil, ErrNotFound
		}
		amount, e := commerce.ParseToman(normalizeDigits(in.AmountToman))
		if e != nil || amount <= 0 {
			return nil, ErrInvalid
		}
		title, description := strings.TrimSpace(in.Title), strings.TrimSpace(in.Description)
		if title == "" || description == "" || !textBound(title, 120) || !textBound(description, 1000) {
			return nil, ErrInvalid
		}
		c := commerceState(w)
		if e = appendWallet(c, id+":credit", in.UserID, "admin-credit:"+id, "admin-credit", amount, now); e != nil {
			return nil, e
		}
		lot := &c.CreditLots[len(c.CreditLots)-1]
		lot.Withdrawable = &in.Withdrawable
		lot.Title = title
		lot.Description = description
		lot.ActorID = a.ID
		balance, held := walletAmounts(c, in.UserID)
		commerceAudit(w, id, a, "wallet.credit", in.UserID, description, now)
		return AdminCreditResult{ID: lot.ID, UserID: in.UserID, CustomerName: w.Users[index].Name, AmountToman: commerce.Toman(amount), Title: title, Description: description, Withdrawable: in.Withdrawable, BalanceToman: commerce.Toman(balance), AvailableToman: commerce.Toman(balance - held), WithdrawableToman: commerce.Toman(withdrawableCredit(c, in.UserID)), At: now}, nil
	})
}

// A cancellation is one positive ledger entry with one returned lot per original
// allocation, retaining its restriction and original FIFO age. Nothing is unspent.
func returnPurchaseCredit(c *commerce.State, q *commerce.Quote, entry string, a identity.User, now time.Time) error {
	if q.Credit == nil {
		return nil
	} // explicitly supported legacy orders, before source snapshots
	allocations := combinedCredit(q)
	if creditSum(allocations) != q.Total {
		return ErrConflict
	}
	if len(c.CreditLots) == 0 || c.CreditLots[len(c.CreditLots)-1].ID != entry {
		return ErrConflict
	}
	sources := map[string]commerce.CreditLot{}
	for _, l := range c.CreditLots {
		sources[l.ID] = l
	}
	c.CreditLots = c.CreditLots[:len(c.CreditLots)-1]
	for i, allocation := range allocations {
		original, ok := sources[allocation.LotID]
		if !ok || original.UserID != q.UserID || allocation.Amount <= 0 {
			return ErrConflict
		}
		withdrawable := original.CanWithdraw()
		priority := original.AvailableSince()
		c.CreditLots = append(c.CreditLots, commerce.CreditLot{ID: fmt.Sprintf("%s:%d", entry, i), EntryID: entry, UserID: q.UserID, Source: "order-cancel:" + q.PurchaseID, OriginLotID: original.ID, Amount: allocation.Amount, Withdrawable: &withdrawable, Title: "بازگشت اعتبار خرید", Description: original.Title, ActorID: a.ID, At: now, PriorityAt: &priority})
	}
	return nil
}

// Distinct safe errors retain the umbrella ErrCoupon for existing callers.
var ErrCouponUnknown = errors.New("coupon_invalid")
var ErrCouponExpired = errors.New("coupon_expired")
var ErrCouponNotStarted = errors.New("coupon_not_started")
var ErrCouponQuota = errors.New("coupon_quota_reached")
var ErrCouponUserQuota = errors.New("coupon_user_limit_reached")
var ErrCouponNoDiscount = errors.New("coupon_no_discount")
