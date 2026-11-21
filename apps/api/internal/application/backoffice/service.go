package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"slices"
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
	case "product.media.update", "category.publish", "banner.media.set", "homepage.save", "catalog.category.create", "catalog.category.update", "catalog.brand.create", "catalog.brand.update", "catalog.product.create", "catalog.product.update", "product.icon.set", "product.icon.remove", "banner.create", "banner.save", "banner.publish", "banner.media.remove", "category.icon.set", "category.icon.remove", "brand.icon.set", "brand.icon.remove", "attribute.create", "attribute.update", "attribute.delete", "category.bindings.set", "variant.create", "offer.create", "offer.price.update", "offer.stock.update", "store.toggle", "store.create", "category.create", "category.archive", "product.publish", "product.reject", "shipping.update":
		return role == identity.RoleOperations
	case "case.start", "case.close", "case.refund", "case.assign", "order.deliver":
		return role == identity.RoleSupport
	case "transfer.resolve":
		return role == identity.RoleFinance
	}
	return false
}
func allowsAction(actor identity.User, action string) bool {
	return roleAllows(actor.Role, action) || (!IsAccessCommand(action) && slices.Contains(actor.Permissions, action))
}

func (s *Service) Read(ctx context.Context, actor identity.User) (Workspace, error) {
	state, err := s.repository.Snapshot(ctx)
	if err != nil {
		return Workspace{}, err
	}
	if actor.Surface == identity.SurfaceSeller {
		return sellerWorkspace(state, actor), nil
	}
	actor, err = currentActor(state, actor)
	if err != nil {
		return Workspace{}, err
	}
	if actor.Role == "" && !slices.Contains(actor.Permissions, WorkspaceRead) {
		return Workspace{}, ErrForbidden
	}
	if actor.Role != identity.RoleSuperAdmin {
		state.Users = nil
		state.Access = nil
		state.Audit = slices.DeleteFunc(state.Audit, func(event Audit) bool {
			return IsAccessCommand(event.Action)
		})
	} else {
		if state.Access == nil {
			state.Access = &AccessDirectory{}
		}
		state.Access.Registry = permissionRegistry()
		state.Access.Memberships = state.StoreMemberships
	}
	if actor.Role != identity.RoleSuperAdmin && actor.Role != identity.RoleFinance && !slices.Contains(actor.Permissions, FinanceRead) {
		state.Transfers = nil
		state.Ledger = nil
		state.Rules = nil
		state.Audit = slices.DeleteFunc(state.Audit, func(event Audit) bool {
			return event.Action == "transfer.resolve" || event.Action == "rule.toggle"
		})
	}
	state.IconMedia = nil
	state.BannerAssets = nil
	state.Banners = nil
	state.Memberships = nil
	state.IconPolicy = CatalogIconPolicy()
	if len(state.HomeSettings.Rails) != 4 {
		state.HomeSettings = DefaultHomeSettings()
	}
	state.CardPreviews = BuildCardPreviews(state)
	return state, nil
}
func (s *Service) Execute(ctx context.Context, actor identity.User, command Command) error {
	return s.ExecuteWithReauthentication(ctx, actor, command, nil)
}

func (s *Service) ExecuteWithReauthentication(ctx context.Context, actor identity.User, command Command, confirm func() error) error {
	if !isProductAction(command.Action) && !allowsAction(actor, command.Action) {
		return ErrForbidden
	}
	if strings.TrimSpace(command.IdempotencyKey) == "" || len(command.IdempotencyKey) > 128 || len(command.Name) > 160 || len(command.Reason) > 2000 {
		return ErrInvalid
	}
	if (command.Action == "user.create" || command.Action == "user.role") && command.Value == "admin" {
		command.Value = string(identity.RoleSuperAdmin)
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
	execute := func(confirmed bool) error {
		return s.repository.Update(ctx, func(state *Workspace) error {
			current, err := catalogActor(*state, actor)
			if err != nil {
				return err
			}
			if isProductAction(command.Action) {
				if err := authorizeProductCommand(*state, actor, command); err != nil {
					return err
				}
			} else if !allowsAction(current, command.Action) {
				return ErrForbidden
			}
			if command.Action == "catalog.category.update" && command.Catalog != nil {
				for _, cat := range state.Categories {
					if cat.ID == command.Target && cat.Published != command.Catalog.Published && !allowsAction(current, "category.publish") {
						return ErrForbidden
					}
				}
			}
			if command.Action == "catalog.category.create" && command.Catalog != nil && command.Catalog.Published && !allowsAction(current, "category.publish") {
				return ErrForbidden
			}
			if command.Action == "catalog.product.update" && command.Catalog != nil {
				p, err := findProduct(state, command.Target)
				if err != nil {
					return err
				}
				if command.Catalog.Published != (p.Status == "published") && !allowsAction(current, "product.publish") {
					return ErrForbidden
				}
			}
			for _, event := range state.Audit {
				if event.ID == command.IdempotencyKey {
					if event.Actor == actor.ID && event.Fingerprint == fingerprint {
						return nil
					}
					return ErrConflict
				}
			}
			event := Audit{Fingerprint: fingerprint, ID: command.IdempotencyKey, Actor: actor.ID, Action: command.Action, Target: command.Target, Reason: command.Reason}
			if err := s.apply(state, command, id, confirmed, &event); err != nil {
				return err
			}
			event.At = s.clock.Now()
			state.Audit = append(state.Audit, event)
			return nil
		})
	}
	err = execute(false)
	if err != ErrReauthenticationRequired || confirm == nil {
		return err
	}
	if err := confirm(); err != nil {
		return ErrInvalidReauthentication
	}
	return execute(true)
}
func (s *Service) apply(state *Workspace, c Command, id string, confirmed bool, event *Audit) error {
	switch c.Action {
	case "homepage.save", "catalog.category.create", "catalog.category.update", "catalog.brand.create", "catalog.brand.update", "catalog.product.create", "catalog.product.update":
		return applyHomeCatalog(state, c, id, s.clock.Now())
	case "product.media.update":
		return applyProductMedia(state, c)
	case "product.icon.remove":
		return removeIcon(state, c)
	case "banner.create", "banner.save", "banner.publish", "banner.media.remove":
		return applyBanner(state, c, id)
	case "variant.create", "offer.create", "offer.price.update", "offer.stock.update":
		return applyProduct(state, c, id)
	case "attribute.create", "attribute.update", "attribute.delete", "category.bindings.set", "settings.card-limit":
		return applyAttributes(state, c, id)
	case "category.icon.remove", "brand.icon.remove":
		return removeIcon(state, c)

	case "user.create", "user.toggle", "user.role":
		return applyUser(state, c, id, confirmed, event)
	case "role.create", "role.update", "role.delete", "user.role-grant", "user.role-revoke", "membership.add", "membership.remove":
		return applyAccess(state, c, id, event)
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

func currentActor(state Workspace, actor identity.User) (identity.User, error) {
	for _, user := range state.Users {
		if user.ID == actor.ID && user.Active && user.AccessVersion == actor.AccessVersion {
			if authenticated, ok := state.BackofficeIdentity(user); ok {
				return authenticated, nil
			}
		}
	}
	return identity.User{}, ErrForbidden
}
