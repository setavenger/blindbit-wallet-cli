package wallet

import (
	"context"
	"fmt"

	"github.com/setavenger/blindbit-lib/logging"
	"github.com/setavenger/blindbit-lib/networking/v2connect"
	"github.com/setavenger/blindbit-lib/scanning/scannerv2"
	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/setavenger/go-bip352"
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

			changeLabel, err := bip352.CreateLabel((*[32]byte)(walletData.Wallet.ScanSecret), 0)
			if err != nil {
				return err
			}

			// var scanSec [32]byte
			// copy(scanSec[:], walletData.Wallet.ScanSecret)
			// changeLabel, err := bip352.CreateLabel(&scanSec, 0)
			// if err != nil {
			// 	return err
			// }

			scanner := scannerv2.NewScannerV2(
				v2client,
				[32]byte(walletData.Wallet.ScanSecret),
				walletData.Wallet.PubKeySpend(),
				[]*bip352.Label{&changeLabel},
			)

			// incompChan := scanner.NewIncompleteUTXOsChan()
			// go func() {
			// 	for utxo := range incompChan {
			// 		if len(utxo) > 0 {
			// 			logging.L.Info().Hex("txid", utxo[0].Txid[:]).Msg("incomplete UTXO")
			// 		}
			// 	}
			// 	// fmt.Println("newUtxosChan passed range", newUtxosChan)
			// }()

			newUtxosChan := scanner.NewOwnedUTXOsChan()
			go func() {
				for utxo := range newUtxosChan {
					logging.L.Info().
						Hex("txid", utxo.Txid[:]).
						Any("utxo", utxo).
						Msg("new UTXO")
				}
				// fmt.Println("newUtxosChan passed range", newUtxosChan)
			}()

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			// err = scanner.ScanParallelShortOutputs(context.TODO(), 892200, 892250)

			// go func() {
			// 	<-time.After(2 * time.Second)
			// 	err = scanner.Stop()
			// 	// cancel()
			// 	if err != nil {
			// 		panic(err)
			// 	}
			// }()

			// err = scanner.Watch(ctx)
			err = scanner.Scan(ctx, 892229, 892340)
			if err != nil {
				return fmt.Errorf("failed to scan: %w", err)
			}
			return nil
		},
	}
}
