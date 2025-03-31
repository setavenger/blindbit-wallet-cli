package config

// Config holds all configuration details.
type Config struct {
	DataDir string `mapstructure:"datadir"`

	ScanHost string `mapstructure:"scan_host"`
	ScanPort int    `mapstructure:"scan_port"`
	ScanUser string `mapstructure:"scan_user"`
	ScanPass string `mapstructure:"scan_pass"`

	// Tor configuration
	UseTor     bool   `mapstructure:"use_tor"`
	TorHost    string `mapstructure:"tor_host"`
	TorPort    int    `mapstructure:"tor_port"`
	TorControl string `mapstructure:"tor_control"`
}

var Global *Config
