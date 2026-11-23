package postgresrepo

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/jackc/pgx/v5"
)

func catalogURLRegistry(ctx context.Context, tx pgx.Tx) (backoffice.CatalogURLRegistry, error) {
	r := backoffice.CatalogURLRegistry{Products: map[string]int64{}, Categories: map[string]string{}}
	rows, e := tx.Query(ctx, "SELECT product_id,number FROM catalog_product_codes")
	if e != nil {
		return r, e
	}
	for rows.Next() {
		var id string
		var n int64
		if e = rows.Scan(&id, &n); e != nil {
			rows.Close()
			return r, e
		}
		r.Products[id] = n
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return r, e
	}
	rows, e = tx.Query(ctx, "SELECT token,category_id FROM catalog_category_urls")
	if e != nil {
		return r, e
	}
	for rows.Next() {
		var token, id string
		if e = rows.Scan(&token, &id); e != nil {
			rows.Close()
			return r, e
		}
		r.Categories[token] = id
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return r, e
	}
	e = tx.QueryRow(ctx, "SELECT last_value FROM catalog_product_code_seq").Scan(&r.LastNumber)
	return r, e
}
func syncCatalogURLs(ctx context.Context, tx pgx.Tx, w *backoffice.Workspace) (bool, error) {
	registry, e := catalogURLRegistry(ctx, tx)
	if e != nil {
		return false, e
	}
	plan, e := backoffice.PlanCatalogURLs(*w, registry)
	if e != nil {
		return false, e
	}
	codes, slugs := map[string]string{}, map[string]string{}
	batch := &pgx.Batch{}
	changed := false
	for _, p := range plan.Products {
		codes[p.ID] = p.Code
		if registry.Products[p.ID] == 0 {
			batch.Queue("INSERT INTO catalog_product_codes(product_id,number) VALUES($1,$2)", p.ID, p.Number)
		}
	}
	for _, c := range plan.Categories {
		slugs[c.ID] = c.Slug
		for _, token := range []string{c.ID, c.Slug} {
			if registry.Categories[token] == "" {
				batch.Queue("INSERT INTO catalog_category_urls(token,category_id) VALUES($1,$2)", token, c.ID)
				registry.Categories[token] = c.ID
			}
		}
	}
	if e = tx.SendBatch(ctx, batch).Close(); e != nil {
		return false, e
	}
	if plan.LastNumber > registry.LastNumber {
		if _, e = tx.Exec(ctx, "SELECT setval('catalog_product_code_seq',$1,true)", plan.LastNumber); e != nil {
			return false, e
		}
	}
	for i := range w.Products {
		p := &w.Products[i]
		if p.PublicCode != codes[p.ID] {
			p.PublicCode = codes[p.ID]
			changed = true
		}
	}
	for i := range w.Categories {
		c := &w.Categories[i]
		if c.Slug != slugs[c.ID] {
			c.Slug = slugs[c.ID]
			changed = true
		}
	}
	return changed, nil
}

// Backfill metadata only. Identity, commerce projections, media and import markers are never replayed.
func (r *Repository) ensureCatalogURLs(ctx context.Context) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		w, e := load(ctx, tx, true)
		if e != nil {
			return e
		}
		changed, e := syncCatalogURLs(ctx, tx, &w)
		if e != nil {
			return e
		}
		if !changed {
			return nil
		}
		var payload []byte
		if e = tx.QueryRow(ctx, "SELECT payload FROM review_workspace WHERE singleton").Scan(&payload); e != nil {
			return e
		}
		var raw map[string]json.RawMessage
		if e = json.Unmarshal(payload, &raw); e != nil {
			return e
		}
		raw["products"], e = json.Marshal(w.Products)
		if e != nil {
			return e
		}
		raw["categories"], e = json.Marshal(w.Categories)
		if e != nil {
			return e
		}
		payload, e = json.Marshal(raw)
		if e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "UPDATE review_workspace SET payload=$1::jsonb WHERE singleton", payload); e != nil {
			return e
		}
		return savePublicCatalog(ctx, tx, w)
	})
}
func (r *Repository) ResolveCatalogToken(ctx context.Context, kind, token string) (string, error) {
	var id string
	var e error
	if kind == "product" {
		e = r.db.QueryRow(ctx, "SELECT product_id FROM catalog_product_codes WHERE product_id=$1 OR public_code=$1", token).Scan(&id)
	} else {
		e = r.db.QueryRow(ctx, "SELECT category_id FROM catalog_category_urls WHERE token=$1", token).Scan(&id)
	}
	if e == pgx.ErrNoRows {
		return "", backoffice.ErrNotFound
	}
	return id, e
}
