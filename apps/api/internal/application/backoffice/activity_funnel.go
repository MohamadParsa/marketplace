package backoffice

import (
	"context"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type ActivityStage struct {
	Key             string `json:"key"`
	Label           string `json:"label"`
	Events          int    `json:"events"`
	KnownIdentities int    `json:"known_identities"`
}
type ActivityFunnel struct {
	Stages              []ActivityStage `json:"stages"`
	Environment         string          `json:"environment"`
	MatchedEvents       int             `json:"matched_events"`
	UnlinkedEvents      int             `json:"unlinked_events"`
	UniquePaidPurchases int             `json:"unique_paid_purchases"`
}

func (s *Service) ReadActivityFunnel(ctx context.Context, a identity.User, q operationlog.Query) (ActivityFunnel, error) {
	if _, err := s.logAdmin(ctx, a); err != nil {
		return ActivityFunnel{}, err
	}
	if q.Validate() != nil || q.Cursor != "" {
		return ActivityFunnel{}, ErrInvalid
	}
	q.Type = "behavior"
	if q.Environment == "" {
		q.Environment = s.operationLogs.Environment()
	}
	out := ActivityFunnel{Environment: q.Environment, Stages: []ActivityStage{
		{Key: "product.viewed", Label: "مشاهده محصول"}, {Key: "search.completed", Label: "جست‌وجوی تکمیل‌شده"}, {Key: "cart.item-added", Label: "افزودن به سبد"}, {Key: "checkout.started", Label: "شروع خرید"}, {Key: "checkout.placed", Label: "ثبت سفارش"}, {Key: "checkout.paid", Label: "پرداخت قطعی سفارش"},
	}}
	if s.operationScanner == nil {
		return out, nil
	}
	seen := map[string]bool{}
	identities := map[string]map[string]bool{}
	stageKeys := map[string]int{}
	paidPurchases := map[string]bool{}
	placed := map[string]bool{}
	quotes := map[string]bool{}
	for i, stage := range out.Stages {
		stageKeys[stage.Key] = i
		identities[stage.Key] = map[string]bool{}
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err := s.operationScanner.Scan(ctx, q, func(e operationlog.Event) {
		if e.Outcome != "succeeded" || seen[e.EventID] {
			return
		}
		seen[e.EventID] = true
		i, known := stageKeys[e.EventName]
		if !known {
			return
		}
		if e.EventName == "checkout.paid" {
			if e.PurchaseID == "" || paidPurchases[e.PurchaseID] {
				return
			}
			paidPurchases[e.PurchaseID] = true
		}
		if e.EventName == "checkout.placed" {
			if placed[e.PurchaseID] {
				return
			}
			placed[e.PurchaseID] = true
		}
		if e.EventName == "checkout.started" {
			if quotes[e.EntityID] {
				return
			}
			quotes[e.EntityID] = true
		}
		out.MatchedEvents++
		out.Stages[i].Events++
		identity := ""
		if e.SubjectID != "" {
			identity = "u:" + e.SubjectID
		} else if e.SessionRef != "" {
			identity = "g:" + e.SessionRef
		}
		if identity != "" {
			identities[e.EventName][identity] = true
		} else {
			out.UnlinkedEvents++
		}
	})
	if err != nil {
		return ActivityFunnel{}, err
	}
	for i := range out.Stages {
		out.Stages[i].KnownIdentities = len(identities[out.Stages[i].Key])
	}
	out.UniquePaidPurchases = len(paidPurchases)
	return out, nil
}
