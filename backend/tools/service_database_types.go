package tools

import "time"

type DatabaseUser struct {
	ID                int64
	Created           time.Time
	Updated           time.Time
	Permissions       int
	EmailAddress      string
	EmailVerified     bool
	IPAddress         string
	MFAEnabled        bool
	MFASecret         *string
	MFACodes          []string
	MFACodesUsed      int
	PasswordHash      *string
	PasswordHistory   []string
	TokenVerify       *string
	TokenVerifyEAT    *time.Time
	TokenLogin        *string
	TokenLoginData    *string
	TokenLoginExpires *time.Time
	TokenReset        *string
	TokenResetEAT     *time.Time
	TokenPasscode     *string
	TokenPasscodeEAT  *time.Time
}

type DatabaseProfile struct {
	ID               int64
	Created          time.Time
	Updated          time.Time
	Username         string
	Displayname      string
	Subtitle         *string
	Biography        *string
	AvatarHash       *string
	BannerHash       *string
	AccentBanner     *int
	AccentBorder     *int
	AccentBackground *int
}

type DatabaseSession struct {
	ID              int64
	Created         time.Time
	Updated         time.Time
	UserID          int64
	Revoked         bool
	Token           string
	ElevatedUntil   int
	DeviceIPAddress string
	DeviceUserAgent string
}

type DatabaseApplication struct {
	ID            int64
	Created       time.Time
	Updated       time.Time
	UserID        int64
	Name          string
	Description   *string
	IconHash      *string
	AuthSecret    string
	AuthRedirects []string
}

type DatabaseConnection struct {
	ID            int64
	Created       time.Time
	Updated       time.Time
	UserID        int64
	ApplicationID int64
	Revoked       bool
	Scopes        int
	TokenAccess   *string
	TokenExpires  time.Time
	TokenRefresh  *string
}

type DatabaseGrant struct {
	ID            int64
	Expires       time.Time
	UserID        int64
	ApplicationID int64
	RedirectURI   string
	Scopes        int
	Code          string
}
