package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/eka026/File-Format-Converter/internal/adapters/cli"
)

func main() {
	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

