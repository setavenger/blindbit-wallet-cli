package wallet

import (
	"encoding/json"
	"time"
)

// Network represents the Bitcoin network type
type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkTestnet Network = "testnet"
	NetworkSignet  Network = "signet"
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

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID         string     `json:"txid"`
	Vout         uint32     `json:"vout"`
	Amount       int64      `json:"amount"`
	ScriptPubKey string     `json:"script_pub_key"`
	Label        *Label     `json:"label,omitempty"`
	Height       int64      `json:"height"`
	Spent        bool       `json:"spent"`
	SpentAt      *time.Time `json:"spent_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Label represents a labeled address
type Label struct {
	PubKey  string `json:"pub_key"`
	Tweak   string `json:"tweak"`
	Address string `json:"address"`
	M       uint32 `json:"m"`
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

// OwnedUTXO represents a UTXO owned by the wallet
type OwnedUTXO struct {
	Txid      [32]byte  `json:"txid"`
	Vout      uint32    `json:"vout"`
	Amount    uint64    `json:"amount"`
	PubKey    string    `json:"pub_key"`
	Timestamp uint64    `json:"timestamp"`
	State     UTXOState `json:"state"`
	Label     *Label    `json:"label,omitempty"`
}

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
		ScanSecret:  string(w.ScanSecret),
		SpendSecret: string(w.SpendSecret),
	})
}

// UnmarshalJSON implements json.Unmarshaler for Wallet
func (w *Wallet) UnmarshalJSON(data []byte) error {
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
	w.ScanSecret = []byte(aux.ScanSecret)
	w.SpendSecret = []byte(aux.SpendSecret)
	return nil
}
