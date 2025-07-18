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

		editor, foundEditor := getEditor()
		if !foundEditor {
			logger.Warn("EDITOR env var not found, using nano")
		}

		logger.Info("Launching editor", zap.String("editor", editor), zap.String("path", path))

		execCmd := exec.CommandContext(cmd.Context(), editor, path)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		return execCmd.Run()
	},
}

func getEditor() (string, bool) {
	if editor := os.Getenv("EDITOR"); editor == "" {
		return "nano", false
	} else {
		return editor, true
	}
}
