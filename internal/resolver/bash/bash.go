package bash

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/goccy/go-yaml"
)

type config struct {
	Command string `yaml:"command"`
}

func Resolve(ctx context.Context, rawConfig []byte) (string, error) {
	var c config
	if err := yaml.Unmarshal(rawConfig, &c); err != nil {
		return "", fmt.Errorf("unamrshaling resolver config: %w", err)
	}

	if c.Command == "" {
		return "", errors.New("missing bash.command value")
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", c.Command)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("executing command: %w", err)
	}

	return string(bytes.TrimSpace(out)), nil
}
