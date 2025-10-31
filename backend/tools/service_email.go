package tools

import (
	"bytes"
	"context"
	"html/template"
	"sync"
	"time"

	"github.com/bakonpancakzz/template-auth/include"
)

type LocalsEmailVerify struct {
	Displayname string
	Token       string
}
type LocalsLoginForgotPassword struct {
	Displayname string
	Token       string
}
type LocalsLoginNewLocation struct {
	Displayname    string
	Token          string
	Timestamp      string
	IpAddress      string
	DeviceBrowser  string
	DeviceLocation string
}
type LocalsLoginNewDevice struct {
	Displayname    string
	Timestamp      string
	IpAddress      string
	DeviceBrowser  string
	DeviceLocation string
}
type LocalsLoginPasscode struct {
	Displayname string
	Code        string
	Lifetime    string
}
type LocalsNotifyUserDeleted struct {
	Displayname string
	Reason      string
}
type LocalsNotifyUserEmailModified struct {
	Displayname string
}
type LocalsNotifyUserPasswordModified struct {
	Displayname string
}

var (
	TemplateEmailVerify                = SetupEmailTemplate[LocalsEmailVerify]("EMAIL_VERIFY", "Verify your Email Address")
	TemplateLoginForgotPassword        = SetupEmailTemplate[LocalsLoginForgotPassword]("LOGIN_FORGOT_PASSWORD", "Forgot Your Password?")
	TemplateLoginNewLocation           = SetupEmailTemplate[LocalsLoginNewLocation]("LOGIN_NEW_LOCATION", "Allow Login from a New Location")
	TemplateLoginNewDevice             = SetupEmailTemplate[LocalsLoginNewDevice]("LOGIN_NEW_DEVICE", "Login from a New Device")
	TemplateLoginPasscode              = SetupEmailTemplate[LocalsLoginPasscode]("LOGIN_PASSCODE", "Your One Time Passcode")
	TemplateNotifyUserDeleted          = SetupEmailTemplate[LocalsNotifyUserDeleted]("NOTIFY_USER_DELETED", "Account Deleted")
	TemplateNotifyUserEmailModified    = SetupEmailTemplate[LocalsNotifyUserEmailModified]("NOTIFY_USER_EMAIL_MODIFIED", "Your Account Password has Changed")
	TemplateNotifyUserPasswordModified = SetupEmailTemplate[LocalsNotifyUserPasswordModified]("NOTIFY_USER_PASS_MODIFIED", "Your Account Email has Changed")
)

type EmailProvider interface {
	Start(stop context.Context, await *sync.WaitGroup) error
	Send(toAddress, subject, html string) error
}

var Email EmailProvider
var EmailLogger LoggerInstance

func SetupEmailProvider(stop context.Context, await *sync.WaitGroup) {
	EmailLogger = Logger.New("email")
	t := time.Now()

	switch EMAIL_PROVIDER {
	case "ses":
		Email = &emailProviderSES{}
	case "emailengine":
		Email = &emailProviderEmailEngine{}
	case "none":
		Email = &emailProviderNone{}
	default:
		EmailLogger.Fatal("Unknown Provider", EMAIL_PROVIDER)
	}

	if err := Email.Start(stop, await); err != nil {
		EmailLogger.Fatal("Startup Failed", err)
	}
	EmailLogger.Info("Ready", map[string]any{
		"time": time.Since(t).String(),
	})
}

func SetupEmailTemplate[L any](filename, subjectLine string) func(emailAddress string, locals L) {

	// Parse Template
	template, err := template.ParseFS(
		include.EmailTemplates,
		"templates/_TEMPLATE.html",
		"templates/"+filename+".html",
	)
	if err != nil {
		panic("cannot parse template: " + err.Error())
	}

	// Send Function
	return func(emailAddress string, locals L) {

		// Render Email
		var buffer bytes.Buffer
		err := template.Execute(&buffer, map[string]any{
			"Host": EMAIL_DEFAULT_HOST,
			"Data": locals,
		})
		if err != nil {
			EmailLogger.Error("Render Failed", map[string]any{
				"address":  emailAddress,
				"template": filename,
				"locals":   locals,
				"error":    err,
			})
			return
		}

		// Send Email
		EmailLogger.Info("Attempt", map[string]any{
			"address":  emailAddress,
			"template": filename,
			"error":    Email.Send(emailAddress, subjectLine, buffer.String()),
		})
	}
}
