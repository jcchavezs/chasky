package environ

import (
	"context"
	"errors"
	"fmt"

	"github.com/jcchavezs/chasky/internal/config"
	"github.com/jcchavezs/chasky/internal/output"
	"github.com/jcchavezs/chasky/internal/output/types"
	"github.com/jcchavezs/chasky/internal/source"
)

type Environment struct {
	WelcomeMsgs []string
	EnvVars     []string
	closers     []func() error
}

func (e Environment) Close() error {
	errs := make([]error, 0, len(e.closers))
	for _, c := range e.closers {
		errs = append(errs, c())
	}

	return errors.Join(errs...)
}

func appendOutput(e *Environment, o types.Output) {
	if o.WelcomeMsg != "" {
		e.WelcomeMsgs = append(e.WelcomeMsgs, o.WelcomeMsg)
	}

	if len(o.EnvVars) > 0 {
		e.EnvVars = append(e.EnvVars, o.EnvVars...)
	}

	if o.Closer != nil {
		e.closers = append(e.closers, o.Closer)
	}
}

func Render(ctx context.Context, tvs []config.EnvironmentValues) (Environment, error) {
	e := Environment{}

	for _, tv := range tvs {
		vars := map[string]string{}

		for name, s := range tv.Values {
			v, err := source.Exec(ctx, s.Type, s.RawConfig)
			if err != nil {
				return Environment{}, fmt.Errorf("resolving value for %s: %w", name, err)
			}

			vars[name] = v
		}

		pre, err := tv.Pre.Exec(ctx, vars)
		if err != nil {
			e.closers = append(e.closers, postHooks(tv, ctx, vars))
			return Environment{closers: e.closers}, fmt.Errorf("running pre-output hooks: %w", err)
		}
		appendOutput(&e, pre)

		o, err := output.Exec(ctx, tv.OutputType, vars)
		if err != nil {
			e.closers = append(e.closers, postHooks(tv, ctx, vars))
			return Environment{closers: e.closers}, fmt.Errorf("executing output %q: %w", tv.OutputType, err)
		}
		appendOutput(&e, o)

		// Post hooks
		e.closers = append(e.closers, postHooks(tv, ctx, vars))
	}

	return e, nil
}

func postHooks(tv config.EnvironmentValues, ctx context.Context, vars map[string]string) func() error {
	return func() error {
		if err := tv.Post.Exec(ctx, vars); err != nil {
			return fmt.Errorf("running post-output hooks: %w", err)
		}

		return nil
	}
}
