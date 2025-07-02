package output

import (
	"context"

	"github.com/jcchavezs/chasky/internal/log"
	"github.com/jcchavezs/chasky/internal/output/env"
	"github.com/jcchavezs/chasky/internal/output/gcloud"
	"github.com/jcchavezs/chasky/internal/output/netrc"
	"github.com/jcchavezs/chasky/internal/output/types"
)

func Exec(ctx context.Context, name string, values map[string]string) (types.Output, error) {
	switch name {
	case "env":
		return env.Exec(ctx, values)
	case "gcloud":
		return gcloud.Exec(ctx, values)
	case "netrc":
		return netrc.Exec(ctx, values)
	default:
		log.Logger.Debug("unknown output type, defaulting to env")
		return env.Exec(ctx, values)
	}
}
