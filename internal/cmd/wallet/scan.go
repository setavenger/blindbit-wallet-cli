package wallet

import (
	"context"
	"fmt"

	"github.com/setavenger/blindbit-lib/networking/v2connect"
	"github.com/setavenger/blindbit-lib/scanning"
	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Do local scanning for transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load wallet data
			datadir := viper.GetString("datadir")
			walletData, err := wallet.LoadData(datadir)
			if err != nil {
				return fmt.Errorf("failed to load wallet: %w", err)
			}
			v2client, err := v2connect.NewClient(context.TODO(), "127.0.0.1:7001")
			if err != nil {
				return fmt.Errorf("failed to create v2 client: %w", err)
			}

			scanner := scanning.NewScannerV2(
				v2client,
				[32]byte(walletData.Wallet.ScanSecret),
				walletData.Wallet.PubKeySpend(),
				nil,
			)

			newUtxosChan := scanner.NewUtxosChan()
			go func() {
				for utxo := range newUtxosChan {
					if len(utxo) > 0 {
						fmt.Printf("New UTXO: %+v\n", utxo[0])
					}
				}
			}()

			err = scanner.ScanParallel(context.TODO(), 883000, 893000)
			if err != nil {
				return fmt.Errorf("failed to scan: %w", err)
			}
			return nil
		},
	}
}
