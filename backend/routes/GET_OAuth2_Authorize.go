package routes

import (
	"net/http"

	"github.com/bakonpancakzz/template-auth/tools"

	"github.com/jackc/pgx/v5"
)

func GET_OAuth2_Authorize(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if session.ApplicationID != tools.SESSION_NO_APPLICATION_ID {
		tools.SendClientError(w, r, tools.ERROR_OAUTH2_USERS_ONLY)
		return
	}
	var Body struct {
		State        *string `query:"state"`
		ClientID     int64   `query:"client_id" validate:"required"`
		ResponseType string  `query:"response_type" validate:"required"`
		RedirectURI  string  `query:"redirect_uri" validate:"required,uri"`
		ScopesString string  `query:"scope" validate:"required"`
	}
	if !tools.ValidateQuery(w, r, &Body) {
		return
	}
	if Body.ResponseType != "code" {
		tools.SendClientError(w, r, tools.ERROR_OAUTH2_FORM_INVALID_RESPONSE_TYPE)
		return
	}

	// Parse Scopes
	ok, requestedScopes := tools.FromStringToScopes(Body.ScopesString)
	if !ok {
		tools.SendClientError(w, r, tools.ERROR_OAUTH2_FORM_INVALID_SCOPE)
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Fetch Relevant Application
	var application tools.DatabaseApplication
	err := tools.Database.QueryRow(ctx,
		`SELECT 
			id, created, name, icon_hash, auth_redirects 
		FROM auth.applications 
		WHERE id = $1`,
		Body.ClientID,
	).Scan(
		&application.ID,
		&application.Created,
		&application.Name,
		&application.IconHash,
		&application.AuthRedirects,
	)
	if err == pgx.ErrNoRows {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_APPLICATION)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Ensure Redirect URI is allowed
	requestedRedirect := ""
	for _, someURI := range application.AuthRedirects {
		if someURI == Body.RedirectURI {
			requestedRedirect = someURI
			break
		}
	}
	if requestedRedirect == "" {
		tools.SendClientError(w, r, tools.ERROR_OAUTH2_FORM_INVALID_REDIRECT_URI)
		return
	}

	// Fetch Profile for Account
	var profile tools.DatabaseProfile
	err = tools.Database.QueryRow(ctx,
		`SELECT 
			id, displayname, avatar_hash 
		FROM auth.profiles 
		WHERE user_id = $1`,
		session.UserID,
	).Scan(
		&profile.ID,
		&profile.Displayname,
		&profile.AvatarHash,
	)
	if err == pgx.ErrNoRows {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Organize Application and Profile
	tools.SendJSON(w, r, map[string]any{
		"redirect": requestedRedirect,
		"scopes":   requestedScopes,
		"state":    Body.State,
		"application": map[string]any{
			"id":      application.ID,
			"created": application.Created,
			"name":    application.Name,
			"icon":    application.IconHash,
		},
		"user": map[string]any{
			"id":          profile.ID,
			"displayname": profile.Displayname,
			"avatar":      profile.AvatarHash,
		},
	})
}
