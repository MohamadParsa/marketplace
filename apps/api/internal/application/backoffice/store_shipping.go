package backoffice

import (
	"context"
	"reflect"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const StoresShippingRead = "stores.shipping.read"
const StoresShippingWrite = "stores.shipping.write"
const StoresShippingReview = "stores.shipping.review"

type StoreShippingTarget struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}
type StoreShippingCommand struct {
	Action      string                `json:"action"`
	MethodID    string                `json:"method_id"`
	BaseVersion int                   `json:"base_version"`
	All         bool                  `json:"all"`
	Stores      []StoreShippingTarget `json:"stores"`
	Method      *ShippingMethodInput  `json:"method,omitempty"`
	Reason      string                `json:"reason"`
}
type StoreShippingRevisionView struct {
	Number int                 `json:"number"`
	Status string              `json:"status"`
	Base   ShippingMethodInput `json:"base"`
	Method ShippingMethodInput `json:"method"`
	Reason string              `json:"reason"`
	At     string              `json:"at"`
}
type StoreShippingMethodView struct {
	Base        ShippingMethodInput         `json:"base"`
	Enabled     bool                        `json:"enabled"`
	Mode        string                      `json:"mode"`
	Effective   *ShippingMethodInput        `json:"effective,omitempty"`
	Live        *StoreShippingRevisionView  `json:"live,omitempty"`
	Draft       *StoreShippingRevisionView  `json:"draft,omitempty"`
	History     []StoreShippingRevisionView `json:"history"`
	BaseChanged bool                        `json:"base_changed"`
}
type StoreShippingView struct {
	ID        string                    `json:"id"`
	Name      string                    `json:"name"`
	Version   int                       `json:"version"`
	CanWrite  bool                      `json:"can_write"`
	CanReview bool                      `json:"can_review"`
	Methods   []StoreShippingMethodView `json:"methods"`
}
type StoreShippingDirectory struct {
	Stores    []StoreShippingView `json:"stores"`
	Geography []commerce.Province `json:"geography"`
}

func storeShippingPolicy(w Workspace, id string) commerce.StoreShippingPolicy {
	if w.Commerce != nil {
		for _, p := range w.Commerce.StoreShipping {
			if p.StoreID == id {
				return p
			}
		}
	}
	return commerce.StoreShippingPolicy{StoreID: id, Bindings: []commerce.StoreShippingBinding{}}
}
func storeShippingBinding(p commerce.StoreShippingPolicy, methodID string) commerce.StoreShippingBinding {
	for _, b := range p.Bindings {
		if b.MethodID == methodID {
			return b
		}
	}
	return commerce.StoreShippingBinding{MethodID: methodID, Enabled: true, Mode: "public", History: []commerce.StoreShippingRevision{}}
}
func storeShippingAllowed(w Workspace, actor identity.User, id, permission string) bool {
	if actor.Surface == identity.SurfaceBackoffice {
		return allowsAction(actor, StoresRead) && allowsAction(actor, StoresShippingRead) && allowsAction(actor, permission)
	}
	if actor.Surface != identity.SurfaceSeller || permission == StoresShippingReview || w.Access == nil {
		return false
	}
	return identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, actor.ID, id, identity.StorePermission(StoresShippingRead)) && identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, actor.ID, id, identity.StorePermission(permission))
}

// The current public policy remains an upper bound even after a custom approval.
func EffectiveStoreShipping(w Workspace, storeID string, base commerce.ShippingMethod) (commerce.ShippingMethod, bool) {
	b := storeShippingBinding(storeShippingPolicy(w, storeID), base.ID)
	if !base.Active || !b.Enabled || len(base.StoreIDs) > 0 && !slices.Contains(base.StoreIDs, storeID) {
		return commerce.ShippingMethod{}, false
	}
	out := base
	if b.Mode == "custom" {
		if b.Live == nil || b.Live.Status != "approved" {
			return out, false
		}
		out = b.Live.Method
		if shippingMethodInput(out).CollectionMode != shippingMethodInput(base).CollectionMode {
			return out, false
		}
		out.StoreRevision = b.Live.Number
	}
	out.BaseVersion = base.Version
	out.DispatchIdentifier = base.DispatchIdentifier
	out.CollectionMode = shippingMethodInput(base).CollectionMode
	if out.CollectionMode == "collect" {
		out.Fee = 0
	}
	return out, true
}
func storeShippingEligible(w Workspace, base, method commerce.ShippingMethod, address commerce.Address, storeID string, lines []commerce.LineSnapshot) bool {
	return ShippingMethodEligible(w, base, address, storeID, lines) && ShippingMethodEligible(w, method, address, storeID, lines)
}
func shippingRevisionView(r *commerce.StoreShippingRevision) *StoreShippingRevisionView {
	if r == nil {
		return nil
	}
	return &StoreShippingRevisionView{Number: r.Number, Status: r.Status, Base: shippingMethodInput(r.Base), Method: shippingMethodInput(r.Method), Reason: r.Reason, At: r.At.Format("2006-01-02T15:04:05Z07:00")}
}
func (s *Service) ReadStoreShipping(ctx context.Context, a identity.User) (StoreShippingDirectory, error) {
	out := StoreShippingDirectory{Stores: []StoreShippingView{}, Geography: []commerce.Province{}}
	if a.Surface != identity.SurfaceBackoffice && a.Surface != identity.SurfaceSeller {
		return out, ErrForbidden
	}
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return out, err
	}
	actor, err := catalogActor(w, a)
	if err != nil {
		return out, err
	}
	if actor.Surface == identity.SurfaceBackoffice && !storeShippingAllowed(w, actor, "", StoresShippingRead) {
		return out, ErrForbidden
	}
	out.Geography = publicGeography(preparationSettings(w).Geography)
	for _, st := range w.Stores {
		if !storeShippingAllowed(w, actor, st.ID, StoresShippingRead) {
			continue
		}
		p := storeShippingPolicy(w, st.ID)
		row := StoreShippingView{ID: st.ID, Name: st.Name, Version: p.Version, CanWrite: storeShippingAllowed(w, actor, st.ID, StoresShippingWrite), CanReview: storeShippingAllowed(w, actor, st.ID, StoresShippingReview), Methods: []StoreShippingMethodView{}}
		for _, base := range preparationSettings(w).Shipping {
			b := storeShippingBinding(p, base.ID)
			v := StoreShippingMethodView{Base: shippingMethodInput(base), Enabled: b.Enabled, Mode: b.Mode, Live: shippingRevisionView(b.Live), Draft: shippingRevisionView(b.Draft), History: []StoreShippingRevisionView{}}
			if b.Draft != nil {
				v.BaseChanged = axisHash(b.Draft.Base) != axisHash(base)
			}
			for _, history := range b.History {
				v.History = append(v.History, *shippingRevisionView(&history))
			}
			if effective, ok := EffectiveStoreShipping(w, st.ID, base); ok {
				in := shippingMethodInput(effective)
				v.Effective = &in
			}
			row.Methods = append(row.Methods, v)
		}
		out.Stores = append(out.Stores, row)
	}
	sort.Slice(out.Stores, func(i, j int) bool { return out.Stores[i].Name < out.Stores[j].Name })
	return out, nil
}

func normalizeStoreShipping(w Workspace, base commerce.ShippingMethod, in ShippingMethodInput) (commerce.ShippingMethod, error) {
	if in.DispatchIdentifier != base.DispatchIdentifier {
		return commerce.ShippingMethod{}, shippingInvalid("dispatch_identifier", "نوع شناسه ارسال از روش عمومی ارث می‌رسد.")
	}
	if in.ID != base.ID || strings.TrimSpace(in.Name) != strings.TrimSpace(base.Name) || in.CollectionMode != shippingMethodInput(base).CollectionMode || !in.Active || len(in.StoreIDs) > 0 {
		return commerce.ShippingMethod{}, shippingInvalid("collection_mode", "نوع دریافت هزینه و هویت روش از عمومی ارث می‌رسند؛ تغییر آن‌ها مجاز نیست.")
	}
	in.Version = base.Version
	method, err := normalizeShippingMethod(w, in, true)
	if err != nil {
		return method, err
	}
	if !base.AllCities {
		ids, unresolved := shippingCityIDs(preparationSettings(w).Geography, base.Cities)
		ids = append(ids, base.CityIDs...)
		if method.AllCities || len(unresolved) > 0 {
			return method, shippingInvalid("city_ids", "پوشش اختصاصی باید محدودتر از روش عمومی باشد.")
		}
		for _, id := range method.CityIDs {
			if !slices.Contains(ids, id) {
				return method, shippingInvalid("city_ids", "شهر خارج از پوشش عمومی است.")
			}
		}
	}
	// Evaluate the entire known tree, including descendants and denied ancestors.
	for _, category := range w.Categories {
		if !category.Archived && ShippingAllowsCategory(w, method, category.ID) && !ShippingAllowsCategory(w, base, category.ID) {
			return method, shippingInvalid("allowed_categories", "پوشش دسته از سقف عمومی بیشتر است؛ ممنوعیت‌های عمومی را نیز حفظ کنید.")
		}
	}
	return method, nil
}

func (s *Service) SaveStoreShipping(ctx context.Context, a identity.User, in StoreShippingCommand, key string) (StoreShippingDirectory, error) {
	if a.Surface != identity.SurfaceBackoffice && a.Surface != identity.SurfaceSeller {
		return StoreShippingDirectory{}, ErrForbidden
	}
	if strings.TrimSpace(key) == "" || len(key) > 128 || len(in.Stores) == 0 || len(in.Stores) > 10000 || in.BaseVersion < 0 {
		return StoreShippingDirectory{}, ErrInvalid
	}
	if !slices.Contains([]string{"public", "disable", "draft", "submit", "approve", "reject"}, in.Action) {
		return StoreShippingDirectory{}, ErrInvalid
	}
	if (in.Action == "draft") != (in.Method != nil) {
		return StoreShippingDirectory{}, ErrInvalid
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if utf8.RuneCountInString(in.Reason) > 1000 || (in.Action == "approve" || in.Action == "reject") && in.Reason == "" {
		return StoreShippingDirectory{}, shippingInvalid("reason", "برای تصمیم بررسی دلیل بنویسید.")
	}
	sort.Slice(in.Stores, func(i, j int) bool { return in.Stores[i].ID < in.Stores[j].ID })
	for i, st := range in.Stores {
		if st.ID == "" || st.Version < 0 || i > 0 && in.Stores[i-1].ID == st.ID {
			return StoreShippingDirectory{}, ErrInvalid
		}
	}
	if in.All && (a.Surface != identity.SurfaceBackoffice || in.Action == "approve" || in.Action == "reject") {
		return StoreShippingDirectory{}, ErrForbidden
	}
	err := s.repository.Update(ctx, func(w *Workspace) error {
		actor, err := catalogActor(*w, a)
		if err != nil {
			return err
		}
		permission := StoresShippingWrite
		if in.Action == "approve" || in.Action == "reject" {
			permission = StoresShippingReview
		}
		for _, st := range in.Stores {
			if !storeShippingAllowed(*w, actor, st.ID, permission) {
				return ErrForbidden
			}
			if _, err = storeByID(w, st.ID); err != nil {
				return err
			}
		}
		fingerprint := axisHash(in)
		for _, event := range w.Audit {
			if event.ID == key {
				if event.Actor != actor.ID || event.Action != "stores.shipping."+in.Action || event.Fingerprint != fingerprint {
					return ErrConflict
				}
				return nil
			}
		}
		if in.All && len(in.Stores) != len(w.Stores) {
			return ErrConflict
		}
		var base commerce.ShippingMethod
		found := false
		for _, m := range preparationSettings(*w).Shipping {
			if m.ID == in.MethodID {
				base = m
				found = true
			}
		}
		if !found {
			return ErrNotFound
		}
		if in.Action != "disable" && in.Action != "reject" && (!base.Active || base.Version != in.BaseVersion) {
			return ErrConflict
		}
		var method commerce.ShippingMethod
		if in.Method != nil {
			method, err = normalizeStoreShipping(*w, base, *in.Method)
			if err != nil {
				return err
			}
		}
		for _, st := range in.Stores {
			if len(base.StoreIDs) > 0 && !slices.Contains(base.StoreIDs, st.ID) && in.Action != "disable" && in.Action != "reject" {
				return ErrForbidden
			}
			p := storeShippingPolicy(*w, st.ID)
			if p.Version != st.Version {
				return ErrConflict
			}
			b := storeShippingBinding(p, base.ID)
			switch in.Action {
			case "public":
				b.Enabled = true
				b.Mode = "public"
				if b.Draft != nil {
					previous := *b.Draft
					previous.Status = "withdrawn"
					previous.Actor = actor.ID
					previous.At = s.clock.Now()
					b.History = append(b.History, previous)
					b.Draft = nil
				}
			case "disable":
				b.Enabled = false
			case "draft":
				number := 1
				for _, r := range b.History {
					number = max(number, r.Number+1)
				}
				if b.Draft != nil {
					number = max(number, b.Draft.Number+1)
					b.History = append(b.History, *b.Draft)
				}
				if b.Live != nil {
					number = max(number, b.Live.Number+1)
				}
				b.Draft = &commerce.StoreShippingRevision{Number: number, Status: "draft", Base: base, Method: method, Actor: actor.ID, At: s.clock.Now()}
				b.Mode = "custom"
				b.Enabled = true
			case "submit", "approve", "reject":
				if b.Draft == nil {
					return ErrConflict
				}
				r := *b.Draft
				if in.Action == "submit" && r.Status != "draft" || in.Action != "submit" && r.Status != "submitted" {
					return ErrConflict
				}
				if in.Action != "reject" {
					if !reflect.DeepEqual(r.Base, base) {
						return ErrConflict
					}
					if _, err := normalizeStoreShipping(*w, base, shippingMethodInput(r.Method)); err != nil {
						return err
					}
				}
				r.Status = map[string]string{"submit": "submitted", "approve": "approved", "reject": "rejected"}[in.Action]
				r.Reason = in.Reason
				r.Actor = actor.ID
				r.At = s.clock.Now()
				b.Draft = &r
				b.History = append(b.History, r)
				if in.Action == "approve" {
					live := r
					b.Live = &live
				}
			}
			if len(b.History) > 2000 {
				return ErrInvalid
			}
			index := slices.IndexFunc(p.Bindings, func(v commerce.StoreShippingBinding) bool { return v.MethodID == base.ID })
			if index < 0 {
				p.Bindings = append(p.Bindings, b)
			} else {
				p.Bindings[index] = b
			}
			p.Version++
			if w.Commerce == nil {
				w.Commerce = &commerce.State{Settings: commerce.DefaultSettings()}
			}
			index = slices.IndexFunc(w.Commerce.StoreShipping, func(v commerce.StoreShippingPolicy) bool { return v.StoreID == st.ID })
			if index < 0 {
				w.Commerce.StoreShipping = append(w.Commerce.StoreShipping, p)
			} else {
				w.Commerce.StoreShipping[index] = p
			}
		}
		w.Audit = append(w.Audit, Audit{ID: key, Actor: actor.ID, Action: "stores.shipping." + in.Action, Target: in.MethodID, Fingerprint: fingerprint, Reason: in.Reason, At: s.clock.Now()})
		return nil
	})
	if err != nil {
		return StoreShippingDirectory{}, err
	}
	return s.ReadStoreShipping(ctx, a)
}
