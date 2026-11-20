package backofficeidentity

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type Users struct{ repository backoffice.Repository }

func New(repository backoffice.Repository) *Users { return &Users{repository: repository} }
func (u *Users) FindByPhone(ctx context.Context, surface identity.Surface, phone string) (identity.User, error) {
	state, err := u.repository.Snapshot(ctx)
	if err != nil {
		return identity.User{}, err
	}
	for _, account := range state.Users {
		if account.Phone != phone {
			continue
		}
		user := identity.User{ID: account.ID, Phone: phone, Surface: surface, Active: account.Active}
		switch surface {
		case identity.SurfaceBackoffice:
			if authenticated, ok := state.BackofficeIdentity(account); ok {
				return authenticated, nil
			}
		case identity.SurfaceCustomer:
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
