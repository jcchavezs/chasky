package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Shows current environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), os.Getenv("CHASKY_ENVNAME"))
		return nil
	},
}
