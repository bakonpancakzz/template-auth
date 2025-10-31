package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/bakonpancakzz/template-auth/tools"
)

func Test_POST_Auth_VerifyEmail(t *testing.T) {
	ResetDatabase()

	if _, err := tools.Database.Exec(context.Background(),
		"UPDATE auth.users SET token_verify = $1, token_verify_eat = $2 WHERE id = $3;",
		testSignedString,
		time.Now().Add(tools.LIFETIME_TOKEN_EMAIL_VERIFY),
		testUserID,
	); err != nil {
		t.Fatal(err)
	}

	t.Run("POST: Verify Email", func(t *testing.T) {
		res := DoJSON("POST", fmt.Sprintf("/auth/verify-email?token=%s", testSignedString), nil, nil)
		if res.StatusCode != http.StatusNoContent {
			t.Errorf("unexpected status code %d", res.StatusCode)
		}

		var emailVerified bool
		if err := tools.Database.QueryRow(context.Background(),
			"SELECT email_verified FROM auth.users WHERE id = $1;",
			testUserID,
		).Scan(&emailVerified); err != nil {
			t.Fatal(err)
		}
		if !emailVerified {
			t.Errorf("email was not verified")
		}
	})

}
