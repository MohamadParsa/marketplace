package bootstrap_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

func TestProductQAGlobalInbox(t *testing.T) {
	for _, kind := range []string{"memory", "postgres"} {
		t.Run(kind, func(t *testing.T) {
			seed := qaFixture()
			for i := range seed.Products {
				seed.Products[i].Content = &catalog.ProductContent{QAEnabled: true}
			}
			var repo backoffice.Repository
			if kind == "postgres" {
				r, _ := cachePostgres(t)
				if err := r.Update(context.Background(), func(w *backoffice.Workspace) error { *w = seed; return nil }); err != nil {
					t.Fatal(err)
				}
				repo = r
			} else {
				repo = backofficememory.New(seed)
			}
			qa := repo.(backoffice.ProductQARepository)
			for i := 0; i < 26; i++ {
				product := "product"
				if i%2 == 1 {
					product = "other"
				}
				status := catalog.QAPending
				if i == 24 {
					status = catalog.QAApproved
				}
				if i == 25 {
					status = catalog.QARejected
				}
				q := catalog.ProductQuestion{ProductQAContent: catalog.ProductQAContent{ID: fmt.Sprintf("global-%d", i), Text: "پرسش ساختگی", Status: status, CreatedAt: time.Now()}, Answers: []catalog.ProductQAContent{}, Version: 1}
				if err := qa.UpdateProductQA(context.Background(), product, q.ID, "seed", q.ID, q.ID, func(backoffice.Workspace, string, *catalog.ProductQuestion) (*catalog.ProductQuestion, error) {
					return &q, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
			admin := login(t, h, "09000000001")
			operator := login(t, h, "09000000002")
			get := func(token, path string, code int) backoffice.ProductQAPage[backoffice.AdminProductQuestion] {
				t.Helper()
				data := commerceHTTP(t, h, token, "", http.MethodGet, path, nil, "", code)
				if code != 200 {
					return backoffice.ProductQAPage[backoffice.AdminProductQuestion]{}
				}
				return readJSON[backoffice.ProductQAPage[backoffice.AdminProductQuestion]](t, data)
			}
			const route = "/backoffice/product-questions"
			get("", route, 401)
			get(operator, route, 403)
			get(admin, route+"?status=invalid", 400)
			get(admin, route+"?after=-1", 400)
			first := get(admin, route+"?status=pending", 200)
			if len(first.Questions) != 20 || first.NextCursor == 0 {
				t.Fatalf("bad first page: %+v", first)
			}
			seen := map[string]bool{}
			products := map[string]bool{}
			for _, q := range first.Questions {
				seen[q.ID] = true
				products[q.ProductID] = true
				if q.ProductName == "" || !q.CanReply || q.Status != catalog.QAPending {
					t.Fatalf("bad projection: %+v", q)
				}
			}
			if len(products) != 2 {
				t.Fatal("global list did not contain both products")
			}
			next := get(admin, fmt.Sprintf("%s?status=pending&after=%d", route, first.NextCursor), 200)
			if len(next.Questions) != 4 || next.NextCursor != 0 {
				t.Fatalf("bad next page: %+v", next)
			}
			for _, q := range next.Questions {
				if seen[q.ID] {
					t.Fatal("duplicate cursor row")
				}
			}
			for _, status := range []string{"approved", "rejected"} {
				p := get(admin, route+"?status="+status, 200)
				if len(p.Questions) != 1 || string(p.Questions[0].Status) != status {
					t.Fatal("status filter applied after limit", p)
				}
			}
			filtered := get(admin, route+"?status=pending&product=other", 200)
			if len(filtered.Questions) != 12 {
				t.Fatal("optional product filter", filtered)
			}
			for _, q := range filtered.Questions {
				if q.ProductID != "other" {
					t.Fatal("wrong product")
				}
			}
			if err := repo.Update(context.Background(), func(w *backoffice.Workspace) error {
				for i := range w.Products {
					if w.Products[i].ID == "other" {
						w.Products[i].Status = "draft"
					}
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			hidden := get(admin, route+"?product=other", 200)
			if len(hidden.Questions) != 13 || hidden.Questions[0].CanReply {
				t.Fatal("hidden history unavailable or reply allowed", hidden)
			}
			qaPublicPage(t, h, "/storefront/products/other/questions", 404)
			public := qaPublicPage(t, h, qaPublicPath, 200)
			if len(public.Questions) != 1 || public.Questions[0].ID != "global-24" {
				t.Fatal("public endpoint leaked private/cross-product questions", public)
			}
		})
	}
}
