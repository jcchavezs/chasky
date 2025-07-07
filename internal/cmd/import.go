package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jcchavezs/chasky/internal/source/keyring"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <source> [key1=val1 [key2=val2 [...]]]",
	Short: "Imports a secret into a source",
	Args:  cobra.MinimumNArgs(2),
	Example: `$ chasky import keyring OPENAI_API_KEY=foo
$ chasky import keyring JIRA_EMAIL=bar JIRA_API_TOKEN=baz
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		if source != "keyring" {
			return errors.New("unsupported source")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		vals := args[1:]

		creds := map[string]string{}
		for _, val := range vals {
			k, v, ok := strings.Cut(val, "=")
			if !ok {
				return fmt.Errorf("invalid importing value %s", val)
			}

			key, err := keyring.Persist(cmd.Context(), k, v)
			if err != nil {
				return fmt.Errorf("persisting value: %w", err)
			}
			creds[k] = key
		}

		cmd.Printf("Credentials successfully imported into %s.\n\n", source)
		cmd.Printf("To use them in a given environment, type `chasky edit` and add:\n")

		yaml := &strings.Builder{}
		_, _ = yaml.WriteString(`
---
# ...
- values:`)

		for name, key := range creds {
			fmt.Fprintf(yaml, `
  - %s:
      type: keyring
      keyring:
        key: %s
`, name, key)
		}

		cmd.Println(yaml)

		return nil
	},
}
