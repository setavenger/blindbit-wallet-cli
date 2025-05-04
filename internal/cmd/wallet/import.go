package wallet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a wallet from a mnemonic",
	Long:  `Import an existing wallet using its mnemonic (seed phrase). The mnemonic will be read securely from stdin. The wallet will be stored in the configured datadir.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		datadir := viper.GetString("datadir")
		walletPath := filepath.Join(datadir, "wallet.json")

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

		fmt.Print("Enter your mnemonic (seed phrase): ")

		// Read password securely
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read mnemonic: %w", err)
		}

		// Convert to string and trim whitespace
		mnemonic := strings.TrimSpace(string(bytePassword))
		fmt.Println() // Add newline after password input

		network := wallet.Network(viper.GetString("network"))
		if cmd.Flags().Changed("network") {
			network = wallet.Network(cmd.Flag("network").Value.String())
		}

		w, err := wallet.Import(datadir, mnemonic, network)
		if err != nil {
			return fmt.Errorf("failed to import wallet: %w", err)
		}

		fmt.Println("\nWallet imported successfully!")
		fmt.Println("Network:", w.Network)
		fmt.Println("Created at:", w.CreatedAt)
		fmt.Printf("Wallet stored in: %s\n", datadir)

		return nil
	},
}

func init() {
	// Add network flag
	importCmd.Flags().String("network", "mainnet", "Network to use (mainnet, testnet, signet, regtest)")
}
