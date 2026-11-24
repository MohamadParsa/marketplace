package customer

import "time"

// RecentlyViewedEntry is the last successful product view, not an analytics event log.
type RecentlyViewedEntry struct {
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	ViewedAt  time.Time `json:"viewed_at"`
}
