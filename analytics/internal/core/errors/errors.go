package core_errors

import "errors"

var (
	ErrParseRecord = errors.New("could not parse entity")
)
