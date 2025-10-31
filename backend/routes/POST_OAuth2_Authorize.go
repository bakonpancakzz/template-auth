package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bakonpancakz/template-auth/tools"

	"github.com/jackc/pgx/v5"
)

func POST_OAuth2_Authorize(w http.ResponseWriter, r *http.Request) {

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

	// Fetch State for Requested Application
	var application tools.DatabaseApplication
	err := tools.Database.QueryRow(ctx,
		"SELECT id, auth_redirects FROM auth.applications WHERE id = $1",
		Body.ClientID,
	).Scan(
		&application.ID,
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

	// Generate Temporary Grant Session
	grantCode := tools.GenerateSignedString()
	if _, err := tools.Database.Exec(ctx,
		`INSERT INTO auth.grants (
			id, expires, user_id, application_id, redirect_uri, scopes, code
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tools.GenerateSnowflake(),
		time.Now().Add(tools.LIFETIME_OAUTH2_GRANT_TOKEN),
		session.UserID,
		application.ID,
		requestedRedirect,
		requestedScopes,
		grantCode,
	); err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Redirect User to Requested URI with Grant
	q := url.Values{}
	q.Add("code", grantCode)
	if Body.State != nil {
		q.Add("state", *Body.State)
	}
	http.Redirect(w, r, fmt.Sprint(requestedRedirect, "?", q.Encode()), http.StatusFound)
}
