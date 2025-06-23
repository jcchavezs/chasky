package resolver

import (
	"context"
	"errors"

	"github.com/jcchavezs/chasky/internal/resolver/bash"
)

func Exec(ctx context.Context, _type string, rawConfig []byte) (string, error) {
	switch _type {
	case "bash":
		return bash.Resolve(ctx, rawConfig)
	}

	return "", errors.New("type not found")
}
