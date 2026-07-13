package entity

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskForbidden      = errors.New("task does not belong to user")
	ErrInvalidTransition  = errors.New("invalid status transition")

	ErrProductNotFound      = errors.New("product not found")
	ErrInvalidIdProduct     = errors.New("invalid product id")
	ErrInvalidProductCreate = errors.New("invalid product create")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrInvalidProductUpdate = errors.New("invalid product update")
	ErrInvalidProductDelete = errors.New("invalid product delete")
	ErrInvalidProductPatch  = errors.New("invalid product patch")

	ErrInternalServerError  = errors.New("internal server error")
	ErrorInvalidRequestBody = errors.New("invalid request body")
)
