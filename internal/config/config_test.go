package config

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalConfig(t *testing.T) {
	cfg := &Config{}
	input := `---
scu: # service catalog updater
# aaa
- output: env
  values:
    OPSLEVEL_API_TOKEN:
      static:
        value: my_token
      type: static
`

	require.NoError(t, yaml.Unmarshal([]byte(input), cfg))
	t.Fail()
}
