package backoffice

import (
	"slices"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
)

const (
	WorkspaceRead = "workspace.read"
	FinanceRead   = "finance.read"
)

func permissionRegistry() map[identity.RoleScope][]string {
	return map[identity.RoleScope][]string{
		identity.ScopeGlobal: {
			WorkspaceRead, FinanceRead, CommerceSettingsRead, CommercePaymentWrite, CommerceShippingWrite, CommerceBusinessWrite, CommerceOrdersRead, CommerceOrdersConfirm, CommerceDepositsRead, CommerceDepositsReview, CommerceRefundsResolve, "category.publish", "banner.read", "banner.create", "banner.save", "banner.publish", "banner.media.set", "banner.media.remove", "category.icon.set", "category.icon.remove", "brand.icon.set", "brand.icon.remove", "attribute.create", "attribute.update", "attribute.delete", "category.bindings.set", "settings.card-limit", "variant.create", "offer.create", "offer.price.update", "offer.stock.update", "homepage.save", "catalog.category.create", "catalog.category.update", "catalog.brand.create", "catalog.brand.update", "catalog.product.create", "catalog.product.update", "product.media.update", "product.icon.set", "product.icon.remove", "store.create", "store.toggle", "category.create",
			"category.archive", "product.publish", "product.reject", "shipping.update",
			"case.start", "case.close", "case.refund", "case.assign", "order.deliver", "transfer.resolve",
			"rule.toggle", "notification.retry",
		},
		identity.ScopeStore: {"orders.read", string(identity.StoreOffersRead), string(identity.StoreStockWrite), "offers.create", "offers.price.write"},
	}
}

func IsAccessCommand(action string) bool {
	return strings.HasPrefix(action, "user.") || strings.HasPrefix(action, "role.") || strings.HasPrefix(action, "membership.")
}

func (w Workspace) BackofficeIdentity(account User) (identity.User, bool) {
	var permissions []string
	if w.Access != nil {
		permissions = identity.GrantedPermissions(w.Access.Roles, w.Access.Grants, account.ID, "")
	}
	return identity.User{
		ID: account.ID, Name: account.Name, Phone: account.Phone, Email: account.Email, Active: account.Active, Role: account.Role, PasswordHash: account.PasswordHash,
		Surface: identity.SurfaceBackoffice, AccessVersion: account.AccessVersion, Permissions: permissions,
	}, account.Role != "" || len(permissions) > 0
}

func applyAccess(state *Workspace, c Command, id string, event *Audit) error {
	if state.Access == nil {
		state.Access = &AccessDirectory{}
	}
	switch c.Action {
	case "role.create", "role.update", "role.delete":
		return applyRoleDefinition(state, c, id, event)
	case "user.role-grant", "user.role-revoke", "membership.add", "membership.remove":
		return applyAssignment(state, c, event)
	default:
		return ErrInvalid
	}
}

func applyRoleDefinition(state *Workspace, c Command, id string, event *Audit) error {
	directory := state.Access
	i := slices.IndexFunc(directory.Roles, func(role identity.RoleDefinition) bool { return role.ID == c.Target })
	if c.Action != "role.create" {
		if i < 0 {
			return ErrNotFound
		}
		role := directory.Roles[i]
		if role.Version != c.Version {
			return ErrConflict
		}
		if c.Scope != "" && c.Scope != role.Scope {
			return ErrInvalid
		}
		c.Scope = role.Scope
		event.RoleID = role.ID
		if c.Action == "role.delete" {
			if slices.ContainsFunc(directory.Grants, func(grant identity.RoleGrant) bool { return grant.RoleID == role.ID }) {
				return ErrConflict
			}
			directory.Roles = slices.Delete(directory.Roles, i, i+1)
			return nil
		}
	}
	name := strings.TrimSpace(c.Name)
	allowed := permissionRegistry()[c.Scope]
	if name == "" || len(c.Permissions) == 0 || len(c.Permissions) > len(allowed) {
		return ErrInvalid
	}
	if slices.ContainsFunc(directory.Roles, func(existing identity.RoleDefinition) bool {
		return (c.Action == "role.create" || existing.ID != c.Target) && existing.Scope == c.Scope && strings.EqualFold(existing.Name, name)
	}) {
		return ErrConflict
	}
	permissions := slices.Clone(c.Permissions)
	slices.Sort(permissions)
	for _, permission := range permissions {
		if !slices.Contains(allowed, permission) {
			return ErrInvalid
		}
	}
	permissions = slices.Compact(permissions)
	event.Permissions = permissions
	if c.Action == "role.create" {
		directory.Roles = append(directory.Roles, identity.RoleDefinition{ID: id, Name: name, Scope: c.Scope, Permissions: permissions, Version: 1})
		event.Target, event.RoleID = id, id
		return nil
	}
	role := &directory.Roles[i]
	if role.Scope == identity.ScopeGlobal && !slices.Equal(role.Permissions, permissions) {
		for _, grant := range directory.Grants {
			if grant.RoleID == role.ID {
				if userIndex := findUser(state, grant.UserID); userIndex >= 0 {
					state.Users[userIndex].AccessVersion++
				}
			}
		}
	}
	role.Name, role.Permissions, role.Version = name, permissions, role.Version+1
	return nil
}

func applyAssignment(state *Workspace, c Command, event *Audit) error {
	userIndex := findUser(state, c.Target)
	if userIndex < 0 {
		return ErrNotFound
	}
	membership := identity.StoreMembership{UserID: c.Target, StoreID: c.StoreID}
	isMember := slices.Contains(state.StoreMemberships, membership)
	event.StoreID = c.StoreID
	switch c.Action {
	case "membership.add":
		if c.StoreID == "" || !slices.ContainsFunc(state.Stores, func(s store.Store) bool { return s.ID == c.StoreID }) {
			return ErrNotFound
		}
		if !isMember {
			state.StoreMemberships = append(state.StoreMemberships, membership)
		}
		return nil
	case "membership.remove":
		if !isMember {
			return ErrNotFound
		}
		state.StoreMemberships = slices.DeleteFunc(state.StoreMemberships, func(current identity.StoreMembership) bool { return current == membership })
		state.Access.Grants = slices.DeleteFunc(state.Access.Grants, func(grant identity.RoleGrant) bool {
			return grant.UserID == c.Target && grant.StoreID == c.StoreID
		})
		return nil
	}
	roleIndex := slices.IndexFunc(state.Access.Roles, func(role identity.RoleDefinition) bool { return role.ID == c.Value })
	if roleIndex < 0 {
		return ErrNotFound
	}
	role := state.Access.Roles[roleIndex]
	if (role.Scope == identity.ScopeGlobal && c.StoreID != "") || (role.Scope == identity.ScopeStore && !isMember) {
		return ErrInvalid
	}
	grant := identity.RoleGrant{UserID: c.Target, StoreID: c.StoreID, RoleID: c.Value}
	event.RoleID = grant.RoleID
	i := slices.Index(state.Access.Grants, grant)
	switch {
	case c.Action == "user.role-grant" && i < 0:
		state.Access.Grants = append(state.Access.Grants, grant)
	case c.Action == "user.role-revoke" && i >= 0:
		state.Access.Grants = slices.Delete(state.Access.Grants, i, i+1)
	default:
		return nil
	}
	if role.Scope == identity.ScopeGlobal {
		state.Users[userIndex].AccessVersion++
	}
	return nil
}

func findUser(state *Workspace, id string) int {
	return slices.IndexFunc(state.Users, func(user User) bool { return user.ID == id })
}
