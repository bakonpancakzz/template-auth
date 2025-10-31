package routes

import (
	"net/http"
	"strconv"

	"github.com/bakonpancakz/template-auth/tools"

	"github.com/jackc/pgx/v5"
)

func POST_OAuth2_Token_Revoke(w http.ResponseWriter, r *http.Request) {

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
		Token string `query:"token" validate:"required,token"`
	}
	if !tools.ValidateQuery(w, r, &Body) {
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Validate Application Secret
	var application tools.DatabaseApplication
	err := tools.Database.QueryRow(ctx,
		"SELECT id, secret_key FROM auth.applications WHERE id = $1",
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

	// Mark Relevant Connection as Revoked
	tag, err := tools.Database.Exec(ctx,
		`UPDATE auth.connections SET 
			updated = CURRENT_TIMESTAMP,
			revoked = true
		WHERE (token_access = $1 OR token_refresh = $1) 
		AND application_id = $2
		AND revoked = false`,
		Body.Token,
		application.ID,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if tag.RowsAffected() == 0 {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_CONNECTION)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
