package keyring

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestResolve(t *testing.T) {
	currentUser, err := getCurrentUser()
	require.NoError(t, err)

	tests := []struct {
		name      string
		rawConfig []byte
		setupFunc func(t *testing.T)
		want      string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "successful keyring retrieval",
			rawConfig: []byte(`key: "test-key"`),
			setupFunc: func(t *testing.T) {
				err := keyring.Set("test-key", currentUser, "secret-value")
				require.NoError(t, err)
			},
			want:    "secret-value",
			wantErr: false,
		},
		{
			name:      "key not found in keyring",
			rawConfig: []byte(`key: "non-existent-key"`),
			setupFunc: func(t *testing.T) {
				// No setup needed - key doesn't exist
			},
			want:    "",
			wantErr: false,
		},
		{
			name:      "missing key in config",
			rawConfig: []byte(`{}`),
			setupFunc: func(t *testing.T) {
				// No setup needed
			},
			want:    "",
			wantErr: true,
			errMsg:  "missing keyring.key value",
		},
		{
			name:      "empty key in config",
			rawConfig: []byte(`key: ""`),
			setupFunc: func(t *testing.T) {
				// No setup needed
			},
			want:    "",
			wantErr: true,
			errMsg:  "missing keyring.key value",
		},
		{
			name:      "invalid yaml config",
			rawConfig: []byte(`invalid yaml: [}`),
			setupFunc: func(t *testing.T) {
				// No setup needed
			},
			want:    "",
			wantErr: true,
			errMsg:  "unmarshaling resolver config",
		},
		{
			name:      "key with special characters",
			rawConfig: []byte(`key: "my-app/api-token"`),
			setupFunc: func(t *testing.T) {
				err := keyring.Set("my-app/api-token", currentUser, "special-token-123")
				require.NoError(t, err)
			},
			want:    "special-token-123",
			wantErr: false,
		},
		{
			name:      "key with unicode characters",
			rawConfig: []byte(`key: "测试密钥"`),
			setupFunc: func(t *testing.T) {
				err := keyring.Set("测试密钥", currentUser, "unicode-value")
				require.NoError(t, err)
			},
			want:    "unicode-value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock keyring for this test
			keyring.MockInit()

			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			ctx := context.Background()
			got, err := Resolve(ctx, tt.rawConfig)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
