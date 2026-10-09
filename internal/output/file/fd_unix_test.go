//go:build unix

package file

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFDSink_WritesOnReleaseAndNumbersDescriptors(t *testing.T) {
	set := NewFDSet()
	ctx := WithFDSet(WithMode(context.Background(), ModeFD), set)

	first, err := NewSink(ctx, ".env")
	require.NoError(t, err)
	second, err := NewSink(ctx, ".netrc")
	require.NoError(t, err)

	// The inherited descriptors are numbered from 3, matching ExtraFiles.
	assert.Equal(t, "/dev/fd/3", first.Path())
	assert.Equal(t, "/dev/fd/4", second.Path())

	// Contents can be accumulated across writes (as the netrc output does).
	_, err = first.Write([]byte("A=1\n"))
	require.NoError(t, err)
	_, err = first.Write([]byte("B=2\n"))
	require.NoError(t, err)

	files := set.Files()
	require.Len(t, files, 2)

	// Nothing is delivered until the contents are released.
	set.Release()

	// Read the read end as the spawned process would through /dev/fd/3.
	b, err := io.ReadAll(files[0])
	require.NoError(t, err)
	assert.Equal(t, "A=1\nB=2\n", string(b))

	require.NoError(t, first.Close())
	require.NoError(t, second.Close())
	set.CloseFiles()
}

func TestFDSink_CloseBeforeRelease(t *testing.T) {
	set := NewFDSet()
	ctx := WithFDSet(WithMode(context.Background(), ModeFD), set)

	sink, err := NewSink(ctx, ".env")
	require.NoError(t, err)

	_, err = sink.Write([]byte("TOKEN=s3cr3t\n"))
	require.NoError(t, err)

	// Closing before Release (e.g. an aborted render) must not leak the pump
	// goroutine nor hang.
	require.NoError(t, sink.Close())
	set.CloseFiles()
}

func TestFDSink_RequiresFDSet(t *testing.T) {
	ctx := WithMode(context.Background(), ModeFD)
	_, err := NewSink(ctx, ".env")
	assert.Error(t, err)
}
