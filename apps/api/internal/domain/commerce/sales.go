package commerce

// SalesSettings is independent from payment/funding and commission rules.
type SalesMethod struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Active           bool     `json:"active"`
	MarkupBPS        int      `json:"markup_bps"`
	PaymentMethodIDs []string `json:"payment_method_ids"`
}
type CategoryMarkup struct {
	CategoryID string `json:"category_id"`
	MethodID   string `json:"method_id"`
	MarkupBPS  int    `json:"markup_bps"`
}
type SalesSettings struct {
	Version   int              `json:"version"`
	Methods   []SalesMethod    `json:"methods"`
	Overrides []CategoryMarkup `json:"overrides"`
}
type MarkupSnapshot struct {
	MethodID   string `json:"method_id"`
	MethodName string `json:"method_name"`
	CategoryID string `json:"category_id,omitempty"`
	BPS        int    `json:"bps"`
	UnitAmount Amount `json:"unit_amount"`
}
