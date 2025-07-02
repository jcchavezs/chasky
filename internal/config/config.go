package config

import (
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
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
		return fmt.Errorf("missing provider configuration for %q", t.Type)
	}

	s.Type = t.Type
	s.RawConfig = pcfg

	return nil
}

func (s Secret) MarshalYAML() ([]byte, error) {
	f, err := parser.ParseBytes(s.RawConfig, 0)
	if err != nil {
		return nil, fmt.Errorf("parsing raw config: %w", err)
	}

	m := map[string]string{}
	if err := yaml.NodeToValue(f.Docs[0].Body, &m); err != nil {
		return nil, fmt.Errorf("marshaling raw config: %w", err)
	}

	return yaml.Marshal(map[string]any{
		"type": s.Type,
		s.Type: m,
	})
}

type ToolValues struct {
	Output string            `yaml:"output,omitempty"`
	Values map[string]Secret `yaml:"values"`
}

type Config map[string][]ToolValues
