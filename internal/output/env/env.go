package env

import (
	"context"
	"fmt"
	"strings"

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
