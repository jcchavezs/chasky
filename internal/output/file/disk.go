package file

import (
	"fmt"
	"os"
)

// diskSink materializes the contents as a temporary file on disk.
type diskSink struct {
	f *os.File
}

func newDiskSink(pattern string) (Sink, error) {
	f, err := os.CreateTemp(os.TempDir(), pattern)
	if err != nil {
		return nil, fmt.Errorf("creating temporary file: %w", err)
	}

	return &diskSink{f: f}, nil
}

func (d *diskSink) Write(p []byte) (int, error) {
	return d.f.Write(p)
}

func (d *diskSink) Path() string {
	return d.f.Name()
}

func (d *diskSink) Close() error {
	_ = d.f.Close()
	return os.Remove(d.f.Name())
}
