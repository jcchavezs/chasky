package netrc

import (
	"context"
	"errors"
	"fmt"

	"os"
	"strings"

	"github.com/jcchavezs/chasky/internal/log"
	"github.com/jcchavezs/chasky/internal/output/types"
)

type netRC struct {
	Machine  string
	Login    string
	Password string
}

func (n netRC) String() string {
	b := []string{}
	if n.Machine == "default" || n.Machine == "" {
		b = append(b, "default")
	} else {
		b = append(b, fmt.Sprintf("machine %s", n.Machine))
	}

	b = append(b, fmt.Sprintf("login %s", n.Login))

	if n.Password != "" {
		b = append(b, fmt.Sprintf("password %s", n.Password))
	}

	return strings.Join(b, " ")
}

var f *os.File

func Exec(ctx context.Context, values map[string]string) (types.Output, error) {
	if len(values) == 0 {
		return types.Output{}, errors.New("empty values")
	}

	creds := netRC{}
	for k, v := range values {
		switch k {
		case "machine":
			creds.Machine = v
		case "login":
			creds.Login = v
		case "password":
			creds.Password = v
		default:
			return types.Output{}, fmt.Errorf("unknown field %s", k)
		}

		if strings.Contains(v, " ") {
			return types.Output{}, fmt.Errorf("%s value should not contain spaces", k)
		}
	}

	if creds.Machine == "" {
		creds.Machine = "default"
	}

	if creds.Login == "" {
		return types.Output{}, fmt.Errorf("login is required")
	}

	log.Logger.Debug("Creating netrc file")
	var (
		err     error
		isFirst = false
	)
	if f == nil {
		isFirst = true
		f, err = os.CreateTemp(os.TempDir(), ".netrc")
		if err != nil {
			return types.Output{}, fmt.Errorf("creating credentials file: %w", err)
		}
	}

	if _, err := f.WriteString(creds.String()); err != nil {
		return types.Output{}, fmt.Errorf("writing credentials: %w", err)
	}

	_, _ = fmt.Fprintln(f, "")

	if !isFirst {
		return types.Output{}, nil
	}

	return types.Output{
		WelcomeMsg: `The location of the .netrc file that has been created can be found in the $NETRC_FILE env var

For example:
$ curl --netrc-file $NETRC_FILE ....`,
		EnvVars: []string{fmt.Sprintf("NETRC_FILE=%s", f.Name())},
		Closer: func() error {
			_ = f.Close()
			log.Logger.Debug("Deleting temporary netrc file")
			return os.Remove(f.Name())
		},
	}, nil
}
