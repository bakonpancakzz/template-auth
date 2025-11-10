package routes

import (
	"errors"
	"net/http"

	"github.com/bakonpancakz/template-auth/tools"

	"github.com/jackc/pgx/v5"
)

// This endpoint wont revoke user sessions to prevent any further PITA

func PATCH_Auth_ResetPassword(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		NewPassword string `json:"password" validate:"required,password"`
		Token       string `json:"token" validate:"required,token"`
	}
	if !tools.ValidateJSON(w, r, &Body) {
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Fetch Relevant Account
	var user tools.DatabaseUser
	err := tools.Database.QueryRow(ctx,
		`SELECT
			id, email_address, password_history
		FROM auth.users
		WHERE token_reset = $1 AND token_reset_eat > NOW()`,
		Body.Token,
	).Scan(
		&user.ID,
		&user.EmailAddress,
		&user.PasswordHistory,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Append Unique Password to History
	for _, oldPassword := range user.PasswordHistory {
		if ok, err := tools.ComparePasswordHash(oldPassword, Body.NewPassword); err != nil {
			tools.SendServerError(w, r, err)
			return
		} else if !ok {
			tools.SendClientError(w, r, tools.ERROR_LOGIN_PASSWORD_ALREADY_USED)
			return
		}
	}
	latestPasswordHash, err := tools.GeneratePasswordHash(Body.NewPassword)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	user.PasswordHistory = append(user.PasswordHistory, latestPasswordHash)
	if len(user.PasswordHistory) > tools.PASSWORD_HISTORY_LIMIT {
		user.PasswordHistory = user.PasswordHistory[1:]
	}

	// Update Account
	tag, err := tools.Database.Exec(ctx,
		`UPDATE auth.users SET
			updated 		 = CURRENT_TIMESTAMP,
			token_reset 	 = NULL,
			token_reset_eat	 = NULL,
			password_hash 	 = $1,
			password_history = $2
		WHERE id = $3`,
		latestPasswordHash,
		user.PasswordHistory,
		user.ID,
	)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if tag.RowsAffected() == 0 {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}

	// Alert Account Owner
	go func() {
		subCtx, subCancel := tools.NewContext()
		defer subCancel()

		// Fetch Displayname
		displayname := tools.EMAIL_DEFAULT_DISPLAYNAME
		tools.Database.
			QueryRow(subCtx, "SELECT displayname FROM auth.profiles WHERE user_id = $1", user.ID).
			Scan(&displayname)

		tools.TemplateNotifyUserPasswordModified(
			user.EmailAddress,
			tools.LocalsNotifyUserPasswordModified{
				Displayname: displayname,
			},
		)
	}()

	w.WriteHeader(http.StatusNoContent)
}
