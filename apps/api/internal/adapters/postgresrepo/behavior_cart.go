package postgresrepo

import (
	"context"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) FindBehaviorCart(ctx context.Context, hash string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `SELECT cart->>'id' FROM review_workspace, jsonb_array_elements(COALESCE(NULLIF(payload->'commerce'->'carts','null'::jsonb),'[]'::jsonb)) cart WHERE singleton AND cart->>'guest_hash'=$1 AND COALESCE(cart->>'merged_into','')='' LIMIT 1`, hash).Scan(&id)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return id, err
}
