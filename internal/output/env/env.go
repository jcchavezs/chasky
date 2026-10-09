package env

import (
	"context"
	"fmt"
	"strings"

	"github.com/jcchavezs/chasky/internal/output/file"
	"github.com/jcchavezs/chasky/internal/output/types"
)

const leap = 2

// anonymizeSecret replaces all but the first and last `leap` characters of a string with asterisks.
// If the string is shorter than or equal to `2 * leap`, it replaces the entire string with asterisks.
// It also limits the number of asterisks to a maximum of 30 for very long strings.
func anonymizeSecret(s string) string {
	if len(s) <= 2*leap {
		return strings.Repeat("*", len(s))
	}
	return s[:leap] + strings.Repeat("*", min(len(s)-2*leap, 30)) + s[len(s)-leap:]
}

var wroteHeader bool

func Exec(ctx context.Context, values map[string]string) (types.Output, error) {
	if len(values) == 0 {
		return types.Output{}, nil
	}

	// When pipe delivery is enabled the secrets must not land in the process
	// environment (where they would be readable through /proc/<pid>/environ);
	// expose them through an on-demand env file instead.
	if file.Enabled(ctx) {
		return execFile(ctx, values)
	}

	return execEnv(values)
}

func execEnv(values map[string]string) (types.Output, error) {
	var env []string
	for k, v := range values {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	wm := &strings.Builder{}
	if !wroteHeader {
		fmt.Fprint(wm, "The following environment variables were set:")
		wroteHeader = true
	}

	for k, v := range values {
		fmt.Fprintf(wm, "\n%s=%s (length: %d)", k, anonymizeSecret(v), len(v))
	}

	return types.Output{
		WelcomeMsg: wm.String(),
		EnvVars:    env,
	}, nil
}

func execFile(ctx context.Context, values map[string]string) (types.Output, error) {
	s := &strings.Builder{}
	for k, v := range values {
		fmt.Fprintf(s, "%s=%s\n", k, v)
	}

	sink, err := file.NewSink(ctx, ".env")
	if err != nil {
		return types.Output{}, fmt.Errorf("creating env file: %w", err)
	}

	if _, err := sink.Write([]byte(s.String())); err != nil {
		_ = sink.Close()
		return types.Output{}, fmt.Errorf("writing env file: %w", err)
	}

	return types.Output{
		WelcomeMsg: `The environment variables were written to the file in the $ENV_FILE env var.

Load them into your shell with:
$ set -a && eval "$(cat "$ENV_FILE")" && set +a

(Read the file with cat rather than sourcing it directly: $ENV_FILE is a pipe,
and some shells read 0 bytes when they size the file up front.)`,
		EnvVars: []string{fmt.Sprintf("ENV_FILE=%s", sink.Path())},
		Closer:  sink.Close,
	}, nil
}
