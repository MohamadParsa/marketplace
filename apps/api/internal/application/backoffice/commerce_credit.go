package backoffice

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

func hasCreditAccount(c *commerce.State, user string) bool {
	for _, a := range c.CreditAccounts {
		if a.UserID == user {
			return true
		}
	}
	return false
}

// Existing history stays immutable. Only its present balance becomes an explicitly
// unattributed opening lot; this is metadata, never a second wallet credit.
func ensureCreditAccount(c *commerce.State, user string, now time.Time) {
	if hasCreditAccount(c, user) {
		return
	}
	entries := []commerce.WalletEntry{}
	var amount commerce.Amount
	for _, e := range c.Entries {
		if e.UserID == user {
			entries = append(entries, e)
			amount += e.Amount
		}
	}
	b, _ := json.Marshal(entries)
	c.CreditAccounts = append(c.CreditAccounts, commerce.CreditAccount{UserID: user, Anchor: hashToken(string(b)), At: now})
	if amount > 0 {
		c.CreditLots = append(c.CreditLots, commerce.CreditLot{ID: "opening:" + user, UserID: user, Source: "legacy-unattributed", Amount: amount, At: now})
	}
}
func creditSum(v []commerce.CreditAllocation) (sum commerce.Amount) {
	for _, a := range v {
		sum += a.Amount
	}
	return
}
func creditHeld(c *commerce.State, user, except string) map[string]commerce.Amount {
	held := map[string]commerce.Amount{}
	for _, q := range c.Quotes {
		if q.UserID != user || q.ID == except || q.Status != "reserved" || q.Credit == nil || q.Credit.Status != "held" {
			continue
		}
		for _, a := range append(append([]commerce.CreditAllocation{}, q.Credit.Existing...), q.Credit.Funding...) {
			held[a.LotID] += a.Amount
		}
	}
	for _, r := range c.Refunds {
		if r.UserID == user && r.ID != except && r.Status == "pending" {
			for _, a := range r.Allocations {
				held[a.LotID] += a.Amount
			}
		}
	}
	return held
}
func allocateCredit(c *commerce.State, user string, amount commerce.Amount, except string) ([]commerce.CreditAllocation, error) {
	return allocateCreditMode(c, user, amount, except, false)
}
func allocateCreditMode(c *commerce.State, user string, amount commerce.Amount, except string, withdrawing bool) ([]commerce.CreditAllocation, error) {
	allocations := []commerce.CreditAllocation{}
	held := creditHeld(c, user, except)
	lots := append([]commerce.CreditLot{}, c.CreditLots...)
	sort.SliceStable(lots, func(i, j int) bool {
		if lots[i].CanWithdraw() != lots[j].CanWithdraw() {
			return !lots[i].CanWithdraw()
		}
		return lots[i].AvailableSince().Before(lots[j].AvailableSince())
	})
	for _, lot := range lots {
		if lot.UserID != user || (withdrawing && !lot.CanWithdraw()) {
			continue
		}
		free := lot.Amount - lot.Spent - held[lot.ID]
		take := min(free, amount)
		if take > 0 {
			allocations = append(allocations, commerce.CreditAllocation{LotID: lot.ID, Amount: take})
			amount -= take
		}
		if amount == 0 {
			break
		}
	}
	if amount != 0 {
		if withdrawing {
			return nil, ErrWithdrawalUnavailable
		}
		return nil, ErrInsufficientWallet
	}
	return allocations, nil
}

// Materialize legacy pending withdrawal holds before any new restricted credit.
// This only adds source metadata; the existing ledger and balances are untouched.
func ensureCreditState(c *commerce.State, user string, now time.Time) error {
	ensureCreditAccount(c, user, now)
	indices := []int{}
	for i, r := range c.Refunds {
		if r.UserID == user && r.Status == "pending" && r.Allocations == nil {
			indices = append(indices, i)
		}
	}
	sort.SliceStable(indices, func(i, j int) bool { return c.Refunds[indices[i]].CreatedAt.Before(c.Refunds[indices[j]].CreatedAt) })
	for _, i := range indices {
		r := &c.Refunds[i]
		a, e := allocateCreditMode(c, user, r.Amount, r.ID, true)
		if e != nil {
			return ErrConflict
		}
		r.Allocations = a
	}
	return nil
}
func withdrawableCredit(c *commerce.State, user string) commerce.Amount {
	balance, heldAmount := walletAmounts(c, user)
	if !hasCreditAccount(c, user) {
		return max(0, balance-heldAmount)
	}
	held := creditHeld(c, user, "")
	var total commerce.Amount
	for _, lot := range c.CreditLots {
		if lot.UserID == user && lot.CanWithdraw() {
			total += max(0, lot.Amount-lot.Spent-held[lot.ID])
		}
	}
	return min(max(0, balance-heldAmount), total)
}
func spendCredit(c *commerce.State, user string, amount commerce.Amount, allocations []commerce.CreditAllocation) error {
	if !hasCreditAccount(c, user) {
		return nil
	}
	if allocations == nil {
		var e error
		allocations, e = allocateCredit(c, user, amount, "")
		if e != nil {
			return e
		}
	}
	if creditSum(allocations) != amount {
		return ErrConflict
	}
	held := creditHeld(c, user, "")
	seen := map[string]bool{}
	for _, a := range allocations {
		if a.Amount <= 0 || seen[a.LotID] {
			return ErrConflict
		}
		seen[a.LotID] = true
		found := false
		for i := range c.CreditLots {
			lot := &c.CreditLots[i]
			if lot.ID == a.LotID && lot.UserID == user {
				if lot.Amount-lot.Spent-held[lot.ID] < a.Amount {
					return ErrInsufficientWallet
				}
				lot.Spent += a.Amount
				found = true
				break
			}
		}
		if !found {
			return ErrConflict
		}
	}
	return nil
}
func reserveCredit(c *commerce.State, q *commerce.Quote, now time.Time) error {
	if q.ContractVersion != 2 || q.Credit == nil {
		return nil
	}
	if e := ensureCreditState(c, q.UserID, now); e != nil {
		return e
	}
	balance, held := walletAmounts(c, q.UserID)
	if q.Credit.Requested > balance-held {
		return ErrInsufficientWallet
	}
	a, e := allocateCredit(c, q.UserID, q.Credit.Requested, q.ID)
	if e != nil {
		return e
	}
	q.Credit.Existing = a
	q.Credit.Status = "held"
	return nil
}
func earmarkDeposit(c *commerce.State, d commerce.Deposit, lotID string) {
	p := findPurchase(c, d.PurchaseID)
	if p == nil || p.Status == "confirmed" || p.Status == "cancelled" {
		return
	}
	q := findQuote(c, p.QuoteID)
	if q == nil || q.Status != "reserved" || q.ContractVersion != 2 || q.Credit == nil || q.Credit.Status != "held" {
		return
	}
	needed := q.Total - q.Credit.Requested - creditSum(q.Credit.Funding)
	if needed > 0 {
		q.Credit.Funding = append(q.Credit.Funding, commerce.CreditAllocation{LotID: lotID, Amount: min(needed, d.Amount)})
	}
}
func combinedCredit(q *commerce.Quote) []commerce.CreditAllocation {
	result := append([]commerce.CreditAllocation{}, q.Credit.Existing...)
	for _, a := range q.Credit.Funding {
		found := false
		for i := range result {
			if result[i].LotID == a.LotID {
				result[i].Amount += a.Amount
				found = true
			}
		}
		if !found {
			result = append(result, a)
		}
	}
	return result
}

func walletEligible(w Workspace, lines []commerce.LineSnapshot, total commerce.Amount) bool {
	for _, m := range commerceState(&w).Settings.Methods {
		if m.Kind == "wallet" && methodEligible(w, m, lines, total) {
			return true
		}
	}
	return false
}
