package locallog

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
)

func testStore(t *testing.T, opts Options) *Store {
	t.Helper()
	s, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	return s
}
func record(s *Store, typ string) *operationlog.Recorder {
	return operationlog.New(s, "api", "test", "storefront")
}
func emit(r *operationlog.Recorder, typ string) {
	r.Record(operationlog.Ensure(context.Background()), operationlog.Event{EventName: "cart.changed", Type: typ, Severity: "info", Outcome: "succeeded", Attributes: map[string]any{"quantity": 2, "password": "PRIVATE"}})
}
func flush(t *testing.T, s *Store) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Flush(ctx); err != nil {
		t.Fatal(err)
	}
}
func TestLocalFilesRestartRotationPrivacyAndTwoWriters(t *testing.T) {
	opts := DefaultOptions(t.TempDir())
	opts.FileBytes = 1400
	opts.MaxBytes = 15000
	opts.QueueSize = 100
	a, b := testStore(t, opts), testStore(t, opts)
	var wg sync.WaitGroup
	for _, s := range []*Store{a, b} {
		wg.Add(1)
		go func(s *Store) {
			defer wg.Done()
			r := record(s, "technical")
			for i := 0; i < 10; i++ {
				emit(r, "technical")
			}
		}(s)
	}
	wg.Wait()
	flush(t, a)
	flush(t, b)
	if a.Stats().Written+b.Stats().Written != 20 || a.Stats().WriteFailures+b.Stats().WriteFailures != 0 {
		t.Fatal(a.Stats(), b.Stats())
	}
	files, err := a.inventory()
	if err != nil || len(files) < 2 {
		t.Fatal("no rotation", err)
	}
	var total int64
	for _, file := range files {
		total += file.size
		raw, err := os.ReadFile(filepath.Join(opts.Directory, file.name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "PRIVATE") {
			t.Fatal("leak")
		}
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			if !json.Valid([]byte(line)) {
				t.Fatal("torn line")
			}
		}
	}
	if total > opts.MaxBytes {
		t.Fatal("budget exceeded")
	}
	reopened := testStore(t, opts)
	page, err := operationlog.Read(context.Background(), reopened, operationlog.Query{Limit: 5})
	if err != nil || len(page.Items) != 5 || page.NextCursor == "" {
		t.Fatal(page, err)
	}
	next, err := operationlog.Read(context.Background(), reopened, operationlog.Query{Limit: 5, Cursor: page.NextCursor})
	if err != nil || len(next.Items) != 5 {
		t.Fatal(next, err)
	}
	ids := map[string]bool{}
	for _, e := range page.Items {
		ids[e.EventID] = true
	}
	for _, e := range next.Items {
		if ids[e.EventID] {
			t.Fatal("pagination repeated event")
		}
	}
}
func TestRetentionAndSharedDiskBudget(t *testing.T) {
	opts := DefaultOptions(t.TempDir())
	opts.QueueSize = 200
	opts.FileBytes = 2000
	opts.MaxBytes = 6000
	s := testStore(t, opts)
	r := record(s, "technical")
	emit(r, "technical")
	emit(r, "behavior")
	flush(t, s)
	files, _ := s.inventory()
	old := time.Now().Add(-10 * 24 * time.Hour)
	for _, f := range files {
		if err := os.Chtimes(filepath.Join(opts.Directory, f.name), old, old); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.withLock(func() error { return s.prune(0) }); err != nil {
		t.Fatal(err)
	}
	files, _ = s.inventory()
	if len(files) != 1 || !strings.HasPrefix(files[0].name, "behavior-") {
		t.Fatal("retention not separated", files)
	}
	for i := 0; i < 100; i++ {
		emit(r, "technical")
	}
	flush(t, s)
	files, _ = s.inventory()
	var total int64
	for _, f := range files {
		total += f.size
	}
	if total > opts.MaxBytes {
		t.Fatal(total)
	}
}
func TestQueueAndDiskFailureAreObservableAndShutdownFlushes(t *testing.T) {
	opts := DefaultOptions(t.TempDir())
	opts.QueueSize = 1
	s := testStore(t, opts)
	f, err := os.OpenFile(filepath.Join(opts.Directory, ".writer.lock"), os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err = lock(f); err != nil {
		t.Fatal(err)
	}
	r := record(s, "technical")
	for i := 0; i < 20; i++ {
		emit(r, "technical")
	}
	if s.Stats().Dropped == 0 || r.Stats().WriteFailures == 0 {
		t.Fatal("queue overflow not observable")
	}
	unlock(f)
	f.Close()
	flush(t, s)
	// Replace the private directory with a file to deterministically simulate
	// an unavailable disk/path on both Windows and Unix without relying on ACLs.
	moved := filepath.Join(t.TempDir(), "moved-logs")
	if err = os.Rename(opts.Directory, moved); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(opts.Directory, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	emit(r, "technical")
	flush(t, s)
	if s.Stats().WriteFailures == 0 {
		t.Fatal("disk failure hidden")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = s.Close(ctx); err != nil {
		t.Fatal(err)
	}
	emit(r, "technical")
	if s.Stats().Dropped == 0 {
		t.Fatal("closed writer accepted")
	}
}
func TestQueryValidationAndCancellation(t *testing.T) {
	s := testStore(t, DefaultOptions(t.TempDir()))
	if _, err := operationlog.Read(context.Background(), s, operationlog.Query{Limit: 201}); err == nil {
		t.Fatal("unbounded query")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	emit(record(s, "behavior"), "behavior")
	flush(t, s)
	if _, err := operationlog.Read(ctx, s, operationlog.Query{}); err == nil {
		t.Fatal("ignored cancellation")
	}
}
