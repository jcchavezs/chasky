package env

import (
	"context"
	"fmt"

	"github.com/jcchavezs/chasky/internal/config"
	"github.com/jcchavezs/chasky/internal/resolver"
)

func GenerateEnv(ctx context.Context, cfg config.ToolSecrets) ([]string, error) {
	if len(cfg) == 0 {
		return nil, nil
	}

	var env []string
	for k, s := range cfg {
		v, err := resolver.Exec(ctx, s.Type, s.RawConfig)
		if err != nil {
			return nil, fmt.Errorf("rendering %s: %w", k, err)
		}

		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env, nil
}
