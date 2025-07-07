package environ

import (
	"context"
	"errors"
	"fmt"

	"github.com/jcchavezs/chasky/internal/config"
	"github.com/jcchavezs/chasky/internal/output"
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

func Render(ctx context.Context, tvs []config.EnvironmentValues) (Environment, error) {
	e := Environment{}

	for _, tv := range tvs {
		vs := map[string]string{}

		for name, s := range tv.Values {
			v, err := source.Exec(ctx, s.Type, s.RawConfig)
			if err != nil {
				return Environment{}, fmt.Errorf("resolving value for %s: %w", name, err)
			}

			vs[name] = v
		}

		o, err := output.Exec(ctx, tv.Output, vs)
		if err != nil {
			return Environment{}, fmt.Errorf("executing output %q: %w", tv.Output, err)
		}

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

	return e, nil
}
