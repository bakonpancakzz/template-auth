package routes

import (
	"net/http"
	"strconv"

	"github.com/bakonpancakz/template-auth/tools"
)

func DELETE_Users_Me_Applications_ID_Reset(w http.ResponseWriter, r *http.Request) {

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
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_APPLICATION)
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Generate New Secret Key for Application
	newSecret := tools.GenerateSignedString()
	tag, err := tools.Database.Exec(ctx,
		`UPDATE auth.applications SET
			updated 	= CURRENT_TIMESTAMP,
			secret_key 	= $1
		WHERE id = $2 AND user_id = $3`,
		newSecret,
		snowflake,
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
	tools.SendJSON(w, r, http.StatusOK, map[string]any{
		"secret_key": newSecret,
	})
}
