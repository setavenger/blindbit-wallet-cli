package wallet

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/tyler-smith/go-bip39"
)

// expandPath expands the path to include the home directory if the path
// is prefixed with '~'. It also handles environment variables.
func expandPath(path string) string {
	if len(path) == 0 {
		return path
	}

	if path[0] == '~' {
		home, err := user.Current()
		if err != nil {
			return path
		}
		path = strings.Replace(path, "~", home.HomeDir, 1)
	}

	return os.ExpandEnv(path)
}

// New creates a new wallet with a random seed phrase and stores it in the datadir
func New(datadir string) (*Wallet, error) {
	expandedDatadir := expandPath(datadir)

	// Generate a random 32-byte seed
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	// Generate a new mnemonic
	mnemonic, err := bip39.NewMnemonic(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	// Generate spend and scan keys
	spendKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate spend key: %w", err)
	}

	scanKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate scan key: %w", err)
	}

	// Create wallet instance
	w := &Wallet{
		Network:     NetworkMainnet,
		Mnemonic:    mnemonic,
		ScanSecret:  scanKey.Serialize(),
		SpendSecret: spendKey.Serialize(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create wallet data with empty UTXOs and labels
	data := &WalletData{
		Wallet:     *w,
		UTXOs:      []UTXO{},
		Labels:     []Label{},
		LastHeight: 0,
	}

	// Ensure datadir exists
	if err := os.MkdirAll(expandedDatadir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create datadir: %w", err)
	}

	// Store wallet data
	walletFile := filepath.Join(expandedDatadir, "wallet.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet data: %w", err)
	}

	if err := os.WriteFile(walletFile, jsonData, 0600); err != nil {
		return nil, fmt.Errorf("failed to write wallet file: %w", err)
	}

	return w, nil
}

// Import creates a wallet from an existing mnemonic and stores it in the datadir
func Import(datadir string, mnemonic string) (*Wallet, error) {
	expandedDatadir := expandPath(datadir)

	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Use first 32 bytes of seed for spend key, next 32 bytes for scan key
	if len(seed) < 64 {
		return nil, fmt.Errorf("seed too short")
	}

	// Generate spend and scan keys from seed
	spendKey, _ := btcec.PrivKeyFromBytes(seed[:32])
	scanKey, _ := btcec.PrivKeyFromBytes(seed[32:64])

	// Create wallet instance
	w := &Wallet{
		Network:     NetworkMainnet,
		Mnemonic:    mnemonic,
		ScanSecret:  scanKey.Serialize(),
		SpendSecret: spendKey.Serialize(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create wallet data with empty UTXOs and labels
	data := &WalletData{
		Wallet:     *w,
		UTXOs:      []UTXO{},
		Labels:     []Label{},
		LastHeight: 0,
	}

	// Ensure datadir exists
	if err := os.MkdirAll(expandedDatadir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create datadir: %w", err)
	}

	// Store wallet data
	walletFile := filepath.Join(expandedDatadir, "wallet.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet data: %w", err)
	}

	if err := os.WriteFile(walletFile, jsonData, 0600); err != nil {
		return nil, fmt.Errorf("failed to write wallet file: %w", err)
	}

	return w, nil
}

// Load loads a wallet from the datadir
func Load(datadir string) (*Wallet, error) {
	expandedDatadir := expandPath(datadir)
	walletFile := filepath.Join(expandedDatadir, "wallet.json")
	data, err := os.ReadFile(walletFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	var walletData WalletData
	if err := json.Unmarshal(data, &walletData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet data: %w", err)
	}

	return &walletData.Wallet, nil
}

// Save saves wallet data to the datadir
func Save(datadir string, data *WalletData) error {
	expandedDatadir := expandPath(datadir)

	// Ensure datadir exists
	if err := os.MkdirAll(expandedDatadir, 0700); err != nil {
		return fmt.Errorf("failed to create datadir: %w", err)
	}

	// Store wallet data
	walletFile := filepath.Join(expandedDatadir, "wallet.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallet data: %w", err)
	}

	if err := os.WriteFile(walletFile, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	return nil
}

// LoadData loads the complete wallet data from the datadir
func LoadData(datadir string) (*WalletData, error) {
	expandedDatadir := expandPath(datadir)
	walletFile := filepath.Join(expandedDatadir, "wallet.json")
	data, err := os.ReadFile(walletFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	var walletData WalletData
	if err := json.Unmarshal(data, &walletData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet data: %w", err)
	}

	return &walletData, nil
}
