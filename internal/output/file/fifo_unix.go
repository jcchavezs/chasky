//go:build unix

package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jcchavezs/chasky/internal/log"
	"go.uber.org/zap"
)

// fifoSink materializes the contents as a UNIX named pipe (FIFO). The contents
// are written to the pipe on demand, only when a reader process opens it, so
// nothing is ever written to persistent storage. The contents are re-served on
// every read, which allows the pipe to be read more than once.
type fifoSink struct {
	dir  string
	path string
	done chan struct{}
	once sync.Once

	mu  sync.Mutex
	buf []byte
}

func newFIFOSink(pattern string) (Sink, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "chasky-fifo-")
	if err != nil {
		return nil, fmt.Errorf("creating pipe directory: %w", err)
	}

	// os.CreateTemp replaces a `*` in the pattern with a random string; the
	// pipe already lives in a unique directory, so just drop the placeholder to
	// keep a predictable basename.
	path := filepath.Join(dir, strings.Replace(pattern, "*", "", 1))

	log.Logger.Debug("Creating named pipe", zap.String("path", path))
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("creating named pipe: %w", err)
	}

	s := &fifoSink{
		dir:  dir,
		path: path,
		done: make(chan struct{}),
	}

	go s.serve()

	return s, nil
}

func (s *fifoSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buf = append(s.buf, p...)
	return len(p), nil
}

func (s *fifoSink) Path() string {
	return s.path
}

func (s *fifoSink) snapshot() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]byte, len(s.buf))
	copy(out, s.buf)
	return out
}

// serve repeatedly offers the current contents to whoever opens the pipe for
// reading. Opening the pipe for writing blocks until a reader is present, so
// each iteration parks until the reader process reads the file, writes the
// contents and starts over to serve the next read.
func (s *fifoSink) serve() {
	for {
		select {
		case <-s.done:
			return
		default:
		}

		wf, err := os.OpenFile(s.path, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			log.Logger.Debug("Stopping named pipe server", zap.Error(err))
			return
		}

		select {
		case <-s.done:
			// Woken up by Close rather than by a genuine reader.
			_ = wf.Close()
			return
		default:
		}

		if _, err := wf.Write(s.snapshot()); err != nil {
			log.Logger.Warn("Failed to write contents to named pipe", zap.Error(err))
		}
		_ = wf.Close()

		// Wait for the reader to drain and close the pipe before offering the
		// contents again. A write-only open succeeds while any reader still
		// holds the pipe open, so without this the loop would reopen and write
		// a second copy into the same reader session.
		s.waitForReaderToClose()
	}
}

// waitForReaderToClose blocks until no reader holds the pipe open. A
// non-blocking write-only open returns ENXIO once the last reader has closed
// its end; until then it succeeds, so poll until it fails. The probe never
// writes, so it cannot add contents to the current reader session.
func (s *fifoSink) waitForReaderToClose() {
	for {
		select {
		case <-s.done:
			return
		default:
		}

		f, err := os.OpenFile(s.path, os.O_WRONLY|syscall.O_NONBLOCK, os.ModeNamedPipe)
		if err != nil {
			if errors.Is(err, syscall.ENXIO) {
				// No reader is attached anymore; ready to serve the next one.
				return
			}
			// The pipe is gone (e.g. removed by Close) or otherwise
			// unavailable; stop waiting and let serve() unwind.
			return
		}
		_ = f.Close()

		select {
		case <-s.done:
			return
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func (s *fifoSink) Close() error {
	s.once.Do(func() {
		close(s.done)
	})

	// A write-only open of a FIFO blocks until a reader shows up, so the serving
	// goroutine may be parked inside OpenFile. Open the read end once
	// (non-blocking) to release it; it then observes `done` and returns instead
	// of serving the contents again.
	if rf, err := os.OpenFile(s.path, os.O_RDONLY|syscall.O_NONBLOCK, os.ModeNamedPipe); err == nil {
		_ = rf.Close()
	}

	log.Logger.Debug("Removing named pipe", zap.String("path", s.path))
	return os.RemoveAll(s.dir)
}
