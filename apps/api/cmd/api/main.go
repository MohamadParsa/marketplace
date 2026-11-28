package main

import (
	"log/slog"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	fx.New(
		fx.Provide(config.Load, bootstrap.NewOperationLogging, newLogger, bootstrap.NewManagedHTTPServer),
		fx.Invoke(bootstrap.RegisterHTTPServer),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),
		fx.StopTimeout(15*time.Second),
	).Run()
}

func newLogger(logs *bootstrap.OperationLogging) *slog.Logger {
	logger := slog.New(operationlog.SlogHandler{Recorder: logs.Recorder})
	slog.SetDefault(logger)
	return logger
}
