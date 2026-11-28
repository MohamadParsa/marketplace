package operationlog

import (
	"context"
	"log/slog"
)

// SlogHandler routes the existing runtime logger through the same safe recorder.
// Runtime messages have fixed codes; arbitrary text/errors are never persisted.
type SlogHandler struct {
	Recorder   *Recorder
	attributes []slog.Attr
}

func (h SlogHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h SlogHandler) Handle(ctx context.Context, r slog.Record) error {
	name := "runtime.message"
	switch r.Message {
	case "public cache statistics":
		name = "runtime.cache-statistics"
	case "public cache warmup":
		name = "runtime.cache-warmup"
	case "recent history cleanup":
		name = "runtime.recent-history-cleanup"
	case "recent history cleanup failed":
		name = "runtime.recent-history-cleanup-failed"
	}
	level := "info"
	outcome := "succeeded"
	if r.Level >= slog.LevelError {
		level = "error"
		outcome = "failed"
	} else if r.Level >= slog.LevelWarn {
		level = "warn"
	} else if r.Level < slog.LevelInfo {
		level = "debug"
	}
	attrs := map[string]any{}
	add := func(a slog.Attr) bool { attrs[a.Key] = a.Value.Any(); return true }
	for _, a := range h.attributes {
		add(a)
	}
	r.Attrs(add)
	h.Recorder.Record(ctx, Event{EventName: name, Type: "technical", Severity: level, Outcome: outcome, Attributes: attrs})
	return nil
}
func (h SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.attributes = append(append([]slog.Attr{}, h.attributes...), attrs...)
	return h
}
func (h SlogHandler) WithGroup(string) slog.Handler { return h }
