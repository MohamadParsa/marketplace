package backoffice

import (
	"context"
	"slices"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

// ActivityOrigin is private correlation metadata, not a financial snapshot or
// session credential. It expires with behavior retention and is never public.
type ActivityOrigin struct {
	OperationID string    `json:"operation_id"`
	SessionRef  string    `json:"session_ref,omitempty"`
	Device      string    `json:"device"`
	At          time.Time `json:"at"`
}
type committedBehavior struct {
	event  operationlog.Event
	origin ActivityOrigin
}

func cartAlias(id string) string {
	if id == "" {
		return ""
	}
	return "g-" + hashToken(id)[:24]
}
func (s *Service) AttachBehaviorSession(ctx context.Context, guest string) {
	if guest == "" || len(guest) < 32 || len(guest) > 128 {
		return
	}
	if lookup, ok := s.repository.(interface {
		FindBehaviorCart(context.Context, string) (string, error)
	}); ok {
		if id, err := lookup.FindBehaviorCart(ctx, hashToken(guest)); err == nil && id != "" {
			operationlog.SetSession(ctx, cartAlias(id))
		}
	}
}
func (s *Service) RecordCustomerActivity(ctx context.Context, name, entity string, attrs map[string]any) {
	ctx = s.observeContext(ctx)
	subject := ""
	trace := operationlog.Context(ctx)
	if trace.ActorKind == "storefront" {
		subject = trace.ActorID
	}
	operationlog.Record(ctx, operationlog.Event{EventName: name, Type: "behavior", Severity: "info", Outcome: "succeeded", SubjectID: subject, EntityID: entity, Attributes: attrs})
}
func originFor(ctx context.Context, cart string) ActivityOrigin {
	d := operationlog.Context(ctx)
	ref := d.SessionRef
	if ref == "" {
		ref = cartAlias(cart)
	}
	device := d.Device
	if device == "" {
		device = "unknown"
	}
	return ActivityOrigin{OperationID: d.OperationID, SessionRef: ref, Device: device, At: time.Now().UTC()}
}
func publishBehavior(ctx context.Context, b committedBehavior) {
	d := operationlog.Context(ctx)
	if b.origin.OperationID != "" && d.OperationID != b.origin.OperationID {
		d.ParentOperationID = d.OperationID
		d.OperationID = b.origin.OperationID
	}
	d.SessionRef = b.origin.SessionRef
	if b.event.Attributes == nil {
		b.event.Attributes = map[string]any{}
	}
	b.event.Attributes["device"] = b.origin.Device
	operationlog.Record(operationlog.WithContext(ctx, d), b.event)
}
func collectBehavior(ctx context.Context, before Workspace, w *Workspace) []committedBehavior {
	if w.Commerce == nil {
		return nil
	}
	c := w.Commerce
	oldC := &commerce.State{}
	if before.Commerce != nil {
		oldC = before.Commerce
	}
	if w.ActivityOrigins == nil {
		w.ActivityOrigins = map[string]ActivityOrigin{}
	}
	cutoff := time.Now().Add(-time.Duration(operationlog.FromContext(ctx).BehaviorRetention()) * 24 * time.Hour)
	for id, origin := range w.ActivityOrigins {
		if origin.At.Before(cutoff) {
			delete(w.ActivityOrigins, id)
		}
	}
	if len(w.ActivityOrigins) > 50000 {
		ids := make([]string, 0, len(w.ActivityOrigins))
		for id := range w.ActivityOrigins {
			ids = append(ids, id)
		}
		slices.SortFunc(ids, func(a, b string) int {
			if w.ActivityOrigins[a].At.Before(w.ActivityOrigins[b].At) {
				return -1
			}
			return 1
		})
		for _, id := range ids[:len(ids)-50000] {
			delete(w.ActivityOrigins, id)
		}
	}
	oldQuotes := map[string]bool{}
	for _, q := range oldC.Quotes {
		oldQuotes[q.ID] = true
	}
	oldPurchases := map[string]string{}
	for _, p := range oldC.Purchases {
		oldPurchases[p.ID] = p.Status
	}
	productCategory := map[string]string{}
	for _, p := range w.Products {
		productCategory[p.ID] = p.CategoryID
	}
	offerProduct := map[string]string{}
	for _, o := range w.Offers {
		offerProduct[o.ID] = o.ProductID
	}
	categoryIDs := func(q commerce.Quote) []string {
		ids := []string{}
		for _, line := range q.Lines {
			if id := productCategory[line.ProductID]; id != "" && !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
		return ids
	}
	events := []committedBehavior{}
	for _, q := range c.Quotes {
		if oldQuotes[q.ID] {
			continue
		}
		origin := originFor(ctx, q.CartID)
		w.ActivityOrigins[q.ID] = origin
		events = append(events, committedBehavior{operationlog.Event{EventName: "checkout.started", Type: "behavior", Severity: "info", Outcome: "succeeded", SubjectID: q.UserID, EntityType: "quote", EntityID: q.ID, Attributes: map[string]any{"quote_id": q.ID, "cart_id": q.CartID, "category_ids": categoryIDs(q)}}, origin})
	}
	consumedCarts := map[string]bool{}
	for _, p := range c.Purchases {
		previous, existed := oldPurchases[p.ID]
		if existed && !(p.Status == "confirmed" && previous != "confirmed") {
			continue
		}
		var q *commerce.Quote
		for i := range c.Quotes {
			if c.Quotes[i].ID == p.QuoteID {
				q = &c.Quotes[i]
				break
			}
		}
		if q == nil {
			continue
		}
		origin, ok := w.ActivityOrigins[q.ID]
		if !ok {
			origin = ActivityOrigin{Device: "unknown", At: time.Now().UTC()}
		}
		attrs := map[string]any{"quote_id": q.ID, "category_ids": categoryIDs(*q), "amount": commerce.Toman(q.Total), "currency": "IRT"}
		base := operationlog.Event{Type: "behavior", Severity: "info", Outcome: "succeeded", SubjectID: p.UserID, EntityType: "purchase", EntityID: p.ID, PurchaseID: p.ID, Attributes: attrs}
		if !existed {
			base.EventName = "checkout.placed"
			events = append(events, committedBehavior{base, origin})
			consumedCarts[q.CartID] = true
		}
		if p.Status == "confirmed" && previous != "confirmed" {
			base.EventName = "checkout.paid"
			events = append(events, committedBehavior{base, origin})
		}
	}
	oldCarts := map[string]commerce.Cart{}
	for _, cart := range oldC.Carts {
		oldCarts[cart.ID] = cart
	}
	merged := map[string]bool{}
	for _, cart := range c.Carts {
		old := oldCarts[cart.ID]
		if cart.MergedInto != "" && old.MergedInto == "" {
			merged[cart.ID] = true
			merged[cart.MergedInto] = true
			events = append(events, committedBehavior{operationlog.Event{EventName: "cart.merged", Type: "behavior", Severity: "info", Outcome: "succeeded", EntityType: "cart", EntityID: cart.MergedInto, Attributes: map[string]any{"cart_id": cart.MergedInto}}, originFor(ctx, cart.ID)})
		}
	}
	for _, cart := range c.Carts {
		if consumedCarts[cart.ID] || merged[cart.ID] || cart.MergedInto != "" {
			continue
		}
		old := oldCarts[cart.ID]
		previous, next := map[string]int{}, map[string]int{}
		for _, line := range old.Lines {
			previous[line.OfferID] = line.Quantity
		}
		for _, line := range cart.Lines {
			next[line.OfferID] = line.Quantity
		}
		ids := []string{}
		for id := range previous {
			ids = append(ids, id)
		}
		for id := range next {
			if !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
		slices.Sort(ids)
		for _, offer := range ids {
			if previous[offer] == next[offer] {
				continue
			}
			name := "cart.item-quantity-changed"
			if previous[offer] == 0 {
				name = "cart.item-added"
			} else if next[offer] == 0 {
				name = "cart.item-removed"
			}
			product := offerProduct[offer]
			events = append(events, committedBehavior{operationlog.Event{EventName: name, Type: "behavior", Severity: "info", Outcome: "succeeded", SubjectID: cart.UserID, EntityType: "cart", EntityID: cart.ID, Attributes: map[string]any{"cart_id": cart.ID, "offer_id": offer, "product_id": product, "category_id": productCategory[product], "quantity": next[offer]}}, originFor(ctx, cart.ID)})
		}
	}
	return events
}
