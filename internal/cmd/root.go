package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jcchavezs/chasky/internal/config"
	"github.com/jcchavezs/chasky/internal/environ"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var LevelIds = map[zapcore.Level][]string{
	zap.DebugLevel: {"debug"},
	zap.InfoLevel:  {"info"},
	zap.WarnLevel:  {"warn"},
	zap.ErrorLevel: {"error"},
}

var loglevel zapcore.Level = zapcore.ErrorLevel

func init() {
	RootCmd.PersistentFlags().Var(
		enumflag.New(&loglevel, "string", LevelIds, enumflag.EnumCaseInsensitive),
		"log-level",
		"Sets the log level",
	)
	RootCmd.AddCommand(EditCmd)
	RootCmd.AddCommand(importCmd)
}

var logger *zap.Logger

var RootCmd = &cobra.Command{
	Use:   "chasky",
	Short: "Chasky is a tool to generate environment variables for various tools",
	Args:  cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger = prettyconsole.NewLogger(loglevel).
			WithOptions(zap.ErrorOutput(zapcore.AddSync(cmd.ErrOrStderr())))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		ctx := cmd.Context()
		var toolName string

		conf, err := config.Parse()
		if err != nil {
			return err
		}

		toolName = args[0]

		s := spinner.New(spinner.CharSets[26], 200*time.Millisecond) // Build our new spinner
		s.Prefix = fmt.Sprintf("Generating env vars for %q ", toolName)
		s.FinalMSG = fmt.Sprintf("Generated env vars for %q successfully\n", toolName)
		s.Start()

		toolValues, ok := conf[toolName]
		if !ok {
			return fmt.Errorf("unknown tool %s", toolName)
		}

		env, err := environ.Render(ctx, toolValues)
		if err != nil {
			return fmt.Errorf("generating env vars: %w", err)
		}
		s.Stop()

		defer func() {
			_ = env.Close()
		}()

		envvars := append(env.EnvVars, fmt.Sprintf("CHASKY_ENV=%s", toolName))

		c := exec.CommandContext(cmd.Context(), os.Getenv("SHELL"))
		c.Env = append(envvars, os.Environ()...)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Stdin = os.Stdin

		return c.Run()
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		_ = logger.Sync()
		return nil
	},
	SilenceUsage:  false,
	SilenceErrors: true,
}
