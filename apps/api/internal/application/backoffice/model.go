package backoffice

import (
	"context"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/finance"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
	"github.com/cherriva/marketplace/apps/api/internal/domain/support"
	"time"
)

var (
	ErrForbidden                = errors.New("forbidden")
	ErrNotFound                 = errors.New("not_found")
	ErrConflict                 = errors.New("version_conflict")
	ErrInvalid                  = errors.New("invalid_request")
	ErrReauthenticationRequired = errors.New("reauthentication_required")
	ErrInvalidReauthentication  = errors.New("invalid_reauthentication")
)

type User struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Phone         string        `json:"phone"`
	Email         string        `json:"email,omitempty"`
	Role          identity.Role `json:"role"`
	Active        bool          `json:"active"`
	AccessVersion int           `json:"-"`
	PasswordHash  string        `json:"-"`
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
	Fingerprint  string        `json:"fingerprint"`
	ID           string        `json:"id"`
	Actor        string        `json:"actor"`
	Action       string        `json:"action"`
	Target       string        `json:"target"`
	Reason       string        `json:"reason"`
	At           time.Time     `json:"at"`
	PreviousRole identity.Role `json:"previous_role,omitempty"`
	Role         identity.Role `json:"role,omitempty"`
	RoleID       string        `json:"role_id,omitempty"`
	StoreID      string        `json:"store_id,omitempty"`
	Permissions  []string      `json:"permissions,omitempty"`
}
type Workspace struct {
	StoreMemberships []identity.StoreMembership `json:"-"`
	Access           *AccessDirectory           `json:"access,omitempty"`
	Purchases        []PurchaseOwner            `json:"-"`
	Users            []User                     `json:"users,omitempty"`
	Stores           []store.Store              `json:"stores,omitempty"`
	Categories       []catalog.Category         `json:"categories,omitempty"`
	Products         []catalog.Product          `json:"products,omitempty"`
	Offers           []catalog.Offer            `json:"offers,omitempty"`
	Orders           []orders.Order             `json:"orders,omitempty"`
	Cases            []support.Case             `json:"cases,omitempty"`
	Assignments      []support.Assignment       `json:"assignments,omitempty"`
	Transfers        []finance.Transfer         `json:"transfers,omitempty"`
	Ledger           []finance.Entry            `json:"ledger,omitempty"`
	Rules            []Rule                     `json:"rules,omitempty"`
	Shipping         []Shipping                 `json:"shipping,omitempty"`
	Notifications    []Notification             `json:"notifications,omitempty"`
	Audit            []Audit                    `json:"audit,omitempty"`
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
	Phone          string             `json:"phone"`
	Action         string             `json:"action"`
	Target         string             `json:"target"`
	Name           string             `json:"name"`
	Value          string             `json:"value"`
	Reason         string             `json:"reason"`
	Amount         int64              `json:"amount"`
	Version        int                `json:"version"`
	IdempotencyKey string             `json:"idempotency_key"`
	Scope          identity.RoleScope `json:"scope,omitempty"`
	Permissions    []string           `json:"permissions,omitempty"`
	StoreID        string             `json:"store_id,omitempty"`
}

// Update must execute the callback atomically and roll back its writes on error.
type Repository interface {
	Snapshot(context.Context) (Workspace, error)
	Update(context.Context, func(*Workspace) error) error
}
type Clock interface{ Now() time.Time }
type Identifier interface{ Token() (string, error) }
