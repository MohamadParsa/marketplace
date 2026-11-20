package backoffice_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func sampleIcon(t *testing.T) []byte {
	t.Helper()
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func TestConcurrentIconUpdatesAreAtomic(t *testing.T) {
	for _, sameKey := range []bool{false, true} {
		t.Run(strconv.FormatBool(sameKey), func(t *testing.T) {
			s, repo := testService()
			ctx := context.Background()
			data := sampleIcon(t)
			var success atomic.Int32
			var wg sync.WaitGroup
			for i := 0; i < 12; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					key := "upload"
					if !sameKey {
						key += strconv.Itoa(i)
					}
					err := s.SetCatalogIcon(ctx, actor(identity.RoleOperations), "category", "c1", 0, key, data)
					if err == nil {
						success.Add(1)
					} else if !errors.Is(err, backoffice.ErrConflict) {
						t.Errorf("unexpected error: %v", err)
					}
				}(i)
			}
			wg.Wait()
			expected := int32(1)
			if sameKey {
				expected = 12
			}
			if success.Load() != expected {
				t.Fatalf("success %d want %d", success.Load(), expected)
			}
			state, _ := repo.Snapshot(ctx)
			if len(state.IconMedia) != 1 || len(state.Audit) != 1 || state.Categories[0].Version != 1 {
				t.Fatal("concurrent/retry updates leaked media or audit")
			}
		})
	}
}

func TestIconAuthorizationRechecksCurrentActor(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].ID == "ops" {
				w.Users[i].Role = identity.RoleViewer
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetCatalogIcon(ctx, actor(identity.RoleOperations), "brand", "b1", 1, "revoked", sampleIcon(t)); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatalf("got %v", err)
	}
	state, _ := repo.Snapshot(ctx)
	if len(state.IconMedia) != 0 || len(state.Audit) != 0 {
		t.Fatal("revoked actor changed data")
	}
}

func TestUnattachedAndArchivedIconsAreNotPublic(t *testing.T) {
	s, repo := testService()
	ctx := context.Background()
	data := sampleIcon(t)
	if err := repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.IconMedia = append(w.IconMedia, catalog.IconMedia{ID: "draft", Data: data})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReadCatalogIcon(ctx, "draft"); !errors.Is(err, backoffice.ErrNotFound) {
		t.Fatal("unattached icon exposed")
	}
	if err := s.SetCatalogIcon(ctx, actor(identity.RoleOperations), "category", "c1", 0, "valid", data); err != nil {
		t.Fatal(err)
	}
	state, _ := repo.Snapshot(ctx)
	id := state.Categories[0].IconID
	if err := s.Execute(ctx, actor(identity.RoleOperations), backoffice.Command{Action: "category.archive", Target: "c1", IdempotencyKey: "archive"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReadCatalogIcon(ctx, id); !errors.Is(err, backoffice.ErrNotFound) {
		t.Fatal("archived record image exposed")
	}
}
