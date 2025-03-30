package config

import (
	"github.com/spf13/cobra"
)

// NewCommand returns the config command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Commands related to configuration",
		Run: func(cmd *cobra.Command, args []string) {
			// Default action when no subcommand is provided.
			cmd.Help()
		},
	}
	// Add subcommands for config if needed
	cmd.AddCommand(newShowCommand())
	return cmd
}
