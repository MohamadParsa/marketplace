package commerce

import "time"

// Immutable adjustment for reconciliation. Commission is deliberately not inferred.
type RefundAdjustment struct {
	OrderID          string `json:"order_id"`
	ItemID           string `json:"item_id,omitempty"`
	StoreID          string `json:"store_id"`
	UnitIndices      []int  `json:"unit_indices,omitempty"`
	Gross            Amount `json:"gross"`
	DiscountPlatform Amount `json:"discount_platform"`
	DiscountSeller   Amount `json:"discount_seller"`
	Net              Amount `json:"net"`
}
type OrderRefund struct {
	ID               string             `json:"id"`
	RequestID        string             `json:"request_id"`
	PurchaseID       string             `json:"purchase_id"`
	UserID           string             `json:"user_id"`
	Method           string             `json:"method"`
	Amount           Amount             `json:"amount"`
	Shipping         Amount             `json:"shipping"`
	ShippingIncluded bool               `json:"shipping_included"`
	OriginalSources  []CreditAllocation `json:"original_sources"`
	CreditEntryID    string             `json:"credit_entry_id,omitempty"`
	WithdrawalID     string             `json:"withdrawal_id,omitempty"`
	Adjustments      []RefundAdjustment `json:"adjustments"`
	SettlementStatus string             `json:"settlement_status"`
	At               time.Time          `json:"at"`
	ActorID          string             `json:"actor_id"`
	Reason           string             `json:"reason"`
}
