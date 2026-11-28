package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const productQAPageSize = 20

type ProductQARepository interface {
	ListProductQA(context.Context, string, int64, int, bool) (Workspace, []catalog.ProductQuestion, error)
	ListAdminProductQA(context.Context, string, catalog.QAStatus, int64, int) (Workspace, []AdminProductQuestion, error)
	// The callback reads current catalog/identity state and Q&A under the same
	// write lock as catalog settings. A nil question means an idempotent replay.
	UpdateProductQA(context.Context, string, string, string, string, string, func(Workspace, string, *catalog.ProductQuestion) (*catalog.ProductQuestion, error)) error
}

type PublicQAContent struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type PublicQuestion struct {
	PublicQAContent
	Answers []PublicQAContent `json:"answers"`
}

type ProductQAPage[T any] struct {
	Questions  []T   `json:"questions"`
	NextCursor int64 `json:"next_cursor,omitempty"`
}

type PublicProductQAPage struct {
	ProductQAPage[PublicQuestion]
	CanAsk bool `json:"can_ask"`
}

type AdminProductQAPage struct {
	ProductQAPage[catalog.ProductQuestion]
	CanReply bool `json:"can_reply"`
}

type AdminProductQuestion struct {
	catalog.ProductQuestion
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	CanReply    bool   `json:"can_reply"`
}

func (s *Service) ListAdminProductQuestions(ctx context.Context, actor identity.User, product string, status catalog.QAStatus, after int64) (ProductQAPage[AdminProductQuestion], error) {
	result := ProductQAPage[AdminProductQuestion]{Questions: []AdminProductQuestion{}}
	if !qaAdmin(actor) {
		return result, ErrForbidden
	}
	if status != "" && status != catalog.QAPending && status != catalog.QAApproved && status != catalog.QARejected {
		return result, ErrInvalid
	}
	if len(product) > 128 || after < 0 {
		return result, ErrInvalid
	}
	store := s.qaRepository()
	if store == nil {
		return result, ErrNotFound
	}
	w, rows, err := store.ListAdminProductQA(ctx, product, status, after, productQAPageSize+1)
	if err != nil {
		return result, err
	}
	if len(rows) > productQAPageSize {
		result.NextCursor = rows[productQAPageSize-1].Cursor
		rows = rows[:productQAPageSize]
	}
	for _, row := range rows {
		if p, err := findProduct(&w, row.ProductID); err == nil {
			row.ProductName = p.Name
			_, visibilityErr := publicProduct(w, row.ProductID)
			row.CanReply = qaAccess(p.Content, visibilityErr == nil).TeamReplies
		}
		result.Questions = append(result.Questions, row)
	}
	return result, nil
}

func (s *Service) qaRepository() ProductQARepository {
	store, _ := s.repository.(ProductQARepository)
	return store
}

func (s *Service) PublicProductQA(ctx context.Context, token string, after int64) (PublicProductQAPage, error) {
	result := PublicProductQAPage{ProductQAPage: ProductQAPage[PublicQuestion]{Questions: []PublicQuestion{}}}
	store := s.qaRepository()
	if store == nil {
		return result, ErrNotFound
	}
	id, err := s.resolveCatalogToken(ctx, "product", token)
	if err != nil {
		return result, err
	}
	state, questions, err := store.ListProductQA(ctx, id, after, productQAPageSize+1, true)
	if err != nil {
		return result, err
	}
	product, err := publicProduct(state, id)
	if err != nil {
		return result, err
	}
	access := qaAccess(product.Content, true)
	result.CanAsk = access.NewSubmissions
	if !access.ApprovedContentVisible {
		return result, nil
	}
	questions, result.NextCursor = productQAPage(questions)
	for _, q := range questions {
		if q.Status != catalog.QAApproved {
			continue
		}
		public := PublicQuestion{PublicQAContent: publicQAContent(q.ProductQAContent), Answers: []PublicQAContent{}}
		for _, answer := range q.Answers {
			if answer.Status == catalog.QAApproved {
				public.Answers = append(public.Answers, publicQAContent(answer))
			}
		}
		result.Questions = append(result.Questions, public)
	}
	return result, nil
}

func publicQAContent(content catalog.ProductQAContent) PublicQAContent {
	return PublicQAContent{ID: content.ID, Text: content.Text, CreatedAt: content.CreatedAt}
}

func (s *Service) AdminProductQA(ctx context.Context, actor identity.User, id string, after int64) (AdminProductQAPage, error) {
	result := AdminProductQAPage{ProductQAPage: ProductQAPage[catalog.ProductQuestion]{Questions: []catalog.ProductQuestion{}}}
	if !qaAdmin(actor) {
		return result, ErrForbidden
	}
	store := s.qaRepository()
	if store == nil {
		return result, ErrNotFound
	}
	state, questions, err := store.ListProductQA(ctx, id, after, productQAPageSize+1, false)
	if err != nil {
		return result, err
	}
	product, err := findProduct(&state, id)
	if err != nil {
		return result, err
	}
	_, visibilityErr := publicProduct(state, id)
	result.CanReply = qaAccess(product.Content, visibilityErr == nil).TeamReplies
	result.Questions, result.NextCursor = productQAPage(questions)
	if result.Questions == nil {
		result.Questions = []catalog.ProductQuestion{}
	}
	return result, nil
}

func productQAPage(questions []catalog.ProductQuestion) ([]catalog.ProductQuestion, int64) {
	if len(questions) > productQAPageSize {
		return questions[:productQAPageSize], questions[productQAPageSize-1].Cursor
	}
	return questions, 0
}

func qaAccess(content *catalog.ProductContent, visible bool) catalog.InteractionAccess {
	return catalog.ProductInteractions(content, visible, catalog.InteractionCapabilities{QA: true}).QA
}

func qaAdmin(actor identity.User) bool {
	return actor.Surface == identity.SurfaceBackoffice && actor.Role == identity.RoleSuperAdmin
}

type ProductQAMutation struct {
	Action   string           `json:"action"`
	Text     string           `json:"text,omitempty"`
	AnswerID string           `json:"answer_id,omitempty"`
	Status   catalog.QAStatus `json:"status,omitempty"`
	Version  int              `json:"version,omitempty"`
}

func (s *Service) AskProductQuestion(ctx context.Context, actor identity.User, token, key, text string) error {
	if actor.Surface != identity.SurfaceCustomer || actor.ID == "" {
		return ErrForbidden
	}
	id, err := s.resolveCatalogToken(ctx, "product", token)
	if err != nil {
		return err
	}
	return s.changeProductQA(ctx, actor, id, "", key, ProductQAMutation{Action: "ask", Text: text})
}

func (s *Service) ManageProductQuestion(ctx context.Context, actor identity.User, product, question, key string, mutation ProductQAMutation) error {
	if !qaAdmin(actor) {
		return ErrForbidden
	}
	if question == "" || len(question) > 128 || len(mutation.AnswerID) > 128 || mutation.Version < 1 {
		return ErrInvalid
	}
	switch mutation.Action {
	case "reply":
		if mutation.AnswerID != "" || mutation.Status != "" {
			return ErrInvalid
		}
	case "moderate":
		if mutation.Text != "" || (mutation.Status != catalog.QAApproved && mutation.Status != catalog.QARejected) {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	return s.changeProductQA(ctx, actor, product, question, key, mutation)
}

func (s *Service) changeProductQA(ctx context.Context, actor identity.User, product, question, key string, mutation ProductQAMutation) error {
	if strings.TrimSpace(key) == "" || len(key) > 128 || strings.ContainsRune(key, 0) || product == "" || len(product) > 128 {
		return ErrInvalid
	}
	mutation.Text = strings.TrimSpace(mutation.Text)
	if mutation.Action != "moderate" && (mutation.Text == "" || !utf8.ValidString(mutation.Text) || strings.ContainsRune(mutation.Text, 0) || utf8.RuneCountInString(mutation.Text) > catalog.ProductQATextLimit) {
		return ErrInvalid
	}
	store := s.qaRepository()
	if store == nil {
		return ErrNotFound
	}
	encoded, err := json.Marshal([]any{product, question, mutation})
	if err != nil {
		return err
	}
	sum := sha256.Sum256(encoded)
	fingerprint := hex.EncodeToString(sum[:])
	return store.UpdateProductQA(ctx, product, question, actor.ID, key, fingerprint, func(w Workspace, previous string, item *catalog.ProductQuestion) (*catalog.ProductQuestion, error) {
		if err := authorizeQAChange(w, actor, product, mutation.Action); err != nil {
			return nil, err
		}
		if previous != "" {
			if previous != fingerprint {
				return nil, ErrConflict
			}
			return nil, nil
		}
		if question != "" {
			if item == nil {
				return nil, ErrNotFound
			}
			if item.Version != mutation.Version {
				return nil, ErrConflict
			}
		}
		return s.applyQAChange(item, mutation)
	})
}

func authorizeQAChange(w Workspace, actor identity.User, product, action string) error {
	if action != "ask" {
		current, err := currentActor(w, actor)
		if err != nil || !qaAdmin(current) {
			return ErrForbidden
		}
	}
	if action == "moderate" {
		// Staff can remove or review historical content on hidden products.
		_, err := findProduct(&w, product)
		return err
	}
	p, err := publicProduct(w, product)
	if err != nil {
		return err
	}
	access := qaAccess(p.Content, true)
	if action == "ask" && !access.NewSubmissions || action == "reply" && !access.TeamReplies {
		return ErrForbidden
	}
	if action == "ask" && !slices.ContainsFunc(w.Users, func(u User) bool {
		return u.ID == actor.ID && u.Active && u.AccessVersion == actor.AccessVersion
	}) {
		return ErrForbidden
	}
	return nil
}

func (s *Service) applyQAChange(item *catalog.ProductQuestion, mutation ProductQAMutation) (*catalog.ProductQuestion, error) {
	if mutation.Action == "moderate" {
		target := &item.ProductQAContent
		if mutation.AnswerID != "" {
			index := slices.IndexFunc(item.Answers, func(a catalog.ProductQAContent) bool { return a.ID == mutation.AnswerID })
			if index < 0 {
				return nil, ErrNotFound
			}
			target = &item.Answers[index]
		}
		target.Status = mutation.Status
		item.Version++
		return item, nil
	}
	id, err := s.identifiers.Token()
	if err != nil {
		return nil, err
	}
	body := catalog.ProductQAContent{ID: id, Text: mutation.Text, Status: catalog.QAPending, CreatedAt: s.clock.Now()}
	if mutation.Action == "ask" {
		return &catalog.ProductQuestion{ProductQAContent: body, Version: 1, Answers: []catalog.ProductQAContent{}}, nil
	}
	item.Answers = append(item.Answers, body)
	item.Version++
	return item, nil
}
