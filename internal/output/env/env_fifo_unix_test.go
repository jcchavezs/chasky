//go:build unix

package env

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jcchavezs/chasky/internal/log"
	"github.com/jcchavezs/chasky/internal/output/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	log.Logger = zap.NewNop()
}

func TestExec_ProducesEnvFileWhenPipeEnabled(t *testing.T) {
	ctx := file.WithMode(context.Background(), file.ModeNamedPipe)

	out, err := Exec(ctx, map[string]string{"S3_BUCKET": "my-bucket", "SECRET_KEY": "super-secret"})
	require.NoError(t, err)
	require.NotNil(t, out.Closer)
	t.Cleanup(func() { _ = out.Closer() })

	// The secrets are exposed through a file, not injected into the environment.
	require.Len(t, out.EnvVars, 1)
	require.True(t, strings.HasPrefix(out.EnvVars[0], "ENV_FILE="))
	assert.Contains(t, out.WelcomeMsg, "$ENV_FILE")

	path := strings.TrimPrefix(out.EnvVars[0], "ENV_FILE=")
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "S3_BUCKET=my-bucket\n")
	assert.Contains(t, string(content), "SECRET_KEY=super-secret\n")
}
