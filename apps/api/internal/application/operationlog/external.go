package operationlog

import (
	"context"
	"errors"
	"net"
	"time"
)

// BeginExternal records one actual provider attempt, including parsing its
// semantic result. End receives only a stable provider code, never body/URL.
func BeginExternal(ctx context.Context, provider, operation string, attempt int) func(int, string, error) {
	ctx = Ensure(ctx)
	started := time.Now()
	base := Event{EventName: "external.call-started", Type: "integration", Severity: "info", Outcome: "started", Provider: provider, ProviderOperation: operation, AttemptNumber: attempt}
	Record(ctx, base)
	return func(status int, code string, err error) {
		base.EventName = "external.call-finished"
		base.HTTPStatus = status
		base.DurationMS = time.Since(started).Milliseconds()
		base.ErrorCode = code
		base.Outcome = "succeeded"
		if err != nil {
			base.Outcome = "failed"
			base.Severity = "error"
			if base.ErrorCode == "" {
				base.ErrorCode = "provider_failed"
			}
			var netErr net.Error
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() {
				base.Outcome = "unknown"
				base.ErrorCode = "provider_timeout"
			}
		}
		Record(ctx, base)
	}
}
