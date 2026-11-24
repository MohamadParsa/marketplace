package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const (
	reviewPending        = "pending"
	reviewApproved       = "approved"
	reviewRejected       = "rejected"
	publicReviewPageSize = 5
	adminReviewPageSize  = 20
)

var ErrReviewSubmissionsClosed = errors.New("review_submissions_closed")

// ProductReview is part of the private workspace aggregate. Use the explicit
// public and administrative projections at the HTTP boundary.
type ProductReview struct {
	ID                      string
	ProductID               string
	Name                    string
	Text                    string
	Rating                  int
	Anonymous               bool
	Status                  string
	CreatedAt               time.Time
	Version                 int
	ReviewedBy              string
	ReviewedAt              time.Time
	RejectionReason         string
	Reply                   string
	RepliedBy               string
	RepliedAt               time.Time
	SubmissionKey           string
	SubmissionFingerprint   string
	LastDecisionKey         string
	LastDecisionFingerprint string
}

type ProductReviewInput struct {
	Name           string `json:"name"`
	Text           string `json:"text"`
	Rating         int    `json:"rating"`
	Anonymous      bool   `json:"anonymous"`
	IdempotencyKey string `json:"idempotency_key"`
}

type ProductReviewDecision struct {
	Action         string `json:"action"`
	Text           string `json:"text"`
	Version        int    `json:"version"`
	IdempotencyKey string `json:"idempotency_key"`
}

type PublicProductReview struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	Text      string    `json:"text"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	Reply     string    `json:"reply,omitempty"`
}

type AdminProductReview struct {
	PublicProductReview
	ProductID       string `json:"product_id"`
	ProductName     string `json:"product_name"`
	Status          string `json:"status"`
	Anonymous       bool   `json:"anonymous"`
	Version         int    `json:"version"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type ProductReviewPage[T any] struct {
	Reviews            []T    `json:"reviews"`
	NextCursor         string `json:"next_cursor,omitempty"`
	SubmissionsEnabled bool   `json:"submissions_enabled"`
}

func reviewAccess(w Workspace, productID string) (catalog.InteractionAccess, error) {
	p, err := publicProduct(indexedCatalog(w), productID)
	if err != nil {
		return catalog.InteractionAccess{}, err
	}
	return catalog.ProductInteractions(p.Content, true, catalog.InteractionCapabilities{Reviews: true}).Reviews, nil
}

func (s *Service) SubmitProductReview(ctx context.Context, productID string, input ProductReviewInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Text = strings.TrimSpace(input.Text)
	if len(productID) > 128 || !validReviewKey(input.IdempotencyKey) || !validReviewText(input.Text, 2000) ||
		input.Rating < 1 || input.Rating > 5 ||
		((!input.Anonymous || input.Name != "") && !validReviewText(input.Name, 80)) {
		return ErrInvalid
	}
	fingerprint, err := reviewFingerprint(input)
	if err != nil {
		return err
	}
	id, err := s.identifiers.Token()
	if err != nil {
		return err
	}
	return s.repository.Update(ctx, func(w *Workspace) error {
		access, err := reviewAccess(*w, productID)
		if err != nil {
			return err
		}
		if !access.NewSubmissions {
			return ErrReviewSubmissionsClosed
		}
		for _, review := range w.ProductReviews {
			if review.ProductID == productID && review.SubmissionKey == input.IdempotencyKey {
				if review.SubmissionFingerprint == fingerprint {
					return nil
				}
				return ErrConflict
			}
		}
		w.ProductReviews = append(w.ProductReviews, ProductReview{
			ID: id, ProductID: productID, Name: input.Name, Text: input.Text,
			Rating: input.Rating, Anonymous: input.Anonymous, Status: reviewPending,
			Version: 1, CreatedAt: s.clock.Now(), SubmissionKey: input.IdempotencyKey,
			SubmissionFingerprint: fingerprint,
		})
		return nil
	})
}

// publicReviewRepository returns the catalog inputs and page from one database
// snapshot. The public path does not need a full workspace or access to identity.
type publicReviewRepository interface {
	ReadPublicReviewPage(context.Context, string, string, int) (Workspace, []PublicProductReview, error)
}

func (s *Service) PublicProductReviews(ctx context.Context, productID, cursor string) (ProductReviewPage[PublicProductReview], error) {
	if len(productID) > 128 || !validReviewCursor(cursor) {
		return ProductReviewPage[PublicProductReview]{}, ErrInvalid
	}
	if repo, ok := s.repository.(publicReviewRepository); ok {
		w, rows, err := repo.ReadPublicReviewPage(ctx, productID, cursor, publicReviewPageSize+1)
		if err != nil {
			return ProductReviewPage[PublicProductReview]{}, err
		}
		access, err := reviewAccess(w, productID)
		if err != nil {
			return ProductReviewPage[PublicProductReview]{}, err
		}
		page := ProductReviewPage[PublicProductReview]{
			Reviews: rows, SubmissionsEnabled: access.NewSubmissions,
		}
		if len(rows) > publicReviewPageSize {
			page.Reviews = rows[:publicReviewPageSize]
			page.NextCursor = page.Reviews[publicReviewPageSize-1].ID
		}
		return page, nil
	}
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return ProductReviewPage[PublicProductReview]{}, err
	}
	access, err := reviewAccess(w, productID)
	if err != nil {
		return ProductReviewPage[PublicProductReview]{}, err
	}
	page := projectReviewPage(w.ProductReviews, cursor, publicReviewPageSize, func(r ProductReview) bool {
		return r.ProductID == productID && r.Status == reviewApproved && access.ApprovedContentVisible
	}, projectPublicReview)
	page.SubmissionsEnabled = access.NewSubmissions
	return page, nil
}

func (s *Service) AdminProductReviews(ctx context.Context, actor identity.User, status, cursor string) (ProductReviewPage[AdminProductReview], error) {
	if !validReviewCursor(cursor) || (status != reviewPending && status != reviewApproved && status != reviewRejected) {
		return ProductReviewPage[AdminProductReview]{}, ErrInvalid
	}
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return ProductReviewPage[AdminProductReview]{}, err
	}
	if err := authorizeReviewAdmin(w, actor); err != nil {
		return ProductReviewPage[AdminProductReview]{}, err
	}
	names := make(map[string]string, len(w.Products))
	for _, product := range w.Products {
		names[product.ID] = product.Name
	}
	return projectReviewPage(w.ProductReviews, cursor, adminReviewPageSize, func(r ProductReview) bool {
		return r.Status == status
	}, func(r ProductReview) AdminProductReview {
		review := projectPublicReview(r)
		review.Name = r.Name
		return AdminProductReview{PublicProductReview: review, ProductID: r.ProductID, ProductName: names[r.ProductID],
			Status: r.Status, Anonymous: r.Anonymous, Version: r.Version, RejectionReason: r.RejectionReason}
	}), nil
}

func (s *Service) DecideProductReview(ctx context.Context, actor identity.User, id string, decision ProductReviewDecision) error {
	decision.Text = strings.TrimSpace(decision.Text)
	if len(id) > 128 || decision.Version < 1 || !validReviewKey(decision.IdempotencyKey) {
		return ErrInvalid
	}
	switch decision.Action {
	case "approve":
		if decision.Text != "" {
			return ErrInvalid
		}
	case "reject", "publish_reply":
		if !validReviewText(decision.Text, 2000) {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	fingerprint, err := reviewFingerprint(decision)
	if err != nil {
		return err
	}
	return s.repository.Update(ctx, func(w *Workspace) error {
		if err := authorizeReviewAdmin(*w, actor); err != nil {
			return err
		}
		for i := range w.ProductReviews {
			r := &w.ProductReviews[i]
			if r.ID != id {
				continue
			}
			if r.LastDecisionKey == decision.IdempotencyKey {
				if r.LastDecisionFingerprint == fingerprint {
					return nil
				}
				return ErrConflict
			}
			if r.Version != decision.Version {
				return ErrConflict
			}
			if err := s.applyReviewDecision(*w, r, actor.ID, decision); err != nil {
				return err
			}
			r.Version++
			r.LastDecisionKey, r.LastDecisionFingerprint = decision.IdempotencyKey, fingerprint
			return nil
		}
		return ErrNotFound
	})
}

func (s *Service) applyReviewDecision(w Workspace, r *ProductReview, actorID string, d ProductReviewDecision) error {
	if d.Action == "publish_reply" {
		access, err := reviewAccess(w, r.ProductID)
		if err != nil {
			return err
		}
		if !access.TeamReplies || r.Status != reviewApproved || r.Reply != "" {
			return ErrConflict
		}
		r.Reply, r.RepliedBy, r.RepliedAt = d.Text, actorID, s.clock.Now()
		return nil
	}
	if r.Status != reviewPending {
		return ErrConflict
	}
	if d.Action == "approve" {
		r.Status = reviewApproved
	} else {
		r.Status, r.RejectionReason = reviewRejected, d.Text
	}
	r.ReviewedBy, r.ReviewedAt = actorID, s.clock.Now()
	return nil
}

func authorizeReviewAdmin(w Workspace, actor identity.User) error {
	if actor.Surface != identity.SurfaceBackoffice {
		return ErrForbidden
	}
	current, err := currentActor(w, actor)
	if err != nil || current.Role != identity.RoleSuperAdmin {
		return ErrForbidden
	}
	return nil
}

func validReviewText(value string, max int) bool {
	return value != "" && !strings.ContainsRune(value, '\x00') && utf8.ValidString(value) && utf8.RuneCountInString(value) <= max
}

func validReviewKey(key string) bool {
	return len(key) >= 16 && validReviewCursor(key)
}

func validReviewCursor(token string) bool {
	return len(token) <= 128 && strings.IndexFunc(token, func(r rune) bool {
		return !(r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r == '-' || r == '_')
	}) < 0
}

func reviewFingerprint(input any) (string, error) {
	encoded, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}

func projectPublicReview(r ProductReview) PublicProductReview {
	out := PublicProductReview{ID: r.ID, Text: r.Text, Rating: r.Rating, CreatedAt: r.CreatedAt, Reply: r.Reply}
	if !r.Anonymous {
		out.Name = r.Name
	}
	return out
}

func projectReviewPage[T any](reviews []ProductReview, cursor string, size int, matches func(ProductReview) bool, project func(ProductReview) T) ProductReviewPage[T] {
	page := ProductReviewPage[T]{Reviews: make([]T, 0, size)}
	lastID := ""
	for i := len(reviews) - 1; i >= 0; i-- {
		r := reviews[i]
		if cursor != "" {
			if r.ID == cursor {
				cursor = ""
			}
			continue
		}
		if !matches(r) {
			continue
		}
		if len(page.Reviews) == size {
			page.NextCursor = lastID
			break
		}
		page.Reviews = append(page.Reviews, project(r))
		lastID = r.ID
	}
	return page
}
