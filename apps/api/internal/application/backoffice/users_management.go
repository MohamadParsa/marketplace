package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"sort"
	"strings"
)

var ErrUserContactConflict = errors.New("duplicate_user_contact")
var ErrLastAdministrator = errors.New("last_administrator")

const UsersRead = "users.read"
const UsersWrite = "users.write"
const UsersPasswordReset = "users.password.reset"

type UserInput struct {
	ID                string `json:"id"`
	Version           int    `json:"version"`
	Phone             string `json:"phone"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
	Gender            string `json:"gender"`
	Active            bool   `json:"active"`
	TemporaryPassword string `json:"temporary_password"`
}
type AdminUserRow struct {
	ID                     string               `json:"id"`
	Name                   string               `json:"name"`
	FirstName              string               `json:"first_name"`
	LastName               string               `json:"last_name"`
	Phone                  string               `json:"phone"`
	Email                  string               `json:"email"`
	Gender                 string               `json:"gender"`
	Active                 bool                 `json:"active"`
	Version                int                  `json:"version"`
	BaseRole               identity.Role        `json:"base_role"`
	PasswordChangeRequired bool                 `json:"password_change_required"`
	HasPassword            bool                 `json:"has_password"`
	StoreIDs               []string             `json:"store_ids"`
	Grants                 []identity.RoleGrant `json:"grants"`
}

// Private retry evidence; never included in an HTTP directory or public projection.
type UserMutation struct {
	Key           string       `json:"key"`
	Actor         string       `json:"actor"`
	Fingerprint   string       `json:"fingerprint"`
	PasswordCheck string       `json:"password_check,omitempty"`
	Result        AdminUserRow `json:"result"`
}
type UserFilter struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Active    string `json:"active"`
	Gender    string `json:"gender"`
	Role      string `json:"role"`
	Store     string `json:"store"`
	Sort      string `json:"sort"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}
type UserDirectory struct {
	Rows             []AdminUserRow            `json:"rows"`
	Total            int                       `json:"total"`
	Page             int                       `json:"page"`
	Size             int                       `json:"size"`
	CanWrite         bool                      `json:"can_write"`
	CanResetPassword bool                      `json:"can_reset_password"`
	CanAssign        bool                      `json:"can_assign"`
	Roles            []identity.RoleDefinition `json:"roles"`
}

func userRow(w Workspace, u User) AdminUserRow {
	r := AdminUserRow{ID: u.ID, Name: u.Name, FirstName: u.FirstName, LastName: u.LastName, Phone: u.Phone, Email: u.Email, Gender: u.Gender, Active: u.Active, Version: u.Version, BaseRole: u.Role, HasPassword: u.PasswordHash != "", PasswordChangeRequired: u.PasswordChangeRequired, StoreIDs: identity.MemberStoreIDs(w.StoreMemberships, u.ID), Grants: []identity.RoleGrant{}}
	if r.StoreIDs == nil {
		r.StoreIDs = []string{}
	}
	if w.Access != nil {
		for _, g := range w.Access.Grants {
			if g.UserID == u.ID {
				r.Grants = append(r.Grants, g)
			}
		}
	}
	return r
}
func (s *Service) identityState(ctx context.Context) (Workspace, error) {
	if r, ok := s.repository.(interface {
		IdentitySnapshot(context.Context) (Workspace, error)
	}); ok {
		return r.IdentitySnapshot(ctx)
	}
	return s.repository.Snapshot(ctx)
}
func requireUserAdmin(w Workspace, a identity.User, permission string) (identity.User, error) {
	if a.Surface != identity.SurfaceBackoffice {
		return identity.User{}, ErrForbidden
	}
	actor, e := currentActor(w, a)
	if e != nil || !(allowsAction(actor, permission) || (permission == UsersRead && allowsAction(actor, CommerceWalletCredit))) {
		return actor, ErrForbidden
	}
	return actor, nil
}
func foldUser(v string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(v), "ي", "ی"), "ك", "ک"))
}
func matchUser(w Workspace, u User, f UserFilter) bool {
	first := u.FirstName
	if first == "" && u.LastName == "" {
		first = u.Name
	}
	for _, v := range [][2]string{{u.ID, f.ID}, {first, f.FirstName}, {u.LastName, f.LastName}, {u.Phone, identity.NormalizePhone(f.Phone)}, {u.Email, f.Email}} {
		if v[1] != "" && !strings.Contains(foldUser(v[0]), foldUser(v[1])) {
			return false
		}
	}
	if f.ID != "" && u.ID != f.ID {
		return false
	}
	if f.Active != "" && u.Active != (f.Active == "true") {
		return false
	}
	if f.Gender != "" && u.Gender != f.Gender {
		return false
	}
	if f.Store != "" {
		found := false
		for _, m := range w.StoreMemberships {
			found = found || (m.UserID == u.ID && m.StoreID == f.Store)
		}
		if !found {
			return false
		}
	}
	if f.Role != "" {
		found := string(u.Role) == f.Role
		if w.Access != nil {
			for _, g := range w.Access.Grants {
				found = found || (g.UserID == u.ID && g.RoleID == f.Role && (f.Store == "" || g.StoreID == "" || g.StoreID == f.Store))
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func (s *Service) ListUsers(ctx context.Context, a identity.User, f UserFilter, export bool, ids []string) (UserDirectory, error) {
	var out UserDirectory
	if f.Page < 0 || f.Size < 0 || f.Size > 100 || (f.Active != "" && f.Active != "true" && f.Active != "false") || (f.Gender != "" && f.Gender != "male" && f.Gender != "female" && f.Gender != "unspecified") || (f.Sort != "" && f.Sort != "name" && f.Sort != "name-desc" && f.Sort != "id") {
		return out, ErrInvalid
	}
	for _, v := range []string{f.ID, f.FirstName, f.LastName, f.Phone, f.Email, f.Role, f.Store} {
		if len(v) > 254 {
			return out, ErrInvalid
		}
	}
	if len(ids) > 10000 {
		return out, ErrInvalid
	}
	w, e := s.identityState(ctx)
	if e != nil {
		return out, e
	}
	actor, e := requireUserAdmin(w, a, UsersRead)
	if e != nil {
		return out, e
	}
	if export {
		if e = s.AuthorizeGridExport(ctx, a); e != nil {
			return out, e
		}
	}
	out = UserDirectory{Rows: []AdminUserRow{}, Page: f.Page, Size: f.Size, CanWrite: allowsAction(actor, UsersWrite), CanResetPassword: allowsAction(actor, UsersPasswordReset), CanAssign: actor.Role == identity.RoleSuperAdmin, Roles: []identity.RoleDefinition{}}
	if out.Size == 0 {
		out.Size = 20
	}
	if w.Access != nil {
		for _, r := range w.Access.Roles {
			r.Permissions = nil
			out.Roles = append(out.Roles, r)
		}
	}
	wanted := map[string]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	for _, u := range w.Users {
		check := f
		if f.Gender == "unspecified" {
			if u.Gender != "" {
				continue
			}
			check.Gender = ""
		}
		if matchUser(w, u, check) && (!export || len(ids) == 0 || wanted[u.ID]) {
			out.Rows = append(out.Rows, userRow(w, u))
		}
	}
	sort.Slice(out.Rows, func(i, j int) bool {
		a, b := out.Rows[i], out.Rows[j]
		if f.Sort == "id" || a.Name == b.Name {
			return a.ID < b.ID
		}
		if f.Sort == "name-desc" {
			return a.Name > b.Name
		}
		return a.Name < b.Name
	})
	out.Total = len(out.Rows)
	if export {
		if out.Total > 10000 {
			return UserDirectory{}, ErrInvalid
		}
		return out, nil
	}
	if out.Page > out.Total/out.Size {
		out.Page = out.Total / out.Size
	}
	if out.Total > 0 && out.Page*out.Size == out.Total {
		out.Page--
	}
	start := out.Page * out.Size
	end := start + out.Size
	if end > out.Total {
		end = out.Total
	}
	out.Rows = out.Rows[start:end]
	return out, nil
}
func (s *Service) SaveUser(ctx context.Context, a identity.User, in UserInput, key string) (AdminUserRow, error) {
	var result AdminUserRow
	in.Phone = identity.NormalizePhone(in.Phone)
	in.Email = identity.NormalizeEmail(in.Email)
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	if len(key) == 0 || len(key) > 128 || !accountPhone.MatchString(in.Phone) || len(in.FirstName) > 160 || len(in.LastName) > 160 || len(in.Email) > 254 || in.Version < 0 || (in.Gender != "" && in.Gender != "male" && in.Gender != "female") {
		return result, ErrInvalid
	}
	if in.Email != "" {
		v, e := mail.ParseAddress(in.Email)
		if e != nil || v.Address != in.Email {
			return result, ErrInvalid
		}
	}
	if in.TemporaryPassword != "" && (len(in.TemporaryPassword) < 8 || len(in.TemporaryPassword) > 72) {
		return result, ErrInvalid
	}
	// Check authority before running the relatively expensive password hash.
	w, e := s.identityState(ctx)
	if e != nil {
		return result, e
	}
	if _, e = requireUserAdmin(w, a, UsersWrite); e != nil {
		return result, e
	}
	password := in.TemporaryPassword
	in.TemporaryPassword = ""
	raw, _ := json.Marshal(in)
	digest := sha256.Sum256(raw)
	fingerprint := hex.EncodeToString(digest[:])
	hash := ""
	if password != "" {
		if _, e = requireUserAdmin(w, a, UsersPasswordReset); e != nil {
			return result, e
		}
		b, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if e != nil {
			return result, e
		}
		hash = string(b)
	}
	id, e := s.identifiers.Token()
	if e != nil {
		return result, e
	}
	e = s.repository.Update(ctx, func(w *Workspace) error {
		actor, e := requireUserAdmin(*w, a, UsersWrite)
		if e != nil {
			return e
		}
		if hash != "" {
			if _, e = requireUserAdmin(*w, a, UsersPasswordReset); e != nil {
				return e
			}
		}
		for _, op := range w.UserMutations {
			if op.Key == key {
				if op.Actor != a.ID || op.Fingerprint != fingerprint || (password == "") != (op.PasswordCheck == "") || (password != "" && bcrypt.CompareHashAndPassword([]byte(op.PasswordCheck), []byte(password)) != nil) {
					return ErrConflict
				}
				result = op.Result
				return nil
			}
		}
		index := findUser(w, in.ID)
		if in.ID != "" && index < 0 {
			return ErrNotFound
		}
		for _, u := range w.Users {
			if u.ID != in.ID && (u.Phone == in.Phone || (in.Email != "" && strings.EqualFold(u.Email, in.Email))) {
				return ErrUserContactConflict
			}
		}
		u := User{ID: id, Phone: in.Phone, Name: "مشتری", Active: in.Active, Version: 0, AccessVersion: 1}
		if index >= 0 {
			u = w.Users[index]
			if u.Version != in.Version {
				return ErrConflict
			}
			if u.Phone != in.Phone {
				return ErrInvalid
			}
			if !u.Active && in.Active && u.Role == identity.RoleSuperAdmin {
				return ErrReauthenticationRequired
			}
			protected := u.Role != ""
			if w.Access != nil {
				for _, g := range w.Access.Grants {
					protected = protected || (g.UserID == u.ID && g.StoreID == "")
				}
			}
			if protected && actor.Role != identity.RoleSuperAdmin {
				return ErrForbidden
			}
			if u.Active && !in.Active && u.Role == identity.RoleSuperAdmin {
				count := 0
				for _, other := range w.Users {
					if other.Active && other.Role == identity.RoleSuperAdmin {
						count++
					}
				}
				if count <= 1 {
					return ErrLastAdministrator
				}
			}
		} else if in.Version != 0 {
			return ErrConflict
		}
		u.FirstName, u.LastName, u.Email, u.Gender = in.FirstName, in.LastName, in.Email, in.Gender
		if name := strings.TrimSpace(in.FirstName + " " + in.LastName); name != "" {
			u.Name = name
		}
		if u.Active != in.Active {
			u.AccessVersion++
		}
		u.Active = in.Active
		if hash != "" {
			u.PasswordHash = hash
			u.PasswordChangeRequired = true
			u.AccessVersion++
		}
		u.Version++
		if index < 0 {
			w.Users = append(w.Users, u)
		} else {
			w.Users[index] = u
		}
		result = userRow(*w, u)
		w.UserMutations = append(w.UserMutations, UserMutation{Key: key, Actor: a.ID, Fingerprint: fingerprint, PasswordCheck: hash, Result: result})
		w.Audit = append(w.Audit, Audit{ID: key, Actor: a.ID, Action: "users.save", Target: u.ID, At: s.clock.Now()})
		return nil
	})
	return result, e
}
