package finance

import "errors"

var ErrCapacity = errors.New("amount exceeds remaining capacity")

type Transfer struct {
	Reserved  bool   `json:"reserved"`
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	OrderID   string `json:"order_id"`
	ItemID    string `json:"item_id"`
	StoreID   string `json:"store_id"`
	Amount    int64  `json:"amount"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	Reference string `json:"reference"`
	Version   int    `json:"version"`
}
type Entry struct {
	ID         string `json:"id"`
	TransferID string `json:"transfer_id"`
	Amount     int64  `json:"amount"`
	Kind       string `json:"kind"`
	Source     string `json:"source"`
}

func ValidOutcome(value string) bool {
	return value == "success" || value == "failure" || value == "unknown"
}
