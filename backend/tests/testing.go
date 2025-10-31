package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/bakonpancakzz/template-auth/core"
	"github.com/bakonpancakzz/template-auth/include"
	"github.com/bakonpancakzz/template-auth/tools"
)

var (
	server           *httptest.Server
	client           *http.Client
	testUserID       int64 = 1
	testSignedString       = tools.GenerateSignedString()
	testEmailAddress       = "anonymous@example.org"
	testUsername           = "anonymous"
	testPassword           = "Password123!"
	testPasswordHash       = "..."
)

func init() {
	// Configure Environment
	tools.EMAIL_PROVIDER = "none"
	tools.STORAGE_PROVIDER = "none"
	tools.RATELIMIT_PROVIDER = "local"
	tools.LOGGER_PROVIDER = "console"
	testPasswordHash, _ = tools.GeneratePasswordHash(testPassword)

	// Startup Services
	stopCtx := context.TODO()
	stopWg := sync.WaitGroup{}
	tools.SetupLogger(stopCtx, &stopWg)
	tools.SetupRatelimitProvider(stopCtx, &stopWg)
	tools.SetupStorageProvider(stopCtx, &stopWg)
	tools.SetupEmailProvider(stopCtx, &stopWg)
	// tools.SetupGeolocation(stopCtx, &stopWg)
	tools.SetupDatabase(stopCtx, &stopWg)

	// Startup HTTP
	server = httptest.NewServer(core.SetupMux())
	client = server.Client()
}

// Reapply the SQL Schema onto the Testing Database, if written correctly then this
// should effectively reset the database, assuming it is also in development mode.
func ResetDatabase() {
	for _, a := range []struct {
		Q string
		A []any
	}{
		{Q: include.DatabaseSchema,
			A: []any{}},
		{Q: "INSERT INTO auth.users (id, email_address, password_hash) VALUES ($1, $2, $3);",
			A: []any{testUserID, testEmailAddress, testPasswordHash}},
		{Q: "INSERT INTO auth.profiles (id, username, displayname) VALUES ($1, $2, $2);",
			A: []any{testUserID, testUsername}},
	} {
		if _, err := tools.Database.Exec(context.Background(), a.Q, a.A...); err != nil {
			panic(err)
		}
	}
}

func DoJSON(method, path string, headers map[string]string, body any) *http.Response {
	// Create Request
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	url := fmt.Sprintf("%s%s", server.URL, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	// Submit Request
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	return res
}
