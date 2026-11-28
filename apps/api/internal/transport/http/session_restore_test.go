package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func TestStaffSessionRestoreRevalidatesAndDoesNotExtendSession(t *testing.T) {
	for _, mode := range []string{"valid", "expired", "other-surface", "inactive", "revoked-access", "temporary-password", "logout"} {
		t.Run(mode, func(t *testing.T) {
			user := identity.User{ID: "staff", Phone: "09000000001", Surface: identity.SurfaceBackoffice, Active: true, AccessVersion: 2, PasswordHash: "private-hash"}
			snapshot := user
			expires := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
			switch mode {
			case "expired":
				expires = time.Now().Add(-time.Hour)
			case "other-surface":
				snapshot.Surface = identity.SurfaceSeller
			case "inactive":
				user.Active = false
			case "revoked-access":
				snapshot.AccessVersion = 1
			case "temporary-password":
				user.PasswordChangeRequired = true
			}
			storage := authmemory.NewStore([]identity.User{user})
			_ = storage.Put(context.Background(), identity.Session{Token: "session-token", User: snapshot, Expires: expires})
			notifier := authmemory.NewNotifier()
			service := auth.NewService(storage, storage, storage, notifier, authmemory.Clock{}, authmemory.Random{})
			router := NewRouter(RouterConfig{Auth: NewAuthHandler(service, notifier, true)})
			if mode == "logout" {
				_ = service.Logout(context.Background(), "session-token")
			}
			req := httptest.NewRequest(http.MethodGet, "/backoffice/auth/session", nil)
			req.Header.Set("Authorization", "Bearer session-token")
			result := httptest.NewRecorder()
			router.ServeHTTP(result, req)
			allowed := mode == "valid" || mode == "temporary-password"
			if allowed && result.Code != 200 || !allowed && result.Code != 401 {
				t.Fatalf("status=%d %s", result.Code, result.Body.String())
			}
			if result.Header().Get("Cache-Control") != "private, no-store" || strings.Contains(result.Body.String(), "private-hash") {
				t.Fatal("private response must not cache/leak hash")
			}
			if allowed && !strings.Contains(result.Body.String(), expires.Format(time.RFC3339)) {
				t.Fatal("restore extended expiry")
			}
			if mode == "temporary-password" {
				if !strings.Contains(result.Body.String(), `"PasswordChangeRequired":true`) {
					t.Fatal("restriction lost")
				}
				if _, e := service.Authenticate(context.Background(), "session-token", identity.SurfaceBackoffice); e != identity.ErrPasswordChangeRequired {
					t.Fatal("restore bypassed password restriction", e)
				}
			}
		})
	}
}
