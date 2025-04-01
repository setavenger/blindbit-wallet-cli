package wallet

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	Comment string
	Given   struct {
		Utxos           []*UTXO
		Recipients      []Recipient
		FeeRate         uint32
		MinChangeAmount uint64
	}
	Expected struct {
		Change             uint64
		NumOfSelectedUTXOs int
		AbsolutFee         uint64
		Err                error
	}
}

var testCases = []TestCase{
	{
		Comment: "Simple case",
		Given: struct {
			Utxos           []*UTXO
			Recipients      []Recipient
			FeeRate         uint32
			MinChangeAmount uint64
		}{
			Utxos: []*UTXO{
				{
					Amount: 20_000,
				}, {
					Amount: 40_000,
				}, {
					Amount: 60_000,
				},
			},
			Recipients: []Recipient{
				&RecipientImpl{
					Address: "bc1qua7e852suw0p74e2lzxwmk2tw8fd2zuzexc866",
					Amount:  5_000,
				},
			},
			FeeRate:         1,
			MinChangeAmount: 5000,
		},
		Expected: struct {
			Change             uint64
			NumOfSelectedUTXOs int
			AbsolutFee         uint64
			Err                error
		}{Change: 14858, NumOfSelectedUTXOs: 1, AbsolutFee: 142, Err: nil},
	},
	{
		Comment: "fits with change amount",
		Given: struct {
			Utxos           []*UTXO
			Recipients      []Recipient
			FeeRate         uint32
			MinChangeAmount uint64
		}{
			Utxos: []*UTXO{
				{
					Amount: 20_000,
				}, {
					Amount: 40_000,
				}, {
					Amount: 60_000,
				},
			},
			Recipients: []Recipient{
				&RecipientImpl{
					Address: "bc1qua7e852suw0p74e2lzxwmk2tw8fd2zuzexc866",
					Amount:  54_000,
				},
			},
			FeeRate:         1,
			MinChangeAmount: 5000,
		},
		Expected: struct {
			Change             uint64
			NumOfSelectedUTXOs int
			AbsolutFee         uint64
			Err                error
		}{Change: 5800, NumOfSelectedUTXOs: 2, AbsolutFee: 200, Err: nil},
	},
	{
		Comment: "Does it work with different fee rates",
		Given: struct {
			Utxos           []*UTXO
			Recipients      []Recipient
			FeeRate         uint32
			MinChangeAmount uint64
		}{
			Utxos: []*UTXO{
				{
					Amount: 20_000,
				}, {
					Amount: 40_000,
				}, {
					Amount: 60_000,
				},
			},
			Recipients: []Recipient{
				&RecipientImpl{
					Address: "bc1qua7e852suw0p74e2lzxwmk2tw8fd2zuzexc866",
					Amount:  50_000,
				},
			},
			FeeRate:         10,
			MinChangeAmount: 5000,
		},
		Expected: struct {
			Change             uint64
			NumOfSelectedUTXOs int
			AbsolutFee         uint64
			Err                error
		}{Change: 8007, NumOfSelectedUTXOs: 2, AbsolutFee: 1993, Err: nil},
	},
	{
		Comment: "fails because not enough funds",
		Given: struct {
			Utxos           []*UTXO
			Recipients      []Recipient
			FeeRate         uint32
			MinChangeAmount uint64
		}{
			Utxos: []*UTXO{
				{
					Amount: 20_000,
				},
			},
			Recipients: []Recipient{
				&RecipientImpl{
					Address: "bc1qua7e852suw0p74e2lzxwmk2tw8fd2zuzexc866",
					Amount:  20_000,
				},
			},
			FeeRate:         10,
			MinChangeAmount: 5000,
		},
		Expected: struct {
			Change             uint64
			NumOfSelectedUTXOs int
			AbsolutFee         uint64
			Err                error
		}{Change: 0, NumOfSelectedUTXOs: 0, AbsolutFee: 0, Err: ErrInsufficientFunds},
	},
	{
		Comment: "sp recipient",
		Given: struct {
			Utxos           []*UTXO
			Recipients      []Recipient
			FeeRate         uint32
			MinChangeAmount uint64
		}{
			Utxos: []*UTXO{
				{
					Amount: 20_000,
				},
			},
			Recipients: []Recipient{
				&RecipientImpl{
					Address: "tsp1qqfqnnv8czppwysafq3uwgwvsc638hc8rx3hscuddh0xa2yd746s7xq36vuz08htp29hyml4u9shtlvcvqxuhjzldxjwyfnxmamz3ft8mh5tzx0hu",
					Amount:  10_000,
				},
			},
			FeeRate:         10,
			MinChangeAmount: 5000,
		},
		Expected: struct {
			Change             uint64
			NumOfSelectedUTXOs int
			AbsolutFee         uint64
			Err                error
		}{Change: 8460, NumOfSelectedUTXOs: 1, AbsolutFee: 1540, Err: nil},
	},
	{
		Comment: "two sp recipient",
		Given: struct {
			Utxos           []*UTXO
			Recipients      []Recipient
			FeeRate         uint32
			MinChangeAmount uint64
		}{
			Utxos: []*UTXO{
				{
					Amount: 20_000,
				},
			},
			Recipients: []Recipient{
				&RecipientImpl{
					Address: "tsp1qqfqnnv8czppwysafq3uwgwvsc638hc8rx3hscuddh0xa2yd746s7xq36vuz08htp29hyml4u9shtlvcvqxuhjzldxjwyfnxmamz3ft8mh5tzx0hu",
					Amount:  10_000,
				},
				&RecipientImpl{
					Address: "tsp1qqfqnnv8czppwysafq3uwgwvsc638hc8rx3hscuddh0xa2yd746s7xq36vuz08htp29hyml4u9shtlvcvqxuhjzldxjwyfnxmamz3ft8mh5tzx0hu",
					Amount:  10_000,
				},
			},
			FeeRate:         10,
			MinChangeAmount: 5000,
		},
		Expected: struct {
			Change             uint64
			NumOfSelectedUTXOs int
			AbsolutFee         uint64
			Err                error
		}{Change: 0, NumOfSelectedUTXOs: 0, AbsolutFee: 0, Err: ErrInsufficientFunds},
	},
}

func TestFeeRateCoinSelector_CoinSelect(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.Comment, func(t *testing.T) {
			selector := NewFeeRateCoinSelector(tc.Given.Utxos, tc.Given.MinChangeAmount, tc.Given.Recipients, &chaincfg.MainNetParams)
			selectedUTXOs, change, err := selector.CoinSelect(tc.Given.FeeRate)

			if tc.Expected.Err != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.Expected.Err, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.Expected.Change, change)
			assert.Equal(t, tc.Expected.NumOfSelectedUTXOs, len(selectedUTXOs))

			// Calculate absolute fee
			var sumSelectedAmounts uint64
			for _, utxo := range selectedUTXOs {
				sumSelectedAmounts += utxo.Amount
			}

			var sumRecipientsAmounts uint64
			for _, recipient := range tc.Given.Recipients {
				sumRecipientsAmounts += recipient.GetAmount()
			}

			absoluteFee := sumSelectedAmounts - (sumRecipientsAmounts + change)
			assert.Equal(t, tc.Expected.AbsolutFee, absoluteFee)
		})
	}
}
