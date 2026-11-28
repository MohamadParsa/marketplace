// Package operationlog defines the private, provider-independent event contract.
// All sinks receive allowlisted data; callers never submit raw HTTP payloads.
package operationlog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Event struct {
	SchemaVersion     int            `json:"schema_version"`
	EventID           string         `json:"event_id"`
	EventName         string         `json:"event_name"`
	Type              string         `json:"log_type"`
	Severity          string         `json:"severity"`
	Outcome           string         `json:"outcome"`
	Stage             string         `json:"stage,omitempty"`
	OccurredAt        time.Time      `json:"occurred_at"`
	RecordedAt        time.Time      `json:"recorded_at"`
	Service           string         `json:"service"`
	InstanceID        string         `json:"instance_id"`
	Environment       string         `json:"environment"`
	Surface           string         `json:"surface"`
	RequestID         string         `json:"request_id,omitempty"`
	OperationID       string         `json:"operation_id"`
	ParentOperationID string         `json:"parent_operation_id,omitempty"`
	ActorKind         string         `json:"actor_kind,omitempty"`
	ActorID           string         `json:"actor_id,omitempty"`
	StoreID           string         `json:"store_id,omitempty"`
	SessionRef        string         `json:"session_ref,omitempty"`
	EntityType        string         `json:"entity_type,omitempty"`
	EntityID          string         `json:"entity_id,omitempty"`
	PurchaseID        string         `json:"purchase_id,omitempty"`
	OrderID           string         `json:"order_id,omitempty"`
	PaymentAttemptID  string         `json:"payment_attempt_id,omitempty"`
	Provider          string         `json:"provider,omitempty"`
	ProviderOperation string         `json:"provider_operation,omitempty"`
	AttemptNumber     int            `json:"attempt_number,omitempty"`
	DurationMS        int64          `json:"duration_ms,omitempty"`
	HTTPStatus        int            `json:"http_status,omitempty"`
	ErrorCode         string         `json:"error_code,omitempty"`
	Attributes        map[string]any `json:"attributes,omitempty"`
}

type Sink interface{ Write(Event) error }
type SinkFunc func(Event) error

func (f SinkFunc) Write(e Event) error { return f(e) }

type ContextData struct {
	RequestID, OperationID, ParentOperationID string
	ActorKind, ActorID, StoreID, SessionRef   string
}
type trace struct {
	mu   sync.Mutex
	data ContextData
}
type contextKey struct{}
type recorderKey struct{}

func WithRecorder(ctx context.Context, r *Recorder) context.Context {
	return context.WithValue(ctx, recorderKey{}, r)
}
func FromContext(ctx context.Context) *Recorder {
	r, _ := ctx.Value(recorderKey{}).(*Recorder)
	return r
}
func Record(ctx context.Context, e Event) { FromContext(ctx).Record(ctx, e) }

// Child keeps the operation while giving a background task an explicit parent.
func Child(ctx context.Context) context.Context {
	d := Context(ctx)
	d.ParentOperationID = d.OperationID
	return WithRecorder(WithContext(context.Background(), d), FromContext(ctx))
}

func ID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("operation log identifier unavailable")
	}
	return hex.EncodeToString(b[:])
}
func WithContext(ctx context.Context, d ContextData) context.Context {
	if !validID(d.OperationID) {
		d.OperationID = ID()
	}
	if d.RequestID != "" && !validID(d.RequestID) {
		d.RequestID = ID()
	}
	d.ParentOperationID = safeID(d.ParentOperationID)
	return context.WithValue(ctx, contextKey{}, &trace{data: d})
}
func Ensure(ctx context.Context) context.Context {
	if _, ok := ctx.Value(contextKey{}).(*trace); ok {
		return ctx
	}
	return WithContext(ctx, ContextData{})
}
func Context(ctx context.Context) ContextData {
	if t, ok := ctx.Value(contextKey{}).(*trace); ok {
		t.mu.Lock()
		defer t.mu.Unlock()
		return t.data
	}
	return ContextData{}
}

// SetActor is called only after server-side authentication/authorization.
func SetActor(ctx context.Context, kind, actor, store string) {
	if t, ok := ctx.Value(contextKey{}).(*trace); ok {
		t.mu.Lock()
		defer t.mu.Unlock()
		t.data.ActorKind = safeCode(kind)
		t.data.ActorID = safeID(actor)
		t.data.StoreID = safeID(store)
	}
}
func SetSession(ctx context.Context, ref string) {
	if t, ok := ctx.Value(contextKey{}).(*trace); ok {
		t.mu.Lock()
		defer t.mu.Unlock()
		t.data.SessionRef = safeID(ref)
	}
}

type Stats struct {
	Recorded      uint64 `json:"recorded"`
	Rejected      uint64 `json:"rejected"`
	WriteFailures uint64 `json:"write_failures"`
}
type Recorder struct {
	sink                                    Sink
	service, environment, surface, instance string
	recorded, rejected, failures            atomic.Uint64
}

func New(sink Sink, service, environment, surface string) *Recorder {
	return &Recorder{sink: sink, service: safeCode(service), environment: safeCode(environment), surface: safeCode(surface), instance: ID()}
}
func (r *Recorder) Stats() Stats {
	if r == nil {
		return Stats{}
	}
	return Stats{r.recorded.Load(), r.rejected.Load(), r.failures.Load()}
}
func (r *Recorder) Record(ctx context.Context, e Event) {
	if r == nil || r.sink == nil {
		return
	}
	d := Context(ctx)
	e.RequestID, e.OperationID, e.ParentOperationID = d.RequestID, d.OperationID, d.ParentOperationID
	e.ActorKind, e.ActorID, e.StoreID, e.SessionRef = d.ActorKind, d.ActorID, d.StoreID, d.SessionRef
	if e.OperationID == "" {
		e.OperationID = ID()
	}
	e.SchemaVersion, e.EventID = 1, ID()
	e.Service, e.Environment, e.Surface, e.InstanceID = r.service, r.environment, r.surface, r.instance
	e.RecordedAt = time.Now().UTC()
	if e.OccurredAt.IsZero() {
		e.OccurredAt = e.RecordedAt
	} else {
		e.OccurredAt = e.OccurredAt.UTC()
	}
	e, err := Sanitize(e)
	if err != nil {
		r.rejected.Add(1)
		return
	}
	if err = r.sink.Write(e); err != nil {
		r.failures.Add(1)
		return
	}
	r.recorded.Add(1)
}

var codePattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,95}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)
var routePattern = regexp.MustCompile(`^(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) /[A-Za-z0-9/{}._-]*$`)

func validID(v string) bool { return idPattern.MatchString(v) }
func safeID(v string) string {
	if validID(v) {
		return v
	}
	return ""
}
func safeCode(v string) string {
	if codePattern.MatchString(v) {
		return v
	}
	return ""
}
func member(v string, choices ...string) bool {
	for _, c := range choices {
		if v == c {
			return true
		}
	}
	return false
}

var ErrEvent = errors.New("invalid_operation_event")

// Sanitize deliberately drops unknown/nested/free-text attributes. It runs
// before every sink (including test/future sinks), never after persistence.
func Sanitize(e Event) (Event, error) {
	if safeCode(e.EventName) == "" || !member(e.Type, "technical", "integration", "business", "security", "behavior") || !member(e.Severity, "debug", "info", "warn", "error") || !member(e.Outcome, "started", "succeeded", "rejected", "failed", "pending", "unknown", "cancelled", "replayed") {
		return Event{}, ErrEvent
	}
	e.Stage, e.EntityType, e.Provider, e.ProviderOperation, e.ErrorCode = safeCode(e.Stage), safeCode(e.EntityType), safeCode(e.Provider), safeCode(e.ProviderOperation), safeCode(e.ErrorCode)
	e.ActorKind = safeCode(e.ActorKind)
	for _, p := range []*string{&e.EventID, &e.InstanceID, &e.RequestID, &e.OperationID, &e.ParentOperationID, &e.ActorID, &e.StoreID, &e.SessionRef, &e.EntityID, &e.PurchaseID, &e.OrderID, &e.PaymentAttemptID} {
		*p = safeID(*p)
	}
	if e.DurationMS < 0 {
		e.DurationMS = 0
	}
	if e.AttemptNumber < 0 {
		e.AttemptNumber = 0
	}
	if e.HTTPStatus < 100 || e.HTTPStatus > 599 {
		e.HTTPStatus = 0
	}
	out := map[string]any{}
	for k, v := range e.Attributes {
		switch k {
		case "quantity", "item_count", "result_count", "version", "previous_version", "changed_fields", "retry_count", "status_code":
			switch n := v.(type) {
			case int:
				out[k] = n
			case int64:
				out[k] = n
			case float64:
				if n >= 0 && n < 1e12 && n == float64(int64(n)) {
					out[k] = int64(n)
				}
			}
		case "product_id", "offer_id", "category_id", "cart_id", "quote_id", "shipment_id", "audit_id":
			if s, ok := v.(string); ok && validID(s) {
				out[k] = s
			}
		case "device":
			if s, ok := v.(string); ok && member(s, "desktop", "mobile", "tablet", "unknown") {
				out[k] = s
			}
		case "method":
			if s, ok := v.(string); ok && member(s, "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS") {
				out[k] = s
			}
		case "route":
			if s, ok := v.(string); ok && len(s) <= 200 && routePattern.MatchString(s) {
				out[k] = s
			}
		case "currency":
			if s, ok := v.(string); ok && member(s, "IRT", "IRR") {
				out[k] = s
			}
		case "replayed", "retryable", "withdrawable":
			if b, ok := v.(bool); ok {
				out[k] = b
			}
		case "amount":
			if s, ok := v.(string); ok && len(s) <= 20 && s != "" && strings.Trim(s, "0123456789.") == "" {
				out[k] = s
			}
		}
	}
	e.Attributes = out
	return e, nil
}
