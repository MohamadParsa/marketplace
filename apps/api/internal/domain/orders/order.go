package orders

type Item struct {
	VariantLabel string `json:"variant_label,omitempty"`
	OfferID      string `json:"offer_id,omitempty"`
	VariantID    string `json:"variant_id,omitempty"`
	Name         string `json:"name,omitempty"`
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	Quantity     int    `json:"quantity"`
	UnitPrice    int64  `json:"unit_price"`
	Status       string `json:"status"`
}
type Order struct {
	PurchaseNumber string   `json:"purchase_number,omitempty"`
	CheckoutID     string   `json:"checkout_id,omitempty"`
	ShippingFee    int64    `json:"shipping_fee,omitempty"`
	ShippingMethod string   `json:"shipping_method,omitempty"`
	DeliveryFrom   string   `json:"delivery_from,omitempty"`
	DeliveryTo     string   `json:"delivery_to,omitempty"`
	ID             string   `json:"id"`
	PurchaseID     string   `json:"purchase_id"`
	StoreID        string   `json:"store_id"`
	City           string   `json:"city"`
	Status         string   `json:"status"`
	Payment        string   `json:"payment"`
	CreatedAt      string   `json:"created_at"`
	Items          []Item   `json:"items"`
	Version        int      `json:"version"`
	Packages       []string `json:"packages"`
}

func (o Order) Total() int64 {
	var total int64
	for _, item := range o.Items {
		total += item.UnitPrice * int64(item.Quantity)
	}
	return total
}
