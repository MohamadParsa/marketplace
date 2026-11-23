package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

const storefrontReviews = "/storefront/products/product/reviews"
const backofficeReviews = "/backoffice/product-reviews"

func reviewRequest(handler http.Handler, token, method, path string, input any) *httptest.ResponseRecorder {
	var body string
	if input != nil {
		raw, _ := json.Marshal(input)
		body = string(raw)
	}
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func ensureReviewStatus(t *testing.T, response *httptest.ResponseRecorder, want int) []byte {
	t.Helper()
	if response.Code != want {
		t.Fatalf("got %d, want %d (%s)", response.Code, want, response.Body)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("review response must not be cached")
	}
	return response.Body.Bytes()
}

func changeReviewFlag(t *testing.T, handler http.Handler, repo backoffice.Repository, token string, enabled bool) {
	t.Helper()
	state, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	p := state.Products[0]
	raw, err := json.Marshal(backoffice.Command{
		Action: "catalog.product.update", Target: p.ID, Version: p.Version,
		IdempotencyKey: fmt.Sprintf("reviews-settings-%d", p.Version),
		Catalog: &backoffice.CatalogInput{Name: p.Name, CategoryID: p.CategoryID, Published: true,
			Content: &catalog.ProductContent{ReviewsEnabled: enabled, QAEnabled: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	requestAs(t, handler, token, http.MethodPost, "/backoffice/commands", string(raw), http.StatusNoContent)
}

func TestProductReviewsHTTP(t *testing.T) {
	testProductReviewsHTTP(t, backofficememory.New(commerceSeed()))
}

func TestProductReviewsPostgresHTTP(t *testing.T) {
	repo, _ := cachePostgres(t)
	if err := repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		*w = commerceSeed()
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	testProductReviewsHTTP(t, repo)
}

func testProductReviewsHTTP(t *testing.T, repo backoffice.Repository) {
	t.Helper()
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	ops := login(t, h, "09000000002")
	seller := loginAt(t, h, "seller", "09120000002")["token"].(string)
	guestReview := backoffice.ProductReviewInput{
		Name: "نام محرمانه", Text: "نظر بدون نام", Rating: 5, Anonymous: true,
		IdempotencyKey: "review-anonymous-guest",
	}
	ensureReviewStatus(t, reviewRequest(h, "", "POST", storefrontReviews, guestReview), 403)
	changeReviewFlag(t, h, repo, admin, true)
	noRating := guestReview
	noRating.Rating = 0
	ensureReviewStatus(t, reviewRequest(h, "", "POST", storefrontReviews, noRating), 400)
	ensureReviewStatus(t, reviewRequest(h, "", "POST", storefrontReviews, guestReview), 204)
	ensureReviewStatus(t, reviewRequest(h, "", "POST", storefrontReviews, guestReview), 204)
	conflicting := guestReview
	conflicting.Rating = 4
	ensureReviewStatus(t, reviewRequest(h, "", "POST", storefrontReviews, conflicting), 409)
	for i := 1; i <= 6; i++ {
		input := backoffice.ProductReviewInput{Name: "مهمان", Text: fmt.Sprintf("نظر آزمایشی %d", i), Rating: i%5 + 1, IdempotencyKey: fmt.Sprintf("review-guest-extra-%d", i)}
		ensureReviewStatus(t, reviewRequest(h, "", "POST", storefrontReviews, input), 204)
	}
	public := func(path string) backoffice.ProductReviewPage[backoffice.PublicProductReview] {
		raw := ensureReviewStatus(t, reviewRequest(h, "", "GET", path, nil), 200)
		if strings.Contains(string(raw), guestReview.Name) || strings.Contains(string(raw), guestReview.IdempotencyKey) {
			t.Fatal("private data in public review")
		}
		return readJSON[backoffice.ProductReviewPage[backoffice.PublicProductReview]](t, raw)
	}
	if initial := public(storefrontReviews); !initial.SubmissionsEnabled || len(initial.Reviews) != 0 {
		t.Fatalf("pending reviews leaked: %+v", initial)
	}
	ensureReviewStatus(t, reviewRequest(h, "", "GET", backofficeReviews, nil), 401)
	ensureReviewStatus(t, reviewRequest(h, seller, "GET", backofficeReviews, nil), 401)
	ensureReviewStatus(t, reviewRequest(h, ops, "GET", backofficeReviews, nil), 403)
	if raw := requestAs(t, h, admin, "GET", "/backoffice/workspace", "", 200); strings.Contains(string(raw), guestReview.Text) {
		t.Fatal("reviews should be read only via the dedicated moderation endpoint")
	}
	raw := ensureReviewStatus(t, reviewRequest(h, admin, "GET", backofficeReviews, nil), 200)
	pending := readJSON[backoffice.ProductReviewPage[backoffice.AdminProductReview]](t, raw)
	if len(pending.Reviews) != 7 {
		t.Fatalf("expected 7 unique reviews, got %d", len(pending.Reviews))
	}
	if private := pending.Reviews[6]; !private.Anonymous || private.Name != guestReview.Name {
		t.Fatal("administrator must see the submitted name alongside the anonymous flag")
	}
	rejected := pending.Reviews[0]
	approve := backoffice.ProductReviewDecision{Action: "approve", Version: rejected.Version, IdempotencyKey: "approve-review-for-test"}
	ensureReviewStatus(t, reviewRequest(h, ops, "POST", backofficeReviews+"/"+rejected.ID+"/decisions", approve), 403)
	reject := backoffice.ProductReviewDecision{Action: "reject", Text: "دلیل داخلی رد", Version: rejected.Version, IdempotencyKey: "reject-review-for-test"}
	ensureReviewStatus(t, reviewRequest(h, admin, "POST", backofficeReviews+"/"+rejected.ID+"/decisions", reject), 204)
	ensureReviewStatus(t, reviewRequest(h, admin, "POST", backofficeReviews+"/"+rejected.ID+"/decisions", approve), 409)
	for _, review := range pending.Reviews[1:] {
		decision := backoffice.ProductReviewDecision{Action: "approve", Version: review.Version, IdempotencyKey: "review-approve-" + review.ID}
		ensureReviewStatus(t, reviewRequest(h, admin, "POST", backofficeReviews+"/"+review.ID+"/decisions", decision), 204)
	}
	first := public(storefrontReviews)
	if len(first.Reviews) != 5 || first.NextCursor == "" || first.Reviews[0].ID == rejected.ID {
		t.Fatalf("invalid first group: %+v", first)
	}
	secondPath := storefrontReviews + "?cursor=" + first.NextCursor
	second := public(secondPath)
	if len(second.Reviews) != 1 || second.NextCursor != "" || second.Reviews[0].Text != guestReview.Text || second.Reviews[0].Name != "" {
		t.Fatalf("invalid second group or anonymous projection: %+v", second)
	}
	changeReviewFlag(t, h, repo, admin, false)
	ensureReviewStatus(t, reviewRequest(h, "", "POST", storefrontReviews, guestReview), 403)
	pdp, _ := pdpRequest(t, h, "/storefront/products/product", 200)
	if !pdp.Content.QAEnabled || pdp.Interactions.QA || pdp.Interactions.Reviews {
		t.Fatal("review toggle changed Q&A or exposed unavailable capabilities")
	}
	reply := backoffice.ProductReviewDecision{Action: "publish_reply", Text: "پاسخ آزمون چریوا", Version: 2, IdempotencyKey: "publication-reply-test"}
	path := backofficeReviews + "/" + second.Reviews[0].ID + "/decisions"
	ensureReviewStatus(t, reviewRequest(h, admin, "POST", path, reply), 204)
	ensureReviewStatus(t, reviewRequest(h, admin, "POST", path, reply), 204)
	disabled := public(secondPath)
	if disabled.SubmissionsEnabled || len(disabled.Reviews) != 1 || disabled.Reviews[0].Reply != reply.Text {
		t.Fatalf("toggle affected approved history or reply: %+v", disabled)
	}
	ensureReviewStatus(t, reviewRequest(h, "", "PATCH", storefrontReviews, guestReview), 405)
	if err := repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		w.Products[0].Status = "draft"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	ensureReviewStatus(t, reviewRequest(h, "", "GET", storefrontReviews, nil), 404)
}

func TestProductReviewsConcurrentSubmissionsAndToggle(t *testing.T) {
	w := commerceSeed()
	w.Products[0].Content = &catalog.ProductContent{ReviewsEnabled: true}
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	input := backoffice.ProductReviewInput{Anonymous: true, Text: "نظر هم‌زمان", Rating: 3, IdempotencyKey: "same-review-concurrency"}
	responses := make(chan int, 8)
	var requests sync.WaitGroup
	for i := 0; i < cap(responses); i++ {
		requests.Add(1)
		go func() {
			defer requests.Done()
			responses <- reviewRequest(h, "", "POST", storefrontReviews, input).Code
		}()
	}
	changeReviewFlag(t, h, repo, admin, false)
	requests.Wait()
	close(responses)
	accepted := false
	for code := range responses {
		switch code {
		case 204:
			accepted = true
		case 403:
		default:
			t.Fatalf("unexpected HTTP status %d", code)
		}
	}
	state, err := repo.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(state.ProductReviews) > 1 || (len(state.ProductReviews) == 1) != accepted {
		t.Fatalf("responses and unique persisted review disagree: accepted %t, saved %d", accepted, len(state.ProductReviews))
	}
	input.IdempotencyKey = "request-after-shutdown"
	ensureReviewStatus(t, reviewRequest(h, "", "POST", storefrontReviews, input), 403)
}
