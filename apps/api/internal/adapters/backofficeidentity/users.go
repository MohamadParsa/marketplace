package backofficeidentity

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
	"time"
)

type Users struct{ repository backoffice.Repository }

func New(repository backoffice.Repository) *Users { return &Users{repository: repository} }
func (u *Users) FindByPhone(ctx context.Context, surface identity.Surface, phone string) (identity.User, error) {
	state, err := u.snapshot(ctx)
	if err != nil {
		return identity.User{}, err
	}
	for _, account := range state.Users {
		if account.Phone != phone {
			continue
		}
		user := identity.User{FirstName: account.FirstName, LastName: account.LastName, Gender: account.Gender, PasswordChangeRequired: account.PasswordChangeRequired, ID: account.ID, Name: account.Name, Phone: phone, Email: account.Email, NationalID: account.NationalID, PasswordHash: account.PasswordHash, Surface: surface, Active: account.Active, AccessVersion: account.AccessVersion}
		switch surface {
		case identity.SurfaceBackoffice:
			if authenticated, ok := state.BackofficeIdentity(account); ok && (!account.SellerOnly || len(authenticated.Permissions) > 0) {
				return authenticated, nil
			}
		case identity.SurfaceCustomer:
			user.Name, user.Email, user.PasswordHash = account.Name, account.Email, account.PasswordHash
			return user, nil
		case identity.SurfaceSeller:
			user.StoreIDs = identity.MemberStoreIDs(state.StoreMemberships, user.ID)
			if len(user.StoreIDs) == 1 {
				user.StoreID = user.StoreIDs[0]
			}
			if len(user.StoreIDs) > 0 {
				if state.Access != nil {
					user.StorePermissions = make(map[string][]string)
					for _, storeID := range user.StoreIDs {
						user.StorePermissions[storeID] = identity.GrantedPermissions(state.Access.Roles, state.Access.Grants, user.ID, storeID)
					}
				}
				return user, nil
			}
		}
		break
	}
	return identity.User{}, identity.ErrUnknown
}

func (u *Users) FindByEmail(ctx context.Context, surface identity.Surface, email string) (identity.User, error) {
	state, err := u.snapshot(ctx)
	if err != nil {
		return identity.User{}, err
	}
	for _, account := range state.Users {
		if account.Active && account.Email != "" && strings.EqualFold(account.Email, identity.NormalizeEmail(email)) {
			return u.FindByPhone(ctx, surface, account.Phone)
		}
	}
	return identity.User{}, identity.ErrUnknown
}

func (u *Users) UpdateProfile(ctx context.Context, surface identity.Surface, userID, name, email, passwordHash string, nationalID *string) error {
	return u.UpdateProfileVersion(ctx, surface, userID, name, email, passwordHash, nationalID, nil)
}
func (u *Users) UpdateProfileVersion(ctx context.Context, surface identity.Surface, userID, name, email, passwordHash string, nationalID *string, version *int) error {
	return u.repository.Update(ctx, func(state *backoffice.Workspace) error {
		i := -1
		for n := range state.Users {
			if state.Users[n].ID == userID && state.Users[n].Active {
				i = n
				break
			}
		}
		if i < 0 {
			return identity.ErrUnknown
		}
		if version != nil && state.Users[i].AccessVersion != *version {
			return identity.ErrInactive
		}
		for n := range state.Users {
			if n != i && email != "" && strings.EqualFold(state.Users[n].Email, identity.NormalizeEmail(email)) {
				return identity.ErrConflict
			}
		}
		if name != "" {
			state.Users[i].Name = strings.TrimSpace(name)

		}
		state.Users[i].Email = identity.NormalizeEmail(email)
		if nationalID != nil {
			state.Users[i].NationalID = *nationalID
		}
		if passwordHash != "" {
			state.Users[i].PasswordHash = passwordHash
			state.Users[i].PasswordChangeRequired = false
			state.Users[i].AccessVersion++
		}
		state.Users[i].Version++
		return nil
	})
}

func (u *Users) RegisterVerifiedCustomer(ctx context.Context, phone, id string) (identity.User, error) {
	err := u.repository.Update(ctx, func(w *backoffice.Workspace) error {
		for _, a := range w.Users {
			if a.Phone == phone {
				if !a.Active {
					return identity.ErrInactive
				}
				return nil
			}
		}
		w.Users = append(w.Users, backoffice.User{ID: id, Name: "مشتری", Phone: phone, Active: true, AccessVersion: 1})
		return nil
	})
	if err != nil {
		return identity.User{}, err
	}
	return u.FindByPhone(ctx, identity.SurfaceCustomer, phone)
}

func (u *Users) snapshot(ctx context.Context) (backoffice.Workspace, error) {
	if r, ok := u.repository.(interface {
		IdentitySnapshot(context.Context) (backoffice.Workspace, error)
	}); ok {
		return r.IdentitySnapshot(ctx)
	}
	return u.repository.Snapshot(ctx)
}

func (u *Users) ChangeTemporaryPassword(ctx context.Context, surface identity.Surface, id string, version int, hash, auditID string, at time.Time) error {
	return u.repository.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Users {
			v := &w.Users[i]
			if v.ID == id {
				if !v.Active || v.AccessVersion != version || !v.PasswordChangeRequired {
					return identity.ErrInactive
				}
				v.PasswordHash = hash
				v.PasswordChangeRequired = false
				w.Audit = append(w.Audit, backoffice.Audit{ID: auditID, Actor: id, Target: id, Action: "users.password.complete", At: at})
				v.AccessVersion++
				v.Version++
				return nil
			}
		}
		return identity.ErrUnknown
	})
}
func (u *Users) SetTemporaryPassword(ctx context.Context, surface identity.Surface, id, hash string, actor identity.User, auditID string, at time.Time) error {
	return u.repository.Update(ctx, func(w *backoffice.Workspace) error {
		authorized := false
		for _, a := range w.Users {
			authorized = authorized || (a.ID == actor.ID && a.Active && a.Role == identity.RoleSuperAdmin && a.AccessVersion == actor.AccessVersion)
		}
		if !authorized {
			return identity.ErrNotAdmin
		}
		for i := range w.Users {
			v := &w.Users[i]
			if v.ID == id {
				v.PasswordHash = hash
				v.PasswordChangeRequired = true
				w.Audit = append(w.Audit, backoffice.Audit{ID: auditID, Actor: actor.ID, Target: id, Action: "users.password.reset", At: at})
				v.AccessVersion++
				v.Version++
				return nil
			}
		}
		return identity.ErrUnknown
	})
}
