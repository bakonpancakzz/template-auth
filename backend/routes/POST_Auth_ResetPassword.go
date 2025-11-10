package routes

import (
	"errors"
	"net/http"
	"time"

	"github.com/bakonpancakz/template-auth/tools"

	"github.com/jackc/pgx/v5"
)

func POST_Auth_ResetPassword(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		Email string `json:"email" validate:"required,email"`
	}
	if !tools.ValidateJSON(w, r, &Body) {
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Update Account matching Given Email
	var (
		resetTokenExpires = time.Now().Add(tools.LIFETIME_TOKEN_EMAIL_RESET)
		resetToken        = tools.GenerateSignedString()
		user              tools.DatabaseUser
	)
	err := tools.Database.QueryRow(ctx,
		`UPDATE auth.users SET
			updated 		= CURRENT_TIMESTAMP,
			token_reset_eat = $1,
			token_reset 	= $2
		WHERE email_address = LOWER($3)
		RETURNING id, email_address`,
		resetTokenExpires,
		resetToken,
		Body.Email,
	).Scan(&user.ID, &user.EmailAddress)
	if errors.Is(err, pgx.ErrNoRows) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Notify Account Owner
	go func() {
		subCtx, subCancel := tools.NewContext()
		defer subCancel()

		// Fetch Displayname
		displayname := tools.EMAIL_DEFAULT_DISPLAYNAME
		tools.Database.
			QueryRow(subCtx, "SELECT displayname FROM auth.profiles WHERE user_id = $1", user.ID).
			Scan(&displayname)

		// Send Email
		tools.TemplateLoginForgotPassword(
			user.EmailAddress,
			tools.LocalsLoginForgotPassword{
				Displayname: displayname,
				Token:       resetToken,
			},
		)
	}()

	w.WriteHeader(http.StatusNoContent)
}
