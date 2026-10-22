package file

import (
	"context"
	"os"
	"testing"

	"github.com/jcchavezs/chasky/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	log.Logger = zap.NewNop()
}

func TestModeFromContext(t *testing.T) {
	assert.Equal(t, ModeDisk, ModeFromContext(context.Background()))
	assert.Equal(t, ModeNamedPipe, ModeFromContext(WithMode(context.Background(), ModeNamedPipe)))
	assert.Equal(t, ModeFD, ModeFromContext(WithMode(context.Background(), ModeFD)))
}

func TestDiskSink(t *testing.T) {
	sink, err := NewSink(context.Background(), ".env")
	require.NoError(t, err)

	_, err = sink.Write([]byte("TOKEN=s3cr3t\n"))
	require.NoError(t, err)

	content, err := os.ReadFile(sink.Path())
	require.NoError(t, err)
	assert.Equal(t, "TOKEN=s3cr3t\n", string(content))

	require.NoError(t, sink.Close())
	_, err = os.Stat(sink.Path())
	assert.ErrorIs(t, err, os.ErrNotExist)
}
