package identity

import "time"

type Surface string

const (
	SurfaceBackoffice Surface = "backoffice"
	SurfaceSeller     Surface = "seller"
	SurfaceCustomer   Surface = "storefront"
)

type Role string

const (
	RoleSuperAdmin Role = "super-admin"
	RoleOperations Role = "operations"
	RoleSupport    Role = "support"
	RoleFinance    Role = "finance"
	RoleViewer     Role = "viewer"
)

type User struct {
	ID               string
	Name             string
	Phone            string
	Email            string
	Surface          Surface
	Role             Role
	StoreID          string
	StoreIDs         []string            `json:",omitempty"`
	Permissions      []string            `json:",omitempty"`
	StorePermissions map[string][]string `json:",omitempty"`
	Active           bool
	AccessVersion    int    `json:"-"`
	PasswordHash     string `json:"-"`
	NationalID       string `json:"-"`
}

type StoreMembership struct {
	UserID  string `json:"user_id"`
	StoreID string `json:"store_id"`
}

type StorePermission string

const (
	StoreOffersRead StorePermission = "offers.read"
	StoreStockWrite StorePermission = "stock.write"
)

type RoleScope string

const (
	ScopeGlobal RoleScope = "global"
	ScopeStore  RoleScope = "store"
)

type RoleDefinition struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Scope       RoleScope `json:"scope"`
	Permissions []string  `json:"permissions"`
	Version     int       `json:"version"`
}

type RoleGrant struct {
	UserID  string `json:"user_id"`
	RoleID  string `json:"role_id"`
	StoreID string `json:"store_id,omitempty"`
}
type Session struct {
	Token   string
	User    User
	Expires time.Time
}
type Challenge struct {
	Phone         string
	Destination   string
	Surface       Surface
	Purpose       string
	Code          string
	Expires       time.Time
	Attempts      int
	IssuedAt      time.Time
	ReauthSession string
}

func (c Challenge) CooldownKey() string {
	key := string(c.Surface) + ":" + c.Phone
	if c.ReauthSession != "" {
		return key + ":admin"
	}
	return key
}
