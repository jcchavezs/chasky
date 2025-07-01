package env

import (
	"context"
	"fmt"

	"github.com/jcchavezs/chasky/internal/output/types"
)

func Exec(ctx context.Context, values map[string]string) (types.Output, error) {
	if len(values) == 0 {
		return types.Output{}, nil
	}

	var env []string
	for k, v := range values {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return types.Output{EnvVars: env}, nil
}
