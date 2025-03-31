package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	client "github.com/setavenger/blindbit-wallet-cli/internal/client"
	"github.com/setavenger/blindbit-wallet-cli/internal/config"
	scanclient "github.com/setavenger/blindbit-wallet-cli/pkg/clients/blindbitscan"
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

		// Create Tor client if enabled
		var torClient *client.TorClient
		if viper.GetBool("use_tor") {
			torClient, err = client.NewTorClient(&config.Config{
				UseTor:     viper.GetBool("use_tor"),
				TorHost:    viper.GetString("tor_host"),
				TorPort:    viper.GetInt("tor_port"),
				TorControl: viper.GetString("tor_control"),
			})
			if err != nil {
				return fmt.Errorf("failed to create Tor client: %w", err)
			}
			defer torClient.Close()
		}

		// Create blindbit-scan client
		host := viper.GetString("scan_host")
		port := viper.GetInt("scan_port")
		username := viper.GetString("scan_user")
		password := viper.GetString("scan_pass")

		// Use the host directly if it already includes a protocol
		baseURL := host
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			baseURL = fmt.Sprintf("http://%s:%d", host, port)
		}
		client := scanclient.NewClient(baseURL, username, password, torClient)

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
			derivedPubKey, err := wallet.DerivePublicKey(w.SpendSecret, u.PrivKeyTweak)
			if err != nil {
				fmt.Printf("Warning: Skipping UTXO %s (vout %d) - ownership verification failed: %v\n",
					hex.EncodeToString(u.Txid[:]), u.Vout, err)
				continue
			}

			// Compare derived public key with UTXO's public key (X-only comparison)
			derivedPubKeyBytes := derivedPubKey.SerializeCompressed()
			if !bytes.Equal(derivedPubKeyBytes[1:], u.PubKey[:]) {
				fmt.Printf("Warning: Skipping UTXO %s (vout %d) - public key mismatch\n",
					hex.EncodeToString(u.Txid[:]), u.Vout)
				fmt.Printf("Derived pubkey: %x\n", derivedPubKeyBytes[1:])
				fmt.Printf("UTXO pubkey: %x\n", u.PubKey[:])
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
