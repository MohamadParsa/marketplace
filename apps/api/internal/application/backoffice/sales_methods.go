package backoffice

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
	"time"
)

const SalesRead = "commerce.sales.read"
const SalesWrite = "commerce.sales.write"

func salesSettings(w Workspace) *commerce.SalesSettings {
	if w.Sales != nil {
		return w.Sales
	}
	if w.Commerce != nil && w.Commerce.Sales != nil {
		return w.Commerce.Sales
	}
	methods := commerce.DefaultSettings().Methods
	if w.Commerce != nil {
		methods = w.Commerce.Settings.Methods
	}
	ids := []string{}
	for _, m := range methods {
		if m.Kind != "wallet" {
			ids = append(ids, m.ID)
		}
	}
	return &commerce.SalesSettings{Version: 1, Methods: []commerce.SalesMethod{{ID: "cash", Name: "نقدی", Active: true, PaymentMethodIDs: ids}}, Overrides: []commerce.CategoryMarkup{}}
}

type SalesDirectory struct {
	commerce.SalesSettings
	PaymentMethods []PaymentMethodInput `json:"payment_methods"`
	CanWrite       bool                 `json:"can_write"`
}

func (s *Service) ReadSales(ctx context.Context, a identity.User) (SalesDirectory, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return SalesDirectory{}, e
	}
	if e = requireCommerceAdmin(w, a, SalesRead); e != nil {
		return SalesDirectory{}, e
	}
	actor, _ := currentActor(w, a)
	out := SalesDirectory{SalesSettings: *salesSettings(w), CanWrite: allowsAction(actor, SalesWrite)}
	for _, m := range commerceState(&w).Settings.Methods {
		if m.Kind != "wallet" {
			out.PaymentMethods = append(out.PaymentMethods, methodInput(m))
		}
	}
	return out, nil
}
func validateSales(w Workspace, in commerce.SalesSettings) error {
	if len(in.Methods) == 0 || len(in.Methods) > 100 || len(in.Overrides) > 20000 {
		return ErrInvalid
	}
	ids, paid := map[string]bool{}, map[string]bool{}
	settings := commerceState(&w).Settings
	for _, m := range in.Methods {
		if m.ID == "" || len(m.ID) > 64 || strings.TrimSpace(m.ID) != m.ID || ids[m.ID] || strings.TrimSpace(m.Name) == "" || !textBound(m.Name, 100) || m.MarkupBPS < 0 || m.MarkupBPS > 10000 || len(m.PaymentMethodIDs) > 100 {
			return ErrInvalid
		}
		ids[m.ID] = true
		for _, p := range m.PaymentMethodIDs {
			if paid[p] {
				return ErrInvalid
			}
			valid := false
			for _, pm := range settings.Methods {
				valid = valid || (pm.ID == p && pm.Kind != "wallet")
			}
			if !valid {
				return ErrInvalid
			}
			paid[p] = true
		}
	}
	seen := map[string]bool{}
	for _, o := range in.Overrides {
		key := o.CategoryID + "\x00" + o.MethodID
		if seen[key] || !ids[o.MethodID] || !categoryIDsValid(w, []string{o.CategoryID}) || o.MarkupBPS < 0 || o.MarkupBPS > 10000 {
			return ErrInvalid
		}
		seen[key] = true
	}
	return nil
}
func (s *Service) SaveSales(ctx context.Context, a identity.User, in commerce.SalesSettings, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "sales", key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, SalesWrite) }, func(w *Workspace, id string, now time.Time) (any, error) {
		current := salesSettings(*w)
		if current.Version != in.Version {
			return nil, ErrConflict
		}
		if e := validateSales(*w, in); e != nil {
			return nil, e
		}
		in.Version++
		stored := in
		stored.Methods = append([]commerce.SalesMethod{}, in.Methods...)
		for i := range stored.Methods {
			stored.Methods[i].PaymentMethodIDs = append([]string{}, stored.Methods[i].PaymentMethodIDs...)
		}
		stored.Overrides = append([]commerce.CategoryMarkup{}, in.Overrides...)
		commerceState(w).Sales = &stored
		commerceAudit(w, id, a, "sales.save", "sales", "", now)
		return in, nil
	})
}
func salesMethodForPayment(w Workspace, payment string) (commerce.SalesMethod, error) {
	for _, m := range salesSettings(w).Methods {
		for _, id := range m.PaymentMethodIDs {
			if id == payment {
				if !m.Active {
					return m, ErrInvalidMethod
				}
				return m, nil
			}
		}
	}
	return commerce.SalesMethod{}, ErrInvalidMethod
}
func resolvedMarkup(w Workspace, m commerce.SalesMethod, category string) commerce.MarkupSnapshot {
	result := commerce.MarkupSnapshot{MethodID: m.ID, MethodName: m.Name, BPS: m.MarkupBPS}
	seen := map[string]bool{}
	for category != "" && !seen[category] {
		seen[category] = true
		for _, o := range salesSettings(w).Overrides {
			if o.MethodID == m.ID && o.CategoryID == category {
				result.BPS, result.CategoryID = o.MarkupBPS, category
				return result
			}
		}
		parent := ""
		for _, c := range w.Categories {
			if c.ID == category {
				parent = c.ParentID
				break
			}
		}
		category = parent
	}
	return result
}
func applySales(w Workspace, q *commerce.Quote, payment string) error {
	q.CashSubtotal = q.Subtotal
	// A wallet-only or no-funding purchase uses cash prices without selecting a funding method.
	if payment == "" {
		return nil
	}
	for _, p := range commerceState(&w).Settings.Methods {
		if p.ID == payment && p.Kind == "wallet" {
			return nil
		}
	}
	m, e := salesMethodForPayment(w, payment)
	if e != nil {
		return e
	}
	q.SalesMethod = &m
	for i := range q.Lines {
		l := &q.Lines[i]
		category := ""
		for _, p := range w.Products {
			if p.ID == l.ProductID {
				category = p.CategoryID
				break
			}
		}
		snapshot := resolvedMarkup(w, m, category)
		snapshot.UnitAmount, e = commerce.Percent(l.UnitPrice, snapshot.BPS)
		if e != nil {
			return ErrInvalid
		}
		cash := l.UnitPrice
		l.CashUnitPrice = &cash
		l.Markup = &snapshot
		l.UnitPrice, e = commerce.Add(cash, snapshot.UnitAmount)
		if e != nil {
			return ErrInvalid
		}
		amount, e := commerce.Multiply(snapshot.UnitAmount, l.Quantity)
		if e != nil {
			return ErrInvalid
		}
		q.MarkupTotal, e = commerce.Add(q.MarkupTotal, amount)
		if e != nil {
			return ErrInvalid
		}
	}
	q.Subtotal, e = commerce.Add(q.CashSubtotal, q.MarkupTotal)
	if e != nil {
		return ErrInvalid
	}
	return nil
}

// Public preview resolves the same method/rate as checkout; only a real visible offer is accepted.
func (s *Service) SalesPrices(ctx context.Context, product, variant, offer string) ([]map[string]any, error) {
	resolved, e := s.resolveCatalogToken(ctx, "product", product)
	if e != nil {
		return nil, e
	}
	product = resolved
	w, e := s.publicSnapshot(ctx, product)
	if e != nil {
		return nil, e
	}
	d, e := BuildProductDetail(w, product, variant, offer)
	if e != nil {
		return nil, e
	}
	out := []map[string]any{}
	category := d.Category.ID
	for _, o := range d.Offers {
		if o.ID != d.SelectedOfferID || !o.Available || o.Price == nil {
			continue
		}
		for _, m := range publicSalesSettings(w).Methods {
			if !m.Active || len(m.PaymentMethodIDs) == 0 || !storeEligibility(w, o.StoreID, category, m.ID) {
				continue
			}
			rate := resolvedMarkup(w, m, category)
			rate.UnitAmount, e = commerce.Percent(commerce.Amount(*o.Price), rate.BPS)
			if e != nil {
				continue
			}
			total, e := commerce.Add(commerce.Amount(*o.Price), rate.UnitAmount)
			if e != nil {
				continue
			}
			payments := w.SalesPayments
			if w.Commerce != nil {
				payments = w.Commerce.Settings.Methods
			}
			if payments == nil {
				payments = commerce.DefaultSettings().Methods
			}
			eligible := false
			for _, id := range m.PaymentMethodIDs {
				for _, p := range payments {
					if p.ID == id && methodEligible(w, p, []commerce.LineSnapshot{{ProductID: product}}, total) {
						eligible = true
					}
				}
			}
			if !eligible {
				continue
			}
			name := o.StoreName
			for _, st := range w.Stores {
				if st.ID == o.StoreID {
					name = storeDisplayName(st, m.ID)
					break
				}
			}
			out = append(out, map[string]any{"store_name": name, "method_id": m.ID, "name": m.Name, "cash_toman": commerce.Toman(commerce.Amount(*o.Price)), "price_toman": commerce.Toman(total), "markup_toman": commerce.Toman(rate.UnitAmount), "markup_bps": rate.BPS, "source_category_id": rate.CategoryID})
		}
	}
	return out, nil
}

func salesPaymentStillAllowed(w Workspace, q commerce.Quote) bool {
	if !storePoliciesStillAllowed(w, q) {
		return false
	}
	if q.PaymentMethod.Kind == "wallet" || q.PaymentMethod.ID == "" {
		return true
	}
	m, e := salesMethodForPayment(w, q.PaymentMethod.ID)
	return e == nil && (q.SalesMethod == nil || q.SalesMethod.ID == m.ID)
}

func publicSalesSettings(w Workspace) *commerce.SalesSettings {
	source := salesSettings(w)
	out := &commerce.SalesSettings{Version: source.Version, Overrides: append([]commerce.CategoryMarkup{}, source.Overrides...)}
	if w.Commerce == nil && w.Sales != nil {
		return source
	}
	settings := commerceState(&w).Settings
	for _, m := range source.Methods {
		ids := []string{}
		for _, id := range m.PaymentMethodIDs {
			for _, p := range settings.Methods {
				if p.ID == id && p.Active && p.Kind != "wallet" {
					ids = append(ids, id)
				}
			}
		}
		m.PaymentMethodIDs = ids
		out.Methods = append(out.Methods, m)
	}
	return out
}
