package cmd

import (
	"maps"
	"slices"

	"github.com/jcchavezs/chasky/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List chasky environs",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.Parse(cmd.Context())
		if err != nil {
			return err
		}

		keys := slices.Sorted(maps.Keys(conf))

		for _, k := range keys {
			desc := conf[k].Description
			if desc == "" {
				cmd.Printf("- %s\n", k)
			} else {
				cmd.Printf("- %s: %s\n", k, desc)
			}
		}

		return nil
	},
}
