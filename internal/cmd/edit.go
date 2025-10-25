package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jcchavezs/chasky/internal/config"
	"github.com/jcchavezs/chasky/internal/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	editCmd.Flags().String("editor", "", "Editor to use for editing the config file")
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit chasky config",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return fmt.Errorf("getting config path: %w", err)
		}

		editor, found := getEditor(cmd)
		if !found {
			log.Logger.Warn("EDITOR env var not found, using nano")
		}

		log.Logger.Info("Launching editor", zap.String("editor", editor), zap.String("path", path))

		execCmd := exec.CommandContext(cmd.Context(), editor, path)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		return execCmd.Run()
	},
}

func getEditor(cmd *cobra.Command) (string, bool) {
	if argEditor, err := cmd.Flags().GetString("editor"); err != nil {
		log.Logger.Error("getting editor from flag", zap.Error(err))
	} else if argEditor != "" {
		return argEditor, true
	}

	if editor := os.Getenv("EDITOR"); editor == "" {
		return "nano", false
	} else {
		return editor, true
	}
}
