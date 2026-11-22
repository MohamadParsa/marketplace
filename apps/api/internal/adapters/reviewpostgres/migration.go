package reviewpostgres

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/jackc/pgx/v5"
	"io/fs"
)

func Migrate(ctx context.Context, db interface {
	Begin(context.Context) (pgx.Tx, error)
}, files fs.FS) error {
	return postgresrepo.Migrate(ctx, db, files)
}
