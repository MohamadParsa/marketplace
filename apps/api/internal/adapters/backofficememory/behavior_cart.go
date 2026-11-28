package backofficememory

import "context"

func (r *Repository) FindBehaviorCart(ctx context.Context, hash string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state.Commerce != nil {
		for _, cart := range r.state.Commerce.Carts {
			if cart.GuestHash == hash && cart.MergedInto == "" {
				return cart.ID, nil
			}
		}
	}
	return "", nil
}
