package backofficememory

import (
	"context"
	"slices"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

type qaReceiptKey struct{ actor, key string }

func (r *Repository) ListAdminProductQA(ctx context.Context, product string, status catalog.QAStatus, after int64, limit int) (backoffice.Workspace, []backoffice.AdminProductQuestion, error) {
	if err := ctx.Err(); err != nil {
		return backoffice.Workspace{}, nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	w, err := clone(r.state)
	if err != nil {
		return w, nil, err
	}
	rows := []backoffice.AdminProductQuestion{}
	for id, questions := range r.qaQuestions {
		if product != "" && product != id {
			continue
		}
		for _, q := range questions {
			if status != "" && q.Status != status || after > 0 && q.Cursor >= after {
				continue
			}
			q.Answers = slices.Clone(q.Answers)
			rows = append(rows, backoffice.AdminProductQuestion{ProductQuestion: q, ProductID: id})
		}
	}
	slices.SortFunc(rows, func(a, b backoffice.AdminProductQuestion) int {
		if a.Cursor > b.Cursor {
			return -1
		}
		if a.Cursor < b.Cursor {
			return 1
		}
		return 0
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return w, rows, nil
}

func (r *Repository) ListProductQA(ctx context.Context, product string, after int64, limit int, public bool) (backoffice.Workspace, []catalog.ProductQuestion, error) {
	if err := ctx.Err(); err != nil {
		return backoffice.Workspace{}, nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	workspace, err := clone(r.state)
	if err != nil {
		return workspace, nil, err
	}
	questions := []catalog.ProductQuestion{}
	rows := r.qaQuestions[product]
	for i := range rows {
		index := i
		if !public {
			index = len(rows) - 1 - i
		}
		question := rows[index]
		if public && (question.Cursor <= after || question.Status != catalog.QAApproved) {
			continue
		}
		if !public && after > 0 && question.Cursor >= after {
			continue
		}
		question.Answers = slices.Clone(question.Answers)
		questions = append(questions, question)
		if len(questions) == limit {
			break
		}
	}
	return workspace, questions, nil
}

func (r *Repository) UpdateProductQA(ctx context.Context, product, question, actor, key, fingerprint string, apply func(backoffice.Workspace, string, *catalog.ProductQuestion) (*catalog.ProductQuestion, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	receipt := qaReceiptKey{actor, key}
	questions := r.qaQuestions[product]
	index := slices.IndexFunc(questions, func(q catalog.ProductQuestion) bool { return q.ID == question })
	var current *catalog.ProductQuestion
	if index >= 0 {
		row := questions[index]
		row.Answers = slices.Clone(row.Answers)
		current = &row
	}
	next, err := apply(r.state, r.qaReceipts[receipt], current)
	if err != nil || next == nil {
		return err
	}
	if index >= 0 {
		questions[index] = *next
	} else {
		r.qaCursor++
		next.Cursor = r.qaCursor
		r.qaQuestions[product] = append(questions, *next)
	}
	r.qaReceipts[receipt] = fingerprint
	return nil
}
