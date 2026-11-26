package backoffice

import (
	"reflect"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

// Compare customer terms, not quota/version counters. A lower quota must honor an existing hold.
func shippingTermsCurrent(w Workspace, ship commerce.ShipmentSnapshot, current commerce.ShippingMethod) bool {
	old, next := shippingMethodInput(ship.Method), shippingMethodInput(current)
	if old.Name != next.Name || old.FeeToman != next.FeeToman || old.CollectionMode != next.CollectionMode || old.DeliveryMode != next.DeliveryMode || old.MinDays != next.MinDays || old.MaxDays != next.MaxDays || old.DayBasis != next.DayBasis {
		return false
	}
	if ship.Preparation == nil {
		return true
	}
	for _, st := range w.Stores {
		if st.ID != ship.StoreID {
			continue
		}
		resolved, _, _, err := shipmentPreparation(preparationSettings(w), st, current, ship.Preparation.SourceAt)
		if err != nil {
			return false
		}
		captured := *ship.Preparation
		captured.Policy.Version, resolved.Policy.Version = 0, 0
		captured.Calendar.Version, resolved.Calendar.Version = 0, 0
		captured.SourceAt, resolved.SourceAt = time.Time{}, time.Time{}
		captured.ReadyAt, resolved.ReadyAt = time.Time{}, time.Time{}
		return reflect.DeepEqual(captured, resolved)
	}
	return false
}
