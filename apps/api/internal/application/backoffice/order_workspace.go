package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type OrderWorkspaceQuery struct {
	Search, Store, Order, Payment, Fulfillment, Mobile string
	Page, Size                                         int
}
type OrderWorkspaceRow struct {
	ID                string   `json:"id"`
	Number            string   `json:"number"`
	Customer          string   `json:"customer"`
	Stores            []string `json:"stores"`
	TotalToman        string   `json:"total_toman"`
	OrderStatus       string   `json:"order_status"`
	PaymentStatus     string   `json:"payment_status"`
	FulfillmentStatus string   `json:"fulfillment_status"`
	Groups            int      `json:"groups"`
	Legacy            bool     `json:"legacy"`
}
type OrderStoreOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type OrderWorkspacePage struct {
	Items  []OrderWorkspaceRow `json:"items"`
	Stores []OrderStoreOption  `json:"stores"`
	Total  int                 `json:"total"`
	Page   int                 `json:"page"`
	Size   int                 `json:"size"`
}
type OrderWorkspaceDetail struct {
	Purchase    *AdminPurchaseView `json:"purchase,omitempty"`
	Permissions map[string]bool    `json:"permissions"`
	Legacy      bool               `json:"legacy"`
}

func orderWorkspaceActor(w Workspace, a identity.User) (identity.User, error) {
	if a.Surface != identity.SurfaceBackoffice {
		return a, ErrForbidden
	}
	current, e := currentActor(w, a)
	if e != nil {
		return a, e
	}
	if current.Surface != identity.SurfaceBackoffice || !allowsAction(current, CommerceOrdersRead) {
		return a, ErrForbidden
	}
	return current, nil
}
func (s *Service) QueryOrderWorkspace(ctx context.Context, a identity.User, q OrderWorkspaceQuery) (OrderWorkspacePage, error) {
	result := OrderWorkspacePage{Items: []OrderWorkspaceRow{}, Stores: []OrderStoreOption{}, Page: q.Page, Size: q.Size}
	if q.Mobile != "" {
		q.Mobile = identity.NormalizePhone(q.Mobile)
		if len(q.Mobile) != 11 || !strings.HasPrefix(q.Mobile, "09") || strings.IndexFunc(q.Mobile, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
			return result, ErrInvalid
		}
	}
	if q.Page < 0 || q.Page > 1000000 || q.Size < 1 || q.Size > 100 || !textBound(q.Search, 150) || !textBound(q.Store, 128) {
		return result, ErrInvalid
	}
	if q.Order != "" && !slices.Contains([]string{"placed", "processing", "completed", "cancelled"}, q.Order) || q.Payment != "" && !slices.Contains([]string{"unpaid", "paid", "returned-to-wallet", "authorized", "failed"}, q.Payment) || q.Fulfillment != "" && !slices.Contains([]string{"awaiting-payment", "awaiting-supply", "supplied", "packing", "shipped", "delivered", "cancelled", "mixed", "on-hold", "ready-to-ship", "awaiting-dispatch"}, q.Fulfillment) {
		return result, ErrInvalid
	}
	if e := s.ReleaseExpiredCommerce(ctx); e != nil {
		return result, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return result, e
	}
	if _, e = orderWorkspaceActor(w, a); e != nil {
		return result, e
	}
	for _, store := range w.Stores {
		result.Stores = append(result.Stores, OrderStoreOption{store.ID, store.Name})
	}
	storeName := func(id string) string {
		for _, store := range w.Stores {
			if store.ID == id {
				return store.Name
			}
		}
		return "فروشگاه سابق"
	}
	matches := func(row OrderWorkspaceRow, storeIDs []string, mobile string) bool {
		if q.Mobile != "" && q.Mobile != identity.NormalizePhone(mobile) {
			return false
		}
		if q.Store != "" && !slices.Contains(storeIDs, q.Store) {
			return false
		}
		if q.Order != "" && row.OrderStatus != q.Order || q.Payment != "" && row.PaymentStatus != q.Payment || q.Fulfillment != "" && row.FulfillmentStatus != q.Fulfillment {
			return false
		}
		needle := strings.ToLower(normalizeDigits(q.Search))
		return needle == "" || strings.Contains(strings.ToLower(row.Number+" "+row.Customer+" "+strings.Join(row.Stores, " ")), needle)
	}
	rows := []OrderWorkspaceRow{}
	c := commerceState(&w)
	for _, p := range c.Purchases {
		mobile := ""
		order, payment := purchaseStatuses(p)
		row := OrderWorkspaceRow{ID: p.ID, Number: strconv.FormatInt(p.Number, 10), OrderStatus: order, PaymentStatus: payment, FulfillmentStatus: p.FulfillmentState, Groups: len(p.OrderIDs), Stores: []string{}}
		for _, user := range w.Users {
			if user.ID == p.UserID {
				row.Customer = user.Name
				mobile = user.Phone
			}
		}
		ids := []string{}
		for _, id := range p.OrderIDs {
			o, _, err := liveOrder(&w, id)
			if err == nil && !slices.Contains(ids, o.StoreID) {
				ids = append(ids, o.StoreID)
				row.Stores = append(row.Stores, storeName(o.StoreID))
			}
		}
		if quote := findQuote(c, p.QuoteID); quote != nil {
			row.TotalToman = commerce.Toman(quote.Total)
		}
		if matches(row, ids, mobile) {
			rows = append(rows, row)
		}
	}
	for _, o := range w.Orders {
		if o.CheckoutID != "" && findPurchase(c, o.CheckoutID) != nil {
			continue
		}
		row := OrderWorkspaceRow{ID: o.ID, Number: strconv.FormatInt(recordNumber(&w, "order", o.ID), 10), Customer: "سابقه قدیمی", Stores: []string{storeName(o.StoreID)}, TotalToman: commerce.Toman(commerce.Amount(o.Total() + o.ShippingFee)), OrderStatus: orderLifecycle(o), PaymentStatus: o.Payment, FulfillmentStatus: o.Status, Groups: 1, Legacy: true}
		if matches(row, []string{o.StoreID}, "") {
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Legacy != rows[j].Legacy {
			return !rows[i].Legacy
		}
		a, _ := strconv.ParseInt(rows[i].Number, 10, 64)
		b, _ := strconv.ParseInt(rows[j].Number, 10, 64)
		return a > b
	})
	result.Total = len(rows)
	start := min(q.Page*q.Size, len(rows))
	end := min(start+q.Size, len(rows))
	result.Items = rows[start:end]
	return result, nil
}

func workspacePurchase(w *Workspace, a identity.User, p commerce.Purchase) AdminPurchaseView {
	c := w.Commerce
	v := purchaseView(c, p)
	if !allowsAction(a, CommerceDepositsRead) {
		v.Deposits = []DepositView{}
		v.PaymentAttempts = nil
		v.Quote.Destination = commerce.Destination{}
		v.Quote.PaymentMethod.Instructions = ""
	}
	if !allowsAction(a, CommerceOrdersConfirm) {
		v.WalletBalanceToman = ""
		v.WalletAvailableToman = ""
		v.ShortfallToman = ""
		v.CreditSources = nil
	}
	entry := AdminPurchaseView{PurchaseView: v, UserID: p.UserID}
	for _, u := range w.Users {
		if u.ID == p.UserID {
			entry.CustomerName = u.Name
			entry.CustomerPhone = u.Phone
		}
	}
	if q := findQuote(c, p.QuoteID); q != nil {
		if q.Discount != nil && allowsAction(a, CouponsRead) {
			var platform, seller commerce.Amount
			for _, allocation := range q.Discount.Allocations {
				platform += allocation.Platform
				seller += allocation.Seller
			}
			entry.DiscountFunding = &DiscountFundingView{commerce.Toman(platform), commerce.Toman(seller), q.Discount.Coupon.Version}
		}
		if q.Credit != nil && allowsAction(a, CommerceOrdersConfirm) {
			entry.CreditSources = creditSourceViews(c, q)
		}
	}
	return entry
}
func (s *Service) ReadOrderWorkspace(ctx context.Context, a identity.User, id string) (OrderWorkspaceDetail, error) {
	result := OrderWorkspaceDetail{Permissions: map[string]bool{}}
	if e := s.ReleaseExpiredCommerce(ctx); e != nil {
		return result, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return result, e
	}
	a, e = orderWorkspaceActor(w, a)
	if e != nil {
		return result, e
	}
	for _, action := range []string{CommerceOrdersConfirm, CommerceDepositsRead, CommerceRefundsResolve} {
		result.Permissions[action] = allowsAction(a, action)
	}
	p, e := requestPurchase(&w, id)
	if e != nil {
		for _, o := range w.Orders {
			if o.ID != id {
				continue
			}
			if o.CheckoutID == "" || findPurchase(w.Commerce, o.CheckoutID) == nil {
				result.Legacy = true
				return result, nil
			}
			p, e = requestPurchase(&w, o.CheckoutID)
			break
		}
	}
	if e != nil {
		return result, e
	}
	view := workspacePurchase(&w, a, *p)
	view.Events = []OrderEventView{}
	result.Purchase = &view
	return result, nil
}
