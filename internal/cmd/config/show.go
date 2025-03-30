package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/setavenger/blindbit-wallet-cli/internal/config"
	"github.com/spf13/cobra"
)

func newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display the current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			conf := config.Global
			if conf == nil {
				fmt.Println("Configuration not loaded")
				os.Exit(1)
			}

			// Pretty-print the configuration as JSON.
			data, err := json.MarshalIndent(conf, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting config: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(data))
		},
	}
}
