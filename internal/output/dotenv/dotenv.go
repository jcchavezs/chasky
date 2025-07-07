package dotenv

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jcchavezs/chasky/internal/output/types"
)

// Exec generates a .env file from the provided values and returns the output
func Exec(ctx context.Context, values map[string]string) (types.Output, error) {
	if len(values) == 0 {
		return types.Output{}, errors.New("empty values")
	}

	s := &strings.Builder{}
	for k, v := range values {
		fmt.Fprintf(s, "%s=%s\n", k, v)
	}

	f, err := os.CreateTemp(os.TempDir(), ".env")
	if err != nil {
		return types.Output{}, fmt.Errorf("creating credentials file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if _, err := f.WriteString(s.String()); err != nil {
		return types.Output{}, fmt.Errorf("writing credentials: %w", err)
	}

	return types.Output{
		WelcomeMsg: `The location of the .env file that has been created can be found in the $DOTENV_FILE env var

For example:
$ docker run --env-file $DOTENV_FILE ....`,
		EnvVars: []string{fmt.Sprintf("DOTENV_FILE=%s", f.Name())},
		Closer: func() error {
			return os.Remove(f.Name())
		},
	}, nil
}
