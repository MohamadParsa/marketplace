package backoffice

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func (s *Service) logAdmin(ctx context.Context, a identity.User) (Workspace, error) {
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return w, err
	}
	current, err := currentActor(w, a)
	if err != nil || current.Surface != identity.SurfaceBackoffice || current.Role != identity.RoleSuperAdmin {
		return Workspace{}, ErrForbidden
	}
	return w, nil
}
func (s *Service) ReadOperationLogs(ctx context.Context, a identity.User, q operationlog.Query) (operationlog.Page, error) {
	if _, err := s.logAdmin(ctx, a); err != nil {
		return operationlog.Page{}, err
	}
	if q.Validate() != nil {
		return operationlog.Page{}, ErrInvalid
	}
	if s.operationScanner == nil {
		return operationlog.Page{Items: []operationlog.Event{}}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return operationlog.Read(ctx, s.operationScanner, q)
}
func (s *Service) ReadOperationLogStats(ctx context.Context, a identity.User) (any, error) {
	if _, err := s.logAdmin(ctx, a); err != nil {
		return nil, err
	}
	if s.operationStats == nil {
		return map[string]any{"status": "disabled"}, nil
	}
	return s.operationStats(), nil
}

type AuditReportRow struct {
	ID            string            `json:"id"`
	Actor         string            `json:"actor"`
	Action        string            `json:"action"`
	Target        string            `json:"target"`
	StoreID       string            `json:"store_id,omitempty"`
	OperationID   string            `json:"operation_id,omitempty"`
	RequestID     string            `json:"request_id,omitempty"`
	Outcome       string            `json:"outcome"`
	Version       int               `json:"version,omitempty"`
	Before        map[string]string `json:"before"`
	After         map[string]string `json:"after"`
	ReasonPresent bool              `json:"reason_present"`
	At            time.Time         `json:"at"`
}
type AuditReport struct {
	Items      []AuditReportRow `json:"items"`
	NextCursor string           `json:"next_cursor,omitempty"`
}

func (s *Service) ReadAuditReport(ctx context.Context, a identity.User, q operationlog.Query, store string) (AuditReport, error) {
	w, err := s.logAdmin(ctx, a)
	if err != nil {
		return AuditReport{}, err
	}
	if q.Validate() != nil {
		return AuditReport{}, ErrInvalid
	}
	limit := q.Limit
	if limit == 0 {
		limit = 100
	}
	rows := []AuditReportRow{}
	for _, v := range w.Audit {
		older, cursorErr := operationlog.BeforeCursor(v.At, auditPublicID(v.ID), q.Cursor)
		if cursorErr != nil {
			return AuditReport{}, ErrInvalid
		}
		if !older {
			continue
		}
		if !q.From.IsZero() && v.At.Before(q.From) || !q.To.IsZero() && v.At.After(q.To) || q.ActorID != "" && q.ActorID != v.Actor || q.OperationID != "" && q.OperationID != v.OperationID || q.EntityID != "" && q.EntityID != v.Target || q.EventName != "" && !strings.Contains(v.Action, q.EventName) || store != "" && store != v.StoreID {
			continue
		}
		outcome := v.Outcome
		if outcome == "" {
			outcome = "succeeded"
		}
		if q.Outcome != "" && q.Outcome != outcome {
			continue
		}
		rows = append(rows, AuditReportRow{ID: auditPublicID(v.ID), Actor: v.Actor, Action: v.Action, Target: v.Target, StoreID: v.StoreID, OperationID: v.OperationID, RequestID: v.RequestID, Outcome: outcome, Version: v.Version, Before: safeAuditProjection(v.Before), After: safeAuditProjection(v.After), ReasonPresent: v.Reason != "", At: v.At})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].At.Equal(rows[j].At) {
			return rows[i].ID > rows[j].ID
		}
		return rows[i].At.After(rows[j].At)
	})
	out := AuditReport{Items: rows}
	if len(rows) > limit {
		out.Items = rows[:limit]
		last := out.Items[limit-1]
		out.NextCursor = operationlog.CursorFor(last.At, last.ID)
	}
	return out, nil
}
func safeAuditProjection(in map[string]string) map[string]string {
	out := map[string]string{}
	for _, k := range []string{"version", "active", "role", "access_version", "password_configured", "price_rial", "stock", "reserved", "status", "content_policy_version", "published", "value", "kind", "allow_wallet_topup", "credential_count", "grants", "permissions", "order_status", "payment", "amount_rial"} {
		maximum := 80
		if k == "grants" || k == "permissions" {
			maximum = 4096
		}
		if v, ok := in[k]; ok && len(v) <= maximum && !strings.ContainsAny(v, "\r\n") {
			out[k] = v
		}
	}
	return out
}
