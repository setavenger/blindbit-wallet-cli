package coinselector

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
)

// UTXO represents a Bitcoin UTXO

// Recipient represents a transaction recipient
type Recipient struct {
	Address  string
	Amount   uint64
	PkScript []byte
}

// Errors
var (
	ErrInvalidFeeRate        = fmt.Errorf("invalid fee rate")
	ErrRecipientAmountIsZero = fmt.Errorf("recipient amount is zero")
	ErrInsufficientFunds     = fmt.Errorf("insufficient funds")
)

// ChainParams is the Bitcoin network parameters
var ChainParams = &chaincfg.MainNetParams
