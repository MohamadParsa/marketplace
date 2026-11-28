package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

var ErrCheckoutChanged = errors.New("checkout_changed")
var ErrWithdrawalUnavailable = errors.New("withdrawal_credit_unavailable")
var ErrWalletExceedsTotal = errors.New("wallet_exceeds_total")
var ErrInsufficientWallet = errors.New("insufficient_wallet")
var ErrInvalidMethod = errors.New("payment_method_not_eligible")
var ErrStock = errors.New("insufficient_stock")

func commerceState(w *Workspace) *commerce.State {
	if w.Commerce == nil {
		w.Commerce = &commerce.State{Settings: commerce.DefaultSettings()}
	}
	w.Commerce.EnsurePurchaseNumbers()
	for index := range w.Commerce.Purchases {
		p := &w.Commerce.Purchases[index]
		for i := range w.Orders {
			if w.Orders[i].PurchaseID == p.ID {
				w.Orders[i].PurchaseNumber = strconv.FormatInt(p.Number, 10)
			}
		}
		refreshPurchaseOrderState(w, p)
	}
	return w.Commerce
}
func customerActor(w Workspace, a identity.User) error {
	if a.Surface != identity.SurfaceCustomer {
		return ErrForbidden
	}
	for _, u := range w.Users {
		if u.ID == a.ID && u.Active && u.AccessVersion == a.AccessVersion {
			return nil
		}
	}
	return ErrForbidden
}
func hashToken(token string) string {
	s := sha256.Sum256([]byte(token))
	return hex.EncodeToString(s[:])
}
func textBound(s string, max int) bool {
	return utf8.RuneCountInString(s) <= max && !strings.ContainsRune(s, '\x00')
}
func normalizeDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '۰' && r <= '۹' {
			return '0' + r - '۰'
		}
		if r >= '٠' && r <= '٩' {
			return '0' + r - '٠'
		}
		return r
	}, strings.TrimSpace(s))
}

func (s *Service) mutateCommerce(ctx context.Context, scope, key string, input any, authorize func(Workspace) error, apply func(*Workspace, string, time.Time) (any, error)) (json.RawMessage, error) {
	if key == "" || len(key) > 128 || scope == "" {
		return nil, ErrInvalid
	}
	b, e := json.Marshal(input)
	if e != nil {
		return nil, ErrInvalid
	}
	fingerprint := hashToken(string(b))
	id, e := s.identifiers.Token()
	if e != nil {
		return nil, e
	}
	var result json.RawMessage
	e = s.repository.Update(ctx, func(w *Workspace) error {
		if e := authorize(*w); e != nil {
			return e
		}
		c := commerceState(w)
		expireCommerce(w, s.clock.Now())
		for _, op := range c.Operations {
			if op.Scope == scope && op.Key == key {
				if op.Fingerprint != fingerprint {
					return ErrConflict
				}
				result = append(json.RawMessage{}, op.Result...)
				return nil
			}
		}
		out, e := apply(w, id, s.clock.Now())
		if e != nil {
			return e
		}
		result, e = json.Marshal(out)
		if e != nil {
			return e
		}
		if len(result) > 128*1024 {
			return ErrInvalid
		}
		c.Operations = append(c.Operations, commerce.Operation{Scope: scope, Key: key, Fingerprint: fingerprint, Result: result, At: s.clock.Now()})
		return nil
	})
	return result, e
}
func findQuote(c *commerce.State, id string) *commerce.Quote {
	for i := range c.Quotes {
		if c.Quotes[i].ID == id {
			return &c.Quotes[i]
		}
	}
	return nil
}
func findPurchase(c *commerce.State, id string) *commerce.Purchase {
	for i := range c.Purchases {
		if c.Purchases[i].ID == id {
			return &c.Purchases[i]
		}
	}
	return nil
}
func releaseQuote(w *Workspace, q *commerce.Quote, now time.Time) {
	if q.Status != "reserved" {
		return
	}
	releaseShipping(w, q.ID, false, now)
	if q.Credit != nil && q.Credit.Status == "held" {
		q.Credit.Status = "released"
	}
	for _, line := range q.Lines {
		for i := range w.Offers {
			if w.Offers[i].ID == line.OfferID {
				w.Offers[i].Reserved -= line.Quantity
				if w.Offers[i].Reserved < 0 {
					w.Offers[i].Reserved = 0
				}
				w.Offers[i].Version++
			}
		}
	}
}
func expireCommerce(w *Workspace, now time.Time) bool {
	if w.Commerce == nil {
		return false
	}
	changed := false
	for i := range w.Commerce.Quotes {
		q := &w.Commerce.Quotes[i]
		if q.Status == "reserved" && !now.Before(q.ReservedUntil) {
			releaseQuote(w, q, now)
			q.Status = "expired"
			if p := findPurchase(w.Commerce, q.PurchaseID); p != nil && p.Status != "confirmed" && p.Status != "cancelled" {
				p.Status = "requires-reconfirmation"
				updateOrderPayment(w.Commerce, q.ID, "failed", "reservation_expired", now)
				p.Version++
				recordPurchaseEvent(p, identity.User{}, "reservation.expired", now)
			}
			changed = true
		}
	}
	return changed
}
func (s *Service) ReleaseExpiredCommerce(ctx context.Context) error {
	if r, ok := s.repository.(interface {
		CommerceExpiryDue(context.Context, time.Time) (bool, error)
	}); ok {
		due, e := r.CommerceExpiryDue(ctx, s.clock.Now())
		if e != nil {
			return e
		}
		if !due {
			return nil
		}
	}

	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return e
	}
	if !expireCommerce(&w, s.clock.Now()) {
		return nil
	}
	return s.repository.Update(ctx, func(w *Workspace) error { expireCommerce(w, s.clock.Now()); return nil })
}
func walletAmounts(c *commerce.State, user string) (balance, held commerce.Amount) {
	for _, e := range c.Entries {
		if e.UserID == user {
			balance += e.Amount
		}
	}
	for _, r := range c.Refunds {
		if r.UserID == user && r.Status == "pending" {
			held += r.Amount
		}
	}
	for _, q := range c.Quotes {
		if q.UserID == user && q.Status == "reserved" && q.Credit != nil && q.Credit.Status == "held" {
			held += creditSum(q.Credit.Existing) + creditSum(q.Credit.Funding)
		}
	}
	return
}
func appendWallet(c *commerce.State, id, user, source, kind string, amount commerce.Amount, now time.Time, exact ...[]commerce.CreditAllocation) error {
	if amount == 0 || amount > commerce.MaxAmount || amount < -commerce.MaxAmount {
		return ErrInvalid
	}
	for _, e := range c.Entries {
		if e.Source == source {
			return ErrConflict
		}
	}
	if err := ensureCreditState(c, user, now); err != nil {
		return err
	}
	balance, held := walletAmounts(c, user)
	if amount > 0 && balance > commerce.MaxAmount-amount {
		return ErrInvalid
	}
	if amount < 0 && balance-held < -amount {
		return ErrInsufficientWallet
	}
	if amount < 0 {
		var allocations []commerce.CreditAllocation
		if len(exact) > 0 {
			allocations = exact[0]
		}
		if kind == "manual-refund" {
			if allocations == nil {
				var err error
				allocations, err = allocateCreditMode(c, user, -amount, "", true)
				if err != nil {
					return err
				}
			}
			for _, allocation := range allocations {
				found := false
				for _, lot := range c.CreditLots {
					if lot.ID == allocation.LotID && lot.UserID == user && lot.CanWithdraw() {
						found = true
						break
					}
				}
				if !found {
					return ErrWithdrawalUnavailable
				}
			}
		}
		if e := spendCredit(c, user, -amount, allocations); e != nil {
			return e
		}
	}
	c.Entries = append(c.Entries, commerce.WalletEntry{ID: id, UserID: user, Source: source, Kind: kind, Amount: amount, At: now})
	if amount > 0 && hasCreditAccount(c, user) {
		withdrawable := true
		c.CreditLots = append(c.CreditLots, commerce.CreditLot{ID: id, EntryID: id, UserID: user, Source: source, Amount: amount, At: now, Withdrawable: &withdrawable})
	}
	return nil
}
func commerceAudit(w *Workspace, id string, a identity.User, action, target, reason string, now time.Time) {
	w.Audit = append(w.Audit, Audit{ID: id, Actor: a.ID, Action: "commerce." + action, Target: target, Reason: reason, At: now})
}

type OrderEventView struct {
	CustomerDescription string    `json:"customer_description,omitempty"`
	Version             int       `json:"version,omitempty"`
	ID                  string    `json:"id"`
	OrderID             string    `json:"order_id,omitempty"`
	RequestID           string    `json:"request_id,omitempty"`
	From                string    `json:"from,omitempty"`
	To                  string    `json:"to,omitempty"`
	Action              string    `json:"action"`
	FlowStatus          string    `json:"flow_status"`
	OrderStatus         string    `json:"order_status"`
	PaymentStatus       string    `json:"payment_status"`
	ActorKind           string    `json:"actor_kind"`
	At                  time.Time `json:"at"`
}

func purchaseStatuses(p commerce.Purchase) (string, string) {
	order := p.OrderState
	if order == "" {
		order = "placed"
	}
	if p.Status == "cancelled" {
		if p.ConfirmedAt != nil {
			return "cancelled", "returned-to-wallet"
		}
		return "cancelled", "unpaid"
	}
	if p.ConfirmedAt != nil {
		return order, "paid"
	}
	return order, "unpaid"
}
func recordPurchaseEvent(p *commerce.Purchase, a identity.User, action string, now time.Time) {
	order, payment := purchaseStatuses(*p)
	kind := "system"
	if a.ID != "" {
		kind = "staff"
		if a.Surface == identity.SurfaceCustomer {
			kind = "customer"
		}
		if a.Surface == identity.SurfaceSeller {
			kind = "seller"
		}
	}
	id := hashToken(p.ID + ":" + strconv.Itoa(len(p.Events)) + ":" + action + ":" + now.Format(time.RFC3339Nano))
	p.Events = append(p.Events, commerce.OrderEvent{CustomerDescription: orderCustomerDescription(action), Version: p.Version, ID: id, Action: action, FlowStatus: p.Status, OrderStatus: order, PaymentStatus: payment, ActorID: a.ID, ActorKind: kind, At: now})
}

func orderCustomerDescription(action string) string {
	switch action {
	case "order.accepted":
		return "سفارش شما ثبت شد؛ وضعیت پرداخت و هر مرسوله جداگانه قابل پیگیری است."
	case "order.paid":
		return "پرداخت سفارش قطعی شد؛ فروشگاه می‌تواند پردازش مرسوله را آغاز کند."
	case "fulfillment.start":
		return "فروشگاه پردازش این مرسوله را آغاز کرد."
	case "fulfillment.supply":
		return "تأمین تمام اقلام باقی‌مانده این مرسوله تأیید شد."
	case "fulfillment.pack":
		return "اقلام این مرسوله در حال آماده‌سازی و بسته‌بندی هستند."
	case "fulfillment.ship":
		return "مرسوله به حمل‌کننده تحویل شد؛ ارسال به معنی تأیید دریافت شما نیست."
	case "fulfillment.deliver":
		return "دریافت تمام اقلام این مرسوله تأیید شد."
	case "fulfillment.correct", "fulfillment.configure-dispatch":
		return "مشخصات ارسال این مرسوله به‌روزرسانی شد."
	case "order.cancelled":
		return "سفارش لغو شد؛ نتیجه بازگشت وجه را در بخش مالی پیگیری کنید."
	case "order.reconfirmation-required":
		return "برای ادامه خرید، شرایط تازه سفارش باید دوباره بررسی و تأیید شوند."
	case "request.created":
		return "درخواست شما ثبت شد و برای رسیدگی در انتظار بررسی است."
	case "request.approve":
		return "درخواست تأیید شد؛ نتیجه مالی و دریافت کالای برگشتی جداگانه پیگیری می‌شوند."
	case "request.reject":
		return "بررسی درخواست به پایان رسید؛ نتیجه آن در بخش درخواست‌ها قابل مشاهده است."
	case "request.receive":
		return "دریافت کالای برگشتی و نتیجه بررسی آن ثبت شد."
	case "refund.wallet":
		return "مبلغ تأییدشده به کیف پول بازگردانده شد؛ محدودیت منابع اولیه اعتبار حفظ می‌شود."
	case "refund.bank":
		return "بازگشت وجه برای انتقال دستی در مسیر رسیدگی مالی ثبت شد."
	default:
		return "تغییری در روند رسیدگی سفارش ثبت شد؛ وضعیت فعلی در جزئیات قابل مشاهده است."
	}
}
