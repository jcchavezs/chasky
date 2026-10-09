package env

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExec_InjectsIntoEnvironmentByDefault(t *testing.T) {
	out, err := Exec(context.Background(), map[string]string{"TOKEN": "s3cr3t"})
	require.NoError(t, err)

	assert.Equal(t, []string{"TOKEN=s3cr3t"}, out.EnvVars)
	assert.Nil(t, out.Closer)
}

func TestExec_EmptyValues(t *testing.T) {
	out, err := Exec(context.Background(), map[string]string{})
	require.NoError(t, err)
	assert.Empty(t, out.EnvVars)
}
