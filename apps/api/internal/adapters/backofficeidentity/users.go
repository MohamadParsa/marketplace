package backofficeidentity

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type Users struct{ repository backoffice.Repository }

func New(repository backoffice.Repository) *Users { return &Users{repository: repository} }
func (u *Users) FindByPhone(ctx context.Context, surface identity.Surface, phone string) (identity.User, error) {
	if surface == identity.SurfaceSeller && phone == "09120000002" {
		return identity.User{ID: "seller-test", Phone: phone, Surface: surface, StoreID: "s1", Active: true}, nil
	}
	if surface != identity.SurfaceBackoffice {
		return identity.User{}, identity.ErrUnknown
	}
	state, err := u.repository.Snapshot(ctx)
	if err != nil {
		return identity.User{}, err
	}
	for _, user := range state.Users {
		if user.Phone == phone {
			return identity.User{ID: user.ID, Phone: phone, Surface: surface, Role: user.Role, Active: user.Active}, nil
		}
	}
	return identity.User{}, identity.ErrUnknown
}
