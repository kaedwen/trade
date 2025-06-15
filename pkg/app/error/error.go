package api_error

import "errors"

var (
	ErrApiBadStatus = errors.New("bad api status code")
)
