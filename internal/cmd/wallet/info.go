package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show wallet information",
	Long:  `Display wallet information including network, scan secret, and spend public key.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		datadir := viper.GetString("datadir")
		w, err := wallet.Load(datadir)
		if err != nil {
			return fmt.Errorf("failed to load wallet: %w", err)
		}

		pubKey := w.PubKeySpend()

		fmt.Println("Wallet Information:")
		fmt.Println("-------------------")
		fmt.Println("Network:", w.Network)
		fmt.Println("Created at:", w.CreatedAt)
		fmt.Println("Scan Secret:", hex.EncodeToString(w.ScanSecret))
		fmt.Println("Spend Public:", hex.EncodeToString(pubKey[:]))

		return nil
	},
}
