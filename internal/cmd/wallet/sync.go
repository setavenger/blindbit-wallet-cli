package wallet

import (
	"encoding/hex"
	"fmt"
	"time"

	client "github.com/setavenger/blindbit-wallet-cli/pkg/clients/blindbitscan"
	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync with blindbit-scan",
	Long:  `Fetch UTXOs and labels from blindbit-scan and update the local wallet data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		datadir := viper.GetString("datadir")
		w, err := wallet.Load(datadir)
		if err != nil {
			return fmt.Errorf("failed to load wallet: %w", err)
		}

		// Create blindbit-scan client
		host := viper.GetString("scan_host")
		port := viper.GetInt("scan_port")
		username := viper.GetString("scan_user")
		password := viper.GetString("scan_pass")

		baseURL := fmt.Sprintf("http://%s:%d", host, port)
		client := client.NewClient(baseURL, username, password)

		// Get current height
		height, err := client.GetCurrentHeight()
		if err != nil {
			return fmt.Errorf("failed to get current height: %w", err)
		}

		// Get UTXOs
		scanUtxos, err := client.GetUTXOs(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to get UTXOs: %w", err)
		}

		// Convert UTXOs to our format
		utxos := make([]wallet.UTXO, len(scanUtxos))
		for i, u := range scanUtxos {
			utxos[i] = wallet.UTXO{
				TxID:         hex.EncodeToString(u.Txid[:]),
				Vout:         u.Vout,
				Amount:       int64(u.Amount),
				ScriptPubKey: u.PubKey,
				Height:       int64(u.Timestamp),
				Spent:        u.State == wallet.StateSpent || u.State == wallet.StateUnconfirmedSpent,
				UpdatedAt:    time.Now(),
			}

			// Add label if present
			if u.Label != nil {
				utxos[i].Label = &wallet.Label{
					PubKey:  u.Label.PubKey,
					Tweak:   u.Label.Tweak,
					Address: u.Label.Address,
					M:       u.Label.M,
				}
			}
		}

		// Update wallet data
		data := &wallet.WalletData{
			Wallet:     *w,
			UTXOs:      utxos,
			LastHeight: int64(height),
		}

		// Save updated wallet data
		if err := wallet.Save(datadir, data); err != nil {
			return fmt.Errorf("failed to save wallet data: %w", err)
		}

		fmt.Println("Wallet synced successfully!")
		fmt.Printf("Current height: %d\n", height)
		fmt.Printf("Found %d UTXOs\n", len(utxos))

		return nil
	},
}
