package auth_test

import (
	"context"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"golang.org/x/crypto/bcrypt"
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

func TestPasswordLoginUsesTheSameUserAndDoesNotRequireOTP(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("test-password-123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	storage := authmemory.NewStore([]identity.User{{ID: "user-1", Phone: "+989120000001", Surface: identity.SurfaceCustomer, Active: true, PasswordHash: string(hash)}})
	service := auth.NewService(storage, storage, storage, authmemory.NewNotifier(), &clock{at: time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)}, authmemory.Random{})
	session, err := service.PasswordLogin(context.Background(), identity.SurfaceCustomer, "۰۹۱۲۰۰۰۰۰۰۱", "test-password-123")
	if err != nil {
		t.Fatal(err)
	}
	if session.User.ID != "user-1" || session.User.PasswordHash != string(hash) {
		t.Fatalf("unexpected session user: %+v", session.User)
	}
	if _, err := service.PasswordLogin(context.Background(), identity.SurfaceCustomer, "09120000001", "wrong"); !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Fatalf("got %v", err)
	}
}

func TestProfilePasswordAndMobileRecovery(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("old-password"), bcrypt.MinCost)
	storage := authmemory.NewStore([]identity.User{{ID: "user-3", Name: "کاربر آزمون", Phone: "09000000003", Surface: identity.SurfaceCustomer, Active: true, PasswordHash: string(hash)}})
	notifier := authmemory.NewNotifier()
	clk := &clock{at: time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)}
	service := auth.NewService(storage, storage, storage, notifier, clk, authmemory.Random{})
	session, err := service.PasswordLogin(context.Background(), identity.SurfaceCustomer, "09000000003", "old-password")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := service.UpdateProfile(context.Background(), identity.SurfaceCustomer, session.Token, "کاربر جدید", "", "new-password")
	if err != nil || updated.Name != "کاربر جدید" {
		t.Fatalf("profile update: %+v %v", updated, err)
	}
	requestID, _, err := service.RequestRecoveryOTP(context.Background(), identity.SurfaceCustomer, "09000000003", "mobile")
	if err != nil {
		t.Fatal(err)
	}
	code, ok := notifier.Code(identity.SurfaceCustomer, "09000000003")
	if !ok {
		t.Fatal("missing recovery code")
	}
	if err := service.RecoverPassword(context.Background(), identity.SurfaceCustomer, requestID, code, "recovered-password"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PasswordLogin(context.Background(), identity.SurfaceCustomer, "09000000003", "recovered-password"); err != nil {
		t.Fatal(err)
	}
}

func TestPasswordFailureLimitUsesSameAccountForEmailAndPhone(t *testing.T) {
	service, notifier, clock := setup()
	id, code := challenge(t, service, notifier)
	session, err := service.VerifyOTP(context.Background(), identity.SurfaceBackoffice, id, code)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.UpdateProfile(context.Background(), identity.SurfaceBackoffice, session.Token, "مدیر آزمون", "admin@example.test", "test-password-27")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		identifier := "09000000001"
		if i%2 == 1 {
			identifier = "admin@example.test"
		}
		_, err = service.PasswordLogin(context.Background(), identity.SurfaceBackoffice, identifier, "wrong")
		if !errors.Is(err, identity.ErrInvalidCredentials) {
			t.Fatal(err)
		}
	}
	if _, err = service.PasswordLogin(context.Background(), identity.SurfaceBackoffice, "09000000001", "test-password-27"); !errors.Is(err, identity.ErrRateLimit) {
		t.Fatal("limit bypass", err)
	}
	clock.at = clock.at.Add(time.Hour + time.Second)
	if _, err = service.PasswordLogin(context.Background(), identity.SurfaceBackoffice, "admin@example.test", "test-password-27"); err != nil {
		t.Fatal(err)
	}
}
