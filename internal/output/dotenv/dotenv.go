package dotenv

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jcchavezs/chasky/internal/output/file"
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

	sink, err := file.NewSink(ctx, ".env")
	if err != nil {
		return types.Output{}, fmt.Errorf("creating credentials file: %w", err)
	}

	if _, err := sink.Write([]byte(s.String())); err != nil {
		_ = sink.Close()
		return types.Output{}, fmt.Errorf("writing credentials: %w", err)
	}

	return types.Output{
		WelcomeMsg: `The location of the .env file that has been created can be found in the $DOTENV_FILE env var

For example:
$ docker run --env-file $DOTENV_FILE ....`,
		EnvVars: []string{fmt.Sprintf("DOTENV_FILE=%s", sink.Path())},
		Closer:  sink.Close,
	}, nil
}
