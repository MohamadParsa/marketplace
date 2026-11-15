package orders

type Item struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
	Status    string `json:"status"`
}
type Order struct {
	ID         string   `json:"id"`
	PurchaseID string   `json:"purchase_id"`
	StoreID    string   `json:"store_id"`
	City       string   `json:"city"`
	Status     string   `json:"status"`
	Payment    string   `json:"payment"`
	CreatedAt  string   `json:"created_at"`
	Items      []Item   `json:"items"`
	Version    int      `json:"version"`
	Packages   []string `json:"packages"`
}

func (o Order) Total() int64 {
	var total int64
	for _, item := range o.Items {
		total += item.UnitPrice * int64(item.Quantity)
	}
	return total
}
