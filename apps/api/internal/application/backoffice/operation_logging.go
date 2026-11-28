package backoffice

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
)

func (s *Service) UseOperationLogs(r *operationlog.Recorder, scanner operationlog.Scanner, stats func() any) {
	s.operationLogs, s.operationScanner, s.operationStats = r, scanner, stats
}
func (s *Service) observeContext(ctx context.Context) context.Context {
	ctx = operationlog.Ensure(ctx)
	if operationlog.FromContext(ctx) == nil && s.operationLogs != nil {
		ctx = operationlog.WithRecorder(ctx, s.operationLogs)
	}
	return ctx
}
func operationOutcome(err error) (string, string) {
	if err == nil {
		return "succeeded", ""
	}
	var gatewayFields *GatewayFieldError
	var shippingField *ShippingFieldError
	if errors.As(err, &gatewayFields) || errors.As(err, &shippingField) {
		return "rejected", "invalid_fields"
	}
	for _, known := range []error{ErrCouponUnknown, ErrCouponExpired, ErrCouponNotStarted, ErrCouponQuota, ErrCouponUserQuota, ErrCouponNoDiscount, ErrCoupon, ErrInvalidMethod, ErrWalletExceedsTotal, ErrWithdrawalUnavailable, ErrTransferReference, ErrTransferDate, ErrRefundReason, ErrAddressUnavailable, ErrProfileIncomplete, ErrAddressIncomplete, ErrShippingFull, ErrShippingExpired} {
		if errors.Is(err, known) {
			return "rejected", known.Error()
		}
	}
	switch {
	case errors.Is(err, ErrForbidden):
		return "rejected", "forbidden"
	case errors.Is(err, ErrInvalid):
		return "rejected", "invalid_request"
	case errors.Is(err, ErrConflict):
		return "rejected", "conflict"
	case errors.Is(err, ErrNotFound):
		return "rejected", "not_found"
	case errors.Is(err, ErrStock):
		return "rejected", "insufficient_stock"
	case errors.Is(err, ErrInsufficientWallet):
		return "rejected", "insufficient_wallet"
	case errors.Is(err, ErrCheckoutChanged):
		return "rejected", "checkout_changed"
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return "unknown", "interrupted"
	default:
		return "failed", "operation_failed"
	}
}
func (s *Service) operationFinish(ctx context.Context, name string, started time.Time, err error, replayed bool, refs ...operationlog.Event) {
	outcome, code := operationOutcome(err)
	if err == nil && replayed {
		outcome = "replayed"
	}
	level := "info"
	if err != nil {
		level = "warn"
		if outcome == "failed" || outcome == "unknown" {
			level = "error"
		}
	}
	e := operationlog.Event{}
	if len(refs) > 0 {
		e = refs[0]
	}
	e.EventName, e.Type, e.Severity, e.Outcome, e.ErrorCode = name, "technical", level, outcome, code
	e.DurationMS = time.Since(started).Milliseconds()
	if e.Attributes == nil {
		e.Attributes = map[string]any{}
	}
	e.Attributes["replayed"] = replayed
	operationlog.Record(ctx, e)
}
func mutationReferences(name string, result json.RawMessage) operationlog.Event {
	var v struct {
		ID               string `json:"id"`
		PurchaseID       string `json:"purchase_id"`
		OrderID          string `json:"order_id"`
		QuoteID          string `json:"quote_id"`
		PaymentAttemptID string `json:"payment_attempt_id"`
	}
	_ = json.Unmarshal(result, &v)
	e := operationlog.Event{EntityID: v.ID, PurchaseID: v.PurchaseID, OrderID: v.OrderID, PaymentAttemptID: v.PaymentAttemptID, Attributes: map[string]any{"quote_id": v.QuoteID}}
	switch {
	case strings.Contains(name, "purchase"):
		e.EntityType = "purchase"
		e.PurchaseID = v.ID
	case strings.Contains(name, "quote"):
		e.EntityType = "quote"
		e.Attributes["quote_id"] = v.ID
	case strings.Contains(name, "cart"):
		e.EntityType = "cart"
		e.Attributes["cart_id"] = v.ID
	case strings.Contains(name, "deposit"):
		e.EntityType = "deposit"
	case strings.Contains(name, "order"):
		e.EntityType = "order"
		e.OrderID = v.ID
	}
	return e
}
