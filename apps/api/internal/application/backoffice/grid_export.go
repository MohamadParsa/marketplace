package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"slices"
)

const GridExport = "grid.export"

func (s *Service) AuthorizeGridExport(ctx context.Context, a identity.User) error {
	var w Workspace
	var e error
	if repo, ok := s.repository.(interface {
		IdentitySnapshot(context.Context) (Workspace, error)
	}); ok {
		w, e = repo.IdentitySnapshot(ctx)
	} else {
		w, e = s.repository.Snapshot(ctx)
	}
	if e != nil {
		return e
	}
	if a.Surface == identity.SurfaceSeller {
		if _, err := catalogActor(w, a); err != nil {
			return err
		}
		// Conservative: exporting a mixed store workspace requires this grant in every
		// member store, not just one store; no other store's data is queried here.
		ids := identity.MemberStoreIDs(w.StoreMemberships, a.ID)
		if len(ids) == 0 {
			return ErrForbidden
		}
		for _, id := range ids {
			if w.Access == nil || !slices.Contains(identity.GrantedPermissions(w.Access.Roles, w.Access.Grants, a.ID, id), GridExport) {
				return ErrForbidden
			}
		}
		return nil
	}
	return requireCommerceAdmin(w, a, GridExport)
}
