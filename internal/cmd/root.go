package cmd

import (
	"fmt"
	"os"
	"path"

	configcmd "github.com/setavenger/blindbit-wallet-cli/internal/cmd/config"
	walletcmd "github.com/setavenger/blindbit-wallet-cli/internal/cmd/wallet"
	"github.com/setavenger/blindbit-wallet-cli/internal/config"
	"github.com/setavenger/blindbit-wallet-cli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	socketPath string
	RootCmd    = &cobra.Command{
		Use:   "blindbit-wallet-cli",
		Short: "A blindbit wallet cli application",
		Long: `BlindBit Wallet Cli is a CLI application to manage a Bitcoin Silent Payment (BIP 352) wallet:

    The cli allows the user to spend coins and manage the wallet. It does NOT scan the chain. A separate deamon like BlindBit Scan is needed to find new coins. 
`,
	}
)

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(configcmd.NewCommand())
	RootCmd.AddCommand(walletcmd.NewCommand())
}

const defaultConfigFileName = "blindbit.toml"

var (
	cfgFile string

	defaultDataDir string = utils.ResolvePath("~/.blindbit-wallet")
)

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", path.Join(defaultDataDir, defaultConfigFileName), "config file (default is $HOME/.blindbit-wallet-config.toml)")

	// Define command flags
	RootCmd.PersistentFlags().String("datadir", defaultDataDir, "datadir default ($HOME/.blindbit-wallet)")
	RootCmd.PersistentFlags().String("scan-host", "localhost", "server host")
	RootCmd.PersistentFlags().Int("scan-port", 8080, "server port")
	RootCmd.PersistentFlags().String("scan-user", "", "blindbit scan user")
	RootCmd.PersistentFlags().String("scan-pass", "", "blindbit scan password for user")

	// Bind the flags to Viper configuration keys
	viper.BindPFlag("datadir", RootCmd.PersistentFlags().Lookup("datadir"))
	viper.BindPFlag("scan_host", RootCmd.PersistentFlags().Lookup("scan_host"))
	viper.BindPFlag("scan_port", RootCmd.PersistentFlags().Lookup("scan_port"))
	viper.BindPFlag("scan_user", RootCmd.PersistentFlags().Lookup("scan_user"))
	viper.BindPFlag("scan_pass", RootCmd.PersistentFlags().Lookup("scan_pass"))
}

func initConfig() {
	// If the user did NOT explicitly set --config,
	// use the datadir flag to build the config file path.
	if !RootCmd.PersistentFlags().Changed("config") {
		// Get the datadir value from Viper (set via the --datadir flag)
		datadir := viper.GetString("datadir")
		// Build the config file path from the datadir
		cfgFile = path.Join(datadir, defaultConfigFileName)
	}

	// Tell Viper to use this config file
	viper.SetConfigFile(cfgFile)

	// Allow environment variables to override values
	viper.AutomaticEnv()

	// Attempt to read the config file (if it exists)
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Unmarshal the configuration into our global config variable
	if err := viper.Unmarshal(&config.Global); err != nil {
		fmt.Printf("Unable to decode config into struct: %v\n", err)
	}
}
