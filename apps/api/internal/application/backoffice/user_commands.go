package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"regexp"
)

var accountPhone = regexp.MustCompile(`^09[0-9]{9}$`)

func validRole(role identity.Role) bool {
	return role == identity.RoleSuperAdmin || role == identity.RoleOperations || role == identity.RoleSupport || role == identity.RoleFinance || role == identity.RoleViewer
}
func applyUser(state *Workspace, c Command, id string, confirmed bool, event *Audit) error {
	role := identity.Role(c.Value)
	if c.Action == "user.create" {
		if c.Name == "" || !accountPhone.MatchString(c.Phone) || !validRole(role) {
			return ErrInvalid
		}
		for _, existing := range state.Users {
			if existing.Phone == c.Phone {
				return ErrConflict
			}
		}
		if role == identity.RoleSuperAdmin && !confirmed {
			return ErrReauthenticationRequired
		}
		state.Users = append(state.Users, User{ID: id, Phone: c.Phone, Name: c.Name, Role: role, Active: true})
		event.Target, event.Role = id, role
		return nil
	}
	for i := range state.Users {
		u := &state.Users[i]
		if u.ID != c.Target {
			continue
		}
		if c.UserVersion != nil && u.Version != *c.UserVersion {
			return ErrConflict
		}
		if c.Action == "user.role" && !validRole(role) && role != "" {
			return ErrInvalid
		}
		toggle := c.Action == "user.toggle"
		if u.Active && u.Role == identity.RoleSuperAdmin && (toggle || role != identity.RoleSuperAdmin) {
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
		grantsAdmin := (!toggle && role == identity.RoleSuperAdmin && u.Role != role) ||
			(toggle && !u.Active && u.Role == identity.RoleSuperAdmin)
		if grantsAdmin && !confirmed {
			return ErrReauthenticationRequired
		}
		event.PreviousRole, event.Role = u.Role, u.Role
		if toggle {
			u.Version++
			u.Active = !u.Active
			u.AccessVersion++
			return nil
		}
		if u.Role != role {
			u.Version++
			u.Role = role
			u.AccessVersion++
		}
		event.Role = u.Role
		return nil
	}
	return ErrNotFound
}
