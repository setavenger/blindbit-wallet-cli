package wallet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/setavenger/blindbit-wallet-cli/internal/config"
	client "github.com/setavenger/blindbit-wallet-cli/pkg/clients/blindbitscan"
	"github.com/setavenger/blindbit-wallet-cli/pkg/utils"
	"github.com/spf13/cobra"
)

func newUtxosCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "utxos",
		Short: "Pull and display the utxos",
		Run: func(cmd *cobra.Command, args []string) {
			conf := config.Global
			if conf == nil {
				fmt.Println("Configuration not loaded")
				os.Exit(1)
			}

			fullUrl := utils.ConstructBaseUrl(conf.ScanHost, conf.ScanPort)
			cl := client.NewClient(fullUrl, conf.ScanUser, conf.ScanPass)

			utxos, err := cl.GetUtxos()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			byteData, err := json.Marshal(utxos)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println(string(byteData))
		},
	}
}
