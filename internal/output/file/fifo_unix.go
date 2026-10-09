//go:build unix

package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

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
// each iteration parks until a reader process opens the file.
//
// Once a reader attaches, the pipe is replaced with a fresh FIFO (same path,
// new inode) before the contents are written. The reader stays attached to the
// old, now-unlinked inode through its open descriptor, so the next blocking
// open parks on the new inode until a genuinely new reader arrives. Without
// this swap the loop would reopen the same inode while the reader it just
// served was still attached and deliver a second copy into that same read.
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

		// Replace the pipe before serving so the next iteration waits for a new
		// reader instead of re-serving this one. Also catches Close waking us up
		// rather than a genuine reader.
		if !s.recreate() {
			_ = wf.Close()
			return
		}

		if _, err := wf.Write(s.snapshot()); err != nil {
			log.Logger.Warn("Failed to write contents to named pipe", zap.Error(err))
		}
		_ = wf.Close()
	}
}

// recreate unlinks the current FIFO and creates a fresh one at the same path,
// giving the next blocking open a new inode to park on. It reports false if the
// sink is closing or the pipe can no longer be created, signalling serve to
// unwind. The caller must already hold an open writer to the previous inode so
// the reader attached there is unaffected by the swap.
func (s *fifoSink) recreate() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return false
	default:
	}

	_ = os.Remove(s.path)
	if err := syscall.Mkfifo(s.path, 0o600); err != nil {
		log.Logger.Warn("Failed to recreate named pipe", zap.Error(err))
		return false
	}
	return true
}

func (s *fifoSink) Close() error {
	s.once.Do(func() {
		close(s.done)
	})

	// A write-only open of a FIFO blocks until a reader shows up, so the serving
	// goroutine may be parked inside OpenFile. Open the read end once
	// (non-blocking) to release it; it then observes `done` and returns instead
	// of serving the contents again. Hold the lock so this does not race with
	// recreate swapping the inode out from under us.
	s.mu.Lock()
	if rf, err := os.OpenFile(s.path, os.O_RDONLY|syscall.O_NONBLOCK, os.ModeNamedPipe); err == nil {
		_ = rf.Close()
	}
	s.mu.Unlock()

	log.Logger.Debug("Removing named pipe", zap.String("path", s.path))
	return os.RemoveAll(s.dir)
}
