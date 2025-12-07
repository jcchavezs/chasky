package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandRunner(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		vars        map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name:    "command with template variables",
			command: "echo '{{.name}}'",
			vars:    map[string]string{"name": "test"},
			wantErr: false,
		},
		{
			name:        "invalid template syntax",
			command:     "echo '{{.name'",
			vars:        map[string]string{"name": "test"},
			wantErr:     true,
			errContains: "parsing pre-command template",
		},
		{
			name:        "command execution failure",
			command:     "exit 1",
			vars:        map[string]string{},
			wantErr:     true,
			errContains: "exit status 1",
		},
		{
			name:    "empty command",
			command: "",
			vars:    map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := commandRunner(ctx, tt.command, tt.vars)

			if tt.wantErr {
				require.Error(t, err, "commandRunner() expected error but got none")
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err, "commandRunner() unexpected error")
			}
		})
	}
}

func TestCommandRunnerContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := commandRunner(ctx, "sleep 10", map[string]string{})
	require.Error(t, err, "commandRunner() expected error due to cancelled context")
}
