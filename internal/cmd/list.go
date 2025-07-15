package cmd

import (
	"github.com/jcchavezs/chasky/internal/config"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List chasky environments",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.Parse()
		if err != nil {
			return err
		}

		for env := range conf {
			cmd.Printf("- %s\n", env)
		}

		return nil
	},
}
