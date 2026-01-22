package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

// NewListFormatsCommand creates the list formats command
func NewListFormatsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "formats",
		Short: "List supported conversion formats",
		Long:  "Display all supported input and output formats",
		RunE:  runListFormats,
	}
}

func runListFormats(cmd *cobra.Command, args []string) error {
	// Implementation will connect to domain core
	fmt.Println("Supported formats:")
	return nil
}

