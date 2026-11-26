package backoffice

import (
	"context"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/content"
	"github.com/cherriva/marketplace/apps/api/internal/domain/customer"
	"github.com/cherriva/marketplace/apps/api/internal/domain/finance"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
	"github.com/cherriva/marketplace/apps/api/internal/domain/support"
	"time"
)

var (
	ErrReauthenticationRequired = errors.New("reauthentication_required")
	ErrInvalidReauthentication  = errors.New("invalid_reauthentication")
	ErrForbidden                = errors.New("forbidden")
	ErrNotFound                 = errors.New("not_found")
	ErrConflict                 = errors.New("version_conflict")
	ErrInvalid                  = errors.New("invalid_request")
)

type User struct {
	FirstName              string        `json:"first_name"`
	LastName               string        `json:"last_name"`
	Gender                 string        `json:"gender"`
	Version                int           `json:"version"`
	PasswordChangeRequired bool          `json:"password_change_required"`
	ID                     string        `json:"id"`
	Name                   string        `json:"name"`
	Phone                  string        `json:"phone"`
	Role                   identity.Role `json:"role"`
	Active                 bool          `json:"active"`
	Email                  string        `json:"email,omitempty"`
	NationalID             string        `json:"national_id,omitempty"`
	AccessVersion          int           `json:"-"`
	PasswordHash           string        `json:"-"`
	SellerOnly             bool          `json:"seller_only,omitempty"`
}

// Minimal test-scope membership permissions; custom role management remains Issue #1.
type StoreMembership struct {
	UserID       string `json:"user_id"`
	StoreID      string `json:"store_id"`
	Active       bool   `json:"active"`
	CreateOffers bool   `json:"create_offers"`
	EditPrice    bool   `json:"edit_price"`
	EditStock    bool   `json:"edit_stock"`
}
type Rule struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Level  string `json:"level"`
	Kind   string `json:"kind"`
	Value  int64  `json:"value"`
	Active bool   `json:"active"`
}
type Shipping struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	City   string `json:"city"`
	Tariff int64  `json:"tariff"`
	Days   int    `json:"days"`
}
type Notification struct {
	ID        string `json:"id"`
	Event     string `json:"event"`
	Recipient string `json:"recipient"`
	Status    string `json:"status"`
	Retries   int    `json:"retries"`
}
type Audit struct {
	PreviousRole identity.Role `json:"previous_role,omitempty"`
	Role         identity.Role `json:"role,omitempty"`
	RoleID       string        `json:"role_id,omitempty"`
	StoreID      string        `json:"store_id,omitempty"`
	Permissions  []string      `json:"permissions,omitempty"`
	ResultID     string        `json:"result_id,omitempty"`
	Fingerprint  string        `json:"fingerprint"`
	ID           string        `json:"id"`
	Actor        string        `json:"actor"`
	Action       string        `json:"action"`
	Target       string        `json:"target"`
	Reason       string        `json:"reason"`
	At           time.Time     `json:"at"`
}
type Workspace struct {
	AxisPlans      []AxisPlan                     `json:"axis_plans,omitempty"`
	UserMutations  []UserMutation                 `json:"user_mutations,omitempty"`
	SalesPayments  []commerce.PaymentMethod       `json:"sales_payments,omitempty"`
	Sales          *commerce.SalesSettings        `json:"sales,omitempty"`
	RecentlyViewed []customer.RecentlyViewedEntry `json:"recently_viewed,omitempty"`
	Wishlist       []customer.WishlistEntry       `json:"wishlist,omitempty"`
	publicIndex    *publicCatalogIndex

	Commerce         *commerce.State               `json:"commerce,omitempty"`
	HomeSettings     HomeSettings                  `json:"home_settings"`
	StoreMemberships []identity.StoreMembership    `json:"-"`
	Access           *AccessDirectory              `json:"access,omitempty"`
	Purchases        []PurchaseOwner               `json:"-"`
	Banners          []content.Banner              `json:"banners,omitempty"`
	BannerAssets     []content.BannerAsset         `json:"banner_assets,omitempty"`
	Users            []User                        `json:"users,omitempty"`
	Stores           []store.Store                 `json:"stores,omitempty"`
	Categories       []catalog.Category            `json:"categories,omitempty"`
	Attributes       []catalog.AttributeDefinition `json:"attributes,omitempty"`
	DisplaySettings  catalog.DisplaySettings       `json:"display_settings"`
	CardPreviews     []catalog.CardPreview         `json:"card_previews,omitempty"`
	Brands           []catalog.Brand               `json:"brands,omitempty"`
	IconMedia        []catalog.IconMedia           `json:"icon_media,omitempty"`
	IconPolicy       *IconPolicy                   `json:"icon_policy,omitempty"`
	Products         []catalog.Product             `json:"products,omitempty"`
	ProductReviews   []ProductReview               `json:"product_reviews,omitempty"`
	Offers           []catalog.Offer               `json:"offers,omitempty"`
	Variants         []catalog.Variant             `json:"variants,omitempty"`
	Memberships      []StoreMembership             `json:"memberships,omitempty"`
	Orders           []orders.Order                `json:"orders,omitempty"`
	Cases            []support.Case                `json:"cases,omitempty"`
	Assignments      []support.Assignment          `json:"assignments,omitempty"`
	Transfers        []finance.Transfer            `json:"transfers,omitempty"`
	Ledger           []finance.Entry               `json:"ledger,omitempty"`
	Rules            []Rule                        `json:"rules,omitempty"`
	Shipping         []Shipping                    `json:"shipping,omitempty"`
	Notifications    []Notification                `json:"notifications,omitempty"`
	Audit            []Audit                       `json:"audit,omitempty"`
}
type AccessDirectory struct {
	Roles       []identity.RoleDefinition       `json:"roles"`
	Grants      []identity.RoleGrant            `json:"grants"`
	Memberships []identity.StoreMembership      `json:"memberships,omitempty"`
	Registry    map[identity.RoleScope][]string `json:"registry,omitempty"`
}
type PurchaseOwner struct {
	PurchaseID string
	UserID     string
}
type Command struct {
	RoleVersion    *int                         `json:"role_version,omitempty"`
	UserVersion    *int                         `json:"user_version,omitempty"`
	ProductMedia   *ProductMediaInput           `json:"product_media,omitempty"`
	Catalog        *CatalogInput                `json:"catalog,omitempty"`
	Home           *HomeSettings                `json:"home,omitempty"`
	Scope          identity.RoleScope           `json:"scope,omitempty"`
	Permissions    []string                     `json:"permissions,omitempty"`
	StoreID        string                       `json:"store_id,omitempty"`
	Banner         *content.BannerConfig        `json:"banner,omitempty"`
	Phone          string                       `json:"phone"`
	Action         string                       `json:"action"`
	Target         string                       `json:"target"`
	Name           string                       `json:"name"`
	Value          string                       `json:"value"`
	Reason         string                       `json:"reason"`
	Amount         int64                        `json:"amount"`
	Version        int                          `json:"version"`
	IdempotencyKey string                       `json:"idempotency_key"`
	Attribute      *catalog.AttributeDefinition `json:"attribute,omitempty"`
	Bindings       []catalog.AttributeBinding   `json:"bindings,omitempty"`
	Variant        *VariantInput                `json:"variant,omitempty"`
	Offer          *OfferInput                  `json:"offer,omitempty"`
}

// Update must execute the callback atomically and roll back its writes on error.
type Repository interface {
	Snapshot(context.Context) (Workspace, error)
	Update(context.Context, func(*Workspace) error) error
}
type Clock interface{ Now() time.Time }
type Identifier interface{ Token() (string, error) }
