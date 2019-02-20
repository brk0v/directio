// +build !linux

package directio

import (
	"errors"
)

// ErrUnsupportedDirectIO is not supported
var ErrUnsupportedDirectIO = errors.New("No DirectIO support")

// stub
func checkDirectIO(fd uintptr) error {
	return ErrUnsupportedDirectIO
}

// stub
func setDirectIO(fd uintptr, dio bool) error {
	return ErrUnsupportedDirectIO
}
