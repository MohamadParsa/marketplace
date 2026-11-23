package bootstrap

import (
	"context"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/rediscache"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"log/slog"
	"net/http"
	"time"
)

func NewHTTPServer(cfg config.Config, lifecycle fx.Lifecycle) (*http.Server, error) {
	server := &http.Server{Addr: cfg.Address, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	review := cfg.Environment == "local" || cfg.Environment == "test"
	if review && cfg.DatabaseURL != "" {
		return nil, errors.New("live database requires a public environment; test OTP is disabled for live data")
	}
	if cfg.ReviewDatabaseURL != "" && cfg.DatabaseURL != "" {
		return nil, errors.New("configure only one database mode")
	}
	if !review && cfg.ReviewDatabaseURL != "" {
		return nil, errors.New("review database is restricted to local/test")
	}
	if !review && cfg.DatabaseURL == "" {
		return nil, errors.New("public environments require DATABASE_URL; memory fixtures are disabled")
	}
	if !review && (cfg.KavenegarAPIKey == "" || cfg.KavenegarTemplate == "") {
		return nil, errors.New("public checkout requires the configured OTP sender")
	}
	dsn := cfg.DatabaseURL
	if dsn == "" {
		dsn = cfg.ReviewDatabaseURL
	}
	if dsn == "" {
		server.Handler = NewHTTPHandler(cfg)
		return server, nil
	}
	options, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		return nil, errors.New("invalid database configuration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, e := pgxpool.NewWithConfig(ctx, options)
	if e != nil {
		return nil, e
	}
	var repo *postgresrepo.Repository
	if cfg.DatabaseURL != "" {
		repo, e = postgresrepo.NewLive(ctx, db)
	} else {
		repo, e = reviewpostgres.New(ctx, db, backofficememory.Fixture())
	}
	if e != nil {
		db.Close()
		return nil, e
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { db.Close(); return nil }})
	var cache *backoffice.PublicCache
	if cfg.PublicCacheEnabled && cfg.RedisURL != "" {
		store, err := rediscache.New(cfg.RedisURL)
		if err != nil {
			slog.Warn("public cache disabled: invalid configuration")
		} else {
			cache = backoffice.NewPublicCache(store)
			cache.ConfigureTTLs(cfg.CacheHomeTTL, cfg.CacheProductTTL, cfg.CacheListingTTL, cfg.CacheSuggestionsTTL)
			lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { return store.Close() }})
		}
	}
	var workspace *backoffice.Service
	server.Handler, workspace = newCatalogHandler(cfg, repo, cache)
	if cache != nil && cfg.WebSurface == "storefront" {
		warmCtx, warmCancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				ctx, cancel := context.WithTimeout(warmCtx, 45*time.Second)
				defer cancel()
				if cfg.CacheWarmupEnabled {
					result := workspace.WarmPublicCache(ctx, 100)
					slog.Info("public cache warmup", "result", result, "cache", cache.Stats(), "reads", repo.CatalogReadStatistics())
				}
				cancel()
				ticker := time.NewTicker(time.Minute)
				defer ticker.Stop()
				for {
					select {
					case <-warmCtx.Done():
						return
					case <-ticker.C:
						slog.Info("public cache statistics", "cache", cache.Stats(), "reads", repo.CatalogReadStatistics())
					}
				}
			}()
			return nil
		}, OnStop: func(ctx context.Context) error {
			warmCancel()
			select {
			case <-done:
			case <-ctx.Done():
			}
			return nil
		}})
	}
	return server, nil
}
