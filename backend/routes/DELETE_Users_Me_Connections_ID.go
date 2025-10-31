package routes

import (
	"net/http"
	"strconv"

	"github.com/bakonpancakz/template-auth/tools"
)

func DELETE_Users_Me_Connections_ID(w http.ResponseWriter, r *http.Request) {

	session := tools.GetSession(r)
	if session.ApplicationID != tools.SESSION_NO_APPLICATION_ID {
		tools.SendClientError(w, r, tools.ERROR_OAUTH2_USERS_ONLY)
		return
	}
	if !session.Elevated {
		tools.SendClientError(w, r, tools.ERROR_MFA_ESCALATION_REQUIRED)
		return
	}

	snowflake, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_CONNECTION)
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Revoke Relevant Connection
	tag, err := tools.Database.Exec(ctx,
		`UPDATE auth.connections SET 
			updated = CURRENT_TIMESTAMP, 
			revoked = true 
		WHERE id = $1 AND user_id = $2`,
		snowflake,
		session.UserID,
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
