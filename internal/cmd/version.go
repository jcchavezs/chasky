package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set at build time via ldflags
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of chasky",
	Run: func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "chasky %s\n", version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
