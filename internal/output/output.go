package output

import (
	"context"

	"github.com/jcchavezs/chasky/internal/output/env"
	"github.com/jcchavezs/chasky/internal/output/gcloud"
	"github.com/jcchavezs/chasky/internal/output/types"
)

func Exec(ctx context.Context, name string, values map[string]string) (types.Output, error) {
	switch name {
	case "env":
		return env.Exec(ctx, values)
	case "gcloud":
		return gcloud.Exec(ctx, values)
	}

	return types.Output{}, nil
}
