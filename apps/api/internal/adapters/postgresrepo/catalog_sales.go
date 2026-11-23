package postgresrepo

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/jackc/pgx/v5"
)

func saveCatalogSales(ctx context.Context, tx pgx.Tx, w backoffice.Workspace) error {
	b, e := json.Marshal(backoffice.BuildCatalogSales(w))
	if e != nil {
		return e
	}
	_, e = tx.Exec(ctx, `INSERT INTO catalog_sales_state(singleton,fingerprint,payload) VALUES(true,$1,$2::jsonb)
 ON CONFLICT(singleton) DO UPDATE SET fingerprint=excluded.fingerprint,payload=excluded.payload,revision=catalog_sales_state.revision+1
 WHERE catalog_sales_state.fingerprint<>excluded.fingerprint`, catalogHash(b), b)
	return e
}
func (r *Repository) ensureCatalogSales(ctx context.Context) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		w, e := load(ctx, tx, true)
		if e != nil {
			return e
		}
		return saveCatalogSales(ctx, tx, w)
	})
}
func (r *Repository) CatalogSalesRevision(ctx context.Context) (int64, error) {
	var n int64
	e := r.db.QueryRow(ctx, "SELECT revision FROM catalog_sales_state WHERE singleton").Scan(&n)
	return n, e
}
func (r *Repository) CatalogSales(ctx context.Context) (backoffice.CatalogSales, error) {
	var b []byte
	var sales backoffice.CatalogSales
	e := r.db.QueryRow(ctx, "SELECT payload FROM catalog_sales_state WHERE singleton").Scan(&b)
	if e == nil {
		e = json.Unmarshal(b, &sales)
	}
	return sales, e
}
