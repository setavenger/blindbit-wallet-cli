package wallet

import (
	"fmt"
	"os"
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
		fmt.Print("Enter your mnemonic (seed phrase): ")

		// Read password securely
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read mnemonic: %w", err)
		}

		// Convert to string and trim whitespace
		mnemonic := strings.TrimSpace(string(bytePassword))
		fmt.Println() // Add newline after password input

		datadir := viper.GetString("datadir")

		w, err := wallet.Import(datadir, mnemonic)
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
