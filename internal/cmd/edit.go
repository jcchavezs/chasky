package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jcchavezs/chasky/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var EditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit chasky config",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return fmt.Errorf("getting config path: %w", err)
		}

		editor := getEditor()

		logger.Info("Launching debug", zap.String("editor", editor), zap.String("path", path))

		execCmd := exec.CommandContext(cmd.Context(), editor, path)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		return execCmd.Run()
	},
}

func getEditor() string {
	if editor := os.Getenv("EDITOR"); editor == "" {
		return "nano"
	} else {
		return editor
	}
}
