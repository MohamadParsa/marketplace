package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"go.uber.org/fx"
)

// RegisterHTTPServer ties the listener and in-flight requests to the Fx lifecycle.
func RegisterHTTPServer(lifecycle fx.Lifecycle, server *http.Server, logger *slog.Logger, shutdowner fx.Shutdowner) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", server.Addr)
			if err != nil {
				return err
			}
			logger.Info("starting api", "addr", listener.Addr().String())
			go func() {
				if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error("http server failed", "error", err)
					if err := shutdowner.Shutdown(fx.ExitCode(1)); err != nil {
						logger.Error("requesting shutdown failed", "error", err)
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Reserve part of Fx's stop budget for forced connection cleanup.
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				return errors.Join(err, server.Close())
			}
			return nil
		},
	})
}
