package tools

type APIError struct {
	Status  int
	Code    int
	Message string
}

var (
	ERROR_GENERIC_SERVER                    = APIError{Status: 500, Code: 0, Message: "Server Error"}
	ERROR_GENERIC_NOT_FOUND                 = APIError{Status: 404, Code: 0, Message: "Endpoint Not Found"}
	ERROR_GENERIC_RATELIMIT                 = APIError{Status: 429, Code: 0, Message: "Too Many Requests"}
	ERROR_GENERIC_UNAUTHORIZED              = APIError{Status: 401, Code: 0, Message: "Unauthorized"}
	ERROR_GENERIC_METHOD_NOT_ALLOWED        = APIError{Status: 405, Code: 0, Message: "Method Not Allowed"}
	ERROR_BODY_EMPTY                        = APIError{Status: 400, Code: 0, Message: "Empty Form Body"}
	ERROR_BODY_MALFORMED                    = APIError{Status: 422, Code: 0, Message: "Malformed Form Body"}
	ERROR_BODY_INVALID                      = APIError{Status: 400, Code: 0, Message: "Invalid Form Body"}
	ERROR_BODY_TOO_LARGE                    = APIError{Status: 413, Code: 0, Message: "Payload Too Large"}
	ERROR_UNKNOWN_USER                      = APIError{Status: 404, Code: 1020, Message: "Unknown User"}
	ERROR_UNKNOWN_TOKEN                     = APIError{Status: 404, Code: 1030, Message: "Unknown Token"}
	ERROR_UNKNOWN_SESSION                   = APIError{Status: 404, Code: 1040, Message: "Unknown Session"}
	ERROR_UNKNOWN_APPLICATION               = APIError{Status: 404, Code: 1050, Message: "Unknown Application"}
	ERROR_UNKNOWN_CONNECTION                = APIError{Status: 404, Code: 1060, Message: "Unknown Connection"}
	ERROR_UNKNOWN_IMAGE                     = APIError{Status: 404, Code: 1070, Message: "Unknown Image"}
	ERROR_IMAGE_UNSUPPORTED                 = APIError{Status: 400, Code: 2010, Message: "Unsupported Image Format (Supports: WEBP, GIF, JPEG, PNG)"}
	ERROR_IMAGE_MALFORMED                   = APIError{Status: 400, Code: 2020, Message: "Invalid or Malformed Image Data"}
	ERROR_ACCESS_REVOKED                    = APIError{Status: 401, Code: 3010, Message: "Access Revoked"}
	ERROR_ACCESS_EXPIRED                    = APIError{Status: 401, Code: 3020, Message: "Access Expired"}
	ERROR_LOGIN_INCORRECT                   = APIError{Status: 401, Code: 4010, Message: "Incorrect Email or Password"}
	ERROR_LOGIN_ACCOUNT_DELETED             = APIError{Status: 401, Code: 4020, Message: "Account Deleted"}
	ERROR_LOGIN_PASSWORD_RESET              = APIError{Status: 401, Code: 4030, Message: "Account Locked. Please reset your password using 'Forgot Password?' on the login page"}
	ERROR_LOGIN_PASSWORD_ALREADY_USED       = APIError{Status: 400, Code: 4040, Message: "Password Already Used"}
	ERROR_SIGNUP_DUPLICATE_USERNAME         = APIError{Status: 409, Code: 4050, Message: "Username is already in use"}
	ERROR_SIGNUP_DUPLICATE_EMAIL            = APIError{Status: 409, Code: 4060, Message: "Email Address is already in use"}
	ERROR_MFA_EMAIL_SENT                    = APIError{Status: 403, Code: 5010, Message: "Email Sent"}
	ERROR_MFA_EMAIL_ALREADY_VERIFIED        = APIError{Status: 400, Code: 5020, Message: "Email Address already Verified"}
	ERROR_MFA_PASSCODE_REQUIRED             = APIError{Status: 403, Code: 5030, Message: "Authenticator Passcode Required"}
	ERROR_MFA_PASSCODE_INCORRECT            = APIError{Status: 401, Code: 5040, Message: "Authenticator Passcode Incorrect"}
	ERROR_MFA_RECOVERY_CODE_USED            = APIError{Status: 401, Code: 5050, Message: "Recovery Code Used"}
	ERROR_MFA_RECOVERY_CODE_INCORRECT       = APIError{Status: 403, Code: 5060, Message: "Recovery Code Incorrect"}
	ERROR_MFA_ESCALATION_REQUIRED           = APIError{Status: 403, Code: 5070, Message: "Escalation Required"}
	ERROR_MFA_PASSWORD_INCORRECT            = APIError{Status: 401, Code: 5080, Message: "Incorrect Password"}
	ERROR_MFA_DISABLED                      = APIError{Status: 412, Code: 5090, Message: "MFA is Disabled"}
	ERROR_MFA_SETUP_ALREADY                 = APIError{Status: 400, Code: 5100, Message: "MFA is Already Setup"}
	ERROR_MFA_SETUP_NOT_INITIALIZED         = APIError{Status: 412, Code: 5110, Message: "MFA Setup not Started"}
	ERROR_OAUTH2_SCOPE_REQUIRED             = APIError{Status: 403, Code: 6010, Message: "Endpoint requires an Additional Scope"}
	ERROR_OAUTH2_USERS_ONLY                 = APIError{Status: 403, Code: 6020, Message: "Endpoint restricted to Users Only"}
	ERROR_OAUTH2_FORM_INVALID_REDIRECT_URI  = APIError{Status: 400, Code: 6030, Message: "Invalid 'redirect_uri'"}
	ERROR_OAUTH2_FORM_INVALID_RESPONSE_TYPE = APIError{Status: 400, Code: 6040, Message: "Invalid 'response_type'"}
	ERROR_OAUTH2_FORM_INVALID_GRANT_TYPE    = APIError{Status: 400, Code: 6050, Message: "Invalid 'grant_type'"}
	ERROR_OAUTH2_FORM_INVALID_CODE          = APIError{Status: 400, Code: 6060, Message: "Invalid 'code'"}
	ERROR_OAUTH2_FORM_INVALID_ACCESS_TOKEN  = APIError{Status: 400, Code: 6070, Message: "Invalid 'access_token'"}
	ERROR_OAUTH2_FORM_INVALID_REFRESH_TOKEN = APIError{Status: 400, Code: 6080, Message: "Invalid 'refresh_token'"}
	ERROR_OAUTH2_FORM_INVALID_SCOPE         = APIError{Status: 400, Code: 6090, Message: "Invalid 'scope'"}
)
