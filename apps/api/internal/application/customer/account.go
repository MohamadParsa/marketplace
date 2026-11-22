package customer

import (
	"context"
	"slices"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
)

type Accounts struct{ repository backoffice.Repository }

func NewAccounts(repository backoffice.Repository) *Accounts {
	return &Accounts{repository: repository}
}

type Profile struct {
	UserID            string `json:"user_id"`
	Name              string `json:"name"`
	Phone             string `json:"phone"`
	Email             string `json:"email,omitempty"`
	PasswordAvailable bool   `json:"password_available"`
}
type Purchase struct {
	ID     string         `json:"id"`
	Orders []orders.Order `json:"orders"`
}

func (s *Accounts) Profile(ctx context.Context, userID string) (Profile, error) {
	_, user, err := s.account(ctx, userID)
	if err != nil {
		return Profile{}, err
	}
	return Profile{UserID: user.ID, Name: user.Name, Phone: user.Phone, Email: user.Email, PasswordAvailable: user.PasswordHash != ""}, nil
}

func (s *Accounts) Purchases(ctx context.Context, userID string) ([]Purchase, error) {
	state, _, err := s.account(ctx, userID)
	if err != nil {
		return nil, err
	}
	purchases := make([]Purchase, 0)
	for _, owner := range state.Purchases {
		if owner.UserID != userID {
			continue
		}
		purchase := Purchase{ID: owner.PurchaseID, Orders: make([]orders.Order, 0)}
		for _, order := range state.Orders {
			if order.PurchaseID == owner.PurchaseID {
				purchase.Orders = append(purchase.Orders, order)
			}
		}
		purchases = append(purchases, purchase)
	}
	return purchases, nil
}

func (s *Accounts) account(ctx context.Context, userID string) (backoffice.Workspace, backoffice.User, error) {
	state, err := s.repository.Snapshot(ctx)
	if err != nil {
		return backoffice.Workspace{}, backoffice.User{}, err
	}
	i := slices.IndexFunc(state.Users, func(user backoffice.User) bool { return user.ID == userID && user.Active })
	if i < 0 {
		return backoffice.Workspace{}, backoffice.User{}, backoffice.ErrForbidden
	}
	return state, state.Users[i], nil
}
