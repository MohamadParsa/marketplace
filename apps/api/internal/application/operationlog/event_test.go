package operationlog

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
)

func TestEverySinkReceivesOnlyAllowlistedData(t *testing.T) {
	var got Event
	r := New(SinkFunc(func(e Event) error { got = e; return nil }), "api", "test", "storefront")
	ctx := WithContext(context.Background(), ContextData{RequestID: "request-1", OperationID: "operation-1", ParentOperationID: "parent-1"})
	SetActor(ctx, "customer", "user-1", "")
	secret := "NEVER-STORE-PRIVATE-VALUE"
	r.Record(ctx, Event{EventName: "checkout.completed", Type: "business", Severity: "info", Outcome: "succeeded", Attributes: map[string]any{
		"password": secret, "client_secret": secret, "authorization": secret, "cookie": secret, "token": secret, "address": secret, "error": errors.New(secret), "body": map[string]any{"quantity": secret}, "reason": secret, "quantity": 2, "offer_id": "offer-1", "device": "mobile", "amount": "100.5", "currency": "IRT",
	}})
	b, _ := json.Marshal(got)
	if strings.Contains(string(b), secret) || len(got.Attributes) != 5 || got.ActorID != "user-1" || got.OperationID != "operation-1" || got.RequestID != "request-1" || got.ParentOperationID != "parent-1" || got.SchemaVersion != 1 || got.RecordedAt.IsZero() {
		t.Fatalf("unsafe or incomplete event: %s", b)
	}
	if r.Stats().Recorded != 1 {
		t.Fatal(r.Stats())
	}
}
func TestInvalidEventsAndFailedSinkCannotLeakOrPanic(t *testing.T) {
	r := New(SinkFunc(func(Event) error { return errors.New("private disk path") }), "api", "test", "backoffice")
	r.Record(context.Background(), Event{EventName: "bad\nname", Type: "technical", Severity: "info", Outcome: "succeeded"})
	r.Record(context.Background(), Event{EventName: "safe.event", Type: "technical", Severity: "info", Outcome: "unknown", Attributes: map[string]any{"offer_id": "bad\npath"}})
	if r.Stats().Rejected != 1 || r.Stats().WriteFailures != 1 || r.Stats().Recorded != 0 {
		t.Fatal(r.Stats())
	}
}
func TestContextWorkersAndConcurrentRecorder(t *testing.T) {
	ctx := Ensure(context.Background())
	root := Context(ctx)
	worker := WithContext(context.Background(), ContextData{OperationID: root.OperationID, ParentOperationID: root.OperationID})
	if Context(worker).OperationID != root.OperationID {
		t.Fatal("worker lost operation")
	}
	r := New(SinkFunc(func(Event) error { return nil }), "api", "test", "storefront")
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			SetActor(ctx, "customer", "actor-1", "")
			r.Record(ctx, Event{EventName: "cart.changed", Type: "behavior", Severity: "info", Outcome: "succeeded"})
		}()
	}
	wg.Wait()
	if r.Stats().Recorded != 40 {
		t.Fatal(r.Stats())
	}
}
func TestRuntimeLoggerUsesSameSanitization(t *testing.T) {
	var got Event
	r := New(SinkFunc(func(e Event) error { got = e; return nil }), "api", "test", "backoffice")
	slog.New(SlogHandler{Recorder: r}).With("password", "PRIVATE").Error("PRIVATE", "error", errors.New("PRIVATE"))
	b, _ := json.Marshal(got)
	if strings.Contains(string(b), "PRIVATE") || got.EventName != "runtime.message" || got.Severity != "error" {
		t.Fatal(string(b))
	}
}
