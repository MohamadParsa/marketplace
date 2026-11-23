package bootstrap

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"go.uber.org/fx"
	"log/slog"
	"time"
)

// This lifecycle is independent of Redis and cannot place private history in public warmup.
func registerRecentHistoryCleanup(lifecycle fx.Lifecycle, service *backoffice.Service) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
		go func() {
			defer close(done)
			clean := func() {
				run, stop := context.WithTimeout(ctx, 10*time.Second)
				defer stop()
				n, e := service.PurgeRecentlyViewed(run)
				if e != nil {
					if ctx.Err() == nil {
						slog.Warn("recent history cleanup failed")
					}
				} else if n > 0 {
					slog.Info("recent history cleanup", "removed", n)
				}
			}
			clean()
			ticker := time.NewTicker(time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					clean()
				}
			}
		}()
		return nil
	}, OnStop: func(stop context.Context) error {
		cancel()
		select {
		case <-done:
		case <-stop.Done():
			return stop.Err()
		}
		return nil
	}})
}
