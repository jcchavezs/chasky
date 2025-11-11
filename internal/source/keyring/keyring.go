package keyring

import (
	"context"
	"errors"
	"fmt"
	"os/user"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/zalando/go-keyring"
)

func getCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	return currentUser.Username, nil
}

type config struct {
	Key string `yaml:"key"`
}

func Resolve(ctx context.Context, rawConfig []byte) (string, error) {
	var c config
	if err := yaml.Unmarshal(rawConfig, &c); err != nil {
		return "", fmt.Errorf("unmarshaling resolver config: %w", err)
	}

	if c.Key == "" {
		return "", errors.New("missing keyring.key value")
	}

	user, err := getCurrentUser()
	if err != nil {
		return "", fmt.Errorf("getting current user: %w", err)
	}

	val, err := keyring.Get(c.Key, user)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("getting keyring value: %w", err)
	}

	return val, nil
}

func Persist(ctx context.Context, key, value string, force bool) (string, error) {
	user, err := getCurrentUser()
	if err != nil {
		return "", fmt.Errorf("getting current user: %w", err)
	}

	key = fmt.Sprintf("com.github.jcchavezs.pakay-%s", strings.ToLower(key))

	if err := keyring.Set(key, user, value); err != nil {
		return "", err
	}

	return key, nil
}
