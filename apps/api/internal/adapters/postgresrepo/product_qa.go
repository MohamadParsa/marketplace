package postgresrepo

import (
	"context"
	"encoding/json"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) ListProductQA(ctx context.Context, product string, after int64, limit int, public bool) (backoffice.Workspace, []catalog.ProductQuestion, error) {
	var workspace backoffice.Workspace
	questions := []catalog.ProductQuestion{}
	err := pgx.BeginTxFunc(ctx, r.db, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		var err error
		workspace, err = productQACatalog(ctx, tx, product)
		if err != nil {
			return err
		}
		query := "SELECT payload,cursor FROM product_qa_questions WHERE product_id=$1 AND ($2=0 OR cursor<$2) ORDER BY cursor DESC LIMIT $3"
		if public {
			query = "SELECT payload,cursor FROM product_qa_questions WHERE product_id=$1 AND cursor>$2 AND status='approved' ORDER BY cursor LIMIT $3"
		}
		rows, err := tx.Query(ctx, query, product, after, limit)
		if err != nil {
			return err
		}
		questions, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (catalog.ProductQuestion, error) {
			var question catalog.ProductQuestion
			err := scanProductQuestion(row, &question)
			return question, err
		})
		return err
	})
	return workspace, questions, err
}

func productQACatalog(ctx context.Context, tx pgx.Tx, id string) (backoffice.Workspace, error) {
	var workspace backoffice.Workspace
	var data []byte
	err := tx.QueryRow(ctx, `SELECT jsonb_build_object('categories',s.payload->'categories',
		'brands',s.payload->'brands','products',p.payload->'products')
		FROM catalog_public_state s LEFT JOIN catalog_public_products p ON p.id=$1 WHERE s.singleton`, id).Scan(&data)
	if err == nil {
		err = json.Unmarshal(data, &workspace)
	}
	return workspace, err
}

func scanProductQuestion(row pgx.Row, question *catalog.ProductQuestion) error {
	var data []byte
	if err := row.Scan(&data, &question.Cursor); err != nil {
		return err
	}
	return json.Unmarshal(data, question)
}

func (r *Repository) UpdateProductQA(ctx context.Context, product, question, actor, key, fingerprint string, apply func(backoffice.Workspace, string, *catalog.ProductQuestion) (*catalog.ProductQuestion, error)) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		// The workspace row lock also serializes toggles and category publication.
		var locked bool
		if err := tx.QueryRow(ctx, "SELECT singleton FROM review_workspace WHERE singleton FOR UPDATE").Scan(&locked); err != nil {
			return err
		}
		workspace, err := productQACatalog(ctx, tx, product)
		if err != nil {
			return err
		}
		var account backoffice.User
		err = tx.QueryRow(ctx, "SELECT id,active,access_version,base_role FROM review_users WHERE id=$1", actor).
			Scan(&account.ID, &account.Active, &account.AccessVersion, &account.Role)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}
		if err == nil {
			workspace.Users = append(workspace.Users, account)
		}
		var previous string
		err = tx.QueryRow(ctx, "SELECT fingerprint FROM product_qa_receipts WHERE actor_id=$1 AND request_key=$2", actor, key).Scan(&previous)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}
		var current *catalog.ProductQuestion
		if question != "" && previous == "" {
			var row catalog.ProductQuestion
			err = scanProductQuestion(tx.QueryRow(ctx, "SELECT payload,cursor FROM product_qa_questions WHERE product_id=$1 AND id=$2", product, question), &row)
			if err != nil && err != pgx.ErrNoRows {
				return err
			}
			if err == nil {
				current = &row
			}
		}
		next, err := apply(workspace, previous, current)
		if err != nil || next == nil {
			return err
		}
		data, err := json.Marshal(next)
		if err != nil {
			return err
		}
		if current == nil {
			_, err = tx.Exec(ctx, "INSERT INTO product_qa_questions (product_id,id,payload) VALUES ($1,$2,$3)", product, next.ID, data)
		} else {
			_, err = tx.Exec(ctx, "UPDATE product_qa_questions SET payload=$3 WHERE product_id=$1 AND id=$2", product, question, data)
		}
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, "INSERT INTO product_qa_receipts (actor_id,request_key,fingerprint) VALUES ($1,$2,$3)", actor, key, fingerprint)
		return err
	})
}
