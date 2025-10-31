package tools

import "context"

// Check Permissions for the Given Session (Queries Database)
func CheckPermissions(ctx context.Context, session SessionData, permissions ...scopeInfo) (bool, error) {
	var user DatabaseUser
	err := Database.
		QueryRow(ctx, "SELECT permissions FROM auth.users WHERE id = $1", session.UserID).
		Scan(&user.Permissions)
	if err != nil {
		return false, err
	}
	for _, p := range permissions {
		if (user.Permissions & p.Flag) == 0 {
			return false, nil
		}
	}
	return true, nil
}
