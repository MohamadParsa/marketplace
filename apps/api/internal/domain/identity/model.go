package identity

import "time"

type Surface string

const (
	SurfaceBackoffice Surface = "backoffice"
	SurfaceSeller     Surface = "seller"
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
	ID      string
	Phone   string
	Surface Surface
	Role    Role
	StoreID string
	Active  bool
}
type Session struct {
	Token   string
	User    User
	Expires time.Time
}
type Challenge struct {
	Phone    string
	Surface  Surface
	Code     string
	Expires  time.Time
	Attempts int
	IssuedAt time.Time
}
