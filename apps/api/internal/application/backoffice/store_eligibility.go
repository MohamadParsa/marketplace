package backoffice

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
)

const StoresEligibilityWrite = "stores.eligibility.write"

type StoreSaleMethodChoice struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}
type StoreEligibilityInput struct {
	Version           int                    `json:"version"`
	CategoryMode      string                 `json:"category_mode"`
	AllowedCategories []string               `json:"allowed_categories"`
	DeniedCategories  []string               `json:"denied_categories"`
	Methods           []store.SalePermission `json:"methods"`
}

func StoreAllowsCategory(w Workspace, st store.Store, category string) bool {
	if st.Eligibility == nil {
		return true
	}
	policy := st.Eligibility
	seen := map[string]bool{}
	allowed := policy.CategoryMode == "all"
	for category != "" {
		if seen[category] {
			return false
		}
		seen[category] = true
		for _, id := range policy.DeniedCategories {
			if id == category {
				return false
			}
		}
		for _, id := range policy.AllowedCategories {
			if id == category {
				allowed = true
			}
		}
		found := false
		parent := ""
		if w.publicIndex != nil {
			if c, ok := w.publicIndex.categories[category]; ok {
				found = true
				parent = c.ParentID
			}
		} else {
			for _, c := range w.Categories {
				if c.ID == category {
					found = true
					parent = c.ParentID
					break
				}
			}
		}
		if !found {
			return false
		}
		category = parent
	}
	return allowed
}
func StoreAllowsMethod(st store.Store, method string) bool {
	if st.Eligibility == nil {
		return true
	}
	for _, m := range st.Eligibility.Methods {
		if m.MethodID == method {
			return true
		}
	}
	return false
}
func storeEligibilityVersion(st store.Store) int {
	if st.Eligibility == nil {
		return 0
	}
	return st.Eligibility.Version
}
func storeEligibility(w Workspace, storeID, category, method string) bool {
	for _, st := range w.Stores {
		if st.ID == storeID {
			return st.Active && StoreAllowsCategory(w, st, category) && StoreAllowsMethod(st, method)
		}
	}
	return false
}
func normalizeStoreEligibility(w Workspace, in StoreEligibilityInput) (StoreEligibilityInput, error) {
	if in.Version < 0 || (in.CategoryMode != "all" && in.CategoryMode != "selected") || len(in.AllowedCategories) > 2000 || len(in.DeniedCategories) > 2000 || len(in.Methods) > 100 {
		return in, ErrInvalid
	}
	normalizeIDs := func(ids []string) ([]string, error) {
		out := []string{}
		seen := map[string]bool{}
		for _, id := range ids {
			resolved, e := ResolvePublicReference("c", id)
			if e != nil || seen[resolved] || !categoryIDsValid(w, []string{resolved}) {
				return nil, ErrInvalid
			}
			seen[resolved] = true
			out = append(out, resolved)
		}
		return out, nil
	}
	var e error
	in.AllowedCategories, e = normalizeIDs(in.AllowedCategories)
	if e != nil {
		return in, e
	}
	in.DeniedCategories, e = normalizeIDs(in.DeniedCategories)
	if e != nil {
		return in, e
	}
	in.Methods = append([]store.SalePermission{}, in.Methods...)
	seen := map[string]bool{}
	for i := range in.Methods {
		m := &in.Methods[i]
		m.DisplayName = strings.TrimSpace(m.DisplayName)
		known := false
		for _, method := range salesSettings(w).Methods {
			known = known || method.ID == m.MethodID
		}
		if !known || seen[m.MethodID] || utf8.RuneCountInString(m.DisplayName) > 160 {
			return in, ErrInvalid
		}
		seen[m.MethodID] = true
	}
	return in, nil
}
func (s *Service) SaveStoreEligibility(ctx context.Context, a identity.User, id string, in StoreEligibilityInput, key string) (AdminStoreRow, error) {
	var out AdminStoreRow
	if strings.TrimSpace(key) == "" || len(key) > 128 {
		return out, ErrInvalid
	}
	e := s.repository.Update(ctx, func(w *Workspace) error {
		actor, e := storeAdmin(*w, a, StoresEligibilityWrite)
		if e != nil {
			return e
		}
		if !allowsAction(actor, StoresRead) {
			return ErrForbidden
		}
		normalized, e := normalizeStoreEligibility(*w, in)
		if e != nil {
			return e
		}
		fingerprint := axisHash([]any{id, normalized})
		st, e := storeByID(w, id)
		if e != nil {
			return e
		}
		for _, event := range w.Audit {
			if event.ID == key {
				if event.Action != "stores.eligibility.save" || event.Actor != a.ID || event.Target != id || event.Fingerprint != fingerprint {
					return ErrConflict
				}
				out = privateStoreRow(*w, *st, allowsAction(actor, StoresMembersRead))
				return nil
			}
		}
		if st.Version != in.Version {
			return ErrConflict
		}
		version := storeEligibilityVersion(*st) + 1
		st.Eligibility = &store.Eligibility{Version: version, CategoryMode: normalized.CategoryMode, AllowedCategories: normalized.AllowedCategories, DeniedCategories: normalized.DeniedCategories, Methods: normalized.Methods}
		st.Version++
		out = privateStoreRow(*w, *st, allowsAction(actor, StoresMembersRead))
		w.Audit = append(w.Audit, Audit{ID: key, Actor: a.ID, Action: "stores.eligibility.save", Target: id, Fingerprint: fingerprint, At: s.clock.Now()})
		return nil
	})
	return out, e
}

func storeHasSale(st store.Store, w Workspace, category string) bool {
	if !st.Active || !StoreAllowsCategory(w, st, category) {
		return false
	}
	if st.Eligibility == nil {
		return true
	}
	for _, method := range salesSettings(w).Methods {
		if StoreAllowsMethod(st, method.ID) {
			return true
		}
	}
	return false
}
func quoteSaleMethod(q commerce.Quote) string {
	if q.SalesMethod != nil {
		return q.SalesMethod.ID
	}
	return "cash"
}
func storeDisplayName(st store.Store, method string) string {
	if method != "" && st.Eligibility != nil {
		for _, m := range st.Eligibility.Methods {
			if m.MethodID == method && m.DisplayName != "" {
				return m.DisplayName
			}
		}
	}
	return st.Name
}
func applyStorePolicies(w Workspace, q *commerce.Quote) error {
	method := quoteSaleMethod(*q)
	for i := range q.Lines {
		l := &q.Lines[i]
		p, e := findProduct(&w, l.ProductID)
		if e != nil {
			return ErrCheckoutChanged
		}
		st, e := storeByID(&w, l.StoreID)
		if e != nil || !storeEligibility(w, l.StoreID, p.CategoryID, method) {
			return ErrInvalidMethod
		}
		l.StorePolicyVersion = storeEligibilityVersion(*st)
	}
	return nil
}
func storePoliciesStillAllowed(w Workspace, q commerce.Quote) bool {
	method := quoteSaleMethod(q)
	for _, l := range q.Lines {
		p, e := findProduct(&w, l.ProductID)
		if e != nil {
			return false
		}
		st, e := storeByID(&w, l.StoreID)
		if e != nil || !storeEligibility(w, l.StoreID, p.CategoryID, method) || l.StorePolicyVersion != storeEligibilityVersion(*st) {
			return false
		}
	}
	return true
}

func quoteDisplayMethod(q commerce.Quote) string {
	if q.SalesMethod != nil {
		return q.SalesMethod.ID
	}
	return ""
}
