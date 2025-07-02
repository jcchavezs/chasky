package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jcchavezs/chasky/internal/config"
	"github.com/jcchavezs/chasky/internal/source/keyring"
	"github.com/spf13/cobra"
)

func init() {
	importCmd.Flags().String("dst-tool", "", "The destination tool to import the secret into")
}

var dstTool string

var importCmd = &cobra.Command{
	Use:   "import <source> [key1=val1 [key2=val2 [...]]]",
	Short: "Imports a secret into a source and adds it to a destination tool",
	Args:  cobra.MinimumNArgs(2),
	Example: `$ chasky import keyring OPENAI_API_KEY=foo --dst-tool mytool
$ chasky import keyring JIRA_EMAIL=bar JIRA_API_TOKEN=baz --dst-tool mytool
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		if source != "keyring" {
			return errors.New("unsupported source")
		}

		var err error
		if dstTool, err = cmd.Flags().GetString("dst-tool"); err != nil {
			return fmt.Errorf("getting destination tool flag: %w", err)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		vals := args[1:]

		tvs := config.ToolValues{
			Values: map[string]config.Secret{},
		}

		for _, val := range vals {
			k, v, ok := strings.Cut(val, "=")
			if !ok {
				return fmt.Errorf("invalid importing value %s", val)
			}

			rawCfg, err := keyring.Persist(cmd.Context(), k, v)
			if err != nil {
				return fmt.Errorf("persisting value: %w", err)
			}

			if dstTool == "" {
				cmd.Printf("Credential %q successfully imported into %s.\n", k, source)
			} else {
				tvs.Values[k] = config.Secret{
					Type:      source,
					RawConfig: json.RawMessage(rawCfg),
				}
			}
		}

		err := config.AppendValues(dstTool, tvs)

		if err != nil {
			return fmt.Errorf("appending config values: %w", err)
		}

		return nil
	},
}
