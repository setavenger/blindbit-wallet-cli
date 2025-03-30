package wallet

import (
	"fmt"

	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new wallet",
	Long:  `Create a new wallet with a new seed phrase. The seed phrase will be displayed and should be backed up securely.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		datadir := viper.GetString("datadir")
		w, err := wallet.New(datadir)
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}

		// Display the seed phrase
		fmt.Println("Your seed phrase (BACK THIS UP SECURELY):")
		fmt.Println(w.Mnemonic)
		fmt.Println("\nNetwork:", w.Network)
		fmt.Println("\nCreated at:", w.CreatedAt)
		fmt.Printf("\nWallet stored in: %s\n", datadir)

		return nil
	},
}
