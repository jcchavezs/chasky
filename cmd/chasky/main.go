package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jcchavezs/chasky/internal/cmd"
)

func main() {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	defer func() {
		signal.Stop(signals)
		cancel()
	}()

	go func() {
		select {
		case <-signals:
			cancel()

		case <-ctx.Done():
		}
	}()

	if err := cmd.RootCmd.ExecuteContext(ctx); err != nil {
		fmt.Printf("ERROR: %v.\n", err)
		exitCode = 1
		return
	}
}
