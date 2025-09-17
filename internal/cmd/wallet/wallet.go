package wallet

import "github.com/spf13/cobra"

// WalletCmd is the root command for wallet operations
var WalletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Commands related to wallet operations",
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided.
		cmd.Help()
	},
}

// NewCommand returns the wallet command.
func NewCommand() *cobra.Command {
	// Add subcommands for wallet operations
	WalletCmd.AddCommand(newCmd)
	WalletCmd.AddCommand(importCmd)
	WalletCmd.AddCommand(infoCmd)
	WalletCmd.AddCommand(syncCmd)
	WalletCmd.AddCommand(utxosCmd)
	WalletCmd.AddCommand(addressCmd)
	WalletCmd.AddCommand(NewSendCmd())
	WalletCmd.AddCommand(NewScanCmd())

	return WalletCmd
}
