package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `Manage the configuration file for the wallet CLI.`,
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Long:  `Create a new configuration file with default values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			datadir := viper.GetString("datadir")
			configFile := filepath.Join(datadir, "blindbit.toml")

			// Create default config
			defaultConfig := `# BlindBit Wallet CLI Configuration

# Data directory for wallet files
datadir = "~/.blindbit-wallet"

# Network configuration (mainnet, testnet, signet, regtest)
network = "mainnet"

# BlindBit Scan connection details
scan_host = "localhost"
scan_port = 8080
scan_user = ""
scan_pass = ""
`

			// Ensure datadir exists
			if err := os.MkdirAll(datadir, 0700); err != nil {
				return fmt.Errorf("failed to create datadir: %w", err)
			}

			// Write config file
			if err := os.WriteFile(configFile, []byte(defaultConfig), 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("Configuration file created at: %s\n", configFile)
			fmt.Println("Please edit the file to set your BlindBit Scan credentials.")

			return nil
		},
	}

	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Display the current configuration values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Current Configuration:")
			fmt.Println("---------------------")
			fmt.Printf("Data Directory: %s\n", viper.GetString("datadir"))
			fmt.Printf("Scan Host: %s\n", viper.GetString("scan_host"))
			fmt.Printf("Scan Port: %d\n", viper.GetInt("scan_port"))
			fmt.Printf("Scan User: %s\n", viper.GetString("scan_user"))
			fmt.Printf("Scan Pass: %s\n", viper.GetString("scan_pass"))
			return nil
		},
	}
)

func init() {
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(showCmd)
}

// NewCommand returns the config command.
func NewCommand() *cobra.Command {
	return configCmd
}
