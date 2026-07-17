package entity

import "errors"

var (
	ErrUserNotFound               = errors.New("user not found")
	ErrUserAlreadyExists          = errors.New("user already exists")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrMissingAuthorizationHeader = errors.New("missing authorization header")
	ErrInvalidAuthorizationHeader = errors.New("invalid authorization header")
	ErrInvalidOrExpiredToken      = errors.New("invalid or expired token")
	ErrUnauthorized               = errors.New("unauthorized")
	ErrTaskNotFound               = errors.New("task not found")
	ErrTaskForbidden              = errors.New("task does not belong to user")
	ErrInvalidTransition          = errors.New("invalid status transition")

	ErrProductNotFound      = errors.New("product not found")
	ErrInvalidIdProduct     = errors.New("invalid product id")
	ErrInvalidProductCreate = errors.New("invalid product create")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrInvalidProductUpdate = errors.New("invalid product update")
	ErrInvalidProductDelete = errors.New("invalid product delete")
	ErrInvalidProductPatch  = errors.New("invalid product patch")

	ErrInternalServerError  = errors.New("internal server error")
	ErrorInvalidRequestBody = errors.New("invalid request body")
	ErrInvalidID            = errors.New("invalid id")
	ErrIdIsRequired         = errors.New("id is required")
	ErrInvalidValidation    = errors.New("invalid validation")
)
