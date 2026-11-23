package content

type BannerConfig struct {
	CatalogCategoryID string `json:"catalog_category_id,omitempty"`

	Placement string `json:"placement,omitempty"`
	DesktopID string `json:"desktop_id"`
	MobileID  string `json:"mobile_id"`
	Alt       string `json:"alt"`
	Link      string `json:"link"`
	Enabled   bool   `json:"enabled"`
}
type PublishedBanner struct {
	Config   BannerConfig `json:"config"`
	Revision int          `json:"revision"`
}
type Banner struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Version   int              `json:"version"`
	Draft     BannerConfig     `json:"draft"`
	Published *PublishedBanner `json:"published,omitempty"`
}

// Binary data stays private in the test repository, never in workspace/list DTOs.
type BannerAsset struct {
	ID     string `json:"id"`
	Data   []byte `json:"data"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
type PublicBanner struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Revision int          `json:"revision"`
	Config   BannerConfig `json:"config"`
}
