package wallet

import (
	"encoding/hex"
	"fmt"

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

		// Convert UTXOs to our format and verify ownership
		utxos := make([]wallet.UTXO, 0, len(scanUtxos))
		for _, u := range scanUtxos {
			// Verify UTXO ownership by checking if we can derive the public key
			derivedPubKey, err := wallet.DerivePublicKey(w.ScanSecret, u.PrivKeyTweak)
			if err != nil {
				fmt.Printf("Warning: Skipping UTXO %s (vout %d) - ownership verification failed: %v\n",
					hex.EncodeToString(u.Txid[:]), u.Vout, err)
				continue
			}

			// Compare derived public key with UTXO's public key
			if !derivedPubKey.IsEqual(u.PubKey) {
				fmt.Printf("Warning: Skipping UTXO %s (vout %d) - public key mismatch\n",
					hex.EncodeToString(u.Txid[:]), u.Vout)
				continue
			}

			utxos = append(utxos, wallet.UTXO(*u))
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
		if len(utxos) != len(scanUtxos) {
			fmt.Printf("Warning: %d UTXOs were skipped due to ownership verification failures\n",
				len(scanUtxos)-len(utxos))
		}

		return nil
	},
}
