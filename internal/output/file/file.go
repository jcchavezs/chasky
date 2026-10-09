// Package file provides a Sink abstraction to materialize output contents as a
// file-like path for the reader process. Depending on the selected delivery
// mode, a Sink is backed by:
//
//   - a temporary file on disk (the default);
//   - a UNIX named pipe (FIFO) that serves the contents on demand so secrets
//     never touch persistent storage, but is reachable by any process of the
//     same user;
//   - an anonymous pipe inherited by the spawned process and exposed as
//     /dev/fd/N, which is reachable only by that process (and its children).
package file

import (
	"context"
	"io"
	"os"
	"sync"
)

// Mode selects how a Sink delivers its contents to the reader process.
type Mode uint8

const (
	// ModeDisk writes the contents to a temporary file on disk. This is the
	// default when FIFO delivery is not requested.
	ModeDisk Mode = iota
	// ModeNamedPipe serves the contents on demand through a UNIX named pipe
	// (FIFO) placed on the filesystem.
	ModeNamedPipe
	// ModeFD serves the contents through an anonymous pipe inherited by the
	// spawned process and exposed as /dev/fd/N, scoping access to that process.
	ModeFD
)

type modeCtxKey struct{}

// WithMode returns a copy of ctx carrying the delivery mode NewSink should use.
func WithMode(ctx context.Context, m Mode) context.Context {
	return context.WithValue(ctx, modeCtxKey{}, m)
}

// ModeFromContext returns the delivery mode stored on ctx, defaulting to
// ModeDisk when none was set.
func ModeFromContext(ctx context.Context) Mode {
	m, _ := ctx.Value(modeCtxKey{}).(Mode)
	return m
}

// Enabled reports whether on-demand pipe delivery (any --fifo mode) was
// requested on ctx, as opposed to the default on-disk file delivery.
func Enabled(ctx context.Context) bool {
	return ModeFromContext(ctx) != ModeDisk
}

// Sink is a file-like destination for output contents. Callers write the
// contents (possibly across several calls), expose Path to the reader process
// and invoke Close once the environment is torn down.
type Sink interface {
	io.Writer
	// Path is the file path to hand over to the reader process.
	Path() string
	// Close releases the underlying resources, removing the file or the pipe.
	Close() error
}

// NewSink creates a Sink for the given file name pattern using the delivery
// mode stored on ctx.
func NewSink(ctx context.Context, pattern string) (Sink, error) {
	switch ModeFromContext(ctx) {
	case ModeFD:
		return newFDSink(ctx, pattern)
	case ModeNamedPipe:
		return newFIFOSink(pattern)
	default:
		return newDiskSink(pattern)
	}
}

// FDSet collects the read ends of the anonymous pipes used by ModeFD sinks so
// the caller can pass them as inherited file descriptors to the spawned
// process, and coordinates when their contents are flushed.
//
// It is safe to create and wire an FDSet unconditionally: when no ModeFD sink
// is used it stays empty and its methods are no-ops.
type FDSet struct {
	mu        sync.Mutex
	files     []*os.File
	releaseCh chan struct{}
	released  sync.Once
	closed    sync.Once
}

// NewFDSet creates an empty FDSet.
func NewFDSet() *FDSet {
	return &FDSet{releaseCh: make(chan struct{})}
}

// add registers a read end to be inherited by the spawned process and returns
// the file descriptor number it will have in that process. The first entry maps
// to fd 3 to match the ordering of exec.Cmd.ExtraFiles.
func (s *FDSet) add(f *os.File) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files = append(s.files, f)
	return 2 + len(s.files)
}

// Files returns the read ends to assign to exec.Cmd.ExtraFiles, in the order
// they were registered.
func (s *FDSet) Files() []*os.File {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.files
}

func (s *FDSet) done() <-chan struct{} {
	return s.releaseCh
}

// Release signals the ModeFD sinks that the contents are final and may be
// written to their pipes. It must be called after the spawned process has
// started so a large payload can drain into it.
func (s *FDSet) Release() {
	s.released.Do(func() {
		close(s.releaseCh)
	})
}

// CloseFiles closes the registered read ends. The caller should call it once
// the spawned process has exited; the process keeps its own inherited copies.
func (s *FDSet) CloseFiles() {
	s.closed.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, f := range s.files {
			_ = f.Close()
		}
	})
}

type fdSetCtxKey struct{}

// WithFDSet returns a copy of ctx carrying the FDSet that ModeFD sinks register
// their inherited read ends with.
func WithFDSet(ctx context.Context, s *FDSet) context.Context {
	return context.WithValue(ctx, fdSetCtxKey{}, s)
}

func fdSetFromContext(ctx context.Context) *FDSet {
	s, _ := ctx.Value(fdSetCtxKey{}).(*FDSet)
	return s
}
