package locallog

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
)

// Reader never creates, deletes, rotates or appends files.
type Reader struct{ store *Store }

func OpenReader(opts Options) (*Reader, error) {
	if !filepath.IsAbs(opts.Directory) || opts.RetentionDays < 1 || opts.BehaviorDays < 1 {
		return nil, ErrStorage
	}
	info, err := os.Lstat(opts.Directory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, ErrStorage
	}
	if opts.Now == nil {
		opts.Now = DefaultOptions(opts.Directory).Now
	}
	return &Reader{store: &Store{options: opts}}, nil
}
func (r *Reader) Scan(ctx context.Context, q operationlog.Query, visit func(operationlog.Event)) error {
	return r.store.Scan(ctx, q, visit)
}
