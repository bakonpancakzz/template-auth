package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bakonpancakzz/template-auth/tools"

	"github.com/jackc/pgx/v5"
)

func PATCH_Users_Me_Applications_ID(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if session.ApplicationID != tools.SESSION_NO_APPLICATION_ID {
		tools.SendClientError(w, r, tools.ERROR_OAUTH2_USERS_ONLY)
		return
	}

	var Body struct {
		Name        *string   `json:"name" validate:"omitempty,displayname"`
		Description *string   `json:"description" validate:"omitempty,description"`
		Redirects   *[]string `json:"redirects"`
	}
	if !tools.ValidateJSON(w, r, &Body) {
		return
	}

	snowflake, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_APPLICATION)
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Fetch Relevant Application
	var application tools.DatabaseApplication
	err = tools.Database.QueryRow(ctx,
		`SELECT 
			id, created, name, description, icon_hash, redirects
		FROM auth.applications 
		WHERE id = $1 AND user_id = $2`,
		snowflake,
		session.UserID,
	).Scan(
		&application.ID,
		&application.Created,
		&application.Name,
		&application.Description,
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

	// Collect Application Edits
	edited := false
	if Body.Name != nil {
		application.Name = *Body.Name
		edited = true
	}
	if Body.Description != nil {
		if len(*Body.Description) == 0 {
			application.Description = nil
		} else {
			application.Description = Body.Description
		}
		edited = true
	}
	if Body.Redirects != nil {
		if len(*Body.Redirects) > tools.REDIRECT_URI_SLICE_LEN_MAX {
			tools.SendFormError(w, r, tools.ValidationError{
				Field:    "redirects",
				Error:    tools.VALIDATOR_SLICE_TOO_MANY_ITEMS,
				Literals: []any{tools.REDIRECT_URI_SLICE_LEN_MAX},
			})
			return
		}
		for i, uri := range *Body.Redirects {
			if _, err := url.Parse(uri); err != nil {
				tools.SendFormError(w, r, tools.ValidationError{
					Field: fmt.Sprintf("redirects[%d]", i),
					Error: tools.VALIDATOR_URI_INVALID,
				})
				return
			}
			if len(uri) > tools.REDIRECT_URI_STRING_LEN_MAX {
				tools.SendFormError(w, r, tools.ValidationError{
					Field:    fmt.Sprintf("redirects[%d]", i),
					Error:    tools.VALIDATOR_STRING_TOO_LONG,
					Literals: []any{tools.REDIRECT_URI_STRING_LEN_MAX},
				})
				return
			}
		}
		application.AuthRedirects = *Body.Redirects
		edited = true
	}

	if !edited {
		tools.SendClientError(w, r, tools.ERROR_BODY_EMPTY)
		return
	}

	// Apply Application Edits
	tag, err := tools.Database.Exec(ctx,
		`UPDATE auth.applications SET
			updated 	= CURRENT_TIMESTAMP,
			name		= $1,
			description = $2,
			redirects   = $3
		WHERE id = $4 and user_id = $5`,
		application.Name,
		application.Description,
		application.AuthRedirects,
		application.ID,
		session.UserID,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if tag.RowsAffected() == 0 {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_APPLICATION)
		return
	}

	// Organize Application
	tools.SendJSON(w, r, map[string]any{
		"id":          application.ID,
		"created":     application.Created,
		"name":        application.Name,
		"description": application.Description,
		"icon":        application.IconHash,
		"redirects":   application.AuthRedirects,
	})
}
