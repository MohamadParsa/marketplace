package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
	"slices"
	"strings"
)

type StorePreparationVersion struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type StorePreparationBatchInput struct {
	All             bool                      `json:"all"`
	Stores          []StorePreparationVersion `json:"stores"`
	LeadDays        int                       `json:"lead_days"`
	DayBasis        shipping.DayBasis         `json:"day_basis"`
	CalendarMode    string                    `json:"calendar_mode"`
	WorkingWeekdays []int                     `json:"working_weekdays"`
}

func (s *Service) SaveStorePreparationBatch(ctx context.Context, a identity.User, in StorePreparationBatchInput, key string) ([]StorePreparationView, error) {
	out := []StorePreparationView{}
	if a.Surface != identity.SurfaceBackoffice {
		return out, ErrForbidden
	}
	if strings.TrimSpace(key) == "" || len(key) > 128 || len(in.Stores) == 0 || len(in.Stores) > 10000 || in.LeadDays < 0 || in.LeadDays > 365 {
		return out, ErrInvalid
	}
	policy, err := shipping.NormalizePreparation(shipping.PreparationPolicy{DayBasis: in.DayBasis, CalendarMode: in.CalendarMode, WorkingWeekdays: in.WorkingWeekdays})
	if err != nil {
		return out, ErrInvalid
	}
	versions := slices.Clone(in.Stores)
	slices.SortFunc(versions, func(a, b StorePreparationVersion) int { return strings.Compare(a.ID, b.ID) })
	for index, item := range versions {
		if item.ID == "" || item.Version < 0 || index > 0 && versions[index-1].ID == item.ID {
			return out, ErrInvalid
		}
	}
	fingerprint := axisHash([]any{in.All, versions, in.LeadDays, policy})
	err = s.repository.Update(ctx, func(w *Workspace) error {
		actor, err := storeAdmin(*w, a, StoresPreparationWrite)
		if err != nil || !allowsAction(actor, StoresRead) {
			return ErrForbidden
		}
		for _, event := range w.Audit {
			if event.ID != key {
				continue
			}
			if event.Actor != actor.ID || event.Action != "stores.preparation.batch" || event.Fingerprint != fingerprint {
				return ErrConflict
			}
			for _, item := range versions {
				st, err := storeByID(w, item.ID)
				if err != nil {
					return err
				}
				view, err := preparationView(*w, actor, *st)
				if err != nil {
					return err
				}
				out = append(out, view)
			}
			return nil
		}
		// The list is also the caller's snapshot of all stores. A newly added store
		// must be reviewed rather than silently omitted from an "all" operation.
		if in.All && len(versions) != len(w.Stores) {
			return ErrConflict
		}
		for _, item := range versions {
			st, err := storeByID(w, item.ID)
			if err != nil {
				return err
			}
			if st.Version != item.Version {
				return ErrConflict
			}
		}
		for _, item := range versions {
			st, _ := storeByID(w, item.ID)
			copy := policy
			copy.Version = preparationPolicy(*st).Version + 1
			st.Preparation = &copy
			st.LeadDays = in.LeadDays
			st.Version++
			view, err := preparationView(*w, actor, *st)
			if err != nil {
				return err
			}
			out = append(out, view)
		}
		w.Audit = append(w.Audit, Audit{ID: key, Actor: actor.ID, Action: "stores.preparation.batch", Target: "stores", Fingerprint: fingerprint, At: s.clock.Now()})
		ids := make([]string, len(versions))
		for index, item := range versions {
			ids[index] = item.ID
		}
		w.Audit[len(w.Audit)-1].Reason = "stores: " + strings.Join(ids, ",")
		return nil
	})
	return out, err
}
