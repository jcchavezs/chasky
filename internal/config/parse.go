package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Secret struct {
	Type      string `yaml:"type"`
	RawConfig json.RawMessage
}

func (s *Secret) UnmarshalYAML(data []byte) error {
	t := struct {
		Type string `yaml:"type"`
	}{}
	if err := yaml.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("unmarshaling type: %w", err)
	}

	cfg := map[string]json.RawMessage{}
	if err := yaml.UnmarshalWithOptions(data, &cfg, yaml.UseJSONUnmarshaler()); err != nil {
		return fmt.Errorf("unmarshaling provider raw configuration: %w", err)
	}

	pcfg, ok := cfg[t.Type]
	if !ok {
		return fmt.Errorf("missing provider configuration")
	}

	s.Type = t.Type
	s.RawConfig = pcfg

	return nil
}

type ToolValues struct {
	Output string            `yaml:"output"`
	Values map[string]Secret `yaml:"values"`
}

type Config map[string][]ToolValues

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

	var m Config
	if err = yaml.Unmarshal(config, &m); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return m, nil
}
