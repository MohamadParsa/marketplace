package reviewpostgres

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Historical review API; live initialization never uses the review fixture.
type Repository = postgresrepo.Repository

func New(ctx context.Context, db *pgxpool.Pool, seed backoffice.Workspace) (*Repository, error) {
	return postgresrepo.NewReview(ctx, db, seed)
}
