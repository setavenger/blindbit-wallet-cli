package coinselector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoinSelector(t *testing.T) {
	// Create test UTXOs
	utxos := UtxoCollection{
		{
			Amount: 1000000, // 0.01 BTC
			Txid:   [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			Vout:   0,
		},
		{
			Amount: 2000000, // 0.02 BTC
			Txid:   [32]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33},
			Vout:   1,
		},
	}

	// Create test recipients
	recipients := []Recipient{
		{
			Address: "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
			Amount:  500000, // 0.005 BTC
		},
		{
			Address: "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
			Amount:  1000000, // 0.01 BTC
		},
	}

	// Test with different fee rates
	feeRates := []uint64{1, 10, 100, 1000}

	for _, feeRate := range feeRates {
		selected, change, err := SelectCoins(utxos, recipients, feeRate)
		assert.NoError(t, err)
		assert.NotNil(t, selected)
		assert.NotNil(t, change)

		// Verify total amount covers recipients and fees
		totalSelected := uint64(0)
		for _, utxo := range selected {
			totalSelected += utxo.Amount
		}

		totalRecipients := uint64(0)
		for _, recipient := range recipients {
			totalRecipients += recipient.Amount
		}

		// Calculate estimated fee (simplified)
		estimatedFee := uint64(len(selected)) * feeRate * 225    // Approximate vbytes per input
		estimatedFee += uint64(len(recipients)+1) * feeRate * 34 // Approximate vbytes per output

		assert.GreaterOrEqual(t, totalSelected, totalRecipients+estimatedFee)
	}
}

func TestInsufficientFunds(t *testing.T) {
	utxos := UtxoCollection{
		{
			Amount: 100000, // 0.001 BTC
			Txid:   [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			Vout:   0,
		},
	}

	recipients := []Recipient{
		{
			Address: "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
			Amount:  1000000, // 0.01 BTC
		},
	}

	_, _, err := SelectCoins(utxos, recipients, 1)
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientFunds, err)
}

func TestInvalidFeeRate(t *testing.T) {
	utxos := UtxoCollection{
		{
			Amount: 1000000,
			Txid:   [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			Vout:   0,
		},
	}

	recipients := []Recipient{
		{
			Address: "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
			Amount:  500000,
		},
	}

	_, _, err := SelectCoins(utxos, recipients, 0)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFeeRate, err)
}
