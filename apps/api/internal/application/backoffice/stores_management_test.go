package backoffice_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func TestStoreConcurrentUpdatesAndCurrentRoleRevocation(t *testing.T) {
	for _, same := range []bool{false, true} {
		t.Run(fmt.Sprint(same), func(t *testing.T) {
			s, repo := testService()
			ctx := context.Background()
			a := actor(identity.RoleSuperAdmin)
			a.Surface = identity.SurfaceBackoffice
			var success atomic.Int32
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					key := "store-once"
					if !same {
						key = fmt.Sprint(key, i)
					}
					_, e := s.SaveStore(ctx, a, backoffice.StoreInput{ID: "s1", Version: 0, Name: "نام ساختگی", City: "تهران", Active: true}, key)
					if e == nil {
						success.Add(1)
					} else if !errors.Is(e, backoffice.ErrConflict) {
						t.Errorf("unexpected %v", e)
					}
				}(i)
			}
			wg.Wait()
			want := int32(1)
			if same {
				want = 10
			}
			if success.Load() != want {
				t.Fatal("concurrent success", success.Load())
			}
			w, _ := repo.Snapshot(ctx)
			if w.Stores[0].Version != 1 || len(w.StoreMutations) != 1 {
				t.Fatal("duplicate mutation")
			}
			repo.Update(ctx, func(w *backoffice.Workspace) error {
				for i := range w.Users {
					if w.Users[i].ID == a.ID {
						w.Users[i].Role = identity.RoleViewer
					}
				}
				return nil
			})
			_, e := s.SaveStore(ctx, a, backoffice.StoreInput{ID: "s1", Version: 1, Name: "ممنوع", Active: true}, "revoked-store")
			if !errors.Is(e, backoffice.ErrForbidden) {
				t.Fatal("stale role accepted", e)
			}
		})
	}
}
