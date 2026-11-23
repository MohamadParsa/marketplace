package customer

import "time"

// WishlistEntry stores a private product bookmark, independent of variants/offers.
type WishlistEntry struct {
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}
