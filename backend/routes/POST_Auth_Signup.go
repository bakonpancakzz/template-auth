package routes

import (
	"net/http"
	"time"

	"github.com/bakonpancakzz/template-auth/tools"
)

func POST_Auth_Signup(w http.ResponseWriter, r *http.Request) {

	var Body struct {
		Email    string `json:"email" validate:"required,email"`
		Username string `json:"username" validate:"required,username"`
		Password string `json:"password" validate:"required,password"`
	}
	if !tools.ValidateJSON(w, r, &Body) {
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Check for Duplicate Email or Username
	var usageEmail int
	var usageUsername int
	if err := tools.Database.QueryRow(ctx,
		`SELECT 
			(SELECT COUNT(*) FROM auth.profiles WHERE username = $1),
			(SELECT COUNT(*) FROM auth.users WHERE email_address = LOWER($2))`,
		Body.Username,
		Body.Email,
	).Scan(
		&usageUsername,
		&usageEmail,
	); err != nil {
		tools.SendServerError(w, r, err)
		return
	}
	if usageEmail > 0 {
		tools.SendClientError(w, r, tools.ERROR_SIGNUP_DUPLICATE_EMAIL)
		return
	}
	if usageUsername > 0 {
		tools.SendClientError(w, r, tools.ERROR_SIGNUP_DUPLICATE_USERNAME)
		return
	}

	// Hash Password
	userVerifyEmail := tools.GenerateSignedString()
	userPasswordHash, err := tools.GeneratePasswordHash(Body.Password)
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Create New Account and Profile
	if _, err := tools.Database.Exec(ctx,
		`BEGIN;
			INSERT INTO auth.users (
				id, email_address, token_verify, token_verify_eat,
				password_hash, password_history, ip_address
			) VALUES ($1, LOWER($2), $3, $4, $5, $6, $7);
			INSERT INTO auth.profiles (
				id, username, displayname
			) VALUES ($1, $8, $8);
		COMMIT;`,
		tools.GenerateSnowflake(),
		Body.Email,
		userVerifyEmail,
		time.Now().Add(tools.LIFETIME_TOKEN_EMAIL_VERIFY),
		userPasswordHash,
		[]string{userPasswordHash},
		tools.GetRemoteIP(r),
		Body.Username,
	); err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Send Verification Email
	go func() {
		tools.TemplateEmailVerify(
			Body.Email,
			tools.LocalsEmailVerify{
				Displayname: Body.Username,
				Token:       userVerifyEmail,
			},
		)
	}()

	w.WriteHeader(http.StatusNoContent)
}
