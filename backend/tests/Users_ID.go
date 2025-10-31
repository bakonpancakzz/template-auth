package tests

import (
	"fmt"
	"net/http"
	"testing"
)

func Test_Users_ID(t *testing.T) {
	ResetDatabase()

	t.Run("GET: Get User", func(t *testing.T) {
		res := DoJSON("GET", fmt.Sprintf("/users/%d", testUserID), nil, nil)
		if res.StatusCode != http.StatusOK {
			t.Errorf("unexpected status code %d", res.StatusCode)
		}
	})

}
