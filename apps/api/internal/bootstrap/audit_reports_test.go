package bootstrap_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
)

func auditContext(events *[]operationlog.Event) context.Context {
	r := operationlog.New(operationlog.SinkFunc(func(e operationlog.Event) error { *events = append(*events, e); return nil }), "api", "test", "backoffice")
	ctx := operationlog.WithRecorder(operationlog.WithContext(context.Background(), operationlog.ContextData{OperationID: "operation-1", RequestID: "request-1"}), r)
	operationlog.SetActor(ctx, "backoffice", "u-admin", "")
	return ctx
}
func TestAuditAtomicProjectionImmutabilityAndFailedCommit(t *testing.T) {
	state := commerceSeed()
	state.Audit = nil
	repo := backofficememory.New(state)
	events := []operationlog.Event{}
	ctx := auditContext(&events)
	err := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Offers[0].Price = 2000000
		w.Offers[0].Version++
		w.Audit = append(w.Audit, backoffice.Audit{ID: "idempotency-PRIVATE", Actor: "u-admin", Action: "offer.price.update", Target: "offer", Reason: "PRIVATE-REASON"})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	saved, _ := repo.Snapshot(ctx)
	a := saved.Audit[0]
	if a.OperationID != "operation-1" || a.RequestID != "request-1" || a.Before["price_rial"] != "1000000" || a.After["price_rial"] != "2000000" || a.Version != 2 {
		t.Fatal("missing safe diff", a)
	}
	encoded, _ := json.Marshal(events)
	if strings.Contains(string(encoded), "PRIVATE") || len(events) != 1 {
		t.Fatal("leak or duplicate event")
	}
	if err = repo.Update(ctx, func(w *backoffice.Workspace) error { w.Offers[0].Price = 1; w.Audit[0].Reason = "altered"; return nil }); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("audit rewrite allowed", err)
	}
	current, _ := repo.Snapshot(ctx)
	if current.Offers[0].Price != 2000000 || current.Audit[0].Reason != "PRIVATE-REASON" {
		t.Fatal("audit guard did not rollback")
	}
	beforeEvents := len(events)
	commitErr := errors.New("simulated commit failure")
	err = backoffice.ObserveWorkspaceUpdate(ctx, func(fn func(*backoffice.Workspace) error) error {
		copy := commerceSeed()
		copy.Audit = nil
		if e := fn(&copy); e != nil {
			return e
		}
		return commitErr
	}, func(w *backoffice.Workspace) error {
		w.Audit = append(w.Audit, backoffice.Audit{ID: "failed", Action: "offer.price.update", Target: "offer"})
		return nil
	})
	if !errors.Is(err, commitErr) || len(events) != beforeEvents {
		t.Fatal("success emitted before commit")
	}
}
func TestAuditAndLocalLogReportsAdminOnlyFilteredAndRedacted(t *testing.T) {
	w := commerceSeed()
	repo := backofficememory.New(w)
	events := []operationlog.Event{}
	ctx := auditContext(&events)
	operationlog.SetActor(ctx, "backoffice", "admin", "")
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Offers[0].Stock--
		w.Offers[0].Version++
		w.Audit = append(w.Audit, backoffice.Audit{ID: "stock-1", Actor: "admin", Action: "offer.stock.update", Target: "offer", Reason: "PRIVATE-REASON"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", OperationLogDir: filepath.Join(t.TempDir(), "logs")}, repo)
	defer closeLogHandler(t, h)
	admin := login(t, h, "09000000001")
	ops := login(t, h, "09000000002")
	finance := login(t, h, "09000000005")
	for _, path := range []string{"/backoffice/observability/audit/query", "/backoffice/observability/events/query"} {
		for _, token := range []string{ops, finance} {
			commerceHTTP(t, h, token, "", "POST", path, map[string]any{}, "", 403)
		}
	}
	for _, token := range []string{ops, finance} {
		commerceHTTP(t, h, token, "", "GET", "/backoffice/observability/stats", nil, "", 403)
	}
	result := commerceHTTP(t, h, admin, "", "POST", "/backoffice/observability/audit/query", map[string]any{"operation_id": "operation-1", "limit": 1}, "", 200)
	if strings.Contains(string(result), "PRIVATE") || strings.Contains(string(result), "fingerprint") {
		t.Fatal("private audit exported")
	}
	var report backoffice.AuditReport
	json.Unmarshal(result, &report)
	if len(report.Items) != 1 || report.Items[0].After["stock"] != "9" || !report.Items[0].ReasonPresent {
		t.Fatal(string(result))
	}
	if report.Items[0].ActorLabel != "مدیر آزمایشی" {
		t.Fatal("actor label missing in authorized report")
	}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/observability/events/query", map[string]any{"limit": 201}, "", 400)
}
func TestAuditPostgresRestartAndRollback(t *testing.T) {
	repo, db := cachePostgres(t)
	if err := repo.Update(context.Background(), func(w *backoffice.Workspace) error { *w = commerceSeed(); w.Audit = nil; return nil }); err != nil {
		t.Fatal(err)
	}
	events := []operationlog.Event{}
	ctx := auditContext(&events)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Offers[0].Stock--
		w.Offers[0].Version++
		w.Audit = append(w.Audit, backoffice.Audit{ID: "pg-audit", Actor: "u-admin", Action: "offer.stock.update", Target: "offer"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error { w.Offers[0].Stock = 1; w.Audit = nil; return nil }); err != backoffice.ErrForbidden {
		t.Fatal(err)
	}
	reopened, err := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := reopened.Snapshot(ctx)
	if err != nil || saved.Offers[0].Stock != 9 || len(saved.Audit) != 1 || saved.Audit[0].OperationID != "operation-1" {
		t.Fatal("lost audit after restart", err)
	}
}
