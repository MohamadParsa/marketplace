package commerce

import "time"

type OrderRequestItem struct {
	OrderID     string `json:"order_id"`
	ItemID      string `json:"item_id"`
	OfferID     string `json:"offer_id"`
	Quantity    int    `json:"quantity"`
	UnitIndices []int  `json:"unit_indices"`
}

type OrderRequest struct {
	ID            string             `json:"id"`
	Number        int64              `json:"number"`
	PurchaseID    string             `json:"purchase_id"`
	UserID        string             `json:"user_id"`
	Kind          string             `json:"kind"`
	Origin        string             `json:"origin"`
	Status        string             `json:"status"`
	Version       int                `json:"version"`
	Reason        string             `json:"reason"`
	CreatedAt     time.Time          `json:"created_at"`
	CreatedBy     string             `json:"created_by"`
	ReviewedAt    *time.Time         `json:"reviewed_at,omitempty"`
	ReviewedBy    string             `json:"reviewed_by,omitempty"`
	ReviewReason  string             `json:"review_reason,omitempty"`
	GoodsReceived bool               `json:"goods_received"`
	ReceivedAt    *time.Time         `json:"received_at,omitempty"`
	Inspected     bool               `json:"inspected"`
	Healthy       bool               `json:"healthy"`
	InspectedAt   *time.Time         `json:"inspected_at,omitempty"`
	StockApplied  bool               `json:"stock_applied"`
	Items         []OrderRequestItem `json:"items"`
}
