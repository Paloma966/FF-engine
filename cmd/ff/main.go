package main

import (
	"fmt"
	"os"

	"github.com/chun/fiction_factory/internal/cli"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ff: error: %v\n", err)
		os.Exit(1)
	}
}
