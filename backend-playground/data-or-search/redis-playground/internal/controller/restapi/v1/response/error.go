package response

// Error -.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CODE
const (
	ErrInvalidRequestBody         = "INVALID_REQUEST_BODY"
	ErrValidation                 = "VALIDATION_ERROR"
	ErrInvalidID                  = "INVALID_ID"
	ErrIdIsRequired               = "ID_IS_REQUIRED"
	ErrUnauthorized               = "UNAUTHORIZED"
	ErrMissingAuthorizationHeader = "MISSING_AUTHORIZATION_HEADER"
	ErrInvalidAuthorizationHeader = "INVALID_AUTHORIZATION_HEADER"
	ErrInvalidOrExpiredToken      = "INVALID_OR_EXPIRED_TOKEN"
	ErrForbidden                  = "FORBIDDEN"
	ErrNotFound                   = "NOT_FOUND"
	ErrInternal                   = "INTERNAL_ERROR"
	ErrInvalidValidation          = "INVALID_VALIDATION"

	ErrUserAlreadyExists  = "USER_ALREADY_EXIST"
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
)
