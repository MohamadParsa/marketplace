package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/locallog"
	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"go.uber.org/fx"
)

type OperationLogging struct {
	Recorder *operationlog.Recorder
	Store    *locallog.Store
	Status   string
}

func buildOperationLogging(cfg config.Config) (*OperationLogging, error) {
	runtime := &OperationLogging{Status: "disabled"}
	if cfg.OperationLogDir == "" {
		return runtime, nil
	}
	dir, err := filepath.Abs(cfg.OperationLogDir)
	if err != nil {
		return nil, errors.New("invalid private log directory")
	}
	if cfg.WebRoot != "" {
		root, _ := filepath.Abs(cfg.WebRoot)
		rel, e := filepath.Rel(root, dir)
		if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, errors.New("log directory must be outside web root")
		}
	}
	opts := locallog.DefaultOptions(dir)
	if cfg.LogRetentionDays > 0 {
		opts.RetentionDays = cfg.LogRetentionDays
	}
	if cfg.BehaviorRetentionDays > 0 {
		opts.BehaviorDays = cfg.BehaviorRetentionDays
	}
	if cfg.LogMaxMB > 0 {
		opts.MaxBytes = int64(cfg.LogMaxMB) * 1024 * 1024
	}
	if cfg.LogFileMB > 0 {
		opts.FileBytes = int64(cfg.LogFileMB) * 1024 * 1024
	}
	if cfg.LogQueueSize > 0 {
		opts.QueueSize = cfg.LogQueueSize
	}
	store, err := locallog.New(opts)
	if err != nil {
		runtime.Status = "storage-unavailable"
		runtime.Recorder = operationlog.New(operationlog.SinkFunc(func(operationlog.Event) error { return locallog.ErrStorage }), "api", cfg.Environment, cfg.WebSurface)
		return runtime, nil
	}
	runtime.Store, runtime.Status = store, "ready"
	runtime.Recorder = operationlog.New(store, "api", cfg.Environment, cfg.WebSurface)
	return runtime, nil
}
func NewOperationLogging(cfg config.Config, lifecycle fx.Lifecycle) (*OperationLogging, error) {
	runtime, err := buildOperationLogging(cfg)
	if err != nil {
		return nil, err
	}
	if lifecycle != nil && runtime.Store != nil {
		lifecycle.Append(fx.Hook{OnStop: runtime.Store.Close})
	}
	return runtime, nil
}
func NewManagedHTTPServer(cfg config.Config, lifecycle fx.Lifecycle, logs *OperationLogging) (*http.Server, error) {
	return newHTTPServer(cfg, lifecycle, logs)
}

type loggedHandler struct {
	http.Handler
	logs *OperationLogging
}

func (h *loggedHandler) Close(ctx context.Context) error {
	if h.logs.Store != nil {
		return h.logs.Store.Close(ctx)
	}
	return nil
}
