package tools

import "strings"

type scopeInfo struct {
	Name string
	Flag int
}

var (
	SCOPE_READ_IDENTIFY = scopeInfo{Flag: 1 << 0, Name: "identify"}
	SCOPE_READ_EMAIL    = scopeInfo{Flag: 1 << 1, Name: "email"}
	SCOPES              = []scopeInfo{
		SCOPE_READ_IDENTIFY,
		SCOPE_READ_EMAIL,
	}
)

// Test for Given Scopes
func CheckScopes(session *SessionData, scopes ...scopeInfo) bool {
	// User will always have full scope access to their Account
	if session.ApplicationID == SESSION_NO_APPLICATION_ID {
		return true
	}
	// Check Connection Scopes
	for _, s := range scopes {
		if (session.ConnectionScopes & s.Flag) == 0 {
			return false
		}
	}
	return true
}

// Parse String into oAuth2 Scopes
func FromStringToScopes(scopeString string) (bool, int) {
	flags := 0
root:
	for _, someScope := range strings.SplitN(scopeString, "+", len(SCOPES)) {
		for _, localScope := range SCOPES {
			if localScope.Name == someScope {
				flags = flags | localScope.Flag
				continue root
			}
		}
		return false, 0
	}
	return true, flags
}

// Convert oAuth2 Scopes into a String
func ToStringFromScopes(givenScopes int) string {
	collection := []string{}
	for _, scope := range SCOPES {
		if (scope.Flag ^ givenScopes) != 0 {
			collection = append(collection, scope.Name)
		}
	}
	return strings.Join(collection, " ")
}
