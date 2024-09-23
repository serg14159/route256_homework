package internal_errors

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrPreconditionFailed  = errors.New("precondition failed")
	ErrInternalServerError = errors.New("internal server error")

	ErrInvalidOrderStatus = errors.New("invalid order status")
)
