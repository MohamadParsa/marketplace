package postgresrepo

import (
	"context"
	"encoding/json"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
)

// Use the same statement snapshot for catalog visibility and reviews. Filtering
// and pagination happen in PostgreSQL, and only public fields leave the query.
func (r *Repository) ReadPublicReviewPage(ctx context.Context, productID, cursor string, limit int) (backoffice.Workspace, []backoffice.PublicProductReview, backoffice.ReviewRatingTotals, error) {
	var catalogJSON, reviewsJSON []byte
	var totals backoffice.ReviewRatingTotals
	err := r.db.QueryRow(ctx, `
		WITH review_rows AS MATERIALIZED (
			SELECT review, position
			FROM review_workspace w
			CROSS JOIN LATERAL jsonb_array_elements(w.payload->'product_reviews')
				WITH ORDINALITY AS rows(review, position)
			WHERE w.singleton AND review->>'ProductID' = $1
        ), approved_ratings AS (
            SELECT CASE WHEN jsonb_typeof(review->'Rating')='number' THEN (review->>'Rating')::numeric END AS rating
            FROM review_rows WHERE review->>'Status'='approved'
        ), valid_ratings AS (
            SELECT rating::bigint AS rating FROM approved_ratings
            WHERE rating BETWEEN 1 AND 5 AND rating=trunc(rating)
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
			COALESCE((SELECT jsonb_agg(public_review ORDER BY position DESC) FROM review_page), '[]'::jsonb),
            (SELECT count(*) FROM valid_ratings), COALESCE((SELECT sum(rating) FROM valid_ratings),0)
		FROM catalog_public_state s LEFT JOIN catalog_public_products p ON p.id = $1
		WHERE s.singleton
	`, productID, cursor, limit).Scan(&catalogJSON, &reviewsJSON, &totals.Count, &totals.Sum)
	var w backoffice.Workspace
	var reviews []backoffice.PublicProductReview
	if err != nil {
		return w, nil, totals, err
	}
	if err := json.Unmarshal(catalogJSON, &w); err != nil {
		return w, nil, totals, err
	}
	err = json.Unmarshal(reviewsJSON, &reviews)
	return w, reviews, totals, err
}
