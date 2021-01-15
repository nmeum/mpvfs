package playlistfs

import (
	"errors"
)

var (
	ErrNoCtl = errors.New("not a playctl command")
	ErrNoVol = errors.New("not a vol command")

	ErrTooFewArgs  = errors.New("too few arguments supplied")
	ErrTooManyArgs = errors.New("too many arguments supplied")

	ErrInvalidVol = errors.New("invalid volume level")
)
