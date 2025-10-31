package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/bakonpancakzz/template-auth/tools"
)

func Test_POST_Auth_ResetPassword(t *testing.T) {
	ResetDatabase()

	t.Run("POST: Request Password Reset", func(t *testing.T) {
		res := DoJSON("POST", "/auth/password-reset", nil, map[string]any{
			"email": testEmailAddress,
		})
		if res.StatusCode != http.StatusNoContent {
			t.Errorf("unexpected status code %d", res.StatusCode)
		}

		var userResetToken *string
		err := tools.Database.QueryRow(context.Background(),
			"SELECT token_reset FROM auth.users WHERE id = $1;",
			testUserID,
		).Scan(&userResetToken)
		if err != nil {
			t.Fatal(err)
		}
		if userResetToken == nil {
			t.Errorf("token was not set")
		}
	})
	// t.Run("PATCH: Update Password using Token", func(t *testing.T) {})
}
