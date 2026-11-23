package backoffice

import (
	"context"
	"sort"
	"time"
)

const DefaultSalesWindowDays = 30
const MaxSalesWindowDays = 3650

type CatalogSale struct {
	ProductID string    `json:"product_id"`
	Quantity  int64     `json:"quantity"`
	PaidAt    time.Time `json:"paid_at"`
}

// Internal read model only: never included in a public response or cached DTO.
type CatalogSales struct {
	WindowDays int           `json:"window_days"`
	Sales      []CatalogSale `json:"sales"`
}
type CatalogSalesRepository interface {
	CatalogSalesRevision(context.Context) (int64, error)
	CatalogSales(context.Context) (CatalogSales, error)
}

func SalesWindowDays(w Workspace) int {
	n := w.DisplaySettings.SalesWindowDays
	if n < 1 || n > MaxSalesWindowDays {
		return DefaultSalesWindowDays
	}
	return n
}
func BuildCatalogSales(w Workspace) CatalogSales {
	result := CatalogSales{WindowDays: SalesWindowDays(w), Sales: []CatalogSale{}}
	if w.Commerce == nil {
		return result
	}
	// A successful full-item refund is counted once even if multiple support cases/transfers reference it.
	refunded := map[string]bool{}
	for _, t := range w.Transfers {
		if t.Kind == "refund" && t.Status == "success" {
			refunded[t.OrderID+"\x00"+t.ItemID] = true
		}
	}
	paid := map[string]time.Time{}
	for _, p := range w.Commerce.Purchases {
		if p.ConfirmedAt != nil && !p.ConfirmedAt.IsZero() && p.Status != "cancelled" {
			paid[p.ID] = p.ConfirmedAt.UTC()
		}
	}
	seen := map[string]bool{}
	for _, order := range w.Orders {
		at, ok := paid[order.CheckoutID]
		if !ok || order.Payment != "paid" || order.Status == "cancelled" || order.Status == "refunded" {
			continue
		}
		for _, item := range order.Items {
			key := order.ID + "\x00" + item.ID
			if seen[key] {
				continue
			}
			seen[key] = true
			if item.Quantity <= 0 || item.ProductID == "" || item.Status == "cancelled" || item.Status == "refunded" || refunded[key] {
				continue
			}
			result.Sales = append(result.Sales, CatalogSale{ProductID: item.ProductID, Quantity: int64(item.Quantity), PaidAt: at})
		}
	}
	sort.Slice(result.Sales, func(i, j int) bool {
		a, b := result.Sales[i], result.Sales[j]
		if a.ProductID != b.ProductID {
			return a.ProductID < b.ProductID
		}
		if !a.PaidAt.Equal(b.PaidAt) {
			return a.PaidAt.Before(b.PaidAt)
		}
		return a.Quantity < b.Quantity
	})
	return result
}
func (s CatalogSales) Counts(now time.Time) map[string]int64 {
	counts := map[string]int64{}
	days := s.WindowDays
	if days < 1 || days > MaxSalesWindowDays {
		days = DefaultSalesWindowDays
	}
	// Rolling interval (now-days*24h, now], independent of server midnight/DST.
	from := now.Add(-time.Duration(days) * 24 * time.Hour)
	for _, v := range s.Sales {
		if v.Quantity > 0 && v.PaidAt.After(from) && !v.PaidAt.After(now) {
			counts[v.ProductID] += v.Quantity
		}
	}
	return counts
}
