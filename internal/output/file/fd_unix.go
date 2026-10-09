//go:build unix

package file

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/jcchavezs/chasky/internal/log"
	"go.uber.org/zap"
)

// fdSink materializes the contents through an anonymous pipe whose read end is
// inherited by the spawned process and exposed as /dev/fd/N. Since an anonymous
// pipe has no name on the filesystem, it can only be reached by the process
// that inherits the descriptor (and its children), not by other processes of
// the same user. The contents are delivered once; a second read observes EOF.
type fdSink struct {
	w    *os.File
	path string

	release <-chan struct{}
	closed  chan struct{}
	once    sync.Once

	mu  sync.Mutex
	buf []byte
}

func newFDSink(ctx context.Context, _ string) (Sink, error) {
	set := fdSetFromContext(ctx)
	if set == nil {
		return nil, errors.New("fd delivery requires a file descriptor set in the context")
	}

	r, w, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("creating pipe: %w", err)
	}

	// The read end is handed to the spawned process through exec.Cmd.ExtraFiles;
	// add reports the descriptor number it will have there.
	fd := set.add(r)

	s := &fdSink{
		w:       w,
		path:    fmt.Sprintf("/dev/fd/%d", fd),
		release: set.done(),
		closed:  make(chan struct{}),
	}

	log.Logger.Debug("Creating inherited pipe", zap.String("path", s.path))
	go s.pump()

	return s, nil
}

func (s *fdSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buf = append(s.buf, p...)
	return len(p), nil
}

func (s *fdSink) Path() string {
	return s.path
}

func (s *fdSink) snapshot() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]byte, len(s.buf))
	copy(out, s.buf)
	return out
}

// pump waits until the contents are final (release) and writes them to the
// pipe, then closes the write end so the reader observes EOF. If the sink is
// closed before being released (e.g. an error aborted the render), it just
// releases the write end.
func (s *fdSink) pump() {
	select {
	case <-s.release:
		if _, err := s.w.Write(s.snapshot()); err != nil {
			log.Logger.Warn("Failed to write contents to inherited pipe", zap.Error(err))
		}
	case <-s.closed:
	}

	_ = s.w.Close()
}

func (s *fdSink) Close() error {
	s.once.Do(func() {
		close(s.closed)
	})
	return nil
}
