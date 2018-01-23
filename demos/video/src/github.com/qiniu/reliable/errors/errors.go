package errors

import (
	"errors"
	"syscall"
)

// --------------------------------------------------------------------

var ErrBadData = errors.New("bad data")
var ErrLineTooLong = errors.New("line too long")
var ErrCrcChecksumError = errors.New("crc checksum error")
var ErrTooManyFails = errors.New("too many fails")
var ErrInvalidArgs = syscall.EINVAL

// --------------------------------------------------------------------
