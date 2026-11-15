package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"regexp"
)

var accountPhone = regexp.MustCompile(`^09[0-9]{9}$`)

func validRole(role identity.Role) bool {
	return role == identity.RoleSuperAdmin || role == identity.RoleOperations || role == identity.RoleSupport || role == identity.RoleFinance || role == identity.RoleViewer
}
func applyUser(state *Workspace, c Command, id string) error {
	if c.Action == "user.create" {
		if c.Name == "" || !accountPhone.MatchString(c.Phone) || !validRole(identity.Role(c.Value)) {
			return ErrInvalid
		}
		for _, existing := range state.Users {
			if existing.Phone == c.Phone {
				return ErrConflict
			}
		}
		state.Users = append(state.Users, User{ID: id, Phone: c.Phone, Name: c.Name, Role: identity.Role(c.Value), Active: true})
		return nil
	}
	for i := range state.Users {
		u := &state.Users[i]
		if u.ID != c.Target {
			continue
		}
		if u.Active && u.Role == identity.RoleSuperAdmin && (c.Action == "user.toggle" || c.Value != string(identity.RoleSuperAdmin)) {
			count := 0
			for _, other := range state.Users {
				if other.Active && other.Role == identity.RoleSuperAdmin {
					count++
				}
			}
			if count <= 1 {
				return ErrConflict
			}
		}
		if c.Action == "user.toggle" {
			u.Active = !u.Active
			return nil
		}
		if !validRole(identity.Role(c.Value)) {
			return ErrInvalid
		}
		u.Role = identity.Role(c.Value)
		return nil
	}
	return ErrNotFound
}
