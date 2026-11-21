package backofficeidentity

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type Users struct{ repository backoffice.Repository }

func New(repository backoffice.Repository) *Users { return &Users{repository: repository} }
func (u *Users) FindByPhone(ctx context.Context, surface identity.Surface, phone string) (identity.User, error) {
	if surface != identity.SurfaceBackoffice && surface != identity.SurfaceSeller {
		return identity.User{}, identity.ErrUnknown
	}
	state, err := u.repository.Snapshot(ctx)
	if err != nil {
		return identity.User{}, err
	}
	for _, user := range state.Users {
		if user.Phone == phone {
			if surface == identity.SurfaceSeller {
				for _, membership := range state.Memberships {
					if membership.UserID == user.ID && membership.Active {
						return identity.User{ID: user.ID, Phone: phone, Surface: surface, Role: user.Role, Active: user.Active}, nil
					}
				}
				return identity.User{}, identity.ErrUnknown
			}
			if user.SellerOnly {
				return identity.User{}, identity.ErrUnknown
			}
			return identity.User{ID: user.ID, Phone: phone, Surface: surface, Role: user.Role, Active: user.Active}, nil
		}
	}
	return identity.User{}, identity.ErrUnknown
}
