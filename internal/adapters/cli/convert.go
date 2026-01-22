package cli

import (
	"github.com/spf13/cobra"
)

// NewConvertCommand creates the convert command
func NewConvertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert a file to another format",
		Long:  "Convert a source file to a target format",
		RunE:  runConvert,
	}

	cmd.Flags().StringP("source", "s", "", "Source file to convert")
	cmd.Flags().StringP("target", "t", "", "Target format (e.g., pdf, webp)")
	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("target")

	return cmd
}

func runConvert(cmd *cobra.Command, args []string) error {
	// Implementation will connect to domain core
	return nil
}

