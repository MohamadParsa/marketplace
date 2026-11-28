package backofficememory

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"slices"
	"sync"
)

// Repository contains synthetic staging data only. A restart restores the fixture.
type Repository struct {
	mu          sync.Mutex
	state       backoffice.Workspace
	qaQuestions map[string][]catalog.ProductQuestion
	qaReceipts  map[qaReceiptKey]string
	qaCursor    int64
}

func New(state backoffice.Workspace) *Repository {
	backoffice.EnsureRecordNumbers(&state)
	return &Repository{state: state, qaQuestions: map[string][]catalog.ProductQuestion{}, qaReceipts: map[qaReceiptKey]string{}}
}
func clone(state backoffice.Workspace) (backoffice.Workspace, error) {
	encoded, err := json.Marshal(state)
	if err != nil {
		return backoffice.Workspace{}, err
	}
	var copy backoffice.Workspace
	err = json.Unmarshal(encoded, &copy)
	copy.Users = slices.Clone(state.Users)
	copy.StoreMemberships = slices.Clone(state.StoreMemberships)
	copy.Purchases = slices.Clone(state.Purchases)
	return copy, err
}
func (r *Repository) Snapshot(ctx context.Context) (backoffice.Workspace, error) {
	if err := ctx.Err(); err != nil {
		return backoffice.Workspace{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return clone(r.state)
}
func (r *Repository) Update(ctx context.Context, apply func(*backoffice.Workspace) error) error {
	return backoffice.ObserveWorkspaceUpdate(ctx, func(fn func(*backoffice.Workspace) error) error { return r.update(ctx, fn) }, apply)
}
func (r *Repository) update(ctx context.Context, apply func(*backoffice.Workspace) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	next, err := clone(r.state)
	if err != nil {
		return err
	}
	if err := apply(&next); err != nil {
		return err
	}
	backoffice.EnsureRecordNumbers(&next)
	r.state = next
	return nil
}
