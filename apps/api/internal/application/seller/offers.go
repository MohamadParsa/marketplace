package seller

import (
	"context"
	"slices"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type Offers struct{ repository backoffice.Repository }

func NewOffers(repository backoffice.Repository) *Offers {
	return &Offers{repository: repository}
}

func (s *Offers) List(ctx context.Context, userID, storeID string) ([]catalog.Offer, error) {
	state, err := s.repository.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	if !allows(state, userID, storeID, identity.StoreOffersRead) {
		return nil, backoffice.ErrForbidden
	}
	offers := make([]catalog.Offer, 0)
	for _, offer := range state.Offers {
		if offer.StoreID == storeID {
			offers = append(offers, offer)
		}
	}
	return offers, nil
}

func (s *Offers) SetStock(ctx context.Context, userID, storeID, offerID string, stock, version int) error {
	return s.repository.Update(ctx, func(state *backoffice.Workspace) error {
		if !allows(*state, userID, storeID, identity.StoreStockWrite) {
			return backoffice.ErrForbidden
		}
		if stock < 0 || version < 0 {
			return backoffice.ErrInvalid
		}
		for i := range state.Offers {
			offer := &state.Offers[i]
			if offer.ID != offerID || offer.StoreID != storeID {
				continue
			}
			if offer.Version != version {
				return backoffice.ErrConflict
			}
			if offer.Stock != stock {
				offer.Stock = stock
				offer.Version++
			}
			return nil
		}
		return backoffice.ErrNotFound
	})
}

func allows(state backoffice.Workspace, userID, storeID string, permission identity.StorePermission) bool {
	active := slices.ContainsFunc(state.Users, func(user backoffice.User) bool {
		return user.ID == userID && user.Active
	})
	return active && state.Access != nil && identity.AllowsStore(state.StoreMemberships, state.Access.Roles, state.Access.Grants, userID, storeID, permission)
}
