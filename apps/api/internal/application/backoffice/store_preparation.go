package backoffice

import (
	"context"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
)

const StoresPreparationRead = "stores.preparation.read"
const StoresPreparationWrite = "stores.preparation.write"

type StorePreparationInput struct {
	Version         int               `json:"version"`
	LeadDays        int               `json:"lead_days"`
	DayBasis        shipping.DayBasis `json:"day_basis"`
	CalendarMode    string            `json:"calendar_mode"`
	WorkingWeekdays []int             `json:"working_weekdays"`
}

type StorePreparationView struct {
	ID                string                     `json:"id"`
	Name              string                     `json:"name"`
	Version           int                        `json:"version"`
	LeadDays          int                        `json:"lead_days"`
	Policy            shipping.PreparationPolicy `json:"policy"`
	SiteCalendar      shipping.Calendar          `json:"site_calendar"`
	EffectiveCalendar shipping.Calendar          `json:"effective_calendar"`
	Timing            shipping.TimingPolicy      `json:"timing"`
	CanWrite          bool                       `json:"can_write"`
}

func preparationPolicy(st store.Store) shipping.PreparationPolicy {
	if st.Preparation == nil {
		return shipping.PreparationPolicy{DayBasis: shipping.WorkingDays, CalendarMode: shipping.Inherit, WorkingWeekdays: []int{}}
	}
	return *st.Preparation
}

func siteShippingCalendar(settings commerce.Settings) (shipping.Calendar, error) {
	working := []int{}
	for day := 0; day < 7; day++ {
		closed := false
		for _, value := range settings.ClosedWeekdays {
			closed = closed || value == day
		}
		if !closed {
			working = append(working, day)
		}
	}
	return shipping.NormalizeCalendar(shipping.Calendar{Version: settings.Version, WorkingWeekdays: working, Holidays: settings.Holidays})
}

func shippingTiming(settings commerce.Settings) shipping.TimingPolicy {
	if settings.ShippingTiming == nil {
		return shipping.DefaultTiming()
	}
	return *settings.ShippingTiming
}

func preparationSettings(w Workspace) commerce.Settings {
	if w.Commerce == nil {
		return commerce.DefaultSettings()
	}
	return w.Commerce.Settings
}

func shipmentPreparation(settings commerce.Settings, st store.Store, method commerce.ShippingMethod, source time.Time) (shipping.PreparationSnapshot, string, string, error) {
	calendar, err := siteShippingCalendar(settings)
	if err != nil {
		return shipping.PreparationSnapshot{}, "", "", ErrInvalid
	}
	calendar, err = shipping.EffectiveCalendar(calendar, st.Preparation)
	if err != nil {
		return shipping.PreparationSnapshot{}, "", "", ErrInvalid
	}
	resolved, err := shipping.ResolvePreparation(source, st.LeadDays, preparationPolicy(st), calendar, shippingTiming(settings))
	if err != nil {
		return resolved, "", "", ErrInvalid
	}
	resolved.TransportBasis = method.DayBasis
	if resolved.TransportBasis == "" {
		resolved.TransportBasis = shipping.WorkingDays
	}
	from, to, err := shipmentDeliveryDates(resolved, method)
	return resolved, from, to, err
}

func shipmentDeliveryDates(snapshot shipping.PreparationSnapshot, method commerce.ShippingMethod) (string, string, error) {
	_, from, to, err := shipping.DeliveryDates(snapshot.ReadyAt, 0, shipping.CalendarDays, method.MinDays, method.MaxDays, snapshot.TransportBasis, snapshot.Calendar)
	if err != nil {
		return "", "", ErrInvalid
	}
	return from.Format("2006-01-02"), to.Format("2006-01-02"), nil
}

func finalizeShipmentCalendars(q *commerce.Quote, paidAt time.Time) error {
	for i := range q.Shipments {
		ship := &q.Shipments[i]
		if ship.Preparation == nil {
			continue
		}
		resolved, err := ship.Preparation.Rebase(paidAt)
		if err != nil {
			return ErrInvalid
		}
		from, to, err := shipmentDeliveryDates(resolved, ship.Method)
		if err != nil {
			return err
		}
		ship.Preparation, ship.DeliveryFrom, ship.DeliveryTo = &resolved, from, to
	}
	return nil
}

func preparationAllowed(w Workspace, actor identity.User, id string, write bool) bool {
	if actor.Surface == identity.SurfaceBackoffice {
		return allowsAction(actor, StoresRead) && (!write || allowsAction(actor, StoresPreparationWrite))
	}
	if actor.Surface != identity.SurfaceSeller || w.Access == nil {
		return false
	}
	read := identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, actor.ID, id, identity.StorePermission(StoresPreparationRead))
	return read && (!write || identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, actor.ID, id, identity.StorePermission(StoresPreparationWrite)))
}

func preparationView(w Workspace, actor identity.User, st store.Store) (StorePreparationView, error) {
	settings := preparationSettings(w)
	site, err := siteShippingCalendar(settings)
	if err != nil {
		return StorePreparationView{}, ErrInvalid
	}
	policy := preparationPolicy(st)
	effective, err := shipping.EffectiveCalendar(site, &policy)
	if err != nil {
		return StorePreparationView{}, ErrInvalid
	}
	return StorePreparationView{ID: st.ID, Name: st.Name, Version: st.Version, LeadDays: st.LeadDays, Policy: policy, SiteCalendar: site, EffectiveCalendar: effective, Timing: shippingTiming(settings), CanWrite: preparationAllowed(w, actor, st.ID, true)}, nil
}

func (s *Service) ReadStorePreparations(ctx context.Context, a identity.User) ([]StorePreparationView, error) {
	out := []StorePreparationView{}
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
	if actor.Surface == identity.SurfaceBackoffice && !allowsAction(actor, StoresRead) {
		return out, ErrForbidden
	}
	for _, st := range w.Stores {
		if !preparationAllowed(w, actor, st.ID, false) {
			continue
		}
		view, err := preparationView(w, actor, st)
		if err != nil {
			return out, err
		}
		out = append(out, view)
	}
	return out, nil
}

func (s *Service) PreviewStorePreparation(ctx context.Context, a identity.User, id string, in StorePreparationInput) (shipping.PreparationSnapshot, error) {
	var out shipping.PreparationSnapshot
	if a.Surface != identity.SurfaceBackoffice && a.Surface != identity.SurfaceSeller {
		return out, ErrForbidden
	}
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return out, err
	}
	actor, err := catalogActor(w, a)
	if err != nil || !preparationAllowed(w, actor, id, false) {
		return out, ErrForbidden
	}
	st, err := storeByID(&w, id)
	if err != nil {
		return out, err
	}
	if st.Version != in.Version {
		return out, ErrConflict
	}
	if in.LeadDays < 0 || in.LeadDays > 365 {
		return out, ErrInvalid
	}
	policy, err := shipping.NormalizePreparation(shipping.PreparationPolicy{DayBasis: in.DayBasis, CalendarMode: in.CalendarMode, WorkingWeekdays: in.WorkingWeekdays})
	if err != nil {
		return out, ErrInvalid
	}
	settings := preparationSettings(w)
	site, err := siteShippingCalendar(settings)
	if err != nil {
		return out, ErrInvalid
	}
	calendar, err := shipping.EffectiveCalendar(site, &policy)
	if err != nil {
		return out, ErrInvalid
	}
	out, err = shipping.ResolvePreparation(s.clock.Now(), in.LeadDays, policy, calendar, shippingTiming(settings))
	if err != nil {
		return out, ErrInvalid
	}
	return out, nil
}

func (s *Service) SaveStorePreparation(ctx context.Context, a identity.User, id string, in StorePreparationInput, key string) (StorePreparationView, error) {
	var out StorePreparationView
	if a.Surface != identity.SurfaceBackoffice && a.Surface != identity.SurfaceSeller {
		return out, ErrForbidden
	}
	if strings.TrimSpace(key) == "" || len(key) > 128 || in.Version < 0 || in.LeadDays < 0 || in.LeadDays > 365 {
		return out, ErrInvalid
	}
	policy, err := shipping.NormalizePreparation(shipping.PreparationPolicy{DayBasis: in.DayBasis, CalendarMode: in.CalendarMode, WorkingWeekdays: in.WorkingWeekdays})
	if err != nil {
		return out, ErrInvalid
	}
	err = s.repository.Update(ctx, func(w *Workspace) error {
		actor, err := catalogActor(*w, a)
		if err != nil {
			return err
		}
		if !preparationAllowed(*w, actor, id, true) {
			return ErrForbidden
		}
		st, err := storeByID(w, id)
		if err != nil {
			return err
		}
		fingerprint := axisHash([]any{id, in.Version, in.LeadDays, policy})
		for _, event := range w.Audit {
			if event.ID != key {
				continue
			}
			if event.Actor != actor.ID || event.Action != "stores.preparation.save" || event.Target != id || event.Fingerprint != fingerprint {
				return ErrConflict
			}
			out, err = preparationView(*w, actor, *st)
			return err
		}
		if st.Version != in.Version {
			return ErrConflict
		}
		policy.Version = preparationPolicy(*st).Version + 1
		st.Preparation = &policy
		st.LeadDays = in.LeadDays
		st.Version++
		out, err = preparationView(*w, actor, *st)
		if err != nil {
			return err
		}
		w.Audit = append(w.Audit, Audit{ID: key, Actor: actor.ID, Action: "stores.preparation.save", Target: id, Fingerprint: fingerprint, At: s.clock.Now()})
		return nil
	})
	return out, err
}
