package config

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/jcchavezs/chasky/internal/log"
	"github.com/jcchavezs/chasky/internal/output/types"
	"go.uber.org/zap"
)

func commandRunner(ctx context.Context, command string, vs map[string]string) error {
	if strings.TrimSpace(command) == "" {
		log.Logger.Warn("Empty command provided, skipping execution")
		return nil
	}

	t, err := template.New(command).Parse(command)
	if err != nil {
		return fmt.Errorf("parsing pre-command template: %w", err)
	}

	s := &strings.Builder{}
	if err := t.Execute(s, vs); err != nil {
		return fmt.Errorf("executing pre-command template: %w", err)
	}

	log.Logger.Debug("Running command", zap.String("command", command))
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", s.String())
	return cmd.Run()
}

type PreList []Pre

func (p PreList) Exec(ctx context.Context, vs map[string]string) (types.Output, error) {
	for _, pre := range p {
		switch pre.Type {
		case "command":
			if err := commandRunner(ctx, pre.Command, vs); err != nil {
				return types.Output{}, fmt.Errorf("running pre command %q: %w", pre.Command, err)
			}
		default:
			log.Logger.Warn("Unknown pre hook type, skipping", zap.String("type", pre.Type))
			continue
		}
	}

	return types.Output{}, nil
}

type Pre struct {
	Type    string `yaml:"type"`
	Command string `yaml:"command"`
}

type PostList []Post

func (p PostList) Exec(ctx context.Context, vs map[string]string) error {
	for _, pre := range p {
		switch pre.Type {
		case "command":
			if err := commandRunner(ctx, pre.Command, vs); err != nil {
				return fmt.Errorf("running post command %q: %w", pre.Command, err)
			}
		default:
			log.Logger.Warn("Unknown post hook type, skipping", zap.String("type", pre.Type))
			continue
		}
	}

	return nil

}

type Post struct {
	Type    string `yaml:"type"`
	Command string `yaml:"command"`
}
