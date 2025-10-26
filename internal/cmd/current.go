package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Shows current environ",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		ename := os.Getenv("CHASKY_ENVNAME")
		if ename == "" {
			return errors.New("no environ loaded")
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout(), os.Getenv("CHASKY_ENVNAME"))
		return nil
	},
}
