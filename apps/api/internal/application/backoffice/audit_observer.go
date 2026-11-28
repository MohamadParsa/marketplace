package backoffice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
)

// ObserveWorkspaceUpdate enriches audit INSIDE the repository transaction and
// publishes telemetry only AFTER its commit. A rollback never emits success.
// The repository's specialized interfaces remain intact (no wrapper hiding them).
func ObserveWorkspaceUpdate(ctx context.Context, update func(func(*Workspace) error) error, apply func(*Workspace) error) error {
	if operationlog.FromContext(ctx) == nil {
		return update(apply)
	}
	var committed []Audit
	var behavior []committedBehavior
	err := update(func(w *Workspace) error {
		committed = nil
		behavior = nil
		before := auditSnapshot(*w)
		count := len(w.Audit)
		original, _ := json.Marshal(w.Audit)
		if err := apply(w); err != nil {
			return err
		}
		if len(w.Audit) < count {
			return ErrForbidden
		}
		afterOriginal, _ := json.Marshal(w.Audit[:count])
		if count > 0 && !bytes.Equal(original, afterOriginal) {
			return ErrForbidden
		}
		appendIdentityAudits(ctx, w, before, count)
		appendCriticalAudits(ctx, w, before, count)
		trace := operationlog.Context(ctx)
		for i := count; i < len(w.Audit); i++ {
			a := &w.Audit[i]
			if a.At.IsZero() {
				a.At = time.Now().UTC()
			}
			target := a.Target
			if target == "" {
				target = a.ResultID
			}
			a.Before = projectAudit(before, target, a.Action)
			a.After = projectAudit(*w, target, a.Action)
			a.OperationID, a.RequestID, a.Outcome = trace.OperationID, trace.RequestID, "succeeded"
			if a.StoreID == "" {
				a.StoreID = trace.StoreID
				for _, o := range w.Offers {
					if o.ID == target {
						a.StoreID = o.StoreID
					}
				}
			}
			a.Version, _ = strconv.Atoi(a.After["version"])
			committed = append(committed, *a)
		}
		behavior = collectBehavior(ctx, before, w)
		return nil
	})
	if err != nil {
		return err
	}
	for _, a := range committed {
		eventCtx := operationlog.WithContext(ctx, operationlog.Context(ctx))
		kind := operationlog.Context(ctx).ActorKind
		if kind == "" {
			kind = "system"
		}
		operationlog.SetActor(eventCtx, kind, a.Actor, a.StoreID)
		target := a.Target
		if target == "" {
			target = a.ResultID
		}
		operationlog.Record(eventCtx, operationlog.Event{EventName: a.Action, Type: "business", Severity: "info", Outcome: "succeeded", EntityID: target, EntityType: "audit", Attributes: map[string]any{"audit_id": auditPublicID(a.ID), "version": a.Version, "changed_fields": auditChangeCount(a.Before, a.After)}})
	}
	for _, b := range behavior {
		publishBehavior(ctx, b)
	}
	return nil
}
func auditPublicID(id string) string {
	sum := sha256.Sum256([]byte(id))
	return "a-" + hex.EncodeToString(sum[:12])
}
func auditChangeCount(a, b map[string]string) int {
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	n := 0
	for k := range keys {
		if a[k] != b[k] {
			n++
		}
	}
	return n
}
func auditSnapshot(w Workspace) Workspace {
	w.Audit = slices.Clone(w.Audit)
	if w.Access != nil {
		a := *w.Access
		a.Roles = slices.Clone(a.Roles)
		for i := range a.Roles {
			a.Roles[i].Permissions = slices.Clone(a.Roles[i].Permissions)
		}
		a.Grants = slices.Clone(a.Grants)
		w.Access = &a
	}
	w.Users = slices.Clone(w.Users)
	w.Offers = slices.Clone(w.Offers)
	w.Products = slices.Clone(w.Products)
	w.Stores = slices.Clone(w.Stores)
	w.Orders = slices.Clone(w.Orders)
	w.Categories = slices.Clone(w.Categories)
	w.Variants = slices.Clone(w.Variants)
	if w.Commerce != nil {
		c := *w.Commerce
		c.Coupons = slices.Clone(c.Coupons)
		c.GatewaySettings = slices.Clone(c.GatewaySettings)
		c.Purchases = slices.Clone(c.Purchases)
		c.Quotes = slices.Clone(c.Quotes)
		c.Carts = slices.Clone(c.Carts)
		for i := range c.Carts {
			c.Carts[i].Lines = slices.Clone(c.Carts[i].Lines)
		}
		c.Entries = slices.Clone(c.Entries)
		w.Commerce = &c
	}
	return w
}
func appendIdentityAudits(ctx context.Context, w *Workspace, before Workspace, start int) {
	trace := operationlog.Context(ctx)
	for _, u := range w.Users {
		var previous *User
		for i := range before.Users {
			if before.Users[i].ID == u.ID {
				previous = &before.Users[i]
				break
			}
		}
		changed := previous == nil || previous.Version != u.Version || previous.Role != u.Role || previous.Active != u.Active || previous.PasswordHash != u.PasswordHash || previous.AccessVersion != u.AccessVersion
		if !changed {
			continue
		}
		exists := false
		for _, a := range w.Audit[start:] {
			if a.Target == u.ID || a.ResultID == u.ID {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		actor := trace.ActorID
		if actor == "" {
			actor = "system"
		}
		action := "identity.user-updated"
		if previous == nil {
			action = "identity.user-created"
		}
		w.Audit = append(w.Audit, Audit{ID: operationlog.ID(), Actor: actor, Action: action, Target: u.ID, At: time.Now().UTC()})
	}
}

type criticalMarker struct{ id, action, signature, store string }

func criticalMarkers(w Workspace) map[string]criticalMarker {
	out := map[string]criticalMarker{}
	add := func(kind, id, signature, store string) {
		out[kind+":"+id] = criticalMarker{id, kind + ".changed", signature, store}
	}
	for _, o := range w.Offers {
		add("offer", o.ID, fmt.Sprint(o.Version, "/", o.Price, "/", o.Stock, "/", o.Reserved), o.StoreID)
	}
	for _, p := range w.Products {
		add("catalog.product", p.ID, fmt.Sprint(p.Version, "/", p.Status), p.OwnerStoreID)
	}
	for _, s := range w.Stores {
		add("store", s.ID, fmt.Sprint(s.Version, "/", s.Active, "/", s.ContentPolicyVersion), s.ID)
	}
	for _, o := range w.Orders {
		add("order", o.ID, fmt.Sprint(o.Version, "/", o.Status, "/", o.OrderStatus, "/", o.Payment), o.StoreID)
	}
	if w.Access != nil {
		for _, r := range w.Access.Roles {
			add("role", r.ID, strconv.Itoa(r.Version), "")
		}
	}
	if c := w.Commerce; c != nil {
		add("commerce.settings", "commerce-settings", strconv.Itoa(c.Settings.Version), "")
		for _, v := range c.Coupons {
			add("coupon", v.ID, fmt.Sprint(v.Version, "/", v.Active, "/", v.Value), "")
		}
		for _, v := range c.GatewaySettings {
			add("gateway.settings", v.ID, strconv.Itoa(v.Version), "")
		}
		for _, v := range c.Entries {
			add("wallet.entry", v.ID, fmt.Sprint(v.Amount, "/", v.Kind), "")
		}
	}
	return out
}
func appendCriticalAudits(ctx context.Context, w *Workspace, before Workspace, start int) {
	old, current := criticalMarkers(before), criticalMarkers(*w)
	covered := map[string]bool{}
	for _, a := range w.Audit[start:] {
		covered[a.Target] = true
		covered[a.ResultID] = true
	}
	actor := operationlog.Context(ctx).ActorID
	if actor == "" {
		actor = "system"
	}
	keys := make([]string, 0, len(current))
	for k := range current {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, key := range keys {
		value := current[key]
		previous, existed := old[key]
		if existed && previous.signature == value.signature || covered[value.id] {
			continue
		}
		w.Audit = append(w.Audit, Audit{ID: operationlog.ID(), Actor: actor, Action: value.action, Target: value.id, StoreID: value.store, At: time.Now().UTC()})
		covered[value.id] = true
	}
}

// projectAudit is a closed field list. It never serializes an entity, body,
// credential, identity contact, address, bank detail or arbitrary display name.
func projectAudit(w Workspace, id, action string) map[string]string {
	value := map[string]string{}
	version := func(n int) { value["version"] = strconv.Itoa(n) }
	for _, u := range w.Users {
		if u.ID == id {
			version(u.Version)
			value["active"] = strconv.FormatBool(u.Active)
			value["role"] = string(u.Role)
			value["access_version"] = strconv.Itoa(u.AccessVersion)
			value["password_configured"] = strconv.FormatBool(u.PasswordHash != "")
			if w.Access != nil {
				grants := []string{}
				for _, g := range w.Access.Grants {
					if g.UserID == id {
						grants = append(grants, g.RoleID+"@"+g.StoreID)
					}
				}
				slices.Sort(grants)
				value["grants"] = strings.Join(grants, ", ")
			}
			return value
		}
	}
	if w.Access != nil {
		for _, r := range w.Access.Roles {
			if r.ID == id {
				version(r.Version)
				permissions := slices.Clone(r.Permissions)
				slices.Sort(permissions)
				value["permissions"] = strings.Join(permissions, ", ")
				return value
			}
		}
	}
	for _, o := range w.Offers {
		if o.ID == id {
			version(o.Version)
			value["price_rial"] = strconv.FormatInt(o.Price, 10)
			value["stock"] = strconv.Itoa(o.Stock)
			value["reserved"] = strconv.Itoa(o.Reserved)
			return value
		}
	}
	for _, p := range w.Products {
		if p.ID == id {
			version(p.Version)
			value["status"] = p.Status
			return value
		}
	}
	for _, s := range w.Stores {
		if s.ID == id {
			version(s.Version)
			value["active"] = strconv.FormatBool(s.Active)
			value["content_policy_version"] = strconv.Itoa(s.ContentPolicyVersion)
			return value
		}
	}
	for _, v := range w.Variants {
		if v.ID == id {
			version(v.Version)
			return value
		}
	}
	for _, o := range w.Orders {
		if o.ID == id {
			version(o.Version)
			value["status"] = o.Status
			value["order_status"] = o.OrderStatus
			value["payment"] = o.Payment
			return value
		}
	}
	for _, c := range w.Categories {
		if c.ID == id {
			value["published"] = strconv.FormatBool(c.Published)
			return value
		}
	}
	if w.Commerce != nil {
		for _, entry := range w.Commerce.Entries {
			if entry.ID == id {
				value["amount_rial"] = strconv.FormatInt(int64(entry.Amount), 10)
				value["kind"] = entry.Kind
				return value
			}
		}
		for _, c := range w.Commerce.Coupons {
			if c.ID == id {
				version(c.Version)
				value["active"] = strconv.FormatBool(c.Active)
				value["value"] = strconv.FormatInt(int64(c.Value), 10)
				value["kind"] = c.Kind
				return value
			}
		}
		for _, p := range w.Commerce.Purchases {
			if p.ID == id {
				version(p.Version)
				value["status"] = p.Status
				return value
			}
		}
		for _, g := range w.Commerce.GatewaySettings {
			if g.ID == id {
				version(g.Version)
				value["active"] = strconv.FormatBool(g.Active)
				value["allow_wallet_topup"] = strconv.FormatBool(g.AllowWalletTopup)
				value["credential_count"] = strconv.Itoa(len(g.CredentialNames))
				return value
			}
		}
	}
	return value
}
