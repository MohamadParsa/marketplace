package backoffice

import (
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

type PaymentAttemptView struct {
	ID          string    `json:"id"`
	Purpose     string    `json:"purpose"`
	MethodName  string    `json:"method_name"`
	AmountToman string    `json:"amount_toman"`
	Status      string    `json:"status"`
	FailureCode string    `json:"failure_code,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func requestOrderPayment(c *commerce.State, p *commerce.Purchase, q *commerce.Quote, now time.Time) {
	id := hashToken("order-payment:" + q.ID)
	for _, v := range c.PaymentAttempts {
		if v.ID == id {
			return
		}
	}
	c.PaymentAttempts = append(c.PaymentAttempts, commerce.PaymentAttempt{ID: id, PurchaseID: p.ID, QuoteID: q.ID, Purpose: "order-payment", SourceID: q.ID, MethodName: "کیف پول", Amount: q.Total, Status: "requested", CreatedAt: now, UpdatedAt: now})
}

func updateOrderPayment(c *commerce.State, quote, status, reason string, now time.Time) {
	for i := range c.PaymentAttempts {
		v := &c.PaymentAttempts[i]
		if v.Purpose == "order-payment" && v.QuoteID == quote && v.Status != "paid" {
			v.Status = status
			v.FailureCode = reason
			v.UpdatedAt = now
		}
	}
}

func updateFundingAttempt(c *commerce.State, d commerce.Deposit, now time.Time) {
	status := "pending"
	if d.ReceiptID != "" {
		status = "awaiting-result"
	}
	if d.Status == "approved" {
		status = "paid"
	}
	if d.Status == "rejected" {
		status = "failed"
	}
	for i := range c.PaymentAttempts {
		v := &c.PaymentAttempts[i]
		if v.Purpose == "wallet-funding" && v.SourceID == d.ID {
			v.Status = status
			v.UpdatedAt = now
			return
		}
	}
	quote := ""
	if p := findPurchase(c, d.PurchaseID); p != nil {
		quote = p.QuoteID
	}
	c.PaymentAttempts = append(c.PaymentAttempts, commerce.PaymentAttempt{ID: hashToken("wallet-funding:" + d.ID), PurchaseID: d.PurchaseID, QuoteID: quote, Purpose: "wallet-funding", SourceID: d.ID, MethodName: d.Method.Name, Amount: d.Amount, Status: status, CreatedAt: d.CreatedAt, UpdatedAt: now})
}

func paymentAttemptViews(c *commerce.State, p commerce.Purchase) ([]PaymentAttemptView, string) {
	result := []PaymentAttemptView{}
	stage := "requested"
	for _, v := range c.PaymentAttempts {
		if v.PurchaseID != p.ID {
			continue
		}
		result = append(result, PaymentAttemptView{ID: v.ID, Purpose: v.Purpose, MethodName: v.MethodName, AmountToman: commerce.Toman(v.Amount), Status: v.Status, FailureCode: v.FailureCode, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt})
		if v.QuoteID == p.QuoteID {
			stage = v.Status
			// Funding success credits the wallet; only the separate order debit is paid.
			if v.Purpose == "wallet-funding" && v.Status == "paid" {
				stage = "pending"
			}
		}
	}
	if p.ConfirmedAt != nil {
		stage = "paid"
	} else if p.Status == "cancelled" || p.Status == "requires-reconfirmation" {
		stage = "failed"
	}
	return result, stage
}
