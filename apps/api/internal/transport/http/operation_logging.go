package httptransport

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
)

type loggedResponse struct {
	http.ResponseWriter
	status int
}

func (w *loggedResponse) WriteHeader(n int) {
	if w.status == 0 {
		w.status = n
		w.ResponseWriter.WriteHeader(n)
	}
}
func (w *loggedResponse) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
func (w *loggedResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }

var operationHeader = regexp.MustCompile(`^[a-f0-9]{32}$`)

// OperationMiddleware records metadata only. Request/response bodies, paths,
// query strings, cookies and headers are never dumped into an event.
func OperationMiddleware(recorder *operationlog.Recorder, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if recorder == nil || strings.HasPrefix(r.URL.Path, "/assets/") || strings.HasPrefix(r.URL.Path, "/health") {
			next.ServeHTTP(w, r)
			return
		}
		op := r.Header.Get("X-Operation-ID")
		if !operationHeader.MatchString(op) {
			op = operationlog.ID()
		}
		ctx := operationlog.WithRecorder(operationlog.WithContext(r.Context(), operationlog.ContextData{RequestID: operationlog.ID(), OperationID: op}), recorder)
		r = r.WithContext(ctx)
		trace := operationlog.Context(ctx)
		w.Header().Set("X-Request-ID", trace.RequestID)
		w.Header().Set("X-Operation-ID", trace.OperationID)
		typ := "technical"
		if strings.Contains(r.URL.Path, "/auth/") {
			typ = "security"
		}
		e := operationlog.Event{EventName: "http.request-started", Type: typ, Severity: "info", Outcome: "started"}
		recorder.Record(ctx, e)
		started := time.Now()
		out := &loggedResponse{ResponseWriter: w}
		defer func() {
			if recovered := recover(); recovered != nil {
				if out.status == 0 {
					writeError(out, 500, "internal_error")
				}
				e.Outcome = "failed"
				e.ErrorCode = "panic_recovered"
				e.Severity = "error"
			} else if out.status >= 500 {
				e.Outcome = "unknown"
				e.ErrorCode = "server_error"
				e.Severity = "error"
			} else if out.status >= 400 {
				e.Outcome = "rejected"
				e.ErrorCode = "request_rejected"
				e.Severity = "warn"
			} else {
				e.Outcome = "succeeded"
			}
			if out.status == 0 {
				out.status = 200
			}
			e.EventName = "http.request-finished"
			e.HTTPStatus = out.status
			e.DurationMS = time.Since(started).Milliseconds()
			e.Attributes = map[string]any{"method": r.Method, "route": r.Pattern, "device": requestDevice(r)}
			recorder.Record(ctx, e)
		}()
		next.ServeHTTP(out, r)
	})
}
func requestDevice(r *http.Request) string {
	ua := strings.ToLower(r.UserAgent())
	if strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet") {
		return "tablet"
	}
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") {
		return "mobile"
	}
	if ua == "" {
		return "unknown"
	}
	return "desktop"
}
