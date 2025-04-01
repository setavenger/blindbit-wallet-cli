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
	cfgFile string
	cfg     *config.Config
)

const (
	defaultConfigFileName = "blindbit.toml"
)

var (
	defaultDataDir string = utils.ResolvePath("~/.blindbit-wallet")
)

var RootCmd = &cobra.Command{
	Use:   "blindbit-wallet-cli",
	Short: "BlindBit Wallet Cli is a CLI application to manage a Bitcoin Silent Payment (BIP 352) wallet",
	Long: `BlindBit Wallet Cli is a CLI application to manage a Bitcoin Silent Payment (BIP 352) wallet:
    The cli allows the user to spend coins and manage the wallet. It does NOT scan the chain. A separate deamon like BlindBit Scan is needed to find new coins.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// use the datadir flag to build the config file path.
		// Get the datadir value from Viper (set via the --datadir flag)
		datadir := viper.GetString("datadir")
		// Expand the datadir path
		expandedDatadir := utils.ResolvePath(datadir)
		// Set the expanded path back in viper
		viper.Set("datadir", expandedDatadir)

		// Build the config file path from the datadir
		cfgFile = path.Join(expandedDatadir, defaultConfigFileName)

		// Load config
		if err := loadConfig(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return nil
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set default config file path
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", path.Join(defaultDataDir, defaultConfigFileName), "config file (default is $HOME/.blindbit-wallet-config.toml)")

	// Set default datadir
	RootCmd.PersistentFlags().String("datadir", defaultDataDir, "datadir default ($HOME/.blindbit-wallet)")

	// Bind flags to viper
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("datadir", RootCmd.PersistentFlags().Lookup("datadir"))

	RootCmd.AddCommand(configcmd.NewCommand())
	RootCmd.AddCommand(walletcmd.NewCommand())
}

func initConfig() {
	// Set default values
	viper.SetDefault("scan_host", "localhost")
	viper.SetDefault("scan_port", 8080)
	viper.SetDefault("scan_user", "")
	viper.SetDefault("scan_pass", "")

	// Tor configuration defaults
	viper.SetDefault("use_tor", false)
	viper.SetDefault("tor_host", "localhost")
	viper.SetDefault("tor_port", 9050)
	viper.SetDefault("tor_control", "")
}

func loadConfig() error {
	// Tell Viper to use this config file
	viper.SetConfigFile(cfgFile)

	// Allow environment variables to override values
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err == nil {
		// todo: leave this for verbose logging
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Unmarshal config into struct
	cfg = &config.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
