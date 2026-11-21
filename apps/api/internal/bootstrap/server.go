package bootstrap

import (
	"context"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"net/http"
	"time"
)

func NewHTTPServer(cfg config.Config, lifecycle fx.Lifecycle) (*http.Server, error) {
	server := &http.Server{Addr: cfg.Address, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	if cfg.ReviewDatabaseURL == "" {
		server.Handler = NewHTTPHandler(cfg)
		return server, nil
	}
	if cfg.Environment != "local" && cfg.Environment != "test" {
		return nil, errors.New("review database is restricted to local/test")
	}
	options, err := pgxpool.ParseConfig(cfg.ReviewDatabaseURL)
	if err != nil {
		return nil, errors.New("invalid REVIEW_DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := pgxpool.NewWithConfig(ctx, options)
	if err != nil {
		return nil, err
	}
	repository, err := reviewpostgres.New(ctx, db, backofficememory.Fixture())
	if err != nil {
		db.Close()
		return nil, err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { db.Close(); return nil }})
	server.Handler = NewHTTPHandlerWithRepository(cfg, repository)
	return server, nil
}
