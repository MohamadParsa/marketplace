package bootstrap

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/customer"
	"go.uber.org/fx"
	"testing"
	"time"
)

type recentTestLifecycle struct{ hooks []fx.Hook }

func (l *recentTestLifecycle) Append(h fx.Hook) { l.hooks = append(l.hooks, h) }
func TestRecentHistoryCleanupLifecycleWithoutRedis(t *testing.T) {
	w := backofficememory.Fixture()
	w.RecentlyViewed = []customer.RecentlyViewedEntry{{UserID: "admin", ProductID: "old", ViewedAt: time.Now().Add(-31 * 24 * time.Hour)}}
	repo := backofficememory.New(w)
	s := backoffice.NewService(repo, authmemory.Clock{}, authmemory.Random{})
	l := &recentTestLifecycle{}
	registerRecentHistoryCleanup(l, s)
	if len(l.hooks) != 1 {
		t.Fatal("cleanup hook missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if e := l.hooks[0].OnStart(ctx); e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := l.hooks[0].OnStop(ctx); e != nil {
			t.Fatal(e)
		}
	}()
	for {
		state, e := repo.Snapshot(ctx)
		if e != nil {
			t.Fatal(e)
		}
		if len(state.RecentlyViewed) == 0 {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("startup purge did not run")
		case <-time.After(5 * time.Millisecond):
		}
	}
}
