package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bakonpancakz/template-auth/tools"

	"github.com/jackc/pgx/v5"
)

const (
	GRANT_CODE    = "authorization_code"
	GRANT_REFRESH = "refresh_token"
)

func POST_OAuth2_Token(w http.ResponseWriter, r *http.Request) {

	var clientID int64
	var clientSecret string
	if user, pass, ok := r.BasicAuth(); !ok {
		tools.SendClientError(w, r, tools.ERROR_GENERIC_UNAUTHORIZED)
		return
	} else if id, err := strconv.ParseInt(user, 10, 64); err != nil {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_APPLICATION)
		return
	} else {
		clientID = id
		clientSecret = pass
	}

	var Body struct {
		GrantType    string `query:"grant_type" validate:"required"`
		RedirectURI  string `query:"redirect_uri"`
		Code         string `query:"code"`
		RefreshToken string `query:"refresh_token"`
	}
	if !tools.ValidateQuery(w, r, &Body) {
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Validate Parameters
	switch Body.GrantType {
	case GRANT_CODE:
		if Body.RedirectURI == "" {
			tools.SendClientError(w, r, tools.ERROR_OAUTH2_FORM_INVALID_REDIRECT_URI)
			return
		}
		if !tools.CompareSignedString(Body.Code) {
			tools.SendClientError(w, r, tools.ERROR_OAUTH2_FORM_INVALID_CODE)
			return
		}
	case GRANT_REFRESH:
		if !tools.CompareSignedString(Body.RefreshToken) {
			tools.SendClientError(w, r, tools.ERROR_OAUTH2_FORM_INVALID_REFRESH_TOKEN)
			return
		}
	}

	// Validate Application Secret
	var application tools.DatabaseApplication
	err := tools.Database.QueryRow(ctx,
		`SELECT
			id, auth_secret
		FROM auth.applications
		WHERE id = $1`,
		clientID,
	).Scan(
		&application.ID,
		&application.AuthSecret,
	)
	if err == pgx.ErrNoRows {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_APPLICATION)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if !tools.CompareStringConstant(clientSecret, application.AuthSecret) {
		tools.SendClientError(w, r, tools.ERROR_GENERIC_UNAUTHORIZED)
		return
	}

	// Complete Grant Request
	switch Body.GrantType {
	case GRANT_CODE:

		// Consume Auth Grant
		var grant tools.DatabaseGrant
		err := tools.Database.QueryRow(ctx,
			`DELETE FROM auth.grants
			WHERE code = $1 AND expires > NOW()
			RETURNING user_id, application_id, redirect_uri, scopes`,
			Body.Code,
		).Scan(
			&grant.UserID,
			&grant.ApplicationID,
			&grant.RedirectURI,
			&grant.Scopes,
		)
		if err == pgx.ErrNoRows {
			tools.SendClientError(w, r, tools.ERROR_UNKNOWN_APPLICATION)
			return
		}
		if err != nil {
			tools.SendServerError(w, r, err)
			return
		}
		if grant.ApplicationID != clientID ||
			grant.RedirectURI != Body.RedirectURI {
			tools.SendClientError(w, r, tools.ERROR_GENERIC_UNAUTHORIZED)
			return
		}

		// Fetch Relevant Connection
		var tokenAccess, tokenRefresh string
		var connection tools.DatabaseConnection
		switch tools.Database.QueryRow(ctx,
			`SELECT token_expires FROM auth.connections
			WHERE application_id = $1 AND user_id = $2`,
			grant.ApplicationID,
			grant.UserID,
		).Scan(&connection.TokenExpires) {

		// Create New Connection
		case pgx.ErrNoRows:
			tokenAccess = tools.GenerateSignedString()
			tokenRefresh = tools.GenerateSignedString()
			if _, err := tools.Database.Exec(ctx,
				`INSERT INTO auth.connections
				(id, user_id, application_id, scopes, token_access, token_expires, token_refresh)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				tools.GenerateSnowflake(),
				grant.UserID,
				grant.ApplicationID,
				grant.Scopes,
				tokenAccess,
				time.Now().Add(tools.LIFETIME_OAUTH2_ACCESS_TOKEN),
				tokenRefresh,
			); err != nil {
				tools.SendServerError(w, r, err)
				return
			}

		// Reset Existing Connection
		case nil:
			tag, err := tools.Database.Exec(ctx,
				`UPDATE auth.connections SET
						updated			= CURRENT_TIMESTAMP,
						revoked 		= FALSE,
						scopes  		= $1,
						token_access 	= $2,
						token_refresh   = $3,
						token_expires	= $4
					WHERE user_id = $5
					AND application_id = $6`,

				grant.Scopes,
				tokenAccess,
				tokenRefresh,
				time.Now().Add(tools.LIFETIME_OAUTH2_ACCESS_TOKEN),
				grant.UserID,
				grant.ApplicationID,
			)
			if err != nil {
				tools.SendServerError(w, r, err)
				return
			}
			if tag.RowsAffected() == 0 {
				tools.SendClientError(w, r, tools.ERROR_UNKNOWN_CONNECTION)
				return
			}

		// Database Error
		default:
			tools.SendServerError(w, r, err)
			return
		}

		// Organize Grant
		tools.SendJSON(w, r, map[string]any{
			"token_type":    tools.TOKEN_PREFIX_BEARER,
			"access_token":  tokenAccess,
			"refresh_token": tokenRefresh,
			"expires_in":    tools.LIFETIME_OAUTH2_ACCESS_TOKEN.Seconds(),
			"scopes":        tools.ToStringFromScopes(grant.Scopes),
		})
		return

	case GRANT_REFRESH:
		// Search for Relevant Connection
		var connection tools.DatabaseConnection
		err = tools.Database.QueryRow(ctx,
			`SELECT id, revoked, scopes
			FROM auth.connections
			WHERE token_refresh = $1
			AND application_id 	= $2`,
			Body.RefreshToken,
			application.ID,
		).Scan(
			&connection.ID,
			&connection.Scopes,
			&connection.Revoked,
		)
		if err == pgx.ErrNoRows {
			tools.SendClientError(w, r, tools.ERROR_GENERIC_UNAUTHORIZED)
			return
		}
		if err != nil {
			tools.SendServerError(w, r, err)
			return
		}
		if connection.Revoked {
			tools.SendClientError(w, r, tools.ERROR_ACCESS_REVOKED)
			return
		}

		// Update Connection Tokens
		var tokenAccess = tools.GenerateSignedString()
		var tokenRefresh = tools.GenerateSignedString()
		tag, err := tools.Database.Exec(ctx,
			`UPDATE auth.connections SET
				updated 	  = CURRENT_TIMESTAMP,
				token_access  = $1,
				token_refresh = $2,
				token_expires = $3
			WHERE id = $4`,
			tokenAccess,
			tokenRefresh,
			time.Now().Add(tools.LIFETIME_OAUTH2_ACCESS_TOKEN),
			connection.ID,
		)
		if err != nil {
			tools.SendServerError(w, r, err)
			return
		}
		if tag.RowsAffected() == 0 {
			tools.SendClientError(w, r, tools.ERROR_UNKNOWN_CONNECTION)
			return
		}

		// Organize Connection
		tools.SendJSON(w, r, map[string]any{
			"token_type":    tools.TOKEN_PREFIX_BEARER,
			"access_token":  tokenAccess,
			"refresh_token": tokenRefresh,
			"expires_in":    tools.LIFETIME_OAUTH2_ACCESS_TOKEN.Seconds(),
			"scopes":        tools.ToStringFromScopes(connection.Scopes),
		})
		return

	// Unknown Grant Type
	default:
		tools.SendClientError(w, r, tools.ERROR_OAUTH2_FORM_INVALID_GRANT_TYPE)
	}
}
