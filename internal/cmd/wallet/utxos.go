package wallet

import (
	"encoding/hex"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	scanwallet "github.com/setavenger/blindbit-scan/pkg/wallet"
	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var utxosCmd = &cobra.Command{
	Use:   "utxos",
	Short: "List UTXOs",
	Long:  `List all UTXOs in the wallet. Use --state to filter by state (unconfirmed, unspent, spent, unconfirmed_spent).`,
	RunE:  runUtxos,
}

func init() {
	utxosCmd.Flags().String("state", "", "Filter by state (unconfirmed, unspent, spent, unconfirmed_spent)")
}

func runUtxos(cmd *cobra.Command, args []string) error {
	datadir := viper.GetString("datadir")
	walletData, err := wallet.LoadData(datadir)
	if err != nil {
		return fmt.Errorf("failed to load wallet data: %w", err)
	}

	state, _ := cmd.Flags().GetString("state")
	var filteredUtxos []wallet.UTXO
	for _, utxo := range walletData.UTXOs {
		if state == "" {
			filteredUtxos = append(filteredUtxos, utxo)
			continue
		}

		switch state {
		case "spent":
			if utxo.State == scanwallet.StateSpent {
				filteredUtxos = append(filteredUtxos, utxo)
			}
		case "unspent":
			if utxo.State == scanwallet.StateUnspent {
				filteredUtxos = append(filteredUtxos, utxo)
			}
		case "unconfirmed":
			if utxo.State == scanwallet.StateUnconfirmed {
				filteredUtxos = append(filteredUtxos, utxo)
			}
		case "unconfirmed_spent":
			if utxo.State == scanwallet.StateUnconfirmedSpent {
				filteredUtxos = append(filteredUtxos, utxo)
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "TXID\tVOUT\tAMOUNT\tTIMESTAMP\tSTATE\tLABEL")
	for _, utxo := range filteredUtxos {
		state := "unspent"
		switch utxo.State {
		case scanwallet.StateSpent:
			state = "spent"
		case scanwallet.StateUnconfirmed:
			state = "unconfirmed"
		case scanwallet.StateUnconfirmedSpent:
			state = "unconfirmed_spent"
		}

		label := ""
		if utxo.Label != nil {
			label = fmt.Sprintf("M=%d", utxo.Label.M)
		}

		// Convert Unix timestamp to human-readable time
		timestamp := time.Unix(int64(utxo.Timestamp), 0).Format("2006-01-02 15:04:05")

		fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\t%s\n",
			hex.EncodeToString(utxo.Txid[:]),
			utxo.Vout,
			utxo.Amount,
			timestamp,
			state,
			label)
	}

	return nil
}
