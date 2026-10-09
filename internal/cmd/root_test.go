package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveCommand(t *testing.T) {
	const shell = "/bin/zsh"

	testCases := map[string]struct {
		args            []string
		dashPos         int
		wantEnvName     string
		wantCommand     string
		wantCommandArg  []string
		wantIsCustomCmd bool
		wantErr         bool
	}{
		"environ only, no dash": {
			args:        []string{"my_app"},
			dashPos:     -1,
			wantEnvName: "my_app",
			wantCommand: shell,
		},
		"bare command after dash": {
			args:            []string{"my_app", "claude"},
			dashPos:         1,
			wantEnvName:     "my_app",
			wantCommand:     "claude",
			wantIsCustomCmd: true,
		},
		"command with args after dash": {
			args:            []string{"my_app", "echo", "hello"},
			dashPos:         1,
			wantEnvName:     "my_app",
			wantCommand:     "echo",
			wantCommandArg:  []string{"hello"},
			wantIsCustomCmd: true,
		},
		"dash with nothing after it": {
			args:        []string{"my_app"},
			dashPos:     1,
			wantEnvName: "my_app",
			wantCommand: shell,
		},
		"extra args without dash": {
			args:    []string{"my_app", "claude"},
			dashPos: -1,
			wantErr: true,
		},
		"no environ before dash": {
			args:    []string{"claude"},
			dashPos: 0,
			wantErr: true,
		},
		"multiple args before dash": {
			args:    []string{"my_app", "extra", "claude"},
			dashPos: 2,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			envName, command, commandArg, isCustomCmd, err := resolveCommand(tc.args, tc.dashPos, shell)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantEnvName, envName)
			require.Equal(t, tc.wantCommand, command)
			require.ElementsMatch(t, tc.wantCommandArg, commandArg)
			require.Equal(t, tc.wantIsCustomCmd, isCustomCmd)
		})
	}
}
