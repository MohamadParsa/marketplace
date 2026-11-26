package commerce

import (
	"encoding/json"
	"time"
)

// Amount is the existing catalog minor unit: one tenth of a toman.
// Public monetary inputs and outputs use exact decimal toman strings.
type Amount int64

const MaxAmount Amount = 9000000000000000
const ReservationMinutes = 30

type State struct {
	Sales              *SalesSettings  `json:"sales,omitempty"`
	Coupons            []Coupon        `json:"coupons,omitempty"`
	CreditLots         []CreditLot     `json:"credit_lots,omitempty"`
	CreditAccounts     []CreditAccount `json:"credit_accounts,omitempty"`
	NextPurchaseNumber int64           `json:"next_purchase_number"`
	Settings           Settings        `json:"settings"`
	Carts              []Cart          `json:"carts"`
	Addresses          []Address       `json:"addresses"`
	Quotes             []Quote         `json:"quotes"`
	Purchases          []Purchase      `json:"purchases"`
	Deposits           []Deposit       `json:"deposits"`
	Receipts           []Receipt       `json:"receipts"`
	Entries            []WalletEntry   `json:"entries"`
	Refunds            []RefundRequest `json:"refunds"`
	Operations         []Operation     `json:"operations"`
}
type City struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}
type Province struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Cities []City `json:"cities"`
}
type Settings struct {
	Geography          []Province       `json:"geography"`
	DestinationHistory []Destination    `json:"destination_history,omitempty"`
	Version            int              `json:"version"`
	Methods            []PaymentMethod  `json:"methods"`
	Shipping           []ShippingMethod `json:"shipping"`
	ClosedWeekdays     []int            `json:"closed_weekdays"`
	Holidays           []string         `json:"holidays"`
	Destination        Destination      `json:"destination"`
	Business           Business         `json:"business"`
}
type PaymentMethod struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Kind               string   `json:"kind"`
	Active             bool     `json:"active"`
	Min                Amount   `json:"min"`
	Max                Amount   `json:"max"`
	AllCategories      bool     `json:"all_categories"`
	AllowedCategories  []string `json:"allowed_categories"`
	ExcludedCategories []string `json:"excluded_categories"`
	Instructions       string   `json:"instructions"`
}
type ShippingMethod struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Active    bool     `json:"active"`
	AllCities bool     `json:"all_cities"`
	Cities    []string `json:"cities"`
	StoreIDs  []string `json:"store_ids"`
	Fee       Amount   `json:"fee"`
	MinDays   int      `json:"min_days"`
	MaxDays   int      `json:"max_days"`
}
type Destination struct {
	ID         string `json:"id"`
	CardNumber string `json:"card_number"`
	Holder     string `json:"holder"`
	Bank       string `json:"bank"`
}
type Business struct {
	Name           string `json:"name"`
	SupportPhone   string `json:"support_phone"`
	SupportEmail   string `json:"support_email"`
	About          string `json:"about"`
	Contact        string `json:"contact"`
	ShippingPolicy string `json:"shipping_policy"`
	ReturnPolicy   string `json:"return_policy"`
	Privacy        string `json:"privacy"`
	Terms          string `json:"terms"`
}
type Cart struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id,omitempty"`
	GuestHash  string     `json:"guest_hash,omitempty"`
	MergedInto string     `json:"merged_into,omitempty"`
	Version    int        `json:"version"`
	Lines      []CartLine `json:"lines"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
type CartLine struct {
	ReviewRequired       bool   `json:"review_required,omitempty"`
	PreviousVariantLabel string `json:"previous_variant_label,omitempty"`
	OfferID              string `json:"offer_id"`
	Quantity             int    `json:"quantity"`
	SeenPrice            Amount `json:"seen_price"`
}
type Address struct {
	DetailVersion       int      `json:"detail_version,omitempty"`
	Title               string   `json:"title,omitempty"`
	FullAddress         string   `json:"full_address,omitempty"`
	HouseNumber         string   `json:"house_number,omitempty"`
	Unit                string   `json:"unit,omitempty"`
	RecipientSelf       *bool    `json:"recipient_self,omitempty"`
	RecipientNationalID string   `json:"recipient_national_id,omitempty"`
	Latitude            *float64 `json:"latitude,omitempty"`
	Longitude           *float64 `json:"longitude,omitempty"`

	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Version    int    `json:"version"`
	Recipient  string `json:"recipient"`
	Phone      string `json:"phone"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Street     string `json:"street"`
	PostalCode string `json:"postal_code"`
}
type LineSnapshot struct {
	CashUnitPrice      *Amount         `json:"cash_unit_price,omitempty"`
	Markup             *MarkupSnapshot `json:"markup,omitempty"`
	ReferenceUnitPrice *Amount         `json:"reference_unit_price,omitempty"`
	DirectDiscount     *Amount         `json:"direct_discount,omitempty"`
	OfferID            string          `json:"offer_id"`
	ProductID          string          `json:"product_id"`
	VariantID          string          `json:"variant_id"`
	StoreID            string          `json:"store_id"`
	Name               string          `json:"name"`
	VariantLabel       string          `json:"variant_label"`
	ImageURL           string          `json:"image_url"`
	UnitPrice          Amount          `json:"unit_price"`
	Quantity           int             `json:"quantity"`
	OfferVersion       int             `json:"offer_version"`
}
type ShipmentSnapshot struct {
	StoreID      string         `json:"store_id"`
	StoreName    string         `json:"store_name"`
	Method       ShippingMethod `json:"method"`
	LeadDays     int            `json:"lead_days"`
	DeliveryFrom string         `json:"delivery_from"`
	DeliveryTo   string         `json:"delivery_to"`
}
type Quote struct {
	SalesMethod        *SalesMethod       `json:"sales_method,omitempty"`
	MarkupTotal        Amount             `json:"markup_total,omitempty"`
	CashSubtotal       Amount             `json:"cash_subtotal,omitempty"`
	ContractVersion    int                `json:"contract_version,omitempty"`
	CouponCode         string             `json:"coupon_code,omitempty"`
	Discount           *DiscountUsage     `json:"discount,omitempty"`
	Credit             *CreditUsage       `json:"credit,omitempty"`
	Destination        Destination        `json:"destination"`
	AcceptedAt         *time.Time         `json:"accepted_at,omitempty"`
	ID                 string             `json:"id"`
	UserID             string             `json:"user_id"`
	CartID             string             `json:"cart_id"`
	CartVersion        int                `json:"cart_version"`
	ReplacesPurchaseID string             `json:"replaces_purchase_id,omitempty"`
	Status             string             `json:"status"`
	Address            Address            `json:"address"`
	Lines              []LineSnapshot     `json:"lines"`
	Shipments          []ShipmentSnapshot `json:"shipments"`
	PaymentMethod      PaymentMethod      `json:"payment_method"`
	Subtotal           Amount             `json:"subtotal"`
	ShippingTotal      Amount             `json:"shipping_total"`
	Total              Amount             `json:"total"`
	CreatedAt          time.Time          `json:"created_at"`
	ExpiresAt          time.Time          `json:"expires_at"`
	ReservedUntil      time.Time          `json:"reserved_until"`
	PurchaseID         string             `json:"purchase_id,omitempty"`
}
type Purchase struct {
	Number                int64        `json:"number"`
	Events                []OrderEvent `json:"events"`
	CancellationRequested bool         `json:"cancellation_requested"`
	CancellationReason    string       `json:"cancellation_reason,omitempty"`
	ID                    string       `json:"id"`
	UserID                string       `json:"user_id"`
	QuoteID               string       `json:"quote_id"`
	PreviousQuotes        []string     `json:"previous_quotes"`
	OrderIDs              []string     `json:"order_ids"`
	Status                string       `json:"status"`
	Version               int          `json:"version"`
	CreatedAt             time.Time    `json:"created_at"`
	ConfirmedAt           *time.Time   `json:"confirmed_at,omitempty"`
	CancelledAt           *time.Time   `json:"cancelled_at,omitempty"`
	Reason                string       `json:"reason,omitempty"`
}
type Deposit struct {
	VerifiedReference string        `json:"verified_reference,omitempty"`
	VerifiedDate      string        `json:"verified_date,omitempty"`
	VerifiedKey       string        `json:"verified_key,omitempty"`
	ID                string        `json:"id"`
	UserID            string        `json:"user_id"`
	PurchaseID        string        `json:"purchase_id,omitempty"`
	Method            PaymentMethod `json:"method"`
	Destination       Destination   `json:"destination"`
	Amount            Amount        `json:"amount"`
	Date              string        `json:"date"`
	Reference         string        `json:"reference"`
	ReferenceKey      string        `json:"reference_key"`
	ReceiptID         string        `json:"receipt_id,omitempty"`
	Status            string        `json:"status"`
	Version           int           `json:"version"`
	CreatedAt         time.Time     `json:"created_at"`
	ReviewedAt        *time.Time    `json:"reviewed_at,omitempty"`
	ReviewedBy        string        `json:"reviewed_by,omitempty"`
	Reason            string        `json:"reason,omitempty"`
}
type Receipt struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	DepositID string `json:"deposit_id"`
	Data      []byte `json:"data"`
}
type WalletEntry struct {
	ID     string    `json:"id"`
	UserID string    `json:"user_id"`
	Source string    `json:"source"`
	Amount Amount    `json:"amount"`
	Kind   string    `json:"kind"`
	At     time.Time `json:"at"`
}
type RefundRequest struct {
	Allocations  []CreditAllocation `json:"credit_allocations,omitempty"`
	ID           string             `json:"id"`
	UserID       string             `json:"user_id"`
	Amount       Amount             `json:"amount"`
	Instructions string             `json:"instructions"`
	Status       string             `json:"status"`
	Version      int                `json:"version"`
	CreatedAt    time.Time          `json:"created_at"`
	ResolvedAt   *time.Time         `json:"resolved_at,omitempty"`
	ResolvedBy   string             `json:"resolved_by,omitempty"`
	Reference    string             `json:"reference,omitempty"`
	Date         string             `json:"date,omitempty"`
	Reason       string             `json:"reason,omitempty"`
}
type Operation struct {
	Scope       string          `json:"scope"`
	Key         string          `json:"key"`
	Fingerprint string          `json:"fingerprint"`
	Result      json.RawMessage `json:"result"`
	At          time.Time       `json:"at"`
}

type OrderEvent struct {
	Action        string    `json:"action"`
	FlowStatus    string    `json:"flow_status"`
	OrderStatus   string    `json:"order_status"`
	PaymentStatus string    `json:"payment_status"`
	ActorID       string    `json:"actor_id"`
	ActorKind     string    `json:"actor_kind"`
	At            time.Time `json:"at"`
}
