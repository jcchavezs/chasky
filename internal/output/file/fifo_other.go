//go:build !unix

package file

import "errors"

// newFIFOSink is not supported on non-unix systems since named pipes (FIFOs)
// are a UNIX concept.
func newFIFOSink(_ string) (Sink, error) {
	return nil, errors.New("--fifo is only supported on unix systems")
}
