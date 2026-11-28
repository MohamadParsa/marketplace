// Package locallog stores private operation events as bounded JSON Lines files.
package locallog

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
)

var ErrStorage = errors.New("local_log_storage_unavailable")
var ErrFull = errors.New("local_log_queue_full")

type Options struct {
	Directory                   string
	RetentionDays, BehaviorDays int
	MaxBytes, FileBytes         int64
	QueueSize                   int
	Now                         func() time.Time
}

func DefaultOptions(dir string) Options {
	return Options{Directory: dir, RetentionDays: 7, BehaviorDays: 30, MaxBytes: 500 * 1024 * 1024, FileBytes: 8 * 1024 * 1024, QueueSize: 1024, Now: time.Now}
}

type Stats struct {
	Written       uint64 `json:"written"`
	Dropped       uint64 `json:"dropped"`
	WriteFailures uint64 `json:"write_failures"`
	CorruptLines  uint64 `json:"corrupt_lines"`
	Queued        int    `json:"queued"`
}
type entry struct {
	event   *operationlog.Event
	barrier chan struct{}
}
type Store struct {
	options                             Options
	instance                            string
	queue                               chan entry
	mu                                  sync.RWMutex
	closed                              bool
	done                                chan struct{}
	written, dropped, failures, corrupt atomic.Uint64
	sequence                            uint64
	files                               map[string]string
}

func New(opts Options) (*Store, error) {
	if opts.RetentionDays < 1 || opts.RetentionDays > 3650 || opts.BehaviorDays < 1 || opts.BehaviorDays > 3650 || opts.MaxBytes < 1024 || opts.FileBytes < 512 || opts.FileBytes > opts.MaxBytes || opts.QueueSize < 1 || opts.QueueSize > 65536 || !filepath.IsAbs(opts.Directory) {
		return nil, ErrStorage
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if err := os.MkdirAll(opts.Directory, 0700); err != nil {
		return nil, ErrStorage
	}
	info, err := os.Lstat(opts.Directory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, ErrStorage
	}
	if err = os.Chmod(opts.Directory, 0700); err != nil {
		return nil, ErrStorage
	}
	s := &Store{options: opts, instance: operationlog.ID(), queue: make(chan entry, opts.QueueSize), done: make(chan struct{}), files: map[string]string{}}
	if err = s.withLock(func() error { return s.prune(0) }); err != nil {
		return nil, err
	}
	go s.run()
	return s, nil
}
func (s *Store) Directory() string { return s.options.Directory }
func (s *Store) Stats() Stats {
	return Stats{s.written.Load(), s.dropped.Load(), s.failures.Load(), s.corrupt.Load(), len(s.queue)}
}
func (s *Store) Write(e operationlog.Event) error {
	e, err := operationlog.Sanitize(e)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		s.dropped.Add(1)
		return ErrStorage
	}
	select {
	case s.queue <- entry{event: &e}:
		return nil
	default:
		s.dropped.Add(1)
		return ErrFull
	}
}
func (s *Store) Flush(ctx context.Context) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		select {
		case <-s.done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	barrier := make(chan struct{})
	select {
	case s.queue <- entry{barrier: barrier}:
		s.mu.RUnlock()
	case <-ctx.Done():
		s.mu.RUnlock()
		return ctx.Err()
	}
	select {
	case <-barrier:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (s *Store) Close(ctx context.Context) error {
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		close(s.queue)
	}
	s.mu.Unlock()
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (s *Store) run() {
	defer close(s.done)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case item, ok := <-s.queue:
			if !ok {
				return
			}
			if item.barrier != nil {
				close(item.barrier)
				continue
			}
			if err := s.append(*item.event); err != nil {
				s.failures.Add(1)
			} else {
				s.written.Add(1)
			}
		case <-ticker.C:
			if err := s.withLock(func() error { return s.prune(0) }); err != nil {
				s.failures.Add(1)
			}
		}
	}
}
func (s *Store) withLock(fn func() error) error {
	f, err := os.OpenFile(filepath.Join(s.options.Directory, ".writer.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return ErrStorage
	}
	defer f.Close()
	if err = lock(f); err != nil {
		return ErrStorage
	}
	defer unlock(f)
	return fn()
}

var filePattern = regexp.MustCompile(`^(technical|integration|business|security|behavior)-[0-9]{8}-[a-f0-9]{32}-[0-9]+\.jsonl$`)

type logFile struct {
	name string
	size int64
	at   time.Time
}

func (s *Store) inventory() ([]logFile, error) {
	entries, err := os.ReadDir(s.options.Directory)
	if err != nil {
		return nil, ErrStorage
	}
	files := []logFile{}
	for _, e := range entries {
		if !filePattern.MatchString(e.Name()) || !e.Type().IsRegular() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, ErrStorage
		}
		files = append(files, logFile{e.Name(), info.Size(), info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].at.Equal(files[j].at) {
			return files[i].name < files[j].name
		}
		return files[i].at.Before(files[j].at)
	})
	return files, nil
}
func (s *Store) prune(incoming int64) error {
	files, err := s.inventory()
	if err != nil {
		return err
	}
	var total int64
	for i := range files {
		days := s.options.RetentionDays
		if strings.HasPrefix(files[i].name, "behavior-") {
			days = s.options.BehaviorDays
		}
		if files[i].at.Before(s.options.Now().Add(-time.Duration(days) * 24 * time.Hour)) {
			if err = os.Remove(filepath.Join(s.options.Directory, files[i].name)); err != nil && !os.IsNotExist(err) {
				return ErrStorage
			}
			files[i].size = 0
		}
		total += files[i].size
	}
	for _, f := range files {
		if total+incoming <= s.options.MaxBytes {
			break
		}
		if f.size == 0 {
			continue
		}
		if err = os.Remove(filepath.Join(s.options.Directory, f.name)); err != nil && !os.IsNotExist(err) {
			return ErrStorage
		}
		total -= f.size
	}
	if total+incoming > s.options.MaxBytes {
		return ErrStorage
	}
	return nil
}
func (s *Store) append(e operationlog.Event) error {
	b, err := json.Marshal(e)
	if err != nil || len(b) > 65535 {
		return ErrStorage
	}
	b = append(b, '\n')
	if int64(len(b)) > s.options.FileBytes {
		return ErrStorage
	}
	return s.withLock(func() error {
		if err := s.prune(int64(len(b))); err != nil {
			return err
		}
		day := s.options.Now().UTC().Format("20060102")
		key := e.Type + "-" + day
		name := s.files[key]
		if name != "" {
			info, err := os.Lstat(filepath.Join(s.options.Directory, name))
			if err != nil || !info.Mode().IsRegular() || info.Size()+int64(len(b)) > s.options.FileBytes {
				name = ""
			}
		}
		if name == "" {
			s.sequence++
			name = fmt.Sprintf("%s-%s-%d.jsonl", key, s.instance, s.sequence)
			if len(s.files) > 64 {
				s.files = map[string]string{}
			}
			s.files[key] = name
		}
		path := filepath.Join(s.options.Directory, name)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return ErrStorage
		}
		n, err := f.Write(b)
		if err == nil && n != len(b) {
			err = io.ErrShortWrite
		}
		if err == nil {
			err = f.Sync()
		}
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			return ErrStorage
		}
		s.files[key] = name
		return nil
	})
}

// Scan streams a bounded file snapshot. Private paths never leave this adapter.
// A partial final line from a concurrent writer is skipped, never exposed raw.
func (s *Store) Scan(ctx context.Context, q operationlog.Query, visit func(operationlog.Event)) error {
	if err := q.Validate(); err != nil {
		return err
	}
	files, err := s.inventory()
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := filepath.Join(s.options.Directory, file.name)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil || !info.Mode().IsRegular() {
			return ErrStorage
		}
		f, err := os.Open(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return ErrStorage
		}
		scanner := bufio.NewScanner(io.LimitReader(f, file.size))
		scanner.Buffer(make([]byte, 4096), 65536)
		for scanner.Scan() {
			if err := ctx.Err(); err != nil {
				f.Close()
				return err
			}
			var e operationlog.Event
			if json.Unmarshal(scanner.Bytes(), &e) != nil {
				s.corrupt.Add(1)
				continue
			}
			e, err = operationlog.Sanitize(e)
			if err != nil {
				s.corrupt.Add(1)
				continue
			}
			days := s.options.RetentionDays
			if e.Type == "behavior" {
				days = s.options.BehaviorDays
			}
			if e.RecordedAt.Before(s.options.Now().Add(-time.Duration(days) * 24 * time.Hour)) {
				continue
			}
			if q.Match(e) {
				visit(e)
			}
		}
		scanErr := scanner.Err()
		f.Close()
		if scanErr != nil {
			s.corrupt.Add(1)
		}
	}
	return nil
}
