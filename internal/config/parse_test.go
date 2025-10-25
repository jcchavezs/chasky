package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseInvalidYAML(t *testing.T) {
	// Create a temporary invalid YAML file
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.yaml")

	invalidContent := `
invalid yaml content:
  - this is not valid: [
    yaml structure
missing closing bracket
`
	err := os.WriteFile(invalidFile, []byte(invalidContent), 0644)
	require.NoError(t, err)

	config, err := parse(invalidFile)
	require.Error(t, err)
	require.Nil(t, config)
	require.Contains(t, err.Error(), "unmarshaling config")
}

func TestParseEmptyFile(t *testing.T) {
	// Create a temporary empty file
	tempDir := t.TempDir()
	emptyFile := filepath.Join(tempDir, "empty.yaml")

	err := os.WriteFile(emptyFile, []byte(""), 0644)
	require.NoError(t, err)

	config, err := parse(emptyFile)
	require.NoError(t, err)
	require.Nil(t, config)
	require.Equal(t, 0, len(config))
}

func TestParseMissingSecretType(t *testing.T) {
	// Create a temporary file with missing secret type
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "missing_type.yaml")

	content := `
test_env:
  - output: env
    values:
      INVALID_SECRET:
        bash:
          command: echo "test"
`
	err := os.WriteFile(invalidFile, []byte(content), 0644)
	require.NoError(t, err)

	config, err := parse(invalidFile)
	require.Error(t, err)
	require.Nil(t, config)
	require.Contains(t, err.Error(), "unmarshaling config")
}

func TestParseValidSecretTypes(t *testing.T) {
	// Create a temporary file with different valid secret types
	tempDir := t.TempDir()
	validFile := filepath.Join(tempDir, "valid_secrets.yaml")

	content := `
test_env:
  - output: env
    values:
      BASH_SECRET:
        type: bash
        bash:
          command: echo "test"
      STATIC_SECRET:
        type: static
        static:
          value: "test_value"
`
	err := os.WriteFile(validFile, []byte(content), 0644)
	require.NoError(t, err)

	config, err := parse(validFile)
	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, 1, len(config))

	envValues, exists := config["test_env"]
	require.True(t, exists)
	require.Equal(t, 1, len(envValues))

	values := envValues[0].Values
	require.Equal(t, 2, len(values))

	// Check bash secret
	bashSecret, exists := values["BASH_SECRET"]
	require.True(t, exists)
	require.Equal(t, "bash", bashSecret.Type)
	require.JSONEq(t, `{"command":"echo \"test\""}`, string(bashSecret.RawConfig))

	// Check static secret
	staticSecret, exists := values["STATIC_SECRET"]
	require.True(t, exists)
	require.Equal(t, "static", staticSecret.Type)
	require.JSONEq(t, `{"value":"test_value"}`, string(staticSecret.RawConfig))
}

func TestParseMultipleOutputsPerEnv(t *testing.T) {
	// Create a temporary file with multiple outputs for the same env
	tempDir := t.TempDir()
	multiOutputFile := filepath.Join(tempDir, "multi_output.yaml")

	content := `
test_env:
  - output: env
    values:
      ENV_SECRET:
        type: static
        static:
          value: "env_value"
  - output: gcloud
    values:
      GCLOUD_SECRET:
        type: bash
        bash:
          command: echo "gcloud_value"
`
	err := os.WriteFile(multiOutputFile, []byte(content), 0644)
	require.NoError(t, err)

	config, err := parse(multiOutputFile)
	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, 1, len(config))

	envValues, exists := config["test_env"]
	require.True(t, exists)
	require.Equal(t, 2, len(envValues))

	// Check first output (env)
	envOutput := envValues[0]
	require.Equal(t, "env", envOutput.OutputType)
	require.Equal(t, 1, len(envOutput.Values))
	envSecret := envOutput.Values["ENV_SECRET"]
	require.Equal(t, "static", envSecret.Type)

	// Check second output (gcloud)
	gcloudOutput := envValues[1]
	require.Equal(t, "gcloud", gcloudOutput.OutputType)
	require.Equal(t, 1, len(gcloudOutput.Values))
	gcloudSecret := gcloudOutput.Values["GCLOUD_SECRET"]
	require.Equal(t, "bash", gcloudSecret.Type)
}
