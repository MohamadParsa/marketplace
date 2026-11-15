package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
)

type Service struct {
	repository  Repository
	clock       Clock
	identifiers Identifier
}

func NewService(repository Repository, clock Clock, identifiers Identifier) *Service {
	return &Service{repository: repository, clock: clock, identifiers: identifiers}
}
func roleAllows(role identity.Role, action string) bool {
	if role == identity.RoleSuperAdmin {
		return true
	}
	switch action {
	case "store.toggle", "store.create", "category.create", "category.archive", "product.publish", "product.reject", "shipping.update":
		return role == identity.RoleOperations
	case "case.start", "case.close", "case.refund", "case.assign", "order.deliver":
		return role == identity.RoleSupport
	case "transfer.resolve":
		return role == identity.RoleFinance
	}
	return false
}
func (s *Service) Read(ctx context.Context, actor identity.User) (Workspace, error) {
	state, err := s.repository.Snapshot(ctx)
	if err != nil {
		return Workspace{}, err
	}
	actor, err = currentActor(state, actor.ID)
	if err != nil {
		return Workspace{}, err
	}
	if actor.Role != identity.RoleSuperAdmin {
		state.Users = nil
	}
	if actor.Role != identity.RoleSuperAdmin && actor.Role != identity.RoleFinance {
		state.Transfers = nil
		state.Ledger = nil
		state.Rules = nil
	}
	return state, nil
}
func (s *Service) Execute(ctx context.Context, actor identity.User, command Command) error {
	if !roleAllows(actor.Role, command.Action) {
		return ErrForbidden
	}
	if strings.TrimSpace(command.IdempotencyKey) == "" || len(command.IdempotencyKey) > 128 || len(command.Name) > 160 || len(command.Reason) > 2000 {
		return ErrInvalid
	}
	id, err := s.identifiers.Token()
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(command)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(encoded)
	fingerprint := hex.EncodeToString(sum[:])
	return s.repository.Update(ctx, func(state *Workspace) error {
		current, err := currentActor(*state, actor.ID)
		if err != nil {
			return err
		}
		if !roleAllows(current.Role, command.Action) {
			return ErrForbidden
		}
		for _, event := range state.Audit {
			if event.ID == command.IdempotencyKey {
				if event.Actor == actor.ID && event.Action == command.Action && event.Target == command.Target && event.Fingerprint == fingerprint {
					return nil
				}
				return ErrConflict
			}
		}
		if err := s.apply(state, command, id); err != nil {
			return err
		}
		state.Audit = append(state.Audit, Audit{Fingerprint: fingerprint, ID: command.IdempotencyKey, Actor: actor.ID, Action: command.Action, Target: command.Target, Reason: command.Reason, At: s.clock.Now()})
		return nil
	})
}
func (s *Service) apply(state *Workspace, c Command, id string) error {
	switch c.Action {
	case "user.create", "user.toggle", "user.role":
		return applyUser(state, c, id)
	case "store.create", "store.toggle", "category.create", "category.archive", "product.publish", "product.reject", "rule.toggle", "shipping.update":
		return applyCatalog(state, c, id)
	case "case.start", "case.close", "case.refund", "case.assign", "order.deliver":
		return applySupport(state, c, id)
	case "transfer.resolve":
		return applyTransfer(state, c, id)
	case "notification.retry":
		for i := range state.Notifications {
			n := &state.Notifications[i]
			if n.ID != c.Target {
				continue
			}
			if n.Retries >= 3 || n.Status == "success" {
				return ErrConflict
			}
			n.Retries++
			n.Status = "success"
			return nil
		}
		return ErrNotFound
	default:
		return ErrInvalid
	}
}

func currentActor(state Workspace, id string) (identity.User, error) {
	for _, user := range state.Users {
		if user.ID == id && user.Active {
			return identity.User{ID: user.ID, Role: user.Role, Active: true, Surface: identity.SurfaceBackoffice}, nil
		}
	}
	return identity.User{}, ErrForbidden
}
