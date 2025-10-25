package netrc

import (
	"os"
	"strings"
	"testing"

	"github.com/jcchavezs/chasky/internal/output/types"
	"github.com/stretchr/testify/require"
)

func TestNetRC_String(t *testing.T) {
	tests := []struct {
		name     string
		netRC    netRC
		expected string
	}{
		{
			name: "complete netrc with machine, login, and password",
			netRC: netRC{
				Machine:  "api.github.com",
				Login:    "myusername",
				Password: "mypassword",
			},
			expected: "machine api.github.com login myusername password mypassword",
		},
		{
			name: "netrc without password",
			netRC: netRC{
				Machine: "api.github.com",
				Login:   "myusername",
			},
			expected: "machine api.github.com login myusername",
		},
		{
			name: "default machine with login and password",
			netRC: netRC{
				Machine:  "default",
				Login:    "defaultuser",
				Password: "defaultpass",
			},
			expected: "default login defaultuser password defaultpass",
		},
		{
			name: "empty machine with login and password",
			netRC: netRC{
				Login:    "defaultuser",
				Password: "defaultpass",
			},
			expected: "default login defaultuser password defaultpass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.netRC.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestNetRC_StringConsistency(t *testing.T) {
	// Test that the String method produces consistent output
	netrc := netRC{
		Machine:  "api.example.com",
		Login:    "testuser",
		Password: "testpass",
	}

	result1 := netrc.String()
	result2 := netrc.String()

	require.Equal(t, result1, result2, "String() method should produce consistent output")
}

func TestNetRC_StringDefaultVsRegularMachine(t *testing.T) {
	// Test the difference between default and regular machine formatting
	defaultNetrc := netRC{
		Machine:  "default",
		Login:    "user",
		Password: "pass",
	}

	regularNetrc := netRC{
		Machine:  "example.com",
		Login:    "user",
		Password: "pass",
	}

	defaultResult := defaultNetrc.String()
	regularResult := regularNetrc.String()

	require.Contains(t, defaultResult, "default login user")
	require.NotContains(t, defaultResult, "machine default")

	require.Contains(t, regularResult, "machine example.com login user")
	require.NotContains(t, regularResult, "default")
}

func TestNetRC_StringPasswordOptional(t *testing.T) {
	// Test that password is optional in the output
	withPassword := netRC{
		Machine:  "api.example.com",
		Login:    "user",
		Password: "secret",
	}

	withoutPassword := netRC{
		Machine: "api.example.com",
		Login:   "user",
	}

	withPasswordResult := withPassword.String()
	withoutPasswordResult := withoutPassword.String()

	require.Contains(t, withPasswordResult, "password secret")
	require.NotContains(t, withoutPasswordResult, "password")

	// Both should contain the machine and login parts
	require.Contains(t, withPasswordResult, "machine api.example.com login user")
	require.Contains(t, withoutPasswordResult, "machine api.example.com login user")
}

func TestExec(t *testing.T) {
	tests := []struct {
		name      string
		values    map[string]string
		wantErr   bool
		errMsg    string
		checkFunc func(t *testing.T, output types.Output)
	}{
		{
			name: "complete netrc with machine, login, and password",
			values: map[string]string{
				"machine":  "api.github.com",
				"login":    "myuser",
				"password": "mypass",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output types.Output) {
				require.Len(t, output.EnvVars, 1)
				require.Contains(t, output.EnvVars[0], "NETRC_FILE=")
				require.NotEmpty(t, output.WelcomeMsg)
				require.NotNil(t, output.Closer)

				// Check file contents
				netrcFile := strings.TrimPrefix(output.EnvVars[0], "NETRC_FILE=")
				content, err := os.ReadFile(netrcFile)
				require.NoError(t, err)
				require.Equal(t, "machine api.github.com login myuser password mypass\n", string(content))
			},
		},
		{
			name:    "empty values map",
			values:  map[string]string{},
			wantErr: true,
			errMsg:  "empty values",
		},
		{
			name: "missing login",
			values: map[string]string{
				"machine":  "api.github.com",
				"password": "mypass",
			},
			wantErr: true,
			errMsg:  "login is required",
		},
		{
			name: "machine with spaces",
			values: map[string]string{
				"machine": "api github.com",
				"login":   "myuser",
			},
			wantErr: true,
			errMsg:  "machine value should not contain spaces",
		},
		{
			name: "unknown field",
			values: map[string]string{
				"machine": "api.github.com",
				"login":   "myuser",
				"unknown": "value",
			},
			wantErr: true,
			errMsg:  "unknown field unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			output, err := Exec(ctx, tt.values)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}

			// Clean up the created file if closer exists
			if output.Closer != nil {
				defer func() {
					err := output.Closer()
					require.NoError(t, err)
				}()
			}
		})
	}
}

func TestExec_FileCleanup(t *testing.T) {
	values := map[string]string{
		"machine": "api.example.com",
		"login":   "testuser",
	}

	ctx := t.Context()
	output, err := Exec(ctx, values)
	require.NoError(t, err)
	require.NotNil(t, output.Closer)

	// Extract file path
	netrcFile := strings.TrimPrefix(output.EnvVars[0], "NETRC_FILE=")

	// Verify file exists
	_, err = os.Stat(netrcFile)
	require.NoError(t, err)

	// Call closer
	err = output.Closer()
	require.NoError(t, err)

	// Verify file is removed
	_, err = os.Stat(netrcFile)
	require.True(t, os.IsNotExist(err))
}

func TestExec_WelcomeMessage(t *testing.T) {
	values := map[string]string{
		"machine": "api.example.com",
		"login":   "testuser",
	}

	ctx := t.Context()
	output, err := Exec(ctx, values)
	require.NoError(t, err)

	defer output.Closer() //nolint

	require.Contains(t, output.WelcomeMsg, "NETRC_FILE")
	require.Contains(t, output.WelcomeMsg, "curl --netrc-file")
	require.Contains(t, output.WelcomeMsg, "$NETRC_FILE")
}
