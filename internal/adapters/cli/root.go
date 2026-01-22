package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root CLI command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "file-format-converter",
		Short: "File Format Converter - Local-First Open Source File Converter",
		Long:  "A command-line tool for converting files between different formats",
	}

	rootCmd.AddCommand(NewConvertCommand())
	rootCmd.AddCommand(NewListFormatsCommand())

	return rootCmd
}

