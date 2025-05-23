package wallet

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewSendCmd() *cobra.Command {
	var (
		feeRate int32
	)

	cmd := &cobra.Command{
		Use:   "send <address:amount>...",
		Short: "Send Bitcoin to one or more addresses",
		Long: `Send Bitcoin to one or more addresses. Each recipient should be specified in the format address:amount.
The amount should be in satoshis. The command supports both regular Bitcoin addresses and silent payment addresses.

Examples:
  blindbit-wallet-cli wallet send bc1q...:1000000
  blindbit-wallet-cli wallet send bc1q...:1000000 sp1q...:2000000 --fee-rate 5`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if feeRate < 0 {
				return fmt.Errorf("please set a fee rate")
			}

			// Load wallet data
			datadir := viper.GetString("datadir")
			walletData, err := wallet.LoadData(datadir)
			if err != nil {
				return fmt.Errorf("failed to load wallet: %w", err)
			}

			// Get network from flag if specified, otherwise use config file value
			if cmd.Flags().Changed("network") {
				walletData.Wallet.Network = wallet.Network(cmd.Flag("network").Value.String())
			} else {
				// Use network from config file
				configNetwork := viper.GetString("network")
				if configNetwork != "" {
					walletData.Wallet.Network = wallet.Network(configNetwork)
				}
			}

			var recipients []wallet.Recipient
			for _, arg := range args {
				rec, err := extractRecipientFromPositionalArg(arg)
				if err != nil {
					return fmt.Errorf("failed to extract recipient: %w", err)
				}
				recipients = append(recipients, rec)
			}

			// Send to recipient
			txBytes, err := wallet.SendToRecipients(
				walletData,
				recipients,
				uint32(feeRate),
			)
			if err != nil {
				return fmt.Errorf("failed to send: %w", err)
			}

			// Print the signed transaction
			fmt.Printf("Signed transaction: %x\n", txBytes)
			return nil
		},
	}

	cmd.Flags().Int32Var(&feeRate, "fee-rate", -1, "Fee rate in sat/vB")
	cmd.Flags().String("network", "", "Network to use (mainnet, testnet, signet, regtest)")

	return cmd
}

func extractRecipientFromPositionalArg(s string) (*wallet.RecipientImpl, error) {
	components := strings.Split(s, ":")
	if len(components) != 2 {
		return nil, fmt.Errorf("bad recipient arg %s", s)
	}
	addr, amt := components[0], components[1]

	amount, err := strconv.ParseUint(amt, 10, 64)
	if err != nil {
		return nil, err
	}

	return &wallet.RecipientImpl{
		Address: addr,
		Amount:  amount,
	}, nil
}
