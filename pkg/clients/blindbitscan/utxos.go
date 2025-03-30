package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"encoding/hex"

	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
)

// GetUTXOs calls the GET /utxos endpoint.
func (c *Client) GetUTXOs(ctx context.Context) ([]*wallet.OwnedUTXO, error) {
	url := fmt.Sprintf("%s/utxos", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var utxos utxosResponse
	if err := json.NewDecoder(resp.Body).Decode(&utxos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert utxosResponse to []*wallet.OwnedUTXO
	result := make([]*wallet.OwnedUTXO, len(utxos))
	for i, u := range utxos {
		var label *wallet.Label
		if u.Label != nil {
			pubKey, _ := hex.DecodeString(u.Label.PubKey)
			tweak, _ := hex.DecodeString(u.Label.Tweak)
			label = &wallet.Label{
				PubKey:  hex.EncodeToString(pubKey),
				Tweak:   hex.EncodeToString(tweak),
				Address: u.Label.Address,
				M:       u.Label.M,
			}
		}

		// Convert state string to UTXOState
		var state wallet.UTXOState
		switch u.State {
		case "spent":
			state = wallet.StateSpent
		case "unspent":
			state = wallet.StateUnspent
		case "unconfirmed":
			state = wallet.StateUnconfirmed
		default:
			state = wallet.StateUnknown
		}

		result[i] = &wallet.OwnedUTXO{
			Txid:      u.Txid,
			Vout:      u.Vout,
			Amount:    u.Amount,
			PubKey:    u.PubKey,
			Timestamp: u.Timestamp,
			State:     state,
			Label:     label,
		}
	}

	return result, nil
}
