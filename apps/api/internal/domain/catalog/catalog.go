package catalog

import "time"

type SEO struct {
	MetaTitle       string `json:"meta_title"`
	MetaKeywords    string `json:"meta_keywords"`
	MetaDescription string `json:"meta_description"`
}

type Category struct {
	DisplayPriority *int               `json:"display_priority,omitempty"`
	Slug            string             `json:"slug,omitempty"`
	SEO             SEO                `json:"seo"`
	EnglishName     string             `json:"english_name,omitempty"`
	Description     string             `json:"description,omitempty"`
	Published       bool               `json:"published"`
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	ParentID        string             `json:"parent_id,omitempty"`
	Bindings        []AttributeBinding `json:"bindings"`
	Archived        bool               `json:"archived"`
	IconID          string             `json:"icon_id,omitempty"`
	Version         int                `json:"version"`
}

type Brand struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	IconID  string `json:"icon_id,omitempty"`
	Version int    `json:"version"`
}

// IconMedia keeps inline bytes in memory; PostgreSQL stores bytes separately and returns a stored reference.
// This collection never appears in a workspace HTTP response.
type IconMedia struct {
	Stored bool   `json:"stored,omitempty"`
	ID     string `json:"id"`
	Data   []byte `json:"data"`
}
type ProductContent struct {
	ShortDescription string             `json:"short_description"`
	FullDescription  string             `json:"full_description"`
	Advantages       []string           `json:"advantages"`
	Disadvantages    []string           `json:"disadvantages"`
	ReviewsEnabled   bool               `json:"reviews_enabled"`
	QAEnabled        bool               `json:"qa_enabled"`
	Shipping         ShippingDimensions `json:"shipping"`
	SEO              SEO                `json:"seo"`
	NoIndex          bool               `json:"noindex"`
	NoFollow         bool               `json:"nofollow"`
}

// Shipping measurements never represent gold mass or a pricing input.
type ShippingDimensions struct {
	Weight        *float64 `json:"weight,omitempty"`
	Length        *float64 `json:"length,omitempty"`
	Width         *float64 `json:"width,omitempty"`
	Height        *float64 `json:"height,omitempty"`
	WeightUnit    string   `json:"weight_unit"`
	DimensionUnit string   `json:"dimension_unit"`
}
type ProductContentRevision struct {
	ID              string           `json:"id"`
	State           string           `json:"state"`
	BaseFingerprint string           `json:"base_fingerprint"`
	SubmittedBy     string           `json:"submitted_by"`
	SubmittedAt     time.Time        `json:"submitted_at"`
	ReviewedBy      string           `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time       `json:"reviewed_at,omitempty"`
	Reason          string           `json:"reason,omitempty"`
	Name            string           `json:"name"`
	EnglishName     string           `json:"english_name"`
	Description     string           `json:"description"`
	BrandID         string           `json:"brand_id"`
	Content         *ProductContent  `json:"content,omitempty"`
	Attributes      []AttributeValue `json:"attributes"`
}

type Product struct {
	PublicCode      string                  `json:"public_code,omitempty"`
	ContentRevision *ProductContentRevision `json:"content_revision,omitempty"`
	Content         *ProductContent         `json:"content,omitempty"`
	GalleryIDs      []string                `json:"gallery_ids,omitempty"`
	CreatedAt       time.Time               `json:"created_at,omitempty"`
	EnglishName     string                  `json:"english_name,omitempty"`
	Description     string                  `json:"description,omitempty"`
	BrandID         string                  `json:"brand_id,omitempty"`
	ImageID         string                  `json:"image_id,omitempty"`
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	CategoryID      string                  `json:"category_id"`
	Model           string                  `json:"model"`
	Status          string                  `json:"status"`
	Version         int                     `json:"version"`
	Attributes      []AttributeValue        `json:"attributes,omitempty"`
	OwnerStoreID    string                  `json:"owner_store_id,omitempty"`
	VariantVersion  int                     `json:"variant_version"`
}

type Variant struct {
	ImageID   string           `json:"image_id,omitempty"`
	ID        string           `json:"id"`
	ProductID string           `json:"product_id"`
	SKU       string           `json:"sku"`
	Values    []AttributeValue `json:"values"`
	Version   int              `json:"version"`
}

// Definitions are shared; usage and ordering belong to an explicit category binding.
type AttributeDefinition struct {
	Groups  []AttributeValueGroup `json:"groups,omitempty"`
	ID      string                `json:"id"`
	Code    string                `json:"code"`
	Name    string                `json:"name"`
	Kind    string                `json:"kind"`
	Unit    string                `json:"unit,omitempty"`
	Options []string              `json:"options"`
	Version int                   `json:"version"`
}

// Display aliases never replace a real variant value or merge its inventory.
type AttributeValueGroup struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Values []string `json:"values"`
	Active bool     `json:"active"`
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
	SalesWindowDays    int `json:"sales_window_days,omitempty"`
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
	RegularPrice int64  `json:"regular_price,omitempty"`
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	StoreID      string `json:"store_id"`
	Stock        int    `json:"stock"`
	Price        int64  `json:"price"`
	VariantID    string `json:"variant_id"`
	Reserved     int    `json:"reserved"`
	Version      int    `json:"version"`
}
