package postgresrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists an explicitly selected live or isolated review workspace.
// Its row serializes commands; identity and financial projections enforce relational constraints.
type Repository struct{ db *pgxpool.Pool }

func NewReview(ctx context.Context, db *pgxpool.Pool, seed backoffice.Workspace) (*Repository, error) {
	return initialize(ctx, db, seed, "review")
}
func NewLive(ctx context.Context, db *pgxpool.Pool) (*Repository, error) {
	return initialize(ctx, db, backoffice.EmptyWorkspace(), "live")
}
func initialize(ctx context.Context, db *pgxpool.Pool, seed backoffice.Workspace, mode string) (*Repository, error) {
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, "INSERT INTO review_workspace(singleton,payload,data_origin) VALUES(true,'{}',$1) ON CONFLICT DO NOTHING", mode)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			var origin string
			if err = tx.QueryRow(ctx, "SELECT data_origin FROM review_workspace WHERE singleton").Scan(&origin); err != nil {
				return err
			}
			if origin != mode {
				return fmt.Errorf("database data origin does not match requested mode")
			}
			return nil
		}
		return save(ctx, tx, seed)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize marketplace repository: %w", err)
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Snapshot(ctx context.Context) (backoffice.Workspace, error) {
	var state backoffice.Workspace
	err := pgx.BeginTxFunc(ctx, r.db, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		var err error
		state, err = load(ctx, tx, false)
		return err
	})
	return state, err
}

func (r *Repository) Update(ctx context.Context, apply func(*backoffice.Workspace) error) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		state, err := load(ctx, tx, true)
		if err != nil {
			return err
		}
		if err := apply(&state); err != nil {
			return err
		}
		return save(ctx, tx, state)
	})
}

func load(ctx context.Context, tx pgx.Tx, write bool) (backoffice.Workspace, error) {
	var state backoffice.Workspace
	query := "SELECT payload FROM review_workspace WHERE singleton"
	if write {
		query += " FOR UPDATE"
	}
	var payload []byte
	if err := tx.QueryRow(ctx, query).Scan(&payload); err != nil {
		return state, err
	}
	if err := json.Unmarshal(payload, &state); err != nil {
		return state, err
	}
	if err := loadIdentity(ctx, tx, &state); err != nil {
		return state, err
	}
	return state, nil
}

func loadIdentity(ctx context.Context, tx pgx.Tx, state *backoffice.Workspace) error {
	jobs := []func() error{
		func() error {
			return scan(ctx, tx, "SELECT id,name,phone,base_role,active,access_version,email,password_hash FROM review_users ORDER BY id", &state.Users, func(row pgx.CollectableRow, u *backoffice.User) error {
				return row.Scan(&u.ID, &u.Name, &u.Phone, &u.Role, &u.Active, &u.AccessVersion, &u.Email, &u.PasswordHash)
			})
		},
		func() error {
			return scan(ctx, tx, "SELECT user_id,store_id FROM review_memberships ORDER BY user_id, store_id", &state.StoreMemberships, func(row pgx.CollectableRow, m *identity.StoreMembership) error {
				return row.Scan(&m.UserID, &m.StoreID)
			})
		},
		func() error {
			return scan(ctx, tx, "SELECT id,name,scope,permissions,version FROM review_roles ORDER BY id", &state.Access.Roles, func(row pgx.CollectableRow, role *identity.RoleDefinition) error {
				return row.Scan(&role.ID, &role.Name, &role.Scope, &role.Permissions, &role.Version)
			})
		},
		func() error {
			return scan(ctx, tx, "SELECT user_id,role_id,coalesce(store_id,'') FROM review_role_grants ORDER BY user_id, role_id, store_id", &state.Access.Grants, func(row pgx.CollectableRow, grant *identity.RoleGrant) error {
				return row.Scan(&grant.UserID, &grant.RoleID, &grant.StoreID)
			})
		},
		func() error {
			return scan(ctx, tx, "SELECT purchase_id,user_id FROM review_purchase_owners ORDER BY purchase_id", &state.Purchases, func(row pgx.CollectableRow, owner *backoffice.PurchaseOwner) error {
				return row.Scan(&owner.PurchaseID, &owner.UserID)
			})
		},
	}
	state.Access = &backoffice.AccessDirectory{}
	for _, job := range jobs {
		if err := job(); err != nil {
			return err
		}
	}
	return nil
}

func scan[T any](ctx context.Context, tx pgx.Tx, sql string, result *[]T, read func(pgx.CollectableRow, *T) error) error {
	rows, err := tx.Query(ctx, sql)
	if err != nil {
		return err
	}
	*result, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (T, error) {
		var value T
		err := read(row, &value)
		return value, err
	})
	return err
}

func save(ctx context.Context, tx pgx.Tx, state backoffice.Workspace) error {
	// The review fixture is small and owned by the workspace transaction.
	// Replacing its access relations also makes revocations and membership removal atomic.
	batch := &pgx.Batch{}
	for _, user := range state.Users {
		batch.Queue(`INSERT INTO review_users (id,phone,name,base_role,active,access_version,email,password_hash) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT(id) DO UPDATE SET phone=excluded.phone, name=excluded.name, base_role=excluded.base_role,
			active=excluded.active, access_version=excluded.access_version,email=excluded.email,password_hash=excluded.password_hash`,
			user.ID, user.Phone, user.Name, user.Role, user.Active, user.AccessVersion, user.Email, user.PasswordHash)
	}
	for _, store := range state.Stores {
		batch.Queue("INSERT INTO review_stores VALUES ($1) ON CONFLICT DO NOTHING", store.ID)
	}
	batch.Queue("DELETE FROM review_role_grants")
	batch.Queue("DELETE FROM review_memberships")
	batch.Queue("DELETE FROM review_roles")
	if state.Access != nil {
		for _, role := range state.Access.Roles {
			batch.Queue("INSERT INTO review_roles VALUES ($1,$2,$3,$4,$5)", role.ID, role.Name, role.Scope, role.Permissions, role.Version)
		}
	}
	for _, membership := range state.StoreMemberships {
		batch.Queue("INSERT INTO review_memberships VALUES ($1,$2)", membership.UserID, membership.StoreID)
	}
	if state.Access != nil {
		for _, grant := range state.Access.Grants {
			batch.Queue("INSERT INTO review_role_grants(user_id,role_id,store_id) VALUES ($1,$2,NULLIF($3,''))", grant.UserID, grant.RoleID, grant.StoreID)
		}
	}
	for _, owner := range state.Purchases {
		batch.Queue("INSERT INTO review_purchase_owners VALUES ($1,$2) ON CONFLICT (purchase_id) DO NOTHING", owner.PurchaseID, owner.UserID)
	}
	state.Users, state.Access = nil, nil
	financialState := state.Commerce
	if financialState != nil {
		copyState := *financialState
		copyState.Receipts = append([]commerce.Receipt{}, financialState.Receipts...)
		for i := range copyState.Receipts {
			copyState.Receipts[i].Data = nil
		}
		state.Commerce = &copyState
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	batch.Queue("UPDATE review_workspace SET payload=$1::jsonb WHERE singleton", payload)
	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		return err
	}
	return saveCommerce(ctx, tx, financialState)
}
