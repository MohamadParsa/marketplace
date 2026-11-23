package backofficeidentity

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
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
		user := identity.User{ID: account.ID, Name: account.Name, Phone: phone, Email: account.Email, PasswordHash: account.PasswordHash, Surface: surface, Active: account.Active, AccessVersion: account.AccessVersion}
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

func (u *Users) UpdateProfile(ctx context.Context, surface identity.Surface, userID, name, email, passwordHash string) error {
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
		for n := range state.Users {
			if n != i && email != "" && strings.EqualFold(state.Users[n].Email, identity.NormalizeEmail(email)) {
				return identity.ErrConflict
			}
		}
		if name != "" {
			state.Users[i].Name = strings.TrimSpace(name)
		}
		state.Users[i].Email = identity.NormalizeEmail(email)
		if passwordHash != "" {
			state.Users[i].PasswordHash = passwordHash
		}
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
