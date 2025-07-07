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
	"github.com/thediveo/enumflag/v2"
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
	Short: "Chasky is a tool to generate shell environments for your apps",
	Args:  cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger = prettyconsole.NewLogger(loglevel).
			WithOptions(zap.ErrorOutput(zapcore.AddSync(cmd.ErrOrStderr())))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		ctx := cmd.Context()
		var envName string

		conf, err := config.Parse()
		if err != nil {
			return err
		}

		envName = args[0]

		s := spinner.New(spinner.CharSets[26], 200*time.Millisecond) // Build our new spinner
		s.Prefix = fmt.Sprintf("Generating the environment for %q ", envName)
		s.FinalMSG = fmt.Sprintf("Generated environment for %q successfully\n", envName)
		s.Suffix = "\n"
		s.Start()

		appValues, ok := conf[envName]
		if !ok {
			return fmt.Errorf("unknown environment %s", envName)
		}

		env, err := environ.Render(ctx, appValues)
		if err != nil {
			return fmt.Errorf("rendering environment: %w", err)
		}
		s.Stop()

		defer func() {
			_ = env.Close()
		}()

		envvars := append(env.EnvVars, fmt.Sprintf("CHASKY_ENVNAME=%s", envName))

		c := exec.CommandContext(cmd.Context(), os.Getenv("SHELL"))
		c.Env = append(envvars, os.Environ()...)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Stdin = os.Stdin

		if err := c.Start(); err != nil {
			return fmt.Errorf("starting environment: %w", err)
		}

		for _, msg := range env.WelcomeMsgs {
			fmt.Println(msg)
		}

		return c.Wait()
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		_ = logger.Sync()
		return nil
	},
	SilenceUsage:  false,
	SilenceErrors: true,
}
