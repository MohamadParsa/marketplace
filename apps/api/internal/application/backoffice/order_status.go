package backoffice

import (
	"slices"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
)

func orderLifecycle(o orders.Order) string {
	switch o.Status {
	case "cancelled", "superseded":
		return "cancelled"
	case "delivered":
		return "completed"
	case "supplied", "packing", "ready-to-ship", "awaiting-dispatch", "shipped":
		return "processing"
	}
	if o.OrderStatus == "processing" {
		return "processing"
	}
	return "placed"
}

// Normalize only witnessed states. A historical payment is not proof that staff
// started processing, and no timestamp/event or financial entry is backfilled.
func refreshPurchaseOrderState(w *Workspace, p *commerce.Purchase) {
	count, completed, cancelled := 0, 0, 0
	processing := false
	flow := ""
	for i := range w.Orders {
		o := &w.Orders[i]
		if !slices.Contains(p.OrderIDs, o.ID) {
			continue
		}
		o.OrderStatus = orderLifecycle(*o)
		count++
		if o.OrderStatus == "completed" {
			completed++
		}
		if o.OrderStatus == "cancelled" {
			cancelled++
		}
		processing = processing || o.OrderStatus == "processing" || o.OrderStatus == "completed"
		if flow == "" {
			flow = o.Status
		} else if flow != o.Status {
			flow = "mixed"
		}
	}
	p.OrderState = "placed"
	if processing {
		p.OrderState = "processing"
	}
	if count > 0 && completed+cancelled == count {
		if completed > 0 {
			p.OrderState = "completed"
		} else {
			p.OrderState = "cancelled"
		}
	}
	if p.Status == "cancelled" {
		p.OrderState = "cancelled"
	}
	p.FulfillmentState = flow
}
