package tools

import (
	"net/url"
	"strings"
)

type ScopeInfo struct {
	Name string
	Flag int
}

var (
	SCOPE_READ_IDENTIFY = ScopeInfo{Flag: 1 << 0, Name: "identify"}
	SCOPE_READ_EMAIL    = ScopeInfo{Flag: 1 << 1, Name: "email"}
	SCOPE_HASH          = map[string]ScopeInfo{
		SCOPE_READ_IDENTIFY.Name: SCOPE_READ_IDENTIFY,
		SCOPE_READ_EMAIL.Name:    SCOPE_READ_EMAIL,
	}
)

// Test for Given Scopes
func OAuth2ScopesContains(session *SessionData, scopes ...ScopeInfo) bool {
	if session.ApplicationID == SESSION_NO_APPLICATION_ID {
		// The User will always have full access to their account
		return true
	}
	// Find missing scope
	for _, s := range scopes {
		if (session.ConnectionScopes & s.Flag) == 0 {
			return false
		}
	}
	return true
}

// Convert oAuth2 Scopes into a String
func OAuth2ScopesToString(givenScopes int) string {
	scopes := make([]string, 0, len(SCOPE_HASH))
	for _, sc := range SCOPE_HASH {
		if (givenScopes & sc.Flag) != 0 {
			scopes = append(scopes, sc.Name)
		}
	}
	return strings.Join(scopes, " ")
}

// Convert String into OAuth2 Scopes
func OAuth2StringToScopes(s string) (bool, int) {
	scopes := strings.Fields(s)
	unique := make(map[string]struct{}, len(scopes))
	flags := 0
	for _, sc := range scopes {
		f, ok := SCOPE_HASH[sc]
		if !ok {
			// Prevent Unknown per standard
			return false, 0
		}
		if _, ok := unique[sc]; ok {
			// Prevent Duplicates per standard
			return false, 0
		}
		unique[sc] = struct{}{}
		flags = flags | f.Flag
	}
	return true, flags
}

// Compare Given URI to list of allowed URIs
func OAauth2ScopesValidateRedirectURI(given string, allowlist []string) (bool, string) {
	givenParsed, err := url.Parse(given)
	if err != nil {
		return false, ""
	}
	for _, allowed := range allowlist {
		allowParsed, err := url.Parse(allowed)
		if err != nil {
			continue
		}
		if givenParsed.Scheme == allowParsed.Scheme &&
			givenParsed.Host == allowParsed.Host &&
			givenParsed.Path == allowParsed.Path {
			return true, allowed
		}
	}
	return false, ""
}
