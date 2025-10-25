package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jcchavezs/chasky/internal/log"
	"go.uber.org/zap"

	"github.com/goccy/go-yaml"
)

var userHomeDir = os.UserHomeDir

func ConfigPath() (string, error) {
	dir, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting homedir: %w", err)
	}

	return filepath.Join(dir, ".chasky.yaml"), nil
}

func Parse() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, fmt.Errorf("getting config path: %w", err)
	}

	cfg, err := parse(path)
	if errors.Is(err, os.ErrNotExist) {
		log.Logger.Debug("config file does not exist, returning empty config", zap.String("path", path))
		return Config{}, nil
	}

	return cfg, err
}

func parse(filepath string) (Config, error) {
	config, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var m = Config{}
	if err = yaml.Unmarshal(config, &m); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return m, nil
}
