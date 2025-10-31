package core

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/bakonpancakzz/template-auth/include"
	"github.com/bakonpancakzz/template-auth/tools"
)

// Renders all Email Templates and then immediately exits
// Intended to help with design or template tweaking

func DebugEmailRenderTemplates() {

	var (
		exampleUsername = tools.EMAIL_DEFAULT_DISPLAYNAME
		exampleAddress  = "127.0.0.1"
		exampleToken    = tools.GenerateSignedString()
		exampleLocation = "Fresno, California, United States"
		exampleBrowser  = "Chrome on Windows 10.0"
		exampleTime     = "10/23/2025 07:45am"
		defaults        = map[string]any{
			"EMAIL_VERIFY.html": tools.LocalsEmailVerify{
				Displayname: exampleUsername,
				Token:       exampleToken,
			},
			"LOGIN_FORGOT_PASSWORD.html": tools.LocalsLoginForgotPassword{
				Displayname: exampleUsername,
				Token:       exampleToken,
			},
			"LOGIN_NEW_LOCATION.html": tools.LocalsLoginNewLocation{
				Displayname:    exampleUsername,
				Token:          exampleToken,
				IpAddress:      exampleAddress,
				Timestamp:      exampleTime,
				DeviceBrowser:  exampleBrowser,
				DeviceLocation: exampleLocation,
			},
			"LOGIN_NEW_DEVICE.html": tools.LocalsLoginNewDevice{
				Displayname:    exampleUsername,
				Timestamp:      exampleTime,
				IpAddress:      exampleAddress,
				DeviceBrowser:  exampleBrowser,
				DeviceLocation: exampleLocation,
			},
			"LOGIN_PASSCODE.html": tools.LocalsLoginPasscode{
				Displayname: exampleUsername,
				Code:        tools.GeneratePasscode(),
				Lifetime:    fmt.Sprint(tools.LIFETIME_TOKEN_EMAIL_PASSCODE.Minutes()),
			},
			"NOTIFY_USER_DELETED.html": tools.LocalsNotifyUserDeleted{
				Displayname: exampleUsername,
				Reason:      "User Request",
			},
			"NOTIFY_USER_EMAIL_MODIFIED.html": tools.LocalsNotifyUserEmailModified{
				Displayname: exampleUsername,
			},
			"NOTIFY_USER_PASS_MODIFIED.html": tools.LocalsNotifyUserPasswordModified{
				Displayname: exampleUsername,
			},
		}
	)

	// Render Templates
	entries, err := include.EmailTemplates.ReadDir("templates")
	if err != nil {
		fmt.Printf("Cannot read embedded directory: %s\n", err)
		return
	}
	for _, ent := range entries {

		// Sanity Checks
		filename := path.Base(ent.Name())
		if strings.HasPrefix(filename, "_") {
			fmt.Printf("Ignoring Template: %s\n", filename)
			continue
		}
		locals, ok := defaults[ent.Name()]
		if !ok {
			fmt.Printf("Ignoring Template, contribute some locals!: %s\n", filename)
			continue
		}

		// Process Template
		template, err := template.ParseFS(
			include.EmailTemplates,
			"templates/_TEMPLATE.html",
			"templates/"+filename,
		)
		if err != nil {
			fmt.Printf("Cannot parse template '%s': %s\n", filename, err)
			return
		}

		// Execute Template
		os.Mkdir("dist", 0666)
		f, err := os.Create("dist/" + filename)
		if err != nil {
			fmt.Printf("Create file error: %s\n", err)
			return
		}
		if err := template.Execute(f, map[string]any{
			"Host": tools.EMAIL_DEFAULT_HOST,
			"Data": locals,
		}); err != nil {
			fmt.Printf("Cannot Render Teamplate '%s': %s\n", filename, err)
			return
		} else {
			fmt.Printf("Rendered Template '%s'\n", filename)
		}

	}

	os.Exit(0)
}
