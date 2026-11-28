package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"reflect"
	"testing"
	"time"
)

func ratingSeed(t *testing.T) backoffice.Workspace {
	w := commerceSeed()
	w.IconMedia[0].Data = iconPNG(t, 8)
	w.Products[0].Content = &catalog.ProductContent{ReviewsEnabled: true}
	w.Products[0].Name = "محصول امتیازدار آزمون"
	w.Products = append(w.Products, catalog.Product{ID: "other-product", Name: "محصول دیگر", CategoryID: "regular", Status: "published", ImageID: "photo", Version: 1})
	return w
}
func seededReview(id, product, status string, rating int) backoffice.ProductReview {
	return backoffice.ProductReview{ID: id, ProductID: product, Name: "نام خصوصی ساختگی", Text: "نظر ساختگی " + id, Rating: rating, Status: status, Anonymous: true, Version: 1, CreatedAt: time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC)}
}
func TestReviewRatingSummaryMemoryAndPostgres(t *testing.T) {
	for _, backend := range []string{"memory", "postgres"} {
		t.Run(backend, func(t *testing.T) {
			ctx := context.Background()
			var repo backoffice.Repository = backofficememory.New(ratingSeed(t))
			var counter *countingPublicRepo
			if backend == "postgres" {
				counter, _ = cachePostgres(t)
				if e := counter.Update(ctx, func(w *backoffice.Workspace) error { *w = ratingSeed(t); return nil }); e != nil {
					t.Fatal(e)
				}
				repo = counter
			}
			h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
			admin := login(t, h, "09000000001")
			read := func(cursor string) backoffice.ProductReviewPage[backoffice.PublicProductReview] {
				path := storefrontReviews
				if cursor != "" {
					path += "?cursor=" + cursor
				}
				return readJSON[backoffice.ProductReviewPage[backoffice.PublicProductReview]](t, ensureReviewStatus(t, reviewRequest(h, "", "GET", path, nil), 200))
			}
			page := read("")
			if page.RatingSummary == nil || page.RatingSummary.Count != 0 || page.RatingSummary.Average != nil {
				t.Fatal("fabricated zero average")
			}
			repo.Update(ctx, func(w *backoffice.Workspace) error {
				w.ProductReviews = []backoffice.ProductReview{seededReview("other", "other-product", "approved", 5), seededReview("pending", "product", "pending", 5), seededReview("rejected", "product", "rejected", 1)}
				return nil
			})
			page = read("")
			if page.RatingSummary.Count != 0 || page.RatingSummary.Average != nil {
				t.Fatal("unapproved or other product included")
			}
			// The existing moderation operation is the only mutation path used here.
			decision := backoffice.ProductReviewDecision{Action: "approve", Version: 1, IdempotencyKey: "rating-approve-review-test"}
			ensureReviewStatus(t, reviewRequest(h, admin, "POST", backofficeReviews+"/pending/decisions", decision), 204)
			page = read("")
			if page.RatingSummary.Count != 1 || *page.RatingSummary.Average != 5 {
				t.Fatal("approval not fresh")
			}
			repo.Update(ctx, func(w *backoffice.Workspace) error {
				w.ProductReviews = append(w.ProductReviews, seededReview("three", "product", "approved", 3))
				return nil
			})
			page = read("")
			if page.RatingSummary.Count != 2 || *page.RatingSummary.Average != 4 {
				t.Fatal("5 and3 must average4")
			}
			repo.Update(ctx, func(w *backoffice.Workspace) error {
				for i := 0; i < 5; i++ {
					w.ProductReviews = append(w.ProductReviews, seededReview(fmt.Sprintf("page-%d", i), "product", "approved", 4))
				}
				w.ProductReviews = append(w.ProductReviews, seededReview("invalid-legacy", "product", "approved", 0))
				return nil
			})
			page = read("")
			next := read(page.NextCursor)
			if len(page.Reviews) != 5 || page.NextCursor == "" || page.RatingSummary.Count != 7 || *page.RatingSummary.Average != 4 || !reflect.DeepEqual(page.RatingSummary, next.RatingSummary) {
				t.Fatal("page dependent average", page.RatingSummary, next.RatingSummary)
			}
			// Changes in stored public eligibility cannot leave a separate stale aggregate.
			repo.Update(ctx, func(w *backoffice.Workspace) error {
				for i := range w.ProductReviews {
					if w.ProductReviews[i].ID == "pending" {
						w.ProductReviews[i].Status = "rejected"
					}
				}
				return nil
			})
			page = read("")
			if page.RatingSummary.Count != 6 || *page.RatingSummary.Average != 3.8 {
				t.Fatal("stale status", page.RatingSummary)
			}
			if counter != nil {
				n := counter.privateReads.Load()
				read("")
				if counter.privateReads.Load() != n {
					t.Fatal("ratings loaded private workspace")
				}
			}
			before, _ := repo.Snapshot(ctx)
			for i := 0; i < 3; i++ {
				read("")
			}
			after, _ := repo.Snapshot(ctx)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("reading ratings mutated state")
			}
			// The public wire result is bounded and carries no moderation/private names.
			raw := ensureReviewStatus(t, reviewRequest(h, "", "GET", storefrontReviews, nil), 200)
			var wire map[string]any
			json.Unmarshal(raw, &wire)
			if wire["rating_summary"] == nil {
				t.Fatal("summary omitted")
			}
		})
	}
}
func TestReviewRatingRoundingAndRestart(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := ratingSeed(t)
	seed.ProductReviews = []backoffice.ProductReview{seededReview("a", "product", "approved", 4), seededReview("b", "product", "approved", 4), seededReview("c", "product", "approved", 5)}
	repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil })
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	raw := ensureReviewStatus(t, reviewRequest(h, "", "GET", storefrontReviews, nil), 200)
	before := readJSON[backoffice.ProductReviewPage[backoffice.PublicProductReview]](t, raw)
	if *before.RatingSummary.Average != 4.3 {
		t.Fatal("one decimal rounding")
	}
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	after := readJSON[backoffice.ProductReviewPage[backoffice.PublicProductReview]](t, ensureReviewStatus(t, reviewRequest(h2, "", "GET", storefrontReviews, nil), 200))
	if !reflect.DeepEqual(before.RatingSummary, after.RatingSummary) {
		t.Fatal("restart summary")
	}
	repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.ProductReviews = []backoffice.ProductReview{seededReview("a", "product", "approved", 1), seededReview("b", "product", "approved", 2), seededReview("c", "product", "approved", 3), seededReview("d", "product", "approved", 3)}
		return nil
	})
	page := readJSON[backoffice.ProductReviewPage[backoffice.PublicProductReview]](t, ensureReviewStatus(t, reviewRequest(h2, "", "GET", storefrontReviews, nil), 200))
	if *page.RatingSummary.Average != 2.3 {
		t.Fatal("half up rounding")
	}
}
