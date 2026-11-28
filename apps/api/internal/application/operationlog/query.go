package operationlog

import (
	"container/heap"
	"context"
	"encoding/base64"
	"encoding/json"
	"sort"
	"time"
)

type Query struct {
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
	Type        string    `json:"log_type"`
	Severity    string    `json:"severity"`
	Outcome     string    `json:"outcome"`
	EventName   string    `json:"event_name"`
	OperationID string    `json:"operation_id"`
	ActorID     string    `json:"actor_id"`
	EntityID    string    `json:"entity_id"`
	Device      string    `json:"device"`
	CategoryID  string    `json:"category_id"`
	Environment string    `json:"environment"`
	Limit       int       `json:"limit"`
	Cursor      string    `json:"cursor"`
}
type Scanner interface {
	Scan(context.Context, Query, func(Event)) error
}

func (q Query) Validate() error {
	if !q.From.IsZero() && !q.To.IsZero() && q.From.After(q.To) {
		return ErrEvent
	}
	if q.Limit < 0 || q.Limit > 200 || len(q.Cursor) > 512 {
		return ErrEvent
	}
	if q.Type != "" && !member(q.Type, "technical", "integration", "business", "security", "behavior") {
		return ErrEvent
	}
	if q.Severity != "" && !member(q.Severity, "debug", "info", "warn", "error") {
		return ErrEvent
	}
	if q.Outcome != "" && !member(q.Outcome, "started", "succeeded", "rejected", "failed", "pending", "unknown", "cancelled", "replayed") {
		return ErrEvent
	}
	for _, v := range []string{q.OperationID, q.ActorID, q.EntityID, q.CategoryID} {
		if v != "" && !validID(v) {
			return ErrEvent
		}
	}
	if q.EventName != "" && safeCode(q.EventName) == "" {
		return ErrEvent
	}
	if q.Environment != "" && safeCode(q.Environment) == "" {
		return ErrEvent
	}
	if q.Device != "" && !member(q.Device, "desktop", "mobile", "tablet", "unknown") {
		return ErrEvent
	}
	if q.Cursor != "" {
		if _, err := decodeCursor(q.Cursor); err != nil {
			return err
		}
	}
	return nil
}
func (q Query) Match(e Event) bool {
	if !q.From.IsZero() && e.RecordedAt.Before(q.From) || !q.To.IsZero() && e.RecordedAt.After(q.To) {
		return false
	}
	if q.Type != "" && e.Type != q.Type || q.Severity != "" && e.Severity != q.Severity || q.Outcome != "" && e.Outcome != q.Outcome || q.EventName != "" && e.EventName != q.EventName || q.OperationID != "" && e.OperationID != q.OperationID || q.ActorID != "" && e.ActorID != q.ActorID || q.EntityID != "" && e.EntityID != q.EntityID || q.Environment != "" && e.Environment != q.Environment {
		return false
	}
	if q.Device != "" && e.Attributes["device"] != q.Device {
		return false
	}
	if q.CategoryID != "" && e.Attributes["category_id"] != q.CategoryID {
		found := false
		if ids, ok := e.Attributes["category_ids"].([]string); ok {
			for _, id := range ids {
				if id == q.CategoryID {
					found = true
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type Page struct {
	Items      []Event `json:"items"`
	NextCursor string  `json:"next_cursor,omitempty"`
}
type cursor struct {
	At time.Time `json:"at"`
	ID string    `json:"id"`
}

func CursorFor(at time.Time, id string) string {
	b, _ := json.Marshal(cursor{at, id})
	return base64.RawURLEncoding.EncodeToString(b)
}
func BeforeCursor(at time.Time, id, value string) (bool, error) {
	if value == "" {
		return true, nil
	}
	c, err := decodeCursor(value)
	if err != nil {
		return false, err
	}
	return at.Before(c.At) || at.Equal(c.At) && id < c.ID, nil
}

func decodeCursor(value string) (cursor, error) {
	var c cursor
	b, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || json.Unmarshal(b, &c) != nil || c.At.IsZero() || !validID(c.ID) {
		return c, ErrEvent
	}
	return c, nil
}
func newer(a, b Event) bool {
	if a.RecordedAt.Equal(b.RecordedAt) {
		return a.EventID > b.EventID
	}
	return a.RecordedAt.After(b.RecordedAt)
}

type eventsHeap []Event

func (h eventsHeap) Len() int           { return len(h) }
func (h eventsHeap) Less(i, j int) bool { return newer(h[j], h[i]) }
func (h eventsHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *eventsHeap) Push(x any)        { *h = append(*h, x.(Event)) }
func (h *eventsHeap) Pop() any          { old := *h; n := len(old); x := old[n-1]; *h = old[:n-1]; return x }
func Read(ctx context.Context, s Scanner, q Query) (Page, error) {
	if err := q.Validate(); err != nil {
		return Page{}, err
	}
	limit := q.Limit
	if limit == 0 {
		limit = 100
	}
	var after cursor
	if q.Cursor != "" {
		after, _ = decodeCursor(q.Cursor)
	}
	h := &eventsHeap{}
	heap.Init(h)
	err := s.Scan(ctx, q, func(e Event) {
		if !after.At.IsZero() && (!e.RecordedAt.Before(after.At) && !(e.RecordedAt.Equal(after.At) && e.EventID < after.ID)) {
			return
		}
		heap.Push(h, e)
		if h.Len() > limit+1 {
			heap.Pop(h)
		}
	})
	if err != nil {
		return Page{}, err
	}
	rows := []Event(*h)
	sort.Slice(rows, func(i, j int) bool { return newer(rows[i], rows[j]) })
	out := Page{Items: rows}
	if len(rows) > limit {
		out.Items = rows[:limit]
		last := out.Items[limit-1]
		b, _ := json.Marshal(cursor{last.RecordedAt, last.EventID})
		out.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return out, nil
}
