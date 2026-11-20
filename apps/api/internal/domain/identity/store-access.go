package identity

import "slices"

func MemberStoreIDs(memberships []StoreMembership, userID string) []string {
	var storeIDs []string
	for _, membership := range memberships {
		if membership.UserID == userID && !slices.Contains(storeIDs, membership.StoreID) {
			storeIDs = append(storeIDs, membership.StoreID)
		}
	}
	return storeIDs
}

func AllowsStore(memberships []StoreMembership, roles []RoleDefinition, grants []RoleGrant, userID, storeID string, permission StorePermission) bool {
	return storeID != "" && slices.Contains(memberships, StoreMembership{UserID: userID, StoreID: storeID}) &&
		slices.Contains(GrantedPermissions(roles, grants, userID, storeID), string(permission))
}

func GrantedPermissions(roles []RoleDefinition, grants []RoleGrant, userID, storeID string) []string {
	scope := ScopeGlobal
	if storeID != "" {
		scope = ScopeStore
	}
	var permissions []string
	for _, grant := range grants {
		if grant.UserID != userID || grant.StoreID != storeID {
			continue
		}
		for _, role := range roles {
			if role.ID == grant.RoleID && role.Scope == scope {
				permissions = append(permissions, role.Permissions...)
				break
			}
		}
	}
	slices.Sort(permissions)
	return slices.Compact(permissions)
}
