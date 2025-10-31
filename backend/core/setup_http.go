package core

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/bakonpancakzz/template-auth/routes"
	"github.com/bakonpancakzz/template-auth/tools"
)

var httpLogger tools.LoggerInstance

func SetupHTTP(stop context.Context, await *sync.WaitGroup) {
	httpLogger = tools.Logger.New("http")

	// Optimized to prevent malicious attacks but shouldn't
	// really bother devices on slower networks :)

	svr := http.Server{
		Handler:           SetupMux(),
		Addr:              tools.HTTP_ADDRESS,
		MaxHeaderBytes:    4096,
		IdleTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       10 * time.Second,
	}
	if tools.HTTP_TLS_ENABLED {
		tls, err := tools.NewTLSConfig(
			tools.HTTP_TLS_CERT,
			tools.HTTP_TLS_KEY,
			tools.HTTP_TLS_CA,
		)
		if err != nil {
			httpLogger.Fatal("Failed to configure TLS", err)
			return
		}
		svr.TLSConfig = tls
	}

	// Shutdown Logic
	await.Add(1)
	go func() {
		defer await.Done()
		<-stop.Done()
		svr.Shutdown(context.Background())
		httpLogger.Info("Closed", nil)
	}()

	// Server Startup
	var err error
	httpLogger.Info("Ready", map[string]any{"address": svr.Addr})
	if tools.HTTP_TLS_ENABLED {
		err = svr.ListenAndServeTLS("", "")
	} else {
		err = svr.ListenAndServe()
	}
	if err != http.ErrServerClosed {
		httpLogger.Fatal("Startup Failed", err)
	}
}

func SetupMux() *http.ServeMux {
	var (
		mux       = http.NewServeMux()
		session   = tools.UseSession
		limitFile = tools.NewBodyLimit(10 * 1024 * 1024) // 10MB
		limitJSON = tools.NewBodyLimit(10 * 1024)        // 10KB
		rateLogin = tools.NewRatelimit(&tools.RatelimitOptions{
			Bucket: "RATE_LOGIN",
			Period: time.Minute,
			Limit:  5,
		})
		rateClientRead = tools.NewRatelimit(&tools.RatelimitOptions{
			Bucket: "RATE_CLIENT_READ",
			Period: time.Minute,
			Limit:  100,
		})
		rateClientWrite = tools.NewRatelimit(&tools.RatelimitOptions{
			Bucket: "RATE_CLIENT_WRITE",
			Period: time.Minute,
			Limit:  10,
		})
		rateClientImage = tools.NewRatelimit(&tools.RatelimitOptions{
			Bucket: "RATE_CLIENT_IMAGE",
			Period: 5 * time.Minute,
			Limit:  3,
		})
		// rateServerWrite = tools.NewRatelimit(&tools.RatelimitOptions{
		// 	Bucket: "RATE_SERVER_WRITE",
		// 	Period: time.Minute,
		// 	Limit:  1000,
		// })
	)

	// // Login Routes
	mux.Handle("/auth/login", tools.MethodHandler{
		// http.MethodPost: tools.Chain(routes.POST_Auth_Login, rateLogin, limitJSON),
	})
	mux.Handle("/auth/signup", tools.MethodHandler{
		// http.MethodPost: tools.Chain(routes.POST_Auth_Signup, rateLogin, limitJSON),
	})
	mux.Handle("/auth/logout", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_Logout, rateLogin, session),
	})
	mux.Handle("/auth/password-reset", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_ResetPassword, rateLogin, limitJSON),
		// http.MethodPatch: tools.Chain(routes.PATCH_Auth_ResetPassword, rateLogin, limitJSON),
	})
	mux.Handle("/auth/verify-login", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_VerifyLogin, rateLogin),
	})
	mux.Handle("/auth/verify-email", tools.MethodHandler{
		http.MethodPost: tools.Chain(routes.POST_Auth_VerifyEmail, rateLogin),
	})

	// oAuth2
	mux.Handle("/oauth2/authorize", tools.MethodHandler{
		// http.MethodGet:  tools.Chain(routes.GET_OAuth2_Authorize, rateClientRead, session),
		// http.MethodPost: tools.Chain(routes.POST_OAuth2_Authorize, rateClientWrite, session),
	})
	mux.Handle("/oauth2/token", tools.MethodHandler{
		// http.MethodPost: tools.Chain(routes.POST_OAuth2_Token, rateServerWrite),
	})
	mux.Handle("/oauth2/token/revoke", tools.MethodHandler{
		// http.MethodPost: tools.Chain(routes.POST_OAuth2_Token_Revoke, rateServerWrite),
	})

	// User
	mux.Handle("/users/{id}", tools.MethodHandler{
		http.MethodGet: tools.Chain(routes.GET_Users_ID, rateClientRead),
	})
	mux.Handle("/users/@me", tools.MethodHandler{
		http.MethodGet: tools.Chain(routes.GET_Users_Me, rateClientRead, session),
		// http.MethodPatch:  tools.Chain(routes.PATCH_Users_Me, rateClientWrite, session, limitJSON),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me, rateClientWrite, session),
	})
	mux.Handle("/users/@me/avatar", tools.MethodHandler{
		// http.MethodPut:    tools.Chain(nil, rateClientImage, session, limitFile), // TODO: Implement endpoint
		// http.MethodDelete: tools.Chain(nil, rateClientWrite, session),            // TODO: Implement endpoint
	})
	mux.Handle("/users/@me/banner", tools.MethodHandler{
		// http.MethodPut:    tools.Chain(nil, rateClientImage, session, limitFile), // TODO: Implement endpoint
		// http.MethodDelete: tools.Chain(nil, rateClientWrite, session),            // TODO: Implement endpoint
	})

	// User Applications
	mux.Handle("/users/@me/applications", tools.MethodHandler{
		http.MethodGet: tools.Chain(routes.GET_Users_Me_Applications, rateClientRead, session),
		// http.MethodPost: tools.Chain(routes.POST_Users_Me_Applications, rateClientWrite, session),
	})
	mux.Handle("/users/@me/applications/{id}", tools.MethodHandler{
		// http.MethodPatch:  tools.Chain(routes.PATCH_Users_Me_Applications_ID, rateClientWrite, session, limitJSON),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Applications_ID, rateClientWrite, session),
	})
	mux.Handle("/users/@me/applications/{id}/icon", tools.MethodHandler{
		http.MethodPut:    tools.Chain(nil, rateClientImage, session, limitFile), // TODO: Implement endpoint
		http.MethodDelete: tools.Chain(nil, rateClientWrite, session),            // TODO: Implement endpoint
	})
	mux.Handle("/users/@me/applications/{id}/reset", tools.MethodHandler{
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Applications_ID_Reset, rateClientWrite, session),
	})

	// User Application Connections
	mux.Handle("/users/@me/connections", tools.MethodHandler{
		// http.MethodGet: tools.Chain(routes.GET_Users_Me_Connections, rateClientRead, session),
	})
	mux.Handle("/users/@me/connections/{id}", tools.MethodHandler{
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Connections_ID, rateClientWrite, session),
	})

	// User Security
	mux.Handle("/users/@me/security/sessions", tools.MethodHandler{
		http.MethodGet: tools.Chain(routes.GET_Users_Me_Security_Sessions, rateClientRead, session),
	})
	mux.Handle("/users/@me/security/sessions/{id}", tools.MethodHandler{
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Security_Sessions_ID, rateClientWrite, session),
	})
	mux.Handle("/users/@me/security/escalate", tools.MethodHandler{
		// http.MethodPost: tools.Chain(routes.POST_Users_Me_Security_Escalate, rateClientWrite, session, limitJSON),
	})
	mux.Handle("/users/@me/security/mfa/setup", tools.MethodHandler{
		http.MethodGet:    tools.Chain(routes.GET_Users_Me_Security_MFA_Setup, rateClientWrite, session),
		http.MethodPost:   tools.Chain(routes.POST_Users_Me_Security_MFA_Setup, rateClientWrite, session, limitJSON),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Security_MFA_Setup, rateClientWrite, session),
	})
	mux.Handle("/users/@me/security/mfa/codes", tools.MethodHandler{
		http.MethodGet:    tools.Chain(routes.GET_Users_Me_Security_MFA_Codes, rateClientRead, session),
		http.MethodDelete: tools.Chain(routes.DELETE_Users_Me_Security_MFA_Codes, rateClientWrite, session),
	})
	mux.Handle("/users/@me/security/password", tools.MethodHandler{
		// http.MethodPatch: tools.Chain(routes.PATCH_Users_Me_Security_Password, rateClientWrite, session, limitJSON),
	})
	mux.Handle("/users/@me/security/email", tools.MethodHandler{
		// http.MethodPost:  tools.Chain(routes.POST_Users_Me_Security_Email, rateClientWrite, session),
		// http.MethodPatch: tools.Chain(routes.PATCH_Users_Me_Security_Email, rateClientWrite, session, limitJSON),
	})

	// Default 404 Handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tools.SendClientError(w, r, tools.ERROR_GENERIC_NOT_FOUND)
	})

	return mux
}
