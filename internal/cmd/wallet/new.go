package wallet

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new wallet",
	Long:  `Generate a new wallet with a random mnemonic phrase.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		datadir := viper.GetString("datadir")
		walletPath := filepath.Join(datadir, "wallet.json")

		// Create datadir if it doesn't exist
		if err := os.MkdirAll(datadir, 0700); err != nil {
			return fmt.Errorf("failed to create datadir: %w", err)
		}

		// Check if wallet already exists
		_, err := os.Stat(walletPath)
		if err == nil {
			fmt.Println("Warning: A wallet already exists at:", walletPath)
			fmt.Println("Creating a new wallet will overwrite the existing one.")

			// Ask for confirmation
			fmt.Print("Do you want to continue? (y/N): ")
			var response string
			fmt.Scanln(&response)

			if response != "y" && response != "Y" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check wallet file: %w", err)
		}

		// Create new wallet
		network := wallet.Network(viper.GetString("network"))
		if cmd.Flags().Changed("network") {
			network = wallet.Network(cmd.Flag("network").Value.String())
		}
		w, err := wallet.New(datadir, network)
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}

		fmt.Println("Wallet created successfully!")
		fmt.Printf("Network: %s\n", w.Network)
		fmt.Printf("Created at: %s\n", w.CreatedAt)
		fmt.Printf("Wallet stored in: %s\n", datadir)
		fmt.Println("\nIMPORTANT: Save your mnemonic phrase securely!")
		fmt.Printf("Mnemonic: %s\n", w.Mnemonic)

		return nil
	},
}

func init() {
	// Add network flag
	newCmd.Flags().String("network", "mainnet", "Network to use (mainnet, testnet, signet, regtest)")
}
