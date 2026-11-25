package bootstrap_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/application/customer"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresReviewIdentitySurvivesRestart(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("set MARKETPLACE_TEST_POSTGRES_URL for PostgreSQL integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	control, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close(context.Background())

	schema := "review_test_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := control.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := control.Exec(cleanup, "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Errorf("cleanup test schema: %v", err)
		}
	}()

	options, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	options.ConnConfig.RuntimeParams["search_path"] = schema
	open := func() *pgxpool.Pool {
		t.Helper()
		pool, err := pgxpool.NewWithConfig(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pool.Close)
		return pool
	}
	migrations := filepath.Join("..", "..", "..", "..", "db", "migrations")
	pool := open()
	migrate := func(pool *pgxpool.Pool) {
		t.Helper()
		if err := reviewpostgres.Migrate(ctx, pool, os.DirFS(migrations)); err != nil {
			t.Fatal(err)
		}
	}
	migrate(pool)
	migrate(pool)
	newHandler := func(pool *pgxpool.Pool) http.Handler {
		t.Helper()
		repository, err := reviewpostgres.New(ctx, pool, backofficememory.Fixture())
		if err != nil {
			t.Fatal(err)
		}
		return bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repository)
	}
	handler := newHandler(pool)
	admin := login(t, handler, "09000000001")
	seller := loginAt(t, handler, "seller", "09000000002")["token"].(string)
	requestAs(t, handler, admin, "POST", "/backoffice/commands", `{"action":"role.create","name":"persisted","scope":"global","permissions":["workspace.read"],"idempotency_key":"pg-role"}`, 204)
	requestAs(t, handler, admin, "POST", "/backoffice/commands", `{"action":"membership.add","target":"seller-test","store_id":"s2","idempotency_key":"pg-member"}`, 204)
	requestAs(t, handler, admin, "POST", "/backoffice/commands", `{"action":"user.role-grant","target":"seller-test","value":"store-view","store_id":"s2","idempotency_key":"pg-grant"}`, 204)
	requestAs(t, handler, seller, "PUT", "/seller/stores/s1/offers/of1/stock", `{"stock":12,"version":1}`, 204)
	requestAs(t, handler, seller, "PUT", "/seller/stores/s2/offers/of2/stock", `{"stock":12,"version":1}`, 403)
	assertConstraint := func(sql, code string) {
		t.Helper()
		_, err := pool.Exec(ctx, sql)
		var violation *pgconn.PgError
		if !errors.As(err, &violation) || violation.Code != code {
			t.Fatalf("expected PostgreSQL constraint %s: %v", code, err)
		}
	}
	assertConstraint("INSERT INTO review_users(id,phone,name,base_role,active,access_version) VALUES ('duplicate','09000000001','duplicate','viewer',true,0)", "23505")
	assertConstraint("INSERT INTO review_role_grants(user_id,role_id,store_id) VALUES ('seller-test','store-view','s4')", "23503")
	assertConstraint("INSERT INTO review_role_grants(user_id,role_id) VALUES ('seller-test','store-view')", "23503")

	pool.Close()
	pool = open()
	migrate(pool)
	restarted := newHandler(pool)
	requestAs(t, restarted, seller, "GET", "/seller/stores/s1/offers", "", 401)
	admin = login(t, restarted, "09000000001")
	state := readJSON[backoffice.Workspace](t, requestAs(t, restarted, admin, "GET", "/backoffice/workspace", "", 200))
	if len(state.Audit) != 3 || len(state.Access.Roles) != len(backofficememory.Fixture().Access.Roles)+1 {
		t.Fatalf("role or audit reset: %+v", state.Access)
	}
	seller = loginAt(t, restarted, "seller", "09000000002")["token"].(string)
	offers := readJSON[[]catalog.Offer](t, requestAs(t, restarted, seller, "GET", "/seller/stores/s1/offers", "", 200))
	var persisted catalog.Offer
	for _, o := range offers {
		if o.StoreID != "s1" {
			t.Fatal("foreign offer leaked")
		}
		if o.ID == "of1" {
			persisted = o
		}
	}
	if persisted.Stock != 12 || persisted.Version != 2 {
		t.Fatalf("stock reset: %+v", offers)
	}
	anotherSeller := loginAt(t, restarted, "seller", "09120000002")["token"].(string)
	requestAs(t, restarted, anotherSeller, "GET", "/seller/stores/s2/offers", "", 200)
	buyer := loginAt(t, restarted, "storefront", "09000000002")["token"].(string)
	purchases := readJSON[[]customer.Purchase](t, requestAs(t, restarted, buyer, "GET", "/storefront/account/purchases", "", 200))
	if len(purchases) != 1 || purchases[0].ID != "BUY-501" || len(purchases[0].Orders) != 2 {
		t.Fatalf("purchase ownership reset: %+v", purchases)
	}
	commerceDown, err := os.ReadFile(filepath.Join(migrations, "000003-commerce-ledger.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	creditPolicyDown, err := os.ReadFile(filepath.Join(migrations, "000010-wallet-credit-policy.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	credentialsDown, err := os.ReadFile(filepath.Join(migrations, "000002-review-credentials.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	down, err := os.ReadFile(filepath.Join(migrations, "000001-review-identity.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, string(creditPolicyDown)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(commerceDown)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(credentialsDown)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(down)); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, "DELETE FROM review_schema_migrations WHERE name IN ('000001-review-identity.up.sql','000002-review-credentials.up.sql','000003-commerce-ledger.up.sql','000010-wallet-credit-policy.up.sql')")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	migrate(pool)
	_ = newHandler(pool)
	pool.Close()
}
