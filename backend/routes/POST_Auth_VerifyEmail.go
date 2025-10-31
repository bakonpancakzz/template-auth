package routes

import (
	"net/http"

	"github.com/bakonpancakz/template-auth/tools"
)

func POST_Auth_VerifyEmail(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		Token string `query:"token" validate:"required,token"`
	}
	if !tools.ValidateQuery(w, r, &Body) {
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Update Account matching Given Token
	tag, err := tools.Database.Exec(ctx,
		`UPDATE auth.users SET 
			updated 		 = CURRENT_TIMESTAMP,
			email_verified   = TRUE, 
			token_verify 	 = NULL,
			token_verify_eat = NULL
		WHERE token_verify = $1 AND token_verify_eat > NOW()`,
		Body.Token,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if tag.RowsAffected() == 0 {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_TOKEN)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
