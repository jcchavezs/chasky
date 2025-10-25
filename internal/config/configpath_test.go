package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigPath_ReturnsConfigFile(t *testing.T) {
	orig := userHomeDir
	defer func() { userHomeDir = orig }()

	tmp := t.TempDir()
	userHomeDir = func() (string, error) { return tmp, nil }

	p, err := ConfigPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmp, ".chasky.yaml"), p)
}

func TestConfigPath_HomeDirError(t *testing.T) {
	orig := userHomeDir
	defer func() { userHomeDir = orig }()

	userHomeDir = func() (string, error) { return "", os.ErrInvalid }

	p, err := ConfigPath()
	require.Error(t, err)
	require.Equal(t, "", p)
}
