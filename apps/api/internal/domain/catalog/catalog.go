package catalog

type Category struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Attributes []string `json:"attributes"`
	Archived   bool     `json:"archived"`
	IconID     string   `json:"icon_id,omitempty"`
	Version    int      `json:"version"`
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
	ID         string `json:"id"`
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
	Model      string `json:"model"`
	Status     string `json:"status"`
	Version    int    `json:"version"`
}
type Offer struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	StoreID   string `json:"store_id"`
	Stock     int    `json:"stock"`
	Price     int64  `json:"price"`
}
