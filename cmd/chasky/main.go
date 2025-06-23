package main

import (
	"fmt"
	"os"

	"github.com/jcchavezs/chasky/internal/cmd"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
