package source

import (
	"context"
	"errors"

	"github.com/jcchavezs/chasky/internal/source/bash"
	"github.com/jcchavezs/chasky/internal/source/static"
)

func Exec(ctx context.Context, _type string, rawConfig []byte) (string, error) {
	switch _type {
	case "bash":
		return bash.Resolve(ctx, rawConfig)
	case "static":
		return static.Resolve(ctx, rawConfig)
	}

	return "", errors.New("type not found")
}
