package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

func AppendValues(tool string, vs ToolValues) error {
	path, err := ConfigPath()
	if err != nil {
		return fmt.Errorf("getting configuration path: %w", err)
	}

	return appendValues(path, tool, vs)
}

func appendValues(filepath string, tool string, vs ToolValues) error {
	cfg, err := parse(filepath)
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	if cfg == nil {
		cfg = Config{}
	}

	if _, ok := cfg[tool]; !ok {
		cfg[tool] = []ToolValues{}
	}

	cfg[tool] = append(cfg[tool], vs)

	b, err := yaml.MarshalWithOptions(cfg, yaml.UseJSONMarshaler())
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(filepath, b, 0666)
}
