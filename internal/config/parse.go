package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

func ConfigPath() (string, error) {
	dir, err := os.UserHomeDir()
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

	return parse(path)
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
