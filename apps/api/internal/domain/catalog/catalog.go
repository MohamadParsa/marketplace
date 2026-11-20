package catalog

type Category struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	ParentID string             `json:"parent_id,omitempty"`
	Bindings []AttributeBinding `json:"bindings"`
	Archived bool               `json:"archived"`
	IconID   string             `json:"icon_id,omitempty"`
	Version  int                `json:"version"`
}

type Brand struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	IconID  string `json:"icon_id,omitempty"`
	Version int    `json:"version"`
}

// IconMedia is held only by the test repository, never in a workspace response.
type IconMedia struct {
	ID   string `json:"id"`
	Data []byte `json:"data"`
}
type Product struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	CategoryID string           `json:"category_id"`
	Model      string           `json:"model"`
	Status     string           `json:"status"`
	Version    int              `json:"version"`
	Attributes []AttributeValue `json:"attributes,omitempty"`
}

// Definitions are shared; usage and ordering belong to an explicit category binding.
type AttributeDefinition struct {
	ID      string   `json:"id"`
	Code    string   `json:"code"`
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
	Unit    string   `json:"unit,omitempty"`
	Options []string `json:"options"`
	Version int      `json:"version"`
}
type AttributeBinding struct {
	AttributeID string `json:"attribute_id"`
	Axis        bool   `json:"axis"`
	Card        bool   `json:"card"`
	Filter      bool   `json:"filter"`
	Position    int    `json:"position"`
}
type AttributeValue struct {
	AttributeID string `json:"attribute_id"`
	Value       string `json:"value"`
}
type DisplaySettings struct {
	CardAttributeLimit int `json:"card_attribute_limit"`
	Version            int `json:"version"`
}
type CardFeature struct {
	AttributeID string `json:"attribute_id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Unit        string `json:"unit,omitempty"`
}
type CardPreview struct {
	ProductID string        `json:"product_id"`
	Features  []CardFeature `json:"features"`
}
type Offer struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	StoreID   string `json:"store_id"`
	Stock     int    `json:"stock"`
	Price     int64  `json:"price"`
}
