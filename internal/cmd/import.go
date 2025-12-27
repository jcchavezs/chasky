package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/jcchavezs/chasky/internal/source/keyring"
	"github.com/jcchavezs/chasky/internal/source/pass"
	"github.com/spf13/cobra"
)

var supportedImportSources = map[string]func(ctx context.Context, key, value string, force bool) (string, error){
	"keyring": keyring.Persist,
	"pass":    pass.Persist,
}

func init() {
	importCmd.Flags().Bool("force", false, "overwrite existing values")
}

var importCmd = &cobra.Command{
	Use:   "import <source> [key1=val1 [key2=val2 [...]]]",
	Short: "Imports a secret into a source",
	Args:  cobra.MinimumNArgs(2),
	Example: `$ chasky import keyring OPENAI_API_KEY=foo
$ chasky import pass JIRA_EMAIL=bar JIRA_API_TOKEN=baz --force
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		if _, found := supportedImportSources[source]; !found {
			return fmt.Errorf("unsupported source %q", source)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		source := args[0]
		vals := args[1:]

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return fmt.Errorf("reading force flag: %w", err)
		}

		persister := supportedImportSources[source]

		creds := map[string]string{}
		for _, val := range vals {
			k, v, ok := strings.Cut(val, "=")
			if !ok {
				return fmt.Errorf("invalid importing value %s", val)
			}

			key, err := persister(cmd.Context(), k, v, force)
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
  - %[1]s:
      type: %[3]s
      %[3]s:
        key: %[2]s
`, name, key, source)
		}

		cmd.Println(yaml)

		return nil
	},
}
