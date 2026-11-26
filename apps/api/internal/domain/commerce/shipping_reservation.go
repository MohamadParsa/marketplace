package commerce

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
	"time"
)

type ShippingReservation struct {
	QuoteID     string                `json:"quote_id"`
	PurchaseID  string                `json:"purchase_id"`
	StoreID     string                `json:"store_id"`
	MethodID    string                `json:"method_id"`
	Plan        shipping.DeliveryPlan `json:"plan"`
	Status      string                `json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	ExpiresAt   time.Time             `json:"expires_at"`
	CommittedAt *time.Time            `json:"committed_at,omitempty"`
	ReleasedAt  *time.Time            `json:"released_at,omitempty"`
}
