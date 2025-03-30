package wallet

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/setavenger/go-bip352"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addressCmd = &cobra.Command{
	Use:   "address",
	Short: "Generate a silent payment address",
	Long: `Generate a silent payment address for receiving payments.
The address can be labeled (M=1,2,3...) for different purposes.
Note: Label 0 is reserved for change addresses.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		datadir := viper.GetString("datadir")
		w, err := wallet.Load(datadir)
		if err != nil {
			return fmt.Errorf("failed to load wallet: %w", err)
		}

		// Get label number from flag
		labelNum, _ := cmd.Flags().GetUint32("label")

		// Convert scan secret to public key
		scanPrivKey, _ := btcec.PrivKeyFromBytes(w.ScanSecret)
		scanPubKey := scanPrivKey.PubKey()

		// Convert spend secret to public key
		spendPrivKey, _ := btcec.PrivKeyFromBytes(w.SpendSecret)
		spendPubKey := spendPrivKey.PubKey()

		// Convert public keys to fixed-size arrays
		var scanPubKeyBytes [33]byte
		var spendPubKeyBytes [33]byte
		copy(scanPubKeyBytes[:], scanPubKey.SerializeCompressed())
		copy(spendPubKeyBytes[:], spendPubKey.SerializeCompressed())

		var address string
		if labelNum > 0 {
			// Create labeled address (user labels start from 1)
			address, err = bip352.CreateLabeledAddress(
				scanPubKeyBytes,
				spendPubKeyBytes,
				w.Network == "mainnet",
				0,          // version 0
				[32]byte{}, // empty tweak (will be generated internally)
				labelNum,
			)
		} else {
			// Create base address (no label)
			address, err = bip352.CreateAddress(
				scanPubKeyBytes,
				spendPubKeyBytes,
				w.Network == "mainnet",
				0, // version 0
			)
		}

		if err != nil {
			return fmt.Errorf("failed to create address: %w", err)
		}

		fmt.Println("Silent Payment Address:")
		fmt.Println(address)
		if labelNum > 0 {
			fmt.Printf("Label: M=%d\n", labelNum)
		}

		return nil
	},
}

func init() {
	// Add label flag (minimum value 1)
	addressCmd.Flags().Uint32("label", 0, "Label number (M=1,2,3...) for the address")
}
