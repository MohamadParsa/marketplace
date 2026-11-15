package auth_test

import (
	"context"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type clock struct{ at time.Time }

func (c *clock) Now() time.Time { return c.at }
func setup() (*auth.Service, *authmemory.Notifier, *clock) {
	storage := authmemory.NewStore([]identity.User{{ID: "admin", Phone: "09000000001", Surface: identity.SurfaceBackoffice, Role: identity.RoleSuperAdmin, Active: true}})
	notifier := authmemory.NewNotifier()
	clock := &clock{at: time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)}
	return auth.NewService(storage, storage, storage, notifier, clock, authmemory.Random{}), notifier, clock
}
func challenge(t *testing.T, service *auth.Service, notifier *authmemory.Notifier) (string, string) {
	t.Helper()
	id, err := service.RequestOTP(context.Background(), identity.SurfaceBackoffice, "09000000001")
	if err != nil {
		t.Fatal(err)
	}
	code, ok := notifier.Code(identity.SurfaceBackoffice, "09000000001")
	if !ok {
		t.Fatal("missing test delivery")
	}
	return id, code
}
func TestOTPIsConsumedExactlyOnceUnderConcurrency(t *testing.T) {
	service, notifier, _ := setup()
	id, code := challenge(t, service, notifier)
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := service.VerifyOTP(context.Background(), identity.SurfaceBackoffice, id, code); err == nil {
				successes.Add(1)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful logins=%d", successes.Load())
	}
}
func TestOTPExpiryAndAttemptLimit(t *testing.T) {
	for _, mode := range []string{"expired", "attempts"} {
		t.Run(mode, func(t *testing.T) {
			service, notifier, clock := setup()
			id, code := challenge(t, service, notifier)
			if mode == "expired" {
				clock.at = clock.at.Add(90 * time.Second)
			} else {
				for i := 0; i < 5; i++ {
					_, _ = service.VerifyOTP(context.Background(), identity.SurfaceBackoffice, id, "invalid")
				}
			}
			if _, err := service.VerifyOTP(context.Background(), identity.SurfaceBackoffice, id, code); !errors.Is(err, identity.ErrInvalidOTP) {
				t.Fatalf("got %v", err)
			}
		})
	}
}
func TestOTPSurfaceIsolationAndResend(t *testing.T) {
	service, notifier, clock := setup()
	id, code := challenge(t, service, notifier)
	if _, err := service.VerifyOTP(context.Background(), identity.SurfaceSeller, id, code); !errors.Is(err, identity.ErrInvalidOTP) {
		t.Fatal(err)
	}
	if _, err := service.RequestOTP(context.Background(), identity.SurfaceBackoffice, "09000000001"); !errors.Is(err, identity.ErrResendSoon) {
		t.Fatal(err)
	}
	clock.at = clock.at.Add(60 * time.Second)
	if _, err := service.RequestOTP(context.Background(), identity.SurfaceBackoffice, "09000000001"); err != nil {
		t.Fatal(err)
	}
}
