package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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

// resolveCommand splits the positional args into the environment name and an
// optional custom command to run inside that environment. Everything after the
// `--` separator is the custom command; dashPos is the number of positional
// args that appeared before `--` (or -1 when no `--` was given, matching
// cobra's Command.ArgsLenAtDash). When no custom command is given, the provided
// shell is used instead.
func resolveCommand(args []string, dashPos int, shell string) (envName, command string, commandArg []string, isCustomCommand bool, err error) {
	command = shell

	envArgs := args
	if dashPos != -1 {
		envArgs = args[:dashPos]
		if cmdArgs := args[dashPos:]; len(cmdArgs) > 0 {
			command = cmdArgs[0]
			commandArg = cmdArgs[1:]
			isCustomCommand = true
		}
	}

	if len(envArgs) != 1 {
		return "", "", nil, false, errors.New("unknown command")
	}

	return envArgs[0], command, commandArg, isCustomCommand, nil
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

		if os.Getenv("CHASKY_ENVNAME") != "" {
			return errors.New("cannot run chasky inside a chasky environment")
		}

		// ArgsLenAtDash reports how many positional args appeared before the
		// `--` separator (or -1 when no `--` was given).
		envName, command, commandArg, isCustomCommand, err := resolveCommand(args, cmd.ArgsLenAtDash(), os.Getenv("SHELL"))
		if err != nil {
			return err
		}

		ctx := cmd.Context()

		conf, err := config.Parse(ctx)
		if err != nil {
			return err
		}

		var afterRender = func() {}
		if !isCustomCommand {
			s := spinner.New(spinner.CharSets[26], 200*time.Millisecond) // Build our new spinner
			s.Prefix = fmt.Sprintf("Generating the environment for %q", envName)
			s.FinalMSG = fmt.Sprintf("Generated environment for %q successfully!\n", envName)
			s.Suffix = "\n"
			s.Start()
			afterRender = s.Stop
		}

		cfg, ok := conf[envName]
		if !ok {
			return fmt.Errorf("unknown environment %s", envName)
		}

		env, err := environ.Render(ctx, cfg.Values)
		defer func() {
			// Ensure we run the post hooks even on render error
			if err = env.Close(); err != nil {
				log.Logger.Warn("Failed to close environment", zap.Error(err))
			}
		}()

		if err != nil {
			return fmt.Errorf("rendering environment: %w", err)
		}

		afterRender()

		envvars := append(env.EnvVars, fmt.Sprintf("CHASKY_ENVNAME=%s", envName))
		c := exec.CommandContext(cmd.Context(), command, commandArg...)
		c.Env = append(envvars, os.Environ()...)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		c.Stdin = os.Stdin

		if err := c.Start(); err != nil {
			return fmt.Errorf("starting environment: %w", err)
		}

		if !isCustomCommand {
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
