package backoffice_test

import (
	"context"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"sync"
	"testing"
)

func testService() (*backoffice.Service, *backofficememory.Repository) {
	repository := backofficememory.New(backofficememory.Fixture())
	return backoffice.NewService(repository, authmemory.Clock{}, authmemory.Random{}), repository
}
func actor(role identity.Role) identity.User {
	id := map[identity.Role]string{identity.RoleSuperAdmin: "admin", identity.RoleOperations: "ops", identity.RoleSupport: "support", identity.RoleFinance: "finance", identity.RoleViewer: "viewer"}[role]
	return identity.User{ID: id, Role: role, Active: true}
}
func TestPermissionsAndRedaction(t *testing.T) {
	service, _ := testService()
	ctx := context.Background()
	viewer := actor(identity.RoleViewer)
	snapshot, err := service.Read(ctx, viewer)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Users) > 0 || len(snapshot.Transfers) > 0 || len(snapshot.Ledger) > 0 {
		t.Fatal("private data leaked to viewer")
	}
	if err := service.Execute(ctx, viewer, backoffice.Command{Action: "store.toggle", Target: "s1", IdempotencyKey: "x"}); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatalf("got %v", err)
	}
}
func TestConcurrentFinancialRetriesProduceOneLedgerEntry(t *testing.T) {
	service, repository := testService()
	ctx := context.Background()
	command := backoffice.Command{Action: "transfer.resolve", Target: "TX-302", Value: "success", Reason: "test", Version: 1, IdempotencyKey: "same-key"}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := service.Execute(ctx, actor(identity.RoleFinance), command); err != nil {
				t.Errorf("retry: %v", err)
			}
		}()
	}
	wg.Wait()
	snapshot, _ := repository.Snapshot(ctx)
	if len(snapshot.Ledger) != 1 || len(snapshot.Audit) != 1 {
		t.Fatalf("ledger=%d audit=%d", len(snapshot.Ledger), len(snapshot.Audit))
	}
}
func TestStaleVersionRollsBack(t *testing.T) {
	service, repository := testService()
	ctx := context.Background()
	err := service.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "product.publish", Target: "p2", Version: 99, IdempotencyKey: "stale"})
	if !errors.Is(err, backoffice.ErrConflict) {
		t.Fatalf("got %v", err)
	}
	snapshot, _ := repository.Snapshot(ctx)
	if snapshot.Products[1].Status != "pending" || len(snapshot.Audit) != 0 {
		t.Fatal("failed transaction changed state")
	}
}
func TestLastSuperAdminCannotBeDisabled(t *testing.T) {
	service, _ := testService()
	err := service.Execute(context.Background(), actor(identity.RoleSuperAdmin), backoffice.Command{Action: "user.toggle", Target: "admin", IdempotencyKey: "admin-disable"})
	if !errors.Is(err, backoffice.ErrConflict) {
		t.Fatalf("got %v", err)
	}
}
func TestUnknownDoesNotPostLedger(t *testing.T) {
	service, repository := testService()
	err := service.Execute(context.Background(), actor(identity.RoleFinance), backoffice.Command{Action: "transfer.resolve", Target: "TX-302", Value: "unknown", Version: 1, Reason: "test", IdempotencyKey: "unknown"})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, _ := repository.Snapshot(context.Background())
	if len(snapshot.Ledger) != 0 || snapshot.Transfers[1].Status != "unknown" {
		t.Fatal("unknown posted a financial effect")
	}
}

func TestSameIdempotencyKeyWithDifferentCommandConflicts(t *testing.T) {
	service, _ := testService()
	ctx := context.Background()
	command := backoffice.Command{Action: "store.toggle", Target: "s1", Reason: "first", IdempotencyKey: "reused"}
	if err := service.Execute(ctx, actor(identity.RoleSuperAdmin), command); err != nil {
		t.Fatal(err)
	}
	command.Reason = "different"
	if err := service.Execute(ctx, actor(identity.RoleSuperAdmin), command); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatalf("got %v", err)
	}
}
func TestFailureReleasesReservation(t *testing.T) {
	service, repository := testService()
	err := service.Execute(context.Background(), actor(identity.RoleFinance), backoffice.Command{Action: "transfer.resolve", Target: "TX-303", Value: "failure", Version: 1, Reason: "test", IdempotencyKey: "failure"})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, _ := repository.Snapshot(context.Background())
	for _, transfer := range snapshot.Transfers {
		if transfer.ID == "TX-303" && transfer.Reserved {
			t.Fatal("definite failure did not release reservation")
		}
	}
	if len(snapshot.Ledger) > 0 {
		t.Fatal("failure posted a ledger entry")
	}
}
