package static

import (
	"context"
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
)

type config struct {
	Value string `yaml:"value"`
}

func Resolve(ctx context.Context, rawConfig []byte) (string, error) {
	var c config
	if err := yaml.Unmarshal(rawConfig, &c); err != nil {
		return "", fmt.Errorf("unamrshaling resolver config: %w", err)
	}

	if c.Value == "" {
		return "", errors.New("missing static.value value")
	}

	return c.Value, nil
}
