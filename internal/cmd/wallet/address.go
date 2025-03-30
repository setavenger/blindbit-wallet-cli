package wallet

import (
	"fmt"

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
		showChange, _ := cmd.Flags().GetBool("change")

		var address string
		if showChange {
			var label bip352.Label
			// Create labeled address (user labels start from 1)
			label, err = wallet.GenerateLabel(*w, 0)
			if err != nil {
				return fmt.Errorf("failed to compute label: %w", err)
			}
			address = label.Address
		} else if labelNum > 0 {
			var label bip352.Label
			// Create labeled address (user labels start from 1)
			label, err = wallet.GenerateLabel(*w, labelNum)
			if err != nil {
				return fmt.Errorf("failed to compute label: %w", err)
			}
			address = label.Address
		} else {

			// Create base address (no label)
			address, err = bip352.CreateAddress(
				w.PubKeyScan(),
				w.PubKeySpend(),
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
	addressCmd.Flags().Bool("change", false, "show change address, overrides label to 0 internally")
}
