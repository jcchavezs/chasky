package config

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

type EnvironmentValues struct {
	OutputType string            `yaml:"output,omitempty"`
	Values     map[string]Secret `yaml:"values"`
}

type ConfigEntry struct {
	Description string
	Values      []EnvironmentValues
}

type Config map[string]ConfigEntry

func makeConfigYAMLUnmarshaler(cm yaml.CommentMap) func(ctx context.Context, t *Config, b []byte) error {
	return func(ctx context.Context, t *Config, b []byte) error {
		var rCfg map[string][]EnvironmentValues

		if err := yaml.Unmarshal(b, &rCfg); err != nil {
			return fmt.Errorf("unmarshaling raw configuration: %w", err)
		}

		var ut = *t
		for k, v := range rCfg {
			var desc string
			if cmts, ok := cm["$."+k]; ok && len(cmts) > 0 {
				desc = strings.TrimSpace(strings.Join(cmts[0].Texts, " "))
			}

			ut[k] = ConfigEntry{
				Values:      v,
				Description: desc,
			}
		}

		return nil
	}
}
