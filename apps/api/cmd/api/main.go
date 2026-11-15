package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	fx.New(
		fx.Provide(config.Load, newLogger, bootstrap.NewHTTPServer),
		fx.Invoke(bootstrap.RegisterHTTPServer),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),
		fx.StopTimeout(15*time.Second),
	).Run()
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
