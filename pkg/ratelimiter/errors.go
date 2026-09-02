package ratelimiter

import "errors"

var (
	ErrLimitExceeded = errors.New("limit exceeded")
)
