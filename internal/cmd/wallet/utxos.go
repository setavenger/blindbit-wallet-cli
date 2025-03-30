package wallet

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var utxosCmd = &cobra.Command{
	Use:   "utxos",
	Short: "List UTXOs",
	Long:  `List all UTXOs in the wallet. Use --state to filter by state (unconfirmed, confirmed, spent).`,
	RunE:  runUtxos,
}

func init() {
	utxosCmd.Flags().String("state", "", "Filter by state (unconfirmed, confirmed, spent)")
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
			if utxo.Spent {
				filteredUtxos = append(filteredUtxos, utxo)
			}
		case "confirmed":
			if !utxo.Spent && utxo.Height > 0 {
				filteredUtxos = append(filteredUtxos, utxo)
			}
		case "unconfirmed":
			if !utxo.Spent && utxo.Height == 0 {
				filteredUtxos = append(filteredUtxos, utxo)
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "TXID\tVOUT\tAMOUNT\tHEIGHT\tSTATE\tLABEL")
	for _, utxo := range filteredUtxos {
		state := "confirmed"
		if utxo.Spent {
			state = "spent"
		} else if utxo.Height == 0 {
			state = "unconfirmed"
		}

		label := ""
		if utxo.Label != nil {
			label = fmt.Sprintf("M=%d", utxo.Label.M)
		}

		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\t%s\n",
			utxo.TxID,
			utxo.Vout,
			utxo.Amount,
			utxo.Height,
			state,
			label)
	}

	return nil
}
