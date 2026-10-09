package pass

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/goccy/go-yaml"
)

type config struct {
	Key string `yaml:"key"`
}

func Resolve(ctx context.Context, rawConfig []byte) (string, error) {
	var c config
	if err := yaml.Unmarshal(rawConfig, &c); err != nil {
		return "", fmt.Errorf("unmarshaling resolver config: %w", err)
	}

	if c.Key == "" {
		return "", errors.New("missing pass.key value")
	}

	cmd := exec.CommandContext(ctx, "pass", "show", c.Key)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("executing command: %w", err)
	}

	return string(bytes.TrimSpace(out)), nil
}

func Persist(ctx context.Context, key, value string, force bool) (string, error) {
	key = fmt.Sprintf("com.github.jcchavezs.pakay-%s", strings.ToLower(key))

	flags := []string{"insert", key}
	if force {
		flags = append(flags, "--force")
	}

	cmd := exec.CommandContext(ctx, "pass", flags...)
	cmd.Stderr = os.Stderr
	_, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("executing command: %w", err)
	}

	return key, nil
}
