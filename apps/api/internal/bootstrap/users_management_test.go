package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"strings"
	"testing"
)

func TestUserManagementTemporaryPasswordAndDirectoryPrivacy(t *testing.T) {
	ctx := context.Background()
	seed := commerceSeed()
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	ops := login(t, h, "09000000002")
	in := backoffice.UserInput{Phone: "۰۹۱۲۰۰۰۰۹۹۱", FirstName: "نام", LastName: "خانوادگی آزمون", Email: "Person@Example.test", Gender: "female", Active: true, TemporaryPassword: "Synthetic-Initial-91"}
	commerceHTTP(t, h, ops, "", "POST", "/backoffice/users/save", in, "unauthorized", 403)
	raw := commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", in, "create-account", 200)
	var u backoffice.AdminUserRow
	json.Unmarshal(raw, &u)
	if u.StoreIDs == nil || u.ID == "" || u.Phone != "09120000991" || u.Email != "person@example.test" || u.Version != 1 || u.BaseRole != "" || !u.PasswordChangeRequired {
		t.Fatal("created user", u.ID)
	}
	again := commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", in, "create-account", 200)
	if string(raw) != string(again) {
		t.Fatal("retry changed response")
	}
	in.TemporaryPassword = "Different-secret-91"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", in, "create-account", 409)
	in.TemporaryPassword = "Synthetic-Initial-91"
	state, _ := repo.Snapshot(ctx)
	for _, v := range state.Users {
		if v.ID == u.ID {
			if v.PasswordHash == in.TemporaryPassword || bcrypt.CompareHashAndPassword([]byte(v.PasswordHash), []byte(in.TemporaryPassword)) != nil {
				t.Fatal("credential not hashed")
			}
		}
	}
	payload := commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/query", map[string]any{"filter": backoffice.UserFilter{Email: "PERSON@EXAMPLE.TEST", Gender: "female", Active: "true"}}, "", 200)
	var list backoffice.UserDirectory
	json.Unmarshal(payload, &list)
	if list.Total != 1 || list.Rows[0].ID != u.ID {
		t.Fatal("combined filters")
	}
	workspace := commerceHTTP(t, h, admin, "", "GET", "/backoffice/workspace", nil, "", 200)
	var shell backoffice.Workspace
	json.Unmarshal(commerceHTTP(t, h, admin, "", "GET", "/backoffice/workspace?user_directory=separate", nil, "", 200), &shell)
	if len(shell.Users) > 0 {
		t.Fatal("new shell downloaded the full user directory")
	}
	for _, data := range [][]byte{raw, payload, workspace} {
		for _, secret := range []string{in.TemporaryPassword, "$2a$", "$2b$", "password_check", "user_mutations"} {
			if strings.Contains(string(data), secret) {
				t.Fatal("private credential retry data leaked")
			}
		}
	}
	// A customer and staff identity is the same account; temporary sessions cannot read either surface.
	pwd := map[string]string{"phone": u.Email, "password": in.TemporaryPassword}
	var restricted map[string]any
	json.Unmarshal(commerceHTTP(t, h, "", "", "POST", "/storefront/auth/password-login", pwd, "", 200), &restricted)
	token := restricted["token"].(string)
	commerceHTTP(t, h, token, "", "GET", "/storefront/account/me", nil, "", 401)
	commerceHTTP(t, h, token, "", "POST", "/storefront/auth/complete-password", map[string]string{"password": in.TemporaryPassword}, "", 400)
	var completed map[string]any
	json.Unmarshal(commerceHTTP(t, h, token, "", "POST", "/storefront/auth/complete-password", map[string]string{"password": "Personal-Synthetic-92"}, "", 200), &completed)
	commerceHTTP(t, h, token, "", "GET", "/storefront/account/me", nil, "", 401)
	commerceHTTP(t, h, completed["token"].(string), "", "GET", "/storefront/account/me", nil, "", 200)
	commerceHTTP(t, h, "", "", "POST", "/storefront/auth/password-login", pwd, "", 401)
	// Resetting it again revokes all previous sessions; OTP must also finish the password gate.
	in.ID = u.ID
	in.Version = 2
	in.Phone = u.Phone
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", in, "reset-account", 200)
	commerceHTTP(t, h, completed["token"].(string), "", "GET", "/storefront/account/me", nil, "", 401)
	otp := loginAt(t, h, "storefront", u.Phone)
	commerceHTTP(t, h, otp["token"].(string), "", "GET", "/storefront/account/me", nil, "", 401)
	commerceHTTP(t, h, otp["token"].(string), "", "POST", "/storefront/auth/complete-password", map[string]string{"password": "Another-Personal-93"}, "", 200)
}
func TestUserManagementConflictLegacyLastAdminAndScopedRoles(t *testing.T) {
	ctx := context.Background()
	repo := backofficememory.New(commerceSeed())
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	ops := login(t, h, "09000000002")
	state, _ := repo.Snapshot(ctx)
	var target backoffice.User
	for _, u := range state.Users {
		if u.ID == "admin" {
			target = u
		}
	}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", backoffice.UserInput{ID: target.ID, Version: target.Version, Phone: target.Phone, Active: false}, "last-admin", 409)
	repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Users = append(w.Users, backoffice.User{ID: "inactive-admin", Phone: "09120000888", Name: "ادمین غیرفعال آزمون", Role: identity.RoleSuperAdmin})
		return nil
	})
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", backoffice.UserInput{ID: "inactive-admin", Version: 0, Phone: "09120000888", Active: true}, "cannot-reactivate-via-profile", 403)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", backoffice.Command{Action: "user.toggle", Target: "inactive-admin", IdempotencyKey: "reactivation-needs-otp"}, "", 403)
	in := backoffice.UserInput{Phone: "09120000992", Active: true, FirstName: "مشترک", Email: "shared@example.test"}
	var row backoffice.AdminUserRow
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", in, "shared-account", 200), &row)
	duplicate := in
	duplicate.Phone = "09120000993"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", duplicate, "duplicate-email", 409)
	in.ID = row.ID
	in.Version = row.Version
	in.Phone = "09120000994"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", in, "change-phone", 400)
	in.Phone = row.Phone
	in.Version = 0
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", in, "stale-user", 409)
	// Existing command RBAC, membership and multi-role union are reused.
	command := func(c backoffice.Command, status int) {
		commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", c, "", status)
	}
	role := identity.RoleDefinition{ID: "test-role", Name: "دو مجوز", Scope: identity.ScopeStore, Permissions: []string{"offers.read", "stock.write"}, Version: 1}
	repo.Update(ctx, func(w *backoffice.Workspace) error { w.Access.Roles = append(w.Access.Roles, role); return nil })
	v := row.Version
	command(backoffice.Command{Action: "user.role-grant", Target: row.ID, Value: role.ID, StoreID: "s1", UserVersion: &v, IdempotencyKey: "not-member"}, 400)
	command(backoffice.Command{Action: "membership.add", Target: row.ID, StoreID: "s1", UserVersion: &v, IdempotencyKey: "membership-one"}, 204)
	v++
	staleRoleVersion := 0
	command(backoffice.Command{Action: "user.role-grant", Target: row.ID, Value: role.ID, StoreID: "s1", UserVersion: &v, RoleVersion: &staleRoleVersion, IdempotencyKey: "stale-role-definition"}, 409)
	command(backoffice.Command{Action: "user.role-grant", Target: row.ID, Value: role.ID, StoreID: "s1", UserVersion: &v, IdempotencyKey: "grant-one"}, 204)
	command(backoffice.Command{Action: "membership.add", Target: row.ID, StoreID: "s2", UserVersion: &v, IdempotencyKey: "stale-membership"}, 409)
	v++
	command(backoffice.Command{Action: "membership.add", Target: row.ID, StoreID: "s2", UserVersion: &v, IdempotencyKey: "membership-two"}, 204)
	seller := loginAt(t, h, "seller", row.Phone)["token"].(string)
	commerceHTTP(t, h, seller, "", "GET", "/seller/stores/s2/offers", nil, "", 403)
	var grants backoffice.Workspace
	json.Unmarshal(commerceHTTP(t, h, seller, "", "GET", "/seller/workspace", nil, "", 200), &grants)
	if grants.Access != nil || len(grants.Users) > 0 {
		t.Fatal("seller saw global directory")
	}
	command(backoffice.Command{Action: "role.delete", Target: role.ID, Version: 1, Scope: role.Scope, IdempotencyKey: "delete-used"}, 409)
	commerceHTTP(t, h, ops, "", "POST", "/backoffice/commands", backoffice.Command{Action: "role.create", Name: "ممنوع", Scope: identity.ScopeGlobal, Permissions: []string{"users.write"}, IdempotencyKey: "forbidden-role"}, "", 403)
	command(backoffice.Command{Action: "role.update", Target: role.ID, Version: 0, Scope: role.Scope, Name: role.Name, Permissions: role.Permissions, IdempotencyKey: "stale-role"}, 409)
	// Removing a store grant takes effect without extending its authority to the other store.
	v++
	command(backoffice.Command{Action: "user.role-revoke", Target: row.ID, Value: role.ID, StoreID: "s1", UserVersion: &v, IdempotencyKey: "revoke-one"}, 204)
	commerceHTTP(t, h, seller, "", "GET", "/seller/stores/s1/offers", nil, "", 403)
}
func TestUserManagementPostgresRestartPreservesExistingIdentityAndFinance(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	w, _ := repo.Snapshot(ctx)
	var adminPhone string
	for _, u := range w.Users {
		if u.Role == identity.RoleSuperAdmin {
			adminPhone = u.Phone
		}
	}
	if adminPhone == "" {
		t.Fatal("admin fixture")
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, adminPhone)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commerce/wallet/credits", backoffice.AdminCreditInput{UserID: w.Users[0].ID, AmountToman: "1234", Title: "اعتبار ساختگی", Description: "آزمون حفظ دفتر هنگام ویرایش کاربران", Withdrawable: false}, "preserve-finance-seed", 200)
	w, _ = repo.Snapshot(ctx)
	in := backoffice.UserInput{Phone: "09120000995", FirstName: "پایدار", LastName: "آزمون", Gender: "male", Email: "persist@example.test", Active: true, TemporaryPassword: "Persist-Synthetic-95"}
	var row backoffice.AdminUserRow
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", in, "pg-user", 200), &row)
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, e := reopened.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(w.Commerce, after.Commerce) || !reflect.DeepEqual(w.Orders, after.Orders) || !reflect.DeepEqual(w.Ledger, after.Ledger) || !reflect.DeepEqual(w.Transfers, after.Transfers) || !reflect.DeepEqual(w.Access, after.Access) {
		t.Fatal("identity mutation rewrote financial/access history")
	}
	for _, old := range w.Users {
		found := false
		for _, now := range after.Users {
			if old.ID == now.ID {
				found = true
				if !reflect.DeepEqual(old, now) {
					t.Fatal("existing identity/credential changed")
				}
			}
		}
		if !found {
			t.Fatal("old identity disappeared")
		}
	}
	if len(after.Users) != len(w.Users)+1 {
		t.Fatal("existing users lost")
	}
	found := false
	for _, u := range after.Users {
		if u.ID == row.ID {
			found = u.FirstName == in.FirstName && u.LastName == in.LastName && u.Version == 1 && u.PasswordChangeRequired && u.Gender == "male"
		}
	}
	if !found {
		t.Fatal("new identity fields not persisted")
	}
	h2 := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, reopened)
	admin2 := login(t, h2, adminPhone)
	commerceHTTP(t, h2, admin2, "", "POST", "/backoffice/users/save", in, "pg-user", 200)
	public, e := reopened.PublicCatalog(ctx, "")
	if e != nil {
		t.Fatal(e)
	}
	if len(public.Users) > 0 || len(public.UserMutations) > 0 {
		t.Fatal("private user data in catalog projection")
	}
}

func TestUserDirectoryPagingExportAndDelegatedPermissionLimits(t *testing.T) {
	ctx := context.Background()
	seed := commerceSeed()
	for i := 0; i < 125; i++ {
		seed.Users = append(seed.Users, backoffice.User{ID: fmt.Sprintf("directory-%03d", i), Name: fmt.Sprintf("نام قبلی %03d", i), Phone: fmt.Sprintf("0913%07d", i), Active: true, Gender: "female"})
	}
	seed.Access.Roles = append(seed.Access.Roles, identity.RoleDefinition{ID: "user-editor", Name: "ویرایش مشتری", Scope: identity.ScopeGlobal, Version: 1, Permissions: []string{backoffice.UsersRead, backoffice.UsersWrite}})
	seed.Access.Grants = append(seed.Access.Grants, identity.RoleGrant{UserID: "viewer", RoleID: "user-editor"})
	repo := backofficememory.New(seed)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	editor := login(t, h, "09000000005")
	var page backoffice.UserDirectory
	input := map[string]any{"filter": backoffice.UserFilter{FirstName: "نام قبلی", Gender: "female", Page: 1, Size: 20}}
	json.Unmarshal(commerceHTTP(t, h, editor, "", "POST", "/backoffice/users/query", input, "", 200), &page)
	if page.Total != 125 || len(page.Rows) != 20 || page.Page != 1 || !page.CanWrite || page.CanResetPassword || page.CanAssign {
		t.Fatal("directory paging/permissions")
	}
	export := map[string]any{"filter": backoffice.UserFilter{FirstName: "نام قبلی", Gender: "female"}}
	commerceHTTP(t, h, editor, "", "POST", "/backoffice/users/export", export, "", 403)
	var all backoffice.UserDirectory
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/export", export, "", 200), &all)
	if len(all.Rows) != 125 {
		t.Fatal("export only loaded page")
	}
	export["ids"] = []string{all.Rows[0].ID, all.Rows[22].ID, "not-visible"}
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/export", export, "", 200), &all)
	if len(all.Rows) != 2 {
		t.Fatal("export selection scope")
	}
	// A delegated identity editor cannot change a staff account or set any credential without its distinct permission.
	state, _ := repo.Snapshot(ctx)
	for _, u := range state.Users {
		if u.ID == "admin" {
			commerceHTTP(t, h, editor, "", "POST", "/backoffice/users/save", backoffice.UserInput{ID: u.ID, Version: u.Version, Phone: u.Phone, Active: true, FirstName: "ممنوع"}, "edit-admin", 403)
		}
	}
	commerceHTTP(t, h, editor, "", "POST", "/backoffice/users/save", backoffice.UserInput{Phone: "09120000997", Active: true, TemporaryPassword: "Not-Authorized-97"}, "credential-permission", 403)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", map[string]any{"phone": "09120000997", "version": 0}, "missing-active", 400)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", map[string]any{"phone": "09120000997", "version": 0, "active": true, "permissions": []string{"super-admin"}}, "injected-grant", 400)
	var unnamed backoffice.AdminUserRow
	json.Unmarshal(commerceHTTP(t, h, admin, "", "POST", "/backoffice/users/save", backoffice.UserInput{Phone: "09120000999", Active: true}, "optional-names", 200), &unnamed)
	if unnamed.Name != "مشتری" || unnamed.FirstName != "" || unnamed.LastName != "" {
		t.Fatal("unnamed accounts must retain the existing incomplete-profile placeholder")
	}
}
