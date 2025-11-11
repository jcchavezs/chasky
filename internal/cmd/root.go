package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jcchavezs/chasky/internal/config"
	"github.com/jcchavezs/chasky/internal/environ"
	"github.com/jcchavezs/chasky/internal/log"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
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
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(editCmd)
	RootCmd.AddCommand(importCmd)
	RootCmd.AddCommand(currentCmd)
}

var RootCmd = &cobra.Command{
	Use:   "chasky [command|environ]",
	Short: "Chasky is a tool to generate shell environs for your apps",
	Example: `$ chasky my_app
$ chasky my_app -- echo "I am ${MY_USER_ENV_VAR}"
$ chasky my_app --log-level=debug -- echo "I am ${MY_USER_ENV_VAR}"`,
	Args: cobra.MinimumNArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.Init(loglevel, cmd.ErrOrStderr())
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		var (
			execCommand bool
			command     = os.Getenv("SHELL")
			commandArg  []string
		)

		if len(args) > 1 {
			if !slices.Contains(os.Args, "--") { // cobra args does not pick up -- separator
				return errors.New("unknown command")
			}

			if len(args) > 2 {
				execCommand = true
				command = args[1]
				commandArg = args[2:]
			}
		}

		ctx := cmd.Context()
		var envName string

		conf, err := config.Parse(ctx)
		if err != nil {
			return err
		}

		envName = args[0]

		s := spinner.New(spinner.CharSets[26], 200*time.Millisecond) // Build our new spinner
		s.Prefix = fmt.Sprintf("Generating the environment for %q", envName)
		s.FinalMSG = fmt.Sprintf("Generated environment for %q successfully!\n", envName)
		s.Suffix = "\n"
		s.Start()

		cfg, ok := conf[envName]
		if !ok {
			return fmt.Errorf("unknown environment %s", envName)
		}

		env, err := environ.Render(ctx, cfg.Values)
		if err != nil {
			return fmt.Errorf("rendering environment: %w", err)
		}
		s.Stop()

		defer func() {
			if err = env.Close(); err != nil {
				log.Logger.Warn("Failed to close environment", zap.Error(err))
			}
		}()

		envvars := append(env.EnvVars, fmt.Sprintf("CHASKY_ENVNAME=%s", envName))
		c := exec.CommandContext(cmd.Context(), command, commandArg...)
		c.Env = append(envvars, os.Environ()...)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Stdin = os.Stdin

		if err := c.Start(); err != nil {
			return fmt.Errorf("starting environment: %w", err)
		}
		if !execCommand {
			if len(env.WelcomeMsgs) > 0 {
				fmt.Println("")
				for _, msg := range env.WelcomeMsgs {
					fmt.Println(msg)
				}
			}
		}

		return c.Wait()
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return log.Close()
	},
	SilenceUsage:  false,
	SilenceErrors: true,
}
