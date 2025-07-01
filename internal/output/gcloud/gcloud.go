package gcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jcchavezs/chasky/internal/output/types"
)

type GCloudCredentials struct {
	Account        string `json:"account,omitempty"`
	ClientID       string `json:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty"`
	QuotaProjectID string `json:"quota_project_id,omitempty"`
	RefreshToken   string `json:"refresh_token,omitempty"`
	Type           string `json:"type,omitempty"`
	UniverseDomain string `json:"universe_domain,omitempty"`
}

func Exec(ctx context.Context, values map[string]string) (types.Output, error) {
	if len(values) == 0 {
		return types.Output{}, nil
	}

	creds := GCloudCredentials{}
	for k, v := range values {
		switch k {
		case "account":
			creds.Account = v
		case "client_id":
			creds.ClientID = v
		case "client_secret":
			creds.ClientSecret = v
		case "quota_project_id":
			creds.QuotaProjectID = v
		case "refresh_token":
			creds.RefreshToken = v
		case "type":
			creds.Type = v
		case "universe_domain":
			creds.UniverseDomain = v
		}
	}

	b, err := json.Marshal(&creds)
	if err != nil {
		return types.Output{}, fmt.Errorf("marshaling credentials: %w", err)
	}

	f, err := os.CreateTemp(os.TempDir(), "application_default_credentials.json")
	if err != nil {
		return types.Output{}, fmt.Errorf("creating credentials file: %w", err)
	}

	if _, err := f.Write(b); err != nil {
		return types.Output{}, fmt.Errorf("writing credentials: %w", err)
	}

	_ = f.Close()

	return types.Output{
		EnvVars: []string{fmt.Sprintf("GOOGLE_APPLICATION_CREDENTIALS=%s", f.Name())},
		Closer: func() error {
			return os.Remove(f.Name())
		},
	}, nil
}
