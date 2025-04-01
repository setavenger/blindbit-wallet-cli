package wallet

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	scanwallet "github.com/setavenger/blindbit-scan/pkg/wallet"
	"github.com/setavenger/go-bip352"
)

// Network represents the Bitcoin network type
type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkTestnet Network = "testnet"
	NetworkSignet  Network = "signet"
	NetworkRegtest Network = "regtest"
)

// UTXOState represents the state of a UTXO
type UTXOState int8

const (
	StateUnknown UTXOState = iota - 1
	StateUnconfirmed
	StateUnspent
	StateSpent
	StateUnconfirmedSpent
)

// Wallet represents the core wallet data
type Wallet struct {
	Network     Network   `json:"network"`
	Mnemonic    string    `json:"mnemonic"`
	ScanSecret  []byte    `json:"scan_secret"`
	SpendSecret []byte    `json:"spend_secret"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WalletData represents the complete wallet data stored on disk
type WalletData struct {
	Wallet     Wallet  `json:"wallet"`
	UTXOs      []UTXO  `json:"utxos"`
	LastHeight int64   `json:"last_height"`
	Labels     []Label `json:"labels"`
}

// ScanOnlyParams represents the parameters needed for scan-only wallets
type ScanOnlyParams struct {
	ScanSecret []byte  `json:"scan_secret"`
	Network    Network `json:"network"`
}

// UTXO represents a UTXO in the wallet
type UTXO = scanwallet.OwnedUTXO

// Label represents a labeled address
type Label = bip352.Label

// NewWallet creates a new wallet with the given mnemonic and network
func NewWallet(mnemonic string, network Network) (*Wallet, error) {
	// TODO: Implement mnemonic to key derivation
	// This will be implemented in the next step
	return &Wallet{
		Network:   network,
		Mnemonic:  mnemonic,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// MarshalJSON implements json.Marshaler for Wallet
func (w *Wallet) MarshalJSON() ([]byte, error) {
	type Alias Wallet
	return json.Marshal(&struct {
		*Alias
		ScanSecret  string `json:"scan_secret"`
		SpendSecret string `json:"spend_secret"`
	}{
		Alias:       (*Alias)(w),
		ScanSecret:  hex.EncodeToString(w.ScanSecret),
		SpendSecret: hex.EncodeToString(w.SpendSecret),
	})
}

// UnmarshalJSON implements json.Unmarshaler for Wallet
func (w *Wallet) UnmarshalJSON(data []byte) (err error) {
	type Alias Wallet
	aux := &struct {
		*Alias
		ScanSecret  string `json:"scan_secret"`
		SpendSecret string `json:"spend_secret"`
	}{
		Alias: (*Alias)(w),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	w.ScanSecret, err = hex.DecodeString(aux.ScanSecret)
	if err != nil {
		return fmt.Errorf("failed to decode scan secret: %w", err)
	}

	w.SpendSecret, err = hex.DecodeString(aux.SpendSecret)
	if err != nil {
		return fmt.Errorf("failed to decode spend secret: %w", err)
	}

	return nil
}

func (w Wallet) PubKeyScan() [33]byte {
	_, scanPubKey := btcec.PrivKeyFromBytes(w.ScanSecret)
	return bip352.ConvertToFixedLength33(scanPubKey.SerializeCompressed())
}

func (w Wallet) PubKeySpend() [33]byte {
	_, spendPubKey := btcec.PrivKeyFromBytes(w.SpendSecret)
	return bip352.ConvertToFixedLength33(spendPubKey.SerializeCompressed())
}

func (w Wallet) ChangeAddress() string {
	label, err := GenerateLabel(w, 0)
	if err != nil {
		panic(err) // should never happen
	}
	return label.Address
}
