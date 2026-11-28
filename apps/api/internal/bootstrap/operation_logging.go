package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
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
	unavailable := func() (*OperationLogging, error) {
		runtime.Status = "storage-unavailable"
		runtime.Recorder = operationlog.New(operationlog.SinkFunc(func(operationlog.Event) error { return locallog.ErrStorage }), "api", cfg.Environment, cfg.WebSurface)
		runtime.Recorder.ConfigureBehaviorRetention(cfg.BehaviorRetentionDays)
		return runtime, nil
	}
	dir, err := resolvedLogPath(cfg.OperationLogDir)
	if err != nil {
		return unavailable()
	}
	if cfg.WebRoot != "" {
		root, e := resolvedLogPath(cfg.WebRoot)
		if e != nil {
			return unavailable()
		}
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
		return unavailable()
	}
	runtime.Store, runtime.Status = store, "ready"
	runtime.Recorder = operationlog.New(store, "api", cfg.Environment, cfg.WebSurface)
	runtime.Recorder.ConfigureBehaviorRetention(opts.BehaviorDays)
	return runtime, nil
}

// Resolve existing ancestors as well: the final log directory may not exist yet,
// and an intermediate junction/symlink must not bypass the public-root check.
func resolvedLogPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	current := abs
	for {
		if _, err = os.Lstat(current); err == nil {
			resolved, err := existingLogPath(current)
			if err != nil {
				return "", err
			}
			rel, err := filepath.Rel(current, abs)
			if err != nil {
				return "", err
			}
			return filepath.Join(resolved, rel), nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return abs, nil
		}
		current = parent
	}
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
func (h *loggedHandler) FlushLogs(ctx context.Context) error {
	if h.logs.Store != nil {
		return h.logs.Store.Flush(ctx)
	}
	return nil
}
