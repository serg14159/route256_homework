package internal_errors

import "errors"

var (
	// HTTP
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrPreconditionFailed  = errors.New("precondition failed")
	ErrInternalServerError = errors.New("internal server error")

	ErrInvalidOrderStatus = errors.New("invalid order status")
	ErrStockReservation   = errors.New("stock reservation failed")

	ErrShardIndexOutOfRange = errors.New("shard index is out of range")
	ErrNoShardsAvailable    = errors.New("no shards available")
)
