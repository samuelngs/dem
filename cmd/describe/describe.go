// Package describe implements the `describe` command
package describe

import (
	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
	return cmd.Usage()
}

// NewCommand returns a new cobra.Command for cluster creation
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of a specific workspace",
		Long:  "Show details of a specific workspace",
		RunE:  run,
	}
	return cmd
}
