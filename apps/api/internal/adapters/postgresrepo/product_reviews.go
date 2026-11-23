package postgresrepo

import (
	"context"
	"encoding/json"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
)

// Use the same statement snapshot for catalog visibility and reviews. Filtering
// and pagination happen in PostgreSQL, and only public fields leave the query.
func (r *Repository) ReadPublicReviewPage(ctx context.Context, productID, cursor string, limit int) (backoffice.Workspace, []backoffice.PublicProductReview, error) {
	var catalogJSON, reviewsJSON []byte
	err := r.db.QueryRow(ctx, `
		WITH review_rows AS MATERIALIZED (
			SELECT review, position
			FROM review_workspace w
			CROSS JOIN LATERAL jsonb_array_elements(w.payload->'product_reviews')
				WITH ORDINALITY AS rows(review, position)
			WHERE w.singleton AND review->>'ProductID' = $1
		), review_page AS (
			SELECT position, jsonb_strip_nulls(jsonb_build_object(
				'id', review->'ID',
				'name', CASE WHEN review->>'Anonymous' = 'false' THEN review->'Name' END,
				'text', review->'Text',
				'rating', review->'Rating',
				'created_at', review->'CreatedAt',
				'reply', NULLIF(review->>'Reply', '')
			)) AS public_review
			FROM review_rows
			WHERE review->>'Status' = 'approved'
			AND ($2 = '' OR position < (SELECT position FROM review_rows WHERE review->>'ID' = $2))
			ORDER BY position DESC
			LIMIT $3
		)
		SELECT jsonb_build_object('categories', s.payload->'categories', 'brands', s.payload->'brands', 'products', p.payload->'products'),
			COALESCE((SELECT jsonb_agg(public_review ORDER BY position DESC) FROM review_page), '[]'::jsonb)
		FROM catalog_public_state s LEFT JOIN catalog_public_products p ON p.id = $1
		WHERE s.singleton
	`, productID, cursor, limit).Scan(&catalogJSON, &reviewsJSON)
	var w backoffice.Workspace
	var reviews []backoffice.PublicProductReview
	if err != nil {
		return w, nil, err
	}
	if err := json.Unmarshal(catalogJSON, &w); err != nil {
		return w, nil, err
	}
	err = json.Unmarshal(reviewsJSON, &reviews)
	return w, reviews, err
}
