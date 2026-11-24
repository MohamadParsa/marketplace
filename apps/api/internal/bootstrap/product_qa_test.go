package bootstrap_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

const qaPublicPath = "/storefront/products/product/questions"
const qaAdminPath = "/backoffice/products/product/questions"

func TestProductQAHTTP(t *testing.T) {
	testProductQAHTTP(t, backofficememory.New(qaFixture()))
}

func TestProductQAPostgres(t *testing.T) {
	repository, db := cachePostgres(t)
	if err := repository.Update(context.Background(), func(state *backoffice.Workspace) error {
		*state = qaFixture()
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	testProductQAHTTP(t, repository)
	reopened, err := reviewpostgres.New(context.Background(), db, backoffice.EmptyWorkspace())
	if err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	public := qaPublicPage(t, h, qaPublicPath, 200)
	if len(public.Questions) != 1 || len(public.Questions[0].Answers) != 1 || public.CanAsk {
		t.Fatalf("approved history or settings did not survive restart: %+v", public)
	}
}

func qaFixture() backoffice.Workspace {
	state := commerceSeed()
	state.Products = append(state.Products, catalog.Product{ID: "other", Name: "محصول دیگر", CategoryID: "regular", Status: "published", Version: 1})
	return state
}

func qaPublicPage(t *testing.T, h http.Handler, path string, status int) backoffice.PublicProductQAPage {
	t.Helper()
	res := cacheRequest(h, path, "")
	if res.Code != status || res.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("GET %s: %d %q %s", path, res.Code, res.Header().Get("Cache-Control"), res.Body)
	}
	if status != 200 {
		return backoffice.PublicProductQAPage{}
	}
	return readJSON[backoffice.PublicProductQAPage](t, res.Body.Bytes())
}

func testProductQAHTTP(t *testing.T, repo backoffice.Repository) {
	t.Helper()
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	operator := login(t, h, "09000000002")
	seller := loginAt(t, h, "seller", "09120000002")["token"].(string)
	customer := loginAt(t, h, "storefront", "09000000005")["token"].(string)
	ask := map[string]string{"text": "سؤال مشتری بدون خرید"}
	post := func(token, path, key string, body any, want int) {
		t.Helper()
		commerceHTTP(t, h, token, "", "POST", path, body, key, want)
	}
	list := func(token string, status int) backoffice.AdminProductQAPage {
		t.Helper()
		data := commerceHTTP(t, h, token, "", "GET", qaAdminPath, nil, "", status)
		if status != 200 {
			return backoffice.AdminProductQAPage{}
		}
		return readJSON[backoffice.AdminProductQAPage](t, data)
	}
	setQA := func(enabled bool) {
		t.Helper()
		if err := repo.Update(context.Background(), func(state *backoffice.Workspace) error {
			state.Products[0].Content = &catalog.ProductContent{ReviewsEnabled: true, QAEnabled: enabled}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}

	post(customer, qaPublicPath, "ask-qa", ask, 403)
	setQA(true)
	post("", qaPublicPath, "ask-qa", ask, 401)
	post(seller, qaPublicPath, "ask-qa", ask, 401)
	post(customer, qaPublicPath, "role-fake", map[string]string{"text": "fake", "role": "super-admin"}, 400)
	post(customer, qaPublicPath, "", ask, 400)
	post(customer, qaPublicPath, "ask-qa", ask, 204)
	post(customer, qaPublicPath, "ask-qa", ask, 204)
	post(customer, qaPublicPath, "ask-qa", map[string]string{"text": "different"}, 409)
	page := qaPublicPage(t, h, qaPublicPath, 200)
	if !page.CanAsk || len(page.Questions) != 0 {
		t.Fatalf("pending question leaked or submit disabled: %+v", page)
	}
	list(customer, 401)
	list(seller, 401)
	list(operator, 403)
	adminPage := list(admin, 200)
	if !adminPage.CanReply || len(adminPage.Questions) != 1 {
		t.Fatalf("duplicate submission or missing admin action: %+v", adminPage)
	}
	question := adminPage.Questions[0]
	reply := backoffice.ProductQAMutation{Action: "reply", Text: "پاسخ رسمی آزمایشی", Version: 1}
	moderate := func(version int, status catalog.QAStatus, answer string) backoffice.ProductQAMutation {
		return backoffice.ProductQAMutation{Action: "moderate", Version: version, Status: status, AnswerID: answer}
	}
	path := qaAdminPath + "/" + question.ID
	post(customer, path, "customer-reply", reply, 401)
	post(seller, path, "seller-reply", reply, 401)
	post(operator, path, "operator-moderate", moderate(1, catalog.QAApproved, ""), 403)
	post(admin, "/backoffice/products/other/questions/"+question.ID, "wrong-product", reply, 404)
	post(admin, path, "approve-question", moderate(1, catalog.QAApproved, ""), 204)

	setQA(false)
	post(customer, qaPublicPath, "disabled", ask, 403)
	post(customer, qaPublicPath, "ask-qa", ask, 403)
	post(admin, path, "stale-reply", reply, 409)
	reply.Version = 2
	post(admin, path, "staff-reply", reply, 204)
	post(admin, path, "staff-reply", reply, 204)
	page = qaPublicPage(t, h, qaPublicPath, 200)
	if page.CanAsk || len(page.Questions) != 1 || len(page.Questions[0].Answers) != 0 {
		t.Fatalf("toggle hid approved history or exposed pending reply: %+v", page)
	}
	detail := cacheRequest(h, "/storefront/products/product", "")
	if detail.Code != 200 || strings.Contains(detail.Body.String(), ask["text"]) || strings.Contains(detail.Body.String(), reply.Text) {
		t.Fatalf("Q&A content entered the product metadata payload: %d %s", detail.Code, detail.Body)
	}

	question = list(admin, 200).Questions[0]
	if len(question.Answers) != 1 {
		t.Fatalf("reply replay duplicated answer: %+v", question)
	}
	answer := question.Answers[0].ID
	post(admin, path, "wrong-answer", moderate(3, catalog.QAApproved, "other-answer"), 404)
	post(admin, path, "approve-answer", moderate(3, catalog.QAApproved, answer), 204)
	page = qaPublicPage(t, h, qaPublicPath, 200)
	if len(page.Questions[0].Answers) != 1 || page.Questions[0].Answers[0].Text != reply.Text {
		t.Fatalf("approved reply not available: %+v", page)
	}
	post(admin, path, "reject-question", moderate(4, catalog.QARejected, ""), 204)
	if page = qaPublicPage(t, h, qaPublicPath, 200); len(page.Questions) != 0 {
		t.Fatalf("rejected question and its approved reply are public: %+v", page)
	}
	post(admin, path, "restore-question", moderate(5, catalog.QAApproved, ""), 204)
	if err := repo.Update(context.Background(), func(state *backoffice.Workspace) error {
		state.Categories[0].ParentID = "gold"
		state.Categories[1].Published = false
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	qaPublicPage(t, h, qaPublicPath, 404)
	if list(admin, 200).CanReply {
		t.Fatal("reply action should follow effective product visibility")
	}
	post(admin, path, "hidden", backoffice.ProductQAMutation{Action: "reply", Text: "invalid", Version: 6}, 404)
	if err := repo.Update(context.Background(), func(state *backoffice.Workspace) error {
		state.Categories[1].Published = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestProductQAPaginationFiltersBeforeLimiting(t *testing.T) {
	t.Run("memory", func(t *testing.T) {
		w := qaFixture()
		w.Products[0].Content = &catalog.ProductContent{QAEnabled: true}
		testProductQAPagination(t, backofficememory.New(w))
	})
	t.Run("postgres", func(t *testing.T) {
		repo, _ := cachePostgres(t)
		if err := repo.Update(context.Background(), func(w *backoffice.Workspace) error {
			*w = qaFixture()
			w.Products[0].Content = &catalog.ProductContent{QAEnabled: true}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		testProductQAPagination(t, repo)
	})
}

func testProductQAPagination(t *testing.T, repo backoffice.Repository) {
	t.Helper()
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	customer := loginAt(t, h, "storefront", "09000000005")["token"].(string)
	admin := login(t, h, "09000000001")
	for i := range 23 {
		text := fmt.Sprintf("پرسش شماره %d", i)
		commerceHTTP(t, h, customer, "", "POST", qaPublicPath, map[string]string{"text": text}, fmt.Sprintf("ask-%d", i), 204)
	}
	readAdmin := func(after int64) backoffice.AdminProductQAPage {
		return readJSON[backoffice.AdminProductQAPage](t, commerceHTTP(t, h, admin, "", "GET", fmt.Sprintf("%s?after=%d", qaAdminPath, after), nil, "", 200))
	}
	first := readJSON[backoffice.AdminProductQAPage](t, commerceHTTP(t, h, admin, "", "GET", qaAdminPath, nil, "", 200))
	if first.NextCursor == 0 || len(first.Questions) != 20 {
		t.Fatalf("wrong admin page: %+v", first)
	}
	next := readAdmin(first.NextCursor)
	if len(next.Questions) != 3 || next.NextCursor != 0 {
		t.Fatalf("wrong admin next page: %+v", next)
	}
	all := append(first.Questions, next.Questions...)
	for _, question := range all[:22] {
		commerceHTTP(t, h, admin, "", "POST", qaAdminPath+"/"+question.ID, backoffice.ProductQAMutation{Action: "moderate", Status: catalog.QAApproved, Version: 1}, "publish-"+question.ID, 204)
	}
	page := qaPublicPage(t, h, qaPublicPath, 200)
	if page.NextCursor == 0 || len(page.Questions) != 20 || page.Questions[0].Text != "پرسش شماره 1" {
		t.Fatalf("filtering or ordering happens after page limit: %+v", page)
	}
	last := qaPublicPage(t, h, fmt.Sprintf("%s?after=%d", qaPublicPath, page.NextCursor), 200)
	if len(last.Questions) != 2 || last.NextCursor != 0 || last.Questions[1].Text != "پرسش شماره 22" {
		t.Fatalf("wrong public next page: %+v", last)
	}
}
