//go:build !unix

package file

import (
	"context"
	"errors"
)

// newFDSink is not supported on non-unix systems since it relies on descriptor
// inheritance and the /dev/fd interface.
func newFDSink(_ context.Context, _ string) (Sink, error) {
	return nil, errors.New("--fifo-mode=fd is only supported on unix systems")
}
