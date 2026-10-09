//go:build unix

package file

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFIFOSink_ServesOnDemandAndRepeatedly(t *testing.T) {
	ctx := WithMode(context.Background(), ModeNamedPipe)
	sink, err := NewSink(ctx, ".env")
	require.NoError(t, err)

	// Accumulate contents across several writes, as the netrc output does.
	_, err = sink.Write([]byte("machine a login x\n"))
	require.NoError(t, err)
	_, err = sink.Write([]byte("machine b login y\n"))
	require.NoError(t, err)

	// The path must be a named pipe, not a regular file.
	fi, err := os.Stat(sink.Path())
	require.NoError(t, err)
	assert.True(t, fi.Mode()&os.ModeNamedPipe != 0, "path should be a named pipe")

	// The contents are served on demand and can be read more than once.
	for i := 0; i < 3; i++ {
		b, err := os.ReadFile(sink.Path())
		require.NoError(t, err)
		assert.Equal(t, "machine a login x\nmachine b login y\n", string(b))
	}

	require.NoError(t, sink.Close())
	_, err = os.Stat(sink.Path())
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestFIFOSink_CloseWithoutReader(t *testing.T) {
	ctx := WithMode(context.Background(), ModeNamedPipe)
	sink, err := NewSink(ctx, ".env")
	require.NoError(t, err)

	_, err = sink.Write([]byte("TOKEN=s3cr3t\n"))
	require.NoError(t, err)

	// Closing must not hang even though no reader ever opened the pipe (the
	// serving goroutine is parked inside OpenFile).
	require.NoError(t, sink.Close())

	_, err = os.Stat(sink.Path())
	assert.ErrorIs(t, err, os.ErrNotExist)
}
