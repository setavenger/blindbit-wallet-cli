package wallet

import (
	"fmt"
	"strconv"

	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewSendCmd() *cobra.Command {
	var (
		feeRate int32
	)

	cmd := &cobra.Command{
		Use:   "send [address] [amount]",
		Short: "Send Bitcoin to an address",
		Long: `Send Bitcoin to an address. The amount should be in satoshis.
The command supports both regular Bitcoin addresses and silent payment addresses.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if feeRate < 0 {
				return fmt.Errorf("please set a fee rate")
			}
			address := args[0]
			amount, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid amount: %s", args[1])
			}

			// Load wallet data
			datadir := viper.GetString("datadir")
			walletData, err := wallet.LoadData(datadir)
			if err != nil {
				return fmt.Errorf("failed to load wallet: %w", err)
			}

			// Create recipient
			recipient := wallet.Recipient{
				Address: address,
				Amount:  amount,
			}

			// Send to recipient
			txBytes, err := wallet.SendToRecipients(walletData, []wallet.Recipient{recipient}, uint32(feeRate))
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}

			// Print the signed transaction
			fmt.Printf("Signed transaction: %x\n", txBytes)
			return nil
		},
	}

	cmd.Flags().Int32Var(&feeRate, "fee-rate", -1, "Fee rate in sat/vB")

	return cmd
}
