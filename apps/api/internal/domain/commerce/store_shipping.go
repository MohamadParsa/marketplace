package commerce

import "time"

type StoreShippingPolicy struct {
	StoreID  string                 `json:"store_id"`
	Version  int                    `json:"version"`
	Bindings []StoreShippingBinding `json:"bindings"`
}

type StoreShippingBinding struct {
	MethodID string                  `json:"method_id"`
	Enabled  bool                    `json:"enabled"`
	Mode     string                  `json:"mode"`
	Live     *StoreShippingRevision  `json:"live,omitempty"`
	Draft    *StoreShippingRevision  `json:"draft,omitempty"`
	History  []StoreShippingRevision `json:"history,omitempty"`
}

type StoreShippingRevision struct {
	Number int            `json:"number"`
	Status string         `json:"status"`
	Base   ShippingMethod `json:"base"`
	Method ShippingMethod `json:"method"`
	Reason string         `json:"reason"`
	Actor  string         `json:"actor"`
	At     time.Time      `json:"at"`
}
